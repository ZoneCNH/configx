package testkit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRequireGolden_MatchesContent(t *testing.T) {
	dir := t.TempDir()
	goldenPath := filepath.Join(dir, "golden.txt")
	content := []byte("expected content\n")
	if err := os.WriteFile(goldenPath, content, 0o644); err != nil {
		t.Fatalf("write golden file: %v", err)
	}
	RequireGolden(t, goldenPath, content)
}

func TestRequireGolden_MatchesContentWithWindowsNewlines(t *testing.T) {
	dir := t.TempDir()
	goldenPath := filepath.Join(dir, "golden.txt")
	content := []byte("expected content\n")
	if err := os.WriteFile(goldenPath, content, 0o644); err != nil {
		t.Fatalf("write golden file: %v", err)
	}
	RequireGolden(t, goldenPath, content)
}

func TestRequireGolden_MismatchCallsFatalf(t *testing.T) {
	dir := t.TempDir()
	goldenPath := filepath.Join(dir, "golden.txt")
	if err := os.WriteFile(goldenPath, []byte("expected"), 0o644); err != nil {
		t.Fatalf("write golden file: %v", err)
	}
	mock := &mockTB{}
	RequireGolden(mock, goldenPath, []byte("actual"))
	if !mock.called {
		t.Fatal("expected Fatalf to be called for mismatch")
	}
}

func TestRequireGolden_MissingFileCallsFatalf(t *testing.T) {
	mock := &mockTB{}
	RequireGolden(mock, "/nonexistent/golden.txt", []byte("actual"))
	if !mock.called {
		t.Fatal("expected Fatalf to be called for missing file")
	}
}
