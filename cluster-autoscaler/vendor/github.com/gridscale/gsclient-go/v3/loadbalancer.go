package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// LoadBalancerOperator provides an interface for operations on load balancers.
type LoadBalancerOperator interface {
	GetLoadBalancerList(ctx context.Context) ([]LoadBalancer, error)
	GetLoadBalancer(ctx context.Context, id string) (LoadBalancer, error)
	CreateLoadBalancer(ctx context.Context, body LoadBalancerCreateRequest) (LoadBalancerCreateResponse, error)
	UpdateLoadBalancer(ctx context.Context, id string, body LoadBalancerUpdateRequest) error
	DeleteLoadBalancer(ctx context.Context, id string) error
	GetLoadBalancerEventList(ctx context.Context, id string) ([]Event, error)
}

// LoadBalancers holds a list of load balancers.
type LoadBalancers struct {
	// Array of load balancers.
	List map[string]LoadBalancerProperties `json:"loadbalancers"`
}

// LoadBalancer represent a single load balancer.
type LoadBalancer struct {
	// Properties of a load balancer.
	Properties LoadBalancerProperties `json:"loadbalancer"`
}

// LoadBalancerProperties holds properties of a load balancer.
type LoadBalancerProperties struct {
	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// Defines the numbering of the Data Centers on a given IATA location (e.g. where fra is the location_iata, the site is then 1, 2, 3, ...).
	LocationSite string `json:"location_site"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// Forwarding rules of a load balancer.
	ForwardingRules []ForwardingRule `json:"forwarding_rules"`

	// Uses IATA airport code, which works as a location identifier.
	LocationIata string `json:"location_iata"`

	// Helps to identify which data center an object belongs to.
	LocationUUID string `json:"location_uuid"`

	// The servers that this Load balancer can communicate with.
	BackendServers []BackendServer `json:"backend_servers"`

	// Defines the date and time of the last object change.
	ChangeTime GSTime `json:"change_time"`

	// Status indicates the status of the object.
	Status string `json:"status"`

	// **DEPRECATED** The price for the current period since the last bill.
	CurrentPrice float64 `json:"current_price"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationCountry string `json:"location_country"`

	// Whether the Load balancer is forced to redirect requests from HTTP to HTTPS.
	RedirectHTTPToHTTPS bool `json:"redirect_http_to_https"`

	// List of labels.
	Labels []string `json:"labels"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationName string `json:"location_name"`

	// Total minutes of cores used.
	UsageInMinutes int `json:"usage_in_minutes"`

	// The algorithm used to process requests. Accepted values: roundrobin / leastconn.
	Algorithm string `json:"algorithm"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// The UUID of the IPv6 address the Load balancer will listen to for incoming requests.
	ListenIPv6UUID string `json:"listen_ipv6_uuid"`

	// The UUID of the IPv4 address the Load balancer will listen to for incoming requests.
	ListenIPv4UUID string `json:"listen_ipv4_uuid"`
}

// BackendServer holds properties telling how a load balancer deals with a backend server.
type BackendServer struct {
	// Weight of the server.
	Weight int `json:"weight"`

	// Host of the server. Can be URL or IP address.
	Host string `json:"host"`
}

// ForwardingRule represents a forwarding rule.
// It tells which port are forwarded to which port.
type ForwardingRule struct {
	// A valid domain name that points to the loadbalancer's IP address.
	LetsencryptSSL *string `json:"letsencrypt_ssl"`

	// The UUID of a custom certificate.
	CertificateUUID string `json:"certificate_uuid,omitempty"`

	// Listen port.
	ListenPort int `json:"listen_port"`

	// Mode of forwarding.
	Mode string `json:"mode"`

	// Target port.
	TargetPort int `json:"target_port"`
}

// LoadBalancerCreateRequest represents a request for creating a load balancer.
type LoadBalancerCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	ListenIPv6UUID string `json:"listen_ipv6_uuid"`

	// The UUID of the IPv4 address the load balancer will listen to for incoming requests.
	ListenIPv4UUID string `json:"listen_ipv4_uuid"`

	// The algorithm used to process requests. Allowed values: `LoadbalancerRoundrobinAlg`, `LoadbalancerLeastConnAlg`.
	Algorithm LoadbalancerAlgorithm `json:"algorithm"`

	// An array of ForwardingRule objects containing the forwarding rules for the load balancer
	ForwardingRules []ForwardingRule `json:"forwarding_rules"`

	// The servers that this load balancer can communicate with.
	BackendServers []BackendServer `json:"backend_servers"`

	// List of labels.
	Labels []string `json:"labels"`

	// Whether the Load balancer is forced to redirect requests from HTTP to HTTPS.
	RedirectHTTPToHTTPS bool `json:"redirect_http_to_https"`

	// Status indicates the status of the object.
	Status string `json:"status,omitempty"`
}

// LoadBalancerUpdateRequest represents a request for updating a load balancer.
type LoadBalancerUpdateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	ListenIPv6UUID string `json:"listen_ipv6_uuid"`

	// The UUID of the IPv4 address the loadbalancer will listen to for incoming requests.
	ListenIPv4UUID string `json:"listen_ipv4_uuid"`

	// The algorithm used to process requests. Allowed values: `LoadbalancerRoundrobinAlg`, `LoadbalancerLeastConnAlg`
	Algorithm LoadbalancerAlgorithm `json:"algorithm"`

	// An array of ForwardingRule objects containing the forwarding rules for the load balancer.
	ForwardingRules []ForwardingRule `json:"forwarding_rules"`

	// The servers that this load balancer can communicate with.
	BackendServers []BackendServer `json:"backend_servers"`

	// List of labels.
	Labels []string `json:"labels"`

	// Whether the Load balancer is forced to redirect requests from HTTP to HTTPS.
	RedirectHTTPToHTTPS bool `json:"redirect_http_to_https"`

	// Status indicates the status of the object.
	Status string `json:"status,omitempty"`
}

// LoadBalancerCreateResponse represents a response for creating a load balancer.
type LoadBalancerCreateResponse struct {
	// Request's UUID.
	RequestUUID string `json:"request_uuid"`

	// UUID of the load balancer being created.
	ObjectUUID string `json:"object_uuid"`
}

// LoadbalancerAlgorithm represents the algorithm that a load balancer uses to balance
// the incoming requests.
type LoadbalancerAlgorithm string

// All available load balancer algorithms.
var (
	LoadbalancerRoundrobinAlg LoadbalancerAlgorithm = "roundrobin"
	LoadbalancerLeastConnAlg  LoadbalancerAlgorithm = "leastconn"
)

// GetLoadBalancerList returns a list of load balancers.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getLoadbalancers
func (c *Client) GetLoadBalancerList(ctx context.Context) ([]LoadBalancer, error) {
	r := gsRequest{
		uri:                 apiLoadBalancerBase,
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response LoadBalancers
	var loadBalancers []LoadBalancer
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		loadBalancers = append(loadBalancers, LoadBalancer{Properties: properties})
	}
	return loadBalancers, err
}

// GetLoadBalancer returns a load balancer of a given UUID.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getLoadbalancer
func (c *Client) GetLoadBalancer(ctx context.Context, id string) (LoadBalancer, error) {
	if !isValidUUID(id) {
		return LoadBalancer{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiLoadBalancerBase, id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response LoadBalancer
	err := r.execute(ctx, *c, &response)
	return response, err
}

// CreateLoadBalancer creates a new load balancer.
//
// Note: A load balancer's algorithm can only be either `LoadbalancerRoundrobinAlg` or `LoadbalancerLeastConnAlg`.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/createLoadbalancer
func (c *Client) CreateLoadBalancer(ctx context.Context, body LoadBalancerCreateRequest) (LoadBalancerCreateResponse, error) {
	if body.Labels == nil {
		body.Labels = make([]string, 0)
	}
	r := gsRequest{
		uri:    apiLoadBalancerBase,
		method: http.MethodPost,
		body:   body,
	}
	var response LoadBalancerCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdateLoadBalancer update configuration of a load balancer.
//
// Note: A load balancer's algorithm can only be either `LoadbalancerRoundrobinAlg` or `LoadbalancerLeastConnAlg`.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateLoadbalancer
func (c *Client) UpdateLoadBalancer(ctx context.Context, id string, body LoadBalancerUpdateRequest) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	if body.Labels == nil {
		body.Labels = make([]string, 0)
	}
	r := gsRequest{
		uri:    path.Join(apiLoadBalancerBase, id),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// GetLoadBalancerEventList retrieves a load balancer's events based on a given load balancer UUID.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getLoadbalancerEvents
func (c *Client) GetLoadBalancerEventList(ctx context.Context, id string) ([]Event, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiLoadBalancerBase, id, "events"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response EventList
	var loadBalancerEvents []Event
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		loadBalancerEvents = append(loadBalancerEvents, Event{Properties: properties})
	}
	return loadBalancerEvents, err
}

// DeleteLoadBalancer removes a load balancer.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deleteLoadbalancer
func (c *Client) DeleteLoadBalancer(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiLoadBalancerBase, id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}
