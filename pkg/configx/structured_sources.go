package configx

import "context"

func LoadTOMLFile(ctx context.Context, path string) (LoadResult, error) {
	return NewLoader().AddSource(NewMapSource("toml", map[string]string{})).Load(ctx)
}
