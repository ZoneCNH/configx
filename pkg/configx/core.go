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
type Sanitizer = foundationx.Sanitizer

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

func Sanitize(result LoadResult) SanitizedResult { return result.Sanitize() }

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

type MergeStrategy string

const (
	MergeLastWins        MergeStrategy = "last_wins"
	MergeFirstWins       MergeStrategy = "first_wins"
	MergeErrorOnConflict MergeStrategy = "error_on_conflict"

	// LastWins preserves the original public constant name while the explicit
	// Merge* names document all supported strategies.
	LastWins = MergeLastWins
)

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
	return loaderOptions{mergeStrategy: MergeLastWins, failFast: true}
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
	if !isSupportedMergeStrategy(l.options.mergeStrategy) {
		return LoadResult{}, validationError(op, "unsupported merge strategy: "+string(l.options.mergeStrategy), nil)
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
			if err := mergeValue(result.Values, key, value, l.options.mergeStrategy); err != nil {
				report.Loaded = false
				report.Error = sanitizeMessage(err.Error())
				if l.options.failFast {
					result.Sources = append(result.Sources, report)
					return result, err
				}
				continue
			}
			report.ValueKeys = append(report.ValueKeys, key)
		}
		result.Sources = append(result.Sources, report)
	}
	return result, nil
}

func Merge(strategy MergeStrategy, maps ...Map) (Map, error) {
	if !isSupportedMergeStrategy(strategy) {
		return nil, validationError("configx.Merge", "unsupported merge strategy: "+string(strategy), nil)
	}
	merged := Map{}
	for _, values := range maps {
		for key, value := range values {
			if value.Key == "" {
				value.Key = key
			}
			if err := mergeValue(merged, key, value, strategy); err != nil {
				return merged, err
			}
		}
	}
	return merged, nil
}

func mergeValue(values Map, key string, next Value, strategy MergeStrategy) error {
	const op = "configx.Merge"
	if values == nil {
		return validationError(op, "target map is required", nil)
	}
	if key == "" {
		return validationError(op, "key is required", nil)
	}
	current, exists := values[key]
	if !exists {
		values[key] = next
		return nil
	}
	switch normalizeMergeStrategy(strategy) {
	case MergeLastWins:
		next.Overridden = true
		values[key] = next
		return nil
	case MergeFirstWins:
		current.Overridden = true
		values[key] = current
		return nil
	case MergeErrorOnConflict:
		return WrapError(ErrorKindConflict, op, "config key conflict: "+key, false, nil)
	default:
		return validationError(op, "unsupported merge strategy: "+string(strategy), nil)
	}
}

func normalizeMergeStrategy(strategy MergeStrategy) MergeStrategy {
	if strategy == "" {
		return MergeLastWins
	}
	return strategy
}

func isSupportedMergeStrategy(strategy MergeStrategy) bool {
	switch normalizeMergeStrategy(strategy) {
	case MergeLastWins, MergeFirstWins, MergeErrorOnConflict:
		return true
	default:
		return false
	}
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

type Validator interface{ Validate() error }

func DecodeMap(values Map, target any) error {
	return Decode(LoadResult{Values: values}, target)
}

func Validate(target any) error {
	validator, ok := target.(Validator)
	if !ok {
		return nil
	}
	return sanitizeTargetError(target, validator.Validate())
}

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
	if err := decodeStruct(result, rv, ""); err != nil {
		return err
	}
	if err := Validate(target); err != nil {
		return validationError(op, "validation failed", sanitizeResultError(result, err))
	}
	return nil
}

func decodeStruct(result LoadResult, rv reflect.Value, prefix string) error {
	const op = "configx.Decode"
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
		key := joinConfigKey(prefix, tag.key)
		if isStructField(field) {
			nested := field
			if field.Kind() == reflect.Pointer {
				if field.IsNil() {
					field.Set(reflect.New(field.Type().Elem()))
				}
				nested = field.Elem()
			}
			if err := decodeStruct(result, nested, key); err != nil {
				return err
			}
			continue
		}
		raw, resolvedKey, ok := findValue(result, key)
		if !ok {
			if tag.defaultValue != "" {
				raw = tag.defaultValue
				ok = true
			}
		}
		if !ok {
			if tag.required {
				return validationError(op, "required config missing: "+key, nil)
			}
			continue
		}
		if err := setField(field, raw); err != nil {
			return validationError(op, "decode "+key+" failed", sanitizeResultError(result, err))
		}
		if tag.secret && resolvedKey != "" {
			value := result.Values[resolvedKey]
			value.Secret = true
			result.Values[resolvedKey] = value
		}
	}
	return nil
}

func joinConfigKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	if key == "" {
		return prefix
	}
	return prefix + "." + key
}

func isStructField(field reflect.Value) bool {
	t := field.Type()
	if field.Kind() == reflect.Pointer {
		t = field.Type().Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	return t != reflect.TypeOf(time.Duration(0)) && t != reflect.TypeOf(SecretString("")) && !reflect.PointerTo(t).Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem())
}

type configTag struct {
	key          string
	defaultValue string
	required     bool
	secret       bool
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
		case option == "secret":
			tag.secret = true
		case strings.HasPrefix(option, "default="):
			tag.defaultValue = strings.TrimPrefix(option, "default=")
		}
	}
}

func isTagOption(option string) bool {
	return option == "required" || strings.HasPrefix(option, "default=") || option == "secret"
}

func findValue(result LoadResult, key string) (string, string, bool) {
	if raw, ok := result.Get(key); ok {
		return raw, key, true
	}
	normalized := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	if normalized != key {
		if raw, ok := result.Get(normalized); ok {
			return raw, normalized, true
		}
	}
	return "", "", false
}

func setField(field reflect.Value, raw string) error {
	if field.Kind() == reflect.Slice {
		return setSliceField(field, raw)
	}
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

func setSliceField(field reflect.Value, raw string) error {
	if field.Type().Elem().Kind() != reflect.String {
		return errors.New("unsupported field type " + field.Type().String())
	}
	if raw == "" {
		field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		return nil
	}
	parts := strings.Split(raw, ",")
	values := reflect.MakeSlice(field.Type(), len(parts), len(parts))
	for i, part := range parts {
		values.Index(i).SetString(strings.TrimSpace(part))
	}
	field.Set(values)
	return nil
}

func IsSecretKey(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	normalized := strings.NewReplacer("-", "_", ".", "_").Replace(k)
	compact := strings.ReplaceAll(normalized, "_", "")
	return strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "password") ||
		strings.Contains(normalized, "passwd") ||
		strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "api_key") ||
		strings.Contains(compact, "apikey") ||
		strings.Contains(normalized, "access_key") ||
		strings.Contains(normalized, "secret_key") ||
		strings.Contains(normalized, "private_key")
}

func sanitizeResultError(result LoadResult, err error) error {
	if err == nil {
		return nil
	}
	message := sanitizeMessage(err.Error())
	for key, value := range result.Values {
		if value.Value == "" || (!value.Secret && !IsSecretKey(key)) {
			continue
		}
		message = strings.ReplaceAll(message, value.Value, redactionMarker)
	}
	return errors.New(message)
}

func sanitizeTargetError(target any, err error) error {
	if err == nil {
		return nil
	}
	message := sanitizeMessage(err.Error())
	for _, secret := range collectTargetSecrets(reflect.ValueOf(target), nil) {
		message = strings.ReplaceAll(message, secret, redactionMarker)
	}
	if message == err.Error() {
		return err
	}
	return errors.New(message)
}

func collectTargetSecrets(value reflect.Value, seen map[uintptr]struct{}) []string {
	if !value.IsValid() {
		return nil
	}
	if seen == nil {
		seen = make(map[uintptr]struct{})
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		ptr := value.Pointer()
		if _, ok := seen[ptr]; ok {
			return nil
		}
		seen[ptr] = struct{}{}
		value = value.Elem()
	}
	if value.Type() == reflect.TypeOf(SecretString("")) {
		secret := value.Interface().(SecretString).Reveal()
		if secret != "" {
			return []string{secret}
		}
		return nil
	}
	if value.Kind() != reflect.Struct {
		return nil
	}
	var secrets []string
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if !field.CanInterface() {
			continue
		}
		sf := value.Type().Field(i)
		if parseConfigTag(sf).secret || sf.Tag.Get("secret") == "true" || field.Type() == reflect.TypeOf(SecretString("")) {
			if field.Type() == reflect.TypeOf(SecretString("")) {
				secret := field.Interface().(SecretString).Reveal()
				if secret != "" {
					secrets = append(secrets, secret)
				}
			} else if field.Kind() == reflect.String && field.String() != "" {
				secrets = append(secrets, field.String())
			}
		}
		secrets = append(secrets, collectTargetSecrets(field, seen)...)
	}
	return secrets
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
