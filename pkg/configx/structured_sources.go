package configx

import (
	"context"
	"os"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func LoadTOMLFile(ctx context.Context, path string) (LoadResult, error) {
	return NewLoader().AddSource(NewTOMLFileSource(path)).Load(ctx)
}

func LoadYAMLFile(ctx context.Context, path string) (LoadResult, error) {
	return NewLoader().AddSource(NewYAMLFileSource(path)).Load(ctx)
}

type TOMLFileSource struct{ name, path string }

func NewTOMLFileSource(path string, opts ...SourceOption) *TOMLFileSource {
	o := sourceOptions{name: "toml"}
	for _, opt := range opts {
		opt(&o)
	}
	return &TOMLFileSource{name: o.name, path: path}
}
func (s *TOMLFileSource) Name() string { return s.name }
func (s *TOMLFileSource) Kind() string { return "toml" }
func (s *TOMLFileSource) Path() string { return s.path }
func (s *TOMLFileSource) Load(ctx context.Context) (Map, error) {
	const op = "configx.TOMLFileSource.Load"
	if ctx == nil {
		return nil, validationError(op, "context is required", nil)
	}
	if s.path == "" {
		return nil, validationError(op, "path is required", nil)
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		return nil, WrapError(ErrorKindConfig, op, "read toml file failed", false, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, contextError(op, err)
	}
	var raw map[string]any
	if err := toml.Unmarshal(b, &raw); err != nil {
		return nil, WrapError(ErrorKindConfig, op, "parse toml file failed", false, err)
	}
	return flattenMap(raw, s.name), nil
}

type YAMLFileSource struct{ name, path string }

func NewYAMLFileSource(path string, opts ...SourceOption) *YAMLFileSource {
	o := sourceOptions{name: "yaml"}
	for _, opt := range opts {
		opt(&o)
	}
	return &YAMLFileSource{name: o.name, path: path}
}
func (s *YAMLFileSource) Name() string { return s.name }
func (s *YAMLFileSource) Kind() string { return "yaml" }
func (s *YAMLFileSource) Path() string { return s.path }
func (s *YAMLFileSource) Load(ctx context.Context) (Map, error) {
	const op = "configx.YAMLFileSource.Load"
	if ctx == nil {
		return nil, validationError(op, "context is required", nil)
	}
	if s.path == "" {
		return nil, validationError(op, "path is required", nil)
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		return nil, WrapError(ErrorKindConfig, op, "read yaml file failed", false, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, contextError(op, err)
	}
	var raw map[string]any
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return nil, WrapError(ErrorKindConfig, op, "parse yaml file failed", false, err)
	}
	return flattenMap(raw, s.name), nil
}
