package configx

import (
	"context"
	"testing"
)

func TestCloseNilClient(t *testing.T) {
	var c *Client
	err := c.Close(context.Background())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCloseNilContext(t *testing.T) {
	c, err := New(context.Background(), Config{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Close(nil)
	if err == nil {
		t.Fatal("expected error for nil context")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCloseExpiredContext(t *testing.T) {
	c, err := New(context.Background(), Config{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = c.Close(ctx)
	if err == nil {
		t.Fatal("expected error for expired context")
	}
	if !IsKind(err, ErrorKindUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func TestCloseAlreadyClosed(t *testing.T) {
	c, err := New(context.Background(), Config{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Second close should be a no-op (returns nil)
	if err := c.Close(context.Background()); err != nil {
		t.Fatalf("second close should return nil, got %v", err)
	}
}

func TestCloseUninitializedClient(t *testing.T) {
	c := &Client{cfg: Config{Name: "test"}}
	err := c.Close(context.Background())
	if err == nil {
		t.Fatal("expected error for uninitialized client")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestNewExpiredContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := New(ctx, Config{Name: "test"})
	if err == nil {
		t.Fatal("expected error for expired context")
	}
	if !IsKind(err, ErrorKindUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func TestNewWithMetrics(t *testing.T) {
	m := &recordingMetrics{}
	c, err := New(context.Background(), Config{Name: "test"}, WithMetrics(m))
	if err != nil {
		t.Fatal(err)
	}
	if !m.hasCounter(MetricClientCreatedTotal) {
		t.Fatal("expected created counter")
	}
	_ = c
}

func TestNewInvalidConfig(t *testing.T) {
	// Config with empty name should fail Validate
	_, err := New(context.Background(), Config{})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}
