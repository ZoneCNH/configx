package configx

import (
	"testing"
	"time"
)

func TestEffectiveConfigHashNilPointerStruct(t *testing.T) {
	type cfg struct {
		Host string
	}
	var p *cfg
	h, err := EffectiveConfigHash(p)
	if err != nil {
		t.Fatal(err)
	}
	if h == "" {
		t.Fatal("expected non-empty hash for nil pointer struct")
	}
}

func TestEffectiveConfigHashPrimitiveTypes(t *testing.T) {
	// Non-struct, non-map types go through json.Marshal fallback
	tests := []struct {
		name string
		val  any
	}{
		{"string", "hello"},
		{"int", 42},
		{"bool", true},
		{"float", 3.14},
		{"slice", []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := EffectiveConfigHash(tt.val)
			if err != nil {
				t.Fatal(err)
			}
			if h == "" {
				t.Fatal("expected non-empty hash")
			}
		})
	}
}

func TestEffectiveConfigHashStructWithVolatileTag(t *testing.T) {
	type cfg struct {
		Host      string
		LoadedAt  time.Time `volatile:"true"`
		CreatedAt time.Time // in VolatileFieldNames
	}
	c1 := cfg{Host: "localhost", LoadedAt: time.Now(), CreatedAt: time.Now()}
	c2 := cfg{Host: "localhost", LoadedAt: c1.LoadedAt.Add(time.Hour), CreatedAt: c1.CreatedAt.Add(time.Hour)}
	h1, _ := EffectiveConfigHash(c1)
	h2, _ := EffectiveConfigHash(c2)
	if h1 != h2 {
		t.Errorf("volatile fields should be excluded: %s != %s", h1, h2)
	}
}

func TestEffectiveConfigHashMapWithLowercaseVolatile(t *testing.T) {
	m := map[string]any{
		"host":      "localhost",
		"loadedat":  time.Now(),
		"timestamp": time.Now(),
	}
	h1, err := EffectiveConfigHash(m)
	if err != nil {
		t.Fatal(err)
	}
	m["loadedat"] = time.Now().Add(time.Hour)
	h2, err := EffectiveConfigHash(m)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Errorf("lowercase volatile keys should be excluded: %s != %s", h1, h2)
	}
}

func TestEffectiveConfigHashStructWithJSONTag(t *testing.T) {
	type cfg struct {
		Host string `json:"host_name"`
		Port int    `json:"port_num"`
	}
	c := cfg{Host: "localhost", Port: 5432}
	h, err := EffectiveConfigHash(c)
	if err != nil {
		t.Fatal(err)
	}
	if h == "" {
		t.Fatal("expected non-empty hash")
	}
}

func TestEffectiveConfigHashStructWithOmitempty(t *testing.T) {
	type cfg struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	}
	c := cfg{Host: "localhost", Port: 5432}
	h, err := EffectiveConfigHash(c)
	if err != nil {
		t.Fatal(err)
	}
	if h == "" {
		t.Fatal("expected non-empty hash")
	}
}

func TestEffectiveConfigHashStructUnexportedFields(t *testing.T) {
	type cfg struct {
		Host    string
		private string //nolint:unused
	}
	c := cfg{Host: "localhost", private: "hidden"}
	h, err := EffectiveConfigHash(c)
	if err != nil {
		t.Fatal(err)
	}
	if h == "" {
		t.Fatal("expected non-empty hash")
	}
}

func TestEffectiveConfigHashNil(t *testing.T) {
	_, err := EffectiveConfigHash(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error kind, got: %v", err)
	}
}
