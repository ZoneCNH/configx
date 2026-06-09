package configx

import (
	"testing"
)

func TestDefaultSecretPolicy(t *testing.T) {
	sp := DefaultSecretPolicy()
	if sp == nil {
		t.Fatal("expected non-nil policy")
	}
	if len(sp.Patterns) != len(DefaultSecretPatterns) {
		t.Fatalf("expected %d patterns, got %d", len(DefaultSecretPatterns), len(sp.Patterns))
	}
}

func TestSecretPolicyIsSecret(t *testing.T) {
	tests := []struct {
		name string
		sp   *SecretPolicy
		key  string
		want bool
	}{
		{"nil policy returns false", nil, "api_token", false},
		{"default policy detects token", DefaultSecretPolicy(), "api_token", true},
		{"default policy detects password", DefaultSecretPolicy(), "DB_PASSWORD", true},
		{"default policy detects secret", DefaultSecretPolicy(), "JWT_SECRET", true},
		{"default policy detects auth", DefaultSecretPolicy(), "AUTH_KEY", true},
		{"default policy detects credential", DefaultSecretPolicy(), "MY_CREDENTIAL", true},
		{"default policy detects _key suffix", DefaultSecretPolicy(), "access_key", true},
		{"default policy detects passwd", DefaultSecretPolicy(), "passwd", true},
		{"default policy allows host", DefaultSecretPolicy(), "HOST", false},
		{"default policy allows port", DefaultSecretPolicy(), "PORT", false},
		{"empty patterns falls back to default", &SecretPolicy{}, "api_token", true},
		{"custom matcher detects", &SecretPolicy{CustomMatcher: func(key string) bool { return key == "MY_CUSTOM" }}, "MY_CUSTOM", true},
		{"custom matcher rejects", &SecretPolicy{CustomMatcher: func(key string) bool { return key == "MY_CUSTOM" }}, "HOST", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sp.IsSecret(tt.key)
			if got != tt.want {
				t.Fatalf("IsSecret(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestSecretPolicySanitizeMap(t *testing.T) {
	sp := DefaultSecretPolicy()

	t.Run("nil map returns nil", func(t *testing.T) {
		got := sp.SanitizeMap(nil)
		if got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("redacts secret keys", func(t *testing.T) {
		m := map[string]any{
			"host":   "localhost",
			"token":  "secret-value",
			"nested": map[string]any{"password": "p@ss", "name": "app"},
		}
		got := sp.SanitizeMap(m)
		if got["host"] != "localhost" {
			t.Fatalf("host = %v, want localhost", got["host"])
		}
		if got["token"] != redactionMarker {
			t.Fatalf("token = %v, want %v", got["token"], redactionMarker)
		}
		nested, ok := got["nested"].(map[string]any)
		if !ok {
			t.Fatalf("nested type = %T", got["nested"])
		}
		if nested["password"] != redactionMarker {
			t.Fatalf("nested.password = %v, want %v", nested["password"], redactionMarker)
		}
		if nested["name"] != "app" {
			t.Fatalf("nested.name = %v", nested["name"])
		}
	})
}

func TestSecretPolicySanitizeStringMap(t *testing.T) {
	sp := DefaultSecretPolicy()

	t.Run("nil map returns nil", func(t *testing.T) {
		got := sp.SanitizeStringMap(nil)
		if got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("redacts secret keys", func(t *testing.T) {
		m := map[string]string{
			"HOST":       "localhost",
			"API_TOKEN":  "tok123",
			"DB_PASSWORD": "p4ss",
		}
		got := sp.SanitizeStringMap(m)
		if got["HOST"] != "localhost" {
			t.Fatalf("HOST = %v", got["HOST"])
		}
		if got["API_TOKEN"] != redactionMarker {
			t.Fatalf("API_TOKEN = %v", got["API_TOKEN"])
		}
		if got["DB_PASSWORD"] != redactionMarker {
			t.Fatalf("DB_PASSWORD = %v", got["DB_PASSWORD"])
		}
	})
}

func TestSecretPolicySecretKeys(t *testing.T) {
	sp := DefaultSecretPolicy()
	m := map[string]any{
		"host":      "localhost",
		"api_token": "tok",
		"password":  "p",
		"name":      "app",
	}
	keys := sp.SecretKeys(m)
	secretSet := make(map[string]bool)
	for _, k := range keys {
		secretSet[k] = true
	}
	if !secretSet["api_token"] {
		t.Fatal("expected api_token in secret keys")
	}
	if !secretSet["password"] {
		t.Fatal("expected password in secret keys")
	}
	if secretSet["host"] {
		t.Fatal("host should not be a secret key")
	}
	if secretSet["name"] {
		t.Fatal("name should not be a secret key")
	}
}
