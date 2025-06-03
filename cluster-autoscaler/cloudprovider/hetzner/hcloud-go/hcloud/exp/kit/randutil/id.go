package randutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateID returns a hex encoded random string with a len of 8 chars similar to
// "2873fce7".
func GenerateID() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		// Should never happen as of go1.24: https://github.com/golang/go/issues/66821
		panic(fmt.Errorf("failed to generate random string: %w", err))
	}
	return hex.EncodeToString(b)
}
