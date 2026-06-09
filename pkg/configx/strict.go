package configx

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// StrictOption configures strict decode behavior.
type StrictOption func(*strictOptions)

type strictOptions struct {
	allowUnknownFields bool
	maxDepth           int
}

// WithAllowUnknownFields allows unknown fields in the input data
// instead of returning an error. By default, unknown fields cause a fail-fast error.
func WithAllowUnknownFields() StrictOption {
	return func(o *strictOptions) { o.allowUnknownFields = true }
}

// WithMaxDepth sets the maximum nesting depth for the input data.
// A depth of 0 means no limit. The default is no limit.
func WithMaxDepth(n int) StrictOption {
	return func(o *strictOptions) { o.maxDepth = n }
}

// StrictDecode decodes JSON data into the target struct with strict validation:
//   - Unknown fields cause an error (unless WithAllowUnknownFields is set)
//   - Duplicate keys cause an error
//   - Type mismatches cause an error
//   - Nesting depth is enforced (if WithMaxDepth is set)
//
// It wraps standard json.Decoder with UseNumber and DisallowUnknownFields.
func StrictDecode(data []byte, target any, opts ...StrictOption) error {
	const op = "configx.StrictDecode"

	if len(data) == 0 {
		return validationError(op, "data must not be empty", nil)
	}
	if target == nil {
		return validationError(op, "target must not be nil", nil)
	}

	options := strictOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	// Check for duplicate keys before decoding.
	if err := checkDuplicateKeys(data); err != nil {
		return validationError(op, err.Error(), err)
	}

	// Check max depth if configured.
	if options.maxDepth > 0 {
		if err := checkMaxDepth(data, options.maxDepth); err != nil {
			return validationError(op, err.Error(), err)
		}
	}

	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()

	if !options.allowUnknownFields {
		dec.DisallowUnknownFields()
	}

	if err := dec.Decode(target); err != nil {
		return validationError(op, "strict decode failed", err)
	}

	// Check for trailing data.
	if dec.More() {
		return validationError(op, "unexpected trailing data after JSON value", nil)
	}

	return nil
}

// checkDuplicateKeys scans raw JSON for duplicate keys at any nesting level.
func checkDuplicateKeys(data []byte) error {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	return walkJSON(dec, nil)
}

// walkJSON walks the JSON token stream and checks for duplicate keys.
func walkJSON(dec *json.Decoder, seen map[string]bool) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}

	delim, ok := tok.(json.Delim)
	if !ok {
		// Primitive value, nothing to check.
		return nil
	}

	switch delim {
	case '{':
		// Object: check each key.
		localSeen := make(map[string]bool, 8)
		for dec.More() {
			keyTok, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyTok.(string)
			if !ok {
				return fmt.Errorf("expected string key, got %T", keyTok)
			}
			if localSeen[key] {
				return fmt.Errorf("duplicate key %q in JSON object", key)
			}
			localSeen[key] = true
			// Recurse into value.
			if err := walkJSON(dec, localSeen); err != nil {
				return err
			}
		}
		// Consume closing delimiter.
		if _, err := dec.Token(); err != nil {
			return err
		}
	case '[':
		// Array: walk each element.
		for dec.More() {
			if err := walkJSON(dec, seen); err != nil {
				return err
			}
		}
		if _, err := dec.Token(); err != nil {
			return err
		}
	}
	return nil
}

// checkMaxDepth checks that the JSON does not exceed the given nesting depth.
func checkMaxDepth(data []byte, maxDepth int) error {
	depth := 0
	maxSeen := 0
	inString := false
	escaped := false

	for i := 0; i < len(data); i++ {
		b := data[i]

		if escaped {
			escaped = false
			continue
		}

		if b == '\\' && inString {
			escaped = true
			continue
		}

		if b == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch b {
		case '{', '[':
			depth++
			if depth > maxSeen {
				maxSeen = depth
			}
			if depth > maxDepth {
				return fmt.Errorf("JSON nesting depth %d exceeds maximum %d", depth, maxDepth)
			}
		case '}', ']':
			depth--
		}
	}

	return nil
}

// StrictDecodeConfig is a convenience function that decodes JSON data into a
// config struct and then validates it using the Validator interface if implemented.
func StrictDecodeConfig(data []byte, target any, opts ...StrictOption) error {
	const op = "configx.StrictDecodeConfig"

	if err := StrictDecode(data, target, opts...); err != nil {
		return err
	}

	rv := reflect.ValueOf(target)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return validationError(op, "target must not be nil", nil)
		}
		rv = rv.Elem()
	}

	if v, ok := rv.Interface().(Validator); ok {
		if err := v.Validate(); err != nil {
			return validationError(op, "validation failed after strict decode", err)
		}
	}

	return nil
}
