//go:build !go1.6
// +build !go1.6

package s3

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/request"

func platformRequestHandlers(r *request.Request) {
}
