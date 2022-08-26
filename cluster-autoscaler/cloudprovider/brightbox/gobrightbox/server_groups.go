package gobrightbox

import (
	"time"
)

// ServerGroup represents a server group
// https://api.gb1.brightbox.com/1.0/#server_group
type ServerGroup struct {
	Id             string
	Name           string
	CreatedAt      *time.Time `json:"created_at"`
	Description    string
	Default        bool
	Fqdn           string
	Account        Account `json:"account"`
	Servers        []Server
	FirewallPolicy *FirewallPolicy `json:"firewall_policy"`
}

// ServerGroupOptions is used in combination with CreateServerGroup and
// UpdateServerGroup to create and update server groups
type ServerGroupOptions struct {
	Id          string  `json:"-"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type serverGroupMemberOptions struct {
	Servers     []serverGroupMember `json:"servers"`
	Destination string              `json:"destination,omitempty"`
}
type serverGroupMember struct {
	Server string `json:"server,omitempty"`
}

// ServerGroups retrieves a list of all server groups
func (c *Client) ServerGroups() ([]ServerGroup, error) {
	var groups []ServerGroup
	_, err := c.MakeApiRequest("GET", "/1.0/server_groups", nil, &groups)
	if err != nil {
		return nil, err
	}
	return groups, err
}

// ServerGroup retrieves a detailed view on one server group
func (c *Client) ServerGroup(identifier string) (*ServerGroup, error) {
	group := new(ServerGroup)
	_, err := c.MakeApiRequest("GET", "/1.0/server_groups/"+identifier, nil, group)
	if err != nil {
		return nil, err
	}
	return group, err
}

// CreateServerGroup creates a new server group
//
// It takes an instance of ServerGroupOptions. Not all attributes can be
// specified at create time (such as Id, which is allocated for you).
func (c *Client) CreateServerGroup(newServerGroup *ServerGroupOptions) (*ServerGroup, error) {
	group := new(ServerGroup)
	_, err := c.MakeApiRequest("POST", "/1.0/server_groups", newServerGroup, &group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// UpdateServerGroup updates an existing server groups's attributes. Not all
// attributes can be changed (such as Id).
//
// Specify the server group you want to update using the ServerGroupOptions Id
// field.
//
// To change group memberships, use AddServersToServerGroup,
// RemoveServersFromServerGroup and MoveServersToServerGroup.
func (c *Client) UpdateServerGroup(updateServerGroup *ServerGroupOptions) (*ServerGroup, error) {
	group := new(ServerGroup)
	_, err := c.MakeApiRequest("PUT", "/1.0/server_groups/"+updateServerGroup.Id, updateServerGroup, &group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// DestroyServerGroup destroys an existing server group
func (c *Client) DestroyServerGroup(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/server_groups/"+identifier, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// AddServersToServerGroup adds servers to an existing server group.
//
// The identifier parameter specifies the destination group.
//
// The serverIds paramater specifies the identifiers of the servers you want to add.
func (c *Client) AddServersToServerGroup(identifier string, serverIds []string) (*ServerGroup, error) {
	group := new(ServerGroup)
	opts := new(serverGroupMemberOptions)
	for _, id := range serverIds {
		opts.Servers = append(opts.Servers, serverGroupMember{Server: id})
	}
	_, err := c.MakeApiRequest("POST", "/1.0/server_groups/"+identifier+"/add_servers", opts, &group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// RemoveServersToServerGroup removes servers from an existing server group.
//
// The identifier parameter specifies the group.
//
// The serverIds paramater specifies the identifiers of the servers you want to remove.
func (c *Client) RemoveServersFromServerGroup(identifier string, serverIds []string) (*ServerGroup, error) {
	group := new(ServerGroup)
	opts := new(serverGroupMemberOptions)
	for _, id := range serverIds {
		opts.Servers = append(opts.Servers, serverGroupMember{Server: id})
	}
	_, err := c.MakeApiRequest("POST", "/1.0/server_groups/"+identifier+"/remove_servers", opts, &group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// MoveServersToServerGroup atomically moves servers from one group to another.
//
// # The src parameter specifies the group to which the servers currently belong
//
// The dst parameter specifies the group to which you want to move the servers.
//
// The serverIds parameter specifies the identifiers of the servers you want to move.
func (c *Client) MoveServersToServerGroup(src string, dst string, serverIds []string) (*ServerGroup, error) {
	group := new(ServerGroup)
	opts := serverGroupMemberOptions{Destination: dst}
	for _, id := range serverIds {
		opts.Servers = append(opts.Servers, serverGroupMember{Server: id})
	}
	_, err := c.MakeApiRequest("POST", "/1.0/server_groups/"+src+"/move_servers", opts, &group)
	if err != nil {
		return nil, err
	}
	return group, nil
}
