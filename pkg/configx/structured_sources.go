package configx

import (
	"context"
	"os"

	toml "github.com/pelletier/go-toml/v2"
	yaml "gopkg.in/yaml.v3"
)

type structuredDecoder func([]byte, any) error

type structuredFileSource struct {
	name   string
	kind   string
	path   string
	decode structuredDecoder
}

func LoadTOMLFile(ctx context.Context, path string) (LoadResult, error) {
	return NewLoader().AddSource(NewTOMLFileSource(path)).Load(ctx)
}

func LoadYAMLFile(ctx context.Context, path string) (LoadResult, error) {
	return NewLoader().AddSource(NewYAMLFileSource(path)).Load(ctx)
}

type TOMLFileSource struct{ structuredFileSource }

type YAMLFileSource struct{ structuredFileSource }

func NewTOMLFileSource(path string, opts ...SourceOption) *TOMLFileSource {
	return &TOMLFileSource{structuredFileSource: newStructuredFileSource("toml", path, toml.Unmarshal, opts...)}
}

func NewYAMLFileSource(path string, opts ...SourceOption) *YAMLFileSource {
	return &YAMLFileSource{structuredFileSource: newStructuredFileSource("yaml", path, yaml.Unmarshal, opts...)}
}

func newStructuredFileSource(kind string, path string, decode structuredDecoder, opts ...SourceOption) structuredFileSource {
	o := sourceOptions{name: kind}
	for _, opt := range opts {
		opt(&o)
	}
	return structuredFileSource{name: o.name, kind: kind, path: path, decode: decode}
}

func (s *structuredFileSource) Name() string { return s.name }
func (s *structuredFileSource) Kind() string { return s.kind }
func (s *structuredFileSource) Path() string { return s.path }

func (s *structuredFileSource) Load(ctx context.Context) (Map, error) {
	op := "configx." + s.kind + "FileSource.Load"
	if ctx == nil {
		return nil, validationError(op, "context is required", nil)
	}
	if s.path == "" {
		return nil, validationError(op, "path is required", nil)
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		return nil, WrapError(ErrorKindConfig, op, "read file failed", false, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, contextError(op, err)
	}
	var raw map[string]any
	if err := s.decode(b, &raw); err != nil {
		return nil, WrapError(ErrorKindConfig, op, "parse file failed", false, err)
	}
	return flattenMap(raw, s.name), nil
}
