package customizations

import (
	"context"
	"fmt"
	"net/url"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
	smithyhttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/transport/http"
)

// AddPredictEndpointMiddleware adds the middleware required to set the endpoint
// based on Predict's PredictEndpoint input member.
func AddPredictEndpointMiddleware(stack *middleware.Stack, endpoint func(interface{}) (*string, error)) error {
	return stack.Serialize.Insert(&predictEndpoint{}, "ResolveEndpoint", middleware.After)
}

// predictEndpoint rewrites the endpoint with whatever is specified in the
// operation input if it is non-nil and non-empty.
type predictEndpoint struct {
	fetchPredictEndpoint func(interface{}) (*string, error)
}

// ID returns the id for the middleware.
func (*predictEndpoint) ID() string { return "MachineLearning:PredictEndpoint" }

// HandleSerialize implements the SerializeMiddleware interface.
func (m *predictEndpoint) HandleSerialize(
	ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler,
) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, &smithy.SerializationError{
			Err: fmt.Errorf("unknown request type %T", in.Request),
		}
	}

	endpoint, err := m.fetchPredictEndpoint(in.Parameters)
	if err != nil {
		return out, metadata, &smithy.SerializationError{
			Err: fmt.Errorf("failed to fetch PredictEndpoint value, %v", err),
		}
	}

	if endpoint != nil && len(*endpoint) != 0 {
		uri, err := url.Parse(*endpoint)
		if err != nil {
			return out, metadata, &smithy.SerializationError{
				Err: fmt.Errorf("unable to parse predict endpoint, %v", err),
			}
		}
		req.URL = uri
	}

	return next.HandleSerialize(ctx, in)
}
