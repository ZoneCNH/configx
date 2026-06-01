package configx

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDotEnvFileSourceSupportsDotEnvName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("APP_NAME=configx\nAPI_TOKEN=hidden\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := LoadEnvFile(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := result.Get("APP_NAME"); !ok || got != "configx" {
		t.Fatalf("APP_NAME=(%q,%v)", got, ok)
	}
	if result.Sanitize().Values["API_TOKEN"].Value != redactionMarker {
		t.Fatal("API_TOKEN should be redacted")
	}
}

func TestTOMLAndYAMLFileSourcesDecodeNestedValues(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "app.toml")
	if err := os.WriteFile(tomlPath, []byte(`
[database]
host = "toml.local"
port = 5432
password = "from-toml"

[server]
debug = false
`), 0o600); err != nil {
		t.Fatal(err)
	}
	yamlPath := filepath.Join(dir, "app.yml")
	if err := os.WriteFile(yamlPath, []byte(`
database:
  host: yaml.local
  port: 6432
server:
  debug: true
`), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := NewLoader().
		AddSource(NewTOMLFileSource(tomlPath)).
		AddSource(NewYAMLFileSource(yamlPath)).
		Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	type databaseConfig struct {
		Host     string       `config:"database.host" required:"true"`
		Port     int          `config:"database.port" required:"true"`
		Password SecretString `config:"database.password" required:"true"`
		Debug    bool         `config:"server.debug"`
	}
	var cfg databaseConfig
	if err := result.Decode(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "yaml.local" || cfg.Port != 6432 || !cfg.Debug {
		t.Fatalf("decoded %#v", cfg)
	}
	if cfg.Password.Reveal() != "from-toml" {
		t.Fatal("secret value was not decoded from toml")
	}
	if result.Sanitize().Values["database.password"].Value != redactionMarker {
		t.Fatal("database password should be redacted")
	}
	if len(result.Sources) != 2 || result.Sources[0].Kind != "toml" || result.Sources[1].Kind != "yaml" {
		t.Fatalf("unexpected source reports: %#v", result.Sources)
	}
}

func TestLoadTOMLFileAndLoadYAMLFileConvenienceFunctions(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "service.toml")
	if err := os.WriteFile(tomlPath, []byte("name = \"toml\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	yamlPath := filepath.Join(dir, "service.yaml")
	if err := os.WriteFile(yamlPath, []byte("name: yaml\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tomlResult, err := LoadTOMLFile(context.Background(), tomlPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := tomlResult.Get("name"); got != "toml" {
		t.Fatalf("toml name=%q", got)
	}
	yamlResult, err := LoadYAMLFile(context.Background(), yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := yamlResult.Get("name"); got != "yaml" {
		t.Fatalf("yaml name=%q", got)
	}
}
