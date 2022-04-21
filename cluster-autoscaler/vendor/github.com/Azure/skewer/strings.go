package skewer

import (
	"strings"
	"unicode"
)

func normalizeLocation(input string) string {
	var output string
	for _, c := range input {
		if !unicode.IsSpace(c) {
			output += string(c)
		}
	}
	return strings.ToLower(output)
}

func locationEquals(a, b string) bool {
	return normalizeLocation(a) == normalizeLocation(b)
}
