package imds

import (
	"context"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
)

type resolveEndpointV2Middleware struct {
	options Options
}

func (*resolveEndpointV2Middleware) ID() string {
	return "ResolveEndpointV2"
}

func (m *resolveEndpointV2Middleware) HandleFinalize(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (
	out middleware.FinalizeOutput, metadata middleware.Metadata, err error,
) {
	return next.HandleFinalize(ctx, in)
}
