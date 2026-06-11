package configx

import (
	"context"
	"encoding/json"
)

const redactionMarker = "***"

// SecretString stores a secret and masks it by default when formatted.
type SecretString string

// NewSecretString creates a new SecretString.
func NewSecretString(value string) SecretString { return SecretString(value) }

func (s SecretString) String() string {
	if s == "" {
		return ""
	}
	return redactionMarker
}

// Reveal returns the underlying secret value.
func (s SecretString) Reveal() string { return string(s) }

// Sanitize masks the secret for safe logging.
func (s SecretString) Sanitize() any { return s.String() }

// IsZero reports whether the secret is empty.
func (s SecretString) IsZero() bool { return s == "" }

// GoString masks the secret in %#v output.
func (s SecretString) GoString() string { return s.String() }

// MarshalText implements encoding.TextMarshaler with redaction.
func (s SecretString) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

// MarshalJSON implements json.Marshaler with redaction.
func (s SecretString) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }

// LoadEnv is a convenience function that loads configuration from environment variables.
func LoadEnv(ctx context.Context, prefix string, keys []string) (LoadResult, error) {
	return NewLoader().AddSource(NewEnvSource(prefix, keys)).Load(ctx)
}

// LoadAllEnv is a convenience function that loads all environment variables with the given prefix.
func LoadAllEnv(ctx context.Context, prefix string) (LoadResult, error) {
	return NewLoader().AddSource(NewAllEnvSource(prefix)).Load(ctx)
}

// LoadEnvFile is a convenience function that loads configuration from a .env file.
func LoadEnvFile(ctx context.Context, path string) (LoadResult, error) {
	return NewLoader().AddSource(NewEnvFileSource(path)).Load(ctx)
}

// LoadJSONFile is a convenience function that loads configuration from a JSON file.
func LoadJSONFile(ctx context.Context, path string) (LoadResult, error) {
	return NewLoader().AddSource(NewJSONFileSource(path)).Load(ctx)
}

// LoadMap is a convenience function that loads configuration from a string map.
func LoadMap(ctx context.Context, name string, values map[string]string) (LoadResult, error) {
	return NewLoader().AddSource(NewMapSource(name, values)).Load(ctx)
}
