// Package foundationx provides the small compatibility surface configx depends on.
//
// It intentionally stays local to this repository so configx can keep an
// explicit, dependency-light boundary while preserving the public helpers callers
// expect from foundationx.
package foundationx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const redacted = "***"

// Clock abstracts time reads for deterministic tests.
type Clock interface {
	Now() time.Time
}

// RealClock reads the current wall clock time.
type RealClock struct{}

// Now returns the current time.
func (RealClock) Now() time.Time {
	return time.Now()
}

// FixedClock always returns the same instant.
type FixedClock struct {
	now time.Time
}

// NewFixedClock creates a clock pinned to now.
func NewFixedClock(now time.Time) FixedClock {
	return FixedClock{now: now}
}

// Now returns the fixed instant.
func (c FixedClock) Now() time.Time {
	return c.now
}

// Sanitizer marks values that can return a safe representation of themselves.
type Sanitizer interface {
	Sanitize() any
}

// SecretString stores secret material while redacting display and marshal paths.
type SecretString string

func NewSecretString(value string) SecretString {
	return SecretString(value)
}

func (s SecretString) String() string {
	if s == "" {
		return ""
	}
	return redacted
}

func (s SecretString) GoString() string {
	return s.String()
}

func (s SecretString) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s SecretString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// Sanitize returns a redacted value suitable for logs and generated evidence.
func (s SecretString) Sanitize() any {
	return s.String()
}

// Reveal returns the raw value for the final integration boundary that needs it.
func (s SecretString) Reveal() string {
	return string(s)
}

// IsZero reports whether the secret is unset.
func (s SecretString) IsZero() bool {
	return s == ""
}

type ErrorKind string

const (
	ErrorKindConfig       ErrorKind = "config"
	ErrorKindValidation   ErrorKind = "validation"
	ErrorKindConnection   ErrorKind = "connection"
	ErrorKindUnavailable  ErrorKind = "unavailable"
	ErrorKindTimeout      ErrorKind = "timeout"
	ErrorKindAuth         ErrorKind = "auth"
	ErrorKindConflict     ErrorKind = "conflict"
	ErrorKindRateLimit    ErrorKind = "rate_limit"
	ErrorKindCanceled     ErrorKind = "canceled"
	ErrorKindNotFound     ErrorKind = "not_found"
	ErrorKindAlreadyExist ErrorKind = "already_exists"
	ErrorKindInternal     ErrorKind = "internal"
)

// Error is the normalized foundation error shape.
type Error struct {
	Kind      ErrorKind `json:"kind"`
	Op        string    `json:"op,omitempty"`
	Message   string    `json:"message"`
	Cause     error     `json:"-"`
	Retryable bool      `json:"retryable"`
}

func NewError(kind ErrorKind, op string, message string) *Error {
	return newError(kind, op, message, nil)
}

func WrapError(kind ErrorKind, op string, message string, cause error) *Error {
	return newError(kind, op, message, cause)
}

func (e *Error) WithRetryable(retryable bool) *Error {
	if e == nil {
		return nil
	}
	e.Retryable = retryable
	return e
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	message := string(e.Kind)
	if e.Op != "" {
		message += ": " + e.Op
	}
	if e.Message != "" {
		message += ": " + e.Message
	}
	if e.Message == "" && e.Cause != nil {
		message += ": " + e.Cause.Error()
	}
	return message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func IsKind(err error, kind ErrorKind) bool {
	var target *Error
	if errors.As(err, &target) {
		return target.Kind == kind
	}
	return false
}

func AsFoundationError(err error) (*Error, bool) {
	var target *Error
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

func newError(kind ErrorKind, op string, message string, cause error) *Error {
	if message == "" && cause != nil {
		message = cause.Error()
	}
	return &Error{Kind: kind, Op: op, Message: message, Cause: cause}
}

type HealthStatusValue string

const (
	HealthHealthy   HealthStatusValue = "healthy"
	HealthDegraded  HealthStatusValue = "degraded"
	HealthUnhealthy HealthStatusValue = "unhealthy"
)

// HealthStatus is the standard health payload shared by base modules.
type HealthStatus struct {
	Name      string            `json:"name"`
	Status    HealthStatusValue `json:"status"`
	Message   string            `json:"message"`
	CheckedAt time.Time         `json:"checked_at"`
	LatencyMs int64             `json:"latency_ms"`
	Metadata  map[string]string `json:"metadata"`
}

func NewHealthStatus(name string, status HealthStatusValue, message string, checkedAt time.Time, latencyMs int64) HealthStatus {
	if latencyMs < 0 {
		latencyMs = 0
	}
	return HealthStatus{
		Name:      name,
		Status:    status,
		Message:   message,
		CheckedAt: checkedAt,
		LatencyMs: latencyMs,
		Metadata:  map[string]string{},
	}
}

func (s HealthStatus) WithMetadata(key string, value string) HealthStatus {
	if s.Metadata == nil {
		s.Metadata = map[string]string{}
	}
	s.Metadata[key] = value
	return s
}

func (s HealthStatus) IsHealthy() bool {
	return s.Status == HealthHealthy
}

// HealthChecker provides a common health-check interface.
type HealthChecker interface {
	HealthCheck(context.Context) HealthStatus
}

// Starter owns startup lifecycle hooks.
type Starter interface {
	Start(context.Context) error
}

// Closer owns shutdown lifecycle hooks.
type Closer interface {
	Close(context.Context) error
}

// Lifecycle combines startup and shutdown hooks.
type Lifecycle interface {
	Starter
	Closer
}

// RetryPolicy describes bounded exponential backoff.
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	BaseDelay   time.Duration `json:"base_delay"`
	MaxDelay    time.Duration `json:"max_delay"`
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond, MaxDelay: time.Second}
}

func (p RetryPolicy) Validate() error {
	if p.MaxAttempts < 1 {
		return NewError(ErrorKindValidation, "foundationx.RetryPolicy.Validate", "max attempts must be at least 1")
	}
	if p.BaseDelay <= 0 {
		return NewError(ErrorKindValidation, "foundationx.RetryPolicy.Validate", "base delay must be positive")
	}
	if p.MaxDelay <= 0 {
		return NewError(ErrorKindValidation, "foundationx.RetryPolicy.Validate", "max delay must be positive")
	}
	if p.MaxDelay < p.BaseDelay {
		return NewError(ErrorKindValidation, "foundationx.RetryPolicy.Validate", "max delay must be greater than or equal to base delay")
	}
	return nil
}

func (p RetryPolicy) Delay(attempt int) time.Duration {
	if attempt <= 1 {
		return p.BaseDelay
	}
	delay := p.BaseDelay
	for i := 1; i < attempt; i++ {
		if delay >= p.MaxDelay/2 {
			delay = p.MaxDelay
			break
		}
		delay *= 2
	}
	if delay > p.MaxDelay {
		return p.MaxDelay
	}
	return delay
}

// VersionInfo is the generated build/version evidence shape.
type VersionInfo struct {
	Module    string `json:"module"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

func NewVersionInfo(module string, version string, commit string, buildTime string, goVersion string) VersionInfo {
	return VersionInfo{Module: module, Version: version, Commit: commit, BuildTime: buildTime, GoVersion: goVersion}
}

func (v VersionInfo) String() string {
	return fmt.Sprintf("%s %s (%s)", v.Module, v.Version, v.Commit)
}
