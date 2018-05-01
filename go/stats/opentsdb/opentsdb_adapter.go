// Package opentsdb has support for opentsdb integration at Pinterest.
//
// Usage:
// $ vtgate <...> -emit_stats -stats_emit_period 1m -stats_backend opentsdb --opentsdb_service vtgate_test
//
// TODO(dweitzman):
// 1) Use an off-the-shelf opentsdb HTTP client instead of the weird code copied from magnus.
//    Allegedly the localhost:18126 client supports HTTP also:
//    https://pinterest.slack.com/archives/C0455L136/p1521827366000043
//
// 2) We're using https://github.com/rcrowley/go-metrics/ to get percentiles for timing stats. We have some flexibility
//    about what algorithm we use. We should maybe think about that. Ideally in the future we'd be able to send
//    histograms directly to tcollector.
package opentsdb

import (
	"encoding/json"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"vitess.io/vitess/go/netutil"
	"vitess.io/vitess/go/stats"
	"vitess.io/vitess/go/vt/log"
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

	percentileBuckets     = []float64{.5, .9, .95, .99, .999}
	percentileBucketNames = []string{"p50", "p90", "p95", "p99", "p999"}
)

// One millisecond in nanoseconds, for ease of conversion.
// Vitess likes to report nanoseconds, but traditionally at pinterest we report millis.
const millisecond = 1000000

// openTSDBBackend implements stats.PushBackend
type openTSDBBackend struct {
	prefix   string
	metadata *Metadata
}

// ByKey imports sort.Interface for []Metric based on the metric key
type ByKey []Metric

func (m ByKey) Len() int           { return len(m) }
func (m ByKey) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m ByKey) Less(i, j int) bool { return m[i].Key < m[j].Key }

// Init attempts to create a singleton openTSDBBackend and register it as a PushBackend.
// If it fails to create one, this is a noop.
func Init(prefix string) {
	// Needs to happen in servenv.OnRun() instead of init because it requires flag parsing and logging
	servenv.OnRun(func() {
		if *openTsdbService == "" {
			return
		}

		metadata := &Metadata{
			GitSha:   buildGitRev,
			Hostname: netutil.FullyQualifiedHostnameOrPanic(),
			Service:  *openTsdbService,
		}

		backend := &openTSDBBackend{
			prefix:   prefix,
			metadata: metadata,
		}

		stats.RegisterPushBackend("opentsdb", backend)

		http.HandleFunc("/debug/opentsdb", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			metrics := (*backend).getMetrics()
			sort.Sort(ByKey(metrics))

			if b, err := json.MarshalIndent(metrics, "", "  "); err != nil {
				w.Write([]byte(err.Error()))
			} else {
				w.Write(b)
			}
		})
	})
}

// PushAll pushes all stats to OpenTSDB
func (backend *openTSDBBackend) PushAll() error {
	if backend.metadata.Service == "" {
		return ErrNoServiceName
	}

	backend.metadata.Timestamp = time.Now().Unix()
	return Send(*openTsdbHost, backend.getMetrics(), backend.metadata)
}

// getMetrics fetches all metrics in an opentsdb-compatible format.
// This is separated from PushAll() so it can be reused for the /debug/opentsdb handler.
func (backend *openTSDBBackend) getMetrics() []Metric {
	metrics := make([]Metric, 0)
	expvar.Do(func(kv expvar.KeyValue) {
		backend.addMetrics(&metrics, kv)
	})
	return metrics
}

// formatInt turns an int into a string that opentsdb can parse.
func formatInt(v int64) string {
	return fmt.Sprintf("%d", v)
}

// formatFloat turns a float into a string that opentsdb can parse.
func formatFloat(v float64) string {
	return strings.Replace(fmt.Sprintf("%g", v), "+", "", -1)
}

// combineMetricName joins parts of a hierachical name with a "."
func combineMetricName(parts ...string) string {
	return strings.Join(parts, ".")
}

// addMetrics adds all the metrics associated with a particular expvar to the list of
// opentsdb metrics. How an expvar is translated depends on its type.
//
// Well-known metric types like histograms and integers are directly converted (saving labels
// as tags).
//
// Generic unrecognized expvars are serialized to json and their int/float values are exported.
// Strings and lists in expvars are not exported.
func (backend *openTSDBBackend) addMetrics(metrics *[]Metric, kv expvar.KeyValue) {
	var k string
	if len(backend.prefix) > 0 {
		k = combineMetricName(backend.prefix, kv.Key)
	} else {
		k = kv.Key
	}
	switch v := kv.Value.(type) {
	case *stats.Float:
		*metrics = append(*metrics, Metric{
			Key:   k,
			Value: formatFloat(v.Get()),
		})
	case stats.FloatFunc:
		*metrics = append(*metrics, Metric{
			Key:   k,
			Value: formatFloat(v()),
		})
	case *stats.Counter:
		*metrics = append(*metrics, Metric{
			Key:   k,
			Value: formatInt(v.Get()),
		})
	case *stats.CounterFunc:
		*metrics = append(*metrics, Metric{
			Key:   k,
			Value: formatFloat(v.Mf.FloatVal()),
		})
	case *stats.Gauge:
		*metrics = append(*metrics, Metric{
			Key:   k,
			Value: formatInt(v.Get()),
		})
	case *stats.GaugeFunc:
		*metrics = append(*metrics, Metric{
			Key:   k,
			Value: formatFloat(v.Mf.FloatVal()),
		})
	case stats.IntFunc:
		*metrics = append(*metrics, Metric{
			Key:   k,
			Value: formatInt(v()),
		})
	case *stats.Duration:
		*metrics = append(*metrics, Metric{
			Key:   k,
			Value: formatInt(int64(v.Get() / time.Millisecond)),
		})
	case stats.DurationFunc:
		*metrics = append(*metrics, Metric{
			Key:   k,
			Value: formatInt(int64(v() / time.Millisecond)),
		})
	case *stats.MultiTimings:
		addTimings(metrics, v.Labels(), &v.Timings, k)
	case *stats.Timings:
		addTimings(metrics, []string{"Histograms"}, v, k)
	case *stats.Histogram:
		addHistogram(metrics, v, 1, k, make(map[string]string))
	case *stats.CountersWithLabels:
		for labelVal, val := range v.Counts() {
			*metrics = append(*metrics, Metric{
				Key:   k,
				Value: formatInt(val),
				Tags:  makeLabel(v.LabelName(), labelVal),
			})
		}
	case *stats.CountersWithMultiLabels:
		for labelVals, val := range v.Counts() {
			*metrics = append(*metrics, Metric{
				Key:   k,
				Value: formatInt(val),
				Tags:  makeLabels(v.Labels(), labelVals),
			})
		}
	case *stats.CountersFuncWithMultiLabels:
		for labelVals, val := range v.Counts() {
			*metrics = append(*metrics, Metric{
				Key:   k,
				Value: formatInt(val),
				Tags:  makeLabels(v.Labels(), labelVals),
			})
		}
	case *stats.GaugesWithMultiLabels:
		for labelVals, val := range v.Counts() {
			*metrics = append(*metrics, Metric{
				Key:   k,
				Value: formatInt(val),
				Tags:  makeLabels(v.Labels(), labelVals),
			})
		}
	case *stats.GaugesFuncWithMultiLabels:
		for labelVals, val := range v.Counts() {
			*metrics = append(*metrics, Metric{
				Key:   k,
				Value: formatInt(val),
				Tags:  makeLabels(v.Labels(), labelVals),
			})
		}
	case *stats.GaugesWithLabels:
		for labelVal, val := range v.Counts() {
			*metrics = append(*metrics, Metric{
				Key:   k,
				Value: formatInt(val),
				Tags:  makeLabel(v.LabelName(), labelVal),
			})
		}
	default:
		// Deal with generic expvars by converting them to JSON and pulling out
		// all the ints and floats. Strings and lists will not be exported to
		// opentsdb.
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(v.String()), &obj); err != nil {
			return
		}

		if len(backend.prefix) > 0 {
			k = combineMetricName(backend.prefix, "expvar", kv.Key)
		} else {
			k = combineMetricName("expvar", kv.Key)
		}

		// Recursive helper function.
		addUnrecognizedExpvars(metrics, k, obj)
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
	for i, v := range labelVals {
		tags[labelNames[i]] = v
	}
	return tags
}

// addUnrecognizedExpvars recurses into a json object to pull out float64 variables to report.
func addUnrecognizedExpvars(metrics *[]Metric, prefix string, obj map[string]interface{}) {
	for k, v := range obj {
		prefix := combineMetricName(prefix, k)
		switch v := v.(type) {
		case map[string]interface{}:
			addUnrecognizedExpvars(metrics, prefix, v)
		case float64:
			*metrics = append(*metrics, Metric{
				Key:   prefix,
				Value: formatFloat(v),
			})
		}
	}
}

// addTimings converts a vitess Timings stat to something opentsdb can deal with.
func addTimings(metrics *[]Metric, labels []string, timings *stats.Timings, prefix string) {
	histograms := timings.Histograms()
	for labelValsCombined, histogram := range histograms {
		addHistogram(metrics, histogram, millisecond, prefix, makeLabels(labels, labelValsCombined))
	}
}

// addHistogram gets the secret go-metrics Timer we've hidden alongside each vitess Histogram
// struct and uses it to calculate percentile metrics (which the vitess Histogram type doesn't do).
func addHistogram(metrics *[]Metric, histogram *stats.Histogram, divideBy int64, prefix string, tags map[string]string) {
	percentilesTimer := stats.GetTimer(histogram)
	if percentilesTimer == nil {
		log.Errorf("Pinterest hook for tracking percentiles failed to register!")
		// TODO: Maybe also increment an internal error counter?
		return
	}
	percentiles := (*percentilesTimer).Percentiles(percentileBuckets)
	for i, v := range percentiles {
		*metrics = append(*metrics, Metric{
			Key:   combineMetricName(prefix, percentileBucketNames[i]),
			Value: formatFloat(v / float64(divideBy)),
			Tags:  tags,
		})
	}
	*metrics = append(*metrics, Metric{
		Key:   combineMetricName(prefix, "average"),
		Value: formatFloat((*percentilesTimer).Mean() / float64(divideBy)),
		Tags:  tags,
	})
	*metrics = append(*metrics, Metric{
		Key:   combineMetricName(prefix, "rate1"),
		Value: formatFloat((*percentilesTimer).Rate1()),
		Tags:  tags,
	})
	*metrics = append(*metrics, Metric{
		Key:   combineMetricName(prefix, histogram.CountLabel()),
		Value: formatInt((*histogram).Count()),
		Tags:  tags,
	})
	*metrics = append(*metrics, Metric{
		Key:   combineMetricName(prefix, histogram.TotalLabel()),
		Value: formatInt((*histogram).Total() / divideBy),
		Tags:  tags,
	})
}
