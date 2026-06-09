package configx

import (
	"fmt"
	"reflect"
	"strings"
)

// FieldError describes a single validation failure for a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// ValidationReport is the result of validating a configuration struct.
// Valid is true only when Errors is empty.
type ValidationReport struct {
	Valid  bool         `json:"valid"`
	Errors []FieldError `json:"errors,omitempty"`
}

// Validate produces a ValidationReport from a configuration struct.
// It checks:
//   - required fields (via `required:"true"` tag) are non-zero
//   - Validator interface if implemented
//   - range constraints (via `min`/`max` tags on numeric fields)
//
// The report is always returned (even on success) so callers can serialise it.
func Validate(cfg any) ValidationReport {
	if cfg == nil {
		return ValidationReport{
			Valid:  false,
			Errors: []FieldError{{Field: "_root", Message: "cfg is nil"}},
		}
	}
	rv := reflect.ValueOf(cfg)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return ValidationReport{
				Valid:  false,
				Errors: []FieldError{{Field: "_root", Message: "cfg is nil pointer"}},
			}
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ValidationReport{
			Valid:  false,
			Errors: []FieldError{{Field: "_root", Message: "cfg must be a struct or pointer to struct"}},
		}
	}

	var errs []FieldError
	errs = append(errs, validateRequired(rv)...)
	errs = append(errs, validateRanges(rv)...)

	// If the struct implements Validator, run that too.
	if v, ok := cfg.(Validator); ok {
		if err := v.Validate(); err != nil {
			errs = append(errs, FieldError{
				Field:   "_validator",
				Message: err.Error(),
			})
		}
	}

	return ValidationReport{
		Valid:  len(errs) == 0,
		Errors: errs,
	}
}

// validateRequired checks that required fields are non-zero.
func validateRequired(rv reflect.Value) []FieldError {
	rt := rv.Type()
	var errs []FieldError
	for i := 0; i < rv.NumField(); i++ {
		sf := rt.Field(i)
		if !sf.IsExported() {
			continue
		}
		if sf.Tag.Get("required") != "true" {
			continue
		}
		fv := rv.Field(i)
		if fv.IsZero() {
			errs = append(errs, FieldError{
				Field:   sf.Name,
				Message: "required field is zero-valued",
				Value:   fv.Interface(),
			})
		}
	}
	return errs
}

// validateRanges checks min/max constraints on numeric fields.
func validateRanges(rv reflect.Value) []FieldError {
	rt := rv.Type()
	var errs []FieldError
	for i := 0; i < rv.NumField(); i++ {
		sf := rt.Field(i)
		if !sf.IsExported() {
			continue
		}
		fv := rv.Field(i)
		minTag := sf.Tag.Get("min")
		maxTag := sf.Tag.Get("max")
		if minTag == "" && maxTag == "" {
			continue
		}
		errs = append(errs, checkRange(sf.Name, fv, minTag, maxTag)...)
	}
	return errs
}

func checkRange(field string, fv reflect.Value, minTag, maxTag string) []FieldError {
	var errs []FieldError
	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val := fv.Int()
		if minTag != "" {
			var min int64
			if _, err := fmt.Sscanf(minTag, "%d", &min); err == nil && val < min {
				errs = append(errs, FieldError{
					Field:   field,
					Message: fmt.Sprintf("value %d is below minimum %d", val, min),
					Value:   val,
				})
			}
		}
		if maxTag != "" {
			var max int64
			if _, err := fmt.Sscanf(maxTag, "%d", &max); err == nil && val > max {
				errs = append(errs, FieldError{
					Field:   field,
					Message: fmt.Sprintf("value %d is above maximum %d", val, max),
					Value:   val,
				})
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val := fv.Uint()
		if minTag != "" {
			var min uint64
			if _, err := fmt.Sscanf(minTag, "%d", &min); err == nil && val < min {
				errs = append(errs, FieldError{
					Field:   field,
					Message: fmt.Sprintf("value %d is below minimum %d", val, min),
					Value:   val,
				})
			}
		}
		if maxTag != "" {
			var max uint64
			if _, err := fmt.Sscanf(maxTag, "%d", &max); err == nil && val > max {
				errs = append(errs, FieldError{
					Field:   field,
					Message: fmt.Sprintf("value %d is above maximum %d", val, max),
					Value:   val,
				})
			}
		}
	case reflect.Float32, reflect.Float64:
		val := fv.Float()
		if minTag != "" {
			var min float64
			if _, err := fmt.Sscanf(minTag, "%f", &min); err == nil && val < min {
				errs = append(errs, FieldError{
					Field:   field,
					Message: fmt.Sprintf("value %g is below minimum %g", val, min),
					Value:   val,
				})
			}
		}
		if maxTag != "" {
			var max float64
			if _, err := fmt.Sscanf(maxTag, "%f", &max); err == nil && val > max {
				errs = append(errs, FieldError{
					Field:   field,
					Message: fmt.Sprintf("value %g is above maximum %g", val, max),
					Value:   val,
				})
			}
		}
	case reflect.String:
		// For strings, min/max are interpreted as length constraints.
		val := fv.Len()
		if minTag != "" {
			var min int
			if _, err := fmt.Sscanf(minTag, "%d", &min); err == nil && val < min {
				errs = append(errs, FieldError{
					Field:   field,
					Message: fmt.Sprintf("length %d is below minimum %d", val, min),
					Value:   fv.String(),
				})
			}
		}
		if maxTag != "" {
			var max int
			if _, err := fmt.Sscanf(maxTag, "%d", &max); err == nil && val > max {
				errs = append(errs, FieldError{
					Field:   field,
					Message: fmt.Sprintf("length %d is above maximum %d", val, max),
					Value:   fv.String(),
				})
			}
		}
	}
	return errs
}

// ValidateWithReport is a convenience that wraps Validate and returns an error
// if the report is invalid, suitable for use in loading pipelines.
func ValidateWithReport(cfg any) (ValidationReport, error) {
	report := Validate(cfg)
	if !report.Valid {
		var msgs []string
		for _, e := range report.Errors {
			msgs = append(msgs, e.Field+": "+e.Message)
		}
		return report, NewError(ErrorKindValidation, "configx.ValidateWithReport",
			strings.Join(msgs, "; "), false)
	}
	return report, nil
}
