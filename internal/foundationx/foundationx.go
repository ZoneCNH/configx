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
	"strings"
	"time"
)

const redacted = "***"

// Sanitizer describes values that can expose a sanitized representation.
type Sanitizer interface {
	Sanitize() any
}

// SecretString stores a secret and masks it by default when formatted.
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

func (s SecretString) Reveal() string                { return string(s) }
func (s SecretString) Sanitize() any                { return s.String() }
func (s SecretString) IsZero() bool                 { return s == "" }
func (s SecretString) GoString() string             { return s.String() }
func (s SecretString) MarshalText() ([]byte, error) { return []byte(s.String()), nil }
func (s SecretString) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }

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
func (e *Error) WithRetryable(retryable bool) *Error {
	if e == nil {
		return nil
	}
	e.Retryable = retryable
	return e
}
func NewError(kind ErrorKind, op, message string) *Error {
	return &Error{Kind: kind, Op: op, Message: message}
}
func WrapError(kind ErrorKind, op, message string, cause error) *Error {
	return &Error{Kind: kind, Op: op, Message: message, Cause: cause}
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

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func NewRealClock() RealClock { return RealClock{} }
func (RealClock) Now() time.Time {
	return time.Now()
}

type FixedClock struct {
	now time.Time
}

func NewFixedClock(now time.Time) FixedClock { return FixedClock{now: now} }
func (c FixedClock) Now() time.Time          { return c.now }

type HealthStatusValue string

const (
	HealthHealthy   HealthStatusValue = "healthy"
	HealthDegraded  HealthStatusValue = "degraded"
	HealthUnhealthy HealthStatusValue = "unhealthy"
)

type HealthStatus struct {
	Name      string            `json:"name"`
	Status    HealthStatusValue `json:"status"`
	Message   string            `json:"message"`
	CheckedAt time.Time         `json:"checked_at"`
	LatencyMs int64             `json:"latency_ms"`
	Metadata  map[string]string `json:"metadata"`
}

type HealthChecker interface {
	Name() string
	Check(ctx context.Context) HealthStatus
}

func NewHealthStatus(name string, status HealthStatusValue, message string, checkedAt time.Time, latencyMs int64) HealthStatus {
	return HealthStatus{
		Name:      name,
		Status:    status,
		Message:   message,
		CheckedAt: checkedAt,
		LatencyMs: latencyMs,
		Metadata:  map[string]string{},
	}
}
func (s HealthStatus) WithMetadata(key, value string) HealthStatus {
	if s.Metadata == nil {
		s.Metadata = map[string]string{}
	}
	s.Metadata[key] = value
	return s
}
func (s HealthStatus) IsHealthy() bool { return s.Status == HealthHealthy }

type Starter interface {
	Start(ctx context.Context) error
}

type Closer interface {
	Close(ctx context.Context) error
}

type Lifecycle interface {
	Starter
	Closer
}

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}
}
func (p RetryPolicy) Validate() error {
	if p.MaxAttempts < 1 {
		return NewError(ErrorKindValidation, "RetryPolicy.Validate", "max attempts must be greater than zero")
	}
	if p.BaseDelay < 0 {
		return NewError(ErrorKindValidation, "RetryPolicy.Validate", "base delay must be non-negative")
	}
	if p.MaxDelay < 0 {
		return NewError(ErrorKindValidation, "RetryPolicy.Validate", "max delay must be non-negative")
	}
	if p.MaxDelay > 0 && p.BaseDelay > p.MaxDelay {
		return NewError(ErrorKindValidation, "RetryPolicy.Validate", "base delay must not exceed max delay")
	}
	return nil
}
func (p RetryPolicy) Delay(attempt int) time.Duration {
	if attempt <= 0 || p.BaseDelay <= 0 {
		return 0
	}

	delay := p.BaseDelay
	const maxDuration time.Duration = 1<<63 - 1
	for i := 1; i < attempt; i++ {
		if p.MaxDelay > 0 && delay >= p.MaxDelay {
			delay = p.MaxDelay
			break
		}
		if delay > maxDuration/2 {
			delay = maxDuration
			break
		}
		delay *= 2
	}

	if p.MaxDelay > 0 && delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	return delay
}

type VersionInfo struct {
	Module    string `json:"module"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

func NewVersionInfo(module, version, commit, buildTime, goVersion string) VersionInfo {
	return VersionInfo{
		Module:    module,
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
		GoVersion: goVersion,
	}
}

func (v VersionInfo) String() string {
	module := v.Module
	if slash := strings.LastIndex(module, "/"); slash >= 0 {
		module = module[slash+1:]
	}
	if module == "" {
		module = "unknown"
	}
	if v.Version == "" {
		return module
	}
	if v.Commit == "" {
		return fmt.Sprintf("%s %s", module, v.Version)
	}
	return fmt.Sprintf("%s %s (%s)", module, v.Version, v.Commit)
}
