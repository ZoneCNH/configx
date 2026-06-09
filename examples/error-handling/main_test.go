package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestMainOutput(t *testing.T) {
	output := captureStdout(t, main)

	// Verify each section header appears.
	for _, want := range []string{
		"=== Missing Config File ===",
		"=== Invalid Format ===",
		"=== Merge Priority ===",
		"=== Secret Redaction ===",
		"=== Validation Error ===",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing section %q", want)
		}
	}

	// Verify missing file error reports kind=config and retryable=false.
	if !strings.Contains(output, "kind=config") {
		t.Errorf("expected 'kind=config' in output for missing file error, got:\n%s", output)
	}
	if !strings.Contains(output, "retryable=false") {
		t.Errorf("expected 'retryable=false' for missing file error, got:\n%s", output)
	}

	// Verify raw secret appears but sanitized does not.
	if !strings.Contains(output, "super-secret-123") {
		t.Errorf("expected raw secret in output, got:\n%s", output)
	}
	if strings.Contains(output, "sanitized: \"super-secret-123\"") {
		t.Errorf("sanitized output should redact the secret, got:\n%s", output)
	}

	// Verify validation error.
	if !strings.Contains(output, "caught validation error as expected") {
		t.Errorf("expected validation error message in output, got:\n%s", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = original
	})

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = original

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return buf.String()
}
