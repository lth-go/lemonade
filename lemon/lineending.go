package lemon

import (
	"regexp"
	"strings"
)

var (
	crlfStep1 = regexp.MustCompile(`\r(.)|\r$`)
	crlfStep2 = regexp.MustCompile(`([^\r])\n|^\n`)
)

// ConvertLineEnding normalizes the line endings of text according to option.
//
// option values (case-insensitive):
//   - "lf":   CRLF/CR -> LF
//   - "crlf": LF/CR   -> CRLF
//   - "":     unchanged
func ConvertLineEnding(text, option string) string {
	switch strings.ToLower(option) {
	case "lf":
		text = strings.ReplaceAll(text, "\r\n", "\n")
		return strings.ReplaceAll(text, "\r", "\n")
	case "crlf":
		text = crlfStep1.ReplaceAllString(text, "\r\n$1")
		return crlfStep2.ReplaceAllString(text, "$1\r\n")
	default:
		return text
	}
}
