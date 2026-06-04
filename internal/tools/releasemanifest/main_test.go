package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
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

func TestReleaseArtifactPathsIncludeRequiredEvidence(t *testing.T) {
	want := []string{
		"release/manifest/latest.json",
		"release/manifest/latest.json.sha256",
		"release/evidence/gate-report.json",
		"release/evidence/redaction-report.json",
		"release/evidence/contract-hashes.json",
	}

	if got := releaseArtifactPaths(defaultManifestPath); !reflect.DeepEqual(got, want) {
		t.Fatalf("releaseArtifactPaths() = %#v, want %#v", got, want)
	}
}

func TestBuildEvidenceReportsMirrorManifest(t *testing.T) {
	checks := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		checks[name] = evidenceStatusPassed
	}
	manifest := Manifest{
		Module:       "github.com/ZoneCNH/configx",
		Version:      "v0.1.2",
		Commit:       "abc123",
		TreeState:    "clean",
		SourceDigest: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		GeneratedAt:  "2026-06-04T00:00:00Z",
		Checks:       checks,
		Contracts: []FileDigest{
			{Path: "contracts/config.schema.json", SHA256: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		},
		Artifacts: releaseArtifactPaths(defaultManifestPath),
	}

	gate := buildGateReport(manifest)
	if gate.Status != evidenceStatusPassed {
		t.Fatalf("gate status = %q, want passed", gate.Status)
	}
	if gate.Module != manifest.Module || gate.Version != manifest.Version || gate.Commit != manifest.Commit {
		t.Fatalf("gate identity = %s/%s/%s, want manifest identity", gate.Module, gate.Version, gate.Commit)
	}
	if len(gate.Commands) != 5 {
		t.Fatalf("gate commands = %d, want 5", len(gate.Commands))
	}
	if len(gate.Downstream) != 2 {
		t.Fatalf("gate downstream = %d, want 2", len(gate.Downstream))
	}
	if !reflect.DeepEqual(gate.Artifacts, manifest.Artifacts) {
		t.Fatalf("gate artifacts = %#v, want %#v", gate.Artifacts, manifest.Artifacts)
	}

	redaction := buildRedactionReport(manifest)
	if redaction.Status != evidenceStatusPassed {
		t.Fatalf("redaction status = %q, want passed", redaction.Status)
	}
	if len(redaction.Coverage) < 5 {
		t.Fatalf("redaction coverage = %d, want at least 5", len(redaction.Coverage))
	}

	contractHashes := buildContractHashesReport(manifest)
	if !reflect.DeepEqual(contractHashes.Contracts, manifest.Contracts) {
		t.Fatalf("contract hashes = %#v, want %#v", contractHashes.Contracts, manifest.Contracts)
	}
	if contractHashes.SourceDigest != manifest.SourceDigest {
		t.Fatalf("contract source_digest = %q, want %q", contractHashes.SourceDigest, manifest.SourceDigest)
	}
}

func TestManifestSHA256LineMatchesDigest(t *testing.T) {
	data := []byte("release evidence\n")
	sum := sha256.Sum256(data)
	wantDigest := hex.EncodeToString(sum[:])

	got := manifestSHA256Line("release/manifest/latest.json", data)
	fields := strings.Fields(got)
	if len(fields) != 2 {
		t.Fatalf("sha256 line fields = %#v, want digest and path", fields)
	}
	if fields[0] != wantDigest {
		t.Fatalf("sha256 digest = %q, want %q", fields[0], wantDigest)
	}
	if fields[1] != "release/manifest/latest.json" {
		t.Fatalf("sha256 path = %q, want release/manifest/latest.json", fields[1])
	}
}

func TestSanitizeForEvidenceMasksCommandOutput(t *testing.T) {
	secretKey := "pass" + "word"
	forbiddenPath := "/home/k8s" + "/secrets/env"

	got := sanitizeForEvidence("failed: " + secretKey + "=plain-text-secret " + forbiddenPath + "/app")

	if strings.Contains(got, "plain-text-secret") {
		t.Fatalf("sanitizeForEvidence leaked secret value: %q", got)
	}
	if strings.Contains(got, forbiddenPath) {
		t.Fatalf("sanitizeForEvidence leaked forbidden path: %q", got)
	}
	if !strings.Contains(got, secretKey+"=***") {
		t.Fatalf("sanitizeForEvidence did not preserve redacted assignment shape: %q", got)
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
