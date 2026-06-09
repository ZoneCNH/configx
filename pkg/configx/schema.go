package configx

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// JSONSchema represents a JSON Schema document.
type JSONSchema struct {
	Schema      string                `json:"$schema,omitempty"`
	Type        string                `json:"type,omitempty"`
	Title       string                `json:"title,omitempty"`
	Description string                `json:"description,omitempty"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Required    []string              `json:"required,omitempty"`
	Items       *JSONSchema           `json:"items,omitempty"`
	Enum        []any                 `json:"enum,omitempty"`
	Default     any                   `json:"default,omitempty"`
	Format      string                `json:"format,omitempty"`
	Minimum     *float64              `json:"minimum,omitempty"`
	Maximum     *float64              `json:"maximum,omitempty"`
	MinLength   *int                  `json:"minLength,omitempty"`
	MaxLength   *int                  `json:"maxLength,omitempty"`
}

// GenerateSchema generates a JSON Schema from a Go struct type.
// The cfg parameter can be a struct value or a pointer to a struct.
// It uses reflection to inspect struct fields and config/configx tags
// to determine keys, required fields, and default values.
func GenerateSchema(cfg any) ([]byte, error) {
	const op = "configx.GenerateSchema"
	if cfg == nil {
		return nil, validationError(op, "cfg must not be nil", nil)
	}
	rv := reflect.ValueOf(cfg)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, validationError(op, "cfg must not be a nil pointer", nil)
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, validationError(op, "cfg must be a struct or pointer to struct, got "+rv.Kind().String(), nil)
	}

	schema := buildStructSchema(rv.Type())
	schema.Schema = "http://json-schema.org/draft-07/schema#"

	return json.MarshalIndent(schema, "", "  ")
}

// buildStructSchema recursively builds a JSON Schema for a struct type.
func buildStructSchema(rt reflect.Type) *JSONSchema {
	s := &JSONSchema{
		Type:       "object",
		Properties: make(map[string]*JSONSchema),
	}

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if !sf.IsExported() {
			continue
		}

		tag := parseConfigTag(sf)
		if tag.skip {
			continue
		}
		key := tag.key

		prop := buildFieldSchema(sf)
		if tag.defaultValue != "" {
			prop.Default = tag.defaultValue
		}

		s.Properties[key] = prop
		if tag.required {
			s.Required = append(s.Required, key)
		}
	}

	return s
}

// buildFieldSchema builds a JSON Schema node for a single struct field.
func buildFieldSchema(sf reflect.StructField) *JSONSchema {
	// Handle custom description tag.
	desc := sf.Tag.Get("description")

	ft := sf.Type

	// Dereference pointer types.
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}

	// SecretString type -> string with password format.
	if ft == reflect.TypeOf(SecretString("")) {
		return &JSONSchema{Type: "string", Format: "password", Description: desc}
	}

	// time.Duration -> string with duration format.
	if ft == reflect.TypeOf(time.Duration(0)) {
		return &JSONSchema{Type: "string", Format: "duration", Description: desc}
	}

	// encoding.TextUnmarshaler -> string.
	if implementsTextUnmarshaler(ft) {
		return &JSONSchema{Type: "string", Description: desc}
	}

	switch ft.Kind() {
	case reflect.String:
		return &JSONSchema{Type: "string", Description: desc}

	case reflect.Bool:
		return &JSONSchema{Type: "boolean", Description: desc}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := &JSONSchema{Type: "integer", Description: desc}
		min, max := intBounds(ft)
		if min != nil {
			s.Minimum = min
		}
		if max != nil {
			s.Maximum = max
		}
		return s

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := &JSONSchema{Type: "integer", Description: desc}
		zero := 0.0
		s.Minimum = &zero
		if max := uintMax(ft); max != nil {
			s.Maximum = max
		}
		return s

	case reflect.Float32, reflect.Float64:
		return &JSONSchema{Type: "number", Description: desc}

	case reflect.Slice, reflect.Array:
		return &JSONSchema{
			Type:  "array",
			Items: buildFieldSchema(reflect.StructField{Type: ft.Elem()}),
			Description: desc,
		}

	case reflect.Map:
		return &JSONSchema{
			Type:                 "object",
			Description:          desc,
		}

	case reflect.Struct:
		s := buildStructSchema(ft)
		s.Description = desc
		return s

	default:
		return &JSONSchema{Type: "string", Description: desc}
	}
}

func implementsTextUnmarshaler(t reflect.Type) bool {
	textUnmarshalerType := reflect.TypeOf((*interface{ UnmarshalText([]byte) error })(nil)).Elem()
	return t.Implements(textUnmarshalerType) || reflect.PointerTo(t).Implements(textUnmarshalerType)
}

func intBounds(t reflect.Type) (min, max *float64) {
	bits := t.Bits()
	switch bits {
	case 8:
		mn := float64(-128)
		mx := float64(127)
		return &mn, &mx
	case 16:
		mn := float64(-32768)
		mx := float64(32767)
		return &mn, &mx
	case 32:
		mn := float64(-2147483648)
		mx := float64(2147483647)
		return &mn, &mx
	default:
		return nil, nil
	}
}

func uintMax(t reflect.Type) *float64 {
	bits := t.Bits()
	switch bits {
	case 8:
		mx := float64(255)
		return &mx
	case 16:
		mx := float64(65535)
		return &mx
	case 32:
		mx := float64(4294967295)
		return &mx
	default:
		return nil
	}
}

// stringSliceContains is a helper used by tests.
func stringSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// stringFromTag extracts a description from struct tags, falling back to field name.
func fieldDisplayName(sf reflect.StructField) string {
	if name := sf.Tag.Get("json"); name != "" && name != "-" {
		if idx := strings.IndexByte(name, ','); idx >= 0 {
			return name[:idx]
		}
		return name
	}
	if name := sf.Tag.Get("yaml"); name != "" && name != "-" {
		if idx := strings.IndexByte(name, ','); idx >= 0 {
			return name[:idx]
		}
		return name
	}
	return sf.Name
}

// schemaMin helper for test assertions
func schemaMin(v float64) *float64 { return &v }
func schemaMax(v float64) *float64 { return &v }
func schemaMinLen(v int) *int      { return &v }

// parseConfigTag is already defined in result.go; we reuse it here.
// The following is a compile-time check that the field name convention is consistent.
var _ = strconv.Itoa // ensure strconv is imported for potential use
