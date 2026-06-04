package configx

import (
	"errors"
	"strings"
)

// IsSecretKey returns true if the key name suggests it contains a secret value.
func IsSecretKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "secret") || strings.Contains(k, "password") || strings.Contains(k, "passwd") || strings.Contains(k, "token") || strings.Contains(k, "access_key") || strings.Contains(k, "secret_key")
}

func sanitizeError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(sanitizeMessage(err.Error()))
}

func sanitizeMessage(message string) string {
	parts := strings.FieldsFunc(message, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ',' || r == '&' || r == ';'
	})
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok || value == "" || !IsSecretKey(strings.Trim(key, `"'`)) {
			continue
		}
		message = strings.ReplaceAll(message, part, key+"="+redactionMarker)
	}
	return message
}
