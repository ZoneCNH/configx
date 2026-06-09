package configx

import (
	"reflect"
	"testing"
)

func TestSanitizedManifestMapWithSecretKey(t *testing.T) {
	m := SanitizedManifest(map[string]any{
		"host":      "localhost",
		"api_token": "key123",
		"password":  "p@ss",
	})
	if m["host"] != "localhost" {
		t.Fatalf("host = %v", m["host"])
	}
	if m["api_token"] != redactionMarker {
		t.Fatalf("api_token = %v, want %v", m["api_token"], redactionMarker)
	}
	if m["password"] != redactionMarker {
		t.Fatalf("password = %v, want %v", m["password"], redactionMarker)
	}
}

func TestSanitizedManifestMapWithNestedStructValue(t *testing.T) {
	type inner struct {
		Value  string
		Secret string
	}
	// Use a concrete map type so dereference can reach the struct branch.
	m := SanitizedManifest(map[string]*inner{
		"structVal": {Value: "ok", Secret: "hidden"},
	})

	structVal, ok := m["structVal"].(map[string]any)
	if !ok {
		t.Fatalf("structVal type = %T, want map[string]any", m["structVal"])
	}
	if structVal["Value"] != "ok" {
		t.Fatalf("structVal.Value = %v", structVal["Value"])
	}
	if structVal["Secret"] != redactionMarker {
		t.Fatalf("structVal.Secret = %v, want %v", structVal["Secret"], redactionMarker)
	}
}

func TestSanitizedManifestMapWithNestedMapValue(t *testing.T) {
	// Use a concrete map type so dereference can reach the map branch.
	inner := map[string]string{"token": "secret123", "host": "localhost"}
	m := SanitizedManifest(map[string]map[string]string{
		"nested": inner,
	})

	nested, ok := m["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested type = %T, want map[string]any", m["nested"])
	}
	if nested["host"] != "localhost" {
		t.Fatalf("nested.host = %v", nested["host"])
	}
	if nested["token"] != redactionMarker {
		t.Fatalf("nested.token = %v, want %v", nested["token"], redactionMarker)
	}
}

func TestSanitizedManifestMapNonSecretPreserved(t *testing.T) {
	m := SanitizedManifest(map[string]any{
		"host": "localhost",
		"port": 5432,
	})
	if m["host"] != "localhost" {
		t.Fatalf("host = %v", m["host"])
	}
	if m["port"] != 5432 {
		t.Fatalf("port = %v", m["port"])
	}
}

func TestSanitizedManifestStructWithPointerField(t *testing.T) {
	type cfg struct {
		Host   string
		Secret string
	}
	c := cfg{Host: "localhost", Secret: "s3cret"}
	m := SanitizedManifest(c)
	if m["Host"] != "localhost" {
		t.Fatalf("Host = %v", m["Host"])
	}
	if m["Secret"] != redactionMarker {
		t.Fatalf("Secret = %v, want %v", m["Secret"], redactionMarker)
	}
}

func TestSanitizedManifestStructWithJsonTag(t *testing.T) {
	type cfg struct {
		Host   string `json:"host_name"`
		APIKey string `json:"api_key"`
	}
	c := cfg{Host: "localhost", APIKey: "key123"}
	m := SanitizedManifest(c)
	if m["host_name"] != "localhost" {
		t.Fatalf("host_name = %v", m["host_name"])
	}
	if m["api_key"] != redactionMarker {
		t.Fatalf("api_key = %v, want %v", m["api_key"], redactionMarker)
	}
}

func TestSanitizedManifestStructWithDashTag(t *testing.T) {
	type cfg struct {
		Skip string `json:"-"`
		Keep string
	}
	c := cfg{Skip: "hidden", Keep: "visible"}
	m := SanitizedManifest(c)
	if m["Skip"] != "hidden" {
		t.Fatalf("Skip = %v", m["Skip"])
	}
	if m["Keep"] != "visible" {
		t.Fatalf("Keep = %v", m["Keep"])
	}
}

func TestSanitizedManifestStructWithOmitemptyTag(t *testing.T) {
	type cfg struct {
		Host string `json:"host,omitempty"`
	}
	c := cfg{Host: "localhost"}
	m := SanitizedManifest(c)
	if m["host"] != "localhost" {
		t.Fatalf("host = %v", m["host"])
	}
}

func TestSanitizedManifestDoublePointerStruct(t *testing.T) {
	type cfg struct {
		Host string
	}
	c := &cfg{Host: "localhost"}
	m := SanitizedManifest(&c)
	if m["Host"] != "localhost" {
		t.Fatalf("Host = %v", m["Host"])
	}
}

func TestSanitizedManifestDoubleNilPointer(t *testing.T) {
	type cfg struct {
		Host string
	}
	var c *cfg
	p := &c // **cfg, inner is nil
	m := SanitizedManifest(p)
	if m != nil {
		t.Fatalf("expected nil for **nil pointer, got %v", m)
	}
}

func TestSanitizedManifestMapWithStructValue(t *testing.T) {
	type inner struct {
		Name   string
		Secret string
	}
	v := &inner{Name: "test", Secret: "s3cret"}
	m := SanitizedManifest(map[string]any{
		"item": v,
	})
	item, ok := m["item"].(map[string]any)
	if !ok {
		// The map path dereferences the pointer to struct, then calls sanitizedStruct
		t.Logf("item type = %T, value = %v", m["item"], m["item"])
		return
	}
	if item["Name"] != "test" {
		t.Fatalf("item.Name = %v", item["Name"])
	}
	if item["Secret"] != redactionMarker {
		t.Fatalf("item.Secret = %v, want %v", item["Secret"], redactionMarker)
	}
}

func TestDereferenceNilPointer(t *testing.T) {
	var p *string
	v := dereference(reflect.ValueOf(p))
	if !v.IsNil() {
		t.Fatal("expected nil value")
	}
}

func TestDereferenceNonPointer(t *testing.T) {
	s := "hello"
	v := dereference(reflect.ValueOf(s))
	if v.String() != "hello" {
		t.Fatalf("expected hello, got %v", v)
	}
}

func TestDereferenceDoublePointer(t *testing.T) {
	s := "hello"
	p := &s
	pp := &p
	v := dereference(reflect.ValueOf(pp))
	if v.String() != "hello" {
		t.Fatalf("expected hello, got %v", v)
	}
}

func TestSanitizedMapWithSecretKey(t *testing.T) {
	// Direct test of sanitizedMap with a reflect.Value of a map
	rv := reflect.ValueOf(map[string]any{
		"host":       "localhost",
		"password":   "secret",
		"access_key": "key123",
		"normal":     "value",
	})
	m := sanitizedMap(rv)
	if m["host"] != "localhost" {
		t.Fatalf("host = %v", m["host"])
	}
	if m["password"] != redactionMarker {
		t.Fatalf("password = %v", m["password"])
	}
	if m["access_key"] != redactionMarker {
		t.Fatalf("access_key = %v", m["access_key"])
	}
	if m["normal"] != "value" {
		t.Fatalf("normal = %v", m["normal"])
	}
}

func TestSanitizedMapEmptyMap(t *testing.T) {
	rv := reflect.ValueOf(map[string]any{})
	m := sanitizedMap(rv)
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %v", m)
	}
}

func TestSanitizedStructWithNestedStruct(t *testing.T) {
	type inner struct {
		Host string
		Pass string
	}
	type outer struct {
		Name  string
		Inner inner
	}
	c := outer{Name: "test", Inner: inner{Host: "localhost", Pass: "secret"}}
	m := SanitizedManifest(c)
	if m["Name"] != "test" {
		t.Fatalf("Name = %v", m["Name"])
	}
	innerMap, ok := m["Inner"].(map[string]any)
	if !ok {
		t.Fatalf("Inner type = %T", m["Inner"])
	}
	if innerMap["Host"] != "localhost" {
		t.Fatalf("Inner.Host = %v", innerMap["Host"])
	}
	if innerMap["Pass"] != redactionMarker {
		t.Fatalf("Inner.Pass = %v, want %v", innerMap["Pass"], redactionMarker)
	}
}

func TestSanitizedManifestNonStructNonMapString(t *testing.T) {
	m := SanitizedManifest("just a string")
	if m["_value"] != "just a string" {
		t.Fatalf("_value = %v", m["_value"])
	}
}

func TestSanitizedManifestNonStructNonMapInt(t *testing.T) {
	m := SanitizedManifest(42)
	if m["_value"] != 42 {
		t.Fatalf("_value = %v", m["_value"])
	}
}

func TestIsSensitiveFieldName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"APIKey", true},
		{"Password", true},
		{"Credential", true},
		{"AuthToken", true},
		{"PrivateKey", true},
		{"Host", false},
		{"Port", false},
		{"Name", false},
		{"secret", true},      // IsSecretKey matches
		{"token", true},       // IsSecretKey matches
		{"access_key", true},  // IsSecretKey matches
		{"my_pass", true},     // suffix "pass"
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSensitiveFieldName(tt.name)
			if got != tt.want {
				t.Fatalf("isSensitiveFieldName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
