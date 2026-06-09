package configx

import (
	"testing"
)

func TestSanitizedManifest_Struct(t *testing.T) {
	type cfg struct {
		Host   string
		Port   int
		APIKey string
		Secret string
	}
	c := cfg{Host: "localhost", Port: 5432, APIKey: "sk-abc123", Secret: "hunter2"}
	m := SanitizedManifest(c)
	if m["Host"] != "localhost" {
		t.Errorf("Host = %v, want localhost", m["Host"])
	}
	if m["Port"] != 5432 {
		t.Errorf("Port = %v, want 5432", m["Port"])
	}
	if m["APIKey"] != "***" {
		t.Errorf("APIKey = %v, want ***", m["APIKey"])
	}
	if m["Secret"] != "***" {
		t.Errorf("Secret = %v, want ***", m["Secret"])
	}
}

func TestSanitizedManifest_StructTag(t *testing.T) {
	type cfg struct {
		Name   string
		ApiKey string `secret:"true"`
	}
	c := cfg{Name: "test", ApiKey: "my-secret"}
	m := SanitizedManifest(c)
	if m["Name"] != "test" {
		t.Errorf("Name = %v, want test", m["Name"])
	}
	if m["ApiKey"] != "***" {
		t.Errorf("ApiKey = %v, want ***", m["ApiKey"])
	}
}

func TestSanitizedManifest_Map(t *testing.T) {
	m := SanitizedManifest(map[string]any{
		"host":        "localhost",
		"port":        5432,
		"secret_key":  "abc",
		"password":    "xyz",
		"access_token": "tok",
	})
	if m["host"] != "localhost" {
		t.Errorf("host = %v, want localhost", m["host"])
	}
	if m["port"] != 5432 {
		t.Errorf("port = %v, want 5432", m["port"])
	}
	if m["secret_key"] != "***" {
		t.Errorf("secret_key = %v, want ***", m["secret_key"])
	}
	if m["password"] != "***" {
		t.Errorf("password = %v, want ***", m["password"])
	}
	if m["access_token"] != "***" {
		t.Errorf("access_token = %v, want ***", m["access_token"])
	}
}

func TestSanitizedManifest_NestedStruct(t *testing.T) {
	type db struct {
		Host string
		Port int
		Pass string
	}
	type cfg struct {
		DB db
	}
	c := cfg{DB: db{Host: "localhost", Port: 5432, Pass: "secret"}}
	m := SanitizedManifest(c)
	dbMap, ok := m["DB"].(map[string]any)
	if !ok {
		t.Fatalf("DB is %T, want map[string]any", m["DB"])
	}
	if dbMap["Host"] != "localhost" {
		t.Errorf("DB.Host = %v", dbMap["Host"])
	}
	if dbMap["Pass"] != "***" {
		t.Errorf("DB.Pass = %v, want ***", dbMap["Pass"])
	}
}

func TestSanitizedManifest_NilInput(t *testing.T) {
	if m := SanitizedManifest(nil); m != nil {
		t.Errorf("expected nil, got %v", m)
	}
}

func TestSanitizedManifest_PointerInput(t *testing.T) {
	type cfg struct {
		Host   string
		Secret string
	}
	c := &cfg{Host: "localhost", Secret: "abc"}
	m := SanitizedManifest(c)
	if m["Host"] != "localhost" {
		t.Errorf("Host = %v", m["Host"])
	}
	if m["Secret"] != "***" {
		t.Errorf("Secret = %v, want ***", m["Secret"])
	}
}

func TestSanitizedManifest_NilPointer(t *testing.T) {
	type cfg struct{ Host string }
	var c *cfg
	if m := SanitizedManifest(c); m != nil {
		t.Errorf("expected nil for nil pointer, got %v", m)
	}
}

func TestSanitizedManifest_NonStructNonMap(t *testing.T) {
	m := SanitizedManifest(42)
	if m["_value"] != 42 {
		t.Errorf("expected _value=42, got %v", m["_value"])
	}
}

func TestSanitizedManifest_EmptySecrets(t *testing.T) {
	type cfg struct {
		Name   string
		Secret string
	}
	c := cfg{Name: "test", Secret: ""}
	m := SanitizedManifest(c)
	// Empty secret should still be redacted (non-empty heuristic applies at field name level).
	if m["Secret"] != "***" {
		t.Errorf("Secret = %v, want *** (field name matches secret pattern)", m["Secret"])
	}
}
