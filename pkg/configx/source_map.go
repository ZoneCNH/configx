package configx

import (
	"context"
	"time"
)

// MapSource loads configuration from an in-memory map.
type MapSource struct {
	name    string
	values  map[string]string
	secrets map[string]bool
}

// NewMapSource creates a MapSource from a string map.
func NewMapSource(name string, values map[string]string) *MapSource {
	return NewSecretMapSource(name, values, nil)
}

// NewSecretMapSource creates a MapSource with explicit secret key marking.
func NewSecretMapSource(name string, values map[string]string, secretKeys []string) *MapSource {
	secrets := map[string]bool{}
	for _, k := range secretKeys {
		secrets[k] = true
	}
	copied := map[string]string{}
	for k, v := range values {
		copied[k] = v
	}
	if name == "" {
		name = "map"
	}
	return &MapSource{name: name, values: copied, secrets: secrets}
}

// Name returns the source name.
func (s *MapSource) Name() string { return s.name }

// Kind returns "map".
func (s *MapSource) Kind() string { return "map" }

// Load loads configuration from the in-memory map.
func (s *MapSource) Load(ctx context.Context) (Map, error) {
	if ctx == nil {
		return nil, validationError("configx.MapSource.Load", "context is required", nil)
	}
	if err := ctx.Err(); err != nil {
		return nil, contextError("configx.MapSource.Load", err)
	}
	out := Map{}
	now := time.Now().UTC()
	for k, v := range s.values {
		out[k] = Value{Key: k, Value: v, Secret: s.secrets[k] || IsSecretKey(k), Source: s.name, LoadedAt: now}
	}
	return out, nil
}
