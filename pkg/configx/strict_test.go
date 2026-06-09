package configx

import (
	"strings"
	"testing"
)

type strictTestConfig struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Debug   bool   `json:"debug"`
	Secret  string `json:"secret"`
}

func TestStrictDecode_Basic(t *testing.T) {
	data := []byte(`{"host":"localhost","port":8080,"debug":true,"secret":"s3cret"}`)
	var cfg strictTestConfig
	if err := StrictDecode(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "localhost" || cfg.Port != 8080 || !cfg.Debug || cfg.Secret != "s3cret" {
		t.Fatalf("unexpected values: %+v", cfg)
	}
}

func TestStrictDecode_EmptyData(t *testing.T) {
	var cfg strictTestConfig
	err := StrictDecode([]byte{}, &cfg)
	if err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestStrictDecode_NilTarget(t *testing.T) {
	err := StrictDecode([]byte(`{"host":"a"}`), nil)
	if err == nil {
		t.Fatal("expected error for nil target")
	}
}

func TestStrictDecode_UnknownField(t *testing.T) {
	data := []byte(`{"host":"localhost","unknown_field":"value"}`)
	var cfg strictTestConfig
	err := StrictDecode(data, &cfg)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
	if !strings.Contains(err.Error(), "unknown") && !strings.Contains(err.Error(), "StrictDecode") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestStrictDecode_UnknownFieldAllowed(t *testing.T) {
	data := []byte(`{"host":"localhost","unknown_field":"value"}`)
	var cfg strictTestConfig
	err := StrictDecode(data, &cfg, WithAllowUnknownFields())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "localhost" {
		t.Fatalf("unexpected host: %s", cfg.Host)
	}
}

func TestStrictDecode_DuplicateKeys(t *testing.T) {
	data := []byte(`{"host":"first","host":"second"}`)
	var cfg strictTestConfig
	err := StrictDecode(data, &cfg)
	if err == nil {
		t.Fatal("expected error for duplicate keys")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected 'duplicate' in error, got: %v", err)
	}
}

func TestStrictDecode_DuplicateKeysNested(t *testing.T) {
	data := []byte(`{"outer":{"key":"a","key":"b"}}`)
	var cfg map[string]map[string]string
	err := StrictDecode(data, &cfg)
	if err == nil {
		t.Fatal("expected error for nested duplicate keys")
	}
}

func TestStrictDecode_TypeMismatch(t *testing.T) {
	// port should be int, not string
	data := []byte(`{"host":"localhost","port":"not_a_number"}`)
	var cfg strictTestConfig
	err := StrictDecode(data, &cfg)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
}

func TestStrictDecode_MaxDepth(t *testing.T) {
	data := []byte(`{"a":{"b":{"c":"deep"}}}`)
	var cfg map[string]any

	// Depth 2 should fail (nesting goes 3 levels).
	err := StrictDecode(data, &cfg, WithMaxDepth(2))
	if err == nil {
		t.Fatal("expected error for exceeding max depth")
	}
	if !strings.Contains(err.Error(), "depth") {
		t.Fatalf("expected 'depth' in error, got: %v", err)
	}

	// Depth 3 should pass.
	err = StrictDecode(data, &cfg, WithMaxDepth(3))
	if err != nil {
		t.Fatal(err)
	}
}

func TestStrictDecode_MaxDepthZeroMeansNoLimit(t *testing.T) {
	// Deeply nested but maxDepth=0 means no limit.
	data := []byte(`{"a":{"b":{"c":{"d":{"e":"deep"}}}}}`)
	var cfg map[string]any
	err := StrictDecode(data, &cfg, WithMaxDepth(0))
	if err != nil {
		t.Fatal(err)
	}
}

func TestStrictDecode_TrailingData(t *testing.T) {
	data := []byte(`{"host":"localhost"} extra`)
	var cfg strictTestConfig
	err := StrictDecode(data, &cfg)
	if err == nil {
		t.Fatal("expected error for trailing data")
	}
}

func TestStrictDecode_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid json}`)
	var cfg strictTestConfig
	err := StrictDecode(data, &cfg)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestStrictDecode_ArrayTopLevel(t *testing.T) {
	type item struct {
		Name string `json:"name"`
	}
	data := []byte(`[{"name":"a"},{"name":"b"}]`)
	var items []item
	if err := StrictDecode(data, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Name != "a" || items[1].Name != "b" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestStrictDecodeConfig_WithValidation(t *testing.T) {
	type validatedConfig struct {
		Name string `json:"name"`
	}

	// This struct doesn't implement Validator, so just decode.
	data := []byte(`{"name":"test"}`)
	var cfg validatedConfig
	if err := StrictDecodeConfig(data, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "test" {
		t.Fatalf("unexpected name: %s", cfg.Name)
	}
}

type failingValidatedConfig struct {
	Name string `json:"name"`
}

func (failingValidatedConfig) Validate() error {
	return &Error{Kind: ErrorKindValidation, Op: "test.Validate", Message: "always fails"}
}

func TestStrictDecodeConfig_ValidationFails(t *testing.T) {
	data := []byte(`{"name":"test"}`)
	var cfg failingValidatedConfig
	err := StrictDecodeConfig(data, &cfg)
	if err == nil {
		t.Fatal("expected error from validation")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got: %v", err)
	}
}

func TestStrictDecode_DeeplyNestedDuplicate(t *testing.T) {
	data := []byte(`{"level1":{"level2":{"dup":"a","dup":"b"}}}`)
	var cfg map[string]any
	err := StrictDecode(data, &cfg)
	if err == nil {
		t.Fatal("expected error for deeply nested duplicate key")
	}
}

func TestStrictDecode_ObjectInArray(t *testing.T) {
	data := []byte(`{"items":[{"name":"a","name":"b"}]}`)
	var cfg map[string]any
	err := StrictDecode(data, &cfg)
	if err == nil {
		t.Fatal("expected error for duplicate key in object inside array")
	}
}
