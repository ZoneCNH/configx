package configx

import (
	"testing"
)

func TestValidateNilCfg(t *testing.T) {
	report := Validate(nil)
	if report.Valid {
		t.Fatal("expected invalid for nil cfg")
	}
	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(report.Errors))
	}
	if report.Errors[0].Field != "_root" {
		t.Fatalf("field = %q", report.Errors[0].Field)
	}
}

func TestValidateNilPointerCfg(t *testing.T) {
	type cfg struct{ Name string }
	var c *cfg
	report := Validate(c)
	if report.Valid {
		t.Fatal("expected invalid for nil pointer")
	}
	if report.Errors[0].Message != "cfg is nil pointer" {
		t.Fatalf("message = %q", report.Errors[0].Message)
	}
}

func TestValidateNonStructCfg(t *testing.T) {
	report := Validate("not a struct")
	if report.Valid {
		t.Fatal("expected invalid for non-struct")
	}
}

func TestValidateRequiredFieldZero(t *testing.T) {
	type cfg struct {
		Name string `required:"true"`
		Port int
	}
	report := Validate(cfg{})
	if report.Valid {
		t.Fatal("expected invalid for missing required field")
	}
	found := false
	for _, e := range report.Errors {
		if e.Field == "Name" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Name error in %v", report.Errors)
	}
}

func TestValidateRequiredFieldPresent(t *testing.T) {
	type cfg struct {
		Name string `required:"true"`
	}
	report := Validate(cfg{Name: "test"})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateRangeChecksInt(t *testing.T) {
	type cfg struct {
		Port int `min:"1" max:"65535"`
	}
	report := Validate(cfg{Port: 0})
	if report.Valid {
		t.Fatal("expected invalid for port below min")
	}

	report = Validate(cfg{Port: 70000})
	if report.Valid {
		t.Fatal("expected invalid for port above max")
	}

	report = Validate(cfg{Port: 8080})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateRangeChecksUint(t *testing.T) {
	type cfg struct {
		Count uint `min:"1" max:"100"`
	}
	report := Validate(cfg{Count: 0})
	if report.Valid {
		t.Fatal("expected invalid for count below min")
	}

	report = Validate(cfg{Count: 200})
	if report.Valid {
		t.Fatal("expected invalid for count above max")
	}
}

func TestValidateRangeChecksFloat(t *testing.T) {
	type cfg struct {
		Rate float64 `min:"0.0" max:"1.0"`
	}
	report := Validate(cfg{Rate: -0.5})
	if report.Valid {
		t.Fatal("expected invalid for rate below min")
	}

	report = Validate(cfg{Rate: 1.5})
	if report.Valid {
		t.Fatal("expected invalid for rate above max")
	}
}

func TestValidateRangeChecksString(t *testing.T) {
	type cfg struct {
		Name string `min:"3" max:"10"`
	}
	report := Validate(cfg{Name: "ab"})
	if report.Valid {
		t.Fatal("expected invalid for name too short")
	}

	report = Validate(cfg{Name: "this is way too long string"})
	if report.Valid {
		t.Fatal("expected invalid for name too long")
	}

	report = Validate(cfg{Name: "hello"})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateWithReportValid(t *testing.T) {
	type cfg struct {
		Name string `required:"true"`
	}
	report, err := ValidateWithReport(cfg{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Valid {
		t.Fatal("expected valid report")
	}
}

func TestValidateWithReportInvalid(t *testing.T) {
	type cfg struct {
		Name string `required:"true"`
	}
	report, err := ValidateWithReport(cfg{})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
	if report.Valid {
		t.Fatal("expected invalid report")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestValidateWithValidatorInterface(t *testing.T) {
	type cfg struct {
		Port int `required:"true"`
	}
	report := Validate(cfg{Port: 8080})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateRangeNoMinMax(t *testing.T) {
	type cfg struct {
		Name string
	}
	report := Validate(cfg{Name: "test"})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateRangeInvalidMinMaxTag(t *testing.T) {
	type cfg struct {
		Port int `min:"abc" max:"xyz"`
	}
	report := Validate(cfg{Port: 8080})
	if !report.Valid {
		t.Fatalf("expected valid with invalid min/max tags (graceful), got %v", report.Errors)
	}
}

func TestCheckRangeUnsupportedType(t *testing.T) {
	type cfg struct {
		Flag bool `min:"1"`
	}
	report := Validate(cfg{Flag: true})
	if !report.Valid {
		t.Fatalf("expected valid (bool not checked for range), got %v", report.Errors)
	}
}

func TestValidateSkipsUnexportedFields(t *testing.T) {
	type cfg struct {
		Name    string `required:"true"`
		private string //nolint:unused
	}
	report := Validate(cfg{Name: "test"})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateStringRangeNoTags(t *testing.T) {
	type cfg struct {
		Name string
	}
	report := Validate(cfg{Name: "anything"})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateUintRangeValid(t *testing.T) {
	type cfg struct {
		Port uint `min:"1" max:"65535"`
	}
	report := Validate(cfg{Port: 8080})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateFloatRangeValid(t *testing.T) {
	type cfg struct {
		Rate float64 `min:"0.0" max:"1.0"`
	}
	report := Validate(cfg{Rate: 0.5})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateStringRangeValid(t *testing.T) {
	type cfg struct {
		Name string `min:"3" max:"10"`
	}
	report := Validate(cfg{Name: "hello"})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateIntRangeOnlyMin(t *testing.T) {
	type cfg struct {
		Port int `min:"1"`
	}
	report := Validate(cfg{Port: 0})
	if report.Valid {
		t.Fatal("expected invalid for port below min")
	}
	report = Validate(cfg{Port: 99999})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}

func TestValidateIntRangeOnlyMax(t *testing.T) {
	type cfg struct {
		Port int `max:"65535"`
	}
	report := Validate(cfg{Port: 70000})
	if report.Valid {
		t.Fatal("expected invalid for port above max")
	}
	report = Validate(cfg{Port: 8080})
	if !report.Valid {
		t.Fatalf("expected valid, got %v", report.Errors)
	}
}
