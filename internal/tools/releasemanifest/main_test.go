package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildChecksUsesGlobalAndSpecificStatus(t *testing.T) {
	t.Setenv("CHECK_STATUS", "passed")
	t.Setenv("LINT_STATUS", "failed")

	checks := buildChecks()

	if checks["fmt"] != "passed" {
		t.Fatalf("fmt status = %q, want passed", checks["fmt"])
	}
	if checks["lint"] != "failed" {
		t.Fatalf("lint status = %q, want failed", checks["lint"])
	}
}

func TestValidateChecksRequiresPassedStatuses(t *testing.T) {
	checks := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		checks[name] = "passed"
	}
	checks["security"] = "unknown"

	failures := validateChecks(checks, true)

	if len(failures) != 1 {
		t.Fatalf("len(failures) = %d, want 1: %v", len(failures), failures)
	}
	if !strings.Contains(failures[0], "checks.security") {
		t.Fatalf("failure = %q, want security check failure", failures[0])
	}
}

func TestFileDigestRecordsPathAndSHA256(t *testing.T) {
	path := t.TempDir() + "/contract.json"
	if err := os.WriteFile(path, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}

	digest, err := fileDigest(path)
	if err != nil {
		t.Fatal(err)
	}

	if digest.Path != path {
		t.Fatalf("path = %q, want %q", digest.Path, path)
	}
	const want = "sha256:ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if digest.SHA256 != want {
		t.Fatalf("sha256 = %q, want %q", digest.SHA256, want)
	}
}

func TestVerifyManifestRejectsExpectedVersionMismatch(t *testing.T) {
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, repoRoot(t))

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.Version = "v1.2.3"

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "v9.9.9")
	if err == nil {
		t.Fatal("verifyManifest() succeeded, want version mismatch error")
	}
	if !strings.Contains(err.Error(), `version mismatch: got "v1.2.3", want "v9.9.9"`) {
		t.Fatalf("verifyManifest() error = %v, want version mismatch", err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			t.Fatal("could not find repo root")
		}
		dir = next
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}
