package configx

import (
	"testing"
)

func TestProvenance_RecordAndGet(t *testing.T) {
	p := NewProvenance()
	p.Record("db.host", "env", 10)

	entry, ok := p.Get("db.host")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if entry.Source != "env" {
		t.Errorf("source = %q, want %q", entry.Source, "env")
	}
	if entry.Priority != 10 {
		t.Errorf("priority = %d, want 10", entry.Priority)
	}
	if len(entry.Overrides) != 0 {
		t.Errorf("overrides = %d, want 0", len(entry.Overrides))
	}
}

func TestProvenance_RecordOverride(t *testing.T) {
	p := NewProvenance()
	p.Record("db.host", "file", 5)
	p.RecordOverride("db.host", "env", 10, "localhost", "10.0.0.1")

	entry, ok := p.Get("db.host")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if entry.Source != "env" {
		t.Errorf("source = %q, want %q", entry.Source, "env")
	}
	if entry.Priority != 10 {
		t.Errorf("priority = %d, want 10", entry.Priority)
	}
	if len(entry.Overrides) != 1 {
		t.Fatalf("overrides = %d, want 1", len(entry.Overrides))
	}
	o := entry.Overrides[0]
	if o.Source != "env" || o.OldValue != "localhost" || o.NewValue != "10.0.0.1" {
		t.Errorf("override = %+v", o)
	}
}

func TestProvenance_MultipleOverrides(t *testing.T) {
	p := NewProvenance()
	p.Record("db.port", "default", 1)
	p.RecordOverride("db.port", "file", 5, "5432", "3306")
	p.RecordOverride("db.port", "env", 10, "3306", "6379")

	entry, _ := p.Get("db.port")
	if len(entry.Overrides) != 2 {
		t.Fatalf("overrides = %d, want 2", len(entry.Overrides))
	}
	if entry.Overrides[1].NewValue != "6379" {
		t.Errorf("last override new value = %q, want %q", entry.Overrides[1].NewValue, "6379")
	}
}

func TestProvenance_GetMissingKey(t *testing.T) {
	p := NewProvenance()
	_, ok := p.Get("nonexistent")
	if ok {
		t.Error("expected false for missing key")
	}
}

func TestProvenance_NilReceiver(t *testing.T) {
	var p *Provenance
	// Should not panic.
	p.Record("k", "s", 1)
	p.RecordOverride("k", "s2", 2, "a", "b")
	_, ok := p.Get("k")
	if ok {
		t.Error("expected false on nil receiver")
	}
	if p.Snapshot() != nil {
		t.Error("expected nil snapshot on nil receiver")
	}
	if p.Keys() != nil {
		t.Error("expected nil keys on nil receiver")
	}
	p.Reset()
}

func TestProvenance_Snapshot(t *testing.T) {
	p := NewProvenance()
	p.Record("a", "env", 10)
	p.Record("b", "file", 5)
	snap := p.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("snapshot len = %d, want 2", len(snap))
	}
	// Mutating snapshot should not affect original.
	snap["c"] = ProvenanceEntry{Source: "x", Priority: 1}
	if _, ok := p.Get("c"); ok {
		t.Error("mutating snapshot should not affect provenance")
	}
}

func TestProvenance_Keys(t *testing.T) {
	p := NewProvenance()
	p.Record("z", "env", 1)
	p.Record("a", "file", 2)
	p.Record("m", "flag", 3)
	keys := p.Keys()
	expected := []string{"a", "m", "z"}
	if len(keys) != len(expected) {
		t.Fatalf("keys = %v, want %v", keys, expected)
	}
	for i, k := range keys {
		if k != expected[i] {
			t.Errorf("keys[%d] = %q, want %q", i, k, expected[i])
		}
	}
}

func TestProvenance_Reset(t *testing.T) {
	p := NewProvenance()
	p.Record("a", "env", 1)
	p.Reset()
	if len(p.Snapshot()) != 0 {
		t.Error("expected empty snapshot after reset")
	}
}
