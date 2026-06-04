package configx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- Loader edge cases ---

func TestLoaderLoadRejectsNilContext(t *testing.T) {
	loader := NewLoader().AddSource(NewMapSource("m", map[string]string{"k": "v"}))
	_, err := loader.Load(nil)
	if err == nil {
		t.Fatal("expected nil context error")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestLoaderLoadRejectsNilLoader(t *testing.T) {
	var loader *Loader
	_, err := loader.Load(context.Background())
	if err == nil {
		t.Fatal("expected nil loader error")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestLoaderAddSourceOnNilLoaderIsNoop(t *testing.T) {
	var loader *Loader
	got := loader.AddSource(NewMapSource("m", nil))
	if got != nil {
		t.Fatal("expected nil return from nil receiver AddSource")
	}
}

func TestLoaderLoadSkipsNilSourceWhenNotFailFast(t *testing.T) {
	loader := NewLoader(WithFailFast(false)).
		AddSource(nil).
		AddSource(NewMapSource("ok", map[string]string{"k": "v"}))
	result, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("expected no error with failFast=false, got %v", err)
	}
	if got, ok := result.Get("k"); !ok || got != "v" {
		t.Fatalf("expected k=v, got (%q,%v)", got, ok)
	}
	if len(result.Sources) != 2 {
		t.Fatalf("expected 2 source reports, got %d", len(result.Sources))
	}
	if result.Sources[0].Error != "source is nil" {
		t.Fatalf("expected nil source error in report, got %q", result.Sources[0].Error)
	}
}

func TestLoaderLoadFailsFastOnNilSource(t *testing.T) {
	loader := NewLoader(WithFailFast(true)).
		AddSource(nil).
		AddSource(NewMapSource("ok", map[string]string{"k": "v"}))
	_, err := loader.Load(context.Background())
	if err == nil {
		t.Fatal("expected failFast error on nil source")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestLoaderLoadSkipsSourceErrorWhenNotFailFast(t *testing.T) {
	loader := NewLoader(WithFailFast(false)).
		AddSource(NewEnvFileSource("/nonexistent/path/file.env")).
		AddSource(NewMapSource("ok", map[string]string{"k": "v"}))
	result, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("expected no error with failFast=false, got %v", err)
	}
	if got, ok := result.Get("k"); !ok || got != "v" {
		t.Fatalf("expected k=v, got (%q,%v)", got, ok)
	}
	if len(result.Sources) != 2 {
		t.Fatalf("expected 2 source reports, got %d", len(result.Sources))
	}
	if result.Sources[0].Error == "" {
		t.Fatal("expected error in first source report")
	}
	if result.Sources[0].Loaded {
		t.Fatal("first source should not be marked loaded")
	}
}

func TestLoaderLoadFailsFastOnSourceError(t *testing.T) {
	loader := NewLoader(WithFailFast(true)).
		AddSource(NewEnvFileSource("/nonexistent/path/file.env"))
	_, err := loader.Load(context.Background())
	if err == nil {
		t.Fatal("expected failFast error on source failure")
	}
	if !IsKind(err, ErrorKindConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestLoaderLoadCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	loader := NewLoader().AddSource(NewMapSource("m", map[string]string{"k": "v"}))
	_, err := loader.Load(ctx)
	if err == nil {
		t.Fatal("expected canceled context error")
	}
	if !IsKind(err, ErrorKindUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func TestWithFailFastOption(t *testing.T) {
	loader := NewLoader(WithFailFast(false))
	if loader.options.failFast {
		t.Fatal("expected failFast=false")
	}
	loader2 := NewLoader(WithFailFast(true))
	if !loader2.options.failFast {
		t.Fatal("expected failFast=true")
	}
}

func TestLoaderConcurrentAddSourceAndLoad(t *testing.T) {
	loader := NewLoader().AddSource(NewMapSource("base", map[string]string{"base": "value"}))
	ctx := context.Background()

	var wg sync.WaitGroup
	errs := make(chan error, 50)

	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			loader.AddSource(NewMapSource(
				fmt.Sprintf("dynamic-%d", i),
				map[string]string{fmt.Sprintf("key-%d", i): fmt.Sprintf("value-%d", i)},
			))
		}(i)
	}

	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := loader.Load(ctx)
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("load: %v", err)
		}
	}
}

// --- EnvSource edge cases ---

func TestEnvSourceLoadNilContext(t *testing.T) {
	src := NewEnvSource("PFX_", []string{"KEY"})
	_, err := src.Load(nil)
	if err == nil {
		t.Fatal("expected nil context error")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestEnvSourceLoadCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	src := NewEnvSource("PFX_", []string{"KEY"})
	_, err := src.Load(ctx)
	if err == nil {
		t.Fatal("expected canceled context error")
	}
	if !IsKind(err, ErrorKindUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func TestAllEnvSourceLoadsAllWithPrefix(t *testing.T) {
	t.Setenv("TESTCFG_HOST", "localhost")
	t.Setenv("TESTCFG_PORT", "8080")
	t.Setenv("OTHER_KEY", "ignored")

	src := NewAllEnvSource("TESTCFG_")
	if src.Name() != "env" {
		t.Fatalf("name=%q", src.Name())
	}
	if src.Kind() != "env" {
		t.Fatalf("kind=%q", src.Kind())
	}
	result, err := src.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := result["HOST"]; !ok || got.Value != "localhost" {
		t.Fatalf("HOST=(%q,%v)", got.Value, ok)
	}
	if got, ok := result["PORT"]; !ok || got.Value != "8080" {
		t.Fatalf("PORT=(%q,%v)", got.Value, ok)
	}
	if _, ok := result["OTHER_KEY"]; ok {
		t.Fatal("expected OTHER_KEY to be filtered out")
	}
}

func TestAllEnvSourceNoPrefixLoadsAll(t *testing.T) {
	t.Setenv("TESTALL_XYZ", "123")
	src := NewAllEnvSource("")
	result, err := src.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result["TESTALL_XYZ"]; !ok {
		t.Fatal("expected TESTALL_XYZ to be present")
	}
}

func TestEnvSourceWithSourceNameOverride(t *testing.T) {
	src := NewEnvSource("PFX_", []string{"K"}, WithSourceName("custom"))
	if src.Name() != "custom" {
		t.Fatalf("name=%q", src.Name())
	}
}

// --- MapSource edge cases ---

func TestMapSourceLoadNilContext(t *testing.T) {
	src := NewMapSource("m", map[string]string{"k": "v"})
	_, err := src.Load(nil)
	if err == nil {
		t.Fatal("expected nil context error")
	}
}

func TestMapSourceLoadCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	src := NewMapSource("m", map[string]string{"k": "v"})
	_, err := src.Load(ctx)
	if err == nil {
		t.Fatal("expected canceled context error")
	}
}

func TestMapSourceDefaultNameWhenEmpty(t *testing.T) {
	src := NewMapSource("", nil)
	if src.Name() != "map" {
		t.Fatalf("expected default name 'map', got %q", src.Name())
	}
}

func TestMapSourceSecretKeysMarked(t *testing.T) {
	src := NewSecretMapSource("s", map[string]string{"api_token": "abc", "host": "localhost"}, []string{"api_token"})
	result, err := src.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result["api_token"].Secret {
		t.Fatal("api_token should be secret")
	}
	if result["host"].Secret {
		t.Fatal("host should not be secret")
	}
}

// --- EnvFileSource edge cases ---

func TestEnvFileSourceLoadNilContext(t *testing.T) {
	src := NewEnvFileSource("/some/path")
	_, err := src.Load(nil)
	if err == nil {
		t.Fatal("expected nil context error")
	}
}

func TestEnvFileSourceLoadEmptyPath(t *testing.T) {
	src := NewEnvFileSource("")
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected empty path error")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestEnvFileSourceLoadFileNotFound(t *testing.T) {
	src := NewEnvFileSource("/nonexistent/path/.env")
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected file not found error")
	}
	if !IsKind(err, ErrorKindConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestEnvFileSourceLoadInvalidLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("VALID_KEY=ok\nno_equals_sign\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := NewEnvFileSource(path)
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected invalid line error")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestEnvFileSourceLoadSkipsCommentsAndBlanks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "# comment\n\n  \nKEY=value\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	src := NewEnvFileSource(path)
	result, err := src.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := result["KEY"]; !ok || got.Value != "value" {
		t.Fatalf("KEY=(%q,%v)", got.Value, ok)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 value, got %d", len(result))
	}
}

func TestEnvFileSourceLoadExportPrefix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("export MY_KEY=my_value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := NewEnvFileSource(path)
	result, err := src.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := result["MY_KEY"]; !ok || got.Value != "my_value" {
		t.Fatalf("MY_KEY=(%q,%v)", got.Value, ok)
	}
}

func TestEnvFileSourceLoadStripsQuotes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("SQ='single'\nDQ=\"double\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := NewEnvFileSource(path)
	result, err := src.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := result["SQ"]; !ok || got.Value != "single" {
		t.Fatalf("SQ=%q", got.Value)
	}
	if got, ok := result["DQ"]; !ok || got.Value != "double" {
		t.Fatalf("DQ=%q", got.Value)
	}
}

func TestEnvFileSourceCanceledContextDuringScan(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("KEY=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	src := NewEnvFileSource(path)
	_, err := src.Load(ctx)
	if err == nil {
		t.Fatal("expected canceled context error")
	}
}

func TestEnvFileSourceWithSourceNameOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("K=V\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := NewEnvFileSource(path, WithSourceName("custom-envfile"))
	if src.Name() != "custom-envfile" {
		t.Fatalf("name=%q", src.Name())
	}
}

// --- JSONFileSource edge cases ---

func TestJSONFileSourceLoadNilContext(t *testing.T) {
	src := NewJSONFileSource("/some/path")
	_, err := src.Load(nil)
	if err == nil {
		t.Fatal("expected nil context error")
	}
}

func TestJSONFileSourceLoadEmptyPath(t *testing.T) {
	src := NewJSONFileSource("")
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected empty path error")
	}
}

func TestJSONFileSourceLoadFileNotFound(t *testing.T) {
	src := NewJSONFileSource("/nonexistent/path.json")
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected file not found error")
	}
	if !IsKind(err, ErrorKindConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestJSONFileSourceLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{invalid json!!"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := NewJSONFileSource(path)
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !IsKind(err, ErrorKindConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestJSONFileSourceCanceledContext(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.json")
	if err := os.WriteFile(path, []byte(`{"k":"v"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	src := NewJSONFileSource(path)
	_, err := src.Load(ctx)
	if err == nil {
		t.Fatal("expected canceled context error")
	}
}

func TestJSONFileSourceWithSourceNameOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.json")
	if err := os.WriteFile(path, []byte(`{"k":"v"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	src := NewJSONFileSource(path, WithSourceName("custom-json"))
	if src.Name() != "custom-json" {
		t.Fatalf("name=%q", src.Name())
	}
}

// --- flattenMap edge cases ---

func TestFlattenMapHandlesMapAnyAny(t *testing.T) {
	raw := map[string]any{
		"nested": map[any]any{
			"key": "value",
		},
	}
	out := flattenMap(raw, "test")
	if got, ok := out["nested.key"]; !ok || got.Value != "value" {
		t.Fatalf("nested.key=(%q,%v)", got.Value, ok)
	}
}

func TestFlattenMapHandlesDeeplyNested(t *testing.T) {
	raw := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "deep",
			},
		},
	}
	out := flattenMap(raw, "test")
	if got, ok := out["a.b.c"]; !ok || got.Value != "deep" {
		t.Fatalf("a.b.c=(%q,%v)", got.Value, ok)
	}
}

// --- TOML/YAML file source edge cases ---

func TestTOMLFileSourceLoadNilContext(t *testing.T) {
	src := NewTOMLFileSource("/some/path")
	_, err := src.Load(nil)
	if err == nil {
		t.Fatal("expected nil context error")
	}
}

func TestTOMLFileSourceLoadEmptyPath(t *testing.T) {
	src := NewTOMLFileSource("")
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected empty path error")
	}
}

func TestTOMLFileSourceLoadFileNotFound(t *testing.T) {
	src := NewTOMLFileSource("/nonexistent/path.toml")
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected file not found error")
	}
	if !IsKind(err, ErrorKindConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestTOMLFileSourceLoadInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("{{{{invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := NewTOMLFileSource(path)
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !IsKind(err, ErrorKindConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestYAMLFileSourceLoadNilContext(t *testing.T) {
	src := NewYAMLFileSource("/some/path")
	_, err := src.Load(nil)
	if err == nil {
		t.Fatal("expected nil context error")
	}
}

func TestYAMLFileSourceLoadEmptyPath(t *testing.T) {
	src := NewYAMLFileSource("")
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected empty path error")
	}
}

func TestYAMLFileSourceLoadFileNotFound(t *testing.T) {
	src := NewYAMLFileSource("/nonexistent/path.yaml")
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected file not found error")
	}
	if !IsKind(err, ErrorKindConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestYAMLFileSourceLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(":\n  :\n    - ][invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := NewYAMLFileSource(path)
	_, err := src.Load(context.Background())
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !IsKind(err, ErrorKindConfig) {
		t.Fatalf("expected config error, got %v", err)
	}
}

// --- Decode/setField edge cases ---

type allTypesConfig struct {
	Str      string        `config:"STR"`
	Bool     bool          `config:"BOOL"`
	Int      int           `config:"INT"`
	Int8     int8          `config:"INT8"`
	Int16    int16         `config:"INT16"`
	Int32    int32         `config:"INT32"`
	Int64    int64         `config:"INT64"`
	Uint     uint          `config:"UINT"`
	Uint8    uint8         `config:"UINT8"`
	Uint16   uint16        `config:"UINT16"`
	Uint32   uint32        `config:"UINT32"`
	Uint64   uint64        `config:"UINT64"`
	Float32  float32       `config:"FLOAT32"`
	Float64  float64       `config:"FLOAT64"`
	Duration time.Duration `config:"DURATION"`
	Secret   SecretString  `config:"SECRET" secret:"true"`
}

func TestDecodeAllSupportedTypes(t *testing.T) {
	result := LoadResult{Values: Map{
		"STR":      {Key: "STR", Value: "hello"},
		"BOOL":     {Key: "BOOL", Value: "true"},
		"INT":      {Key: "INT", Value: "-42"},
		"INT8":     {Key: "INT8", Value: "127"},
		"INT16":    {Key: "INT16", Value: "32767"},
		"INT32":    {Key: "INT32", Value: "2147483647"},
		"INT64":    {Key: "INT64", Value: "9223372036854775807"},
		"UINT":     {Key: "UINT", Value: "42"},
		"UINT8":    {Key: "UINT8", Value: "255"},
		"UINT16":   {Key: "UINT16", Value: "65535"},
		"UINT32":   {Key: "UINT32", Value: "4294967295"},
		"UINT64":   {Key: "UINT64", Value: "18446744073709551615"},
		"FLOAT32":  {Key: "FLOAT32", Value: "3.14"},
		"FLOAT64":  {Key: "FLOAT64", Value: "2.718281828"},
		"DURATION": {Key: "DURATION", Value: "5s"},
		"SECRET":   {Key: "SECRET", Value: "shh", Secret: true},
	}}
	var cfg allTypesConfig
	if err := Decode(result, &cfg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if cfg.Str != "hello" {
		t.Fatalf("str=%q", cfg.Str)
	}
	if !cfg.Bool {
		t.Fatal("expected bool=true")
	}
	if cfg.Int != -42 {
		t.Fatalf("int=%d", cfg.Int)
	}
	if cfg.Uint != 42 {
		t.Fatalf("uint=%d", cfg.Uint)
	}
	if cfg.Float64 != 2.718281828 {
		t.Fatalf("float64=%f", cfg.Float64)
	}
	if cfg.Duration != 5*time.Second {
		t.Fatalf("duration=%v", cfg.Duration)
	}
	if cfg.Secret.Reveal() != "shh" {
		t.Fatalf("secret=%q", cfg.Secret.Reveal())
	}
}

func TestDecodeDurationAsIntegerNanoseconds(t *testing.T) {
	type durCfg struct {
		D time.Duration `config:"D"`
	}
	result := LoadResult{Values: Map{"D": {Key: "D", Value: "1000000000"}}}
	var cfg durCfg
	if err := Decode(result, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.D != time.Second {
		t.Fatalf("duration=%v", cfg.D)
	}
}

func TestDecodeInvalidBoolReturnsError(t *testing.T) {
	type boolCfg struct {
		B bool `config:"B"`
	}
	result := LoadResult{Values: Map{"B": {Key: "B", Value: "notabool"}}}
	var cfg boolCfg
	err := Decode(result, &cfg)
	if err == nil {
		t.Fatal("expected invalid bool error")
	}
}

func TestDecodeInvalidIntReturnsError(t *testing.T) {
	type intCfg struct {
		I int `config:"I"`
	}
	result := LoadResult{Values: Map{"I": {Key: "I", Value: "notanint"}}}
	var cfg intCfg
	err := Decode(result, &cfg)
	if err == nil {
		t.Fatal("expected invalid int error")
	}
}

func TestDecodeInvalidUintReturnsError(t *testing.T) {
	type uintCfg struct {
		U uint `config:"U"`
	}
	result := LoadResult{Values: Map{"U": {Key: "U", Value: "notanuint"}}}
	var cfg uintCfg
	err := Decode(result, &cfg)
	if err == nil {
		t.Fatal("expected invalid uint error")
	}
}

func TestDecodeInvalidFloatReturnsError(t *testing.T) {
	type floatCfg struct {
		F float64 `config:"F"`
	}
	result := LoadResult{Values: Map{"F": {Key: "F", Value: "notafloat"}}}
	var cfg floatCfg
	err := Decode(result, &cfg)
	if err == nil {
		t.Fatal("expected invalid float error")
	}
}

func TestDecodeInvalidDurationReturnsError(t *testing.T) {
	type durCfg struct {
		D time.Duration `config:"D"`
	}
	result := LoadResult{Values: Map{"D": {Key: "D", Value: "not-a-duration-and-not-a-number"}}}
	var cfg durCfg
	err := Decode(result, &cfg)
	if err == nil {
		t.Fatal("expected invalid duration error")
	}
}

func TestDecodeNilTargetReturnsError(t *testing.T) {
	err := Decode(LoadResult{}, nil)
	if err == nil {
		t.Fatal("expected nil target error")
	}
}

func TestDecodeNonPointerTargetReturnsError(t *testing.T) {
	err := Decode(LoadResult{}, struct{}{})
	if err == nil {
		t.Fatal("expected non-pointer error")
	}
}

func TestDecodePointerToNonStructReturnsError(t *testing.T) {
	s := "not a struct"
	err := Decode(LoadResult{}, &s)
	if err == nil {
		t.Fatal("expected pointer-to-non-struct error")
	}
}

func TestDecodeNilPointerTargetReturnsError(t *testing.T) {
	var p *struct{}
	err := Decode(LoadResult{}, p)
	if err == nil {
		t.Fatal("expected nil pointer error")
	}
}

func TestDecodeUnsupportedFieldType(t *testing.T) {
	type badCfg struct {
		Ch chan int `config:"CH"`
	}
	result := LoadResult{Values: Map{"CH": {Key: "CH", Value: "value"}}}
	var cfg badCfg
	err := Decode(result, &cfg)
	if err == nil {
		t.Fatal("expected unsupported type error")
	}
}

func TestDecodeSkippedFieldWithTag(t *testing.T) {
	type skipCfg struct {
		SkipMe string `config:"-"`
		Keep   string `config:"KEEP"`
	}
	result := LoadResult{Values: Map{"KEEP": {Key: "KEEP", Value: "yes"}}}
	var cfg skipCfg
	if err := Decode(result, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.SkipMe != "" {
		t.Fatalf("skip field should be empty, got %q", cfg.SkipMe)
	}
	if cfg.Keep != "yes" {
		t.Fatalf("keep=%q", cfg.Keep)
	}
}

func TestDecodeConfigxTagSkipField(t *testing.T) {
	type skipCfg struct {
		S string `configx:"-"`
		K string `configx:"KEEP"`
	}
	result := LoadResult{Values: Map{"KEEP": {Key: "KEEP", Value: "yes"}}}
	var cfg skipCfg
	if err := Decode(result, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.S != "" {
		t.Fatalf("skip field should be empty, got %q", cfg.S)
	}
	if cfg.K != "yes" {
		t.Fatalf("keep=%q", cfg.K)
	}
}

func TestDecodeConfigxTagWithOptions(t *testing.T) {
	type optCfg struct {
		H string `configx:"HOST,default=localhost,required"`
	}
	result := LoadResult{Values: Map{}}
	var cfg optCfg
	if err := Decode(result, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.H != "localhost" {
		t.Fatalf("host=%q", cfg.H)
	}
}

func TestDecodeNormalizedKeyLookup(t *testing.T) {
	type normCfg struct {
		DatabaseHost string `config:"database.host"`
	}
	result := LoadResult{Values: Map{"DATABASE_HOST": {Key: "DATABASE_HOST", Value: "db.example.com"}}}
	var cfg normCfg
	if err := Decode(result, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseHost != "db.example.com" {
		t.Fatalf("databaseHost=%q", cfg.DatabaseHost)
	}
}

// --- TextUnmarshaler path ---

type textUnmarshalerField struct {
	V mockTextUnmarshaler `config:"V"`
}

type mockTextUnmarshaler struct{ val string }

func (m *mockTextUnmarshaler) UnmarshalText(text []byte) error {
	m.val = string(text)
	return nil
}

func (m mockTextUnmarshaler) String() string { return m.val }

func TestDecodeTextUnmarshalerField(t *testing.T) {
	result := LoadResult{Values: Map{"V": {Key: "V", Value: "unmarshaled"}}}
	var cfg textUnmarshalerField
	if err := Decode(result, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.V.val != "unmarshaled" {
		t.Fatalf("val=%q", cfg.V.val)
	}
}

// --- isTagOption ---

func TestIsTagOption(t *testing.T) {
	tests := []struct {
		option string
		want   bool
	}{
		{"required", true},
		{"default=val", true},
		{"secret", true},
		{"other", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isTagOption(tt.option); got != tt.want {
			t.Fatalf("isTagOption(%q)=%v, want %v", tt.option, got, tt.want)
		}
	}
}

// --- parseConfigTag edge cases ---

func TestParseConfigTagWithConfigxTagKey(t *testing.T) {
	type cfg struct {
		F string `configx:"MY_KEY"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	tag := parseConfigTag(sf)
	if tag.key != "MY_KEY" {
		t.Fatalf("key=%q", tag.key)
	}
}

func TestParseConfigTagWithConfigxTagSkip(t *testing.T) {
	type cfg struct {
		F string `configx:"-"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	tag := parseConfigTag(sf)
	if !tag.skip {
		t.Fatal("expected skip=true")
	}
}

func TestParseConfigTagConfigOverridesFieldName(t *testing.T) {
	type cfg struct {
		F string `config:"OVERRIDDEN"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	tag := parseConfigTag(sf)
	if tag.key != "OVERRIDDEN" {
		t.Fatalf("key=%q", tag.key)
	}
}

func TestParseConfigTagDefault(t *testing.T) {
	type cfg struct {
		F string `config:"F" default:"fallback"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	tag := parseConfigTag(sf)
	if tag.defaultValue != "fallback" {
		t.Fatalf("default=%q", tag.defaultValue)
	}
}

func TestParseConfigTagRequired(t *testing.T) {
	type cfg struct {
		F string `config:"F,required"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	tag := parseConfigTag(sf)
	if !tag.required {
		t.Fatal("expected required=true")
	}
}

func TestParseConfigTagNoConfigTagUsesFieldName(t *testing.T) {
	type cfg struct {
		FieldName string
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	tag := parseConfigTag(sf)
	if tag.key != "FieldName" {
		t.Fatalf("key=%q", tag.key)
	}
}

// --- Secret/sanitization edge cases ---

func TestSanitizeErrorNilReturnsNil(t *testing.T) {
	if got := sanitizeError(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestSanitizeMessageRedactsSecretKeyValue(t *testing.T) {
	msg := "error with password=supersecret and host=localhost"
	got := sanitizeMessage(msg)
	if strings.Contains(got, "supersecret") {
		t.Fatalf("secret leaked: %q", got)
	}
	if !strings.Contains(got, "host=localhost") {
		t.Fatalf("non-secret should be preserved: %q", got)
	}
}

func TestSanitizeErrorPreservesConfigxErrorSemantics(t *testing.T) {
	cause := errors.New("driver failed password=supersecret")
	err := WrapError(ErrorKindTimeout, "configx.Source.Load", "token=mytoken failed", true, cause)

	sanitized := sanitizeError(err)
	var got *Error
	if !errors.As(sanitized, &got) {
		t.Fatalf("expected sanitized *Error, got %T", sanitized)
	}
	if got.Kind != ErrorKindTimeout || got.Op != "configx.Source.Load" || !got.Retryable {
		t.Fatalf("metadata not preserved: %#v", got)
	}
	if !IsKind(sanitized, ErrorKindTimeout) {
		t.Fatalf("kind lookup failed for sanitized error: %v", sanitized)
	}
	if got.Cause == nil {
		t.Fatal("expected sanitized cause to be preserved")
	}
	if strings.Contains(got.Error(), "mytoken") || strings.Contains(got.Cause.Error(), "supersecret") {
		t.Fatalf("secret leaked in sanitized error: %q cause=%q", got.Error(), got.Cause.Error())
	}
}

func TestSanitizeErrorPreservesNestedConfigxErrorCauseSemantics(t *testing.T) {
	inner := WrapError(ErrorKindAuth, "inner", "password=innersecret", true, errors.New("token=causesecret"))
	outer := WrapError(ErrorKindConfig, "outer", "secret_key=outersecret", false, inner)

	sanitized := sanitizeError(outer)
	var outerGot *Error
	if !errors.As(sanitized, &outerGot) {
		t.Fatalf("expected outer *Error, got %T", sanitized)
	}
	var innerGot *Error
	if !errors.As(outerGot.Cause, &innerGot) {
		t.Fatalf("expected nested *Error cause, got %T", outerGot.Cause)
	}
	if innerGot.Kind != ErrorKindAuth || innerGot.Op != "inner" || !innerGot.Retryable {
		t.Fatalf("nested metadata not preserved: %#v", innerGot)
	}
	for _, leaked := range []string{"outersecret", "innersecret", "causesecret"} {
		if strings.Contains(sanitized.Error(), leaked) || strings.Contains(innerGot.Error(), leaked) || strings.Contains(innerGot.Cause.Error(), leaked) {
			t.Fatalf("secret %q leaked: outer=%q inner=%q cause=%q", leaked, sanitized.Error(), innerGot.Error(), innerGot.Cause.Error())
		}
	}
}

func TestSanitizeMessageHandlesTokenKey(t *testing.T) {
	msg := "token=mytoken123"
	got := sanitizeMessage(msg)
	if strings.Contains(got, "mytoken123") {
		t.Fatalf("token leaked: %q", got)
	}
}

func TestSanitizeMessageHandlesSecretKeyValue(t *testing.T) {
	msg := "secret_key=abcdef"
	got := sanitizeMessage(msg)
	if strings.Contains(got, "abcdef") {
		t.Fatalf("secret_key leaked: %q", got)
	}
}

func TestSanitizeMessageIgnoresEmptyValue(t *testing.T) {
	msg := "password="
	got := sanitizeMessage(msg)
	if got != msg {
		t.Fatalf("expected no change for empty value, got %q", got)
	}
}

func TestIsSecretKeyVariants(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"api_token", true},
		{"DB_PASSWORD", true},
		{"my_secret", true},
		{"passwd", true},
		{"access_key", true},
		{"secret_key", true},
		{"HOST", false},
		{"PORT", false},
		{"name", false},
	}
	for _, tt := range tests {
		if got := IsSecretKey(tt.key); got != tt.want {
			t.Fatalf("IsSecretKey(%q)=%v, want %v", tt.key, got, tt.want)
		}
	}
}

// --- Error edge cases ---

func TestErrorNilReceiver(t *testing.T) {
	var e *Error
	if got := e.Error(); got != "" {
		t.Fatalf("expected empty string for nil receiver, got %q", got)
	}
	if got := e.Unwrap(); got != nil {
		t.Fatalf("expected nil unwrap for nil receiver, got %v", got)
	}
}

func TestErrorErrorStringWithCauseButNoMessage(t *testing.T) {
	cause := errors.New("root cause")
	err := WrapError(ErrorKindConfig, "op", "", false, cause)
	got := err.Error()
	if !strings.Contains(got, "root cause") {
		t.Fatalf("expected cause in error string, got %q", got)
	}
}

func TestIsKindWithNonErrorReturnsFalse(t *testing.T) {
	if IsKind(errors.New("plain"), ErrorKindValidation) {
		t.Fatal("expected false for non-Error type")
	}
}

func TestErrorKindWithNonErrorReturnsInternal(t *testing.T) {
	got := errorKind(errors.New("plain"))
	if got != ErrorKindInternal {
		t.Fatalf("expected internal, got %q", got)
	}
}

func TestErrorKindWithErrorReturnsKind(t *testing.T) {
	err := NewError(ErrorKindTimeout, "op", "msg", true)
	got := errorKind(err)
	if got != ErrorKindTimeout {
		t.Fatalf("expected timeout, got %q", got)
	}
}

// --- mergeValue edge cases ---

func TestMergeValueUnknownStrategy(t *testing.T) {
	m := Map{"k": {Key: "k", Value: "old"}}
	err := mergeValue(m, "k", Value{Key: "k", Value: "new"}, MergeStrategy(99))
	if err == nil {
		t.Fatal("expected unknown strategy error")
	}
}

func TestMergeValueFirstWinsNoOverwrite(t *testing.T) {
	m := Map{"k": {Key: "k", Value: "first"}}
	err := mergeValue(m, "k", Value{Key: "k", Value: "second"}, FirstWins)
	if err != nil {
		t.Fatal(err)
	}
	if m["k"].Value != "first" {
		t.Fatalf("expected first value, got %q", m["k"].Value)
	}
}

func TestMergeValueLastWinsOverwrites(t *testing.T) {
	m := Map{"k": {Key: "k", Value: "first"}}
	err := mergeValue(m, "k", Value{Key: "k", Value: "second"}, LastWins)
	if err != nil {
		t.Fatal(err)
	}
	if m["k"].Value != "second" {
		t.Fatalf("expected second value, got %q", m["k"].Value)
	}
	if !m["k"].Overridden {
		t.Fatal("expected winning value to be marked overridden")
	}
}

func TestMergeValueNewKeyNoConflict(t *testing.T) {
	m := Map{}
	err := mergeValue(m, "new", Value{Key: "new", Value: "val"}, ErrorOnConflict)
	if err != nil {
		t.Fatal(err)
	}
	if m["new"].Value != "val" {
		t.Fatalf("expected val, got %q", m["new"].Value)
	}
}

// --- Convenience function edge cases ---

func TestLoadEnvConvenience(t *testing.T) {
	t.Setenv("CVT_HOST", "localhost")
	result, err := LoadEnv(context.Background(), "CVT_", []string{"HOST"})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := result.Get("HOST"); got != "localhost" {
		t.Fatalf("HOST=%q", got)
	}
}

func TestLoadAllEnvConvenience(t *testing.T) {
	t.Setenv("CVTALL_X", "1")
	result, err := LoadAllEnv(context.Background(), "CVTALL_")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result.Get("X"); !ok {
		t.Fatal("expected X")
	}
}

func TestLoadJSONFileConvenience(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.json")
	if err := os.WriteFile(path, []byte(`{"name":"test"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := LoadJSONFile(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := result.Get("name"); got != "test" {
		t.Fatalf("name=%q", got)
	}
}

func TestLoadMapConvenience(t *testing.T) {
	result, err := LoadMap(context.Background(), "test", map[string]string{"k": "v"})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := result.Get("k"); got != "v" {
		t.Fatalf("k=%q", got)
	}
}

// --- Client edge cases ---

func TestNewRejectsNilContext(t *testing.T) {
	_, err := New(nil, Config{Name: "test"})
	if err == nil {
		t.Fatal("expected nil context error")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCloseRejectsNilContext(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	err = client.Close(nil)
	if err == nil {
		t.Fatal("expected nil context error")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCloseWithExpiredContext(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	err = client.Close(ctx)
	if err == nil {
		t.Fatal("expected expired context error")
	}
	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

// --- Unexported field handling in Decode ---

func TestDecodeSkipsUnexportedFields(t *testing.T) {
	type cfg struct {
		Public  string `config:"PUBLIC"`
		private string //nolint:unused // intentionally unexported
	}
	result := LoadResult{Values: Map{"PUBLIC": {Key: "PUBLIC", Value: "visible"}}}
	var c cfg
	if err := Decode(result, &c); err != nil {
		t.Fatal(err)
	}
	if c.Public != "visible" {
		t.Fatalf("public=%q", c.Public)
	}
}

// --- LoadResult.Get edge case ---

func TestLoadResultGetMissingKey(t *testing.T) {
	result := LoadResult{Values: Map{}}
	_, ok := result.Get("missing")
	if ok {
		t.Fatal("expected missing key")
	}
}

// --- Sanitize edge case ---

func TestSanitizeNonSecretValuePreserved(t *testing.T) {
	result := LoadResult{Values: Map{
		"HOST": {Key: "HOST", Value: "localhost", Secret: false},
	}}
	sanitized := result.Sanitize()
	if sanitized.Values["HOST"].Value != "localhost" {
		t.Fatalf("expected non-secret preserved, got %q", sanitized.Values["HOST"].Value)
	}
}

// --- LoadResult.Decode convenience ---

func TestLoadResultDecodeConvenience(t *testing.T) {
	type cfg struct {
		Name string `config:"NAME"`
	}
	result := LoadResult{Values: Map{"NAME": {Key: "NAME", Value: "test"}}}
	var c cfg
	if err := result.Decode(&c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "test" {
		t.Fatalf("name=%q", c.Name)
	}
}
