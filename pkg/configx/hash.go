package configx

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// VolatileFieldNames lists struct field names that are excluded from the hash
// computation because they change between loads without reflecting actual
// configuration changes (e.g., timestamps, load metadata).
var VolatileFieldNames = []string{
	"LoadedAt",
	"Timestamp",
	"UpdatedAt",
	"CreatedAt",
	"Timestamps",
}

// EffectiveConfigHash computes a SHA-256 hex digest of the configuration value.
// Volatile fields (timestamps, load metadata) are stripped before hashing so
// that the fingerprint only reflects meaningful configuration state.
//
// The input cfg can be:
//   - a struct (volatile-tagged or named fields are excluded)
//   - a map[string]any (volatile keys are excluded)
//   - a LoadResult (values are hashed by key sorted alphabetically)
//
// Returns the hex-encoded SHA-256 string or an error if cfg cannot be
// marshalled.
func EffectiveConfigHash(cfg any) (string, error) {
	if cfg == nil {
		return "", NewError(ErrorKindValidation, "configx.EffectiveConfigHash", "cfg is nil", false)
	}
	sorted, err := canonicalJSON(cfg)
	if err != nil {
		return "", WrapError(ErrorKindInternal, "configx.EffectiveConfigHash", "canonical json failed", false, err)
	}
	sum := sha256.Sum256(sorted)
	return fmt.Sprintf("%x", sum), nil
}

// canonicalJSON produces a deterministic JSON representation of cfg with
// volatile fields removed and map keys sorted.
func canonicalJSON(cfg any) ([]byte, error) {
	switch v := cfg.(type) {
	case LoadResult:
		return marshalLoadResult(v)
	default:
		return marshalGeneric(cfg)
	}
}

// marshalLoadResult hashes a LoadResult by sorting keys and excluding
// volatile Value fields.
func marshalLoadResult(r LoadResult) ([]byte, error) {
	keys := make([]string, 0, len(r.Values))
	for k := range r.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	type entry struct {
		Key    string `json:"key"`
		Value  string `json:"value"`
		Secret bool   `json:"secret"`
		Source string `json:"source"`
	}
	entries := make([]entry, 0, len(keys))
	for _, k := range keys {
		v := r.Values[k]
		entries = append(entries, entry{
			Key:    v.Key,
			Value:  v.Value,
			Secret: v.Secret,
			Source: v.Source,
		})
	}
	return json.Marshal(entries)
}

// marshalGeneric handles structs and maps by stripping volatile fields.
func marshalGeneric(cfg any) ([]byte, error) {
	rv := reflect.ValueOf(cfg)
	// Dereference pointer
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return []byte("null"), nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return marshalStruct(rv)
	case reflect.Map:
		return marshalMap(rv)
	default:
		return json.Marshal(cfg)
	}
}

func marshalStruct(rv reflect.Value) ([]byte, error) {
	rt := rv.Type()
	volatileSet := make(map[string]struct{}, len(VolatileFieldNames))
	for _, name := range VolatileFieldNames {
		volatileSet[name] = struct{}{}
	}
	// Also honor `volatile:"true"` struct tag.
	type kv struct {
		key string
		val any
	}
	var fields []kv
	for i := 0; i < rv.NumField(); i++ {
		sf := rt.Field(i)
		if !sf.IsExported() {
			continue
		}
		if _, skip := volatileSet[sf.Name]; skip {
			continue
		}
		if sf.Tag.Get("volatile") == "true" {
			continue
		}
		name := sf.Tag.Get("json")
		if name == "" || name == "-" {
			name = sf.Name
		}
		// Strip omitempty etc.
		if idx := strings.IndexByte(name, ','); idx >= 0 {
			name = name[:idx]
		}
		fields = append(fields, kv{key: name, val: rv.Field(i).Interface()})
	}
	// Sort by key for determinism.
	sort.Slice(fields, func(i, j int) bool { return fields[i].key < fields[j].key })
	m := make(map[string]any, len(fields))
	for _, f := range fields {
		m[f.key] = f.val
	}
	return json.Marshal(m)
}

func marshalMap(rv reflect.Value) ([]byte, error) {
	volatileSet := make(map[string]struct{}, len(VolatileFieldNames))
	for _, name := range VolatileFieldNames {
		volatileSet[name] = struct{}{}
		volatileSet[strings.ToLower(name)] = struct{}{}
	}
	keys := rv.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})
	type kv struct {
		key string
		val any
	}
	var fields []kv
	for _, k := range keys {
		ks := k.String()
		if _, skip := volatileSet[ks]; skip {
			continue
		}
		fields = append(fields, kv{key: ks, val: rv.MapIndex(k).Interface()})
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].key < fields[j].key })
	m := make(map[string]any, len(fields))
	for _, f := range fields {
		m[f.key] = f.val
	}
	return json.Marshal(m)
}
