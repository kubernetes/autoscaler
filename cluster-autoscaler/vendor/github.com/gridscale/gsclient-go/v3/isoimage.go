package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// ISOImageOperator provides an interface for operations on ISO images.
type ISOImageOperator interface {
	GetISOImageList(ctx context.Context) ([]ISOImage, error)
	GetISOImage(ctx context.Context, id string) (ISOImage, error)
	CreateISOImage(ctx context.Context, body ISOImageCreateRequest) (ISOImageCreateResponse, error)
	UpdateISOImage(ctx context.Context, id string, body ISOImageUpdateRequest) error
	DeleteISOImage(ctx context.Context, id string) error
	GetISOImageEventList(ctx context.Context, id string) ([]Event, error)
	GetISOImagesByLocation(ctx context.Context, id string) ([]ISOImage, error)
	GetDeletedISOImages(ctx context.Context) ([]ISOImage, error)
}

// ISOImageList hold a list of ISO images.
type ISOImageList struct {
	// List of ISO-images.
	List map[string]ISOImageProperties `json:"isoimages"`
}

// DeletedISOImageList holds a list of deleted ISO images.
type DeletedISOImageList struct {
	// List of deleted ISO-images.
	List map[string]ISOImageProperties `json:"deleted_isoimages"`
}

// ISOImage represent a single ISO image.
type ISOImage struct {
	// Properties of an ISO-image.
	Properties ISOImageProperties `json:"isoimage"`
}

// ISOImageProperties holds properties of an ISO image.
// an ISO image can be retrieved and attached to servers via ISO image's UUID.
type ISOImageProperties struct {
	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// The information about other object which are related to this ISO image.
	Relations ISOImageRelation `json:"relations"`

	// Description of the ISO image release.
	Description string `json:"description"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationName string `json:"location_name"`

	// Contains the source URL of the ISO image that it was originally fetched from.
	SourceURL string `json:"source_url"`

	// List of labels.
	Labels []string `json:"labels"`

	// Uses IATA airport code, which works as a location identifier.
	LocationIata string `json:"location_iata"`

	// Helps to identify which data center an object belongs to.
	LocationUUID string `json:"location_uuid"`

	// Status indicates the status of the object.
	Status string `json:"status"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// Upstream version of the ISO image release.
	Version string `json:"version"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationCountry string `json:"location_country"`

	// Total minutes the object has been running.
	UsageInMinutes int `json:"usage_in_minutes"`

	// Whether the ISO image is private or not.
	Private bool `json:"private"`

	// Defines the date and time of the last object change.
	ChangeTime GSTime `json:"change_time"`

	// The capacity of an ISO image in GB.
	Capacity int `json:"capacity"`

	// **DEPRECATED** The price for the current period since the last bill.
	CurrentPrice float64 `json:"current_price"`
}

// ISOImageRelation represents a list of ISO image-server relations.
type ISOImageRelation struct {
	// Array of object (ServerinIsoimage).
	Servers []ServerinISOImage `json:"servers"`
}

// ServerinISOImage represents a relation between an ISO image and a Server.
type ServerinISOImage struct {
	// Whether the server boots from this iso image or not.
	Bootdevice bool `json:"bootdevice"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	ObjectName string `json:"object_name"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`
}

// ISOImageCreateRequest represents a request for creating an ISO image.
type ISOImageCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The source URL from which the ISO image should be downloaded.
	SourceURL string `json:"source_url"`

	// List of labels. Can be leave empty.
	Labels []string `json:"labels,omitempty"`
}

// ISOImageCreateResponse represents a response for creating an ISO image.
type ISOImageCreateResponse struct {
	// Request's UUID
	RequestUUID string `json:"request_uuid"`

	// UUID of an ISO-image being created.
	ObjectUUID string `json:"object_uuid"`
}

// ISOImageUpdateRequest represents a request for updating an ISO image.
type ISOImageUpdateRequest struct {
	// New name. Leave it if you do not want to update the name.
	Name string `json:"name,omitempty"`

	// List of labels. Leave it if you do not want to update the list of labels.
	Labels *[]string `json:"labels,omitempty"`
}

// GetISOImageList returns a list of available ISO images.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getIsoimages
func (c *Client) GetISOImageList(ctx context.Context) ([]ISOImage, error) {
	r := gsRequest{
		uri:                 path.Join(apiISOBase),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response ISOImageList
	var isoImages []ISOImage
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		isoImages = append(isoImages, ISOImage{Properties: properties})
	}
	return isoImages, err
}

// GetISOImage returns a specific ISO image based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getIsoimage
func (c *Client) GetISOImage(ctx context.Context, id string) (ISOImage, error) {
	if !isValidUUID(id) {
		return ISOImage{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiISOBase, id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response ISOImage
	err := r.execute(ctx, *c, &response)
	return response, err
}

// CreateISOImage creates an ISO image.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/createIsoimage
func (c *Client) CreateISOImage(ctx context.Context, body ISOImageCreateRequest) (ISOImageCreateResponse, error) {
	r := gsRequest{
		uri:    path.Join(apiISOBase),
		method: http.MethodPost,
		body:   body,
	}
	var response ISOImageCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdateISOImage updates a specific ISO Image.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateIsoimage
func (c *Client) UpdateISOImage(ctx context.Context, id string, body ISOImageUpdateRequest) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiISOBase, id),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeleteISOImage removes a specific ISO image.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deleteIsoimage
func (c *Client) DeleteISOImage(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiISOBase, id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// GetISOImageEventList returns a list of events of an ISO image.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getIsoimageEvents
func (c *Client) GetISOImageEventList(ctx context.Context, id string) ([]Event, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiISOBase, id, "events"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response EventList
	var isoImageEvents []Event
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		isoImageEvents = append(isoImageEvents, Event{Properties: properties})
	}
	return isoImageEvents, err
}

// GetISOImagesByLocation gets a list of ISO images by location.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getLocationIsoimages
func (c *Client) GetISOImagesByLocation(ctx context.Context, id string) ([]ISOImage, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiLocationBase, id, "isoimages"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response ISOImageList
	var isoImages []ISOImage
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		isoImages = append(isoImages, ISOImage{Properties: properties})
	}
	return isoImages, err
}

// GetDeletedISOImages gets a list of deleted ISO images.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getDeletedIsoimages
func (c *Client) GetDeletedISOImages(ctx context.Context) ([]ISOImage, error) {
	r := gsRequest{
		uri:                 path.Join(apiDeletedBase, "isoimages"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response DeletedISOImageList
	var isoImages []ISOImage
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		isoImages = append(isoImages, ISOImage{Properties: properties})
	}
	return isoImages, err
}
