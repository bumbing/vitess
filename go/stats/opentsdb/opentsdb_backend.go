// Package opentsdb has support for opentsdb integration at Pinterest.
//
// Usage:
// $ vtgate <...> -emit_stats-stats_emit_period 1m -stats_backend opentsdb --opentsdb_service vtgate_test
//
// TODO(dweitzman):
// 1) Use an off-the-shelf opentsdb HTTP client instead of the weird code copied from magnus.
//    Allegedly the localhost:18126 client supports HTTP also:
//    https://pinterest.slack.com/archives/C0455L136/p1521827366000043
//
// 2) When re-implementing with the HTTP client, we can support histograms. For a Java service, if
//    you run "curl http://localhost:9999/stats.txt" the format for timing stats looks something like
//    this, which the ostrich sidecar will somehow interpret and upload to statsboard:
//    serviceframework.pepsi.UpdatePinnerLists.responseTime client_name=ngapi: (average=58, count=1, maximum=57, minimum=57, p50=57, p90=57, p95=57, p99=57, p999=57, p9999=5
//    https://github.com/simonratner/twitter-ostrich/blob/master/src/main/scala/com/twitter/ostrich/stats/Histogram.scala
//    is the code that actually secretly tracks histograms and converts them to percentiles under the covers for Java.
//
//    In vitess the default set of timing buckets is less extensive, probably from the expectation that queries will tend to be
//    faster. The buckets in go/stats/timings.go are .5ms, 1ms, 5ms, 10ms, 50ms, 100ms, .5s, 1s, 5s, 10s
//
// 3) Someday in the future if enough versions of opentsdb / statsboard are released and Pinterest upgrades,
//    we could use a histogram API calls (http://opentsdb.net/docs/build/html/api_http/histogram.html).
//    However that's not yet supported at Pinterest:
//    https://pinterest.slack.com/archives/C0455L136/p1521828598000596?thread_ts=1521827831.000400&cid=C0455L136
package opentsdb

import (
	"errors"
	"flag"
	"runtime"
	"time"

	"vitess.io/vitess/go/netutil"
	"vitess.io/vitess/go/stats"
	"vitess.io/vitess/go/vt/servenv"
)

var (
	// Most servers at pinterest have the opentsdb telnet protocol available on port 18126.
	openTsdbHost    = flag.String("opentsdb_host", "localhost:18126", "the opentsdb host (with port)")
	openTsdbService = flag.String("opentsdb_service", "", "the service name for opentsdb stats")
	buildGitRev     = "unknown"

	// ErrNoServiceName means pushing to OpenTSDB failed because -opentsdb_service was not passed on the command
	// line.
	ErrNoServiceName = errors.New("you must set -opentdsb_service <service name> when using -stats_backend OpenTSDB")
)

// openTSDBBackend implements stats.PushBackend
type openTSDBBackend struct {
	metadata *Metadata
}

// init attempts to create a singleton openTSDBBackend and register it as a PushBackend.
// If it fails to create one, this is a noop.
func init() {
	// Needs to happen in servenv.OnRun() instead of init because it requires flag parsing and logging
	servenv.OnRun(func() {
		if *openTsdbService == "" {
			return
		}

		metadata := &Metadata{
			GitSha:    buildGitRev,
			Hostname:  netutil.FullyQualifiedHostnameOrPanic(),
			Service:   *openTsdbService,
			GoVersion: runtime.Version(),
		}

		stats.RegisterPushBackend("opentsdb", &openTSDBBackend{
			metadata: metadata,
		})
	})
}

// PushAll pushes all expvar stats to OpenTSDB
func (backend *openTSDBBackend) PushAll() error {
	if backend.metadata.Service == "" {
		return ErrNoServiceName
	}

	backend.metadata.Timestamp = time.Now().Unix()
	metrics, err := GetExpVars()
	if err != nil {
		return err
	}
	return Send(*openTsdbHost, metrics, backend.metadata)
}
