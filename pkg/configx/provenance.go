package configx

import (
	"sort"
	"sync"
)

// OverrideEntry records a single override in a key's provenance chain.
type OverrideEntry struct {
	Source   string
	OldValue string
	NewValue string
}

// ProvenanceEntry records the origin, priority, and override history of a configuration key.
type ProvenanceEntry struct {
	Source   string
	Priority int
	Overrides []OverrideEntry
}

// Provenance tracks the provenance of every loaded configuration key.
// It is safe for concurrent use.
type Provenance struct {
	mu      sync.RWMutex
	entries map[string]ProvenanceEntry
}

// NewProvenance creates an empty Provenance tracker.
func NewProvenance() *Provenance {
	return &Provenance{entries: make(map[string]ProvenanceEntry)}
}

// Record records the initial source for a key.
func (p *Provenance) Record(key, source string, priority int) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.entries[key] = ProvenanceEntry{
		Source:   source,
		Priority: priority,
	}
}

// RecordOverride records that a key was overridden by a new source.
func (p *Provenance) RecordOverride(key, newSource string, priority int, oldValue, newValue string) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	entry := p.entries[key]
	entry.Overrides = append(entry.Overrides, OverrideEntry{
		Source:   newSource,
		OldValue: oldValue,
		NewValue: newValue,
	})
	entry.Source = newSource
	entry.Priority = priority
	p.entries[key] = entry
}

// Get returns the ProvenanceEntry for a key and whether it exists.
func (p *Provenance) Get(key string) (ProvenanceEntry, bool) {
	if p == nil {
		return ProvenanceEntry{}, false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	entry, ok := p.entries[key]
	return entry, ok
}

// Snapshot returns a copy of the entire provenance map, sorted by key.
func (p *Provenance) Snapshot() map[string]ProvenanceEntry {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make(map[string]ProvenanceEntry, len(p.entries))
	for k, v := range p.entries {
		out[k] = v
	}
	return out
}

// Keys returns a sorted slice of all tracked keys.
func (p *Provenance) Keys() []string {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	keys := make([]string, 0, len(p.entries))
	for k := range p.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Reset clears all provenance entries.
func (p *Provenance) Reset() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.entries = make(map[string]ProvenanceEntry)
}
