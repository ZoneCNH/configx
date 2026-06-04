package configx

import (
	"encoding"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Value holds a single configuration value with its metadata.
type Value struct {
	Key        string
	Value      string
	Secret     bool
	Source     string
	LoadedAt   time.Time
	Overridden bool
}

// Map is a map of configuration keys to Values.
type Map map[string]Value

// LoadResult holds the result of loading configuration from all sources.
type LoadResult struct {
	Values   Map
	Sources  []SourceReport
	LoadedAt time.Time
}

// SourceReport records metadata about a single source load attempt.
type SourceReport struct {
	Name      string
	Kind      string
	Path      string
	Loaded    bool
	Error     string
	LoadedAt  time.Time
	ValueKeys []string
}

// SanitizedValue is a Value with the secret masked for safe display.
type SanitizedValue struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	Secret     bool   `json:"secret"`
	Source     string `json:"source"`
	Overridden bool   `json:"overridden"`
}

// SanitizedResult is a LoadResult with all secrets masked.
type SanitizedResult struct {
	Values  map[string]SanitizedValue `json:"values"`
	Sources []SourceReport            `json:"sources"`
}

// Get returns the value for the given key and whether it was found.
func (r LoadResult) Get(key string) (string, bool) {
	v, ok := r.Values[key]
	return v.Value, ok
}

// Decode decodes the loaded configuration into the target struct.
func (r LoadResult) Decode(target any) error { return Decode(r, target) }

// Sanitize returns a copy of the result with all secret values masked.
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

// Validator is implemented by config structs that need custom validation.
type Validator interface{ Validate() error }

// Decode decodes a LoadResult into the target struct using config tags.
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
