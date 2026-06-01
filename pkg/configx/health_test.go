package configx

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestHealthCheckHealthy(t *testing.T) {
	metrics := &recordingMetrics{}
	client, err := New(context.Background(), Config{Name: "configx"}, WithMetrics(metrics))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	status := client.HealthCheck(context.Background())
	if status.Status != HealthHealthy {
		t.Fatalf("expected healthy status, got %q", status.Status)
	}
	if status.Name != "configx" {
		t.Fatalf("expected configx health name, got %q", status.Name)
	}
	if status.LatencyMs < 0 {
		t.Fatalf("expected non-negative latency, got %d", status.LatencyMs)
	}
	if !metrics.hasGauge(MetricClientHealthStatus) {
		t.Fatalf("expected health status gauge, got %#v", metrics.gauges)
	}
	if !metrics.hasHistogram(MetricClientHealthLatencyMS) {
		t.Fatalf("expected health latency histogram, got %#v", metrics.histograms)
	}
}

func TestHealthCheckClosedClientUnhealthy(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "configx"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Close(context.Background()); err != nil {
		t.Fatalf("close client: %v", err)
	}

	status := client.HealthCheck(context.Background())
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
}

func TestHealthCheckCanceledContextUnhealthy(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "configx"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	status := client.HealthCheck(ctx)
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
	if !strings.Contains(status.Message, context.Canceled.Error()) {
		t.Fatalf("expected canceled message, got %q", status.Message)
	}
}

func TestHealthCheckZeroValueClientUnhealthy(t *testing.T) {
	var client Client

	status := client.HealthCheck(context.Background())
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy status, got %q", status.Status)
	}
	if status.Name != "configx" {
		t.Fatalf("expected fallback health name, got %q", status.Name)
	}
}

func TestHealthStatusJSONContract(t *testing.T) {
	payload, err := json.Marshal(HealthStatus{
		Name:      "configx",
		Status:    HealthHealthy,
		LatencyMs: 7,
	})
	if err != nil {
		t.Fatalf("marshal health status: %v", err)
	}
	encoded := string(payload)
	for _, field := range []string{"name", "status", "checked_at", "latency_ms"} {
		if !strings.Contains(encoded, `"`+field+`"`) {
			t.Fatalf("expected JSON field %q in %s", field, encoded)
		}
	}
	if strings.Contains(encoded, "CheckedAt") || strings.Contains(encoded, "LatencyMs") {
		t.Fatalf("expected snake_case JSON fields, got %s", encoded)
	}
}
