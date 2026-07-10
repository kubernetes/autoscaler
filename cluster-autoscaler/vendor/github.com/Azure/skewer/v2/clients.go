package skewer

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/pkg/errors"
)

// wrappedResourceClient defines a wrapper for the typical Azure client
// signature to collect all resource skus from the iterator returned by NewListPager().
type wrappedResourceClient struct {
	client ResourceClient
}

func newWrappedResourceClient(client ResourceClient) *wrappedResourceClient {
	return &wrappedResourceClient{client}
}

func (w *wrappedResourceClient) List(ctx context.Context, filter, includeExtendedLocations string) ([]*armcompute.ResourceSKU, error) {
	options := &armcompute.ResourceSKUsClientListOptions{}
	if filter != "" {
		options.Filter = &filter
	}
	if includeExtendedLocations != "" {
		options.IncludeExtendedLocations = &includeExtendedLocations
	}
	pager := w.client.NewListPager(options)
	var skus []*armcompute.ResourceSKU
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "could not list resource skus")
		}
		skus = append(skus, page.Value...)
	}
	return skus, nil
}
