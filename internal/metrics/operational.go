package metrics

import "github.com/prometheus/client_golang/prometheus"

type OperationalMetrics struct {
	RevisionGaps prometheus.Counter
	Resyncs      *prometheus.CounterVec
	OutboxEvents *prometheus.CounterVec
}

func NewOperational(registerer prometheus.Registerer) *OperationalMetrics {
	m := &OperationalMetrics{
		RevisionGaps: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "feature_flag_revision_gaps_total",
			Help: "Total number of detected revision gaps.",
		}),
		Resyncs: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "feature_flag_resyncs_total",
			Help: "Total number of full resync attempts.",
		}, []string{"result"}),
		OutboxEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "feature_flag_outbox_events_total",
			Help: "Total number of outbox events processed.",
		}, []string{"result"}),
	}
	registerer.MustRegister(m.RevisionGaps, m.Resyncs, m.OutboxEvents)
	return m
}

func (m *OperationalMetrics) ObserveGap() {
	m.RevisionGaps.Inc()
}

func (m *OperationalMetrics) ObserveResync(success bool) {
	m.Resyncs.WithLabelValues(result(success)).Inc()
}

func (m *OperationalMetrics) ObserveOutbox(success bool) {
	m.OutboxEvents.WithLabelValues(result(success)).Inc()
}

func result(success bool) string {
	if success {
		return "success"
	}
	return "failure"
}
