package configx

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func BenchmarkLoaderLoad(b *testing.B) {
	ctx := context.Background()
	defaults := map[string]string{
		"APP_NAME":  "bench",
		"APP_PORT":  "8080",
		"APP_HOST":  "localhost",
		"APP_DEBUG": "false",
		"DB_HOST":   "localhost",
		"DB_PORT":   "5432",
		"DB_NAME":   "testdb",
		"DB_USER":   "postgres",
	}
	override := map[string]string{
		"APP_PORT":  "9090",
		"APP_DEBUG": "true",
		"DB_HOST":   "remote-host",
	}
	keys := []string{"APP_NAME", "APP_PORT", "APP_HOST", "APP_DEBUG", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		loader := NewLoader().
			AddSource(NewMapSource("defaults", defaults)).
			AddSource(NewSecretMapSource("override", override, []string{"DB_USER"})).
			AddSource(NewEnvSource("BENCH_", keys))
		_, err := loader.Load(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	type benchConfig struct {
		Name   string `config:"APP_NAME" required:"true"`
		Port   int    `config:"APP_PORT" default:"8080"`
		Debug  bool   `config:"APP_DEBUG"`
		Host   string `config:"APP_HOST" default:"localhost"`
		DBHost string `config:"DB_HOST"`
		DBPort int    `config:"DB_PORT" default:"5432"`
		DBName string `config:"DB_NAME"`
		DBUser string `config:"DB_USER"`
	}

	result := LoadResult{
		Values: Map{
			"APP_NAME":  {Key: "APP_NAME", Value: "bench"},
			"APP_PORT":  {Key: "APP_PORT", Value: "9090"},
			"APP_DEBUG": {Key: "APP_DEBUG", Value: "true"},
			"APP_HOST":  {Key: "APP_HOST", Value: "remote-host"},
			"DB_HOST":   {Key: "DB_HOST", Value: "db.example.com"},
			"DB_PORT":   {Key: "DB_PORT", Value: "5432"},
			"DB_NAME":   {Key: "DB_NAME", Value: "proddb"},
			"DB_USER":   {Key: "DB_USER", Value: "admin"},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg benchConfig
		if err := Decode(result, &cfg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMergeMaps(b *testing.B) {
	keys := []string{"key1", "key2", "key3", "key4", "key5", "key6", "key7", "key8"}
	makeMap := func(prefix string) Map {
		m := Map{}
		for _, k := range keys {
			m[k] = Value{Key: k, Value: prefix + "_" + k, Source: prefix}
		}
		return m
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		base := makeMap("base")
		override := makeMap("override")
		for key, value := range override {
			if err := mergeValue(base, key, value, LastWins); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkSecretDetection(b *testing.B) {
	keys := []string{
		"APP_NAME", "APP_PORT", "APP_HOST", "APP_DEBUG",
		"DB_PASSWORD", "API_TOKEN", "SECRET_KEY", "ACCESS_KEY",
		"DATABASE_PASSWD", "AUTH_SECRET", "OAUTH_TOKEN", "S3_SECRET_KEY",
		"REDIS_HOST", "REDIS_PORT", "CACHE_TTL", "LOG_LEVEL",
	}

	messages := []string{
		"config error: DB_PASSWORD=supersecret123 HOST=localhost",
		"connection failed: API_TOKEN=abc123def RETRY=true",
		"auth error: SECRET_KEY=mysecret ACCESS_KEY=mykey PORT=8080",
		"no secrets here: HOST=localhost PORT=8080 DEBUG=true",
	}

	b.Run("IsSecretKey", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, k := range keys {
				IsSecretKey(k)
			}
		}
	})

	b.Run("SanitizeMessage", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, msg := range messages {
				sanitizeMessage(msg)
			}
		}
	})
}

// BenchmarkLoaderLoadMultiSource benchmarks loading with many sources to
// measure scaling behavior as source count grows.
func BenchmarkLoaderLoadMultiSource(b *testing.B) {
	ctx := context.Background()
	sources := 10
	keysPerSource := 20

	keys := make([]string, keysPerSource)
	for i := range keys {
		keys[i] = fmt.Sprintf("KEY_%d", i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		loader := NewLoader()
		for s := 0; s < sources; s++ {
			vals := make(map[string]string, keysPerSource)
			for _, k := range keys {
				vals[k] = fmt.Sprintf("source%d_%s", s, strings.ToLower(k))
			}
			loader.AddSource(NewMapSource(fmt.Sprintf("source-%d", s), vals))
		}
		_, err := loader.Load(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
