package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRequireNonEmptyEmptyValue(t *testing.T) {
	var failures []string
	requireNonEmpty(&failures, "field", "")
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(failures))
	}
	if !strings.Contains(failures[0], "field is required") {
		t.Fatalf("failure = %q", failures[0])
	}
}

func TestRequireNonEmptyWhitespaceValue(t *testing.T) {
	var failures []string
	requireNonEmpty(&failures, "field", "   ")
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure for whitespace, got %d", len(failures))
	}
}

func TestRequireNonEmptyNonEmptyValue(t *testing.T) {
	var failures []string
	requireNonEmpty(&failures, "field", "value")
	if len(failures) != 0 {
		t.Fatalf("expected 0 failures, got %d", len(failures))
	}
}

func TestContainsFound(t *testing.T) {
	if !contains([]string{"a", "b", "c"}, "b") {
		t.Fatal("expected contains to find 'b'")
	}
}

func TestContainsNotFound(t *testing.T) {
	if contains([]string{"a", "b", "c"}, "d") {
		t.Fatal("expected contains to not find 'd'")
	}
}

func TestContainsEmptySlice(t *testing.T) {
	if contains(nil, "a") {
		t.Fatal("expected contains to not find 'a' in nil slice")
	}
	if contains([]string{}, "a") {
		t.Fatal("expected contains to not find 'a' in empty slice")
	}
}

func TestFirstLineSingleLine(t *testing.T) {
	got := firstLine("hello world")
	if got != "hello world" {
		t.Fatalf("firstLine = %q", got)
	}
}

func TestFirstLineMultiLine(t *testing.T) {
	got := firstLine("first\nsecond\nthird")
	if got != "first" {
		t.Fatalf("firstLine = %q", got)
	}
}

func TestFirstLineWithWhitespace(t *testing.T) {
	got := firstLine("  hello  \n  world  ")
	// firstLine trims outer whitespace, then takes first line
	// "  hello  \n  world  " -> trim -> "hello  \n  world  " -> first line -> "hello  "
	if got != "hello  " {
		t.Fatalf("firstLine = %q", got)
	}
}

func TestEnvDefaultWithEnv(t *testing.T) {
	t.Setenv("TEST_ENV_DEFAULT", "set")
	got := envDefault("TEST_ENV_DEFAULT", "fallback")
	if got != "set" {
		t.Fatalf("envDefault = %q, want set", got)
	}
}

func TestEnvDefaultWithoutEnv(t *testing.T) {
	t.Setenv("TEST_ENV_DEFAULT_MISSING", "")
	got := envDefault("TEST_ENV_DEFAULT_MISSING", "fallback")
	if got != "fallback" {
		t.Fatalf("envDefault = %q, want fallback", got)
	}
}

func TestEnvDefaultWhitespaceOnly(t *testing.T) {
	t.Setenv("TEST_ENV_WS", "   ")
	got := envDefault("TEST_ENV_WS", "fallback")
	if got != "fallback" {
		t.Fatalf("envDefault = %q, want fallback", got)
	}
}

func TestRunTrimmedDefault(t *testing.T) {
	got := runTrimmedDefault("fallback", "echo", "hello")
	if got != "hello" {
		t.Fatalf("runTrimmedDefault = %q", got)
	}
}

func TestRunTrimmedDefaultFails(t *testing.T) {
	got := runTrimmedDefault("fallback", "/nonexistent/command")
	if got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestRunTrimmed(t *testing.T) {
	got, err := runTrimmed("echo", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("runTrimmed = %q", got)
	}
}

func TestRunTrimmedFails(t *testing.T) {
	_, err := runTrimmed("/nonexistent/command")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunRaw(t *testing.T) {
	got, err := runRaw("echo", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "hello") {
		t.Fatalf("runRaw = %q", got)
	}
}

func TestRunRawFails(t *testing.T) {
	_, err := runRaw("/nonexistent/command")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestToolVersionMissing(t *testing.T) {
	got := toolVersion("/nonexistent-tool", "--version")
	if got != "missing" {
		t.Fatalf("toolVersion = %q, want missing", got)
	}
}

func TestToolVersionExists(t *testing.T) {
	got := toolVersion("go", "version")
	if got == "" || got == "missing" {
		t.Fatalf("toolVersion(go) = %q", got)
	}
}

func TestTreeStateClean(t *testing.T) {
	// This will be "dirty" in the test environment, but let's verify it returns a valid string.
	state := treeState()
	if state != "clean" && state != "dirty" && state != "unknown" {
		t.Fatalf("treeState = %q, want clean/dirty/unknown", state)
	}
}

func TestValidateChecksMissingField(t *testing.T) {
	checks := map[string]string{
		"fmt": "passed",
		// Missing other required checks
	}
	failures := validateChecks(checks, false)
	if len(failures) == 0 {
		t.Fatal("expected failures for missing checks")
	}
}

func TestValidateChecksNotRequirePassed(t *testing.T) {
	checks := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		checks[name] = "unknown"
	}
	failures := validateChecks(checks, false)
	if len(failures) != 0 {
		t.Fatalf("expected 0 failures when not requiring passed, got %d", len(failures))
	}
}

func TestValidateChecksAllPassed(t *testing.T) {
	checks := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		checks[name] = "passed"
	}
	failures := validateChecks(checks, true)
	if len(failures) != 0 {
		t.Fatalf("expected 0 failures when all passed, got %d", len(failures))
	}
}

func TestWriteManifest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "manifest.json")
	m := Manifest{
		Module:  "test",
		Version: "v0.1.0",
	}
	if err := writeManifest(path, m); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "test") {
		t.Fatal("expected module name in output")
	}
}

func TestWriteManifestInvalidPath(t *testing.T) {
	// Try to write to a path where the parent is a file, not a dir
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(blocker, "sub", "manifest.json")
	err := writeManifest(path, Manifest{})
	if err == nil {
		t.Fatal("expected error writing to invalid path")
	}
}

func TestFileDigestNonexistentFile(t *testing.T) {
	_, err := fileDigest("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestContractDigestsFromRepoRoot(t *testing.T) {
	chdir(t, repoRoot(t))
	digests, err := contractDigests()
	if err != nil {
		t.Fatal(err)
	}
	if len(digests) == 0 {
		t.Fatal("expected at least one contract digest")
	}
	// Each digest should have a non-empty path and sha256
	for _, d := range digests {
		if d.Path == "" {
			t.Fatal("expected non-empty path")
		}
		if !strings.HasPrefix(d.SHA256, "sha256:") {
			t.Fatalf("expected sha256 prefix, got %q", d.SHA256)
		}
	}
}

func TestSourceDigestFromRepoRoot(t *testing.T) {
	chdir(t, repoRoot(t))
	digest, count, err := sourceDigest()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("expected sha256 prefix, got %q", digest)
	}
	if count == 0 {
		t.Fatal("expected at least one tracked file")
	}
}

func TestModuleDigestsFromRepoRoot(t *testing.T) {
	chdir(t, repoRoot(t))
	modules, err := moduleDigests()
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) == 0 {
		t.Fatal("expected at least one module")
	}
	// At least one module should be Main=true
	foundMain := false
	for _, m := range modules {
		if m.Main {
			foundMain = true
			break
		}
	}
	if !foundMain {
		t.Fatal("expected at least one main module")
	}
}

func TestBuildManifestFromRepoRoot(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")
	m, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	if m.Module == "" {
		t.Fatal("expected non-empty module")
	}
	if m.Version == "" {
		t.Fatal("expected non-empty version")
	}
	if m.Commit == "" {
		t.Fatal("expected non-empty commit")
	}
	if m.TreeSHA == "" {
		t.Fatal("expected non-empty tree sha")
	}
	if m.SourceDigest == "" {
		t.Fatal("expected non-empty source digest")
	}
	if m.GoVersion == "" {
		t.Fatal("expected non-empty go version")
	}
	if m.GeneratedAt == "" {
		t.Fatal("expected non-empty generated at")
	}
	if m.GeneratedBy == "" {
		t.Fatal("expected non-empty generated by")
	}
	if m.TreeState == "" {
		t.Fatal("expected non-empty tree state")
	}
	if len(m.Contracts) == 0 {
		t.Fatal("expected contracts")
	}
	if len(m.Dependencies) == 0 {
		t.Fatal("expected dependencies")
	}
	if m.Notes.BreakingChanges != "none" {
		t.Fatalf("breaking changes = %q", m.Notes.BreakingChanges)
	}
}

func TestBuildChecksFromRepoRoot(t *testing.T) {
	t.Setenv("CHECK_STATUS", "unknown")
	t.Setenv("FMT_STATUS", "passed")
	t.Setenv("VET_STATUS", "passed")
	checks := buildChecks()
	if checks["fmt"] != "passed" {
		t.Fatalf("fmt = %q", checks["fmt"])
	}
	if checks["vet"] != "passed" {
		t.Fatalf("vet = %q", checks["vet"])
	}
	if checks["lint"] != "unknown" {
		t.Fatalf("lint = %q", checks["lint"])
	}
}

func TestVerifyManifestFromRepoRoot(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	// Verify should pass with no flags
	if err := verifyManifest(path, false, false, ""); err != nil {
		t.Fatalf("verifyManifest failed: %v", err)
	}
}

func TestVerifyManifestMissingFile(t *testing.T) {
	err := verifyManifest("/nonexistent/manifest.json", false, false, "")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestVerifyManifestInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestVerifyManifestRequirePassed(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "unknown")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	// Verify with requirePassed=true should fail because checks are "unknown"
	err = verifyManifest(path, true, false, "")
	if err == nil {
		t.Fatal("expected error with requirePassed=true and unknown checks")
	}
}

func TestVerifyManifestRequireCleanDirty(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	// Force tree state to dirty
	manifest.TreeState = "dirty"

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, true, "")
	if err == nil {
		t.Fatal("expected error with requireClean=true and dirty tree")
	}
}

func TestVerifyManifestMissingArtifact(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	// Remove the required artifact
	manifest.Artifacts = []string{}

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for missing artifact")
	}
}

func TestVerifyManifestEmptyModule(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	manifest.Module = ""

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for empty module")
	}
}

func TestVerifyManifestInvalidGeneratedAt(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	manifest.GeneratedAt = "not-rfc3339"

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for invalid generated_at")
	}
}

func TestVerifyManifestEmptyGoVersion(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	manifest.Tools = map[string]string{"go": ""}

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for empty tools.go")
	}
}
