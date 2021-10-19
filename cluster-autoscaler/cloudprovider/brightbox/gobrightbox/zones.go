package gobrightbox

import (
	"fmt"
)

type Zone struct {
	Id     string
	Handle string
}

func (c *Client) Zones() ([]Zone, error) {
	var zones []Zone
	_, err := c.MakeApiRequest("GET", "/1.0/zones", nil, &zones)
	if err != nil {
		return nil, err
	}
	return zones, err
}

func (c *Client) Zone(identifier string) (*Zone, error) {
	zone := new(Zone)
	_, err := c.MakeApiRequest("GET", "/1.0/zones/"+identifier, nil, zone)
	if err != nil {
		return nil, err
	}
	return zone, err
}

func (c *Client) ZoneByHandle(handle string) (*Zone, error) {
	zones, err := c.Zones()
	if err != nil {
		return nil, err
	}
	for _, zone := range zones {
		if zone.Handle == handle {
			return &zone, nil
		}
	}
	return nil, fmt.Errorf("Zone with handle '%s' doesn't exist", handle)
}
