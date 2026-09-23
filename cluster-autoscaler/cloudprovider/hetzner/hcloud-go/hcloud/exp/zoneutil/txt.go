package zoneutil

import (
	"fmt"
	"slices"
	"strings"
)

// IsTXTRecordQuoted returns wether a the given string starts and ends with quotes.
//
// - hello world	=> false
// - "hello world"	=> true
func IsTXTRecordQuoted(value string) bool {
	return strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)
}

// FormatTXTRecord splits a long string in chunks of 255 characters, and quotes each of
// the chunks. Existing quotes will be escaped.
//
// - hello world	=> "hello world"
// - hello "world"	=> "hello \"world\""
//
// This function can be reversed with [ParseTXTValue].
func FormatTXTRecord(value string) string {
	// Escape existing quotes
	value = strings.ReplaceAll(value, "\"", "\\\"")

	// Split in chunks of 255 characters
	parts := []string{}
	for chunk := range slices.Chunk([]byte(value), 255) {
		parts = append(parts, fmt.Sprintf(`"%s"`, chunk))
	}
	return strings.Join(parts, " ")
}

// ParseTXTRecord joins multiple quoted strings into a single unquoted string.
//
// - "hello world"		=> hello world
// - "hello \"world\""	=> hello "world"
func ParseTXTRecord(value string) string {
	return strings.Join(parseTXTStrings(value), "")
}

// parseTXTStrings splits the given string at whitespaces not escaped by double quotation marks:
//   - "hello" "world" -> []string{"hello", "world"}
//   - "hello" world   -> []string{"hello", "world"}
//   - hello "world"   -> []string{"hello", "world"}
//   - hello world     -> []string{"hello", "world"}
//
// Double quotation marks escaped with \" are ignored:
//   - hello\" world   -> []string{"hello \"", "world"}
//   - hello wo\"rld   -> []string{"hello", "wo\"rld"}
func parseTXTStrings(value string) []string {
	var result []string

	var cur strings.Builder
	var quoted, escapeNext bool

	for _, c := range value {
		if escapeNext {
			// This character is escaped by a previous \.
			cur.WriteRune(c)
			escapeNext = false
			continue
		}

		switch c {
		case '\\':
			// Next character is escaped.
			escapeNext = true
		case '"':
			quoted = !quoted
		case ' ':
			if quoted {
				// Inside double quotation marks, add character to current string.
				cur.WriteRune(c)
			} else if cur.Len() > 0 {
				// Whitespace after ending quotation mark, end of current string.
				result = append(result, cur.String())
				cur.Reset()
			}
		default:
			if quoted {
				cur.WriteRune(c)
			}
		}
	}

	// Remaining characters are last string.
	if cur.Len() > 0 {
		result = append(result, cur.String())
	}

	return result
}
