package hcloud

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud/exp/ctxutil"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud/schema"
)

// Pricing specifies pricing information for various resources.
type Pricing struct {
	Image ImagePricing
	// Deprecated: [Pricing.FloatingIP] is deprecated, use [Pricing.FloatingIPs] instead.
	FloatingIP  FloatingIPPricing
	FloatingIPs []FloatingIPTypePricing
	PrimaryIPs  []PrimaryIPPricing
	// Deprecated: [Pricing.Traffic] is deprecated and will report 0 after 2024-08-05.
	// Use traffic pricing from [Pricing.ServerTypes] or [Pricing.LoadBalancerTypes] instead.
	Traffic           TrafficPricing
	ServerBackup      ServerBackupPricing
	ServerTypes       []ServerTypePricing
	LoadBalancerTypes []LoadBalancerTypePricing
	Volume            VolumePricing
}

// Price represents a price. Net amount, gross amount, as well as VAT rate are
// specified as strings and it is the user's responsibility to convert them to
// appropriate types for calculations.
type Price struct {
	Currency string
	VATRate  string
	Net      string
	Gross    string
}

// PrimaryIPPrice represents a price. Net amount and gross amount are
// specified as strings and it is the user's responsibility to convert them to
// appropriate types for calculations.
type PrimaryIPPrice struct {
	Net   string
	Gross string
}

// ImagePricing provides pricing information for imaegs.
type ImagePricing struct {
	PerGBMonth Price
}

// FloatingIPPricing provides pricing information for Floating IPs.
type FloatingIPPricing struct {
	Monthly Price
}

// FloatingIPTypePricing provides pricing information for Floating IPs per Type.
type FloatingIPTypePricing struct {
	Type     FloatingIPType
	Pricings []FloatingIPTypeLocationPricing
}

// PrimaryIPTypePricing defines the schema of pricing information for a primary IP
// type at a datacenter.
type PrimaryIPTypePricing struct {
	Datacenter string // Deprecated: the API does not return pricing for the individual DCs anymore
	Location   string
	Hourly     PrimaryIPPrice
	Monthly    PrimaryIPPrice
}

// PrimaryIPTypePricing provides pricing information for PrimaryIPs.
type PrimaryIPPricing struct {
	Type     string
	Pricings []PrimaryIPTypePricing
}

// FloatingIPTypeLocationPricing provides pricing information for a Floating IP type
// at a location.
type FloatingIPTypeLocationPricing struct {
	Location *Location
	Monthly  Price
}

// TrafficPricing provides pricing information for traffic.
type TrafficPricing struct {
	PerTB Price
}

// VolumePricing provides pricing information for a Volume.
type VolumePricing struct {
	PerGBMonthly Price
}

// ServerBackupPricing provides pricing information for server backups.
type ServerBackupPricing struct {
	Percentage string
}

// ServerTypePricing provides pricing information for a server type.
type ServerTypePricing struct {
	ServerType *ServerType
	Pricings   []ServerTypeLocationPricing
}

// ServerTypeLocationPricing provides pricing information for a server type
// at a location.
type ServerTypeLocationPricing struct {
	Location *Location
	Hourly   Price
	Monthly  Price

	// IncludedTraffic is the free traffic per month in bytes
	IncludedTraffic uint64
	PerTBTraffic    Price
}

// LoadBalancerTypePricing provides pricing information for a Load Balancer type.
type LoadBalancerTypePricing struct {
	LoadBalancerType *LoadBalancerType
	Pricings         []LoadBalancerTypeLocationPricing
}

// LoadBalancerTypeLocationPricing provides pricing information for a Load Balancer type
// at a location.
type LoadBalancerTypeLocationPricing struct {
	Location *Location
	Hourly   Price
	Monthly  Price

	// IncludedTraffic is the free traffic per month in bytes
	IncludedTraffic uint64
	PerTBTraffic    Price
}

// PricingClient is a client for the pricing API.
type PricingClient struct {
	client *Client
}

// Get retrieves pricing information.
func (c *PricingClient) Get(ctx context.Context) (Pricing, *Response, error) {
	const opPath = "/pricing"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := opPath

	respBody, resp, err := getRequest[schema.PricingGetResponse](ctx, c.client, reqPath)
	if err != nil {
		return Pricing{}, resp, err
	}

	return PricingFromSchema(respBody.Pricing), resp, nil
}
