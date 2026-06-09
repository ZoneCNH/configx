package configx

import (
	"context"
	"errors"
	"testing"
)

func TestErrorNilReceiverUnwrap(t *testing.T) {
	var e *Error
	if e.Unwrap() != nil {
		t.Fatal("expected nil")
	}
}

func TestErrorErrorStringAllFields(t *testing.T) {
	e := &Error{
		Kind:    ErrorKindConfig,
		Op:      "test.Op",
		Message: "something failed",
	}
	got := e.Error()
	if got != "config: test.Op: something failed" {
		t.Fatalf("error string = %q", got)
	}
}

func TestErrorErrorStringNoOp(t *testing.T) {
	e := &Error{
		Kind:    ErrorKindValidation,
		Message: "bad input",
	}
	got := e.Error()
	if got != "validation: bad input" {
		t.Fatalf("error string = %q", got)
	}
}

func TestErrorErrorStringNoMessageWithCause(t *testing.T) {
	cause := errors.New("root cause")
	e := &Error{
		Kind:  ErrorKindInternal,
		Op:    "test.Op",
		Cause: cause,
	}
	got := e.Error()
	if got != "internal: test.Op: root cause" {
		t.Fatalf("error string = %q", got)
	}
}

func TestErrorErrorStringKindOnly(t *testing.T) {
	e := &Error{Kind: ErrorKindTimeout}
	got := e.Error()
	if got != "timeout" {
		t.Fatalf("error string = %q", got)
	}
}

func TestNewError(t *testing.T) {
	e := NewError(ErrorKindAuth, "op", "msg", true)
	if e.Kind != ErrorKindAuth {
		t.Fatalf("kind = %q", e.Kind)
	}
	if !e.Retryable {
		t.Fatal("expected retryable")
	}
}

func TestWrapError(t *testing.T) {
	cause := errors.New("inner")
	e := WrapError(ErrorKindConfig, "op", "msg", false, cause)
	if e.Cause != cause {
		t.Fatal("cause not preserved")
	}
}

func TestWrapErrorEmptyMessageUsesCause(t *testing.T) {
	cause := errors.New("root cause")
	e := WrapError(ErrorKindInternal, "op", "", false, cause)
	if e.Message != "root cause" {
		t.Fatalf("message = %q, want cause", e.Message)
	}
}

func TestContextErrorDeadlineExceeded(t *testing.T) {
	e := contextError("op", context.DeadlineExceeded)
	if e.Kind != ErrorKindTimeout {
		t.Fatalf("kind = %q, want timeout", e.Kind)
	}
	if !e.Retryable {
		t.Fatal("expected retryable")
	}
}

func TestContextErrorCanceled(t *testing.T) {
	e := contextError("op", context.Canceled)
	if e.Kind != ErrorKindUnavailable {
		t.Fatalf("kind = %q, want unavailable", e.Kind)
	}
	if e.Retryable {
		t.Fatal("expected not retryable")
	}
}
