package redact

import (
	"regexp"
	"strings"
)

const Replacement = "[REDACTED]"

var (
	bearerPattern  = regexp.MustCompile(`(?i)(authorization\s*[:=]\s*bearer\s+)[^\s,"}]+`)
	jsonKeyPattern = regexp.MustCompile(`(?i)("(?:api_?key|token|secret|password)"\s*:\s*")[^"]*(")`)
	keyPattern     = regexp.MustCompile(`\blago_(?:test_|live_)?[A-Za-z0-9_-]{12,}\b`)
)

type Redactor struct {
	secrets []string
}

func New(secrets ...string) *Redactor {
	r := &Redactor{}
	for _, secret := range secrets {
		if strings.TrimSpace(secret) != "" {
			r.secrets = append(r.secrets, secret)
		}
	}
	return r
}

func (r *Redactor) String(value string) string {
	for _, secret := range r.secrets {
		value = strings.ReplaceAll(value, secret, Replacement)
	}
	value = bearerPattern.ReplaceAllString(value, `${1}`+Replacement)
	value = jsonKeyPattern.ReplaceAllString(value, `${1}`+Replacement+`${2}`)
	return keyPattern.ReplaceAllString(value, Replacement)
}
