package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// MarketplaceApplicationOperator aprovides an interface for operations on marketplace applications.
type MarketplaceApplicationOperator interface {
	GetMarketplaceApplicationList(ctx context.Context) ([]MarketplaceApplication, error)
	GetMarketplaceApplication(ctx context.Context, id string) (MarketplaceApplication, error)
	CreateMarketplaceApplication(ctx context.Context, body MarketplaceApplicationCreateRequest) (MarketplaceApplicationCreateResponse, error)
	ImportMarketplaceApplication(ctx context.Context, body MarketplaceApplicationImportRequest) (MarketplaceApplicationCreateResponse, error)
	UpdateMarketplaceApplication(ctx context.Context, id string, body MarketplaceApplicationUpdateRequest) error
	DeleteMarketplaceApplication(ctx context.Context, id string) error
	GetMarketplaceApplicationEventList(ctx context.Context, id string) ([]Event, error)
}

// MarketplaceApplicationList holds a list of market applications.
type MarketplaceApplicationList struct {
	// Array of market applications.
	List map[string]MarketplaceApplicationProperties `json:"applications"`
}

// MarketplaceApplication represent a single market application.
type MarketplaceApplication struct {
	// Properties of a market application.
	Properties MarketplaceApplicationProperties `json:"application"`
}

// MarketplaceApplicationProperties holds properties of a market application.
type MarketplaceApplicationProperties struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// Unique hash to allow user to import the self-created marketplace application.
	UniqueHash string `json:"unique_hash"`

	// Path to the images of the application.
	ObjectStoragePath string `json:"object_storage_path"`

	// Whether the you are the owner of application or not.
	IsApplicationOwner bool `json:"application_owner"`

	// Setup of the application.
	Setup MarketplaceApplicationSetup `json:"setup"`

	// Whether the template is published by the partner to their tenant.
	Published bool `json:"published"`

	// The date when the template is published into other tenant in the same partner.
	PublishedDate GSTime `json:"published_date"`

	// Whether the tenants want their template to be published or not.
	PublishRequested bool `json:"publish_requested"`

	// The date when the tenant requested their template to be published.
	PublishRequestedDate GSTime `json:"publish_requested_date"`

	// Whether a partner wants their tenant template published to other partners.
	PublishGlobalRequested bool `json:"publish_global_requested"`

	// The date when a partner requested their tenants template to be published.
	PublishGlobalRequestedDate GSTime `json:"publish_global_requested_date"`

	// Whether a template is published to other partner or not.
	PublishedGlobal bool `json:"published_global"`

	// The date when a template is published to other partner.
	PublishedGlobalDate GSTime `json:"published_global_date"`

	// Enum:"CMS", "project management", "Adminpanel", "Collaboration", "Cloud Storage", "Archiving".
	// Category of marketplace application.
	Category string `json:"category"`

	// Metadata of the Application.
	Metadata MarketplaceApplicationMetadata `json:"metadata"`

	// Defines the date and time of the last object change.
	ChangeTime GSTime `json:"change_time"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// Status indicates the status of the object.
	Status string `json:"status"`

	// The type of template.
	ApplicationType string `json:"application_type"`
}

// MarketplaceApplicationSetup represents marketplace application's setup.
type MarketplaceApplicationSetup struct {
	// Number of server cores.
	Cores int `json:"cores"`

	// The capacity of server memory in GB.
	Memory int `json:"memory"`

	// The capacity of a storage in GB.
	Capacity int `json:"capacity"`
}

// MarketplaceApplicationMetadata holds metadata of a marketplace application.
type MarketplaceApplicationMetadata struct {
	License    string   `json:"license"`
	OS         string   `json:"os"`
	Components []string `json:"components"`
	Overview   string   `json:"overview"`
	Hints      string   `json:"hints"`
	Icon       string   `json:"icon"`
	Features   string   `json:"features"`
	TermsOfUse string   `json:"terms_of_use"`
	Author     string   `json:"author"`
	Advices    string   `json:"advices"`
}

// MarketplaceApplicationCreateRequest represents a request for creating a marketplace application.
type MarketplaceApplicationCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// Path to the images for the application, must be in .gz format and started with "s3//"".
	ObjectStoragePath string `json:"object_storage_path"`

	// Category of the marketplace application. Allowed values: not-set, MarketplaceApplicationCMSCategory, MarketplaceApplicationProjectManagementCategory, MarketplaceApplicationAdminpanelCategory,
	// MarketplaceApplicationCollaborationCategory, MarketplaceApplicationCloudStorageCategory, MarketplaceApplicationArchivingCategory. Optional.
	Category MarketplaceApplicationCategory `json:"category,omitempty"`

	// whether you want to publish your application or not. Optional.
	Publish *bool `json:"publish,omitempty"`

	// Application's setup, consist the number of resource for creating the application.
	Setup MarketplaceApplicationSetup `json:"setup"`

	// Metadata of application.
	Metadata *MarketplaceApplicationMetadata `json:"metadata,omitempty"`
}

// MarketplaceApplicationImportRequest represents a request for importing a marketplace application.
type MarketplaceApplicationImportRequest struct {
	// Unique hash for importing this marketplace application.
	UniqueHash string `json:"unique_hash"`
}

// MarketplaceApplicationCreateResponse represents a response for a marketplace application's creation.
type MarketplaceApplicationCreateResponse struct {
	// UUID of the object being created.
	ObjectUUID string `json:"object_uuid"`

	// UUID of the request.
	RequestUUID string `json:"request_uuid"`

	// Unique hash for importing this marketplace application.
	UniqueHash string `json:"unique_hash"`
}

// MarketplaceApplicationUpdateRequest represents a request for updating a marketplace application.
type MarketplaceApplicationUpdateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters. Optional.
	Name string `json:"name,omitempty"`

	// Path to the images for the application, must be in .gz format and started with s3// . Optional.
	ObjectStoragePath string `json:"object_storage_path,omitempty"`

	// Category of the marketplace application. Allowed values: not-set, MarketplaceApplicationCMSCategory, MarketplaceApplicationProjectManagementCategory, MarketplaceApplicationAdminpanelCategory,
	// MarketplaceApplicationCollaborationCategory, MarketplaceApplicationCloudStorageCategory, MarketplaceApplicationArchivingCategory. Optional.
	Category MarketplaceApplicationCategory `json:"category,omitempty"`

	// Whether you want to publish your application or not. Optional.
	Publish *bool `json:"publish,omitempty"`

	// Application's setup, consist the number of resource for creating the application.
	Setup *MarketplaceApplicationSetup `json:"setup,omitempty"`

	// Metadata of application.
	Metadata *MarketplaceApplicationMetadata `json:"metadata,omitempty"`
}

// MarketplaceApplicationCategory represents the category in which a market application is.
type MarketplaceApplicationCategory string

// All allowed Marketplace application category's values.
var (
	MarketplaceApplicationCMSCategory               MarketplaceApplicationCategory = "CMS"
	MarketplaceApplicationProjectManagementCategory MarketplaceApplicationCategory = "project management"
	MarketplaceApplicationAdminpanelCategory        MarketplaceApplicationCategory = "Adminpanel"
	MarketplaceApplicationCollaborationCategory     MarketplaceApplicationCategory = "Collaboration"
	MarketplaceApplicationCloudStorageCategory      MarketplaceApplicationCategory = "Cloud Storage"
	MarketplaceApplicationArchivingCategory         MarketplaceApplicationCategory = "Archiving"
)

// GetMarketplaceApplicationList gets a list of available marketplace applications.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getMarketplaceApplications
func (c *Client) GetMarketplaceApplicationList(ctx context.Context) ([]MarketplaceApplication, error) {
	r := gsRequest{
		uri:                 apiMarketplaceApplicationBase,
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response MarketplaceApplicationList
	var marketApps []MarketplaceApplication
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		marketApps = append(marketApps, MarketplaceApplication{
			Properties: properties,
		})
	}
	return marketApps, err
}

// GetMarketplaceApplication gets a marketplace application.
//
// See https://gridscale.io/en//api-documentation/index.html#operation/getMarketplaceApplication
func (c *Client) GetMarketplaceApplication(ctx context.Context, id string) (MarketplaceApplication, error) {
	if !isValidUUID(id) {
		return MarketplaceApplication{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiMarketplaceApplicationBase, id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response MarketplaceApplication
	err := r.execute(ctx, *c, &response)
	return response, err
}

// CreateMarketplaceApplication creates a new marketplace application. Allowed
// values for Category are `nil`, `MarketplaceApplicationCMSCategory`,
// `MarketplaceApplicationProjectManagementCategory`,
// `MarketplaceApplicationAdminpanelCategory`,
// `MarketplaceApplicationCollaborationCategory`, `MarketplaceApplicationCloudStorageCategory`, `MarketplaceApplicationArchivingCategory`.
//
//See https://gridscale.io/en//api-documentation/index.html#operation/createMarketplaceApplication.
func (c *Client) CreateMarketplaceApplication(ctx context.Context, body MarketplaceApplicationCreateRequest) (MarketplaceApplicationCreateResponse, error) {
	r := gsRequest{
		uri:    apiMarketplaceApplicationBase,
		method: http.MethodPost,
		body:   body,
	}
	var response MarketplaceApplicationCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// ImportMarketplaceApplication imports a marketplace application. Allowed
// values for Category are `nil`, `MarketplaceApplicationCMSCategory`,
// `MarketplaceApplicationProjectManagementCategory`,
// `MarketplaceApplicationAdminpanelCategory`,
// `MarketplaceApplicationCollaborationCategory`,
// `MarketplaceApplicationCloudStorageCategory`,
// `MarketplaceApplicationArchivingCategory`.
//
// See https://gridscale.io/en//api-documentation/index.html#operation/createMarketplaceApplication.
func (c *Client) ImportMarketplaceApplication(ctx context.Context, body MarketplaceApplicationImportRequest) (MarketplaceApplicationCreateResponse, error) {
	r := gsRequest{
		uri:    apiMarketplaceApplicationBase,
		method: http.MethodPost,
		body:   body,
	}
	var response MarketplaceApplicationCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdateMarketplaceApplication updates a marketplace application.
//
// See https://gridscale.io/en//api-documentation/index.html#operation/updateMarketplaceApplication.
func (c *Client) UpdateMarketplaceApplication(ctx context.Context, id string, body MarketplaceApplicationUpdateRequest) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiMarketplaceApplicationBase, id),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeleteMarketplaceApplication removes a marketplace application.
//
// See https://gridscale.io/en//api-documentation/index.html#operation/deleteMarketplaceApplication.
func (c *Client) DeleteMarketplaceApplication(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiMarketplaceApplicationBase, id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// GetMarketplaceApplicationEventList gets list of a marketplace application's events.
//
// See https://gridscale.io/en//api-documentation/index.html#operation/getStorageEvents.
func (c *Client) GetMarketplaceApplicationEventList(ctx context.Context, id string) ([]Event, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiMarketplaceApplicationBase, id, "events"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response EventList
	var marketAppEvents []Event
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		marketAppEvents = append(marketAppEvents, Event{Properties: properties})
	}
	return marketAppEvents, err
}
