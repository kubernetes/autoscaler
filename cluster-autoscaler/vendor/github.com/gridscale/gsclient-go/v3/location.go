package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// LocationOperator provides an interface for operations on locations.
type LocationOperator interface {
	GetLocationList(ctx context.Context) ([]Location, error)
	GetLocation(ctx context.Context, id string) (Location, error)
	CreateLocation(ctx context.Context, body LocationCreateRequest) (CreateResponse, error)
	UpdateLocation(ctx context.Context, id string, body LocationUpdateRequest) error
	DeleteLocation(ctx context.Context, id string) error
}

// LocationList holds a list of locations.
type LocationList struct {
	// Array of locations.
	List map[string]LocationProperties `json:"locations"`
}

// Location represent a single location.
type Location struct {
	// Properties of a location.
	Properties LocationProperties `json:"location"`
}

// LocationProperties holds properties of a location.
type LocationProperties struct {
	// Uses IATA airport code, which works as a location identifier.
	Iata string `json:"iata"`

	// Status indicates the status of the object. DEPRECATED
	Status string `json:"status"`

	// List of labels.
	Labels []string `json:"labels"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Country string `json:"country"`

	// True if the location is active.
	Active bool `json:"active"`

	// Request change.
	ChangeRequested LocationRequestedChange `json:"change_requested"`

	// The number of dedicated cpunodes to assign to the private location.
	CPUNodeCount int `json:"cpunode_count"`

	// If this location is publicly available or a private location.
	Public bool `json:"public"`

	// The product number of a valid and available dedicated cpunode article.
	ProductNo int `json:"product_no"`

	// More detail about the location.
	LocationInformation LocationInformation `json:"location_information"`

	// Feature information of the location.
	Features LocationFeatures `json:"features"`
}

// LocationRequestedChange represents a location's requested change.
type LocationRequestedChange struct {
	// The requested number of dedicated cpunodes.
	CPUNodeCount int `json:"cpunode_count"`

	// The product number of a valid and available dedicated cpunode article.
	ProductNo int `json:"product_no"`

	// The location_uuid of an existing public location in which to create the private location.
	ParentLocationUUID string `json:"parent_location_uuid"`
}

// LocationInformation represents a location's detail information.
type LocationInformation struct {
	// List of certifications.
	CertificationList string `json:"certification_list"`

	// City of the locations.
	City string `json:"city"`

	// Data protection agreement.
	DataProtectionAgreement string `json:"data_protection_agreement"`

	// Geo Location.
	GeoLocation string `json:"geo_location"`

	// Green energy.
	GreenEnergy string `json:"green_energy"`

	// List of operator certificate.
	OperatorCertificationList string `json:"operator_certification_list"`

	// Owner of the location.
	Owner string `json:"owner"`

	// Website of the owner.
	OwnerWebsite string `json:"owner_website"`

	// The name of site.
	SiteName string `json:"site_name"`
}

// LocationFeatures represent a location's list of features.
type LocationFeatures struct {
	// List of available hardware profiles.
	HardwareProfiles string `json:"hardware_profiles"`

	// TRUE if the location has rocket storage feature.
	HasRocketStorage string `json:"has_rocket_storage"`

	// TRUE if the location has permission to provision server.
	HasServerProvisioning string `json:"has_server_provisioning"`

	// Region of object storage.
	ObjectStorageRegion string `json:"object_storage_region"`

	// Backup location UUID.
	BackupCenterLocationUUID string `json:"backup_center_location_uuid"`
}

// LocationCreateRequest represent a payload of a request for creating a new location.
type LocationCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 charset, with a maximum of 64 characters.
	Name string `json:"name"`

	// List of labels.
	Labels []string `json:"labels,omitempty"`

	// The location_uuid of an existing public location in which to create the private location.
	ParentLocationUUID string `json:"parent_location_uuid"`

	// The number of dedicated cpunodes to assigne to the private location.
	CPUNodeCount int `json:"cpunode_count"`

	// The product number of a valid and available dedicated cpunode article.
	ProductNo int `json:"product_no"`
}

// LocationUpdateRequest represents a request for updating a location.
type LocationUpdateRequest struct {
	// Name is the human-readable name of the object. Name is an optional field.
	Name string `json:"name,omitempty"`

	// List of labels. Labels is an optional field.
	Labels *[]string `json:"labels,omitempty"`

	// The number of dedicated cpunodes to assigne to the private location.
	// CPUNodeCount is an optional field.
	CPUNodeCount *int `json:"cpunode_count,omitempty"`
}

// GetLocationList gets a list of available locations.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getLocations
func (c *Client) GetLocationList(ctx context.Context) ([]Location, error) {
	r := gsRequest{
		uri:                 apiLocationBase,
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response LocationList
	var locations []Location
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		locations = append(locations, Location{Properties: properties})
	}
	return locations, err
}

// GetLocation gets a specific location.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getLocation
func (c *Client) GetLocation(ctx context.Context, id string) (Location, error) {
	if !isValidUUID(id) {
		return Location{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiLocationBase, id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var location Location
	err := r.execute(ctx, *c, &location)
	return location, err
}

// CreateLocation creates a new location.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/createLocation
func (c *Client) CreateLocation(ctx context.Context, body LocationCreateRequest) (CreateResponse, error) {
	r := gsRequest{
		uri:    apiLocationBase,
		method: http.MethodPost,
		body:   body,
	}
	var response CreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdateLocation updates a specific location based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updateLocation
func (c *Client) UpdateLocation(ctx context.Context, id string, body LocationUpdateRequest) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiLocationBase, id),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeleteLocation removes a single Location.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deleteLocation
func (c *Client) DeleteLocation(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiLocationBase, id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}
