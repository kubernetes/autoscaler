package testing

import (
	"fmt"
	"math/rand"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/internal/sdkio"
)

var randBytes = func() []byte {
	rr := rand.New(rand.NewSource(0))
	b := make([]byte, 10*sdkio.MebiByte)

	if _, err := rr.Read(b); err != nil {
		panic(fmt.Sprintf("failed to read random bytes, %v", err))
	}
	return b
}()

// GetTestBytes returns a pseudo-random []byte of length size
func GetTestBytes(size int) []byte {
	if len(randBytes) >= size {
		return randBytes[:size]
	}

	b := append(randBytes, GetTestBytes(size-len(randBytes))...)
	return b
}
