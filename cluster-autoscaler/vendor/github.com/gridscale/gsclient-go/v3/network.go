package gsclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
)

// NetworkOperator provides an interface for operations on networks.
type NetworkOperator interface {
	GetNetwork(ctx context.Context, id string) (Network, error)
	GetNetworkList(ctx context.Context) ([]Network, error)
	CreateNetwork(ctx context.Context, body NetworkCreateRequest) (NetworkCreateResponse, error)
	DeleteNetwork(ctx context.Context, id string) error
	UpdateNetwork(ctx context.Context, id string, body NetworkUpdateRequest) error
	GetNetworkEventList(ctx context.Context, id string) ([]Event, error)
	GetNetworkPublic(ctx context.Context) (Network, error)
	GetNetworksByLocation(ctx context.Context, id string) ([]Network, error)
	GetDeletedNetworks(ctx context.Context) ([]Network, error)
	GetPinnedServerList(ctx context.Context, networkUUID string) (PinnedServerList, error)
	UpdateNetworkPinnedServer(ctx context.Context, networkUUID, serverUUID string, body PinServerRequest) error
	DeleteNetworkPinnedServer(ctx context.Context, networkUUID, serverUUID string) error
}

// NetworkList holds a list of available networks.
type NetworkList struct {
	// Array of networks.
	List map[string]NetworkProperties `json:"networks"`
}

// DeletedNetworkList holds a list of deleted networks.
type DeletedNetworkList struct {
	// Array of deleted networks.
	List map[string]NetworkProperties `json:"deleted_networks"`
}

// Network represents a single network.
type Network struct {
	// Properties of a network.
	Properties NetworkProperties `json:"network"`
}

// NetworkProperties holds properties of a network.
// A network can be retrieved and attached to servers via the network UUID.
type NetworkProperties struct {
	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationCountry string `json:"location_country"`

	// Helps to identify which data center an object belongs to.
	LocationUUID string `json:"location_uuid"`

	// True if the network is public. If private it will be false.
	PublicNet bool `json:"public_net"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// One of 'network', 'network_high' or 'network_insane'.
	NetworkType string `json:"network_type"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// Status indicates the status of the object.
	Status string `json:"status"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// Defines information about MAC spoofing protection (filters layer2 and ARP traffic based on MAC source).
	// It can only be (de-)activated on a private network - the public network always has l2security enabled.
	// It will be true if the network is public, and false if the network is private.
	L2Security bool `json:"l2security"`

	// Defines the information if dhcp is activated for this network or not.
	DHCPActive bool `json:"dhcp_active"`

	// The general IP Range configured for this network (/24 for private networks).
	DHCPRange string `json:"dhcp_range"`

	// The ip reserved and communicated by the dhcp service to be the default gateway.
	DHCPGateway string `json:"dhcp_gateway"`

	DHCPDNS string `json:"dhcp_dns"`

	// Subrange within the ip range.
	DHCPReservedSubnet []string `json:"dhcp_reserved_subnet"`

	// Contains ips of all servers in the network which got a designated IP by the DHCP server.
	AutoAssignedServers []ServerWithIP `json:"auto_assigned_servers"`

	// Contains ips of all servers in the network which got a designated IP by the user.
	PinnedServers []ServerWithIP `json:"pinned_servers"`

	// Defines the date and time of the last object change.
	ChangeTime GSTime `json:"change_time"`

	// Uses IATA airport code, which works as a location identifier.
	LocationIata string `json:"location_iata"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationName string `json:"location_name"`

	// Defines if the object is administratively blocked. If true, it can not be deleted by the user.
	DeleteBlock bool `json:"delete_block"`

	// List of labels.
	Labels []string `json:"labels"`

	// The information about other object which are related to this network. the object could be servers and/or vlans.
	Relations NetworkRelations `json:"relations"`
}

// ServerWithIP holds a server's UUID and a corresponding IP address
type ServerWithIP struct {
	// UUID of the server
	ServerUUID string `json:"server_uuid"`

	// IP which is assigned to the server
	IP string `json:"ip"`
}

// NetworkRelations holds a list of a network's relations.
// The relation tells which VLANs/Servers/PaaS security zones relate to the network.
type NetworkRelations struct {
	// Array of object (NetworkVlan).
	Vlans []NetworkVlan `json:"vlans"`

	// Array of object (NetworkServer).
	Servers []NetworkServer `json:"servers"`

	// Array of object (NetworkPaaSSecurityZone).
	PaaSSecurityZones []NetworkPaaSSecurityZone `json:"paas_security_zones"`

	// Array of PaaS services that are connected to this network.
	PaaSServices []NetworkPaaSService `json:"paas_services"`
}

// NetworkVlan represents a relation between a network and a VLAN.
type NetworkVlan struct {
	// Vlan.
	Vlan int `json:"vlan"`

	// Name of tenant.
	TenantName string `json:"tenant_name"`

	// UUID of tenant.
	TenantUUID string `json:"tenant_uuid"`
}

// NetworkServer represents a relation between a network and a server.
type NetworkServer struct {
	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// Network_mac defines the MAC address of the network interface.
	Mac string `json:"mac"`

	// Whether the server boots from this iso image or not.
	Bootdevice bool `json:"bootdevice"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// Defines information about IP prefix spoof protection (it allows source traffic only from the IPv4/IPv4 network prefixes).
	// If empty, it allow no IPv4/IPv6 source traffic. If set to null, l3security is disabled (default).
	L3security []string `json:"l3security"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	ObjectName string `json:"object_name"`

	// The UUID of the network you're requesting.
	NetworkUUID string `json:"network_uuid"`

	// The ordering of the network interfaces. Lower numbers have lower PCI-IDs.
	Ordering int `json:"ordering"`
}

// NetworkPaaSSecurityZone represents a relation between a network and a PaaS security zone.
type NetworkPaaSSecurityZone struct {
	// IPv6 prefix of the PaaS service.
	IPv6Prefix string `json:"ipv6_prefix"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	ObjectName string `json:"object_name"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`
}

// NetworkPaaSService represents a relation between a network and a Network.
type NetworkPaaSService struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	ObjectName string `json:"object_name"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// Category of the PaaS service.
	ServiceTemplateCategory string `json:"service_template_category"`

	// The template used to create the service, you can find an available list at the /service_templates endpoint.
	ServiceTemplateUUID string `json:"service_template_uuid"`

	// Contains the IPv6/IPv4 address and port that the Service will listen to,
	// you can use these details to connect internally to a service.
	ListenPorts map[string]map[string]int `json:"listen_ports"`
}

// NetworkCreateRequest represents a request for creating a network.
type NetworkCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// List of labels. Can be empty.
	Labels []string `json:"labels,omitempty"`

	// Defines information about MAC spoofing protection (filters layer2 and ARP traffic based on MAC source).
	// It can only be (de-)activated on a private network - the public network always has l2security enabled.
	// It will be true if the network is public, and false if the network is private.
	L2Security bool `json:"l2security,omitempty"`

	// Defines the information if dhcp is activated for this network or not.
	DHCPActive bool `json:"dhcp_active,omitempty"`

	// The general IP Range configured for this network (/24 for private networks).
	DHCPRange string `json:"dhcp_range,omitempty"`

	// The ip reserved and communicated by the dhcp service to be the default gateway.
	DHCPGateway string `json:"dhcp_gateway,omitempty"`

	DHCPDNS string `json:"dhcp_dns,omitempty"`

	// Subrange within the ip range.
	DHCPReservedSubnet []string `json:"dhcp_reserved_subnet,omitempty"`
}

// NetworkCreateResponse represents a response for creating a network.
type NetworkCreateResponse struct {
	// UUID of the network being created.
	ObjectUUID string `json:"object_uuid"`

	// UUID of the request.
	RequestUUID string `json:"request_uuid"`
}

// NetworkUpdateRequest represents a request for updating a network.
type NetworkUpdateRequest struct {
	// New name. Leave it if you do not want to update the name.
	Name string `json:"name,omitempty"`

	// L2Security. Leave it if you do not want to update the l2 security.
	L2Security bool `json:"l2security"`

	// List of labels. Can be empty.
	Labels *[]string `json:"labels,omitempty"`

	// Defines the information if dhcp is activated for this network or not.
	DHCPActive *bool `json:"dhcp_active,omitempty"`

	// The general IP Range configured for this network (/24 for private networks).
	DHCPRange *string `json:"dhcp_range,omitempty"`

	// The ip reserved and communicated by the dhcp service to be the default gateway.
	DHCPGateway *string `json:"dhcp_gateway,omitempty"`

	DHCPDNS *string `json:"dhcp_dns,omitempty"`

	// Subrange within the ip range.
	DHCPReservedSubnet *[]string `json:"dhcp_reserved_subnet,omitempty"`
}

// PinnedServerList hold a list of pinned server with corresponding DCHP IP.
type PinnedServerList struct {
	// List of server and it's assigned DHCP IP
	List []ServerWithIP `json:"pinned_servers"`
}

// PinServerRequest represents a request assigning DHCP IP to a server
type PinServerRequest struct {
	// IP which is assigned to the server
	IP string `json:"ip"`
}

// GetNetwork get a specific network based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getNetwork
func (c *Client) GetNetwork(ctx context.Context, id string) (Network, error) {
	if !isValidUUID(id) {
		return Network{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiNetworkBase, id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response Network
	err := r.execute(ctx, *c, &response)
	return response, err
}

// CreateNetwork creates a network.
//
// See: https://gridscale.io/en//api-documentation/index.html#tag/network
func (c *Client) CreateNetwork(ctx context.Context, body NetworkCreateRequest) (NetworkCreateResponse, error) {
	r := gsRequest{
		uri:    apiNetworkBase,
		method: http.MethodPost,
		body:   body,
	}
	var response NetworkCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// DeleteNetwork removed a specific network based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deleteNetwork
func (c *Client) DeleteNetwork(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiNetworkBase, id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// UpdateNetwork updates a specific network based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateNetwork
func (c *Client) UpdateNetwork(ctx context.Context, id string, body NetworkUpdateRequest) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiNetworkBase, id),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// GetNetworkList gets a list of available networks.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getNetworks
func (c *Client) GetNetworkList(ctx context.Context) ([]Network, error) {
	r := gsRequest{
		uri:                 apiNetworkBase,
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response NetworkList
	var networks []Network
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		networks = append(networks, Network{
			Properties: properties,
		})
	}
	return networks, err
}

// GetNetworkEventList gets a list of a network's events.
//
// See: https://gridscale.io/en//api-documentation/index.html#tag/network
func (c *Client) GetNetworkEventList(ctx context.Context, id string) ([]Event, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiNetworkBase, id, "events"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response EventList
	var networkEvents []Event
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		networkEvents = append(networkEvents, Event{Properties: properties})
	}
	return networkEvents, err
}

// GetNetworkPublic gets the public network.
func (c *Client) GetNetworkPublic(ctx context.Context) (Network, error) {
	networks, err := c.GetNetworkList(ctx)
	if err != nil {
		return Network{}, err
	}
	for _, network := range networks {
		if network.Properties.PublicNet {
			return Network{Properties: network.Properties}, nil
		}
	}
	return Network{}, fmt.Errorf("public network not found")
}

// GetNetworksByLocation gets a list of networks by location.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getDeletedNetworks
func (c *Client) GetNetworksByLocation(ctx context.Context, id string) ([]Network, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiLocationBase, id, "networks"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response NetworkList
	var networks []Network
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		networks = append(networks, Network{Properties: properties})
	}
	return networks, err
}

// GetDeletedNetworks gets a list of deleted networks.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getDeletedNetworks
func (c *Client) GetDeletedNetworks(ctx context.Context) ([]Network, error) {
	r := gsRequest{
		uri:                 path.Join(apiDeletedBase, "networks"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response DeletedNetworkList
	var networks []Network
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		networks = append(networks, Network{Properties: properties})
	}
	return networks, err
}

// GetPinnedServerList returns a list of pinned servers in a specific network.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getNetworkPinnedServers
func (c *Client) GetPinnedServerList(ctx context.Context, networkUUID string) (PinnedServerList, error) {
	if !isValidUUID(networkUUID) {
		return PinnedServerList{}, errors.New("'networkUUID' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiNetworkBase, networkUUID, "pinned_servers"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response PinnedServerList
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdateNetworkPinnedServer assigns DHCP IP to a server.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateNetworkPinnedServer
func (c *Client) UpdateNetworkPinnedServer(ctx context.Context, networkUUID, serverUUID string, body PinServerRequest) error {
	if !isValidUUID(networkUUID) {
		return errors.New("'networkUUID' is invalid")
	}
	if !isValidUUID(serverUUID) {
		return errors.New("'serverUUID' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiNetworkBase, networkUUID, "pinned_servers", serverUUID),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeleteNetworkPinnedServer removes DHCP IP from a server.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateNetworkPinnedServer
func (c *Client) DeleteNetworkPinnedServer(ctx context.Context, networkUUID, serverUUID string) error {
	if !isValidUUID(networkUUID) {
		return errors.New("'networkUUID' is invalid")
	}
	if !isValidUUID(serverUUID) {
		return errors.New("'serverUUID' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiNetworkBase, networkUUID, "pinned_servers", serverUUID),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}
