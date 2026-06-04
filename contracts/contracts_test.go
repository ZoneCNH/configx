package contracts

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ZoneCNH/configx/pkg/configx"
	foundationx "github.com/ZoneCNH/foundationx"
)

type schemaProperty struct {
	Type    string   `json:"type"`
	Format  string   `json:"format"`
	Enum    []string `json:"enum"`
	Minimum *int     `json:"minimum"`
}

type objectSchema struct {
	Required             []string                  `json:"required"`
	Properties           map[string]schemaProperty `json:"properties"`
	AdditionalProperties *bool                     `json:"additionalProperties"`
}

func TestErrorKindContractMatchesPublicConstants(t *testing.T) {
	schema := readSchema(t, "error.schema.json")

	expected := sortedStrings(
		string(configx.ErrorKindConfig),
		string(configx.ErrorKindValidation),
		string(configx.ErrorKindConnection),
		string(configx.ErrorKindUnavailable),
		string(configx.ErrorKindTimeout),
		string(configx.ErrorKindAuth),
		string(configx.ErrorKindConflict),
		string(configx.ErrorKindRateLimit),
		string(configx.ErrorKindCanceled),
		string(configx.ErrorKindNotFound),
		string(configx.ErrorKindAlreadyExists),
		string(configx.ErrorKindInternal),
	)
	actual := sortedStrings(schema.Properties["kind"].Enum...)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("error kind contract drift:\nactual:   %#v\nexpected: %#v", actual, expected)
	}
	requireNoAdditionalProperties(t, schema)
	requireFields(t, schema.Required, "kind", "op", "message", "retryable")
	if kindType := schema.Properties["kind"].Type; kindType != "string" {
		t.Fatalf("error kind schema type = %q, want string", kindType)
	}
}

func TestHealthStatusContractMatchesPublicConstants(t *testing.T) {
	schema := readSchema(t, "health.schema.json")

	expected := sortedStrings(
		string(configx.HealthHealthy),
		string(configx.HealthDegraded),
		string(configx.HealthUnhealthy),
	)
	actual := sortedStrings(schema.Properties["status"].Enum...)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("health status contract drift:\nactual:   %#v\nexpected: %#v", actual, expected)
	}
	requireNoAdditionalProperties(t, schema)
	requireFields(t, schema.Required, "name", "status", "checked_at")
	if statusType := schema.Properties["status"].Type; statusType != "string" {
		t.Fatalf("health status schema type = %q, want string", statusType)
	}
	if checkedAtFormat := schema.Properties["checked_at"].Format; checkedAtFormat != "date-time" {
		t.Fatalf("checked_at schema format = %q, want date-time", checkedAtFormat)
	}
	if minimum := schema.Properties["latency_ms"].Minimum; minimum == nil || *minimum != 0 {
		t.Fatalf("latency_ms must define minimum 0, got %#v", minimum)
	}
}

func TestConfigContractMatchesPublicConfig(t *testing.T) {
	schema := readSchema(t, "config.schema.json")
	requireFields(t, schema.Required, "name")

	configType := reflect.TypeOf(configx.Config{})
	requireSchemaFieldMapsToStructField(t, schema, configType, "name", "Name", "string")
	requireSchemaFieldMapsToStructField(t, schema, configType, "timeout_ms", "Timeout", "integer")
	requireSchemaFieldMapsToStructField(t, schema, configType, "secret", "Secret", "string")

	if timeoutField, ok := configType.FieldByName("Timeout"); !ok || timeoutField.Type != reflect.TypeOf(time.Duration(0)) {
		t.Fatalf("Config.Timeout must remain time.Duration, got %v", timeoutField.Type)
	}
	if minimum := schema.Properties["timeout_ms"].Minimum; minimum == nil || *minimum != 0 {
		t.Fatalf("timeout_ms must define minimum 0, got %#v", minimum)
	}
}

func TestMetricsContractDocumentsPublicConstants(t *testing.T) {
	content, err := os.ReadFile("metrics.md")
	if err != nil {
		t.Fatalf("read metrics contract: %v", err)
	}
	text := string(content)
	for _, metric := range []string{
		configx.MetricClientCreatedTotal,
		configx.MetricClientClosedTotal,
		configx.MetricClientErrorsTotal,
		configx.MetricClientHealthStatus,
		configx.MetricClientHealthLatencyMS,
		configx.MetricClientRequestsTotal,
		configx.MetricClientRequestDurationSeconds,
		configx.MetricClientRetriesTotal,
		configx.MetricClientInflight,
	} {
		if !strings.Contains(text, "`"+metric+"`") {
			t.Fatalf("metrics contract does not document %q", metric)
		}
	}
}

func TestVersionContractMatchesFoundationXVersionInfo(t *testing.T) {
	schema := readSchema(t, "version.schema.json")
	requireNoAdditionalProperties(t, schema)
	requireFields(t, schema.Required, "module", "version", "commit", "build_time", "go_version")

	versionType := reflect.TypeOf(foundationx.VersionInfo{})
	requireSchemaFieldMapsToStructField(t, schema, versionType, "module", "Module", "string")
	requireSchemaFieldMapsToStructField(t, schema, versionType, "version", "Version", "string")
	requireSchemaFieldMapsToStructField(t, schema, versionType, "commit", "Commit", "string")
	requireSchemaFieldMapsToStructField(t, schema, versionType, "build_time", "BuildTime", "string")
	requireSchemaFieldMapsToStructField(t, schema, versionType, "go_version", "GoVersion", "string")
}

func TestManifestSchemaPinsReleaseEvidenceShape(t *testing.T) {
	schema := readSchema(t, "manifest.schema.json")
	requireNoAdditionalProperties(t, schema)
	requireFields(t, schema.Required,
		"module",
		"version",
		"commit",
		"tree_sha",
		"source_digest",
		"tracked_file_count",
		"go_version",
		"generated_at",
		"generated_by",
		"tree_state",
		"checks",
		"contracts",
		"dependencies",
		"tools",
		"artifacts",
		"notes",
	)

	for field, schemaType := range map[string]string{
		"version":            "string",
		"commit":             "string",
		"tree_sha":           "string",
		"source_digest":      "string",
		"tracked_file_count": "integer",
		"go_version":         "string",
		"generated_at":       "string",
		"generated_by":       "string",
		"checks":             "object",
		"contracts":          "array",
		"dependencies":       "array",
		"tools":              "object",
		"artifacts":          "array",
		"notes":              "object",
	} {
		if got := schema.Properties[field].Type; got != schemaType {
			t.Fatalf("manifest schema property %q type = %q, want %q", field, got, schemaType)
		}
	}
	if got := sortedStrings(schema.Properties["tree_state"].Enum...); !reflect.DeepEqual(got, []string{"clean", "dirty"}) {
		t.Fatalf("manifest tree_state enum = %#v, want clean/dirty", got)
	}
}

func TestReleaseManifestTemplateCoversVerificationContracts(t *testing.T) {
	content, err := os.ReadFile("../release/manifest/template.json")
	if err != nil {
		t.Fatalf("read release manifest template: %v", err)
	}
	var template struct {
		Checks    map[string]string `json:"checks"`
		Contracts []struct {
			Path string `json:"path"`
		} `json:"contracts"`
	}
	if err := json.Unmarshal(content, &template); err != nil {
		t.Fatalf("parse release manifest template: %v", err)
	}

	requireMapKeys(t, template.Checks,
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
	)

	var contractPaths []string
	for _, contract := range template.Contracts {
		contractPaths = append(contractPaths, contract.Path)
	}
	requireFields(t, contractPaths,
		"contracts/config.schema.json",
		"contracts/error.schema.json",
		"contracts/health.schema.json",
		"contracts/version.schema.json",
		"contracts/metrics.md",
		"contracts/manifest.schema.json",
		"release/manifest/template.json",
	)
}

func requireSchemaFieldMapsToStructField(t *testing.T, schema objectSchema, structType reflect.Type, schemaField string, structField string, schemaType string) {
	t.Helper()

	property, ok := schema.Properties[schemaField]
	if !ok {
		t.Fatalf("schema missing property %q", schemaField)
	}
	if property.Type != schemaType {
		t.Fatalf("schema property %q type = %q, want %q", schemaField, property.Type, schemaType)
	}
	if _, ok := structType.FieldByName(structField); !ok {
		t.Fatalf("%s missing field %s required by schema property %q", structType.Name(), structField, schemaField)
	}
}

func requireNoAdditionalProperties(t *testing.T, schema objectSchema) {
	t.Helper()
	if schema.AdditionalProperties == nil || *schema.AdditionalProperties {
		t.Fatalf("schema must set additionalProperties false, got %#v", schema.AdditionalProperties)
	}
}

func readSchema(t *testing.T, path string) objectSchema {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var schema objectSchema
	if err := json.Unmarshal(content, &schema); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return schema
}

func requireFields(t *testing.T, actual []string, expected ...string) {
	t.Helper()
	fields := make(map[string]bool, len(actual))
	for _, field := range actual {
		fields[field] = true
	}
	for _, field := range expected {
		if !fields[field] {
			t.Fatalf("required fields missing %q from %#v", field, actual)
		}
	}
}

func requireMapKeys[T any](t *testing.T, actual map[string]T, expected ...string) {
	t.Helper()
	for _, key := range expected {
		if _, ok := actual[key]; !ok {
			t.Fatalf("map keys missing %q from %#v", key, actual)
		}
	}
}

func sortedStrings(values ...string) []string {
	copied := append([]string(nil), values...)
	sort.Strings(copied)
	return copied
}
