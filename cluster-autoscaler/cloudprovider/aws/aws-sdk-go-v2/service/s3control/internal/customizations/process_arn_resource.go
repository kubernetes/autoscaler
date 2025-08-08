package customizations

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
	smithyhttp "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/transport/http"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
	awsarn "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/arn"
	awsmiddleware "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/middleware"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/internal/s3shared"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/internal/s3shared/arn"
	s3endpoints "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/s3control/internal/endpoints/s3"
)

const (
	// outpost id header
	outpostIDHeader = "x-amz-outpost-id"

	// account id header
	accountIDHeader = "x-amz-account-id"
)

// processARNResource is used to process an ARN resource.
type processARNResource struct {

	// CopyInput creates a copy of input to be modified, this ensures the original input is not modified.
	CopyInput func(interface{}) (interface{}, error)

	// UpdateARNField points to a function that takes in a copy of input, updates the ARN field with
	// the provided value and returns the input
	UpdateARNField func(interface{}, string) error

	// UseARNRegion indicates if region parsed from an ARN should be used.
	UseARNRegion bool

	// EndpointResolver used to resolve endpoints. This may be a custom endpoint resolver
	EndpointResolver EndpointResolver

	// EndpointResolverOptions used by endpoint resolver
	EndpointResolverOptions EndpointResolverOptions
}

// ID returns the middleware ID.
func (*processARNResource) ID() string { return "S3Control:ProcessARNResourceMiddleware" }

func (m *processARNResource) HandleSerialize(
	ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler,
) (
	out middleware.SerializeOutput, metadata middleware.Metadata, err error,
) {
	if !awsmiddleware.GetRequiresLegacyEndpoints(ctx) {
		return next.HandleSerialize(ctx, in)
	}

	// if arn region resolves to custom endpoint that is mutable
	if smithyhttp.GetHostnameImmutable(ctx) {
		return next.HandleSerialize(ctx, in)
	}

	// check if arn was provided, if not skip this middleware
	arnValue, ok := s3shared.GetARNResourceFromContext(ctx)
	if !ok {
		return next.HandleSerialize(ctx, in)
	}

	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unknown request type %T", req)
	}

	// parse arn into an endpoint arn wrt to service
	resource, err := parseEndpointARN(arnValue)
	if err != nil {
		return out, metadata, err
	}

	resourceRequest := s3shared.ResourceRequest{
		Resource:      resource,
		RequestRegion: awsmiddleware.GetRegion(ctx),
		SigningRegion: awsmiddleware.GetSigningRegion(ctx),
		PartitionID:   awsmiddleware.GetPartitionID(ctx),
		UseARNRegion:  m.UseARNRegion,
	}

	// validate resource request
	if err := validateResourceRequest(resourceRequest); err != nil {
		return out, metadata, err
	}

	// if not done already, clone the input and reassign it to in.Parameters
	if !s3shared.IsClonedInput(ctx) {
		in.Parameters, err = m.CopyInput(in.Parameters)
		if err != nil {
			return out, metadata, fmt.Errorf("error creating a copy of input while processing arn")
		}
		// set copy input key on context
		ctx = s3shared.SetClonedInputKey(ctx, true)
	}

	// switch to correct endpoint updater
	switch tv := resource.(type) {
	case arn.OutpostAccessPointARN:
		// validations
		// check if dual stack
		if m.EndpointResolverOptions.UseDualStackEndpoint == aws.DualStackEndpointStateEnabled {
			return out, metadata, s3shared.NewClientConfiguredForDualStackError(tv,
				resourceRequest.PartitionID, resourceRequest.RequestRegion, nil)
		}

		// Disable endpoint host prefix for s3-control
		ctx = smithyhttp.DisableEndpointHostPrefix(ctx, true)

		if m.UpdateARNField == nil {
			return out, metadata, fmt.Errorf("error updating arnable field while serializing")
		}

		// update the arnable field with access point name
		err = m.UpdateARNField(in.Parameters, tv.AccessPointName)
		if err != nil {
			return out, metadata, fmt.Errorf("error updating arnable field while serializing")
		}

		// Add outpostID header
		req.Header.Add(outpostIDHeader, tv.OutpostID)

		// build outpost access point request
		ctx, err = buildOutpostAccessPointRequest(ctx, outpostAccessPointOptions{
			processARNResource: *m,
			request:            req,
			resource:           tv,
			partitionID:        resourceRequest.PartitionID,
			requestRegion:      resourceRequest.RequestRegion,
		})
		if err != nil {
			return out, metadata, err
		}

	// process outpost accesspoint ARN
	case arn.OutpostBucketARN:
		// check if dual stack
		if m.EndpointResolverOptions.UseDualStackEndpoint == aws.DualStackEndpointStateEnabled {
			return out, metadata, s3shared.NewClientConfiguredForDualStackError(tv,
				resourceRequest.PartitionID, resourceRequest.RequestRegion, nil)
		}

		// Disable endpoint host prefix for s3-control
		ctx = smithyhttp.DisableEndpointHostPrefix(ctx, true)

		if m.UpdateARNField == nil {
			return out, metadata, fmt.Errorf("error updating arnable field while serializing")
		}

		// update the arnable field with bucket name
		err = m.UpdateARNField(in.Parameters, tv.BucketName)
		if err != nil {
			return out, metadata, fmt.Errorf("error updating arnable field while serializing")
		}

		// Add outpostID header
		req.Header.Add(outpostIDHeader, tv.OutpostID)

		// build outpost bucket request
		ctx, err = buildOutpostBucketRequest(ctx, outpostBucketOptions{
			processARNResource: *m,
			request:            req,
			resource:           tv,
			partitionID:        resourceRequest.PartitionID,
			requestRegion:      resourceRequest.RequestRegion,
		})
		if err != nil {
			return out, metadata, err
		}

	default:
		return out, metadata, s3shared.NewInvalidARNError(resource, nil)
	}

	// Add account-id header for the request if not present.
	// SDK must always send the x-amz-account-id header for all requests
	// where an accountId has been extracted from an ARN or the accountId field modeled as a header.
	if h := req.Header.Get(accountIDHeader); len(h) == 0 {
		req.Header.Add(accountIDHeader, resource.GetARN().AccountID)
	}

	return next.HandleSerialize(ctx, in)
}

// validate if s3 resource and request config is compatible.
func validateResourceRequest(resourceRequest s3shared.ResourceRequest) error {
	// check if resourceRequest leads to a cross partition error
	v, err := resourceRequest.IsCrossPartition()
	if err != nil {
		return err
	}
	if v {
		// if cross partition
		return s3shared.NewClientPartitionMismatchError(resourceRequest.Resource,
			resourceRequest.PartitionID, resourceRequest.RequestRegion, nil)
	}

	// check if resourceRequest leads to a cross region error
	if !resourceRequest.AllowCrossRegion() && resourceRequest.IsCrossRegion() {
		// if cross region, but not use ARN region is not enabled
		return s3shared.NewClientRegionMismatchError(resourceRequest.Resource,
			resourceRequest.PartitionID, resourceRequest.RequestRegion, nil)
	}

	return nil
}

// Used by shapes with members decorated as endpoint ARN.
func parseEndpointARN(v awsarn.ARN) (arn.Resource, error) {
	return arn.ParseResource(v, resourceParser)
}

func resourceParser(a awsarn.ARN) (arn.Resource, error) {
	resParts := arn.SplitResource(a.Resource)
	switch resParts[0] {
	case "outpost":
		return arn.ParseOutpostARNResource(a, resParts[1:])
	default:
		return nil, arn.InvalidARNError{ARN: a, Reason: "unknown resource type"}
	}
}

// ====== Outpost Accesspoint ========

type outpostAccessPointOptions struct {
	processARNResource
	request       *smithyhttp.Request
	resource      arn.OutpostAccessPointARN
	partitionID   string
	requestRegion string
}

func buildOutpostAccessPointRequest(ctx context.Context, options outpostAccessPointOptions) (context.Context, error) {
	tv := options.resource
	req := options.request

	// Build outpost access point resource
	resolveRegion := tv.Region
	resolveService := tv.Service

	endpointsID := resolveService
	if resolveService == "s3-outposts" {
		endpointsID = "s3"
	}

	// resolve regional endpoint for resolved region.
	var endpoint aws.Endpoint
	var err error

	endpointSource := awsmiddleware.GetEndpointSource(ctx)

	eo := options.EndpointResolverOptions
	eo.Logger = middleware.GetLogger(ctx)
	eo.ResolvedRegion = ""

	if endpointsID == "s3" && endpointSource == aws.EndpointSourceServiceMetadata {
		// use s3 endpoint resolver
		endpoint, err = s3endpoints.New().ResolveEndpoint(resolveRegion, s3endpoints.Options{
			LogDeprecated:        eo.LogDeprecated,
			DisableHTTPS:         eo.DisableHTTPS,
			UseFIPSEndpoint:      eo.UseFIPSEndpoint,
			UseDualStackEndpoint: eo.UseDualStackEndpoint,
		})
	} else {
		endpoint, err = options.EndpointResolver.ResolveEndpoint(resolveRegion, eo)
	}

	if err != nil {
		return ctx, s3shared.NewFailedToResolveEndpointError(
			tv,
			options.partitionID,
			options.requestRegion,
			err,
		)
	}

	req.URL, err = url.Parse(endpoint.URL)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse endpoint URL: %w", err)
	}

	// redirect signer to use resolved endpoint signing name and region
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
		ctx = awsmiddleware.SetSigningRegion(ctx, resolveRegion)
	}

	// skip arn processing, if arn region resolves to a immutable endpoint
	if endpoint.HostnameImmutable {
		return ctx, nil
	}

	// add url host as s3-outposts
	cfgHost := req.URL.Host
	if strings.HasPrefix(cfgHost, endpointsID) {
		req.URL.Host = resolveService + cfgHost[len(endpointsID):]

		// update serviceID to resolved service
		ctx = awsmiddleware.SetServiceID(ctx, resolveService)
	}

	// validate the endpoint host
	if err := smithyhttp.ValidateEndpointHost(req.URL.Host); err != nil {
		return ctx, s3shared.NewInvalidARNError(tv, err)
	}

	// Disable endpoint host prefix for s3-control
	ctx = smithyhttp.DisableEndpointHostPrefix(ctx, true)

	return ctx, nil
}

// ======= Outpost Bucket =========
type outpostBucketOptions struct {
	processARNResource
	request       *smithyhttp.Request
	resource      arn.OutpostBucketARN
	partitionID   string
	requestRegion string
}

func buildOutpostBucketRequest(ctx context.Context, options outpostBucketOptions) (context.Context, error) {
	tv := options.resource
	req := options.request

	// Build endpoint from outpost bucket arn
	resolveRegion := tv.Region
	resolveService := tv.Service
	// Outpost bucket resource uses `s3-control` as serviceEndpointLabel
	endpointsID := "s3-control"

	// resolve regional endpoint for resolved region.
	eo := options.EndpointResolverOptions
	eo.Logger = middleware.GetLogger(ctx)
	eo.ResolvedRegion = ""

	endpoint, err := options.EndpointResolver.ResolveEndpoint(resolveRegion, eo)
	if err != nil {
		return ctx, s3shared.NewFailedToResolveEndpointError(
			tv,
			options.partitionID,
			options.requestRegion,
			err,
		)
	}

	// assign resolved endpoint url to request url
	req.URL, err = url.Parse(endpoint.URL)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse endpoint URL: %w", err)
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
		ctx = awsmiddleware.SetSigningRegion(ctx, resolveRegion)
	}

	// skip arn processing, if arn region resolves to a immutable endpoint
	if endpoint.HostnameImmutable {
		return ctx, nil
	}

	cfgHost := req.URL.Host
	if strings.HasPrefix(cfgHost, endpointsID) {
		// replace service endpointID label with resolved service
		req.URL.Host = resolveService + cfgHost[len(endpointsID):]

		// update serviceID to resolved service
		ctx = awsmiddleware.SetServiceID(ctx, resolveService)
	}

	// validate the endpoint host
	if err := smithyhttp.ValidateEndpointHost(req.URL.Host); err != nil {
		return ctx, s3shared.NewInvalidARNError(tv, err)
	}

	// Disable endpoint host prefix for s3-control
	ctx = smithyhttp.DisableEndpointHostPrefix(ctx, true)

	return ctx, nil
}
