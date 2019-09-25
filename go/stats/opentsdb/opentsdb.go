// Package opentsdb adds support for pushing stats to opentsdb.
//
// Usage:
// $ vtgate <...> -emit_stats -stats_emit_period 1m -stats_backend opentsdb --opentsdb_service vtgate_test
//
// Visit /debug/opentsdb on a vtgate, vttablet, or vtctl server to see what would be reported to opentsdb.
//
// We're using https://github.com/rcrowley/go-metrics/ to get percentiles for timing stats. We have some flexibility
// about what algorithm we use. We should maybe think about that. Ideally in the future we'd be able to send
// histograms directly to tcollector (allegedly this support is forthcoming).
package opentsdb

import (
	"bytes"
	"encoding/json"
	"expvar"
	"flag"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"vitess.io/vitess/go/stats"
	"vitess.io/vitess/go/vt/log"
	"vitess.io/vitess/go/vt/servenv"
)

var (
	// Most servers at pinterest have the opentsdb telnet protocol available on port 18126.
	openTsdbURI = flag.String("opentsdb_uri", "http://localhost:18126/api/put?details", "URI of opentsdb /api/put method")
)

// dataPoint represents a single OpenTSDB data point.
type dataPoint struct {
	// Example: sys.cpu.nice
	Metric string `json:"metric"`
	// Seconds or milliseconds since unix epoch.
	Timestamp float64           `json:"timestamp"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
}

// sendDataPoints pushes a list of data points to openTSDB.
// All other code in this file is just to support getting this function called
// with all stats represented as data points.
func sendDataPoints(data []dataPoint) error {
	json, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(*openTsdbURI, "application/json", bytes.NewReader(json))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// openTSDBBackend implements stats.PushBackend
type openTSDBBackend struct {
	// The prefix is the name of the binary (vtgate, vttablet, etc.) and will be
	// prepended to all the stats reported.
	prefix string
	// Tags that should be included with every data point. If there's a tag name
	// collision between the common tags and a single data point's tags, the data
	// point tag will override the common tag.
	commonTags map[string]string
}

// dataCollector tracks state for a single pass of stats reporting / data collection.
type dataCollector struct {
	settings   *openTSDBBackend
	timestamp  int64
	dataPoints []dataPoint
}

// Init attempts to create a singleton openTSDBBackend and register it as a PushBackend.
// If it fails to create one, this is a noop. The prefix argument is an optional string
// to prepend to the name of every data point reported.
func Init(prefix string) {
	// Needs to happen in servenv.OnRun() instead of init because it requires flag parsing and logging
	servenv.OnRun(func() {
		if *openTsdbService == "" {
			return
		}

		gitRev := buildGitRev
		if gitRev == "" {
			gitRev = "empty"
		}

		hostname, err := os.Hostname()
		if err != nil {
			log.Exitf("Unable to determine hostname: %v", err)
		}

		backend := &openTSDBBackend{
			prefix: prefix,
			// If you want to global service values like host, service name, git revision, etc,
			// this is the place to do it.
			commonTags: map[string]string{
				"version": gitRev,
				"host":    hostname,
				"service": *openTsdbService,
			},
		}

		stats.RegisterPushBackend("opentsdb", backend)

		http.HandleFunc("/debug/opentsdb", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			dataPoints := (*backend).getDataPoints()
			sort.Sort(byMetric(dataPoints))

			if b, err := json.MarshalIndent(dataPoints, "", "  "); err != nil {
				w.Write([]byte(err.Error()))
			} else {
				w.Write(b)
			}
		})
	})
}

// PushAll pushes all stats to OpenTSDB
func (backend *openTSDBBackend) PushAll() error {
	return sendDataPoints(backend.getDataPoints())
}

// getDataPoints fetches all stats in an opentsdb-compatible format.
// This is separated from PushAll() so it can be reused for the /debug/opentsdb handler.
func (backend *openTSDBBackend) getDataPoints() []dataPoint {
	dataCollector := &dataCollector{
		settings:  backend,
		timestamp: time.Now().Unix(),
	}

	expvar.Do(func(kv expvar.KeyValue) {
		dataCollector.addExpVar(kv)
	})

	dataCollector.fillCommonTags()

	return dataCollector.dataPoints
}

// combineMetricName joins parts of a hierarchical name with a "."
func combineMetricName(parts ...string) string {
	return strings.Join(parts, ".")
}

func (dc *dataCollector) addInt(metric string, val int64, tags map[string]string) {
	dc.addFloat(metric, float64(val), tags)
}

func (dc *dataCollector) addFloat(metric string, val float64, tags map[string]string) {
	var fullMetric string
	if len(dc.settings.prefix) > 0 {
		fullMetric = combineMetricName(dc.settings.prefix, metric)
	} else {
		fullMetric = metric
	}

	fullTags := make(map[string]string)
	for k, v := range tags {
		fullTags[sanitize(k)] = sanitize(v)
	}

	dp := dataPoint{
		Metric:    sanitize(fullMetric),
		Value:     val,
		Timestamp: float64(dc.timestamp),
		Tags:      fullTags,
	}
	dc.dataPoints = append(dc.dataPoints, dp)
}

// addExpVar adds all the data points associated with a particular expvar to the list of
// opentsdb data points. How an expvar is translated depends on its type.
//
// Well-known metric types like histograms and integers are directly converted (saving labels
// as tags).
//
// Generic unrecognized expvars are serialized to json and their int/float values are exported.
// Strings and lists in expvars are not exported.
func (dc *dataCollector) addExpVar(kv expvar.KeyValue) {
	k := kv.Key
	switch v := kv.Value.(type) {
	case stats.FloatFunc:
		dc.addFloat(k, v(), nil)
	case *stats.Counter:
		dc.addInt(k, v.Get(), nil)
	case *stats.CounterFunc:
		dc.addInt(k, v.F(), nil)
	case *stats.Gauge:
		dc.addInt(k, v.Get(), nil)
	case *stats.GaugeFunc:
		dc.addInt(k, v.F(), nil)
	case *stats.CounterDuration:
		dc.addInt(k, int64(v.Get()/time.Millisecond), nil)
	case *stats.CounterDurationFunc:
		dc.addInt(k, int64(v.F()/time.Millisecond), nil)
	case *stats.MultiTimings:
		dc.addTimings(v.Labels(), &v.Timings, k)
	case *stats.Timings:
		dc.addTimings([]string{"Histograms"}, v, k)
	case *stats.Histogram:
		dc.addHistogram(v, 1, k, make(map[string]string))
	case *stats.CountersWithSingleLabel:
		for labelVal, val := range v.Counts() {
			dc.addInt(k, val, makeLabel(v.Label(), labelVal))
		}
	case *stats.CountersWithMultiLabels:
		for labelVals, val := range v.Counts() {
			dc.addInt(k, val, makeLabels(v.Labels(), labelVals))
		}
	case *stats.CountersFuncWithMultiLabels:
		for labelVals, val := range v.Counts() {
			dc.addInt(k, val, makeLabels(v.Labels(), labelVals))
		}
	case *stats.GaugesWithMultiLabels:
		for labelVals, val := range v.Counts() {
			dc.addInt(k, val, makeLabels(v.Labels(), labelVals))
		}
	case *stats.GaugesFuncWithMultiLabels:
		for labelVals, val := range v.Counts() {
			dc.addInt(k, val, makeLabels(v.Labels(), labelVals))
		}
	case *stats.GaugesWithSingleLabel:
		for labelVal, val := range v.Counts() {
			dc.addInt(k, val, makeLabel(v.Label(), labelVal))
		}
	case *stats.String:
		val := v.Get()
		switch k {
		case "TabletType":
			dc.settings.commonTags["dbtype"] = val
		case "TabletShard":
			dc.settings.commonTags["shard"] = val
		case "TabletKeyspace":
			dc.settings.commonTags["keyspace"] = val
		}
	default:
		// Deal with generic expvars by converting them to JSON and pulling out
		// all the floats. Strings and lists will not be exported to opentsdb.
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(v.String()), &obj); err != nil {
			return
		}

		// Recursive helper function.
		dc.addUnrecognizedExpvars(combineMetricName("expvar", k), obj)
	}
}

// makeLabel builds a tag list with a single label + value.
func makeLabel(labelName string, labelVal string) map[string]string {
	return map[string]string{labelName: labelVal}
}

// makeLabels takes the vitess stat representation of label values ("."-separated list) and breaks it
// apart into a map of label name -> label value.
func makeLabels(labelNames []string, labelValsCombined string) map[string]string {
	tags := make(map[string]string)
	labelVals := strings.Split(labelValsCombined, ".")

	if len(labelVals) != len(labelNames) {
		log.Fatalf("Internal error: expected tags %v but got wrong number: %s", labelNames, labelValsCombined)
		return tags
	}

	for i, v := range labelVals {
		tags[labelNames[i]] = v
	}
	return tags
}

// addUnrecognizedExpvars recurses into a json object to pull out float64 variables to report.
func (dc *dataCollector) addUnrecognizedExpvars(prefix string, obj map[string]interface{}) {
	for k, v := range obj {
		prefix := combineMetricName(prefix, k)
		switch v := v.(type) {
		case map[string]interface{}:
			dc.addUnrecognizedExpvars(prefix, v)
		case float64:
			dc.addFloat(prefix, v, nil)
		}
	}
}

// addTimings converts a vitess Timings stat to something opentsdb can deal with.
func (dc *dataCollector) addTimings(labels []string, timings *stats.Timings, prefix string) {
	histograms := timings.Histograms()
	for labelValsCombined, histogram := range histograms {
		// If you prefer millisecond timings over nanoseconds you can pass 1000000 here instead of 1.
		dc.addHistogram(histogram, millisecond, prefix, makeLabels(labels, labelValsCombined))
	}
}

// addHistogram gets the secret go-metrics Timer we've hidden alongside each vitess Histogram
// struct and uses it to calculate percentile metrics (which the vitess Histogram type doesn't do).
func (dc *dataCollector) addHistogram(histogram *stats.Histogram, divideBy int64, prefix string, tags map[string]string) {
	rawTimer := (*stats.GetTimer(histogram))

	// Only report histogram data if we saw queries within the last minute.
	if rawTimer.Rate1() >= 1.0/60.0 {
		percentilesTimer := rawTimer
		if percentilesTimer == nil {
			log.Errorf("Pinterest hook for tracking percentiles failed to register!")
			// TODO: Maybe also increment an internal error counter?
			return
		}

		percentiles := percentilesTimer.Percentiles(percentileBuckets)
		for i, v := range percentiles {
			dc.addFloat(combineMetricName(prefix, percentileBucketNames[i]), v/float64(divideBy), tags)
		}
		dc.addFloat(
			combineMetricName(prefix, "average"),
			percentilesTimer.Mean()/float64(divideBy),
			tags,
		)
		dc.addFloat(
			combineMetricName(prefix, "rate1"),
			percentilesTimer.Rate1(),
			tags,
		)
	}

	dc.addInt(
		combineMetricName(prefix, histogram.CountLabel()),
		(*histogram).Count(),
		tags,
	)
	dc.addInt(
		combineMetricName(prefix, histogram.TotalLabel()),
		(*histogram).Total()/divideBy,
		tags,
	)
}

// byMetric implements sort.Interface for []dataPoint based on the metric key
// and then tag values (prioritized in tag name order). Having a consistent sort order
// is convenient when refreshing /debug/opentsdb or for encoding and comparing JSON directly
// in the tests.
type byMetric []dataPoint

func (m byMetric) Len() int      { return len(m) }
func (m byMetric) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m byMetric) Less(i, j int) bool {
	if m[i].Metric < m[j].Metric {
		return true
	}

	if m[i].Metric > m[j].Metric {
		return false
	}

	// Metric names are the same. We can use tag values to figure out the sort order.
	// The deciding tag will be the lexicographically earliest tag name where tag values differ.
	decidingTagName := ""
	result := false
	for tagName, iVal := range m[i].Tags {
		jVal, ok := m[j].Tags[tagName]
		if !ok {
			// We'll arbitrarily declare that if i has any tag name that j doesn't then it sorts earlier.
			// This shouldn't happen in practice, though, if metric code is correct...
			return true
		}

		if iVal != jVal && (tagName < decidingTagName || decidingTagName == "") {
			decidingTagName = tagName
			result = iVal < jVal
		}
	}
	return result
}
