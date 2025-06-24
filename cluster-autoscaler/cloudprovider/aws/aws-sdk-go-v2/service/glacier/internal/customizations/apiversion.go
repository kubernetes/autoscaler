package customizations

import (
	"context"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
	smithyhttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/transport/http"
)

const glacierAPIVersionHeaderKey = "X-Amz-Glacier-Version"

// AddGlacierAPIVersionMiddleware explicitly add handling for the Glacier api version
// middleware to the operation stack.
func AddGlacierAPIVersionMiddleware(stack *middleware.Stack, apiVersion string) error {
	return stack.Serialize.Add(&GlacierAPIVersion{apiVersion: apiVersion}, middleware.Before)
}

// GlacierAPIVersion handles automatically setting Glacier's API version header.
type GlacierAPIVersion struct {
	apiVersion string
}

// ID returns the id for the middleware.
func (*GlacierAPIVersion) ID() string {
	return "Glacier:APIVersion"
}

// HandleSerialize implements the SerializeMiddleware interface
func (m *GlacierAPIVersion) HandleSerialize(
	ctx context.Context, input middleware.SerializeInput, next middleware.SerializeHandler,
) (
	output middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	req, ok := input.Request.(*smithyhttp.Request)
	if !ok {
		return output, metadata, &smithy.SerializationError{
			Err: fmt.Errorf("unknown request type %T", input.Request),
		}
	}

	req.Header.Set(glacierAPIVersionHeaderKey, m.apiVersion)

	return next.HandleSerialize(ctx, input)
}
