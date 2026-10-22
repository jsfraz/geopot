package utils

import (
	"strings"
	"unicode/utf8"
)

// Maximum number of runes stored or logged for a captured username/password.
const MaxCredentialLength = 1024

// Sanitizes attacker-controlled credentials before they are logged or stored.
// It replaces invalid UTF-8, strips control characters (which would allow log
// forging or break text handling) and truncates very long input.
//
//	@param value
//	@return string
func SanitizeCredential(value string) string {
	value = strings.ToValidUTF8(value, "\uFFFD")
	value = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, value)

	if utf8.RuneCountInString(value) > MaxCredentialLength {
		runes := []rune(value)
		value = string(runes[:MaxCredentialLength])
	}
	return value
}
