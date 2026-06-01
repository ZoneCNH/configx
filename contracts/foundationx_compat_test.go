package contracts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	foundationx "github.com/ZoneCNH/foundationx"
)

func TestFoundationXCompatibilitySurface(t *testing.T) {
	kinds := []foundationx.ErrorKind{
		foundationx.ErrorKindConfig,
		foundationx.ErrorKindValidation,
		foundationx.ErrorKindConnection,
		foundationx.ErrorKindUnavailable,
		foundationx.ErrorKindTimeout,
		foundationx.ErrorKindAuth,
		foundationx.ErrorKindConflict,
		foundationx.ErrorKindRateLimit,
		foundationx.ErrorKindCanceled,
		foundationx.ErrorKindNotFound,
		foundationx.ErrorKindAlreadyExist,
		foundationx.ErrorKindInternal,
	}
	seen := map[foundationx.ErrorKind]bool{}
	for _, kind := range kinds {
		if kind == "" {
			t.Fatal("foundationx error kind must not be empty")
		}
		if seen[kind] {
			t.Fatalf("duplicate foundationx error kind %q", kind)
		}
		seen[kind] = true
	}

	cause := errors.New("upstream unavailable")
	err := foundationx.WrapError(foundationx.ErrorKindUnavailable, "Client.Ping", "ping failed", cause).WithRetryable(true)
	if !foundationx.IsKind(fmt.Errorf("outer: %w", err), foundationx.ErrorKindUnavailable) {
		t.Fatalf("IsKind did not identify wrapped foundationx error: %v", err)
	}
	if got, ok := foundationx.AsFoundationError(err); !ok || got != err {
		t.Fatalf("AsFoundationError() = (%v, %v), want original error", got, ok)
	}
	if !errors.Is(err, cause) || !err.Retryable {
		t.Fatalf("foundationx error did not preserve cause/retryable: %#v", err)
	}
}

func TestFoundationXErrorFormattingMatchesUpstream(t *testing.T) {
	cause := errors.New("cause")
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "operation",
			err:  foundationx.NewError(foundationx.ErrorKindValidation, "Config.Validate", "bad value"),
			want: "validation: Config.Validate: bad value",
		},
		{
			name: "no operation",
			err:  foundationx.NewError(foundationx.ErrorKindValidation, "", "bad value"),
			want: "validation: bad value",
		},
		{
			name: "empty message does not fall back to cause",
			err:  foundationx.WrapError(foundationx.ErrorKindValidation, "Config.Validate", "", cause),
			want: "validation: Config.Validate: ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFoundationXSecretAndUtilityContracts(t *testing.T) {
	const raw = "super-secret"
	secret := foundationx.NewSecretString(raw)
	var _ foundationx.Sanitizer = secret
	if got := fmt.Sprint(secret); got != "***" || strings.Contains(got, raw) {
		t.Fatalf("SecretString leaked raw value through fmt: %q", got)
	}
	if got := secret.Sanitize(); got != "***" {
		t.Fatalf("SecretString.Sanitize() = %v, want ***", got)
	}
	if secret.IsZero() || !foundationx.NewSecretString("").IsZero() {
		t.Fatal("SecretString.IsZero mismatch")
	}

	fixed := time.Date(2026, 6, 1, 1, 2, 3, 0, time.UTC)
	var clock foundationx.Clock = foundationx.NewFixedClock(fixed)
	if got := clock.Now(); !got.Equal(fixed) {
		t.Fatalf("FixedClock.Now() = %s, want %s", got, fixed)
	}
	var _ foundationx.Clock = foundationx.NewRealClock()

	policy := foundationx.DefaultRetryPolicy()
	if err := policy.Validate(); err != nil {
		t.Fatalf("DefaultRetryPolicy().Validate(): %v", err)
	}
	if got := policy.Delay(2); got != 200*time.Millisecond {
		t.Fatalf("RetryPolicy.Delay(2) = %s, want 200ms", got)
	}
	if err := (foundationx.RetryPolicy{MaxAttempts: 0}).Validate(); !foundationx.IsKind(err, foundationx.ErrorKindValidation) {
		t.Fatalf("invalid RetryPolicy error = %v, want validation kind", err)
	}
}

func TestFoundationXHealthLifecycleAndVersionContracts(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	baseStatus := foundationx.NewHealthStatus("config", foundationx.HealthHealthy, "ok", now, 7)
	baseStatus.Metadata["existing"] = "keep"
	status := baseStatus.WithMetadata("source", "explicit")
	if _, ok := baseStatus.Metadata["source"]; ok {
		t.Fatalf("WithMetadata mutated original metadata: %#v", baseStatus.Metadata)
	}
	if !status.IsHealthy() || status.Metadata["existing"] != "keep" || status.Metadata["source"] != "explicit" || !status.CheckedAt.Equal(now) {
		t.Fatalf("unexpected health status: %#v", status)
	}
	encodedStatus, err := json.Marshal(foundationx.HealthStatus{Name: "empty", Status: foundationx.HealthHealthy})
	if err != nil {
		t.Fatalf("json.Marshal(HealthStatus): %v", err)
	}
	if !strings.Contains(string(encodedStatus), `"metadata":{}`) {
		t.Fatalf("HealthStatus JSON metadata = %s, want empty object", encodedStatus)
	}
	var _ foundationx.HealthChecker = staticFoundationHealthChecker{}
	if got := (staticFoundationHealthChecker{}).Check(context.Background()); got.Status != foundationx.HealthHealthy {
		t.Fatalf("HealthChecker.Check() = %#v", got)
	}

	var _ foundationx.Lifecycle = (*foundationLifecycle)(nil)
	lifecycle := &foundationLifecycle{}
	if err := lifecycle.Start(context.Background()); err != nil {
		t.Fatalf("Start(): %v", err)
	}
	if err := lifecycle.Close(context.Background()); err != nil {
		t.Fatalf("Close(): %v", err)
	}
	if !lifecycle.started || !lifecycle.closed {
		t.Fatalf("lifecycle flags = started:%v closed:%v", lifecycle.started, lifecycle.closed)
	}

	version := foundationx.NewVersionInfo("github.com/ZoneCNH/foundationx", "v0.0.0", "abcdef0", now.Format(time.RFC3339), "go1.23")
	if version.Module != "github.com/ZoneCNH/foundationx" || version.Version != "v0.0.0" {
		t.Fatalf("unexpected version info: %#v", version)
	}
}

type staticFoundationHealthChecker struct{}

func (staticFoundationHealthChecker) Name() string { return "config" }
func (staticFoundationHealthChecker) Check(context.Context) foundationx.HealthStatus {
	return foundationx.NewHealthStatus("config", foundationx.HealthHealthy, "ok", time.Time{}, 0)
}

type foundationLifecycle struct {
	started bool
	closed  bool
}

func (l *foundationLifecycle) Start(context.Context) error {
	l.started = true
	return nil
}

func (l *foundationLifecycle) Close(context.Context) error {
	l.closed = true
	return nil
}
