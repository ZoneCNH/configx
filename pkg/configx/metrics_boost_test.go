package configx

import (
	"testing"
)

func TestNoopMetricsMethods(t *testing.T) {
	var m NoopMetrics
	// All methods should be callable without panic.
	m.IncCounter("test", nil)
	m.ObserveHistogram("test", 1.0, nil)
	m.SetGauge("test", 1.0, nil)
}

func TestNoopMetricsWithLabels(t *testing.T) {
	var m NoopMetrics
	labels := map[string]string{"key": "value"}
	m.IncCounter("test", labels)
	m.ObserveHistogram("test", 1.0, labels)
	m.SetGauge("test", 1.0, labels)
}
