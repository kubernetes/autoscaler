package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// FirewallOperator provides an interface for operations on firewalls.
type FirewallOperator interface {
	GetFirewallList(ctx context.Context) ([]Firewall, error)
	GetFirewall(ctx context.Context, id string) (Firewall, error)
	CreateFirewall(ctx context.Context, body FirewallCreateRequest) (FirewallCreateResponse, error)
	UpdateFirewall(ctx context.Context, id string, body FirewallUpdateRequest) error
	DeleteFirewall(ctx context.Context, id string) error
	GetFirewallEventList(ctx context.Context, id string) ([]Event, error)
}

// FirewallList holds a list of firewalls.
type FirewallList struct {
	// Array of firewalls.
	List map[string]FirewallProperties `json:"firewalls"`
}

// Firewall represents a single firewall.
type Firewall struct {
	// Properties of a firewall.
	Properties FirewallProperties `json:"firewall"`
}

// FirewallProperties holds the properties of a firewall.
// Firewall's UUID can be used to attach a firewall to servers.
type FirewallProperties struct {
	// Status indicates the status of the object.
	Status string `json:"status"`

	// List of labels.
	Labels []string `json:"labels"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// Defines the date and time of the last object change.
	ChangeTime GSTime `json:"change_time"`

	// FirewallRules.
	Rules FirewallRules `json:"rules"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// If this is a private or public Firewall-Template.
	Private bool `json:"private"`

	// The information about other object which are related to this Firewall. The object could be Network.
	Relations FirewallRelation `json:"relations"`

	// Description of the ISO image release.
	Description string `json:"description"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationName string `json:"location_name"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`
}

// FirewallRules represents a list of firewall's rules.
type FirewallRules struct {
	// Firewall template rules for inbound traffic - covers IPv6 addresses.
	RulesV6In []FirewallRuleProperties `json:"rules-v6-in,omitempty"`

	// Firewall template rules for outbound traffic - covers IPv6 addresses.
	RulesV6Out []FirewallRuleProperties `json:"rules-v6-out,omitempty"`

	// Firewall template rules for inbound traffic - covers IPv4 addresses.
	RulesV4In []FirewallRuleProperties `json:"rules-v4-in,omitempty"`

	// Firewall template rules for outbound traffic - covers IPv4 addresses.
	RulesV4Out []FirewallRuleProperties `json:"rules-v4-out,omitempty"`
}

// FirewallRuleProperties represents properties of a firewall rule.
type FirewallRuleProperties struct {
	// Enum:"udp" "tcp". Allowed values: `TCPTransport`, `UDPTransport`.
	Protocol TransportLayerProtocol `json:"protocol"`

	// A Number between 1 and 65535, port ranges are separated by a colon for FTP.
	DstPort string `json:"dst_port,omitempty"`

	// A Number between 1 and 65535, port ranges are separated by a colon for FTP.
	SrcPort string `json:"src_port,omitempty"`

	// A Number between 1 and 65535, port ranges are separated by a colon for FTP.
	SrcCidr string `json:"src_cidr,omitempty"`

	// Enum:"accept" "drop". This defines what the firewall will do. Either accept or drop.
	Action string `json:"action"`

	// Description.
	Comment string `json:"comment,omitempty"`

	// Either an IPv4/6 address or and IP Network in CIDR format. If this field is empty then all IPs have access to this service.
	DstCidr string `json:"dst_cidr,omitempty"`

	// The order at which the firewall will compare packets against its rules,
	// a packet will be compared against the first rule, it will either allow it to pass
	// or block it and it won t be matched against any other rules.
	// However, if it does no match the rule, then it will proceed onto rule 2.
	// Packets that do not match any rules are blocked by default.
	Order int `json:"order"`
}

// FirewallRelation holds a list of firewall-network-server relations.
type FirewallRelation struct {
	// Array of object (NetworkinFirewall).
	Networks []NetworkInFirewall `json:"networks"`
}

// NetworkInFirewall represents properties of a firewall-network-server relation.
// A firewall-network-server relation tells which server uses the queried firewall
// and in which network that the firewall is enabled.
type NetworkInFirewall struct {
	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// The UUID of the network you're requesting.
	NetworkUUID string `json:"network_uuid"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	NetworkName string `json:"network_name"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	ObjectName string `json:"object_name"`
}

// FirewallCreateRequest represents a request for creating a new firewall.
type FirewallCreateRequest struct {
	// Name of firewall being created.
	Name string `json:"name"`

	// Labels. Can be nil.
	Labels []string `json:"labels,omitempty"`

	// FirewallRules.
	Rules FirewallRules `json:"rules"`
}

// FirewallCreateResponse represents a response for creating a firewall.
type FirewallCreateResponse struct {
	// Request UUID.
	RequestUUID string `json:"request_uuid"`

	// The UUID of the firewall being created.
	ObjectUUID string `json:"object_uuid"`
}

// FirewallUpdateRequest represent a request for updating a firewall.
type FirewallUpdateRequest struct {
	// New name. Leave it if you do not want to update the name.
	Name string `json:"name,omitempty"`

	// New list of labels. Leave it if you do not want to update the Labels.
	Labels *[]string `json:"labels,omitempty"`

	// FirewallRules. Leave it if you do not want to update the firewall rules.
	Rules *FirewallRules `json:"rules,omitempty"`
}

// TransportLayerProtocol represents a layer 4 protocol.
type TransportLayerProtocol string

// All available transport protocols.
var (
	TCPTransport TransportLayerProtocol = "tcp"
	UDPTransport TransportLayerProtocol = "udp"
)

// GetFirewallList gets a list of available firewalls.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getFirewalls
func (c *Client) GetFirewallList(ctx context.Context) ([]Firewall, error) {
	r := gsRequest{
		uri:                 path.Join(apiFirewallBase),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response FirewallList
	var firewalls []Firewall
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		firewalls = append(firewalls, Firewall{Properties: properties})
	}
	return firewalls, err
}

// GetFirewall gets a specific firewall based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getFirewall
func (c *Client) GetFirewall(ctx context.Context, id string) (Firewall, error) {
	if !isValidUUID(id) {
		return Firewall{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiFirewallBase, id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response Firewall
	err := r.execute(ctx, *c, &response)
	return response, err
}

// CreateFirewall creates a new firewall.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/createFirewall
func (c *Client) CreateFirewall(ctx context.Context, body FirewallCreateRequest) (FirewallCreateResponse, error) {
	r := gsRequest{
		uri:    path.Join(apiFirewallBase),
		method: http.MethodPost,
		body:   body,
	}
	var response FirewallCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdateFirewall update a specific firewall.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateFirewall
func (c *Client) UpdateFirewall(ctx context.Context, id string, body FirewallUpdateRequest) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiFirewallBase, id),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeleteFirewall removes a specific firewall.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deleteFirewall
func (c *Client) DeleteFirewall(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiFirewallBase, id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// GetFirewallEventList get list of a firewall's events.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getFirewallEvents
func (c *Client) GetFirewallEventList(ctx context.Context, id string) ([]Event, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiFirewallBase, id, "events"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response EventList
	var firewallEvents []Event
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		firewallEvents = append(firewallEvents, Event{Properties: properties})
	}
	return firewallEvents, err
}
