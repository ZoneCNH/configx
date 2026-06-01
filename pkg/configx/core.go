package configx

import (
	"bufio"
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	foundationx "github.com/ZoneCNH/foundationx"
)

const redactionMarker = "***"

type SecretString = foundationx.SecretString

func NewSecretString(value string) SecretString { return foundationx.NewSecretString(value) }

type Value struct {
	Key        string
	Value      string
	Secret     bool
	Source     string
	LoadedAt   time.Time
	Overridden bool
}

type Map map[string]Value

type LoadResult struct {
	Values   Map
	Sources  []SourceReport
	LoadedAt time.Time
}

type SourceReport struct {
	Name      string
	Kind      string
	Path      string
	Loaded    bool
	Error     string
	LoadedAt  time.Time
	ValueKeys []string
}

type SanitizedValue struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	Secret     bool   `json:"secret"`
	Source     string `json:"source"`
	Overridden bool   `json:"overridden"`
}

type SanitizedResult struct {
	Values  map[string]SanitizedValue `json:"values"`
	Sources []SourceReport            `json:"sources"`
}

func (r LoadResult) Get(key string) (string, bool) {
	v, ok := r.Values[key]
	return v.Value, ok
}

func (r LoadResult) Decode(target any) error { return Decode(r, target) }

func LoadEnv(ctx context.Context, prefix string, keys []string) (LoadResult, error) {
	return NewLoader().AddSource(NewEnvSource(prefix, keys)).Load(ctx)
}

func LoadAllEnv(ctx context.Context, prefix string) (LoadResult, error) {
	return NewLoader().AddSource(NewAllEnvSource(prefix)).Load(ctx)
}

func LoadEnvFile(ctx context.Context, path string) (LoadResult, error) {
	return NewLoader().AddSource(NewEnvFileSource(path)).Load(ctx)
}

func LoadJSONFile(ctx context.Context, path string) (LoadResult, error) {
	return NewLoader().AddSource(NewJSONFileSource(path)).Load(ctx)
}

func LoadMap(ctx context.Context, name string, values map[string]string) (LoadResult, error) {
	return NewLoader().AddSource(NewMapSource(name, values)).Load(ctx)
}

func (r LoadResult) Sanitize() SanitizedResult {
	values := make(map[string]SanitizedValue, len(r.Values))
	for key, value := range r.Values {
		out := value.Value
		if value.Secret {
			out = redactionMarker
		}
		values[key] = SanitizedValue{Key: value.Key, Value: out, Secret: value.Secret, Source: value.Source, Overridden: value.Overridden}
	}
	return SanitizedResult{Values: values, Sources: r.Sources}
}

type Source interface {
	Name() string
	Kind() string
	Load(ctx context.Context) (Map, error)
}

type SourceOption func(*sourceOptions)
type sourceOptions struct{ name string }

func WithSourceName(name string) SourceOption { return func(o *sourceOptions) { o.name = name } }

type MergeStrategy int

const LastWins MergeStrategy = iota

type Loader struct {
	sources []Source
	options loaderOptions
}

type LoaderOption func(*loaderOptions)
type loaderOptions struct {
	mergeStrategy MergeStrategy
	failFast      bool
}

func defaultLoaderOptions() loaderOptions {
	return loaderOptions{mergeStrategy: LastWins, failFast: true}
}
func WithMergeStrategy(strategy MergeStrategy) LoaderOption {
	return func(o *loaderOptions) { o.mergeStrategy = strategy }
}
func WithFailFast(failFast bool) LoaderOption {
	return func(o *loaderOptions) { o.failFast = failFast }
}

func NewLoader(opts ...LoaderOption) *Loader {
	options := defaultLoaderOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return &Loader{options: options}
}

func (l *Loader) AddSource(source Source) *Loader {
	if l == nil {
		return l
	}
	l.sources = append(l.sources, source)
	return l
}

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
			if prev, ok := result.Values[key]; ok {
				prev.Overridden = true
				result.Values[key] = prev
			}
			if value.Key == "" {
				value.Key = key
			}
			if value.Source == "" {
				value.Source = report.Name
			}
			if value.LoadedAt.IsZero() {
				value.LoadedAt = report.LoadedAt
			}
			result.Values[key] = value
			report.ValueKeys = append(report.ValueKeys, key)
		}
		result.Sources = append(result.Sources, report)
	}
	return result, nil
}

type EnvSource struct {
	name, prefix string
	keys         []string
	all          bool
}

func NewEnvSource(prefix string, keys []string, opts ...SourceOption) *EnvSource {
	o := sourceOptions{name: "env"}
	for _, opt := range opts {
		opt(&o)
	}
	return &EnvSource{name: o.name, prefix: prefix, keys: append([]string(nil), keys...)}
}
func NewAllEnvSource(prefix string, opts ...SourceOption) *EnvSource {
	s := NewEnvSource(prefix, nil, opts...)
	s.all = true
	return s
}
func (s *EnvSource) Name() string { return s.name }
func (s *EnvSource) Kind() string { return "env" }
func (s *EnvSource) Load(ctx context.Context) (Map, error) {
	if ctx == nil {
		return nil, validationError("configx.EnvSource.Load", "context is required", nil)
	}
	if err := ctx.Err(); err != nil {
		return nil, contextError("configx.EnvSource.Load", err)
	}
	out := Map{}
	now := time.Now().UTC()
	if !s.all {
		for _, key := range s.keys {
			name := s.prefix + key
			if value, ok := os.LookupEnv(name); ok {
				out[key] = Value{Key: key, Value: value, Secret: IsSecretKey(key), Source: s.name, LoadedAt: now}
			}
		}
		return out, nil
	}
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		if s.prefix != "" {
			if !strings.HasPrefix(key, s.prefix) {
				continue
			}
			key = strings.TrimPrefix(key, s.prefix)
		}
		out[key] = Value{Key: key, Value: parts[1], Secret: IsSecretKey(key), Source: s.name, LoadedAt: now}
	}
	return out, nil
}

type EnvFileSource struct{ name, path string }

func NewEnvFileSource(path string, opts ...SourceOption) *EnvFileSource {
	o := sourceOptions{name: "envfile"}
	for _, opt := range opts {
		opt(&o)
	}
	return &EnvFileSource{name: o.name, path: path}
}
func (s *EnvFileSource) Name() string { return s.name }
func (s *EnvFileSource) Kind() string { return "envfile" }
func (s *EnvFileSource) Path() string { return s.path }
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

type JSONFileSource struct{ name, path string }

func NewJSONFileSource(path string, opts ...SourceOption) *JSONFileSource {
	o := sourceOptions{name: "json"}
	for _, opt := range opts {
		opt(&o)
	}
	return &JSONFileSource{name: o.name, path: path}
}
func (s *JSONFileSource) Name() string { return s.name }
func (s *JSONFileSource) Kind() string { return "json" }
func (s *JSONFileSource) Path() string { return s.path }
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

type MapSource struct {
	name    string
	values  map[string]string
	secrets map[string]bool
}

func NewMapSource(name string, values map[string]string) *MapSource {
	return NewSecretMapSource(name, values, nil)
}
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
func (s *MapSource) Name() string { return s.name }
func (s *MapSource) Kind() string { return "map" }
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
		default:
			out[prefix] = Value{Key: prefix, Value: fmt.Sprint(x), Secret: IsSecretKey(prefix), Source: source, LoadedAt: now}
		}
	}
	for k, v := range raw {
		walk(k, v)
	}
	return out
}

type Validator interface{ Validate() error }

func Decode(result LoadResult, target any) error {
	const op = "configx.Decode"
	if target == nil {
		return validationError(op, "target is required", nil)
	}
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return validationError(op, "target must be a non-nil pointer", nil)
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return validationError(op, "target must point to a struct", nil)
	}
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		sf := rt.Field(i)
		if !field.CanSet() {
			continue
		}
		tag := parseConfigTag(sf)
		if tag.skip {
			continue
		}
		raw, ok := findValue(result, tag.key)
		if !ok {
			if tag.defaultValue != "" {
				raw = tag.defaultValue
				ok = true
			}
		}
		if !ok {
			if tag.required {
				return validationError(op, "required config missing: "+tag.key, nil)
			}
			continue
		}
		if err := setField(field, raw); err != nil {
			return validationError(op, "decode "+tag.key+" failed", sanitizeError(err))
		}
	}
	if validator, ok := target.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return validationError(op, "validation failed", err)
		}
	}
	return nil
}

type configTag struct {
	key          string
	defaultValue string
	required     bool
	skip         bool
}

func parseConfigTag(sf reflect.StructField) configTag {
	tag := configTag{key: sf.Name, defaultValue: sf.Tag.Get("default")}
	if sf.Tag.Get("required") == "true" {
		tag.required = true
	}
	if raw := sf.Tag.Get("config"); raw != "" {
		parts := splitTag(raw)
		if len(parts) > 0 {
			if parts[0] == "-" {
				tag.skip = true
				return tag
			}
			if parts[0] != "" {
				tag.key = parts[0]
			}
			applyTagOptions(&tag, parts[1:])
		}
	}
	if raw := sf.Tag.Get("configx"); raw != "" {
		parts := splitTag(raw)
		if len(parts) > 0 {
			if parts[0] == "-" {
				tag.skip = true
				return tag
			}
			if parts[0] != "" && !isTagOption(parts[0]) {
				tag.key = parts[0]
				parts = parts[1:]
			}
			applyTagOptions(&tag, parts)
		}
	}
	return tag
}

func splitTag(raw string) []string {
	parts := strings.Split(raw, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func applyTagOptions(tag *configTag, options []string) {
	for _, option := range options {
		switch {
		case option == "required":
			tag.required = true
		case strings.HasPrefix(option, "default="):
			tag.defaultValue = strings.TrimPrefix(option, "default=")
		}
	}
}

func isTagOption(option string) bool {
	return option == "required" || strings.HasPrefix(option, "default=") || option == "secret"
}

func findValue(result LoadResult, key string) (string, bool) {
	if raw, ok := result.Get(key); ok {
		return raw, true
	}
	normalized := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	if normalized != key {
		return result.Get(normalized)
	}
	return "", false
}

func setField(field reflect.Value, raw string) error {
	if field.CanAddr() {
		if u, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
			if err := u.UnmarshalText([]byte(raw)); err != nil {
				return sanitizeError(err)
			}
			return nil
		}
	}
	if field.Type() == reflect.TypeOf(time.Duration(0)) {
		d, err := time.ParseDuration(raw)
		if err != nil {
			i, ierr := strconv.ParseInt(raw, 10, 64)
			if ierr != nil {
				return errors.New("invalid duration")
			}
			d = time.Duration(i)
		}
		field.SetInt(int64(d))
		return nil
	}
	if field.Type() == reflect.TypeOf(SecretString("")) {
		field.Set(reflect.ValueOf(NewSecretString(raw)))
		return nil
	}
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return errors.New("invalid bool")
		}
		field.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(raw, 10, field.Type().Bits())
		if err != nil {
			return errors.New("invalid integer")
		}
		field.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(raw, 10, field.Type().Bits())
		if err != nil {
			return errors.New("invalid unsigned integer")
		}
		field.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(raw, field.Type().Bits())
		if err != nil {
			return errors.New("invalid float")
		}
		field.SetFloat(v)
	default:
		return errors.New("unsupported field type " + field.Type().String())
	}
	return nil
}

func IsSecretKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "secret") || strings.Contains(k, "password") || strings.Contains(k, "passwd") || strings.Contains(k, "token") || strings.Contains(k, "access_key") || strings.Contains(k, "secret_key")
}

func sanitizeError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(sanitizeMessage(err.Error()))
}

func sanitizeMessage(message string) string {
	parts := strings.FieldsFunc(message, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ',' || r == '&' || r == ';'
	})
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok || value == "" || !IsSecretKey(strings.Trim(key, `"'`)) {
			continue
		}
		message = strings.ReplaceAll(message, part, key+"="+redactionMarker)
	}
	return message
}
