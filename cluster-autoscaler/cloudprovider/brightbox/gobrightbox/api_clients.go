package gobrightbox

import (
	"time"
)

// ApiClient represents an API client.
// https://api.gb1.brightbox.com/1.0/#api_client
type ApiClient struct {
	Id               string
	Name             string
	Description      string
	Secret           string
	PermissionsGroup string     `json:"permissions_group"`
	RevokedAt        *time.Time `json:"revoked_at"`
	Account          Account
}

// ApiClientOptions is used in conjunction with CreateApiClient and
// UpdateApiClient to create and update api clients
type ApiClientOptions struct {
	Id               string  `json:"-"`
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	PermissionsGroup *string `json:"permissions_group,omitempty"`
}

// ApiClients retrieves a list of all API clients
func (c *Client) ApiClients() ([]ApiClient, error) {
	var apiClients []ApiClient
	_, err := c.MakeApiRequest("GET", "/1.0/api_clients", nil, &apiClients)
	if err != nil {
		return nil, err
	}
	return apiClients, err
}

// ApiClient retrieves a detailed view of one API client
func (c *Client) ApiClient(identifier string) (*ApiClient, error) {
	apiClient := new(ApiClient)
	_, err := c.MakeApiRequest("GET", "/1.0/api_clients/"+identifier, nil, apiClient)
	if err != nil {
		return nil, err
	}
	return apiClient, err
}

// CreateApiClient creates a new API client.
//
// It takes a ApiClientOptions struct for specifying name and other
// attributes. Not all attributes can be specified at create time
// (such as Id, which is allocated for you)
func (c *Client) CreateApiClient(options *ApiClientOptions) (*ApiClient, error) {
	ac := new(ApiClient)
	_, err := c.MakeApiRequest("POST", "/1.0/api_clients", options, &ac)
	if err != nil {
		return nil, err
	}
	return ac, nil
}

// UpdateApiClient updates an existing api client.
//
// It takes a ApiClientOptions struct for specifying Id, name and other
// attributes. Not all attributes can be specified at update time.
func (c *Client) UpdateApiClient(options *ApiClientOptions) (*ApiClient, error) {
	ac := new(ApiClient)
	_, err := c.MakeApiRequest("PUT", "/1.0/api_clients/"+options.Id, options, &ac)
	if err != nil {
		return nil, err
	}
	return ac, nil
}

// DestroyApiClient issues a request to deletes an existing api client
func (c *Client) DestroyApiClient(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/api_clients/"+identifier, nil, nil)
	return err
}

// ResetSecretForApiClient requests a snapshot of an existing api client
func (c *Client) ResetSecretForApiClient(identifier string) (*ApiClient, error) {
	ac := new(ApiClient)
	_, err := c.MakeApiRequest("POST", "/1.0/api_clients/"+identifier+"/reset_secret", nil, &ac)
	if err != nil {
		return nil, err
	}
	return ac, nil
}
