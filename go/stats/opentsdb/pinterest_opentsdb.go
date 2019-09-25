package opentsdb

import (
	"bytes"
	"flag"
	"strings"
	"unicode"
)

var (
	openTsdbService = flag.String("opentsdb_service", "", "the service name for opentsdb stats")
	// This will be filled in at build time (see tools/build_version_flags.sh)
	buildGitRev = ""

	percentileBuckets     = []float64{.5, .9, .95, .99, .999}
	percentileBucketNames = []string{"p50", "p90", "p95", "p99", "p999"}
)

// A millisecond in nanoseconds, for ease of conversion.
// Vitess likes to report nanoseconds, but traditionally at pinterest we report millis.
//
// NOTE(dweitzman): Ideally it seems simpler to report all times as nanoseconds, but
// that makes writing queries against the data in a way that reports milliseconds
// trickier.
const millisecond = 1000000

func (dc *dataCollector) fillCommonTags() {
	for idx := range dc.dataPoints {
		for k, v := range dc.settings.commonTags {
			if _, ok := dc.dataPoints[idx].Tags[k]; ok {
				// Let any explicit tag on a metric override a common tag.
				continue
			}
			dc.dataPoints[idx].Tags[sanitize(k)] = sanitize(v)
		}
	}
}

// Based on upstream opentsdb.go for use in fillCommonTags(), but also fills "empty"
//
// Restrict metric and tag name/values to legal characters:
// http://opentsdb.net/docs/build/html/user_guide/writing.html#metrics-and-tags
//
// Also make everything lowercase, since opentsdb is case sensitive and lowercase
// simplifies the convention.
func sanitize(text string) string {
	if text == "" {
		return "empty"
	}

	var b bytes.Buffer
	for _, r := range text {
		if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '-' || r == '_' || r == '/' || r == '.' {
			b.WriteRune(r)
		} else {
			// For characters that would cause errors, write underscore instead
			b.WriteRune('_')
		}
	}
	return strings.ToLower(b.String())
}
