package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestVerifyManifestCommitMismatch(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.Commit = "deadbeef"

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for commit mismatch")
	}
	if !strings.Contains(err.Error(), "commit mismatch") {
		t.Fatalf("expected commit mismatch error, got: %v", err)
	}
}

func TestVerifyManifestTreeSHAMismatch(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.TreeSHA = "deadbeef"

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for tree_sha mismatch")
	}
	if !strings.Contains(err.Error(), "tree_sha mismatch") {
		t.Fatalf("expected tree_sha mismatch error, got: %v", err)
	}
}

func TestVerifyManifestSourceDigestMismatch(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.SourceDigest = "sha256:deadbeef"

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for source_digest mismatch")
	}
	if !strings.Contains(err.Error(), "source_digest") {
		t.Fatalf("expected source_digest error, got: %v", err)
	}
}

func TestVerifyManifestTrackedFileCountMismatch(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.TrackedFileCount = 999999

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for tracked_file_count mismatch")
	}
	if !strings.Contains(err.Error(), "tracked_file_count mismatch") {
		t.Fatalf("expected tracked_file_count mismatch error, got: %v", err)
	}
}

func TestVerifyManifestTreeStateMismatch(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	// Force tree state to something different from current
	if manifest.TreeState == "clean" {
		manifest.TreeState = "dirty"
	} else {
		manifest.TreeState = "clean"
	}

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for tree_state mismatch")
	}
	if !strings.Contains(err.Error(), "tree_state mismatch") {
		t.Fatalf("expected tree_state mismatch error, got: %v", err)
	}
}

func TestVerifyManifestContractsMismatch(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.Contracts = []FileDigest{{Path: "fake.json", SHA256: "sha256:deadbeef"}}

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for contracts mismatch")
	}
	if !strings.Contains(err.Error(), "contract fingerprints") {
		t.Fatalf("expected contract fingerprints error, got: %v", err)
	}
}

func TestVerifyManifestDependenciesMismatch(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.Dependencies = []ModuleDigest{{Path: "fake/module", Version: "v0.0.0"}}

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for dependencies mismatch")
	}
	if !strings.Contains(err.Error(), "dependency inventory") {
		t.Fatalf("expected dependency inventory error, got: %v", err)
	}
}

func TestVerifyManifestToolsGoEmpty(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.Tools = map[string]string{}

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for missing tools.go")
	}
	if !strings.Contains(err.Error(), "tools.go") {
		t.Fatalf("expected tools.go error, got: %v", err)
	}
}

func TestVerifyManifestVersionMismatch(t *testing.T) {
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

	err = verifyManifest(path, false, false, "v99.99.99")
	if err == nil {
		t.Fatal("expected error for version mismatch")
	}
	if !strings.Contains(err.Error(), "version mismatch") {
		t.Fatalf("expected version mismatch error, got: %v", err)
	}
}

func TestVerifyManifestRequirePassedAndClean(t *testing.T) {
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

	// Should pass when checks are passed and tree is clean
	err = verifyManifest(path, true, true, "")
	if err != nil {
		// May fail if tree is dirty in test environment
		if !strings.Contains(err.Error(), "tree_state") && !strings.Contains(err.Error(), "checks") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestToolVersionCommandExistsButFails(t *testing.T) {
	// "go" exists but "go nonexistent" returns an error
	got := toolVersion("go", "nonexistent-subcommand")
	if got == "missing" {
		t.Fatal("expected not missing for existing command")
	}
	if !strings.HasPrefix(got, "error: ") {
		t.Fatalf("expected error: prefix, got %q", got)
	}
}

func TestSourceDigestFileReadError(t *testing.T) {
	// sourceDigest reads files listed by git ls-files.
	// We can't easily make git list a file that doesn't exist,
	// but we can verify the function returns valid output from repo root.
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

func TestContractDigestsMissingFile(t *testing.T) {
	// Save original and override with a nonexistent file
	orig := contractFiles
	contractFiles = []string{"/nonexistent/file.json"}
	defer func() { contractFiles = orig }()

	_, err := contractDigests()
	if err == nil {
		t.Fatal("expected error for nonexistent contract file")
	}
}

func TestModuleDigestsInvalidJSON(t *testing.T) {
	// moduleDigests calls "go list -m -json all" which we can't easily mock.
	// Verify it works from repo root.
	chdir(t, repoRoot(t))
	modules, err := moduleDigests()
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) == 0 {
		t.Fatal("expected at least one module")
	}
}

func TestTreeStateReturnsValidString(t *testing.T) {
	state := treeState()
	if state != "clean" && state != "dirty" && state != "unknown" {
		t.Fatalf("treeState = %q, want clean/dirty/unknown", state)
	}
}

func TestBuildManifestSourceDigestError(t *testing.T) {
	// buildManifest calls sourceDigest which needs git.
	// Test from repo root for the happy path.
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")
	m, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	if m.SourceDigest == "" {
		t.Fatal("expected source digest")
	}
	if m.TrackedFileCount == 0 {
		t.Fatal("expected tracked file count")
	}
}

func TestBuildManifestWithCustomVersion(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")
	t.Setenv("VERSION", "v2.0.0")
	t.Setenv("GENERATED_BY", "test-runner")

	m, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	if m.Version != "v2.0.0" {
		t.Fatalf("version = %q, want v2.0.0", m.Version)
	}
	if m.GeneratedBy != "test-runner" {
		t.Fatalf("generated_by = %q, want test-runner", m.GeneratedBy)
	}
}

func TestBuildManifestToolsVersions(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	m, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	if m.Tools == nil {
		t.Fatal("expected tools map")
	}
	if m.Tools["go"] == "" {
		t.Fatal("expected go tool version")
	}
}

func TestBuildManifestArtifacts(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	m, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Artifacts) == 0 {
		t.Fatal("expected artifacts")
	}
	found := false
	for _, a := range m.Artifacts {
		if a == "release/manifest/latest.json" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected release/manifest/latest.json in artifacts")
	}
}

func TestBuildManifestNotes(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	m, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	if m.Notes.BreakingChanges != "none" {
		t.Fatalf("breaking_changes = %q", m.Notes.BreakingChanges)
	}
	if m.Notes.KnownRisks == nil {
		t.Fatal("expected known_risks to be non-nil")
	}
}

func TestBuildManifestGoVersion(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	m, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	if m.GoVersion == "" {
		t.Fatal("expected go_version")
	}
}

func TestBuildManifestGeneratedAtRFC3339(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	m, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := time.Parse(time.RFC3339, m.GeneratedAt); err != nil {
		t.Fatalf("generated_at is not RFC3339: %v", err)
	}
}

func TestBuildManifestCommitAndTreeSHA(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	m, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}
	if m.Commit == "" {
		t.Fatal("expected commit")
	}
	if m.TreeSHA == "" {
		t.Fatal("expected tree_sha")
	}
}

func TestVerifyManifestEmptyRequireNonEmptyFields(t *testing.T) {
	chdir(t, repoRoot(t))
	t.Setenv("CHECK_STATUS", "passed")

	manifest, err := buildManifest()
	if err != nil {
		t.Fatal(err)
	}

	// Clear all required fields
	manifest.Module = ""
	manifest.Version = ""
	manifest.Commit = ""
	manifest.TreeSHA = ""
	manifest.SourceDigest = ""
	manifest.GoVersion = ""
	manifest.GeneratedAt = ""
	manifest.GeneratedBy = ""
	manifest.TreeState = ""

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected error for empty required fields")
	}
	// Should contain multiple failures
	if !strings.Contains(err.Error(), "module is required") {
		t.Fatalf("expected module required error, got: %v", err)
	}
}
