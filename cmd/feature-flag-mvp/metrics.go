package main

import (
	"github.com/prometheus/client_golang/prometheus"
	appmetrics "github.com/richardnogueira01/feature-flag-mvp/internal/metrics"
)

var operationalMetrics = appmetrics.NewOperational(prometheus.DefaultRegisterer)
