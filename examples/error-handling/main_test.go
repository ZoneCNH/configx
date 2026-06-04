package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestExampleRuns(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("example panicked or failed: %v\noutput:\n%s", err, out)
	}

	output := string(out)

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
}

func TestMissingFileError(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("example failed: %v\noutput:\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "kind=config") {
		t.Errorf("expected 'kind=config' in output for missing file error, got:\n%s", output)
	}
	if !strings.Contains(output, "retryable=false") {
		t.Errorf("expected 'retryable=false' for missing file error, got:\n%s", output)
	}
}

func TestSecretRedaction(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("example failed: %v\noutput:\n%s", err, out)
	}

	output := string(out)
	// Raw value should appear.
	if !strings.Contains(output, "super-secret-123") {
		t.Errorf("expected raw secret in output, got:\n%s", output)
	}
	// Sanitized value should NOT contain the secret.
	if strings.Contains(output, "sanitized: \"super-secret-123\"") {
		t.Errorf("sanitized output should redact the secret, got:\n%s", output)
	}
}

func TestValidationError(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("example failed: %v\noutput:\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "caught validation error as expected") {
		t.Errorf("expected validation error message in output, got:\n%s", output)
	}
}
