package skewer

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
)

// ResourceClient is the required Azure client interface used to populate skewer's data.
type ResourceClient interface {
	NewListPager(options *armcompute.ResourceSKUsClientListOptions) *runtime.Pager[armcompute.ResourceSKUsClientListResponse]
}

var _ ResourceClient = &armcompute.ResourceSKUsClient{}

// client defines the internal interface required by the skewer Cache.
type client interface {
	List(ctx context.Context, filter, includeExtendedLocations string) ([]*armcompute.ResourceSKU, error)
}
