package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
)

var checkNames = []string{
	"fmt",
	"vet",
	"lint",
	"unit_test",
	"race_test",
	"boundary",
	"secret_scan",
	"security",
	"contract",
	"integration",
}

var checkEnvNames = map[string]string{
	"fmt":         "FMT_STATUS",
	"vet":         "VET_STATUS",
	"lint":        "LINT_STATUS",
	"unit_test":   "UNIT_TEST_STATUS",
	"race_test":   "RACE_TEST_STATUS",
	"boundary":    "BOUNDARY_STATUS",
	"secret_scan": "SECRET_SCAN_STATUS",
	"security":    "SECURITY_STATUS",
	"contract":    "CONTRACT_STATUS",
	"integration": "INTEGRATION_STATUS",
}

var contractFiles = []string{
	"contracts/config.schema.json",
	"contracts/error.schema.json",
	"contracts/health.schema.json",
	"contracts/metrics.md",
}

type Manifest struct {
	Module           string            `json:"module"`
	Version          string            `json:"version"`
	Commit           string            `json:"commit"`
	TreeSHA          string            `json:"tree_sha"`
	SourceDigest     string            `json:"source_digest"`
	TrackedFileCount int               `json:"tracked_file_count"`
	GoVersion        string            `json:"go_version"`
	GeneratedAt      string            `json:"generated_at"`
	GeneratedBy      string            `json:"generated_by"`
	TreeState        string            `json:"tree_state"`
	Checks           map[string]string `json:"checks"`
	Contracts        []FileDigest      `json:"contracts"`
	Dependencies     []ModuleDigest    `json:"dependencies"`
	Tools            map[string]string `json:"tools"`
	Artifacts        []string          `json:"artifacts"`
	Notes            Notes             `json:"notes"`
}

type FileDigest struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type ModuleDigest struct {
	Path    string         `json:"path"`
	Version string         `json:"version,omitempty"`
	Main    bool           `json:"main,omitempty"`
	Replace *ModuleReplace `json:"replace,omitempty"`
}

type ModuleReplace struct {
	Path    string `json:"path"`
	Version string `json:"version,omitempty"`
}

type Notes struct {
	BreakingChanges string   `json:"breaking_changes"`
	KnownRisks      []string `json:"known_risks"`
}

func main() {
	out := flag.String("out", "release/manifest/latest.json", "release manifest output path")
	verify := flag.String("verify", "", "verify an existing release manifest instead of generating one")
	requirePassed := flag.Bool("require-passed", false, "require all release checks to be passed during verification")
	requireClean := flag.Bool("require-clean", false, "require a clean git tree during verification")
	flag.Parse()

	if *verify != "" {
		if err := verifyManifest(*verify, *requirePassed, *requireClean); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("release evidence verified: %s\n", *verify)
		return
	}

	manifest, err := buildManifest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	if err := writeManifest(*out, manifest); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("generated %s\n", *out)
}

func buildManifest() (Manifest, error) {
	module, err := runTrimmed("go", "list", "-m")
	if err != nil {
		return Manifest{}, err
	}

	sourceDigest, trackedFileCount, err := sourceDigest()
	if err != nil {
		return Manifest{}, err
	}
	contracts, err := contractDigests()
	if err != nil {
		return Manifest{}, err
	}
	dependencies, err := moduleDigests()
	if err != nil {
		return Manifest{}, err
	}

	return Manifest{
		Module:           module,
		Version:          envDefault("VERSION", "v0.1.0"),
		Commit:           runTrimmedDefault("unknown", "git", "rev-parse", "HEAD"),
		TreeSHA:          runTrimmedDefault("unknown", "git", "rev-parse", "HEAD^{tree}"),
		SourceDigest:     sourceDigest,
		TrackedFileCount: trackedFileCount,
		GoVersion:        runtime.Version(),
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		GeneratedBy:      envDefault("GENERATED_BY", "scripts/generate_manifest.sh"),
		TreeState:        treeState(),
		Checks:           buildChecks(),
		Contracts:        contracts,
		Dependencies:     dependencies,
		Tools: map[string]string{
			"go":            firstLine(runTrimmedDefault(runtime.Version(), "go", "version")),
			"golangci-lint": toolVersion("golangci-lint", "--version"),
			"govulncheck":   toolVersion("govulncheck", "-version"),
		},
		Artifacts: []string{
			"release/manifest/latest.json",
		},
		Notes: Notes{
			BreakingChanges: "none",
			KnownRisks:      []string{},
		},
	}, nil
}

func verifyManifest(path string, requirePassed bool, requireClean bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var got Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		return err
	}

	current, err := buildManifest()
	if err != nil {
		return err
	}

	var failures []string
	requireNonEmpty(&failures, "module", got.Module)
	requireNonEmpty(&failures, "version", got.Version)
	requireNonEmpty(&failures, "commit", got.Commit)
	requireNonEmpty(&failures, "tree_sha", got.TreeSHA)
	requireNonEmpty(&failures, "source_digest", got.SourceDigest)
	requireNonEmpty(&failures, "go_version", got.GoVersion)
	requireNonEmpty(&failures, "generated_at", got.GeneratedAt)
	requireNonEmpty(&failures, "generated_by", got.GeneratedBy)
	requireNonEmpty(&failures, "tree_state", got.TreeState)

	if _, err := time.Parse(time.RFC3339, got.GeneratedAt); err != nil {
		failures = append(failures, "generated_at must be RFC3339")
	}
	if got.Module != current.Module {
		failures = append(failures, fmt.Sprintf("module mismatch: got %q, want %q", got.Module, current.Module))
	}
	if got.Commit != current.Commit {
		failures = append(failures, fmt.Sprintf("commit mismatch: got %q, want %q", got.Commit, current.Commit))
	}
	if got.TreeSHA != current.TreeSHA {
		failures = append(failures, fmt.Sprintf("tree_sha mismatch: got %q, want %q", got.TreeSHA, current.TreeSHA))
	}
	if got.SourceDigest != current.SourceDigest {
		failures = append(failures, "source_digest does not match current tracked file contents")
	}
	if got.TrackedFileCount != current.TrackedFileCount {
		failures = append(failures, fmt.Sprintf("tracked_file_count mismatch: got %d, want %d", got.TrackedFileCount, current.TrackedFileCount))
	}
	if got.TreeState != current.TreeState {
		failures = append(failures, fmt.Sprintf("tree_state mismatch: got %q, want %q", got.TreeState, current.TreeState))
	}
	if requireClean && got.TreeState != "clean" {
		failures = append(failures, fmt.Sprintf("tree_state must be clean, got %q", got.TreeState))
	}
	if !reflect.DeepEqual(got.Contracts, current.Contracts) {
		failures = append(failures, "contract fingerprints do not match current contract files")
	}
	if !reflect.DeepEqual(got.Dependencies, current.Dependencies) {
		failures = append(failures, "dependency inventory does not match go list -m -json all")
	}
	if !contains(got.Artifacts, "release/manifest/latest.json") {
		failures = append(failures, "artifacts must include release/manifest/latest.json")
	}
	if got.Tools["go"] == "" {
		failures = append(failures, "tools.go must be recorded")
	}
	failures = append(failures, validateChecks(got.Checks, requirePassed)...)

	if len(failures) > 0 {
		return errors.New("release evidence verification failed:\n - " + strings.Join(failures, "\n - "))
	}
	return nil
}

func writeManifest(path string, manifest Manifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func buildChecks() map[string]string {
	defaultStatus := envDefault("CHECK_STATUS", "unknown")
	checks := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		checks[name] = envDefault(checkEnvNames[name], defaultStatus)
	}
	return checks
}

func validateChecks(checks map[string]string, requirePassed bool) []string {
	var failures []string
	for _, name := range checkNames {
		status := strings.TrimSpace(checks[name])
		if status == "" {
			failures = append(failures, "checks."+name+" is required")
			continue
		}
		if requirePassed && status != "passed" {
			failures = append(failures, fmt.Sprintf("checks.%s must be passed, got %q", name, status))
		}
	}
	return failures
}

func sourceDigest() (string, int, error) {
	raw, err := runRaw("git", "ls-files", "-z")
	if err != nil {
		return "", 0, err
	}
	parts := strings.Split(string(raw), "\x00")
	files := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			files = append(files, part)
		}
	}
	sort.Strings(files)

	digest := sha256.New()
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", 0, err
		}
		fileSum := sha256.Sum256(data)
		digest.Write([]byte(path))
		digest.Write([]byte{0})
		digest.Write([]byte(hex.EncodeToString(fileSum[:])))
		digest.Write([]byte{0})
	}

	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), len(files), nil
}

func contractDigests() ([]FileDigest, error) {
	digests := make([]FileDigest, 0, len(contractFiles))
	for _, path := range contractFiles {
		digest, err := fileDigest(path)
		if err != nil {
			return nil, err
		}
		digests = append(digests, digest)
	}
	return digests, nil
}

func fileDigest(path string) (FileDigest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileDigest{}, err
	}
	sum := sha256.Sum256(data)
	return FileDigest{
		Path:   path,
		SHA256: "sha256:" + hex.EncodeToString(sum[:]),
	}, nil
}

func moduleDigests() ([]ModuleDigest, error) {
	raw, err := runRaw("go", "list", "-m", "-json", "all")
	if err != nil {
		return nil, err
	}

	type goModule struct {
		Path    string
		Version string
		Main    bool
		Replace *struct {
			Path    string
			Version string
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	var modules []ModuleDigest
	for {
		var module goModule
		if err := decoder.Decode(&module); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		digest := ModuleDigest{
			Path:    module.Path,
			Version: module.Version,
			Main:    module.Main,
		}
		if module.Replace != nil {
			digest.Replace = &ModuleReplace{
				Path:    module.Replace.Path,
				Version: module.Replace.Version,
			}
		}
		modules = append(modules, digest)
	}
	return modules, nil
}

func treeState() string {
	status, err := runTrimmed("git", "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return "unknown"
	}
	if status == "" {
		return "clean"
	}
	return "dirty"
}

func toolVersion(name string, args ...string) string {
	if _, err := exec.LookPath(name); err != nil {
		return "missing"
	}
	output, err := runTrimmed(name, args...)
	if err != nil {
		return "error: " + firstLine(err.Error())
	}
	return firstLine(output)
}

func runTrimmedDefault(fallback string, name string, args ...string) string {
	output, err := runTrimmed(name, args...)
	if err != nil {
		return fallback
	}
	return output
}

func runTrimmed(name string, args ...string) (string, error) {
	output, err := runRaw(name, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func runRaw(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func envDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if idx := strings.IndexByte(value, '\n'); idx >= 0 {
		return value[:idx]
	}
	return value
}

func requireNonEmpty(failures *[]string, field string, value string) {
	if strings.TrimSpace(value) == "" {
		*failures = append(*failures, field+" is required")
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
