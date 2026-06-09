package configx

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTOMLFileSourceCanceledContext(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	if err := os.WriteFile(path, []byte("key = \"value\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	src := NewTOMLFileSource(path)
	_, err := src.Load(ctx)
	if err == nil {
		t.Fatal("expected canceled context error")
	}
}

func TestTOMLFileSourceNameKindPath(t *testing.T) {
	src := NewTOMLFileSource("/some/path.toml", WithSourceName("custom-toml"))
	if src.Name() != "custom-toml" {
		t.Fatalf("name = %q", src.Name())
	}
	if src.Kind() != "toml" {
		t.Fatalf("kind = %q", src.Kind())
	}
	if src.Path() != "/some/path.toml" {
		t.Fatalf("path = %q", src.Path())
	}
}

func TestTOMLFileSourceValidLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	if err := os.WriteFile(path, []byte("host = \"localhost\"\nport = 8080\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := LoadTOMLFile(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := result.Get("host"); got != "localhost" {
		t.Fatalf("host = %q", got)
	}
}

func TestYAMLFileSourceCanceledContext(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte("key: value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	src := NewYAMLFileSource(path)
	_, err := src.Load(ctx)
	if err == nil {
		t.Fatal("expected canceled context error")
	}
}

func TestYAMLFileSourceNameKindPath(t *testing.T) {
	src := NewYAMLFileSource("/some/path.yaml", WithSourceName("custom-yaml"))
	if src.Name() != "custom-yaml" {
		t.Fatalf("name = %q", src.Name())
	}
	if src.Kind() != "yaml" {
		t.Fatalf("kind = %q", src.Kind())
	}
	if src.Path() != "/some/path.yaml" {
		t.Fatalf("path = %q", src.Path())
	}
}

func TestYAMLFileSourceValidLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte("host: localhost\nport: 8080\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := LoadYAMLFile(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := result.Get("host"); got != "localhost" {
		t.Fatalf("host = %q", got)
	}
}

func TestJSONFileSourceNameKindPath(t *testing.T) {
	src := NewJSONFileSource("/some/path.json", WithSourceName("custom-json"))
	if src.Name() != "custom-json" {
		t.Fatalf("name = %q", src.Name())
	}
	if src.Kind() != "json" {
		t.Fatalf("kind = %q", src.Kind())
	}
	if src.Path() != "/some/path.json" {
		t.Fatalf("path = %q", src.Path())
	}
}

func TestEnvFileSourceKind(t *testing.T) {
	src := NewEnvFileSource("/path")
	if src.Kind() != "envfile" {
		t.Fatalf("kind = %q", src.Kind())
	}
	if src.Path() != "/path" {
		t.Fatalf("path = %q", src.Path())
	}
}

func TestEnvSourceKind(t *testing.T) {
	src := NewEnvSource("PFX_", []string{"KEY"})
	if src.Kind() != "env" {
		t.Fatalf("kind = %q", src.Kind())
	}
}
