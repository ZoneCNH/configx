package configx_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ZoneCNH/configx/pkg/configx"
	"github.com/ZoneCNH/configx/testkit"
)

// TestSecretLeakGolden verifies that Sanitize replaces all secret values
// with the redaction marker and that non-secret fields are preserved.
func TestSecretLeakGolden(t *testing.T) {
	tests := []struct {
		name       string
		keys       map[string]string
		secretKeys []string
		wantRedact []string
		wantClear  []string
	}{
		{
			name: "password and token fields are redacted",
			keys: map[string]string{
				"APP_NAME":    "myapp",
				"DB_PASSWORD": "s3cret-db-pass",
				"API_KEY":     "ak-12345",
				"APP_DEBUG":   "false",
			},
			secretKeys: []string{"DB_PASSWORD", "API_KEY"},
			wantRedact: []string{"DB_PASSWORD", "API_KEY"},
			wantClear:  []string{"APP_NAME", "APP_DEBUG"},
		},
		{
			name: "auto-detected secret key names are redacted",
			keys: map[string]string{
				"JWT_SECRET":    "jwt-abc",
				"ACCESS_TOKEN":  "tok-xyz",
				"SECRET_CONFIG": "internal",
				"APP_HOST":      "localhost",
			},
			secretKeys: nil,
			wantRedact: []string{"JWT_SECRET", "ACCESS_TOKEN", "SECRET_CONFIG"},
			wantClear:  []string{"APP_HOST"},
		},
		{
			name: "empty secrets produce no redaction",
			keys: map[string]string{
				"APP_NAME": "clean",
				"DB_PORT":  "5432",
			},
			secretKeys: nil,
			wantRedact: nil,
			wantClear:  []string{"APP_NAME", "DB_PORT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := configx.NewSecretMapSource("defaults", tt.keys, tt.secretKeys)
			result, err := configx.NewLoader().AddSource(source).Load(context.Background())
			if err != nil {
				t.Fatalf("load: %v", err)
			}

			sanitized := result.Sanitize()

			for _, key := range tt.wantRedact {
				sv, ok := sanitized.Values[key]
				if !ok {
					t.Fatalf("expected key %q in sanitized result", key)
				}
				if sv.Value != "***" {
					t.Errorf("key %q: expected redacted value, got %q", key, sv.Value)
				}
				if !sv.Secret {
					t.Errorf("key %q: expected Secret=true", key)
				}
			}

			for _, key := range tt.wantClear {
				sv, ok := sanitized.Values[key]
				if !ok {
					t.Fatalf("expected key %q in sanitized result", key)
				}
				if sv.Value == "***" {
					t.Errorf("key %q: non-secret value was incorrectly redacted", key)
				}
				if sv.Value != tt.keys[key] {
					t.Errorf("key %q: expected %q, got %q", key, tt.keys[key], sv.Value)
				}
				if sv.Secret {
					t.Errorf("key %q: expected Secret=false", key)
				}
			}

			payload, err := json.Marshal(sanitized)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			jsonStr := string(payload)
			for _, key := range tt.wantRedact {
				original := tt.keys[key]
				if original != "" && strings.Contains(jsonStr, original) {
					t.Errorf("JSON output leaked secret value for key %q", key)
				}
			}
		})
	}
}

// TestSecretLeakGoldenOutput compares the sanitized Values map against a golden file.
// We compare only Values (not Sources) because Sources contains non-deterministic timestamps.
func TestSecretLeakGoldenOutput(t *testing.T) {
	result, err := configx.NewLoader().
		AddSource(configx.NewSecretMapSource("defaults", map[string]string{
			"APP_NAME":      "myapp",
			"APP_HOST":      "localhost",
			"APP_DEBUG":     "false",
			"JWT_SECRET":    "jwt-abc-123",
			"SECRET_CONFIG": "internal-val",
		}, []string{"JWT_SECRET"})).
		AddSource(configx.NewSecretMapSource("override", map[string]string{
			"APP_PORT":     "9090",
			"DB_PASSWORD":  "db-p@ss",
			"DB_USER":      "admin",
			"API_KEY":      "ak-live-key",
			"ACCESS_TOKEN": "at-live-tok",
		}, []string{"DB_PASSWORD", "API_KEY"})).
		AddSource(configx.NewSecretMapSource("env", map[string]string{
			"ACCESS_TOKEN": "at-env-tok",
		}, []string{"ACCESS_TOKEN"})).
		Load(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	payload, err := json.MarshalIndent(result.Sanitize().Values, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	payload = append(payload, '\n')

	testkit.RequireGolden(t, "testdata/golden/secret_leak_sanitized.json", payload)
}
