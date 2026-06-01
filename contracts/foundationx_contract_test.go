package contracts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	foundationx "github.com/ZoneCNH/foundationx"
)

func TestFoundationxCompatibilitySurface(t *testing.T) {
	secret := foundationx.NewSecretString("contract-secret")
	if got := fmt.Sprint(secret); got != "***" {
		t.Fatalf("secret string should redact, got %q", got)
	}
	encodedSecret, err := json.Marshal(secret)
	if err != nil {
		t.Fatalf("marshal secret: %v", err)
	}
	if strings.Contains(string(encodedSecret), secret.Reveal()) {
		t.Fatalf("secret JSON leaked raw value: %s", encodedSecret)
	}
	if _, ok := any(secret).(foundationx.Sanitizer); !ok {
		t.Fatal("SecretString must implement Sanitizer")
	}

	cause := errors.New("downstream unavailable")
	foundationErr := foundationx.WrapError(foundationx.ErrorKindTimeout, "contracts", "timed out", cause).WithRetryable(true)
	if !foundationx.IsKind(foundationErr, foundationx.ErrorKindTimeout) || !errors.Is(foundationErr, cause) {
		t.Fatalf("foundation error helpers did not preserve kind/cause: %v", foundationErr)
	}
	if got, ok := foundationx.AsFoundationError(fmt.Errorf("outer: %w", foundationErr)); !ok || !got.Retryable {
		t.Fatalf("AsFoundationError mismatch: %#v ok=%v", got, ok)
	}
	encodedError, err := json.Marshal(foundationErr)
	if err != nil {
		t.Fatalf("marshal foundation error: %v", err)
	}
	for _, want := range []string{`"kind"`, `"op"`, `"message"`, `"retryable"`} {
		if !strings.Contains(string(encodedError), want) {
			t.Fatalf("foundation error JSON missing %s: %s", want, encodedError)
		}
	}
	if strings.Contains(string(encodedError), "downstream unavailable") {
		t.Fatalf("foundation error JSON leaked cause: %s", encodedError)
	}

	now := time.Unix(0, 0).UTC()
	status := foundationx.NewHealthStatus("configx", foundationx.HealthHealthy, "ok", now, 1).WithMetadata("module", "configx")
	if !status.IsHealthy() || status.Metadata["module"] != "configx" {
		t.Fatalf("health status mismatch: %#v", status)
	}
	if _, ok := any((*mockHealthChecker)(nil)).(foundationx.HealthChecker); !ok {
		t.Fatal("HealthChecker interface is not compatible")
	}

	policy := foundationx.DefaultRetryPolicy()
	if err := policy.Validate(); err != nil {
		t.Fatalf("default retry policy invalid: %v", err)
	}
	if got := policy.Delay(3); got != 400*time.Millisecond {
		t.Fatalf("retry delay attempt 3 = %s", got)
	}
	clock := foundationx.NewFixedClock(now)
	if !clock.Now().Equal(now) {
		t.Fatalf("fixed clock returned %s", clock.Now())
	}
	info := foundationx.NewVersionInfo("github.com/ZoneCNH/foundationx", "v0.1.0", "deadbeef", "2026-06-01T00:00:00Z", "go1.23")
	if !strings.Contains(info.String(), "foundationx v0.1.0") {
		t.Fatalf("version string mismatch: %s", info.String())
	}
}

func TestFoundationxErrorKindContract(t *testing.T) {
	got := []string{
		string(foundationx.ErrorKindConfig),
		string(foundationx.ErrorKindValidation),
		string(foundationx.ErrorKindConnection),
		string(foundationx.ErrorKindUnavailable),
		string(foundationx.ErrorKindTimeout),
		string(foundationx.ErrorKindAuth),
		string(foundationx.ErrorKindConflict),
		string(foundationx.ErrorKindRateLimit),
		string(foundationx.ErrorKindCanceled),
		string(foundationx.ErrorKindNotFound),
		string(foundationx.ErrorKindAlreadyExist),
		string(foundationx.ErrorKindInternal),
	}
	want := []string{"already_exists", "auth", "canceled", "config", "conflict", "connection", "internal", "not_found", "rate_limit", "timeout", "unavailable", "validation"}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("foundationx error kinds mismatch\ngot:  %v\nwant: %v", got, want)
	}
}

type mockHealthChecker struct{}

func (*mockHealthChecker) Name() string { return "mock" }

func (*mockHealthChecker) Check(context.Context) foundationx.HealthStatus {
	return foundationx.NewHealthStatus("mock", foundationx.HealthHealthy, "ok", time.Now(), 0)
}
