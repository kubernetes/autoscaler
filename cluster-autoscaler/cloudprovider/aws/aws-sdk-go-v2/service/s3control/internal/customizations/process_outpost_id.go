package customizations

import (
	"context"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
	"net/url"
	"strings"

	awsmiddleware "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/middleware"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
	smithyhttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/transport/http"
)

// processOutpostIDMiddleware is special customization middleware to be applied for operations
// CreateBucket, ListRegionalBuckets which must resolve endpoint to s3-outposts.{region}.amazonaws.com
// with region as client region and signed by s3-control if an outpost id is provided.
type processOutpostIDMiddleware struct {
	// GetOutpostID points to a function that processes an input and returns an outpostID as string ptr,
	// and bool indicating if outpostID is supported or set.
	GetOutpostID func(interface{}) (*string, bool)

	// EndpointResolver used to resolve endpoints. This may be a custom endpoint resolver
	EndpointResolver EndpointResolver

	// EndpointResolverOptions used by endpoint resolver
	EndpointResolverOptions EndpointResolverOptions
}

// ID returns the middleware ID.
func (*processOutpostIDMiddleware) ID() string { return "S3Control:ProcessOutpostIDMiddleware" }

// HandleSerialize adds a serialize step, this has to be before operation serializer and arn endpoint logic.
// Ideally this step will be ahead of ARN customization for CreateBucket, ListRegionalBucket operation.
func (m *processOutpostIDMiddleware) HandleSerialize(
	ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler,
) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	if !awsmiddleware.GetRequiresLegacyEndpoints(ctx) {
		return next.HandleSerialize(ctx, in)
	}

	// if host name is immutable, skip this customization
	if smithyhttp.GetHostnameImmutable(ctx) {
		return next.HandleSerialize(ctx, in)
	}

	// attempt to fetch an outpost id
	outpostID, ok := m.GetOutpostID(in.Parameters)
	if !ok {
		return next.HandleSerialize(ctx, in)
	}

	// check if outpostID was not set or is empty
	if outpostID == nil || len(strings.TrimSpace(*outpostID)) == 0 {
		return next.HandleSerialize(ctx, in)
	}

	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown request type %T", req)
	}

	requestRegion := awsmiddleware.GetRegion(ctx)

	ero := m.EndpointResolverOptions

	// validate if dualstack
	if ero.UseDualStackEndpoint == aws.DualStackEndpointStateEnabled {
		return out, metadata, fmt.Errorf("dualstack is not supported for outposts request")
	}

	endpoint, err := m.EndpointResolver.ResolveEndpoint(requestRegion, ero)
	if err != nil {
		return out, metadata, err
	}

	// resolved endpoint with endpoint-id s3-control
	endpointsID := "s3-control"

	// resolved service label that must be used in case endpointsID is s3-control
	resolveService := "s3-outposts"

	req.URL, err = url.Parse(endpoint.URL)
	if err != nil {
		return out, metadata, fmt.Errorf("failed to parse endpoint URL: %w", err)
	}

	if len(endpoint.SigningName) != 0 {
		ctx = awsmiddleware.SetSigningName(ctx, endpoint.SigningName)
	} else {
		// assign resolved service from arn as signing name
		ctx = awsmiddleware.SetSigningName(ctx, resolveService)
	}

	if len(endpoint.SigningRegion) != 0 {
		// redirect signer to use resolved endpoint signing name and region
		ctx = awsmiddleware.SetSigningRegion(ctx, endpoint.SigningRegion)
	} else {
		ctx = awsmiddleware.SetSigningRegion(ctx, requestRegion)
	}

	// add url host as s3-outposts
	cfgHost := req.URL.Host
	if strings.HasPrefix(cfgHost, endpointsID) {
		req.URL.Host = resolveService + cfgHost[len(endpointsID):]
		ctx = awsmiddleware.SetServiceID(ctx, resolveService)
	}

	// Disable endpoint host prefix for s3-control
	ctx = smithyhttp.DisableEndpointHostPrefix(ctx, true)

	return next.HandleSerialize(ctx, in)
}
