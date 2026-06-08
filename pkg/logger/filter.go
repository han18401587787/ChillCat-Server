package logger

import "strings"

var sensitiveFields = []string{"password", "token", "secret", "authorization", "api_key", "apikey"}

// Sanitize masks sensitive field values in a log message
func Sanitize(msg string) string {
	for _, field := range sensitiveFields {
		msg = maskJSONField(msg, field)
	}
	return msg
}

func maskJSONField(msg, field string) string {
	// JSON pattern: "field":"value" or "field": "value"
	lower := strings.ToLower(msg)
	idx := 0
	for {
		pos := strings.Index(lower[idx:], `"`+field+`"`)
		if pos == -1 {
			break
		}
		abs := idx + pos + len(field) + 2
		// Skip colon and optional whitespace
		for abs < len(msg) && (msg[abs] == ':' || msg[abs] == ' ' || msg[abs] == '"') {
			if msg[abs] == '"' {
				abs++
				// Find closing quote
				end := strings.Index(msg[abs:], `"`)
				if end == -1 {
					return msg
				}
				msg = msg[:abs] + "***" + msg[abs+end:]
				lower = strings.ToLower(msg)
				break
			}
			abs++
		}
		idx = abs
	}
	return msg
}
