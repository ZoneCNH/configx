package configx

import "strings"

// DefaultSecretPatterns contains the default patterns for identifying secret keys.
// A key is considered secret if its lowercased form contains any of these substrings.
var DefaultSecretPatterns = []string{
	"secret",
	"password",
	"passwd",
	"token",
	"_key",
	"credential",
	"auth",
}

// SecretPolicy defines rules for identifying and sanitizing secret values
// in configuration maps. It can be used to uniformly detect secret keys
// across all configuration sources.
type SecretPolicy struct {
	// Patterns is a list of substrings (lowercased) that indicate a key is secret.
	// If empty, DefaultSecretPatterns is used.
	Patterns []string

	// CustomMatcher is an optional function for additional secret detection logic.
	// It receives the original key (not lowercased) and returns true if the key
	// is considered a secret. This runs in addition to pattern matching.
	CustomMatcher func(key string) bool
}

// DefaultSecretPolicy returns a SecretPolicy with the default patterns.
func DefaultSecretPolicy() *SecretPolicy {
	return &SecretPolicy{Patterns: DefaultSecretPatterns}
}

// IsSecret returns true if the given key matches any of the policy's patterns
// or the custom matcher.
func (sp *SecretPolicy) IsSecret(key string) bool {
	if sp == nil {
		return false
	}

	patterns := sp.Patterns
	if len(patterns) == 0 {
		patterns = DefaultSecretPatterns
	}

	lower := strings.ToLower(key)
	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}

	if sp.CustomMatcher != nil && sp.CustomMatcher(key) {
		return true
	}

	return false
}

// SanitizeMap returns a copy of the map with all secret values replaced
// by the redaction marker ("***"). Non-secret keys are preserved as-is.
func (sp *SecretPolicy) SanitizeMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}

	out := make(map[string]any, len(m))
	for k, v := range m {
		if sp.IsSecret(k) {
			out[k] = redactionMarker
		} else {
			// Recursively sanitize nested maps.
			if nested, ok := v.(map[string]any); ok {
				out[k] = sp.SanitizeMap(nested)
			} else {
				out[k] = v
			}
		}
	}
	return out
}

// SanitizeStringMap is a convenience method that works on map[string]string.
func (sp *SecretPolicy) SanitizeStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}

	out := make(map[string]string, len(m))
	for k, v := range m {
		if sp.IsSecret(k) {
			out[k] = redactionMarker
		} else {
			out[k] = v
		}
	}
	return out
}

// SecretKeys returns a list of keys from the given map that are identified as secrets.
func (sp *SecretPolicy) SecretKeys(m map[string]any) []string {
	var keys []string
	for k := range m {
		if sp.IsSecret(k) {
			keys = append(keys, k)
		}
	}
	return keys
}
