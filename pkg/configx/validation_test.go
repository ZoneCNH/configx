package configx

import (
	"errors"
	"testing"
)

func TestValidate_ValidStruct(t *testing.T) {
	type cfg struct {
		Host string `required:"true"`
		Port int
	}
	report := Validate(cfg{Host: "localhost", Port: 5432})
	if !report.Valid {
		t.Errorf("expected valid, got errors: %+v", report.Errors)
	}
}

func TestValidate_MissingRequired(t *testing.T) {
	type cfg struct {
		Host string `required:"true"`
		Port int
	}
	report := Validate(cfg{Port: 5432})
	if report.Valid {
		t.Error("expected invalid for missing required field")
	}
	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(report.Errors))
	}
	if report.Errors[0].Field != "Host" {
		t.Errorf("error field = %q, want Host", report.Errors[0].Field)
	}
}

func TestValidate_NilInput(t *testing.T) {
	report := Validate(nil)
	if report.Valid {
		t.Error("expected invalid for nil input")
	}
	if len(report.Errors) != 1 || report.Errors[0].Field != "_root" {
		t.Errorf("errors = %+v", report.Errors)
	}
}

func TestValidate_NilPointer(t *testing.T) {
	type cfg struct{ Host string }
	var c *cfg
	report := Validate(c)
	if report.Valid {
		t.Error("expected invalid for nil pointer")
	}
}

func TestValidate_NonStructInput(t *testing.T) {
	report := Validate(42)
	if report.Valid {
		t.Error("expected invalid for non-struct")
	}
}

func TestValidate_RangeConstraints_Int(t *testing.T) {
	type cfg struct {
		Port int `min:"1" max:"65535"`
	}
	// Valid.
	report := Validate(cfg{Port: 5432})
	if !report.Valid {
		t.Errorf("expected valid, got %+v", report.Errors)
	}
	// Below min.
	report = Validate(cfg{Port: 0})
	if report.Valid {
		t.Error("expected invalid for port below minimum")
	}
	// Above max.
	report = Validate(cfg{Port: 70000})
	if report.Valid {
		t.Error("expected invalid for port above maximum")
	}
}

func TestValidate_RangeConstraints_Uint(t *testing.T) {
	type cfg struct {
		Workers uint `min:"1" max:"16"`
	}
	report := Validate(cfg{Workers: 0})
	if report.Valid {
		t.Error("expected invalid for 0 workers")
	}
	report = Validate(cfg{Workers: 32})
	if report.Valid {
		t.Error("expected invalid for 32 workers")
	}
	report = Validate(cfg{Workers: 4})
	if !report.Valid {
		t.Errorf("expected valid, got %+v", report.Errors)
	}
}

func TestValidate_RangeConstraints_Float(t *testing.T) {
	type cfg struct {
		Rate float64 `min:"0.0" max:"1.0"`
	}
	report := Validate(cfg{Rate: -0.1})
	if report.Valid {
		t.Error("expected invalid for negative rate")
	}
	report = Validate(cfg{Rate: 1.5})
	if report.Valid {
		t.Error("expected invalid for rate above max")
	}
	report = Validate(cfg{Rate: 0.5})
	if !report.Valid {
		t.Errorf("expected valid, got %+v", report.Errors)
	}
}

func TestValidate_RangeConstraints_String(t *testing.T) {
	type cfg struct {
		Name string `min:"3" max:"10"`
	}
	report := Validate(cfg{Name: "ab"})
	if report.Valid {
		t.Error("expected invalid for short name")
	}
	report = Validate(cfg{Name: "this is way too long"})
	if report.Valid {
		t.Error("expected invalid for long name")
	}
	report = Validate(cfg{Name: "hello"})
	if !report.Valid {
		t.Errorf("expected valid, got %+v", report.Errors)
	}
}

func TestValidate_WithValidatorInterface(t *testing.T) {
	type cfg struct {
		Host string `required:"true"`
		Port int    `required:"true"`
	}
	// Note: cfg here doesn't implement Validator, so we test with Config which does.
	c := Config{}
	report := Validate(c)
	if report.Valid {
		t.Error("expected invalid for empty Config (name required)")
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	type cfg struct {
		Host string `required:"true"`
		Port int    `min:"1" max:"65535"`
	}
	report := Validate(cfg{Port: 0})
	if report.Valid {
		t.Error("expected invalid")
	}
	if len(report.Errors) < 2 {
		t.Errorf("expected at least 2 errors, got %d: %+v", len(report.Errors), report.Errors)
	}
}

func TestValidateWithReport_Valid(t *testing.T) {
	type cfg struct {
		Host string `required:"true"`
	}
	report, err := ValidateWithReport(cfg{Host: "localhost"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !report.Valid {
		t.Error("expected valid")
	}
}

func TestValidateWithReport_Invalid(t *testing.T) {
	type cfg struct {
		Host string `required:"true"`
	}
	report, err := ValidateWithReport(cfg{})
	if err == nil {
		t.Error("expected error for invalid config")
	}
	if report.Valid {
		t.Error("expected invalid report")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Errorf("expected validation error kind, got %v", err)
	}
}

func TestFieldError_Fields(t *testing.T) {
	e := FieldError{Field: "Port", Message: "out of range", Value: -1}
	if e.Field != "Port" || e.Message != "out of range" || e.Value != -1 {
		t.Errorf("unexpected field error: %+v", e)
	}
}

func TestValidationReport_NoErrors(t *testing.T) {
	report := ValidationReport{Valid: true}
	if !report.Valid {
		t.Error("expected valid")
	}
	if report.Errors != nil {
		t.Errorf("expected nil errors, got %v", report.Errors)
	}
}

func TestValidate_ZeroValueRequiredInt(t *testing.T) {
	type cfg struct {
		Count int `required:"true"`
	}
	report := Validate(cfg{Count: 0})
	if report.Valid {
		t.Error("expected invalid for zero-value required int")
	}
}

// Ensure that errors package integration works with Error.
func TestValidationReport_ErrorIntegration(t *testing.T) {
	_, err := ValidateWithReport(struct {
		Name string `required:"true"`
	}{})
	if err == nil {
		t.Fatal("expected error")
	}
	var target *Error
	if !errors.As(err, &target) {
		t.Error("expected *Error type")
	}
	if target.Kind != ErrorKindValidation {
		t.Errorf("kind = %q, want validation", target.Kind)
	}
}
