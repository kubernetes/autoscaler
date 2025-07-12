package testing

import (
	"context"
	"net/url"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/s3"
	smithyendpoints "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/endpoints"
)

// EndpointResolverV2 is a mock s3 endpoint resolver v2 for testing
type EndpointResolverV2 struct {
	URL string
}

// ResolveEndpoint returns the given endpoint url
func (r EndpointResolverV2) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (smithyendpoints.Endpoint, error) {
	u, err := url.Parse(r.URL)
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}
	return smithyendpoints.Endpoint{
		URI: *u,
	}, nil
}
