package hcloud

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud/schema"
)

// Datacenter represents a datacenter in the Hetzner Cloud.
type Datacenter struct {
	ID          int64
	Name        string
	Description string
	Location    *Location
	ServerTypes DatacenterServerTypes
}

// DatacenterServerTypes represents the server types available and supported in a datacenter.
type DatacenterServerTypes struct {
	Supported             []*ServerType
	AvailableForMigration []*ServerType
	Available             []*ServerType
}

// DatacenterClient is a client for the datacenter API.
type DatacenterClient struct {
	client *Client
}

// GetByID retrieves a datacenter by its ID. If the datacenter does not exist, nil is returned.
func (c *DatacenterClient) GetByID(ctx context.Context, id int64) (*Datacenter, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("/datacenters/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.DatacenterGetResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}
	return DatacenterFromSchema(body.Datacenter), resp, nil
}

// GetByName retrieves a datacenter by its name. If the datacenter does not exist, nil is returned.
func (c *DatacenterClient) GetByName(ctx context.Context, name string) (*Datacenter, *Response, error) {
	if name == "" {
		return nil, nil, nil
	}
	datacenters, response, err := c.List(ctx, DatacenterListOpts{Name: name})
	if len(datacenters) == 0 {
		return nil, response, err
	}
	return datacenters[0], response, err
}

// Get retrieves a datacenter by its ID if the input can be parsed as an integer, otherwise it
// retrieves a datacenter by its name. If the datacenter does not exist, nil is returned.
func (c *DatacenterClient) Get(ctx context.Context, idOrName string) (*Datacenter, *Response, error) {
	if id, err := strconv.ParseInt(idOrName, 10, 64); err == nil {
		return c.GetByID(ctx, id)
	}
	return c.GetByName(ctx, idOrName)
}

// DatacenterListOpts specifies options for listing datacenters.
type DatacenterListOpts struct {
	ListOpts
	Name string
	Sort []string
}

func (l DatacenterListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of datacenters for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *DatacenterClient) List(ctx context.Context, opts DatacenterListOpts) ([]*Datacenter, *Response, error) {
	path := "/datacenters?" + opts.values().Encode()
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.DatacenterListResponse
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}
	datacenters := make([]*Datacenter, 0, len(body.Datacenters))
	for _, i := range body.Datacenters {
		datacenters = append(datacenters, DatacenterFromSchema(i))
	}
	return datacenters, resp, nil
}

// All returns all datacenters.
func (c *DatacenterClient) All(ctx context.Context) ([]*Datacenter, error) {
	return c.AllWithOpts(ctx, DatacenterListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all datacenters for the given options.
func (c *DatacenterClient) AllWithOpts(ctx context.Context, opts DatacenterListOpts) ([]*Datacenter, error) {
	allDatacenters := []*Datacenter{}

	err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		datacenters, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allDatacenters = append(allDatacenters, datacenters...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allDatacenters, nil
}
