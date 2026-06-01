package contracts

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bytechainx/configx/pkg/configx"
)

type schemaProperty struct {
	Type    string   `json:"type"`
	Enum    []string `json:"enum"`
	Minimum *int     `json:"minimum"`
}

type objectSchema struct {
	Required   []string                  `json:"required"`
	Properties map[string]schemaProperty `json:"properties"`
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
		string(configx.ErrorKindInternal),
	)
	actual := sortedStrings(schema.Properties["kind"].Enum...)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("error kind contract drift:\nactual:   %#v\nexpected: %#v", actual, expected)
	}
	requireFields(t, schema.Required, "kind", "op", "message", "retryable")
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
	requireFields(t, schema.Required, "name", "status", "checked_at")
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

func sortedStrings(values ...string) []string {
	copied := append([]string(nil), values...)
	sort.Strings(copied)
	return copied
}
