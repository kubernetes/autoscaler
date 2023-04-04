package gsclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
)

// TemplateOperator provides an interface for operations on OS templates.
type TemplateOperator interface {
	GetTemplate(ctx context.Context, id string) (Template, error)
	GetTemplateByName(ctx context.Context, name string) (Template, error)
	GetTemplateList(ctx context.Context) ([]Template, error)
	CreateTemplate(ctx context.Context, body TemplateCreateRequest) (CreateResponse, error)
	UpdateTemplate(ctx context.Context, id string, body TemplateUpdateRequest) error
	DeleteTemplate(ctx context.Context, id string) error
	GetDeletedTemplates(ctx context.Context) ([]Template, error)
	GetTemplateEventList(ctx context.Context, id string) ([]Event, error)
}

// TemplateList holds a list of templates.
type TemplateList struct {
	// Array of templates.
	List map[string]TemplateProperties `json:"templates"`
}

// DeletedTemplateList Holds a list of deleted templates.
type DeletedTemplateList struct {
	// Array of deleted templates.
	List map[string]TemplateProperties `json:"deleted_templates"`
}

// Template represents a single OS template.
type Template struct {
	// Properties of a template.
	Properties TemplateProperties `json:"template"`
}

// TemplateProperties holds the properties of an OS template. OS templates can
// be selected by a user when creating new storages and attaching them to
// servers. Usually there are a fixed number of OS templates available and you
// would reference them by name or ObjectUUID.
type TemplateProperties struct {
	// Status indicates the status of the object.
	Status string `json:"status"`

	// Status indicates the status of the object.
	Ostype string `json:"ostype"`

	// Helps to identify which data center an object belongs to.
	LocationUUID string `json:"location_uuid"`

	// A version string for this template.
	Version string `json:"version"`

	// Description of the template.
	LocationIata string `json:"location_iata"`

	// Defines the date and time of the last change.
	ChangeTime GSTime `json:"change_time"`

	// Whether the object is private, the value will be true. Otherwise the value will be false.
	Private bool `json:"private"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// If a template has been used that requires a license key (e.g. Windows Servers)
	// this shows the product_no of the license (see the /prices endpoint for more details).
	LicenseProductNo int `json:"license_product_no"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// Total minutes the object has been running.
	UsageInMinutes int `json:"usage_in_minutes"`

	// The capacity of a storage/ISO image/template/snapshot in GiB.
	Capacity int `json:"capacity"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationName string `json:"location_name"`

	// The OS distribution of this template.
	Distro string `json:"distro"`

	// Description of the template.
	Description string `json:"description"`

	// **DEPRECATED** The price for the current period since the last bill.
	CurrentPrice float64 `json:"current_price"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationCountry string `json:"location_country"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// List of labels.
	Labels []string `json:"labels"`
}

// TemplateCreateRequest represents the request for creating a new OS template from an existing storage snapshot.
type TemplateCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// Snapshot UUID for template.
	SnapshotUUID string `json:"snapshot_uuid"`

	// List of labels. Optional.
	Labels []string `json:"labels,omitempty"`
}

// TemplateUpdateRequest represents a request to update a OS template.
type TemplateUpdateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	// Optional.
	Name string `json:"name,omitempty"`

	// List of labels. Optional.
	Labels *[]string `json:"labels,omitempty"`
}

// GetTemplate gets an OS template object by a given ID.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getTemplate
func (c *Client) GetTemplate(ctx context.Context, id string) (Template, error) {
	if !isValidUUID(id) {
		return Template{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiTemplateBase, id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response Template
	err := r.execute(ctx, *c, &response)
	return response, err
}

// GetTemplateList gets a list of OS templates.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getTemplates
func (c *Client) GetTemplateList(ctx context.Context) ([]Template, error) {
	r := gsRequest{
		uri:                 apiTemplateBase,
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response TemplateList
	var templates []Template
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		templates = append(templates, Template{
			Properties: properties,
		})
	}
	return templates, err
}

// GetTemplateByName retrieves a single template by its name. Use GetTemplate to
// retrieve a single template by it's ID.
func (c *Client) GetTemplateByName(ctx context.Context, name string) (Template, error) {
	if name == "" {
		return Template{}, errors.New("'name' is required")
	}
	templates, err := c.GetTemplateList(ctx)
	if err != nil {
		return Template{}, err
	}
	for _, template := range templates {
		if template.Properties.Name == name {
			return Template{Properties: template.Properties}, nil
		}
	}
	return Template{}, fmt.Errorf("Template %v not found", name)
}

// CreateTemplate creates a new OS template.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/createTemplate
func (c *Client) CreateTemplate(ctx context.Context, body TemplateCreateRequest) (CreateResponse, error) {
	r := gsRequest{
		uri:    apiTemplateBase,
		method: http.MethodPost,
		body:   body,
	}
	var response CreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdateTemplate updates an existing OS template's properties.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateTemplate
func (c *Client) UpdateTemplate(ctx context.Context, id string, body TemplateUpdateRequest) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiTemplateBase, id),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeleteTemplate removes a single OS template.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deleteTemplate
func (c *Client) DeleteTemplate(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiTemplateBase, id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// GetTemplateEventList gets the list of events that are associated with the
// given template.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getTemplateEvents
func (c *Client) GetTemplateEventList(ctx context.Context, id string) ([]Event, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiTemplateBase, id, "events"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response EventList
	var templateEvents []Event
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		templateEvents = append(templateEvents, Event{Properties: properties})
	}
	return templateEvents, err
}

// GetTemplatesByLocation gets a list of templates by location.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getLocationTemplates
func (c *Client) GetTemplatesByLocation(ctx context.Context, id string) ([]Template, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiLocationBase, id, "templates"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response TemplateList
	var templates []Template
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		templates = append(templates, Template{Properties: properties})
	}
	return templates, err
}

// GetDeletedTemplates gets a list of deleted templates.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getDeletedTemplates
func (c *Client) GetDeletedTemplates(ctx context.Context) ([]Template, error) {
	r := gsRequest{
		uri:                 path.Join(apiDeletedBase, "templates"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response DeletedTemplateList
	var templates []Template
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		templates = append(templates, Template{Properties: properties})
	}
	return templates, err
}
