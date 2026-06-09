package configx

import (
	"testing"
)

func TestStrictDecodeConfigNilPointer(t *testing.T) {
	var p *struct{ Name string }
	err := StrictDecodeConfig([]byte(`{"Name":"test"}`), p)
	if err == nil {
		t.Fatal("expected error for nil pointer target")
	}
}

func TestCheckMaxDepthEscapeInString(t *testing.T) {
	// JSON with braces inside a string should not count
	data := []byte(`{"key":"value with { and } inside"}`)
	var m map[string]any
	err := StrictDecode(data, &m, WithMaxDepth(1))
	if err != nil {
		t.Fatalf("braces inside string should not count: %v", err)
	}
}

func TestCheckMaxDepthArrayNesting(t *testing.T) {
	data := []byte(`[[1,2],[3,4]]`)
	var m any
	err := StrictDecode(data, &m, WithMaxDepth(1))
	if err == nil {
		t.Fatal("expected depth error for nested arrays")
	}
}

func TestCheckMaxDepthEscapedQuote(t *testing.T) {
	// String with escaped quote: {"key":"val\"ue"}
	data := []byte(`{"key":"val\"ue"}`)
	var m map[string]any
	err := StrictDecode(data, &m, WithMaxDepth(1))
	if err != nil {
		t.Fatalf("escaped quote in string should not break parser: %v", err)
	}
}

func TestStrictDecodeEmptyObject(t *testing.T) {
	data := []byte(`{}`)
	var m map[string]any
	err := StrictDecode(data, &m)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStrictDecodeEmptyArray(t *testing.T) {
	data := []byte(`[]`)
	var s []any
	err := StrictDecode(data, &s)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckDuplicateKeysNoObject(t *testing.T) {
	// A primitive at the top level — walkJSON sees a non-delimiter token
	err := checkDuplicateKeys([]byte(`"just a string"`))
	if err != nil {
		t.Fatalf("primitive should have no duplicate key issue: %v", err)
	}
}
