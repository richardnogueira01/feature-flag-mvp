package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsUsesBoundedLabels(t *testing.T) {
	metrics := New(prometheus.NewRegistry())
	metrics.ObserveEvaluation(true, true, time.Now())
	metrics.ObserveEvaluation(false, false, time.Now())
}
