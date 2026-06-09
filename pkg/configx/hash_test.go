package configx

import (
	"testing"
	"time"
)

func TestEffectiveConfigHash_LoadResult(t *testing.T) {
	r := LoadResult{
		Values: Map{
			"db.host": {Key: "db.host", Value: "localhost", Source: "env"},
			"db.port": {Key: "db.port", Value: "5432", Source: "file"},
		},
		LoadedAt: time.Now(), // volatile — should be excluded
	}
	h1, err := EffectiveConfigHash(r)
	if err != nil {
		t.Fatal(err)
	}
	if h1 == "" {
		t.Fatal("expected non-empty hash")
	}
	// Same values, different LoadedAt → same hash.
	r.LoadedAt = r.LoadedAt.Add(time.Hour)
	h2, err := EffectiveConfigHash(r)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Errorf("hash changed with different LoadedAt: %s != %s", h1, h2)
	}
}

func TestEffectiveConfigHash_LoadResultDifferentValues(t *testing.T) {
	r1 := LoadResult{
		Values: Map{
			"db.host": {Key: "db.host", Value: "localhost", Source: "env"},
		},
	}
	r2 := LoadResult{
		Values: Map{
			"db.host": {Key: "db.host", Value: "remotehost", Source: "env"},
		},
	}
	h1, _ := EffectiveConfigHash(r1)
	h2, _ := EffectiveConfigHash(r2)
	if h1 == h2 {
		t.Error("different values should produce different hashes")
	}
}

func TestEffectiveConfigHash_Struct(t *testing.T) {
	type cfg struct {
		Host      string
		Port      int
		LoadedAt  time.Time `volatile:"true"`
		Timestamp time.Time
	}
	c1 := cfg{Host: "localhost", Port: 5432, LoadedAt: time.Now(), Timestamp: time.Now()}
	c2 := cfg{Host: "localhost", Port: 5432, LoadedAt: c1.LoadedAt.Add(time.Hour), Timestamp: c1.Timestamp.Add(time.Hour)}
	h1, err := EffectiveConfigHash(c1)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := EffectiveConfigHash(c2)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Errorf("volatile fields should be excluded: %s != %s", h1, h2)
	}
}

func TestEffectiveConfigHash_StructDifferent(t *testing.T) {
	type cfg struct {
		Host string
		Port int
	}
	c1 := cfg{Host: "localhost", Port: 5432}
	c2 := cfg{Host: "localhost", Port: 3306}
	h1, _ := EffectiveConfigHash(c1)
	h2, _ := EffectiveConfigHash(c2)
	if h1 == h2 {
		t.Error("different configs should produce different hashes")
	}
}

func TestEffectiveConfigHash_Map(t *testing.T) {
	m1 := map[string]any{"host": "localhost", "port": 5432, "LoadedAt": time.Now()}
	m2 := map[string]any{"host": "localhost", "port": 5432, "LoadedAt": m1["LoadedAt"].(time.Time).Add(time.Hour)}
	h1, err := EffectiveConfigHash(m1)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := EffectiveConfigHash(m2)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Errorf("volatile keys should be excluded: %s != %s", h1, h2)
	}
}

func TestEffectiveConfigHash_NilInput(t *testing.T) {
	_, err := EffectiveConfigHash(nil)
	if err == nil {
		t.Error("expected error for nil input")
	}
}

func TestEffectiveConfigHash_PointerStruct(t *testing.T) {
	type cfg struct {
		Host string
		Port int
	}
	c := &cfg{Host: "localhost", Port: 5432}
	h, err := EffectiveConfigHash(c)
	if err != nil {
		t.Fatal(err)
	}
	if h == "" {
		t.Fatal("expected non-empty hash")
	}
}

func TestEffectiveConfigHash_Determinism(t *testing.T) {
	type cfg struct {
		A string
		B string
		C string
	}
	c := cfg{A: "1", B: "2", C: "3"}
	// Run 100 times — should always produce the same hash.
	h0, _ := EffectiveConfigHash(c)
	for i := 0; i < 100; i++ {
		hi, _ := EffectiveConfigHash(c)
		if hi != h0 {
			t.Fatalf("non-deterministic hash on iteration %d: %s != %s", i, hi, h0)
		}
	}
}
