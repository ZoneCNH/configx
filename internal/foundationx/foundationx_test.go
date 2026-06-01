package foundationx

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSecretStringRedactsDisplayAndMarshalPaths(t *testing.T) {
	secret := NewSecretString("raw-secret-value")

	checks := []string{
		secret.String(),
		fmt.Sprintf("%#v", secret),
		string(mustMarshalText(t, secret)),
		string(mustMarshalJSON(t, secret)),
		fmt.Sprint(secret.Sanitize()),
	}
	for _, got := range checks {
		if strings.Contains(got, secret.Reveal()) {
			t.Fatalf("secret leaked through safe path: %q", got)
		}
	}
	if secret.Reveal() != "raw-secret-value" {
		t.Fatalf("reveal returned %q", secret.Reveal())
	}
	if NewSecretString("").String() != "" || !NewSecretString("").IsZero() {
		t.Fatal("empty secret should stay empty and report zero")
	}
}

func TestErrorHelpersExposeCompatibleShape(t *testing.T) {
	cause := errors.New("network unavailable")
	err := WrapError(ErrorKindUnavailable, "foundationx.test", "not available", cause).WithRetryable(true)

	if !IsKind(err, ErrorKindUnavailable) {
		t.Fatal("IsKind did not match wrapped foundation error")
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped cause was not preserved")
	}
	got, ok := AsFoundationError(fmt.Errorf("outer: %w", err))
	if !ok || got.Kind != ErrorKindUnavailable || !got.Retryable {
		t.Fatalf("AsFoundationError mismatch: %#v ok=%v", got, ok)
	}
	encoded := string(mustMarshalJSON(t, err))
	for _, want := range []string{`"kind"`, `"op"`, `"message"`, `"retryable"`} {
		if !strings.Contains(encoded, want) {
			t.Fatalf("encoded error missing %s: %s", want, encoded)
		}
	}
	if strings.Contains(encoded, "Cause") || strings.Contains(encoded, "network unavailable") {
		t.Fatalf("encoded error leaked cause detail: %s", encoded)
	}
}

func TestHealthRetryClockLifecycleAndVersionSurface(t *testing.T) {
	now := time.Unix(10, 0).UTC()
	clock := NewFixedClock(now)
	if !clock.Now().Equal(now) {
		t.Fatalf("fixed clock returned %s", clock.Now())
	}

	status := NewHealthStatus("configx", HealthHealthy, "ok", now, -1).WithMetadata("module", "configx")
	if !status.IsHealthy() || status.LatencyMs != 0 || status.Metadata["module"] != "configx" {
		t.Fatalf("unexpected health status: %#v", status)
	}

	policy := DefaultRetryPolicy()
	if err := policy.Validate(); err != nil {
		t.Fatalf("default policy should validate: %v", err)
	}
	if got := policy.Delay(3); got != 400*time.Millisecond {
		t.Fatalf("delay attempt 3 = %s", got)
	}
	if err := (RetryPolicy{MaxAttempts: 0, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}).Validate(); !IsKind(err, ErrorKindValidation) {
		t.Fatalf("invalid policy should return validation error, got %v", err)
	}

	info := NewVersionInfo("github.com/bytechainx/foundationx", "v0.1.0", "deadbeef", "2026-06-01T00:00:00Z", "go1.23")
	if !strings.Contains(info.String(), "github.com/bytechainx/foundationx v0.1.0") {
		t.Fatalf("unexpected version string: %s", info.String())
	}
}

func mustMarshalText(t *testing.T, secret SecretString) []byte {
	t.Helper()
	data, err := secret.MarshalText()
	if err != nil {
		t.Fatalf("marshal text: %v", err)
	}
	return data
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return data
}
