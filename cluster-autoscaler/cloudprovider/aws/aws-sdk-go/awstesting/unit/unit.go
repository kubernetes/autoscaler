// Package unit performs initialization and validation for unit tests
package unit

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/credentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/session"
)

// Session is a shared session for unit tests to use.
var Session = session.Must(session.NewSession(&aws.Config{
	Credentials: credentials.NewStaticCredentials("AKID", "SECRET", "SESSION"),
	Region:      aws.String("mock-region"),
	SleepDelay:  func(time.Duration) {},
}))
