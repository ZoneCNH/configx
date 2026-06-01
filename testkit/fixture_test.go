package testkit

import (
	"testing"
	"time"
)

func TestConfigBuildsValidFixture(t *testing.T) {
	cfg := Config("fixture")
	if cfg.Name != "fixture" {
		t.Fatalf("unexpected name: %q", cfg.Name)
	}
	if cfg.Timeout != time.Second {
		t.Fatalf("unexpected timeout: %s", cfg.Timeout)
	}
	RequireNoError(t, cfg.Validate())
}

func TestRequireNoErrorAcceptsNil(t *testing.T) {
	RequireNoError(t, nil)
}
