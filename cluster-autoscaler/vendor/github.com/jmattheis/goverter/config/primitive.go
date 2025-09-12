package config

import (
	"fmt"
	"strings"
)

func parseCommand(value string) (string, string) {
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func parseEnum(name string, empty bool, remaining string, values ...string) (string, error) {
	fields := strings.Fields(remaining)

	switch {
	case len(fields) == 0 && empty:
		return "", nil
	case len(fields) == 1:
		for _, value := range values {
			if fields[0] == value {
				return value, nil
			}
		}

		return "", fmt.Errorf("invalid %s value: '%s' must be one of '%s'", name, fields[0], strings.Join(values, "', '"))
	default:
		return "", fmt.Errorf("invalid %s value: expected one value but got %d: %s", name, len(fields), fields)
	}
}

func parseBool(remaining string) (bool, error) {
	val, err := parseEnum("boolean", true, remaining, "yes", "no")
	return val == "" || val == "yes", err
}

func parseString(remaining string) (string, error) {
	fields := strings.Fields(remaining)
	if len(fields) != 1 {
		return "", fmt.Errorf("must have one value but got %d: %#v", len(fields), remaining)
	}
	return fields[0], nil
}
