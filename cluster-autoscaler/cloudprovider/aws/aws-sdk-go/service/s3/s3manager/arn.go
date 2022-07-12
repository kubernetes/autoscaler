package s3manager

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/arn"
)

func validateSupportedARNType(bucket string) error {
	if !arn.IsARN(bucket) {
		return nil
	}

	parsedARN, err := arn.Parse(bucket)
	if err != nil {
		return err
	}

	if parsedARN.Service == "s3-object-lambda" {
		return fmt.Errorf("manager does not support s3-object-lambda service ARNs")
	}

	return nil
}
