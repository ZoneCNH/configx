package configx

import (
	"reflect"
	"strings"
)

// isSensitiveFieldName returns true if the struct field name suggests a secret
// or credential. This is broader than IsSecretKey (which is tuned for error
// message redaction) and covers common naming conventions like *Key, *Pass,
// *Credential, *Auth, etc.
func isSensitiveFieldName(name string) bool {
	if IsSecretKey(name) {
		return true
	}
	lower := strings.ToLower(name)
	sensitiveSuffixes := []string{
		"key", "pass", "credential", "auth", "private",
	}
	for _, s := range sensitiveSuffixes {
		if strings.HasSuffix(lower, s) {
			return true
		}
	}
	return false
}

// SanitizedManifest generates a safe snapshot of the configuration suitable
// for inclusion in logs, health endpoints, or CI artifacts. Fields identified
// as secrets (by field name heuristic or struct tag) have their values replaced
// with "***".
//
// The input cfg can be any struct or map value.
func SanitizedManifest(cfg any) map[string]any {
	if cfg == nil {
		return nil
	}
	rv := reflect.ValueOf(cfg)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return sanitizedStruct(rv)
	case reflect.Map:
		return sanitizedMap(rv)
	default:
		return map[string]any{"_value": cfg}
	}
}

func sanitizedStruct(rv reflect.Value) map[string]any {
	rt := rv.Type()
	out := make(map[string]any, rv.NumField())
	for i := 0; i < rv.NumField(); i++ {
		sf := rt.Field(i)
		if !sf.IsExported() {
			continue
		}
		name := sf.Tag.Get("json")
		if name == "" || name == "-" {
			name = sf.Name
		}
		if idx := strings.IndexByte(name, ','); idx >= 0 {
			name = name[:idx]
		}

		fv := rv.Field(i)
		// Determine if this field is a secret.
		isSecret := sf.Tag.Get("secret") == "true" || isSensitiveFieldName(sf.Name)
		if isSecret {
			out[name] = redactionMarker
			continue
		}
		// Recurse into nested structs.
		deref := dereference(fv)
		if deref.Kind() == reflect.Struct {
			out[name] = sanitizedStruct(deref)
			continue
		}
		out[name] = fv.Interface()
	}
	return out
}

func sanitizedMap(rv reflect.Value) map[string]any {
	out := make(map[string]any, rv.Len())
	for _, k := range rv.MapKeys() {
		key := k.String()
		val := rv.MapIndex(k)
		if IsSecretKey(key) {
			out[key] = redactionMarker
			continue
		}
		deref := dereference(val)
		if deref.Kind() == reflect.Struct {
			out[key] = sanitizedStruct(deref)
			continue
		}
		if deref.Kind() == reflect.Map {
			out[key] = sanitizedMap(deref)
			continue
		}
		out[key] = val.Interface()
	}
	return out
}

// dereference unwraps pointers to get the underlying value.
func dereference(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return v
		}
		v = v.Elem()
	}
	return v
}
