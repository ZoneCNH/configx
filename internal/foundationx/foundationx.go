// Package foundationx is a local compatibility surface used while the upstream
// module is unavailable to this repository checkout.
package foundationx

import (
	"encoding/json"
	"fmt"
)

const redacted = "***"

type SecretString string

func NewSecretString(value string) SecretString { return SecretString(value) }
func (s SecretString) Reveal() string           { return string(s) }
func (s SecretString) String() string {
	if s == "" {
		return ""
	}
	return redacted
}
func (s SecretString) GoString() string             { return s.String() }
func (s SecretString) MarshalText() ([]byte, error) { return []byte(s.String()), nil }
func (s SecretString) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }

type ErrorKind string

const (
	ErrorKindConfig     ErrorKind = "config"
	ErrorKindValidation ErrorKind = "validation"
	ErrorKindInternal   ErrorKind = "internal"
)

type Error struct {
	Kind    ErrorKind
	Op      string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Op == "" {
		return fmt.Sprintf("%s: %s", e.Kind, e.Message)
	}
	return fmt.Sprintf("%s: %s: %s", e.Kind, e.Op, e.Message)
}
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}
func NewError(kind ErrorKind, op, message string, cause error) *Error {
	return &Error{Kind: kind, Op: op, Message: message, Cause: cause}
}
