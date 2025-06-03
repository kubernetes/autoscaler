package envutil

import (
	"fmt"
	"os"
	"strings"
)

// LookupEnvWithFile retrieves the value of the environment variable named by the key (e.g.
// HCLOUD_TOKEN). If the previous environment variable is not set, it retrieves the
// content of the file located by a second environment variable named by the key +
// '_FILE' (.e.g HCLOUD_TOKEN_FILE).
//
// For both cases, the returned value may be empty.
//
// The value from the environment takes precedence over the value from the file.
func LookupEnvWithFile(key string) (string, error) {
	// Check if the value is set in the environment (e.g. HCLOUD_TOKEN)
	value, ok := os.LookupEnv(key)
	if ok {
		return value, nil
	}

	key += "_FILE"

	// Check if the value is set via a file (e.g. HCLOUD_TOKEN_FILE)
	valueFile, ok := os.LookupEnv(key)
	if !ok {
		// Validation of the value happens outside of this function
		return "", nil
	}

	// Read the content of the file
	valueBytes, err := os.ReadFile(valueFile)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", key, err)
	}

	return strings.TrimSpace(string(valueBytes)), nil
}
