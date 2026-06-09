package testkit

import (
	"errors"
	"testing"
)

func TestRequireNoError_NilError(t *testing.T) {
	RequireNoError(t, nil)
}

// mockTB is a minimal testing.TB that records Fatalf calls.
type mockTB struct {
	testing.TB
	fatalfMsg string
	called    bool
}

func (m *mockTB) Helper() {}
func (m *mockTB) Fatalf(format string, args ...interface{}) {
	m.called = true
	m.fatalfMsg = format
}

func TestRequireNoError_NonNilError(t *testing.T) {
	mock := &mockTB{}
	RequireNoError(mock, errors.New("test error"))
	if !mock.called {
		t.Fatal("expected Fatalf to be called")
	}
}
