package foundationx

import (
	"fmt"
	"strings"
	"testing"
)

func TestSecretStringCompatibility(t *testing.T) {
	var _ Sanitizer = SecretString("")

	raw := "super-secret"
	secret := NewSecretString(raw)

	if got := secret.Reveal(); got != raw {
		t.Fatalf("Reveal() = %q, want raw secret", got)
	}
	if got := secret.String(); got != redacted {
		t.Fatalf("String() = %q, want redacted", got)
	}
	if got := secret.Sanitize(); got != redacted {
		t.Fatalf("Sanitize() = %v, want redacted", got)
	}
	if secret.IsZero() {
		t.Fatal("non-empty secret must not be zero")
	}
	if !NewSecretString("").IsZero() {
		t.Fatal("empty secret must be zero")
	}
	if strings.Contains(fmt.Sprint(secret), raw) {
		t.Fatal("fmt.Sprint leaked the raw secret")
	}
}
