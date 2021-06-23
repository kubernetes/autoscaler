package gobrightbox

import (
	"fmt"
)

type ServerType struct {
	Id       string
	Name     string
	Status   string
	Handle   string
	Cores    int
	Ram      int
	DiskSize int `json:"disk_size"`
}

func (c *Client) ServerTypes() ([]ServerType, error) {
	var servertypes []ServerType
	_, err := c.MakeApiRequest("GET", "/1.0/server_types", nil, &servertypes)
	if err != nil {
		return nil, err
	}
	return servertypes, err
}

func (c *Client) ServerType(identifier string) (*ServerType, error) {
	servertype := new(ServerType)
	_, err := c.MakeApiRequest("GET", "/1.0/server_types/"+identifier, nil, servertype)
	if err != nil {
		return nil, err
	}
	return servertype, err
}

func (c *Client) ServerTypeByHandle(handle string) (*ServerType, error) {
	servertypes, err := c.ServerTypes()
	if err != nil {
		return nil, err
	}
	for _, servertype := range servertypes {
		if servertype.Handle == handle {
			return &servertype, nil
		}
	}
	return nil, fmt.Errorf("ServerType with handle '%s' doesn't exist", handle)
}
