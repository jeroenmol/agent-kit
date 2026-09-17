// Package evidence writes the bounded, sanitized evidence needed to describe a
// run outcome. State publication remains the caller's responsibility.
package evidence

import "strings"

const redactedValue = "[REDACTED]"

// Redactor removes exact secret values before data reaches durable storage.
// Empty values are ignored so they cannot alter ordinary text.
type Redactor struct {
	secrets []string
}

// NewRedactor returns a redactor for the supplied secret values.
func NewRedactor(secrets ...string) Redactor {
	values := make([]string, 0, len(secrets))
	for _, secret := range secrets {
		if secret != "" {
			values = append(values, secret)
		}
	}
	return Redactor{secrets: values}
}

// Redact replaces each configured secret with a stable marker.
func (r Redactor) Redact(value string) string {
	for _, secret := range r.secrets {
		value = strings.ReplaceAll(value, secret, redactedValue)
	}
	return value
}
