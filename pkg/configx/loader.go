package configx

import (
	"context"
	"time"
)

// Loader orchestrates loading configuration from multiple sources.
type Loader struct {
	sources []Source
	options loaderOptions
}

// LoaderOption configures a Loader.
type LoaderOption func(*loaderOptions)

type loaderOptions struct {
	mergeStrategy MergeStrategy
	failFast      bool
}

func defaultLoaderOptions() loaderOptions {
	return loaderOptions{mergeStrategy: LastWins, failFast: true}
}

// WithMergeStrategy sets the merge strategy for the loader.
func WithMergeStrategy(strategy MergeStrategy) LoaderOption {
	return func(o *loaderOptions) { o.mergeStrategy = strategy }
}

// WithFailFast sets whether the loader should fail on the first error.
func WithFailFast(failFast bool) LoaderOption {
	return func(o *loaderOptions) { o.failFast = failFast }
}

// NewLoader creates a new Loader with the given options.
func NewLoader(opts ...LoaderOption) *Loader {
	options := defaultLoaderOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return &Loader{options: options}
}

// AddSource adds a configuration source to the loader.
func (l *Loader) AddSource(source Source) *Loader {
	if l == nil {
		return l
	}
	l.sources = append(l.sources, source)
	return l
}

// Load loads configuration from all added sources.
func (l *Loader) Load(ctx context.Context) (LoadResult, error) {
	const op = "configx.Loader.Load"
	if ctx == nil {
		return LoadResult{}, validationError(op, "context is required", nil)
	}
	if l == nil {
		return LoadResult{}, validationError(op, "loader is nil", nil)
	}
	loadedAt := time.Now().UTC()
	result := LoadResult{Values: Map{}, LoadedAt: loadedAt}
	for _, source := range l.sources {
		if err := ctx.Err(); err != nil {
			return result, contextError(op, err)
		}
		report := SourceReport{LoadedAt: time.Now().UTC()}
		if source == nil {
			report.Error = "source is nil"
			result.Sources = append(result.Sources, report)
			if l.options.failFast {
				return result, validationError(op, report.Error, nil)
			}
			continue
		}
		report.Name, report.Kind = source.Name(), source.Kind()
		if pather, ok := source.(interface{ Path() string }); ok {
			report.Path = pather.Path()
		}
		values, err := source.Load(ctx)
		if err != nil {
			report.Error = sanitizeMessage(err.Error())
			result.Sources = append(result.Sources, report)
			if l.options.failFast {
				return result, WrapError(ErrorKindConfig, op, report.Error, false, sanitizeError(err))
			}
			continue
		}
		report.Loaded = true
		for key, value := range values {
			if value.Key == "" {
				value.Key = key
			}
			if value.Source == "" {
				value.Source = report.Name
			}
			if value.LoadedAt.IsZero() {
				value.LoadedAt = report.LoadedAt
			}
			report.ValueKeys = append(report.ValueKeys, key)
			if err := mergeValue(result.Values, key, value, l.options.mergeStrategy); err != nil {
				report.Error = err.Error()
				result.Sources = append(result.Sources, report)
				return result, WrapError(ErrorKindConflict, op, report.Error, false, err)
			}
		}
		result.Sources = append(result.Sources, report)
	}
	return result, nil
}
