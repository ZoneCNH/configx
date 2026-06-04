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
	"regexp"
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
	"contracts/version.schema.json",
	"contracts/metrics.md",
	"contracts/manifest.schema.json",
	"release/manifest/template.json",
}

const (
	defaultManifestPath  = "release/manifest/latest.json"
	gateReportPath       = "release/evidence/gate-report.json"
	redactionReportPath  = "release/evidence/redaction-report.json"
	contractHashesPath   = "release/evidence/contract-hashes.json"
	evidenceStatusPassed = "passed"
)

var evidenceSecretAssignments = regexp.MustCompile(`(?i)\b(password|passwd|token|access_key|secret_key)(\s*[:=]\s*)("?)\S+`)

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

type GateReport struct {
	Module      string            `json:"module"`
	Version     string            `json:"version"`
	Commit      string            `json:"commit"`
	TreeState   string            `json:"tree_state"`
	Status      string            `json:"status"`
	GeneratedAt string            `json:"generated_at"`
	Checks      map[string]string `json:"checks"`
	Commands    []EvidenceCommand `json:"commands"`
	Downstream  []DownstreamCheck `json:"downstream"`
	Artifacts   []string          `json:"artifacts"`
}

type EvidenceCommand struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Status  string `json:"status"`
}

type DownstreamCheck struct {
	Name     string `json:"name"`
	Module   string `json:"module"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

type RedactionReport struct {
	Module      string           `json:"module"`
	Version     string           `json:"version"`
	Commit      string           `json:"commit"`
	Status      string           `json:"status"`
	GeneratedAt string           `json:"generated_at"`
	Coverage    []RedactionCheck `json:"coverage"`
}

type RedactionCheck struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

type ContractHashesReport struct {
	Module       string       `json:"module"`
	Version      string       `json:"version"`
	Commit       string       `json:"commit"`
	SourceDigest string       `json:"source_digest"`
	GeneratedAt  string       `json:"generated_at"`
	Contracts    []FileDigest `json:"contracts"`
}

func main() {
	out := flag.String("out", defaultManifestPath, "release manifest output path")
	verify := flag.String("verify", "", "verify an existing release manifest instead of generating one")
	requirePassed := flag.Bool("require-passed", false, "require all release checks to be passed during verification")
	requireClean := flag.Bool("require-clean", false, "require a clean git tree during verification")
	expectVersion := flag.String("expect-version", "", "require the manifest version to match this value during verification")
	flag.Parse()

	if *verify != "" {
		if err := verifyManifest(*verify, *requirePassed, *requireClean, *expectVersion); err != nil {
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
	manifest.Artifacts = releaseArtifactPaths(*out)
	if err := writeReleaseEvidence(*out, manifest); err != nil {
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
		Artifacts: releaseArtifactPaths(defaultManifestPath),
		Notes: Notes{
			BreakingChanges: "none",
			KnownRisks:      []string{},
		},
	}, nil
}

func verifyManifest(path string, requirePassed bool, requireClean bool, expectVersion string) error {
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
	if expectVersion = strings.TrimSpace(expectVersion); expectVersion != "" && got.Version != expectVersion {
		failures = append(failures, fmt.Sprintf("version mismatch: got %q, want %q", got.Version, expectVersion))
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
	failures = append(failures, validateReleaseArtifacts(path, data, got, requirePassed)...)
	if got.Tools["go"] == "" {
		failures = append(failures, "tools.go must be recorded")
	}
	failures = append(failures, validateChecks(got.Checks, requirePassed)...)

	if len(failures) > 0 {
		return errors.New("release evidence verification failed:\n - " + strings.Join(failures, "\n - "))
	}
	return nil
}

func releaseArtifactPaths(manifestPath string) []string {
	manifestPath = filepath.ToSlash(manifestPath)
	return []string{
		manifestPath,
		manifestPath + ".sha256",
		gateReportPath,
		redactionReportPath,
		contractHashesPath,
	}
}

func writeReleaseEvidence(path string, manifest Manifest) error {
	if err := writeManifest(path, manifest); err != nil {
		return err
	}
	manifestBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := writeTextFile(path+".sha256", manifestSHA256Line(path, manifestBytes)); err != nil {
		return err
	}
	if err := writeJSONFile(gateReportPath, buildGateReport(manifest)); err != nil {
		return err
	}
	if err := writeJSONFile(redactionReportPath, buildRedactionReport(manifest)); err != nil {
		return err
	}
	return writeJSONFile(contractHashesPath, buildContractHashesReport(manifest))
}

func writeManifest(path string, manifest Manifest) error {
	return writeJSONFile(path, manifest)
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func writeTextFile(path string, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(value), 0o644)
}

func manifestSHA256Line(path string, data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x  %s\n", sum[:], filepath.ToSlash(path))
}

func buildGateReport(manifest Manifest) GateReport {
	return GateReport{
		Module:      manifest.Module,
		Version:     manifest.Version,
		Commit:      manifest.Commit,
		TreeState:   manifest.TreeState,
		Status:      aggregateStatus(manifest.Checks),
		GeneratedAt: manifest.GeneratedAt,
		Checks:      orderedChecks(manifest.Checks),
		Commands: []EvidenceCommand{
			{Name: "unit", Command: "GOWORK=off go test ./...", Status: manifest.Checks["unit_test"]},
			{Name: "ci", Command: "GOWORK=off make ci", Status: aggregateNamedStatus(manifest.Checks, "fmt", "vet", "lint", "unit_test", "race_test", "boundary", "secret_scan", "security", "contract")},
			{Name: "ci_extended", Command: "GOWORK=off make ci-extended", Status: aggregateNamedStatus(manifest.Checks, "fmt", "vet", "lint", "unit_test", "race_test", "boundary", "secret_scan", "security", "contract")},
			{Name: "release_check", Command: "XLIB_CONTEXT=release_verify GOWORK=off make release-check", Status: aggregateStatus(manifest.Checks)},
			{Name: "release_final_check", Command: "XLIB_CONTEXT=release_verify GOWORK=off make release-final-check", Status: aggregateStatus(manifest.Checks)},
		},
		Downstream: []DownstreamCheck{
			{Name: "baselibx", Module: "github.com/ZoneCNH/baselibx", Status: manifest.Checks["integration"], Evidence: "scripts/run_integration.sh"},
			{Name: "corekit", Module: "example.com/acme/corekit", Status: manifest.Checks["integration"], Evidence: "scripts/run_integration.sh"},
		},
		Artifacts: append([]string(nil), manifest.Artifacts...),
	}
}

func buildRedactionReport(manifest Manifest) RedactionReport {
	status := aggregateNamedStatus(manifest.Checks, "unit_test", "secret_scan", "contract")
	return RedactionReport{
		Module:      manifest.Module,
		Version:     manifest.Version,
		Commit:      manifest.Commit,
		Status:      status,
		GeneratedAt: manifest.GeneratedAt,
		Coverage: []RedactionCheck{
			{Name: "secret_string_stringer", Status: status, Evidence: "internal/foundationx/foundationx_test.go"},
			{Name: "secret_string_json", Status: status, Evidence: "contracts/foundationx_contract_test.go"},
			{Name: "load_result_sanitize", Status: status, Evidence: "pkg/configx/core_test.go"},
			{Name: "decode_secret_tag", Status: status, Evidence: "pkg/configx/core_test.go"},
			{Name: "release_evidence_secret_scan", Status: manifest.Checks["secret_scan"], Evidence: "scripts/check_release_evidence.sh"},
		},
	}
}

func buildContractHashesReport(manifest Manifest) ContractHashesReport {
	return ContractHashesReport{
		Module:       manifest.Module,
		Version:      manifest.Version,
		Commit:       manifest.Commit,
		SourceDigest: manifest.SourceDigest,
		GeneratedAt:  manifest.GeneratedAt,
		Contracts:    append([]FileDigest(nil), manifest.Contracts...),
	}
}

func aggregateStatus(checks map[string]string) string {
	return aggregateNamedStatus(checks, checkNames...)
}

func aggregateNamedStatus(checks map[string]string, names ...string) string {
	for _, name := range names {
		if strings.TrimSpace(checks[name]) != evidenceStatusPassed {
			return "unknown"
		}
	}
	return evidenceStatusPassed
}

func orderedChecks(checks map[string]string) map[string]string {
	ordered := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		ordered[name] = checks[name]
	}
	return ordered
}

func validateReleaseArtifacts(manifestPath string, manifestData []byte, manifest Manifest, requirePassed bool) []string {
	var failures []string
	for _, required := range releaseArtifactPaths(manifestPath) {
		if !contains(manifest.Artifacts, required) {
			failures = append(failures, "artifacts must include "+required)
		}
	}
	for _, artifact := range manifest.Artifacts {
		if strings.TrimSpace(artifact) == "" {
			failures = append(failures, "artifacts must not contain empty paths")
			continue
		}
		if _, err := os.Stat(artifact); err != nil {
			failures = append(failures, fmt.Sprintf("artifact %s is not readable: %v", artifact, err))
		}
	}
	failures = append(failures, validateManifestChecksum(manifestPath, manifestData)...)
	failures = append(failures, validateGateReport(manifest, requirePassed)...)
	failures = append(failures, validateRedactionReport(manifest, requirePassed)...)
	failures = append(failures, validateContractHashesReport(manifest)...)
	return failures
}

func validateManifestChecksum(manifestPath string, manifestData []byte) []string {
	shaPath := manifestPath + ".sha256"
	data, err := os.ReadFile(shaPath)
	if err != nil {
		return []string{fmt.Sprintf("artifact %s is not readable: %v", shaPath, err)}
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return []string{shaPath + " must contain a sha256 digest"}
	}
	sum := sha256.Sum256(manifestData)
	want := hex.EncodeToString(sum[:])
	var failures []string
	if fields[0] != want {
		failures = append(failures, fmt.Sprintf("%s digest mismatch: got %q, want %q", shaPath, fields[0], want))
	}
	if len(fields) > 1 && fields[1] != filepath.ToSlash(manifestPath) {
		failures = append(failures, fmt.Sprintf("%s path mismatch: got %q, want %q", shaPath, fields[1], filepath.ToSlash(manifestPath)))
	}
	return failures
}

func validateGateReport(manifest Manifest, requirePassed bool) []string {
	var report GateReport
	if failures := readEvidenceJSON(gateReportPath, &report); len(failures) > 0 {
		return failures
	}
	var failures []string
	failures = append(failures, requireEvidenceIdentity(gateReportPath, report.Module, report.Version, report.Commit, manifest)...)
	if report.TreeState != manifest.TreeState {
		failures = append(failures, fmt.Sprintf("%s tree_state mismatch: got %q, want %q", gateReportPath, report.TreeState, manifest.TreeState))
	}
	if !reflect.DeepEqual(report.Checks, orderedChecks(manifest.Checks)) {
		failures = append(failures, gateReportPath+" checks do not match manifest checks")
	}
	if !reflect.DeepEqual(report.Artifacts, manifest.Artifacts) {
		failures = append(failures, gateReportPath+" artifacts do not match manifest artifacts")
	}
	for _, command := range report.Commands {
		if command.Name == "" || command.Command == "" || command.Status == "" {
			failures = append(failures, gateReportPath+" commands must include name, command, and status")
		}
		if requirePassed && command.Status != evidenceStatusPassed {
			failures = append(failures, fmt.Sprintf("%s command %s must be passed, got %q", gateReportPath, command.Name, command.Status))
		}
	}
	if len(report.Downstream) == 0 {
		failures = append(failures, gateReportPath+" must include downstream adoption evidence")
	}
	for _, downstream := range report.Downstream {
		if downstream.Status == "" || downstream.Evidence == "" {
			failures = append(failures, gateReportPath+" downstream entries must include status and evidence")
		}
		if requirePassed && downstream.Status != evidenceStatusPassed {
			failures = append(failures, fmt.Sprintf("%s downstream %s must be passed, got %q", gateReportPath, downstream.Name, downstream.Status))
		}
	}
	if requirePassed && report.Status != evidenceStatusPassed {
		failures = append(failures, fmt.Sprintf("%s status must be passed, got %q", gateReportPath, report.Status))
	}
	return failures
}

func validateRedactionReport(manifest Manifest, requirePassed bool) []string {
	var report RedactionReport
	if failures := readEvidenceJSON(redactionReportPath, &report); len(failures) > 0 {
		return failures
	}
	var failures []string
	failures = append(failures, requireEvidenceIdentity(redactionReportPath, report.Module, report.Version, report.Commit, manifest)...)
	if len(report.Coverage) == 0 {
		failures = append(failures, redactionReportPath+" must include redaction coverage")
	}
	if requirePassed && report.Status != evidenceStatusPassed {
		failures = append(failures, fmt.Sprintf("%s status must be passed, got %q", redactionReportPath, report.Status))
	}
	for _, coverage := range report.Coverage {
		if coverage.Name == "" || coverage.Status == "" || coverage.Evidence == "" {
			failures = append(failures, redactionReportPath+" coverage entries must include name, status, and evidence")
		}
		if requirePassed && coverage.Status != evidenceStatusPassed {
			failures = append(failures, fmt.Sprintf("%s coverage %s must be passed, got %q", redactionReportPath, coverage.Name, coverage.Status))
		}
	}
	return failures
}

func validateContractHashesReport(manifest Manifest) []string {
	var report ContractHashesReport
	if failures := readEvidenceJSON(contractHashesPath, &report); len(failures) > 0 {
		return failures
	}
	var failures []string
	failures = append(failures, requireEvidenceIdentity(contractHashesPath, report.Module, report.Version, report.Commit, manifest)...)
	if report.SourceDigest != manifest.SourceDigest {
		failures = append(failures, fmt.Sprintf("%s source_digest mismatch", contractHashesPath))
	}
	if !reflect.DeepEqual(report.Contracts, manifest.Contracts) {
		failures = append(failures, contractHashesPath+" contracts do not match manifest contracts")
	}
	return failures
}

func readEvidenceJSON(path string, target any) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("artifact %s is not readable: %v", path, err)}
	}
	if err := json.Unmarshal(data, target); err != nil {
		return []string{fmt.Sprintf("artifact %s is not valid JSON: %v", path, err)}
	}
	return nil
}

func requireEvidenceIdentity(path string, module string, version string, commit string, manifest Manifest) []string {
	var failures []string
	if module != manifest.Module {
		failures = append(failures, fmt.Sprintf("%s module mismatch: got %q, want %q", path, module, manifest.Module))
	}
	if version != manifest.Version {
		failures = append(failures, fmt.Sprintf("%s version mismatch: got %q, want %q", path, version, manifest.Version))
	}
	if commit != manifest.Commit {
		failures = append(failures, fmt.Sprintf("%s commit mismatch: got %q, want %q", path, commit, manifest.Commit))
	}
	return failures
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
		return "error: " + firstLine(sanitizeForEvidence(err.Error()))
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
		return nil, fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, sanitizeForEvidence(strings.TrimSpace(string(output))))
	}
	return output, nil
}

func sanitizeForEvidence(value string) string {
	value = evidenceSecretAssignments.ReplaceAllString(value, `$1$2$3***`)
	for _, forbidden := range []string{
		"/home/k8s" + "/secrets/env",
		"." + "env",
		"production" + ".yaml",
		"production" + ".yml",
		"config.local" + ".yaml",
		"config.local" + ".yml",
	} {
		value = strings.ReplaceAll(value, forbidden, redactionMarker())
	}
	return value
}

func redactionMarker() string {
	return "***"
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
