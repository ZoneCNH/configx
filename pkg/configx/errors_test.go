package configx

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewErrorFormatsKindOpAndMessage(t *testing.T) {
	err := NewError(ErrorKindValidation, "configx.Test", "bad input", false)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Kind != ErrorKindValidation {
		t.Fatalf("expected validation kind, got %q", err.Kind)
	}
	if err.Retryable {
		t.Fatal("expected non-retryable error")
	}
	if got := err.Error(); !strings.Contains(got, "validation: configx.Test: bad input") {
		t.Fatalf("unexpected error string: %q", got)
	}
}

func TestWrapErrorPreservesCauseAndKind(t *testing.T) {
	cause := context.DeadlineExceeded
	err := WrapError(ErrorKindTimeout, "configx.Test", "", true, cause)

	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout kind, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped cause, got %v", err)
	}
	if !err.Retryable {
		t.Fatal("expected retryable error")
	}
}

func TestContextErrorClassifiesDeadlineAsRetryableTimeout(t *testing.T) {
	err := contextError("configx.Test", context.DeadlineExceeded)
	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout kind, got %v", err)
	}
	if !err.Retryable {
		t.Fatal("expected deadline errors to be retryable")
	}
}
