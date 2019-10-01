package discovery

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"
	querypb "vitess.io/vitess/go/vt/proto/query"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
	"vitess.io/vitess/go/vt/topo"
)

func TestHealthCheck(t *testing.T) {
	tablet := topo.NewTablet(0, "cell", "a")
	tablet.PortMap["vt"] = 1
	input := make(chan *querypb.StreamHealthResponse)
	createFakeConn(tablet, input)
	t.Logf(`createFakeConn({Host: "a", PortMap: {"vt": 1}}, c)`)
	l := newListener()
	hc := NewHealthCheck(1*time.Millisecond, time.Hour).(*HealthCheckImpl)
	hc.SetListener(l, true)
	testChecksum(t, 0, hc.stateChecksum())
	hc.AddTablet(tablet, "")
	t.Logf(`hc = HealthCheck(); hc.AddTablet({Host: "a", PortMap: {"vt": 1}}, "")`)

	// Immediately after AddTablet() there will be the first notification.
	want := &TabletStats{
		Key:     "a,vt:1",
		Tablet:  tablet,
		Target:  &querypb.Target{},
		Up:      true,
		Serving: false,
	}
	res := <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}
	testChecksum(t, 401258919, hc.stateChecksum())

	// one tablet after receiving a StreamHealthResponse
	shr := &querypb.StreamHealthResponse{
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		Serving:                             true,
		TabletExternallyReparentedTimestamp: 10,
		RealtimeStats:                       &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
	}
	want = &TabletStats{
		Key:                                 "a,vt:1",
		Tablet:                              tablet,
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		Up:                                  true,
		Serving:                             true,
		Stats:                               &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
		TabletExternallyReparentedTimestamp: 10,
	}
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: MASTER}, Serving: true, TabletExternallyReparentedTimestamp: 10, {SecondsBehindMaster: 1, CpuUsage: 0.2}}`)
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}

	// Verify that the error count is initialized to 0 after the first tablet response.
	if err := checkErrorCounter("k", "s", topodatapb.TabletType_MASTER, 0); err != nil {
		t.Errorf("%v", err)
	}

	tcsl := hc.CacheStatus()
	tcslWant := TabletsCacheStatusList{{
		Cell:   "cell",
		Target: &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		TabletsStats: TabletStatsList{{
			Key:                                 "a,vt:1",
			Tablet:                              tablet,
			Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
			Up:                                  true,
			Serving:                             true,
			Stats:                               &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
			TabletExternallyReparentedTimestamp: 10,
		}},
	}}
	if !reflect.DeepEqual(tcsl, tcslWant) {
		t.Errorf("hc.CacheStatus() =\n%+v; want\n%+v", tcsl[0], tcslWant[0])
	}
	testChecksum(t, 1562785705, hc.stateChecksum())

	// TabletType changed, should get both old and new event
	shr = &querypb.StreamHealthResponse{
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Serving:                             true,
		TabletExternallyReparentedTimestamp: 0,
		RealtimeStats:                       &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.5},
	}
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: REPLICA}, Serving: true, TabletExternallyReparentedTimestamp: 0, {SecondsBehindMaster: 1, CpuUsage: 0.5}}`)
	want = &TabletStats{
		Key:                                 "a,vt:1",
		Tablet:                              tablet,
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		Up:                                  false,
		Serving:                             true,
		Stats:                               &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
		TabletExternallyReparentedTimestamp: 10,
	}
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}
	want = &TabletStats{
		Key:                                 "a,vt:1",
		Tablet:                              tablet,
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Up:                                  true,
		Serving:                             true,
		Stats:                               &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.5},
		TabletExternallyReparentedTimestamp: 0,
	}
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}

	if err := checkErrorCounter("k", "s", topodatapb.TabletType_REPLICA, 0); err != nil {
		t.Errorf("%v", err)
	}
	testChecksum(t, 1906892404, hc.stateChecksum())

	// Serving & RealtimeStats changed
	shr = &querypb.StreamHealthResponse{
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Serving:                             false,
		TabletExternallyReparentedTimestamp: 0,
		RealtimeStats:                       &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.3},
	}
	want = &TabletStats{
		Key:                                 "a,vt:1",
		Tablet:                              tablet,
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Up:                                  true,
		Serving:                             false,
		Stats:                               &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.3},
		TabletExternallyReparentedTimestamp: 0,
	}
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: REPLICA}, TabletExternallyReparentedTimestamp: 0, {SecondsBehindMaster: 1, CpuUsage: 0.3}}`)
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}
	testChecksum(t, 1200695592, hc.stateChecksum())

	// HealthError
	shr = &querypb.StreamHealthResponse{
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Serving:                             true,
		TabletExternallyReparentedTimestamp: 0,
		RealtimeStats:                       &querypb.RealtimeStats{HealthError: "some error", SecondsBehindMaster: 1, CpuUsage: 0.3},
	}
	want = &TabletStats{
		Key:                                 "a,vt:1",
		Tablet:                              tablet,
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Up:                                  true,
		Serving:                             false,
		Stats:                               &querypb.RealtimeStats{HealthError: "some error", SecondsBehindMaster: 1, CpuUsage: 0.3},
		TabletExternallyReparentedTimestamp: 0,
		LastError:                           fmt.Errorf("vttablet error: some error"),
	}
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: REPLICA}, Serving: true, TabletExternallyReparentedTimestamp: 0, {HealthError: "some error", SecondsBehindMaster: 1, CpuUsage: 0.3}}`)
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}
	testChecksum(t, 1200695592, hc.stateChecksum()) // unchanged

	// remove tablet
	hc.deleteConn(tablet)
	t.Logf(`hc.RemoveTablet({Host: "a", PortMap: {"vt": 1}})`)
	want = &TabletStats{
		Key:                                 "a,vt:1",
		Tablet:                              tablet,
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_REPLICA},
		Up:                                  false,
		Serving:                             false,
		Stats:                               &querypb.RealtimeStats{HealthError: "some error", SecondsBehindMaster: 1, CpuUsage: 0.3},
		TabletExternallyReparentedTimestamp: 0,
		LastError:                           context.Canceled,
	}
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf("<-l.output:\n%+v; want\n%+v", res, want)
	}
	testChecksum(t, 0, hc.stateChecksum())

	// close healthcheck
	hc.Close()
}

func TestHealthCheckTimeout(t *testing.T) {
	timeout := 500 * time.Millisecond
	tablet := topo.NewTablet(0, "cell", "a")
	tablet.PortMap["vt"] = 1
	input := make(chan *querypb.StreamHealthResponse)
	fc := createFakeConn(tablet, input)
	t.Logf(`createFakeConn({Host: "a", PortMap: {"vt": 1}}, c)`)
	l := newListener()
	hc := NewHealthCheck(1*time.Millisecond, timeout).(*HealthCheckImpl)
	hc.SetListener(l, false)
	hc.AddTablet(tablet, "")
	t.Logf(`hc = HealthCheck(); hc.AddTablet({Host: "a", PortMap: {"vt": 1}}, "")`)

	// Immediately after AddTablet() there will be the first notification.
	want := &TabletStats{
		Key:     "a,vt:1",
		Tablet:  tablet,
		Target:  &querypb.Target{},
		Up:      true,
		Serving: false,
	}
	res := <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}

	// one tablet after receiving a StreamHealthResponse
	shr := &querypb.StreamHealthResponse{
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		Serving:                             true,
		TabletExternallyReparentedTimestamp: 10,
		RealtimeStats:                       &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
	}
	want = &TabletStats{
		Key:                                 "a,vt:1",
		Tablet:                              tablet,
		Target:                              &querypb.Target{Keyspace: "k", Shard: "s", TabletType: topodatapb.TabletType_MASTER},
		Up:                                  true,
		Serving:                             true,
		Stats:                               &querypb.RealtimeStats{SecondsBehindMaster: 1, CpuUsage: 0.2},
		TabletExternallyReparentedTimestamp: 10,
	}
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: MASTER}, Serving: true, TabletExternallyReparentedTimestamp: 10, {SecondsBehindMaster: 1, CpuUsage: 0.2}}`)
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}

	if err := checkErrorCounter("k", "s", topodatapb.TabletType_MASTER, 0); err != nil {
		t.Errorf("%v", err)
	}

	// wait for timeout period
	time.Sleep(2 * timeout)
	t.Logf(`Sleep(2 * timeout)`)
	res = <-l.output
	if res.Serving {
		t.Errorf(`<-l.output: %+v; want not serving`, res)
	}

	if err := checkErrorCounter("k", "s", topodatapb.TabletType_MASTER, 1); err != nil {
		t.Errorf("%v", err)
	}

	if !fc.isCanceled() {
		t.Errorf("StreamHealth should be canceled after timeout, but is not")
	}

	// repeat the wait. It will timeout one more time trying to get the connection.
	fc.resetCanceledFlag()
	time.Sleep(timeout)
	t.Logf(`Sleep(2 * timeout)`)

	res = <-l.output
	if res.Serving {
		t.Errorf(`<-l.output: %+v; want not serving`, res)
	}

	if err := checkErrorCounter("k", "s", topodatapb.TabletType_MASTER, 2); err != nil {
		t.Errorf("%v", err)
	}

	if !fc.isCanceled() {
		t.Errorf("StreamHealth should be canceled again after timeout")
	}

	// send a healthcheck response, it should be serving again
	fc.resetCanceledFlag()
	input <- shr
	t.Logf(`input <- {{Keyspace: "k", Shard: "s", TabletType: MASTER}, Serving: true, TabletExternallyReparentedTimestamp: 10, {SecondsBehindMaster: 1, CpuUsage: 0.2}}`)

	// wait for the exponential backoff to wear off and health monitoring to resume.
	time.Sleep(timeout)
	res = <-l.output
	if !reflect.DeepEqual(res, want) {
		t.Errorf(`<-l.output: %+v; want %+v`, res, want)
	}

	// close healthcheck
	hc.Close()
}
