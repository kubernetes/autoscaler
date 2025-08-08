package integrationtest

import (
	"context"
	"crypto/rand"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
	"io"
	"log"
	"os"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/config"
)

// LoadConfigWithDefaultRegion loads the default configuration for the SDK, and
// falls back to a default region if one is not specified.
func LoadConfigWithDefaultRegion(defaultRegion string) (cfg aws.Config, err error) {
	var lm aws.ClientLogMode

	if strings.EqualFold(os.Getenv("AWS_DEBUG_REQUEST"), "true") {
		lm |= aws.LogRequest

	} else if strings.EqualFold(os.Getenv("AWS_DEBUG_REQUEST_BODY"), "true") {
		lm |= aws.LogRequestWithBody
	}

	cfg, err = config.LoadDefaultConfig(context.Background(),
		config.WithClientLogMode(lm),
		config.WithAPIOptions([]func(*middleware.Stack) error{
			RemoveOperationInputValidationMiddleware,
		}),
		config.WithDefaultRegion(defaultRegion),
	)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

type logger struct{}

func (logger) Logf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// UniqueID returns a unique UUID-like identifier for use in generating
// resources for integration tests.
func UniqueID() string {
	uuid := make([]byte, 16)
	io.ReadFull(rand.Reader, uuid)
	return fmt.Sprintf("%x", uuid)
}
