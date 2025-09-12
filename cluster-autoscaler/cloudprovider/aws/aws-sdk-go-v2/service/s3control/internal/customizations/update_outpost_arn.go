package customizations

import (
	"context"
	"fmt"

	awsarn "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/arn"
	s3arn "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/internal/s3shared/arn"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
)

type updateOutpostARN struct {
}

func (*updateOutpostARN) ID() string {
	return "setArnFieldName"
}

// updateOutpostARN handles updating the relevant operation member
// whose value is an S3 Outposts ARN provided by the customer.
func (m *updateOutpostARN) HandleSerialize(
	ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler,
) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	v, ok := s3arn.GetARNField(in.Parameters)
	if ok && awsarn.IsARN(*v) {

		av, err := awsarn.Parse(*v)
		if err != nil {
			return out, metadata, fmt.Errorf("error parsing arn: %w", err)
		}
		resource, err := s3arn.ParseResource(av, resourceParser)
		if err != nil {
			return out, metadata, err
		}

		switch tv := resource.(type) {
		case s3arn.OutpostAccessPointARN:
			s3arn.SetARNField(in.Parameters, tv.AccessPointName)
		case s3arn.OutpostBucketARN:
			s3arn.SetARNField(in.Parameters, tv.BucketName)
		}
	}
	return next.HandleSerialize(ctx, in)
}

// AddUpdateOutpostARN is used by operation runtimes to add
// this middleware to their middleware stack.
func AddUpdateOutpostARN(stack *middleware.Stack) error {
	return stack.Serialize.Insert(
		&updateOutpostARN{},
		"OperationSerializer",
		middleware.Before,
	)
}
