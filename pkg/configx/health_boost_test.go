package configx

import (
	"context"
	"testing"
)

func TestHealthCheckNilContext(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	status := client.HealthCheck(nil)
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy, got %q", status.Status)
	}
	if status.Message != "context is required" {
		t.Fatalf("message = %q", status.Message)
	}
}

func TestHealthCheckNilClient(t *testing.T) {
	var c *Client
	status := c.HealthCheck(context.Background())
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy, got %q", status.Status)
	}
	if status.Name != "configx" {
		t.Fatalf("name = %q, want configx", status.Name)
	}
}

func TestHealthGaugeValue(t *testing.T) {
	if healthGaugeValue(HealthHealthy) != 1 {
		t.Fatal("expected 1 for healthy")
	}
	if healthGaugeValue(HealthUnhealthy) != 0 {
		t.Fatal("expected 0 for unhealthy")
	}
	if healthGaugeValue(HealthDegraded) != 0 {
		t.Fatal("expected 0 for degraded")
	}
}

func TestRecordHealthMetricNilMetrics(t *testing.T) {
	// Should not panic
	recordHealthMetric(nil, HealthStatus{Status: HealthHealthy})
}

func TestHealthCheckWithCustomName(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "myapp"})
	if err != nil {
		t.Fatal(err)
	}
	status := client.HealthCheck(context.Background())
	if status.Name != "myapp" {
		t.Fatalf("name = %q, want myapp", status.Name)
	}
}

func TestHealthCheckUninitializedClient(t *testing.T) {
	// Zero-value client is not initialized
	c := &Client{cfg: Config{Name: "test"}}
	status := c.HealthCheck(context.Background())
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy, got %q", status.Status)
	}
	if status.Message != "client is not initialized" {
		t.Fatalf("message = %q", status.Message)
	}
}
