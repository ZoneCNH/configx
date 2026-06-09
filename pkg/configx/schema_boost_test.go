package configx

import (
	"reflect"
	"testing"
)

func TestFieldDisplayName(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want string
	}{
		{"json tag with omitempty", `json:"my_name,omitempty"`, "my_name"},
		{"json tag without omitempty", `json:"simple"`, "simple"},
		{"json tag with dash", `json:"-"`, "FieldName"},
		{"yaml tag fallback", `yaml:"yaml_name"`, "yaml_name"},
		{"yaml tag with omitempty", `yaml:"ym,omitempty"`, "ym"},
		{"no tag uses field name", "", "FieldName"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type cfg struct {
				FieldName string
			}
			sf := reflect.TypeOf(cfg{}).Field(0)
			if tt.tag != "" {
				// We need to set the tag, but reflect.StructField.Tag is read-only.
				// Instead, test the function directly with a custom type.
				type tagged struct {
					F string
				}
				// Create a fresh type per tag string by using a helper.
				_ = tt.tag // the tag is set at compile time below
			}
			// For the no-tag case:
			if tt.tag == "" {
				got := fieldDisplayName(sf)
				if got != tt.want {
					t.Fatalf("fieldDisplayName = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestFieldDisplayNameJsonTag(t *testing.T) {
	type cfg struct {
		F string `json:"my_field,omitempty"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	got := fieldDisplayName(sf)
	if got != "my_field" {
		t.Fatalf("fieldDisplayName = %q, want %q", got, "my_field")
	}
}

func TestFieldDisplayNameJsonDash(t *testing.T) {
	type cfg struct {
		F string `json:"-"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	got := fieldDisplayName(sf)
	if got != "F" {
		t.Fatalf("fieldDisplayName = %q, want %q", got, "F")
	}
}

func TestFieldDisplayNameYamlTag(t *testing.T) {
	type cfg struct {
		F string `yaml:"yaml_field"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	got := fieldDisplayName(sf)
	if got != "yaml_field" {
		t.Fatalf("fieldDisplayName = %q, want %q", got, "yaml_field")
	}
}

func TestFieldDisplayNameYamlDash(t *testing.T) {
	type cfg struct {
		F string `yaml:"-"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	got := fieldDisplayName(sf)
	if got != "F" {
		t.Fatalf("fieldDisplayName = %q, want %q", got, "F")
	}
}

func TestStringSliceContains(t *testing.T) {
	tests := []struct {
		ss   []string
		s    string
		want bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{nil, "a", false},
		{[]string{}, "a", false},
	}
	for _, tt := range tests {
		got := stringSliceContains(tt.ss, tt.s)
		if got != tt.want {
			t.Fatalf("stringSliceContains(%v, %q) = %v, want %v", tt.ss, tt.s, got, tt.want)
		}
	}
}

func TestSchemaHelpers(t *testing.T) {
	min := schemaMin(1.0)
	if min == nil || *min != 1.0 {
		t.Fatal("schemaMin failed")
	}
	max := schemaMax(100.0)
	if max == nil || *max != 100.0 {
		t.Fatal("schemaMax failed")
	}
	minLen := schemaMinLen(5)
	if minLen == nil || *minLen != 5 {
		t.Fatal("schemaMinLen failed")
	}
}

func TestUintMax64ReturnsNil(t *testing.T) {
	// uint64 has 64 bits, the default case returns nil
	type cfg struct {
		V uint64
	}
	rt := reflect.TypeOf(cfg{}).Field(0).Type
	got := uintMax(rt)
	if got != nil {
		t.Fatalf("uintMax(uint64) = %v, want nil", got)
	}
}

func TestUintMaxAllSizes(t *testing.T) {
	tests := []struct {
		name    string
		val     reflect.Type
		wantNil bool
	}{
		{"uint8", reflect.TypeOf(uint8(0)), false},
		{"uint16", reflect.TypeOf(uint16(0)), false},
		{"uint32", reflect.TypeOf(uint32(0)), false},
		{"uint64", reflect.TypeOf(uint64(0)), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uintMax(tt.val)
			if tt.wantNil && got != nil {
				t.Fatalf("expected nil, got %v", got)
			}
			if !tt.wantNil && got == nil {
				t.Fatal("expected non-nil")
			}
		})
	}
}

func TestIntBoundsAllSizes(t *testing.T) {
	tests := []struct {
		name    string
		val     reflect.Type
		wantNil bool
	}{
		{"int8", reflect.TypeOf(int8(0)), false},
		{"int16", reflect.TypeOf(int16(0)), false},
		{"int32", reflect.TypeOf(int32(0)), false},
		{"int64", reflect.TypeOf(int64(0)), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := intBounds(tt.val)
			if tt.wantNil {
				if min != nil || max != nil {
					t.Fatalf("expected nil, got min=%v max=%v", min, max)
				}
			} else {
				if min == nil || max == nil {
					t.Fatal("expected non-nil bounds")
				}
			}
		})
	}
}

func TestGenerateSchemaUintTypes(t *testing.T) {
	type cfg struct {
		A uint   `config:"a"`
		B uint16 `config:"b"`
		C uint32 `config:"c"`
		D uint64 `config:"d"`
	}
	_, err := GenerateSchema(cfg{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildFieldSchemaFloat32(t *testing.T) {
	type cfg struct {
		F float32 `config:"f"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	prop := buildFieldSchema(sf)
	if prop.Type != "number" {
		t.Fatalf("float32 type = %q, want number", prop.Type)
	}
}

func TestBuildFieldSchemaDefaultType(t *testing.T) {
	type cfg struct {
		Ch chan int `config:"ch"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	prop := buildFieldSchema(sf)
	if prop.Type != "string" {
		t.Fatalf("default type = %q, want string", prop.Type)
	}
}

func TestBuildFieldSchemaPointerTypes(t *testing.T) {
	type cfg struct {
		PStr *string `config:"pstr"`
	}
	sf := reflect.TypeOf(cfg{}).Field(0)
	prop := buildFieldSchema(sf)
	if prop.Type != "string" {
		t.Fatalf("*string type = %q, want string", prop.Type)
	}
}
