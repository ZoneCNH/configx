package configx

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGenerateSchema_NilPanics(t *testing.T) {
	_, err := GenerateSchema(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestGenerateSchema_NilPointer(t *testing.T) {
	var s *struct{ Name string }
	_, err := GenerateSchema(s)
	if err == nil {
		t.Fatal("expected error for nil pointer")
	}
}

func TestGenerateSchema_NonStruct(t *testing.T) {
	_, err := GenerateSchema("not a struct")
	if err == nil {
		t.Fatal("expected error for non-struct")
	}
}

type schemaTestConfig struct {
	Host     string        `config:"host,required"`
	Port     int           `config:"port,default=8080"`
	Debug    bool          `config:"debug"`
	Timeout  time.Duration `config:"timeout"`
	Secret   SecretString  `config:"secret"`
	Skip     string        `config:"-"`
	Tags     []string      `config:"tags"`
	Settings map[string]string `config:"settings"`
	Nested   schemaNested `config:"nested"`
}

type schemaNested struct {
	Name string `config:"name"`
}

func TestGenerateSchema_BasicStruct(t *testing.T) {
	schemaBytes, err := GenerateSchema(schemaTestConfig{})
	if err != nil {
		t.Fatal(err)
	}

	var schema JSONSchema
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		t.Fatal(err)
	}

	if schema.Type != "object" {
		t.Fatalf("expected type 'object', got %q", schema.Type)
	}
	if schema.Schema != "http://json-schema.org/draft-07/schema#" {
		t.Fatalf("unexpected $schema: %s", schema.Schema)
	}

	// host should exist and be required
	hostProp, ok := schema.Properties["host"]
	if !ok {
		t.Fatal("missing 'host' property")
	}
	if hostProp.Type != "string" {
		t.Fatalf("host type: expected 'string', got %q", hostProp.Type)
	}
	if !stringSliceContains(schema.Required, "host") {
		t.Fatal("expected 'host' in required")
	}

	// port should have default
	portProp, ok := schema.Properties["port"]
	if !ok {
		t.Fatal("missing 'port' property")
	}
	if portProp.Type != "integer" {
		t.Fatalf("port type: expected 'integer', got %q", portProp.Type)
	}
	if portProp.Default != "8080" {
		t.Fatalf("port default: expected '8080', got %v", portProp.Default)
	}

	// debug should be boolean
	debugProp, ok := schema.Properties["debug"]
	if !ok {
		t.Fatal("missing 'debug' property")
	}
	if debugProp.Type != "boolean" {
		t.Fatalf("debug type: expected 'boolean', got %q", debugProp.Type)
	}

	// timeout should be string with duration format
	timeoutProp, ok := schema.Properties["timeout"]
	if !ok {
		t.Fatal("missing 'timeout' property")
	}
	if timeoutProp.Type != "string" {
		t.Fatalf("timeout type: expected 'string', got %q", timeoutProp.Type)
	}
	if timeoutProp.Format != "duration" {
		t.Fatalf("timeout format: expected 'duration', got %q", timeoutProp.Format)
	}

	// secret should be string with password format
	secretProp, ok := schema.Properties["secret"]
	if !ok {
		t.Fatal("missing 'secret' property")
	}
	if secretProp.Format != "password" {
		t.Fatalf("secret format: expected 'password', got %q", secretProp.Format)
	}

	// skip should not be in properties
	if _, ok := schema.Properties["skip"]; ok {
		t.Fatal("'skip' should not be in properties")
	}

	// tags should be array
	tagsProp, ok := schema.Properties["tags"]
	if !ok {
		t.Fatal("missing 'tags' property")
	}
	if tagsProp.Type != "array" {
		t.Fatalf("tags type: expected 'array', got %q", tagsProp.Type)
	}
	if tagsProp.Items == nil {
		t.Fatal("tags items should not be nil")
	}
	if tagsProp.Items.Type != "string" {
		t.Fatalf("tags items type: expected 'string', got %q", tagsProp.Items.Type)
	}

	// settings should be object
	settingsProp, ok := schema.Properties["settings"]
	if !ok {
		t.Fatal("missing 'settings' property")
	}
	if settingsProp.Type != "object" {
		t.Fatalf("settings type: expected 'object', got %q", settingsProp.Type)
	}

	// nested should be object with sub-properties
	nestedProp, ok := schema.Properties["nested"]
	if !ok {
		t.Fatal("missing 'nested' property")
	}
	if nestedProp.Type != "object" {
		t.Fatalf("nested type: expected 'object', got %q", nestedProp.Type)
	}
	if nestedProp.Properties == nil {
		t.Fatal("nested properties should not be nil")
	}
	if _, ok := nestedProp.Properties["name"]; !ok {
		t.Fatal("nested should have 'name' property")
	}
}

func TestGenerateSchema_PointerInput(t *testing.T) {
	cfg := &schemaTestConfig{Host: "localhost"}
	_, err := GenerateSchema(cfg)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateSchema_IntBounds(t *testing.T) {
	type intConfig struct {
		A int8  `config:"a"`
		B int16 `config:"b"`
		C int32 `config:"c"`
		D uint8 `config:"d"`
	}

	schemaBytes, err := GenerateSchema(intConfig{})
	if err != nil {
		t.Fatal(err)
	}
	var schema JSONSchema
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		t.Fatal(err)
	}

	a := schema.Properties["a"]
	if a.Minimum == nil || *a.Minimum != -128 {
		t.Fatal("int8 min should be -128")
	}
	if a.Maximum == nil || *a.Maximum != 127 {
		t.Fatal("int8 max should be 127")
	}

	d := schema.Properties["d"]
	if d.Minimum == nil || *d.Minimum != 0 {
		t.Fatal("uint8 min should be 0")
	}
	if d.Maximum == nil || *d.Maximum != 255 {
		t.Fatal("uint8 max should be 255")
	}
}

func TestGenerateSchema_Float(t *testing.T) {
	type floatConfig struct {
		Rate float64 `config:"rate"`
	}
	schemaBytes, err := GenerateSchema(floatConfig{})
	if err != nil {
		t.Fatal(err)
	}
	var schema JSONSchema
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		t.Fatal(err)
	}
	if schema.Properties["rate"].Type != "number" {
		t.Fatal("float64 should map to number")
	}
}

func TestGenerateSchema_ValidJSON(t *testing.T) {
	schemaBytes, err := GenerateSchema(schemaTestConfig{})
	if err != nil {
		t.Fatal(err)
	}
	// Must be valid JSON
	var raw any
	if err := json.Unmarshal(schemaBytes, &raw); err != nil {
		t.Fatalf("generated schema is not valid JSON: %v", err)
	}
}
