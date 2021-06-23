package gobrightbox

import (
	"time"
)

// LoadBalancer represents a Load Balancer
// https://api.gb1.brightbox.com/1.0/#load_balancer
type LoadBalancer struct {
	Id                string
	Name              string
	Status            string
	CreatedAt         *time.Time `json:"created_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
	Locked            bool
	HttpsRedirect     bool   `json:"https_redirect"`
	SslMinimumVersion string `json:"ssl_minimum_version"`
	Account           Account
	Nodes             []Server
	CloudIPs          []CloudIP `json:"cloud_ips"`
	Policy            string
	BufferSize        int `json:"buffer_size"`
	Listeners         []LoadBalancerListener
	Healthcheck       LoadBalancerHealthcheck
	Certificate       *LoadBalancerCertificate
	Acme              *LoadBalancerAcme
}

// LoadBalancerCertificate represents a certificate on a LoadBalancer
type LoadBalancerCertificate struct {
	ExpiresAt time.Time `json:"expires_at"`
	ValidFrom time.Time `json:"valid_from"`
	SslV3     bool      `json:"sslv3"`
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
}

// LoadBalancerAcme represents an ACME object on a LoadBalancer
type LoadBalancerAcme struct {
	Certificate *LoadBalancerAcmeCertificate `json:"certificate"`
	Domains     []LoadBalancerAcmeDomain     `json:"domains"`
}

// LoadBalancerAcmeCertificate represents an ACME issued certificate on
// a LoadBalancer
type LoadBalancerAcmeCertificate struct {
	Fingerprint string    `json:"fingerprint"`
	ExpiresAt   time.Time `json:"expires_at"`
	IssuedAt    time.Time `json:"issued_at"`
}

// LoadBalancerAcmeDomain represents a domain for which ACME support
// has been requested
type LoadBalancerAcmeDomain struct {
	Identifier  string `json:"identifier"`
	Status      string `json:"status"`
	LastMessage string `json:"last_message"`
}

// LoadBalancerHealthcheck represents a health check on a LoadBalancer
type LoadBalancerHealthcheck struct {
	Type          string `json:"type"`
	Port          int    `json:"port"`
	Request       string `json:"request,omitempty"`
	Interval      int    `json:"interval,omitempty"`
	Timeout       int    `json:"timeout,omitempty"`
	ThresholdUp   int    `json:"threshold_up,omitempty"`
	ThresholdDown int    `json:"threshold_down,omitempty"`
}

// LoadBalancerListener represents a listener on a LoadBalancer
type LoadBalancerListener struct {
	Protocol      string `json:"protocol,omitempty"`
	In            int    `json:"in,omitempty"`
	Out           int    `json:"out,omitempty"`
	Timeout       int    `json:"timeout,omitempty"`
	ProxyProtocol string `json:"proxy_protocol,omitempty"`
}

// LoadBalancerOptions is used in conjunction with CreateLoadBalancer and
// UpdateLoadBalancer to create and update load balancers
type LoadBalancerOptions struct {
	Id                    string                   `json:"-"`
	Name                  *string                  `json:"name,omitempty"`
	Nodes                 []LoadBalancerNode       `json:"nodes,omitempty"`
	Policy                *string                  `json:"policy,omitempty"`
	BufferSize            *int                     `json:"buffer_size,omitempty"`
	Listeners             []LoadBalancerListener   `json:"listeners,omitempty"`
	Healthcheck           *LoadBalancerHealthcheck `json:"healthcheck,omitempty"`
	Domains               *[]string                `json:"domains,omitempty"`
	CertificatePem        *string                  `json:"certificate_pem,omitempty"`
	CertificatePrivateKey *string                  `json:"certificate_private_key,omitempty"`
	SslMinimumVersion     *string                  `json:"ssl_minimum_version,omitempty"`
	SslV3                 *bool                    `json:"sslv3,omitempty"`
	HttpsRedirect         *bool                    `json:"https_redirect,omitempty"`
}

// LoadBalancerNode is used in conjunction with LoadBalancerOptions,
// AddNodesToLoadBalancer, RemoveNodesFromLoadBalancer to specify a list of
// servers to use as load balancer nodes. The Node parameter should be a server
// identifier.
type LoadBalancerNode struct {
	Node string `json:"node"`
}

// LoadBalancers retrieves a list of all load balancers
func (c *Client) LoadBalancers() ([]LoadBalancer, error) {
	var lbs []LoadBalancer
	_, err := c.MakeApiRequest("GET", "/1.0/load_balancers", nil, &lbs)
	if err != nil {
		return nil, err
	}
	return lbs, err
}

// LoadBalancer retrieves a detailed view of one load balancer
func (c *Client) LoadBalancer(identifier string) (*LoadBalancer, error) {
	lb := new(LoadBalancer)
	_, err := c.MakeApiRequest("GET", "/1.0/load_balancers/"+identifier, nil, lb)
	if err != nil {
		return nil, err
	}
	return lb, err
}

// CreateLoadBalancer creates a new load balancer.
//
// It takes a LoadBalancerOptions struct for specifying name and other
// attributes.  Not all attributes can be specified at create time (such as Id,
// which is allocated for you)
func (c *Client) CreateLoadBalancer(newLB *LoadBalancerOptions) (*LoadBalancer, error) {
	lb := new(LoadBalancer)
	_, err := c.MakeApiRequest("POST", "/1.0/load_balancers", newLB, &lb)
	if err != nil {
		return nil, err
	}
	return lb, nil
}

// UpdateLoadBalancer updates an existing load balancer.
//
// It takes a LoadBalancerOptions struct for specifying name and other
// attributes. Provide the identifier using the Id attribute.
func (c *Client) UpdateLoadBalancer(newLB *LoadBalancerOptions) (*LoadBalancer, error) {
	lb := new(LoadBalancer)
	_, err := c.MakeApiRequest("PUT", "/1.0/load_balancers/"+newLB.Id, newLB, &lb)
	if err != nil {
		return nil, err
	}
	return lb, nil
}

// DestroyLoadBalancer issues a request to destroy the load balancer
func (c *Client) DestroyLoadBalancer(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/load_balancers/"+identifier, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// AddNodesToLoadBalancer adds nodes to an existing load balancer.
func (c *Client) AddNodesToLoadBalancer(loadBalancerID string, nodes []LoadBalancerNode) (*LoadBalancer, error) {
	lb := new(LoadBalancer)
	_, err := c.MakeApiRequest("POST", "/1.0/load_balancers/"+loadBalancerID+"/add_nodes", nodes, &lb)
	if err != nil {
		return nil, err
	}
	return lb, nil
}

// RemoveNodesFromLoadBalancer removes nodes from an existing load balancer.
func (c *Client) RemoveNodesFromLoadBalancer(loadBalancerID string, nodes []LoadBalancerNode) (*LoadBalancer, error) {
	lb := new(LoadBalancer)
	_, err := c.MakeApiRequest("POST", "/1.0/load_balancers/"+loadBalancerID+"/remove_nodes", nodes, &lb)
	if err != nil {
		return nil, err
	}
	return lb, nil
}

// AddListenersToLoadBalancer adds listeners to an existing load balancer.
func (c *Client) AddListenersToLoadBalancer(loadBalancerID string, listeners []LoadBalancerListener) (*LoadBalancer, error) {
	lb := new(LoadBalancer)
	_, err := c.MakeApiRequest("POST", "/1.0/load_balancers/"+loadBalancerID+"/add_listeners", listeners, &lb)
	if err != nil {
		return nil, err
	}
	return lb, nil
}

// RemoveListenersFromLoadBalancer removes listeners to an existing load balancer.
func (c *Client) RemoveListenersFromLoadBalancer(loadBalancerID string, listeners []LoadBalancerListener) (*LoadBalancer, error) {
	lb := new(LoadBalancer)
	_, err := c.MakeApiRequest("POST", "/1.0/load_balancers/"+loadBalancerID+"/remove_listeners", listeners, &lb)
	if err != nil {
		return nil, err
	}
	return lb, nil
}
