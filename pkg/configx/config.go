package configx

import (
	"errors"
	"fmt"
	"time"
)

// Config holds the configuration for a configx client.
type Config struct {
	Name    string
	Timeout time.Duration
	Secret  string
}

// SanitizedConfig is a Config with the secret masked.
type SanitizedConfig struct {
	Name    string
	Timeout time.Duration
	Secret  string
}

// Validate validates the Config fields.
func (c Config) Validate() error {
	if err := requireNonEmpty("name", c.Name); err != nil {
		return validationError("Config.Validate", err.Error(), err)
	}
	if c.Timeout < 0 {
		err := errors.New("timeout must not be negative")
		return validationError("Config.Validate", err.Error(), err)
	}
	return nil
}

// Sanitize returns a copy of the Config with the secret masked.
func (c Config) Sanitize() SanitizedConfig {
	return SanitizedConfig{
		Name:    c.Name,
		Timeout: c.Timeout,
		Secret:  sanitizeSecret(c.Secret),
	}
}

// sanitizeSecret masks a secret value, returning "***" for non-empty values.
func sanitizeSecret(value string) string {
	if value == "" {
		return ""
	}
	return "***"
}

// requireNonEmpty returns an error if the value is empty.
func requireNonEmpty(field string, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}
