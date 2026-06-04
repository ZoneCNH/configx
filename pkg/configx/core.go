package configx

import (
	"context"

	foundationx "github.com/ZoneCNH/foundationx"
)

const redactionMarker = "***"

// SecretString is an alias for foundationx.SecretString.
type SecretString = foundationx.SecretString

// NewSecretString creates a new SecretString.
func NewSecretString(value string) SecretString { return foundationx.NewSecretString(value) }

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
