package configx_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ZoneCNH/configx/pkg/configx"
	"github.com/ZoneCNH/configx/testkit"
)

// TestSourcePrecedenceGolden verifies that when the same key is set across
// multiple sources, the highest-priority source wins (default < file < env < flag).
// Provenance metadata (Source, Overridden) is checked, and the full result
// is compared against a golden file.
func TestSourcePrecedenceGolden(t *testing.T) {
	tests := []struct {
		name       string
		sources    []configx.Source
		wantValues map[string]string
		wantSource map[string]string
		wantOver   map[string]bool
	}{
		{
			name: "env overrides default",
			sources: []configx.Source{
				configx.NewMapSource("defaults", map[string]string{
					"APP_NAME": "default-name",
					"DB_PORT":  "5432",
				}),
				configx.NewMapSource("env", map[string]string{
					"APP_NAME": "env-name",
				}),
			},
			wantValues: map[string]string{
				"APP_NAME": "env-name",
				"DB_PORT":  "5432",
			},
			wantSource: map[string]string{
				"APP_NAME": "env",
				"DB_PORT":  "defaults",
			},
			wantOver: map[string]bool{
				"APP_NAME": true,
				"DB_PORT":  false,
			},
		},
		{
			name: "flag overrides env which overrides default",
			sources: []configx.Source{
				configx.NewMapSource("defaults", map[string]string{
					"APP_PORT":  "8080",
					"APP_DEBUG": "false",
					"DB_HOST":   "localhost",
				}),
				configx.NewMapSource("file", map[string]string{
					"APP_PORT": "3000",
					"DB_HOST":  "file-host",
				}),
				configx.NewMapSource("env", map[string]string{
					"APP_PORT": "9090",
					"DB_HOST":  "env-db",
				}),
				configx.NewMapSource("flags", map[string]string{
					"APP_PORT": "9999",
				}),
			},
			wantValues: map[string]string{
				"APP_PORT":  "9999",
				"APP_DEBUG": "false",
				"DB_HOST":   "env-db",
			},
			wantSource: map[string]string{
				"APP_PORT":  "flags",
				"APP_DEBUG": "defaults",
				"DB_HOST":   "env",
			},
			wantOver: map[string]bool{
				"APP_PORT":  true,
				"APP_DEBUG": false,
				"DB_HOST":   true,
			},
		},
		{
			name: "single source has no overrides",
			sources: []configx.Source{
				configx.NewMapSource("defaults", map[string]string{
					"KEY_A": "val-a",
					"KEY_B": "val-b",
				}),
			},
			wantValues: map[string]string{
				"KEY_A": "val-a",
				"KEY_B": "val-b",
			},
			wantSource: map[string]string{
				"KEY_A": "defaults",
				"KEY_B": "defaults",
			},
			wantOver: map[string]bool{
				"KEY_A": false,
				"KEY_B": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := configx.NewLoader()
			for _, src := range tt.sources {
				loader.AddSource(src)
			}
			result, err := loader.Load(context.Background())
			if err != nil {
				t.Fatalf("load: %v", err)
			}

			for key, want := range tt.wantValues {
				got, ok := result.Get(key)
				if !ok {
					t.Fatalf("key %q not found in result", key)
				}
				if got != want {
					t.Errorf("key %q: got %q, want %q", key, got, want)
				}
			}

			for key, wantSrc := range tt.wantSource {
				v, ok := result.Values[key]
				if !ok {
					t.Fatalf("key %q not in Values", key)
				}
				if v.Source != wantSrc {
					t.Errorf("key %q source: got %q, want %q", key, v.Source, wantSrc)
				}
			}

			for key, wantOver := range tt.wantOver {
				v, ok := result.Values[key]
				if !ok {
					t.Fatalf("key %q not in Values", key)
				}
				if v.Overridden != wantOver {
					t.Errorf("key %q Overridden: got %v, want %v", key, v.Overridden, wantOver)
				}
			}
		})
	}
}

// TestSourcePrecedenceGoldenOutput compares the full precedence result
// against a golden file to catch regressions in merge behavior.
func TestSourcePrecedenceGoldenOutput(t *testing.T) {
	result, err := configx.NewLoader().
		AddSource(configx.NewMapSource("defaults", map[string]string{
			"APP_NAME":  "default-app",
			"APP_HOST":  "default-host",
			"APP_PORT":  "8080",
			"APP_DEBUG": "false",
			"DB_HOST":   "default-db",
			"DB_PORT":   "5432",
			"LOG_LEVEL": "warn",
		})).
		AddSource(configx.NewMapSource("file", map[string]string{
			"APP_HOST":  "file-host",
			"APP_PORT":  "3000",
			"DB_HOST":   "file-db",
			"LOG_LEVEL": "file-level",
		})).
		AddSource(configx.NewMapSource("env", map[string]string{
			"APP_NAME": "env-name",
			"APP_PORT": "9090",
			"DB_HOST":  "env-db",
		})).
		AddSource(configx.NewMapSource("flags", map[string]string{
			"APP_PORT":  "9999",
			"APP_DEBUG": "true",
		})).
		Load(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	payload, err := json.MarshalIndent(result.Sanitize().Values, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	payload = append(payload, '\n')

	testkit.RequireGolden(t, "testdata/golden/source_precedence.json", payload)
}
