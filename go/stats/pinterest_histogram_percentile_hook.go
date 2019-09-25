package stats

import (
	"sync"
	"time"

	metrics "github.com/rcrowley/go-metrics"
)

var (
	histogramToTimerMu sync.RWMutex
	histogramToTimer   = make(map[*Histogram]*metrics.Timer)
	fakeTimer          *metrics.Timer
)

// GetTimer maps a Histogram to its associated Timer, which has percentiles, mean, min, and max.
// Background: The vitess "Histogram" type is a plain old histogram. OpenTSDB doesn't have any
// graceful way to ingest histogram data and calculate percentiles, so there's some hackery here
// to also increment the open source go-metrics library's Timer object alongside the Histogram
// type. There's a mapping to get the Timer associated with a Histogram, and the timer can
// return percentile information that the histogram can't / won't.
func GetTimer(histogram *Histogram) *metrics.Timer {
	histogramToTimerMu.RLock()
	defer histogramToTimerMu.RUnlock()
	return histogramToTimer[histogram]
}

// SetFakeTimerForTest sets a fake timer that will always be associated
// with histograms.
func SetFakeTimerForTest(timer *metrics.Timer) {
	fakeTimer = timer
}

func makeHistogramHook(histogramPtr *Histogram) func(int64) {
	var percentilesTimer metrics.Timer
	if fakeTimer != nil {
		percentilesTimer = *fakeTimer
	} else {
		percentilesTimer = metrics.NewTimer()
	}

	histogramToTimerMu.Lock()
	histogramToTimer[histogramPtr] = &percentilesTimer
	histogramToTimerMu.Unlock()

	return func(duration int64) {
		percentilesTimer.Update(time.Duration(duration))
	}
}
