package configx

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type appConfig struct {
	Name    string        `config:"APP_NAME" required:"true"`
	Port    int           `config:"PORT" default:"8080"`
	Debug   bool          `config:"DEBUG"`
	Timeout time.Duration `config:"TIMEOUT"`
	Token   SecretString  `config:"API_TOKEN" secret:"true"`
}

func TestLoaderMergesSourcesLastWinsAndSanitizesSecrets(t *testing.T) {
	loader := NewLoader().
		AddSource(NewMapSource("defaults", map[string]string{"APP_NAME": "base", "PORT": "1000", "API_TOKEN": "one"})).
		AddSource(NewSecretMapSource("override", map[string]string{"PORT": "2000", "API_TOKEN": "two"}, []string{"API_TOKEN"}))
	result, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got, _ := result.Get("PORT"); got != "2000" {
		t.Fatalf("PORT=%q", got)
	}
	if result.Values["PORT"].Secret {
		t.Fatal("PORT should not be secret")
	}
	if !result.Values["API_TOKEN"].Secret {
		t.Fatal("API_TOKEN should be secret")
	}
	if result.Sanitize().Values["API_TOKEN"].Value != redactionMarker {
		t.Fatalf("secret leaked in sanitize")
	}
}

func TestEnvSourceReadsOnlyExplicitKeys(t *testing.T) {
	t.Setenv("CFG_NAME", "configx")
	t.Setenv("CFG_OTHER", "ignored")
	result, err := NewLoader().AddSource(NewEnvSource("CFG_", []string{"NAME"})).Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := result.Get("NAME"); !ok || got != "configx" {
		t.Fatalf("NAME=(%q,%v)", got, ok)
	}
	if _, ok := result.Get("OTHER"); ok {
		t.Fatal("unexpected implicit env discovery")
	}
}

func TestEnvFileAndJSONSourcesRequireExplicitPaths(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, "app.env")
	if err := os.WriteFile(envPath, []byte("APP_NAME=file\nAPI_TOKEN=hidden\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(dir, "app.json")
	if err := os.WriteFile(jsonPath, []byte(`{"PORT":9090,"DEBUG":true,"TIMEOUT":"2s"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := NewLoader().AddSource(NewEnvFileSource(envPath)).AddSource(NewJSONFileSource(jsonPath)).Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var cfg appConfig
	if err := Decode(result, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "file" || cfg.Port != 9090 || !cfg.Debug || cfg.Timeout != 2*time.Second {
		t.Fatalf("decoded %#v", cfg)
	}
	if cfg.Token.String() != redactionMarker || strings.Contains(cfg.Token.String(), "hidden") {
		t.Fatalf("token leaked: %s", cfg.Token.String())
	}
}

func TestDecodeRequiredAndDefaults(t *testing.T) {
	var cfg appConfig
	err := Decode(LoadResult{Values: Map{}}, &cfg)
	if err == nil || !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	result := LoadResult{Values: Map{"APP_NAME": {Key: "APP_NAME", Value: "configx"}}}
	if err := Decode(result, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("default port=%d", cfg.Port)
	}
}

func TestNoImplicitConfigDiscovery(t *testing.T) {
	loader := NewLoader()
	result, err := loader.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Values) != 0 {
		t.Fatalf("unexpected values: %#v", result.Values)
	}
}

func TestDecodeConfigTagOptionsAndSecretJSONRedaction(t *testing.T) {
	type optionConfig struct {
		Password SecretString `config:"database.password,required,secret"`
		Host     string       `config:"database.host,default=localhost"`
	}
	result := LoadResult{Values: Map{
		"DATABASE_PASSWORD": {Key: "DATABASE_PASSWORD", Value: "super-secret", Secret: true},
	}}
	var cfg optionConfig
	if err := result.Decode(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Password.Reveal() != "super-secret" {
		t.Fatal("secret value was not decoded")
	}
	if cfg.Host != "localhost" {
		t.Fatalf("default host=%q", cfg.Host)
	}
	encoded, err := json.Marshal(cfg.Password)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "super-secret") || string(encoded) != `"`+redactionMarker+`"` {
		t.Fatalf("secret leaked through json: %s", encoded)
	}
	if strings.Contains(cfg.Password.GoString(), "super-secret") {
		t.Fatalf("secret leaked through GoString: %#v", cfg.Password)
	}
}

func TestDecodeErrorDoesNotExposeSecretValue(t *testing.T) {
	type badConfig struct {
		PasswordLength int `config:"DB_PASSWORD,required,secret"`
	}
	result := LoadResult{Values: Map{
		"DB_PASSWORD": {Key: "DB_PASSWORD", Value: "super-secret", Secret: true},
	}}
	var cfg badConfig
	err := Decode(result, &cfg)
	if err == nil {
		t.Fatal("expected decode error")
	}
	if strings.Contains(err.Error(), "super-secret") {
		t.Fatalf("secret leaked in error: %v", err)
	}
	if cause := errors.Unwrap(err); cause != nil && strings.Contains(cause.Error(), "super-secret") {
		t.Fatalf("secret leaked in cause: %v", cause)
	}
}
