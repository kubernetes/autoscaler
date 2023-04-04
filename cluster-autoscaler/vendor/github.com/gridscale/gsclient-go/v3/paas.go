package gsclient

import (
	"context"
	"errors"
	"net/http"
	"path"
)

// PaaSOperator provides an interface for operations on PaaS-service-related resource.
type PaaSOperator interface {
	GetPaaSServiceList(ctx context.Context) ([]PaaSService, error)
	GetPaaSService(ctx context.Context, id string) (PaaSService, error)
	CreatePaaSService(ctx context.Context, body PaaSServiceCreateRequest) (PaaSServiceCreateResponse, error)
	UpdatePaaSService(ctx context.Context, id string, body PaaSServiceUpdateRequest) error
	DeletePaaSService(ctx context.Context, id string) error
	GetPaaSServiceMetrics(ctx context.Context, id string) ([]PaaSServiceMetric, error)
	GetPaaSTemplateList(ctx context.Context) ([]PaaSTemplate, error)
	GetDeletedPaaSServices(ctx context.Context) ([]PaaSService, error)
	RenewK8sCredentials(ctx context.Context, id string) error
	GetPaaSSecurityZoneList(ctx context.Context) ([]PaaSSecurityZone, error)
	GetPaaSSecurityZone(ctx context.Context, id string) (PaaSSecurityZone, error)
	CreatePaaSSecurityZone(ctx context.Context, body PaaSSecurityZoneCreateRequest) (PaaSSecurityZoneCreateResponse, error)
	UpdatePaaSSecurityZone(ctx context.Context, id string, body PaaSSecurityZoneUpdateRequest) error
	DeletePaaSSecurityZone(ctx context.Context, id string) error
}

// PaaSServices holds a list of available PaaS services.
type PaaSServices struct {
	// Array of PaaS services
	List map[string]PaaSServiceProperties `json:"paas_services"`
}

// DeletedPaaSServices provides a list of deleted PaaS services.
type DeletedPaaSServices struct {
	// Array of deleted PaaS services.
	List map[string]PaaSServiceProperties `json:"deleted_paas_services"`
}

// PaaSService represents a single PaaS service.
type PaaSService struct {
	// Properties of a PaaS service.
	Properties PaaSServiceProperties `json:"paas_service"`
}

// PaaSServiceProperties holds properties of a single PaaS service.
type PaaSServiceProperties struct {
	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// List of labels.
	Labels []string `json:"labels"`

	// Contains the initial setup credentials for Service.
	Credentials []Credential `json:"credentials"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// Contains the IPv6/IPv4 address and port that the Service will listen to,
	// you can use these details to connect internally to a service.
	ListenPorts map[string]map[string]int `json:"listen_ports"`

	// The UUID of the security zone that the service is attached to.
	SecurityZoneUUID string `json:"security_zone_uuid"`

	// The UUID of the network that the service is attached to.
	NetworkUUID string `json:"network_uuid"`

	// The template used to create the service, you can find an available list at the /service_templates endpoint.
	ServiceTemplateUUID string `json:"service_template_uuid"`

	// The template category used to create the service.
	ServiceTemplateCategory string `json:"service_template_category"`

	// Total minutes the object has been running.
	UsageInMinutes int `json:"usage_in_minutes"`

	// **DEPRECATED** The price for the current period since the last bill.
	CurrentPrice float64 `json:"current_price"`

	// Defines the date and time of the last object change.
	ChangeTime GSTime `json:"change_time"`

	// Status indicates the status of the object.
	Status string `json:"status"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// A list of service resource limits.
	ResourceLimits []ResourceLimit `json:"resource_limits"`

	// Contains the service parameters for the service.
	Parameters map[string]interface{} `json:"parameters"`
}

// Credential represents credential used to access a PaaS service.
type Credential struct {
	// The initial username to authenticate the Service.
	Username string `json:"username"`

	// The initial password to authenticate the Service.
	Password string `json:"password"`

	// The type of Service.
	Type string `json:"type"`

	// If the PaaS service is a k8s cluster, this field will be set.
	KubeConfig string `json:"kubeconfig"`

	// Expiration time of k8s credential.
	ExpirationTime GSTime `json:"expiration_time"`
}

// PaaSServiceCreateRequest represents a request for creating a PaaS service.
type PaaSServiceCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The template used to create the service, you can find an available list at the /service_templates endpoint.
	PaaSServiceTemplateUUID string `json:"paas_service_template_uuid"`

	// The list of labels.
	Labels []string `json:"labels,omitempty"`

	// The UUID of the security zone that the service is attached to.
	PaaSSecurityZoneUUID string `json:"paas_security_zone_uuid,omitempty"`

	// The UUID of the network that the service is attached to.
	NetworkUUID string `json:"network_uuid,omitempty"`

	// A list of service resource limits.
	ResourceLimits []ResourceLimit `json:"resource_limits,omitempty"`

	// Contains the service parameters for the service.
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ResourceLimit represents a resource limit.
// It is used to limit a specific computational resource in a PaaS service.
// e.g. it can be used to limit cpu count.
type ResourceLimit struct {
	// The name of the resource you would like to cap.
	Resource string `json:"resource"`

	// The maximum number of the specific resource your service can use.
	Limit int `json:"limit"`
}

// PaaSServiceCreateResponse represents a response for creating a PaaS service.
type PaaSServiceCreateResponse struct {
	// UUID of the request.
	RequestUUID string `json:"request_uuid"`

	// Contains the IPv6 address and port that the Service will listen to, you can use these details to connect internally to a service.
	ListenPorts map[string]map[string]int `json:"listen_ports"`

	// The template used to create the service, you can find an available list at the /service_templates endpoint.
	PaaSServiceUUID string `json:"paas_service_uuid"`

	// Contains the initial setup credentials for Service.
	Credentials []Credential `json:"credentials"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// A list of service resource limits.
	ResourceLimits []ResourceLimit `json:"resource_limits"`

	// Contains the service parameters for the service.
	Parameters map[string]interface{} `json:"parameters"`
}

// PaaSTemplates holds a list of PaaS Templates.
type PaaSTemplates struct {
	// Array of PaaS templates.
	List map[string]PaaSTemplateProperties `json:"paas_service_templates"`
}

// PaaSTemplate represents a single PaaS Template.
type PaaSTemplate struct {
	// Properties of a PaaS template.
	Properties PaaSTemplateProperties `json:"paas_service_template"`
}

// PaaSTemplateProperties holds properties of a PaaS template.
// A PaaS template can be retrieved and used to create a new PaaS service via the PaaS template UUID.
type PaaSTemplateProperties struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// Describes the category of the service.
	Category string `json:"category"`

	// Product No.
	ProductNo int `json:"product_no"`

	// Discounted product number related to the service template.
	DiscountProductNo int `json:"discount_product_no"`

	// Time period (seconds) for which the discounted product number is valid.
	DiscountPeriod int64 `json:"discount_period"`

	// List of labels.
	Labels []string `json:"labels"`

	// The amount of concurrent connections for the service.
	Resources Resource `json:"resources"`

	// Status indicates the status of the object.
	Status string `json:"status"`

	// A definition of possible service template parameters (python-cerberus compatible).
	ParametersSchema map[string]Parameter `json:"parameters_schema"`

	// Describes the flavour of the service. E.g. kubernetes, redis-store, postgres, etc.
	Flavour string `json:"flavour"`

	// Describes the version of the service.
	Version string `json:"version"`

	// Describes the release of the service.
	Release string `json:"release"`

	// Describes the performance class of the service.
	PerformanceClass string `json:"performance_class"`

	// List of service template uuids to which a performance class update is allowed.
	PerformanceClassUpdates []string `json:"performance_class_updates"`

	// List of service template uuids to which an upgrade is allowed.
	VersionUpgrades []string `json:"version_upgrades"`

	// List of service template uuids to which a patch update is allowed.
	PatchUpdates []string `json:"patch_updates"`

	// Values of the autoscaling resources.
	Autoscaling AutoscalingProperties `json:"autoscaling"`
}

// AutoscalingProperties holds properties of resource autoscalings.
type AutoscalingProperties struct {
	// Limit values of CPU core autoscaling.
	Cores AutoscalingResourceProperties `json:"cores"`

	// Limit values of storage autoscaling.
	Storage AutoscalingResourceProperties `json:"storage"`
}

// AutoscalingResourceProperties holds properties (Min/Max values)
// of a resource autoscaling.
type AutoscalingResourceProperties struct {
	// Min value of a resource autoscaling.
	Min int `json:"min"`

	// Max value of a resource autoscaling.
	Max int `json:"max"`
}

// Parameter represents a parameter used in PaaS template.
// Each type of PaaS service has diffrent set of parameters.
// Use method `GetPaaSTemplateList` to get infomation about
// parameters of a PaaS template.
type Parameter struct {
	// Is required.
	Required bool `json:"required"`

	// Is empty.
	Empty bool `json:"empty"`

	// Description of parameter.
	Description string `json:"description"`

	// Maximum.
	Max int `json:"max"`

	// Minimum.
	Min int `json:"min"`

	// Default value.
	Default interface{} `json:"default"`

	// Type of parameter.
	Type string `json:"type"`

	// Allowed values.
	Allowed []string `json:"allowed"`

	// Regex.
	Regex string `json:"regex"`

	// Immutable.
	Immutable bool `json:"immutable"`
}

// Resource represents the amount of concurrent connections for the service.
type Resource struct {
	// The amount of memory required by the service, either RAM(MB) or SSD Storage(GB).
	Memory int `json:"memory"`

	// The amount of concurrent connections for the service.
	Connections int `json:"connections"`

	// Storage type (one of storage, storage_high, storage_insane).
	StorageType string `json:"storage_type"`
}

// PaaSServiceUpdateRequest represetns a request for updating a PaaS service.
type PaaSServiceUpdateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	// Leave it if you do not want to update the name.
	Name string `json:"name,omitempty"`

	// List of labels. Leave it if you do not want to update the list of labels.
	Labels *[]string `json:"labels,omitempty"`

	// Contains the service parameters for the service. Leave it if you do not want to update the parameters.
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// A list of service resource limits. Leave it if you do not want to update the resource limits.
	ResourceLimits []ResourceLimit `json:"resource_limits,omitempty"`

	// The template that you want to use in the service, you can find an available list at the /service_templates endpoint.
	PaaSServiceTemplateUUID string `json:"service_template_uuid,omitempty"`

	// The UUID of the network that the service is attached to.
	NetworkUUID string `json:"network_uuid,omitempty"`
}

// PaaSServiceMetrics represents a list of metrics of a PaaS service.
type PaaSServiceMetrics struct {
	// Array of a PaaS service's metrics.
	List []PaaSMetricProperties `json:"paas_service_metrics"`
}

// PaaSServiceMetric represents a single metric of a PaaS service.
type PaaSServiceMetric struct {
	// Properties of a PaaS service metric.
	Properties PaaSMetricProperties `json:"paas_service_metric"`
}

// PaaSMetricProperties holds properties of a PaaS service metric.
type PaaSMetricProperties struct {
	// Defines the begin of the time range.
	BeginTime GSTime `json:"begin_time"`

	// Defines the end of the time range.
	EndTime GSTime `json:"end_time"`

	// The UUID of an object is always unique, and refers to a specific object.
	PaaSServiceUUID string `json:"paas_service_uuid"`

	// CPU core usage.
	CoreUsage PaaSMetricValue `json:"core_usage"`

	// Storage usage.
	StorageSize PaaSMetricValue `json:"storage_size"`
}

// PaaSMetricValue represents a PaaS metric value.
type PaaSMetricValue struct {
	// Value.
	Value float64 `json:"value"`

	// Unit of the value.
	Unit string `json:"unit"`
}

// PaaSSecurityZones holds a list of PaaS security zones.
type PaaSSecurityZones struct {
	// Array of security zones.
	List map[string]PaaSSecurityZoneProperties `json:"paas_security_zones"`
}

// PaaSSecurityZone represents a single PaaS security zone.
type PaaSSecurityZone struct {
	// Properties of a security zone.
	Properties PaaSSecurityZoneProperties `json:"paas_security_zone"`
}

// PaaSSecurityZoneProperties holds properties of a PaaS security zone.
// PaaS security zone can be retrieved and attached to PaaS services via security zone UUID,
// or attached to servers via the security zone's network UUID. To get the security zone's network UUID,
// check out methods `GetNetwork` and `GetNetworkList`, and retrieve the network relations.
type PaaSSecurityZoneProperties struct {
	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationCountry string `json:"location_country"`

	// Defines the date and time the object was initially created.
	CreateTime GSTime `json:"create_time"`

	// Uses IATA airport code, which works as a location identifier.
	LocationIata string `json:"location_iata"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`

	// List of labels.
	Labels []string `json:"labels"`

	// The human-readable name of the location. It supports the full UTF-8 character set, with a maximum of 64 characters.
	LocationName string `json:"location_name"`

	// Status indicates the status of the object.
	Status string `json:"status"`

	// Helps to identify which data center an object belongs to.
	LocationUUID string `json:"location_uuid"`

	// Defines the date and time of the last object change.
	ChangeTime GSTime `json:"change_time"`

	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`

	// object (PaaSRelationService).
	Relation PaaSRelationService `json:"relations"`
}

// PaaSRelationService represents a relation between a PaaS security zone and PaaS services.
type PaaSRelationService struct {
	// Array of object (ServiceObject).
	Services []ServiceObject `json:"services"`
}

// ServiceObject represents the UUID of a PaaS service relating to a PaaS security zone.
type ServiceObject struct {
	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`
}

// PaaSSecurityZoneCreateRequest represents a request for creating a PaaS security zone.
type PaaSSecurityZoneCreateRequest struct {
	// The human-readable name of the object. It supports the full UTF-8 character set, with a maximum of 64 characters.
	Name string `json:"name"`
}

// PaaSSecurityZoneCreateResponse represents a response for creating a PaaS security zone.
type PaaSSecurityZoneCreateResponse struct {
	// UUID of the request.
	RequestUUID string `json:"request_uuid"`

	// UUID of the security zone being created.
	PaaSSecurityZoneUUID string `json:"paas_security_zone_uuid"`

	// The UUID of an object is always unique, and refers to a specific object.
	ObjectUUID string `json:"object_uuid"`
}

// PaaSSecurityZoneUpdateRequest represents a request for updating a PaaS security zone.
type PaaSSecurityZoneUpdateRequest struct {
	// The new name you give to the security zone. Leave it if you do not want to update the name.
	Name string `json:"name,omitempty"`

	// The UUID for the security zone you would like to update. Leave it if you do not want to update the security zone.
	PaaSSecurityZoneUUID string `json:"paas_security_zone_uuid,omitempty"`
}

// GetPaaSServiceList returns a list of available PaaS Services.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getPaasServices
func (c *Client) GetPaaSServiceList(ctx context.Context) ([]PaaSService, error) {
	r := gsRequest{
		uri:                 path.Join(apiPaaSBase, "services"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response PaaSServices
	var paasServices []PaaSService
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		paasServices = append(paasServices, PaaSService{
			Properties: properties,
		})
	}
	return paasServices, err
}

// CreatePaaSService creates a new PaaS service.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/createPaasService
func (c *Client) CreatePaaSService(ctx context.Context, body PaaSServiceCreateRequest) (PaaSServiceCreateResponse, error) {
	r := gsRequest{
		uri:    path.Join(apiPaaSBase, "services"),
		method: http.MethodPost,
		body:   body,
	}
	var response PaaSServiceCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// GetPaaSService returns a specific PaaS Service based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getPaasService
func (c *Client) GetPaaSService(ctx context.Context, id string) (PaaSService, error) {
	if !isValidUUID(id) {
		return PaaSService{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiPaaSBase, "services", id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response PaaSService
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdatePaaSService updates a specific PaaS Service based on a given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updatePaasService
func (c *Client) UpdatePaaSService(ctx context.Context, id string, body PaaSServiceUpdateRequest) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiPaaSBase, "services", id),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeletePaaSService removes a PaaS service.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deletePaasService
func (c *Client) DeletePaaSService(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiPaaSBase, "services", id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// GetPaaSServiceMetrics get a specific PaaS Service's metrics based on a given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getPaasServiceMetrics
func (c *Client) GetPaaSServiceMetrics(ctx context.Context, id string) ([]PaaSServiceMetric, error) {
	if !isValidUUID(id) {
		return nil, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiPaaSBase, "services", id, "metrics"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response PaaSServiceMetrics
	var metrics []PaaSServiceMetric
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		metrics = append(metrics, PaaSServiceMetric{
			Properties: properties,
		})
	}
	return metrics, err
}

// RenewK8sCredentials renews credentials for gridscale Kubernetes PaaS service templates.
// The credentials of a PaaS service will be renewed (if supported by service template).
//
// See:https://gridscale.io/en/api-documentation/index.html#operation/renewPaasServiceCredentials
func (c *Client) RenewK8sCredentials(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiPaaSBase, "services", id, "renew_credentials"),
		method: http.MethodPatch,
		body:   emptyStruct{},
	}
	return r.execute(ctx, *c, nil)
}

// GetPaaSTemplateList returns a list of PaaS service templates.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getPaasServiceTemplates
func (c *Client) GetPaaSTemplateList(ctx context.Context) ([]PaaSTemplate, error) {
	r := gsRequest{
		uri:                 path.Join(apiPaaSBase, "service_templates"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response PaaSTemplates
	var paasTemplates []PaaSTemplate
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		paasTemplate := PaaSTemplate{
			Properties: properties,
		}
		paasTemplates = append(paasTemplates, paasTemplate)
	}
	return paasTemplates, err
}

// GetPaaSSecurityZoneList gets available security zones.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getPaasSecurityZones
func (c *Client) GetPaaSSecurityZoneList(ctx context.Context) ([]PaaSSecurityZone, error) {
	r := gsRequest{
		uri:                 path.Join(apiPaaSBase, "security_zones"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response PaaSSecurityZones
	var securityZones []PaaSSecurityZone
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		securityZones = append(securityZones, PaaSSecurityZone{
			Properties: properties,
		})
	}
	return securityZones, err
}

// CreatePaaSSecurityZone creates a new PaaS security zone.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/createPaasSecurityZone
func (c *Client) CreatePaaSSecurityZone(ctx context.Context, body PaaSSecurityZoneCreateRequest) (PaaSSecurityZoneCreateResponse, error) {
	r := gsRequest{
		uri:    path.Join(apiPaaSBase, "security_zones"),
		method: http.MethodPost,
		body:   body,
	}
	var response PaaSSecurityZoneCreateResponse
	err := r.execute(ctx, *c, &response)
	return response, err
}

// GetPaaSSecurityZone gets a specific PaaS Security Zone based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getPaasSecurityZone
func (c *Client) GetPaaSSecurityZone(ctx context.Context, id string) (PaaSSecurityZone, error) {
	if !isValidUUID(id) {
		return PaaSSecurityZone{}, errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:                 path.Join(apiPaaSBase, "security_zones", id),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response PaaSSecurityZone
	err := r.execute(ctx, *c, &response)
	return response, err
}

// UpdatePaaSSecurityZone updates a specific PaaS security zone based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/updatePaasSecurityZone
func (c *Client) UpdatePaaSSecurityZone(ctx context.Context, id string, body PaaSSecurityZoneUpdateRequest) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiPaaSBase, "security_zones", id),
		method: http.MethodPatch,
		body:   body,
	}
	return r.execute(ctx, *c, nil)
}

// DeletePaaSSecurityZone removes a specific PaaS Security Zone based on given id.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/deletePaasSecurityZone
func (c *Client) DeletePaaSSecurityZone(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	r := gsRequest{
		uri:    path.Join(apiPaaSBase, "security_zones", id),
		method: http.MethodDelete,
	}
	return r.execute(ctx, *c, nil)
}

// GetDeletedPaaSServices returns a list of deleted PaaS Services.
//
// See: https://gridscale.io/en//api-documentation/index.html#operation/getDeletedPaasServices
func (c *Client) GetDeletedPaaSServices(ctx context.Context) ([]PaaSService, error) {
	r := gsRequest{
		uri:                 path.Join(apiDeletedBase, "paas_services"),
		method:              http.MethodGet,
		skipCheckingRequest: true,
	}
	var response DeletedPaaSServices
	var paasServices []PaaSService
	err := r.execute(ctx, *c, &response)
	for _, properties := range response.List {
		paasServices = append(paasServices, PaaSService{
			Properties: properties,
		})
	}
	return paasServices, err
}
