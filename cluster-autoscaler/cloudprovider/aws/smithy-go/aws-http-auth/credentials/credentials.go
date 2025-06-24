// Package credentials exposes container types for AWS credentials.
package credentials

import (
	"time"
)

// Credentials describes a shared-secret AWS credential identity.
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expires         time.Time
}
