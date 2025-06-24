package customizations

import (
	"context"
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/transport/http"
)

// AddAcceptHeader is the helper function used to add middleware to the stack
func AddAcceptHeader(stack *middleware.Stack) error {
	return stack.Build.Add(&acceptHeader{}, middleware.After)
}

type acceptHeader struct{}

// ID returns the middleware ID.
func (*acceptHeader) ID() string {
	return "APIGATEWAY:AcceptHeader"
}

// HandleBuild handles the associated build step of middleware stack
func (m *acceptHeader) HandleBuild(
	ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (
	out middleware.BuildOutput, metadata middleware.Metadata, err error,
) {
	req, ok := in.Request.(*http.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown transport type %T", in.Request)
	}

	// all APIGateway operations must have Accept header set to application/json
	const conHeader = "Accept"
	req.Header[conHeader] = append(req.Header[conHeader][:0], "application/json")

	return next.HandleBuild(ctx, in)
}
