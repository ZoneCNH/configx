package configx

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// EnvFileSource loads configuration from a .env file.
type EnvFileSource struct{ name, path string }

// NewEnvFileSource creates an EnvFileSource for the given path.
func NewEnvFileSource(path string, opts ...SourceOption) *EnvFileSource {
	o := sourceOptions{name: "envfile"}
	for _, opt := range opts {
		opt(&o)
	}
	return &EnvFileSource{name: o.name, path: path}
}

// Name returns the source name.
func (s *EnvFileSource) Name() string { return s.name }

// Kind returns "envfile".
func (s *EnvFileSource) Kind() string { return "envfile" }

// Path returns the file path.
func (s *EnvFileSource) Path() string { return s.path }

// Load loads configuration from the .env file.
func (s *EnvFileSource) Load(ctx context.Context) (Map, error) {
	const op = "configx.EnvFileSource.Load"
	if ctx == nil {
		return nil, validationError(op, "context is required", nil)
	}
	if s.path == "" {
		return nil, validationError(op, "path is required", nil)
	}
	f, err := os.Open(s.path)
	if err != nil {
		return nil, WrapError(ErrorKindConfig, op, "read env file failed", false, err)
	}
	defer func() { _ = f.Close() }()
	out := Map{}
	now := time.Now().UTC()
	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		if err := ctx.Err(); err != nil {
			return nil, contextError(op, err)
		}
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		key, val, ok := strings.Cut(raw, "=")
		if !ok {
			return nil, validationError(op, fmt.Sprintf("invalid env file line %d", line), nil)
		}
		key = strings.TrimSpace(strings.TrimPrefix(key, "export "))
		val = strings.TrimSpace(val)
		val = strings.Trim(val, "'\"")
		out[key] = Value{Key: key, Value: val, Secret: IsSecretKey(key), Source: s.name, LoadedAt: now}
	}
	if err := scanner.Err(); err != nil {
		return nil, WrapError(ErrorKindConfig, op, "scan env file failed", false, err)
	}
	return out, nil
}

// JSONFileSource loads configuration from a JSON file.
type JSONFileSource struct{ name, path string }

// NewJSONFileSource creates a JSONFileSource for the given path.
func NewJSONFileSource(path string, opts ...SourceOption) *JSONFileSource {
	o := sourceOptions{name: "json"}
	for _, opt := range opts {
		opt(&o)
	}
	return &JSONFileSource{name: o.name, path: path}
}

// Name returns the source name.
func (s *JSONFileSource) Name() string { return s.name }

// Kind returns "json".
func (s *JSONFileSource) Kind() string { return "json" }

// Path returns the file path.
func (s *JSONFileSource) Path() string { return s.path }

// Load loads configuration from the JSON file.
func (s *JSONFileSource) Load(ctx context.Context) (Map, error) {
	const op = "configx.JSONFileSource.Load"
	if ctx == nil {
		return nil, validationError(op, "context is required", nil)
	}
	if s.path == "" {
		return nil, validationError(op, "path is required", nil)
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		return nil, WrapError(ErrorKindConfig, op, "read json file failed", false, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, contextError(op, err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, WrapError(ErrorKindConfig, op, "parse json file failed", false, err)
	}
	return flattenMap(raw, s.name), nil
}

// flattenMap recursively flattens a nested map into dot-separated keys.
func flattenMap(raw map[string]any, source string) Map {
	out := Map{}
	now := time.Now().UTC()
	var walk func(string, any)
	walk = func(prefix string, v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, vv := range x {
				key := k
				if prefix != "" {
					key = prefix + "." + k
				}
				walk(key, vv)
			}
		case map[any]any:
			for k, vv := range x {
				key := fmt.Sprint(k)
				if prefix != "" {
					key = prefix + "." + key
				}
				walk(key, vv)
			}
		default:
			out[prefix] = Value{Key: prefix, Value: fmt.Sprint(x), Secret: IsSecretKey(prefix), Source: source, LoadedAt: now}
		}
	}
	for k, v := range raw {
		walk(k, v)
	}
	return out
}
