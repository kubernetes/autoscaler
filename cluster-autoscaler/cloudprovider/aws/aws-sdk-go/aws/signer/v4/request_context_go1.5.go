//go:build !go1.7
// +build !go1.7

package v4

import (
	"net/http"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
)

func requestContext(r *http.Request) aws.Context {
	return aws.BackgroundContext()
}
