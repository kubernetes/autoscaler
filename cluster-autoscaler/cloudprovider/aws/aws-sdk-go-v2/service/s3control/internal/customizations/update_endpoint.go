package customizations

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/internal/s3shared"

	internalendpoints "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/s3control/internal/endpoints"
)

// EndpointResolver interface for resolving service endpoints.
type EndpointResolver interface {
	ResolveEndpoint(region string, options EndpointResolverOptions) (aws.Endpoint, error)
}

// EndpointResolverOptions is the service endpoint resolver options
type EndpointResolverOptions = internalendpoints.Options

// UpdateEndpointParameterAccessor represents accessor functions used by the middleware
type UpdateEndpointParameterAccessor struct {
	// GetARNInput points to a function that processes an input and returns ARN as string ptr,
	// and bool indicating if ARN is supported or set.
	GetARNInput func(interface{}) (*string, bool)

	// GetOutpostIDInput points to a function that processes an input and returns a outpostID as string ptr,
	// and bool indicating if outpostID is supported or set.
	GetOutpostIDInput func(interface{}) (*string, bool)

	// CopyInput creates a copy of input to be modified, this ensures the original input is not modified.
	CopyInput func(interface{}) (interface{}, error)

	// BackfillAccountID points to a function that validates the input for accountID. If absent, it populates the
	// accountID. If present, but different than passed in accountID value throws an error
	BackfillAccountID func(interface{}, string) error

	// UpdateARNField points to a function that takes in a copy of input, updates the ARN field with
	// the provided value and returns any error
	UpdateARNField func(interface{}, string) error
}

// UpdateEndpointOptions provides the options for the UpdateEndpoint middleware setup.
type UpdateEndpointOptions struct {

	// Accessor are parameter accessors used by the middleware
	Accessor UpdateEndpointParameterAccessor

	// UseARNRegion indicates if region parsed from an ARN should be used.
	UseARNRegion bool

	// EndpointResolver used to resolve endpoints. This may be a custom endpoint resolver
	EndpointResolver EndpointResolver

	// EndpointResolverOptions used by endpoint resolver
	EndpointResolverOptions EndpointResolverOptions
}

// UpdateEndpoint adds the middleware to the middleware stack based on the UpdateEndpointOptions.
func UpdateEndpoint(stack *middleware.Stack, options UpdateEndpointOptions) (err error) {
	// validate and backfill account id from ARN
	err = stack.Initialize.Add(&BackfillInput{
		CopyInput:         options.Accessor.CopyInput,
		BackfillAccountID: options.Accessor.BackfillAccountID,
	}, middleware.Before)
	if err != nil {
		return err
	}

	// initial arn look up middleware should be before BackfillInput
	err = stack.Initialize.Insert(&s3shared.ARNLookup{
		GetARNValue: options.Accessor.GetARNInput,
	}, (*BackfillInput)(nil).ID(), middleware.Before)
	if err != nil {
		return err
	}

	// process arn
	err = stack.Serialize.Insert(&processARNResource{
		CopyInput:               options.Accessor.CopyInput,
		UpdateARNField:          options.Accessor.UpdateARNField,
		UseARNRegion:            options.UseARNRegion,
		EndpointResolver:        options.EndpointResolver,
		EndpointResolverOptions: options.EndpointResolverOptions,
	}, "OperationSerializer", middleware.Before)
	if err != nil {
		return err
	}

	// outpostID middleware
	err = stack.Serialize.Insert(&processOutpostIDMiddleware{
		GetOutpostID:            options.Accessor.GetOutpostIDInput,
		EndpointResolver:        options.EndpointResolver,
		EndpointResolverOptions: options.EndpointResolverOptions,
	}, (*processARNResource)(nil).ID(), middleware.Before)
	if err != nil {
		return err
	}
	return err
}
