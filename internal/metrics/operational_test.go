package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestOperationalMetricsCanObserveBoundedSeries(t *testing.T) {
	metrics := NewOperational(prometheus.NewRegistry())
	metrics.ObserveGap()
	metrics.ObserveResync(true)
	metrics.ObserveResync(false)
	metrics.ObserveOutbox(true)
	metrics.ObserveOutbox(false)
}
