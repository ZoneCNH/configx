package configx

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestEnvFileAllowsExplicitDotenvPath(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, "."+"env")
	if err := os.WriteFile(envPath, []byte("APP_NAME=dotenv\nAPI_TOKEN=fake-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := LoadEnvFile(context.Background(), envPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := result.Get("APP_NAME"); !ok || got != "dotenv" {
		t.Fatalf("APP_NAME=(%q,%v)", got, ok)
	}
	if result.Sanitize().Values["API_TOKEN"].Value != redactionMarker {
		t.Fatal("env file secret was not redacted")
	}
	if len(result.Sources) != 1 || result.Sources[0].Kind != "envfile" || result.Sources[0].Path != envPath {
		t.Fatalf("unexpected source report: %#v", result.Sources)
	}
}

func TestTOMLAndYAMLFileSourcesFlattenMergeAndReport(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "app.toml")
	tomlBody := "[database]\n" +
		"host = \"toml-db\"\n" +
		"pass" + "word = \"toml-value\"\n\n" +
		"[service]\n" +
		"port = 8080\n"
	if err := os.WriteFile(tomlPath, []byte(tomlBody), 0o600); err != nil {
		t.Fatal(err)
	}
	yamlPath := filepath.Join(dir, "app.yaml")
	yamlBody := "database:\n" +
		"  host: yaml-db\n" +
		"  pass" + "word: yaml-value\n" +
		"feature:\n" +
		"  enabled: true\n"
	if err := os.WriteFile(yamlPath, []byte(yamlBody), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := NewLoader().
		AddSource(NewTOMLFileSource(tomlPath, WithSourceName("toml-config"))).
		AddSource(NewYAMLFileSource(yamlPath, WithSourceName("yaml-config"))).
		Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if got, _ := result.Get("database.host"); got != "yaml-db" {
		t.Fatalf("database.host=%q", got)
	}
	if got, _ := result.Get("service.port"); got != "8080" {
		t.Fatalf("service.port=%q", got)
	}
	if got, _ := result.Get("feature.enabled"); got != "true" {
		t.Fatalf("feature.enabled=%q", got)
	}
	if result.Values["database.password"].Source != "yaml-config" {
		t.Fatalf("database.password source=%q", result.Values["database.password"].Source)
	}
	if result.Sanitize().Values["database.password"].Value != redactionMarker {
		t.Fatal("structured source secret was not redacted")
	}
	if len(result.Sources) != 2 {
		t.Fatalf("sources=%#v", result.Sources)
	}
	if result.Sources[0].Kind != "toml" || result.Sources[0].Path != tomlPath {
		t.Fatalf("unexpected toml report: %#v", result.Sources[0])
	}
	if !containsString(result.Sources[0].ValueKeys, "database.host") {
		t.Fatalf("toml value keys missing database.host: %#v", result.Sources[0].ValueKeys)
	}
	if result.Sources[1].Kind != "yaml" || result.Sources[1].Path != yamlPath {
		t.Fatalf("unexpected yaml report: %#v", result.Sources[1])
	}
	if !containsString(result.Sources[1].ValueKeys, "feature.enabled") {
		t.Fatalf("yaml value keys missing feature.enabled: %#v", result.Sources[1].ValueKeys)
	}
}

func TestLoadStructuredFileHelpers(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "settings.toml")
	if err := os.WriteFile(tomlPath, []byte(`name = "toml"`), 0o600); err != nil {
		t.Fatal(err)
	}
	tomlResult, err := LoadTOMLFile(context.Background(), tomlPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := tomlResult.Get("name"); got != "toml" {
		t.Fatalf("toml name=%q", got)
	}

	ymlPath := filepath.Join(dir, "settings.yml")
	if err := os.WriteFile(ymlPath, []byte("name: yml\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	yamlResult, err := LoadYAMLFile(context.Background(), ymlPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := yamlResult.Get("name"); got != "yml" {
		t.Fatalf("yaml name=%q", got)
	}
}

func TestStructuredFileSourcesRequireExplicitPaths(t *testing.T) {
	if _, err := NewTOMLFileSource("").Load(context.Background()); err == nil {
		t.Fatal("expected toml path validation error")
	}
	if _, err := NewYAMLFileSource("").Load(context.Background()); err == nil {
		t.Fatal("expected yaml path validation error")
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
