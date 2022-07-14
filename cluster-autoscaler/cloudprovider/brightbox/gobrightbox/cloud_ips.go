package gobrightbox

import (
	"fmt"
)

// CloudIP represents a Cloud IP
// https://api.gb1.brightbox.com/1.0/#cloud_ip
type CloudIP struct {
	Id              string
	Name            string
	PublicIP        string `json:"public_ip"`
	PublicIPv4      string `json:"public_ipv4"`
	PublicIPv6      string `json:"public_ipv6"`
	Status          string
	ReverseDns      string           `json:"reverse_dns"`
	PortTranslators []PortTranslator `json:"port_translators"`
	Account         Account
	Fqdn            string
	Interface       *ServerInterface
	Server          *Server
	ServerGroup     *ServerGroup    `json:"server_group"`
	LoadBalancer    *LoadBalancer   `json:"load_balancer"`
	DatabaseServer  *DatabaseServer `json:"database_server"`
}

// PortTranslator represents a port translator on a Cloud IP
type PortTranslator struct {
	Incoming int    `json:"incoming"`
	Outgoing int    `json:"outgoing"`
	Protocol string `json:"protocol"`
}

// CloudIPOptions is used in conjunction with CreateCloudIP and UpdateCloudIP to
// create and update cloud IPs.
type CloudIPOptions struct {
	Id              string           `json:"-"`
	ReverseDns      *string          `json:"reverse_dns,omitempty"`
	Name            *string          `json:"name,omitempty"`
	PortTranslators []PortTranslator `json:"port_translators,omitempty"`
}

// CloudIPs retrieves a list of all cloud ips
func (c *Client) CloudIPs() ([]CloudIP, error) {
	var cloudips []CloudIP
	_, err := c.MakeApiRequest("GET", "/1.0/cloud_ips", nil, &cloudips)
	if err != nil {
		return nil, err
	}
	return cloudips, err
}

// CloudIP retrieves a detailed view of one cloud ip
func (c *Client) CloudIP(identifier string) (*CloudIP, error) {
	cloudip := new(CloudIP)
	_, err := c.MakeApiRequest("GET", "/1.0/cloud_ips/"+identifier, nil, cloudip)
	if err != nil {
		return nil, err
	}
	return cloudip, err
}

// DestroyCloudIP issues a request to destroy the cloud ip
func (c *Client) DestroyCloudIP(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/cloud_ips/"+identifier, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// CreateCloudIP creates a new Cloud IP.
//
// It takes a CloudIPOptions struct for specifying name and other attributes.
// Not all attributes can be specified at create time (such as Id, which is
// allocated for you)
func (c *Client) CreateCloudIP(newCloudIP *CloudIPOptions) (*CloudIP, error) {
	cloudip := new(CloudIP)
	_, err := c.MakeApiRequest("POST", "/1.0/cloud_ips", newCloudIP, &cloudip)
	if err != nil {
		return nil, err
	}
	return cloudip, nil
}

// UpdateCloudIP updates an existing cloud ip's attributes. Not all attributes
// can be changed after creation time (such as Id, which is allocated for you).
//
// Specify the cloud ip you want to update using the CloudIPOptions Id field
func (c *Client) UpdateCloudIP(updateCloudIP *CloudIPOptions) (*CloudIP, error) {
	cip := new(CloudIP)
	_, err := c.MakeApiRequest("PUT", "/1.0/cloud_ips/"+updateCloudIP.Id, updateCloudIP, &cip)
	if err != nil {
		return nil, err
	}
	return cip, nil
}

// MapCloudIP issues a request to map the cloud ip to the destination. The
// destination can be an identifier of any resource capable of receiving a Cloud
// IP, such as a server interface, a load balancer, or a cloud sql instace.
//
// To map a Cloud IP to a server, first lookup the server to get it's interface
// identifier (or use the MapCloudIPtoServer convenience method)
func (c *Client) MapCloudIP(identifier string, destination string) error {
	_, err := c.MakeApiRequest("POST", "/1.0/cloud_ips/"+identifier+"/map",
		map[string]string{"destination": destination}, nil)
	if err != nil {
		return err
	}
	return nil
}

// MapCloudIPtoServer is a convenience method to map a Cloud IP to a
// server. First looks up the server to get the network interface id. Uses the
// first interface found.
func (c *Client) MapCloudIPtoServer(identifier string, serverid string) error {
	server, err := c.Server(serverid)
	if err != nil {
		return err
	}
	if len(server.Interfaces) == 0 {
		return fmt.Errorf("Server %s has no interfaces to map cloud ip %s to", server.Id, identifier)
	}
	destination := server.Interfaces[0].Id
	err = c.MapCloudIP(identifier, destination)
	if err != nil {
		return err
	}
	return nil
}

// UnMapCloudIP issues a request to unmap the cloud ip.
func (c *Client) UnMapCloudIP(identifier string) error {
	_, err := c.MakeApiRequest("POST", "/1.0/cloud_ips/"+identifier+"/unmap", nil, nil)
	if err != nil {
		return err
	}
	return nil
}
