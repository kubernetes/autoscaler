/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package v2 provides primitives to interact the openapi HTTP API.
//
package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/deepmap/oapi-codegen/pkg/runtime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CdnConfiguration defines model for cdn-configuration.
type CdnConfiguration struct {
	Bucket *string `json:"bucket,omitempty"`
	Fqdn   *string `json:"fqdn,omitempty"`
	Status *string `json:"status,omitempty"`
}

// Healthcheck defines model for healthcheck.
type Healthcheck struct {
	Interval *int64  `json:"interval,omitempty"`
	Mode     *string `json:"mode,omitempty"`
	Port     *int64  `json:"port,omitempty"`
	Retries  *int64  `json:"retries,omitempty"`
	Timeout  *int64  `json:"timeout,omitempty"`
	TlsSni   *string `json:"tls-sni,omitempty"`
	Uri      *string `json:"uri,omitempty"`
}

// Instance defines model for instance.
type Instance struct {
	Id   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// InstanceType defines model for instance-type.
type InstanceType struct {
	Authorized *bool   `json:"authorized,omitempty"`
	Cpus       *int64  `json:"cpus,omitempty"`
	Family     *string `json:"family,omitempty"`
	Gpus       *int64  `json:"gpus,omitempty"`
	Id         *string `json:"id,omitempty"`
	Memory     *int64  `json:"memory,omitempty"`
	Size       *string `json:"size,omitempty"`
}

// LoadBalancer defines model for load-balancer.
type LoadBalancer struct {
	CreatedAt   *time.Time             `json:"created-at,omitempty"`
	Description *string                `json:"description,omitempty"`
	Id          *string                `json:"id,omitempty"`
	Ip          *string                `json:"ip,omitempty"`
	Name        *string                `json:"name,omitempty"`
	Services    *[]LoadBalancerService `json:"services,omitempty"`
	State       *string                `json:"state,omitempty"`
}

// LoadBalancerServerStatus defines model for load-balancer-server-status.
type LoadBalancerServerStatus struct {
	PublicIp *string `json:"public-ip,omitempty"`
	Status   *string `json:"status,omitempty"`
}

// LoadBalancerService defines model for load-balancer-service.
type LoadBalancerService struct {
	Description       *string                     `json:"description,omitempty"`
	Healthcheck       *Healthcheck                `json:"healthcheck,omitempty"`
	HealthcheckStatus *[]LoadBalancerServerStatus `json:"healthcheck-status,omitempty"`
	Id                *string                     `json:"id,omitempty"`
	InstancePool      *Resource                   `json:"instance-pool,omitempty"`
	Name              *string                     `json:"name,omitempty"`
	Port              *int64                      `json:"port,omitempty"`
	Protocol          *string                     `json:"protocol,omitempty"`
	State             *string                     `json:"state,omitempty"`
	Strategy          *string                     `json:"strategy,omitempty"`
	TargetPort        *int64                      `json:"target-port,omitempty"`
}

// Operation defines model for operation.
type Operation struct {
	Id        *string    `json:"id,omitempty"`
	Message   *string    `json:"message,omitempty"`
	Reason    *string    `json:"reason,omitempty"`
	Reference *Reference `json:"reference,omitempty"`
	State     *string    `json:"state,omitempty"`
}

// Reference defines model for reference.
type Reference struct {
	Command *string `json:"command,omitempty"`
	Id      *string `json:"id,omitempty"`
	Link    *string `json:"link,omitempty"`
}

// Resource defines model for resource.
type Resource struct {
	Id   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// SecurityGroup defines model for security-group.
type SecurityGroup struct {
	Description *string              `json:"description,omitempty"`
	Id          *string              `json:"id,omitempty"`
	Name        *string              `json:"name,omitempty"`
	Rules       *[]SecurityGroupRule `json:"rules,omitempty"`
}

// SecurityGroupResource defines model for security-group-resource.
type SecurityGroupResource struct {
	Id   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// SecurityGroupRule defines model for security-group-rule.
type SecurityGroupRule struct {
	Description   *string `json:"description,omitempty"`
	EndPort       *int64  `json:"end-port,omitempty"`
	FlowDirection *string `json:"flow-direction,omitempty"`
	Icmp          *struct {
		Code *int64 `json:"code,omitempty"`
		Type *int64 `json:"type,omitempty"`
	} `json:"icmp,omitempty"`
	Id            *string                `json:"id,omitempty"`
	Network       *string                `json:"network,omitempty"`
	Protocol      *string                `json:"protocol,omitempty"`
	SecurityGroup *SecurityGroupResource `json:"security-group,omitempty"`
	StartPort     *int64                 `json:"start-port,omitempty"`
}

// Snapshot defines model for snapshot.
type Snapshot struct {
	CreatedAt   *time.Time `json:"created-at,omitempty"`
	Description *string    `json:"description,omitempty"`
	Id          *string    `json:"id,omitempty"`
	Instance    *Instance  `json:"instance,omitempty"`
	Name        *string    `json:"name,omitempty"`
	State       *string    `json:"state,omitempty"`
}

// SnapshotExport defines model for snapshot-export.
type SnapshotExport struct {
	Id           *string `json:"id,omitempty"`
	Md5sum       *string `json:"md5sum,omitempty"`
	PresignedUrl *string `json:"presigned-url,omitempty"`
}

// Template defines model for template.
type Template struct {
	Build           *string    `json:"build,omitempty"`
	CreatedAt       *time.Time `json:"created-at,omitempty"`
	DefaultUser     *string    `json:"default-user,omitempty"`
	Description     *string    `json:"description,omitempty"`
	Family          *string    `json:"family,omitempty"`
	Id              *string    `json:"id,omitempty"`
	Name            *string    `json:"name,omitempty"`
	PasswordEnabled *bool      `json:"password-enabled,omitempty"`
	SshkeyEnabled   *bool      `json:"sshkey-enabled,omitempty"`
	Url             *string    `json:"url,omitempty"`
	Version         *string    `json:"version,omitempty"`
	Visibility      *string    `json:"visibility,omitempty"`
}

// Zone defines model for zone.
type Zone struct {
	Name *string `json:"name,omitempty"`
}

// CreateCdnConfigurationJSONBody defines parameters for CreateCdnConfiguration.
type CreateCdnConfigurationJSONBody CdnConfiguration

// CreateInstanceJSONBody defines parameters for CreateInstance.
type CreateInstanceJSONBody Instance

// CreateInstanceParams defines parameters for CreateInstance.
type CreateInstanceParams struct {
	Start *bool `json:"start,omitempty"`
}

// CreateLoadBalancerJSONBody defines parameters for CreateLoadBalancer.
type CreateLoadBalancerJSONBody LoadBalancer

// UpdateLoadBalancerJSONBody defines parameters for UpdateLoadBalancer.
type UpdateLoadBalancerJSONBody struct {
	Description *string `json:"description,omitempty"`
	Name        *string `json:"name,omitempty"`
}

// AddServiceToLoadBalancerJSONBody defines parameters for AddServiceToLoadBalancer.
type AddServiceToLoadBalancerJSONBody LoadBalancerService

// UpdateLoadBalancerServiceJSONBody defines parameters for UpdateLoadBalancerService.
type UpdateLoadBalancerServiceJSONBody struct {
	Description *string      `json:"description,omitempty"`
	Healthcheck *Healthcheck `json:"healthcheck,omitempty"`
	Name        *string      `json:"name,omitempty"`
	Port        *int64       `json:"port,omitempty"`
	Protocol    *string      `json:"protocol,omitempty"`
	Strategy    *string      `json:"strategy,omitempty"`
	TargetPort  *int64       `json:"target-port,omitempty"`
}

// CreateSecurityGroupJSONBody defines parameters for CreateSecurityGroup.
type CreateSecurityGroupJSONBody SecurityGroup

// AddRuleToSecurityGroupJSONBody defines parameters for AddRuleToSecurityGroup.
type AddRuleToSecurityGroupJSONBody SecurityGroupRule

// CreateCdnConfigurationRequestBody defines body for CreateCdnConfiguration for application/json ContentType.
type CreateCdnConfigurationJSONRequestBody CreateCdnConfigurationJSONBody

// CreateInstanceRequestBody defines body for CreateInstance for application/json ContentType.
type CreateInstanceJSONRequestBody CreateInstanceJSONBody

// CreateLoadBalancerRequestBody defines body for CreateLoadBalancer for application/json ContentType.
type CreateLoadBalancerJSONRequestBody CreateLoadBalancerJSONBody

// UpdateLoadBalancerRequestBody defines body for UpdateLoadBalancer for application/json ContentType.
type UpdateLoadBalancerJSONRequestBody UpdateLoadBalancerJSONBody

// AddServiceToLoadBalancerRequestBody defines body for AddServiceToLoadBalancer for application/json ContentType.
type AddServiceToLoadBalancerJSONRequestBody AddServiceToLoadBalancerJSONBody

// UpdateLoadBalancerServiceRequestBody defines body for UpdateLoadBalancerService for application/json ContentType.
type UpdateLoadBalancerServiceJSONRequestBody UpdateLoadBalancerServiceJSONBody

// CreateSecurityGroupRequestBody defines body for CreateSecurityGroup for application/json ContentType.
type CreateSecurityGroupJSONRequestBody CreateSecurityGroupJSONBody

// AddRuleToSecurityGroupRequestBody defines body for AddRuleToSecurityGroup for application/json ContentType.
type AddRuleToSecurityGroupJSONRequestBody AddRuleToSecurityGroupJSONBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A callback for modifying requests which are generated before sending over
	// the network.
	RequestEditor RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = http.DefaultClient
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditor = fn
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// ListCdnConfigurations request
	ListCdnConfigurations(ctx context.Context) (*http.Response, error)

	// CreateCdnConfiguration request  with any body
	CreateCdnConfigurationWithBody(ctx context.Context, contentType string, body io.Reader) (*http.Response, error)

	CreateCdnConfiguration(ctx context.Context, body CreateCdnConfigurationJSONRequestBody) (*http.Response, error)

	// DeleteCdnConfiguration request
	DeleteCdnConfiguration(ctx context.Context, bucket string) (*http.Response, error)

	// CreateInstance request  with any body
	CreateInstanceWithBody(ctx context.Context, params *CreateInstanceParams, contentType string, body io.Reader) (*http.Response, error)

	CreateInstance(ctx context.Context, params *CreateInstanceParams, body CreateInstanceJSONRequestBody) (*http.Response, error)

	// ListInstanceTypes request
	ListInstanceTypes(ctx context.Context) (*http.Response, error)

	// GetInstanceType request
	GetInstanceType(ctx context.Context, id string) (*http.Response, error)

	// CreateSnapshot request
	CreateSnapshot(ctx context.Context, id string) (*http.Response, error)

	// ListLoadBalancers request
	ListLoadBalancers(ctx context.Context) (*http.Response, error)

	// CreateLoadBalancer request  with any body
	CreateLoadBalancerWithBody(ctx context.Context, contentType string, body io.Reader) (*http.Response, error)

	CreateLoadBalancer(ctx context.Context, body CreateLoadBalancerJSONRequestBody) (*http.Response, error)

	// DeleteLoadBalancer request
	DeleteLoadBalancer(ctx context.Context, id string) (*http.Response, error)

	// GetLoadBalancer request
	GetLoadBalancer(ctx context.Context, id string) (*http.Response, error)

	// UpdateLoadBalancer request  with any body
	UpdateLoadBalancerWithBody(ctx context.Context, id string, contentType string, body io.Reader) (*http.Response, error)

	UpdateLoadBalancer(ctx context.Context, id string, body UpdateLoadBalancerJSONRequestBody) (*http.Response, error)

	// AddServiceToLoadBalancer request  with any body
	AddServiceToLoadBalancerWithBody(ctx context.Context, id string, contentType string, body io.Reader) (*http.Response, error)

	AddServiceToLoadBalancer(ctx context.Context, id string, body AddServiceToLoadBalancerJSONRequestBody) (*http.Response, error)

	// DeleteLoadBalancerService request
	DeleteLoadBalancerService(ctx context.Context, id string, serviceId string) (*http.Response, error)

	// GetLoadBalancerService request
	GetLoadBalancerService(ctx context.Context, id string, serviceId string) (*http.Response, error)

	// UpdateLoadBalancerService request  with any body
	UpdateLoadBalancerServiceWithBody(ctx context.Context, id string, serviceId string, contentType string, body io.Reader) (*http.Response, error)

	UpdateLoadBalancerService(ctx context.Context, id string, serviceId string, body UpdateLoadBalancerServiceJSONRequestBody) (*http.Response, error)

	// GetOperation request
	GetOperation(ctx context.Context, id string) (*http.Response, error)

	// Ping request
	Ping(ctx context.Context) (*http.Response, error)

	// ListSecurityGroups request
	ListSecurityGroups(ctx context.Context) (*http.Response, error)

	// CreateSecurityGroup request  with any body
	CreateSecurityGroupWithBody(ctx context.Context, contentType string, body io.Reader) (*http.Response, error)

	CreateSecurityGroup(ctx context.Context, body CreateSecurityGroupJSONRequestBody) (*http.Response, error)

	// DeleteSecurityGroup request
	DeleteSecurityGroup(ctx context.Context, id string) (*http.Response, error)

	// GetSecurityGroup request
	GetSecurityGroup(ctx context.Context, id string) (*http.Response, error)

	// AddRuleToSecurityGroup request  with any body
	AddRuleToSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader) (*http.Response, error)

	AddRuleToSecurityGroup(ctx context.Context, id string, body AddRuleToSecurityGroupJSONRequestBody) (*http.Response, error)

	// DeleteRuleFromSecurityGroup request
	DeleteRuleFromSecurityGroup(ctx context.Context, id string, ruleId string) (*http.Response, error)

	// ListSnapshots request
	ListSnapshots(ctx context.Context) (*http.Response, error)

	// DeleteSnapshot request
	DeleteSnapshot(ctx context.Context, id string) (*http.Response, error)

	// GetSnapshot request
	GetSnapshot(ctx context.Context, id string) (*http.Response, error)

	// GetExportSnapshot request
	GetExportSnapshot(ctx context.Context, id string) (*http.Response, error)

	// ExportSnapshot request
	ExportSnapshot(ctx context.Context, id string) (*http.Response, error)

	// GetTemplate request
	GetTemplate(ctx context.Context, id string) (*http.Response, error)

	// Version request
	Version(ctx context.Context) (*http.Response, error)

	// ListZones request
	ListZones(ctx context.Context) (*http.Response, error)
}

func (c *Client) ListCdnConfigurations(ctx context.Context) (*http.Response, error) {
	req, err := NewListCdnConfigurationsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) CreateCdnConfigurationWithBody(ctx context.Context, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewCreateCdnConfigurationRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) CreateCdnConfiguration(ctx context.Context, body CreateCdnConfigurationJSONRequestBody) (*http.Response, error) {
	req, err := NewCreateCdnConfigurationRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteCdnConfiguration(ctx context.Context, bucket string) (*http.Response, error) {
	req, err := NewDeleteCdnConfigurationRequest(c.Server, bucket)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) CreateInstanceWithBody(ctx context.Context, params *CreateInstanceParams, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewCreateInstanceRequestWithBody(c.Server, params, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) CreateInstance(ctx context.Context, params *CreateInstanceParams, body CreateInstanceJSONRequestBody) (*http.Response, error) {
	req, err := NewCreateInstanceRequest(c.Server, params, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) ListInstanceTypes(ctx context.Context) (*http.Response, error) {
	req, err := NewListInstanceTypesRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetInstanceType(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewGetInstanceTypeRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) CreateSnapshot(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewCreateSnapshotRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) ListLoadBalancers(ctx context.Context) (*http.Response, error) {
	req, err := NewListLoadBalancersRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) CreateLoadBalancerWithBody(ctx context.Context, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewCreateLoadBalancerRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) CreateLoadBalancer(ctx context.Context, body CreateLoadBalancerJSONRequestBody) (*http.Response, error) {
	req, err := NewCreateLoadBalancerRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteLoadBalancer(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewDeleteLoadBalancerRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetLoadBalancer(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewGetLoadBalancerRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateLoadBalancerWithBody(ctx context.Context, id string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewUpdateLoadBalancerRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateLoadBalancer(ctx context.Context, id string, body UpdateLoadBalancerJSONRequestBody) (*http.Response, error) {
	req, err := NewUpdateLoadBalancerRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) AddServiceToLoadBalancerWithBody(ctx context.Context, id string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewAddServiceToLoadBalancerRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) AddServiceToLoadBalancer(ctx context.Context, id string, body AddServiceToLoadBalancerJSONRequestBody) (*http.Response, error) {
	req, err := NewAddServiceToLoadBalancerRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteLoadBalancerService(ctx context.Context, id string, serviceId string) (*http.Response, error) {
	req, err := NewDeleteLoadBalancerServiceRequest(c.Server, id, serviceId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetLoadBalancerService(ctx context.Context, id string, serviceId string) (*http.Response, error) {
	req, err := NewGetLoadBalancerServiceRequest(c.Server, id, serviceId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateLoadBalancerServiceWithBody(ctx context.Context, id string, serviceId string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewUpdateLoadBalancerServiceRequestWithBody(c.Server, id, serviceId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateLoadBalancerService(ctx context.Context, id string, serviceId string, body UpdateLoadBalancerServiceJSONRequestBody) (*http.Response, error) {
	req, err := NewUpdateLoadBalancerServiceRequest(c.Server, id, serviceId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetOperation(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewGetOperationRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) Ping(ctx context.Context) (*http.Response, error) {
	req, err := NewPingRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) ListSecurityGroups(ctx context.Context) (*http.Response, error) {
	req, err := NewListSecurityGroupsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) CreateSecurityGroupWithBody(ctx context.Context, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewCreateSecurityGroupRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) CreateSecurityGroup(ctx context.Context, body CreateSecurityGroupJSONRequestBody) (*http.Response, error) {
	req, err := NewCreateSecurityGroupRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteSecurityGroup(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewDeleteSecurityGroupRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetSecurityGroup(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewGetSecurityGroupRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) AddRuleToSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewAddRuleToSecurityGroupRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) AddRuleToSecurityGroup(ctx context.Context, id string, body AddRuleToSecurityGroupJSONRequestBody) (*http.Response, error) {
	req, err := NewAddRuleToSecurityGroupRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteRuleFromSecurityGroup(ctx context.Context, id string, ruleId string) (*http.Response, error) {
	req, err := NewDeleteRuleFromSecurityGroupRequest(c.Server, id, ruleId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) ListSnapshots(ctx context.Context) (*http.Response, error) {
	req, err := NewListSnapshotsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteSnapshot(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewDeleteSnapshotRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetSnapshot(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewGetSnapshotRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetExportSnapshot(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewGetExportSnapshotRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) ExportSnapshot(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewExportSnapshotRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) GetTemplate(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewGetTemplateRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) Version(ctx context.Context) (*http.Response, error) {
	req, err := NewVersionRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) ListZones(ctx context.Context) (*http.Response, error) {
	req, err := NewListZonesRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if c.RequestEditor != nil {
		err = c.RequestEditor(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

// NewListCdnConfigurationsRequest generates requests for ListCdnConfigurations
func NewListCdnConfigurationsRequest(server string) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/cdn-configuration")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateCdnConfigurationRequest calls the generic CreateCdnConfiguration builder with application/json body
func NewCreateCdnConfigurationRequest(server string, body CreateCdnConfigurationJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateCdnConfigurationRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateCdnConfigurationRequestWithBody generates requests for CreateCdnConfiguration with any type of body
func NewCreateCdnConfigurationRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/cdn-configuration")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// NewDeleteCdnConfigurationRequest generates requests for DeleteCdnConfiguration
func NewDeleteCdnConfigurationRequest(server string, bucket string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "bucket", bucket)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/cdn-configuration/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateInstanceRequest calls the generic CreateInstance builder with application/json body
func NewCreateInstanceRequest(server string, params *CreateInstanceParams, body CreateInstanceJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateInstanceRequestWithBody(server, params, "application/json", bodyReader)
}

// NewCreateInstanceRequestWithBody generates requests for CreateInstance with any type of body
func NewCreateInstanceRequestWithBody(server string, params *CreateInstanceParams, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/instance")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	queryValues := queryUrl.Query()

	if params.Start != nil {

		if queryFrag, err := runtime.StyleParam("form", true, "start", *params.Start); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryUrl.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("POST", queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// NewListInstanceTypesRequest generates requests for ListInstanceTypes
func NewListInstanceTypesRequest(server string) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/instance-type")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetInstanceTypeRequest generates requests for GetInstanceType
func NewGetInstanceTypeRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/instance-type/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateSnapshotRequest generates requests for CreateSnapshot
func NewCreateSnapshotRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/instance/%s:create-snapshot", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListLoadBalancersRequest generates requests for ListLoadBalancers
func NewListLoadBalancersRequest(server string) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/load-balancer")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateLoadBalancerRequest calls the generic CreateLoadBalancer builder with application/json body
func NewCreateLoadBalancerRequest(server string, body CreateLoadBalancerJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateLoadBalancerRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateLoadBalancerRequestWithBody generates requests for CreateLoadBalancer with any type of body
func NewCreateLoadBalancerRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/load-balancer")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// NewDeleteLoadBalancerRequest generates requests for DeleteLoadBalancer
func NewDeleteLoadBalancerRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/load-balancer/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetLoadBalancerRequest generates requests for GetLoadBalancer
func NewGetLoadBalancerRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/load-balancer/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateLoadBalancerRequest calls the generic UpdateLoadBalancer builder with application/json body
func NewUpdateLoadBalancerRequest(server string, id string, body UpdateLoadBalancerJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateLoadBalancerRequestWithBody(server, id, "application/json", bodyReader)
}

// NewUpdateLoadBalancerRequestWithBody generates requests for UpdateLoadBalancer with any type of body
func NewUpdateLoadBalancerRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/load-balancer/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// NewAddServiceToLoadBalancerRequest calls the generic AddServiceToLoadBalancer builder with application/json body
func NewAddServiceToLoadBalancerRequest(server string, id string, body AddServiceToLoadBalancerJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewAddServiceToLoadBalancerRequestWithBody(server, id, "application/json", bodyReader)
}

// NewAddServiceToLoadBalancerRequestWithBody generates requests for AddServiceToLoadBalancer with any type of body
func NewAddServiceToLoadBalancerRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/load-balancer/%s/service", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// NewDeleteLoadBalancerServiceRequest generates requests for DeleteLoadBalancerService
func NewDeleteLoadBalancerServiceRequest(server string, id string, serviceId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParam("simple", false, "service-id", serviceId)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/load-balancer/%s/service/%s", pathParam0, pathParam1)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetLoadBalancerServiceRequest generates requests for GetLoadBalancerService
func NewGetLoadBalancerServiceRequest(server string, id string, serviceId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParam("simple", false, "service-id", serviceId)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/load-balancer/%s/service/%s", pathParam0, pathParam1)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateLoadBalancerServiceRequest calls the generic UpdateLoadBalancerService builder with application/json body
func NewUpdateLoadBalancerServiceRequest(server string, id string, serviceId string, body UpdateLoadBalancerServiceJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateLoadBalancerServiceRequestWithBody(server, id, serviceId, "application/json", bodyReader)
}

// NewUpdateLoadBalancerServiceRequestWithBody generates requests for UpdateLoadBalancerService with any type of body
func NewUpdateLoadBalancerServiceRequestWithBody(server string, id string, serviceId string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParam("simple", false, "service-id", serviceId)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/load-balancer/%s/service/%s", pathParam0, pathParam1)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// NewGetOperationRequest generates requests for GetOperation
func NewGetOperationRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/operation/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPingRequest generates requests for Ping
func NewPingRequest(server string) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/ping")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListSecurityGroupsRequest generates requests for ListSecurityGroups
func NewListSecurityGroupsRequest(server string) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/security-group")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateSecurityGroupRequest calls the generic CreateSecurityGroup builder with application/json body
func NewCreateSecurityGroupRequest(server string, body CreateSecurityGroupJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateSecurityGroupRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateSecurityGroupRequestWithBody generates requests for CreateSecurityGroup with any type of body
func NewCreateSecurityGroupRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/security-group")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// NewDeleteSecurityGroupRequest generates requests for DeleteSecurityGroup
func NewDeleteSecurityGroupRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/security-group/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetSecurityGroupRequest generates requests for GetSecurityGroup
func NewGetSecurityGroupRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/security-group/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewAddRuleToSecurityGroupRequest calls the generic AddRuleToSecurityGroup builder with application/json body
func NewAddRuleToSecurityGroupRequest(server string, id string, body AddRuleToSecurityGroupJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewAddRuleToSecurityGroupRequestWithBody(server, id, "application/json", bodyReader)
}

// NewAddRuleToSecurityGroupRequestWithBody generates requests for AddRuleToSecurityGroup with any type of body
func NewAddRuleToSecurityGroupRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/security-group/%s/rules", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)
	return req, nil
}

// NewDeleteRuleFromSecurityGroupRequest generates requests for DeleteRuleFromSecurityGroup
func NewDeleteRuleFromSecurityGroupRequest(server string, id string, ruleId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParam("simple", false, "rule-id", ruleId)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/security-group/%s/rules/%s", pathParam0, pathParam1)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListSnapshotsRequest generates requests for ListSnapshots
func NewListSnapshotsRequest(server string) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/snapshot")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeleteSnapshotRequest generates requests for DeleteSnapshot
func NewDeleteSnapshotRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/snapshot/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetSnapshotRequest generates requests for GetSnapshot
func NewGetSnapshotRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/snapshot/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetExportSnapshotRequest generates requests for GetExportSnapshot
func NewGetExportSnapshotRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/snapshot/%s:export", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewExportSnapshotRequest generates requests for ExportSnapshot
func NewExportSnapshotRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/snapshot/%s:export", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetTemplateRequest generates requests for GetTemplate
func NewGetTemplateRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParam("simple", false, "id", id)
	if err != nil {
		return nil, err
	}

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/template/%s", pathParam0)
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewVersionRequest generates requests for Version
func NewVersionRequest(server string) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/version")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListZonesRequest generates requests for ListZones
func NewListZonesRequest(server string) (*http.Request, error) {
	var err error

	queryUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	basePath := fmt.Sprintf("/zone")
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	queryUrl, err = queryUrl.Parse(basePath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// ListCdnConfigurations request
	ListCdnConfigurationsWithResponse(ctx context.Context) (*ListCdnConfigurationsResponse, error)

	// CreateCdnConfiguration request  with any body
	CreateCdnConfigurationWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader) (*CreateCdnConfigurationResponse, error)

	CreateCdnConfigurationWithResponse(ctx context.Context, body CreateCdnConfigurationJSONRequestBody) (*CreateCdnConfigurationResponse, error)

	// DeleteCdnConfiguration request
	DeleteCdnConfigurationWithResponse(ctx context.Context, bucket string) (*DeleteCdnConfigurationResponse, error)

	// CreateInstance request  with any body
	CreateInstanceWithBodyWithResponse(ctx context.Context, params *CreateInstanceParams, contentType string, body io.Reader) (*CreateInstanceResponse, error)

	CreateInstanceWithResponse(ctx context.Context, params *CreateInstanceParams, body CreateInstanceJSONRequestBody) (*CreateInstanceResponse, error)

	// ListInstanceTypes request
	ListInstanceTypesWithResponse(ctx context.Context) (*ListInstanceTypesResponse, error)

	// GetInstanceType request
	GetInstanceTypeWithResponse(ctx context.Context, id string) (*GetInstanceTypeResponse, error)

	// CreateSnapshot request
	CreateSnapshotWithResponse(ctx context.Context, id string) (*CreateSnapshotResponse, error)

	// ListLoadBalancers request
	ListLoadBalancersWithResponse(ctx context.Context) (*ListLoadBalancersResponse, error)

	// CreateLoadBalancer request  with any body
	CreateLoadBalancerWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader) (*CreateLoadBalancerResponse, error)

	CreateLoadBalancerWithResponse(ctx context.Context, body CreateLoadBalancerJSONRequestBody) (*CreateLoadBalancerResponse, error)

	// DeleteLoadBalancer request
	DeleteLoadBalancerWithResponse(ctx context.Context, id string) (*DeleteLoadBalancerResponse, error)

	// GetLoadBalancer request
	GetLoadBalancerWithResponse(ctx context.Context, id string) (*GetLoadBalancerResponse, error)

	// UpdateLoadBalancer request  with any body
	UpdateLoadBalancerWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader) (*UpdateLoadBalancerResponse, error)

	UpdateLoadBalancerWithResponse(ctx context.Context, id string, body UpdateLoadBalancerJSONRequestBody) (*UpdateLoadBalancerResponse, error)

	// AddServiceToLoadBalancer request  with any body
	AddServiceToLoadBalancerWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader) (*AddServiceToLoadBalancerResponse, error)

	AddServiceToLoadBalancerWithResponse(ctx context.Context, id string, body AddServiceToLoadBalancerJSONRequestBody) (*AddServiceToLoadBalancerResponse, error)

	// DeleteLoadBalancerService request
	DeleteLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string) (*DeleteLoadBalancerServiceResponse, error)

	// GetLoadBalancerService request
	GetLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string) (*GetLoadBalancerServiceResponse, error)

	// UpdateLoadBalancerService request  with any body
	UpdateLoadBalancerServiceWithBodyWithResponse(ctx context.Context, id string, serviceId string, contentType string, body io.Reader) (*UpdateLoadBalancerServiceResponse, error)

	UpdateLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string, body UpdateLoadBalancerServiceJSONRequestBody) (*UpdateLoadBalancerServiceResponse, error)

	// GetOperation request
	GetOperationWithResponse(ctx context.Context, id string) (*GetOperationResponse, error)

	// Ping request
	PingWithResponse(ctx context.Context) (*PingResponse, error)

	// ListSecurityGroups request
	ListSecurityGroupsWithResponse(ctx context.Context) (*ListSecurityGroupsResponse, error)

	// CreateSecurityGroup request  with any body
	CreateSecurityGroupWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader) (*CreateSecurityGroupResponse, error)

	CreateSecurityGroupWithResponse(ctx context.Context, body CreateSecurityGroupJSONRequestBody) (*CreateSecurityGroupResponse, error)

	// DeleteSecurityGroup request
	DeleteSecurityGroupWithResponse(ctx context.Context, id string) (*DeleteSecurityGroupResponse, error)

	// GetSecurityGroup request
	GetSecurityGroupWithResponse(ctx context.Context, id string) (*GetSecurityGroupResponse, error)

	// AddRuleToSecurityGroup request  with any body
	AddRuleToSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader) (*AddRuleToSecurityGroupResponse, error)

	AddRuleToSecurityGroupWithResponse(ctx context.Context, id string, body AddRuleToSecurityGroupJSONRequestBody) (*AddRuleToSecurityGroupResponse, error)

	// DeleteRuleFromSecurityGroup request
	DeleteRuleFromSecurityGroupWithResponse(ctx context.Context, id string, ruleId string) (*DeleteRuleFromSecurityGroupResponse, error)

	// ListSnapshots request
	ListSnapshotsWithResponse(ctx context.Context) (*ListSnapshotsResponse, error)

	// DeleteSnapshot request
	DeleteSnapshotWithResponse(ctx context.Context, id string) (*DeleteSnapshotResponse, error)

	// GetSnapshot request
	GetSnapshotWithResponse(ctx context.Context, id string) (*GetSnapshotResponse, error)

	// GetExportSnapshot request
	GetExportSnapshotWithResponse(ctx context.Context, id string) (*GetExportSnapshotResponse, error)

	// ExportSnapshot request
	ExportSnapshotWithResponse(ctx context.Context, id string) (*ExportSnapshotResponse, error)

	// GetTemplate request
	GetTemplateWithResponse(ctx context.Context, id string) (*GetTemplateResponse, error)

	// Version request
	VersionWithResponse(ctx context.Context) (*VersionResponse, error)

	// ListZones request
	ListZonesWithResponse(ctx context.Context) (*ListZonesResponse, error)
}

type ListCdnConfigurationsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		CdnConfigurations *[]CdnConfiguration `json:"cdn-configurations,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListCdnConfigurationsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListCdnConfigurationsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateCdnConfigurationResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateCdnConfigurationResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateCdnConfigurationResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteCdnConfigurationResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteCdnConfigurationResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteCdnConfigurationResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateInstanceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateInstanceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateInstanceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListInstanceTypesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		InstanceTypes *[]InstanceType `json:"instance-types,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListInstanceTypesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListInstanceTypesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetInstanceTypeResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *InstanceType
}

// Status returns HTTPResponse.Status
func (r GetInstanceTypeResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetInstanceTypeResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateSnapshotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateSnapshotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateSnapshotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListLoadBalancersResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		LoadBalancers *[]LoadBalancer `json:"load-balancers,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListLoadBalancersResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListLoadBalancersResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateLoadBalancerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateLoadBalancerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateLoadBalancerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteLoadBalancerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteLoadBalancerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteLoadBalancerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetLoadBalancerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *LoadBalancer
}

// Status returns HTTPResponse.Status
func (r GetLoadBalancerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetLoadBalancerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateLoadBalancerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdateLoadBalancerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateLoadBalancerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type AddServiceToLoadBalancerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r AddServiceToLoadBalancerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r AddServiceToLoadBalancerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteLoadBalancerServiceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteLoadBalancerServiceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteLoadBalancerServiceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetLoadBalancerServiceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *LoadBalancerService
}

// Status returns HTTPResponse.Status
func (r GetLoadBalancerServiceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetLoadBalancerServiceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateLoadBalancerServiceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdateLoadBalancerServiceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateLoadBalancerServiceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetOperationResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r GetOperationResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetOperationResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PingResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *string
}

// Status returns HTTPResponse.Status
func (r PingResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PingResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListSecurityGroupsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListSecurityGroupsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListSecurityGroupsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *SecurityGroup
}

// Status returns HTTPResponse.Status
func (r GetSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type AddRuleToSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r AddRuleToSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r AddRuleToSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteRuleFromSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteRuleFromSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteRuleFromSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListSnapshotsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Snapshots *[]Snapshot `json:"snapshots,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListSnapshotsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListSnapshotsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteSnapshotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteSnapshotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteSnapshotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetSnapshotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Snapshot
}

// Status returns HTTPResponse.Status
func (r GetSnapshotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetSnapshotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetExportSnapshotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *SnapshotExport
}

// Status returns HTTPResponse.Status
func (r GetExportSnapshotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetExportSnapshotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ExportSnapshotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ExportSnapshotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ExportSnapshotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetTemplateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Template
}

// Status returns HTTPResponse.Status
func (r GetTemplateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetTemplateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type VersionResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *string
}

// Status returns HTTPResponse.Status
func (r VersionResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r VersionResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListZonesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Zones *[]Zone `json:"zones,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListZonesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListZonesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ListCdnConfigurationsWithResponse request returning *ListCdnConfigurationsResponse
func (c *ClientWithResponses) ListCdnConfigurationsWithResponse(ctx context.Context) (*ListCdnConfigurationsResponse, error) {
	rsp, err := c.ListCdnConfigurations(ctx)
	if err != nil {
		return nil, err
	}
	return ParseListCdnConfigurationsResponse(rsp)
}

// CreateCdnConfigurationWithBodyWithResponse request with arbitrary body returning *CreateCdnConfigurationResponse
func (c *ClientWithResponses) CreateCdnConfigurationWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader) (*CreateCdnConfigurationResponse, error) {
	rsp, err := c.CreateCdnConfigurationWithBody(ctx, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseCreateCdnConfigurationResponse(rsp)
}

func (c *ClientWithResponses) CreateCdnConfigurationWithResponse(ctx context.Context, body CreateCdnConfigurationJSONRequestBody) (*CreateCdnConfigurationResponse, error) {
	rsp, err := c.CreateCdnConfiguration(ctx, body)
	if err != nil {
		return nil, err
	}
	return ParseCreateCdnConfigurationResponse(rsp)
}

// DeleteCdnConfigurationWithResponse request returning *DeleteCdnConfigurationResponse
func (c *ClientWithResponses) DeleteCdnConfigurationWithResponse(ctx context.Context, bucket string) (*DeleteCdnConfigurationResponse, error) {
	rsp, err := c.DeleteCdnConfiguration(ctx, bucket)
	if err != nil {
		return nil, err
	}
	return ParseDeleteCdnConfigurationResponse(rsp)
}

// CreateInstanceWithBodyWithResponse request with arbitrary body returning *CreateInstanceResponse
func (c *ClientWithResponses) CreateInstanceWithBodyWithResponse(ctx context.Context, params *CreateInstanceParams, contentType string, body io.Reader) (*CreateInstanceResponse, error) {
	rsp, err := c.CreateInstanceWithBody(ctx, params, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseCreateInstanceResponse(rsp)
}

func (c *ClientWithResponses) CreateInstanceWithResponse(ctx context.Context, params *CreateInstanceParams, body CreateInstanceJSONRequestBody) (*CreateInstanceResponse, error) {
	rsp, err := c.CreateInstance(ctx, params, body)
	if err != nil {
		return nil, err
	}
	return ParseCreateInstanceResponse(rsp)
}

// ListInstanceTypesWithResponse request returning *ListInstanceTypesResponse
func (c *ClientWithResponses) ListInstanceTypesWithResponse(ctx context.Context) (*ListInstanceTypesResponse, error) {
	rsp, err := c.ListInstanceTypes(ctx)
	if err != nil {
		return nil, err
	}
	return ParseListInstanceTypesResponse(rsp)
}

// GetInstanceTypeWithResponse request returning *GetInstanceTypeResponse
func (c *ClientWithResponses) GetInstanceTypeWithResponse(ctx context.Context, id string) (*GetInstanceTypeResponse, error) {
	rsp, err := c.GetInstanceType(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseGetInstanceTypeResponse(rsp)
}

// CreateSnapshotWithResponse request returning *CreateSnapshotResponse
func (c *ClientWithResponses) CreateSnapshotWithResponse(ctx context.Context, id string) (*CreateSnapshotResponse, error) {
	rsp, err := c.CreateSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseCreateSnapshotResponse(rsp)
}

// ListLoadBalancersWithResponse request returning *ListLoadBalancersResponse
func (c *ClientWithResponses) ListLoadBalancersWithResponse(ctx context.Context) (*ListLoadBalancersResponse, error) {
	rsp, err := c.ListLoadBalancers(ctx)
	if err != nil {
		return nil, err
	}
	return ParseListLoadBalancersResponse(rsp)
}

// CreateLoadBalancerWithBodyWithResponse request with arbitrary body returning *CreateLoadBalancerResponse
func (c *ClientWithResponses) CreateLoadBalancerWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader) (*CreateLoadBalancerResponse, error) {
	rsp, err := c.CreateLoadBalancerWithBody(ctx, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseCreateLoadBalancerResponse(rsp)
}

func (c *ClientWithResponses) CreateLoadBalancerWithResponse(ctx context.Context, body CreateLoadBalancerJSONRequestBody) (*CreateLoadBalancerResponse, error) {
	rsp, err := c.CreateLoadBalancer(ctx, body)
	if err != nil {
		return nil, err
	}
	return ParseCreateLoadBalancerResponse(rsp)
}

// DeleteLoadBalancerWithResponse request returning *DeleteLoadBalancerResponse
func (c *ClientWithResponses) DeleteLoadBalancerWithResponse(ctx context.Context, id string) (*DeleteLoadBalancerResponse, error) {
	rsp, err := c.DeleteLoadBalancer(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseDeleteLoadBalancerResponse(rsp)
}

// GetLoadBalancerWithResponse request returning *GetLoadBalancerResponse
func (c *ClientWithResponses) GetLoadBalancerWithResponse(ctx context.Context, id string) (*GetLoadBalancerResponse, error) {
	rsp, err := c.GetLoadBalancer(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseGetLoadBalancerResponse(rsp)
}

// UpdateLoadBalancerWithBodyWithResponse request with arbitrary body returning *UpdateLoadBalancerResponse
func (c *ClientWithResponses) UpdateLoadBalancerWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader) (*UpdateLoadBalancerResponse, error) {
	rsp, err := c.UpdateLoadBalancerWithBody(ctx, id, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseUpdateLoadBalancerResponse(rsp)
}

func (c *ClientWithResponses) UpdateLoadBalancerWithResponse(ctx context.Context, id string, body UpdateLoadBalancerJSONRequestBody) (*UpdateLoadBalancerResponse, error) {
	rsp, err := c.UpdateLoadBalancer(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return ParseUpdateLoadBalancerResponse(rsp)
}

// AddServiceToLoadBalancerWithBodyWithResponse request with arbitrary body returning *AddServiceToLoadBalancerResponse
func (c *ClientWithResponses) AddServiceToLoadBalancerWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader) (*AddServiceToLoadBalancerResponse, error) {
	rsp, err := c.AddServiceToLoadBalancerWithBody(ctx, id, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseAddServiceToLoadBalancerResponse(rsp)
}

func (c *ClientWithResponses) AddServiceToLoadBalancerWithResponse(ctx context.Context, id string, body AddServiceToLoadBalancerJSONRequestBody) (*AddServiceToLoadBalancerResponse, error) {
	rsp, err := c.AddServiceToLoadBalancer(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return ParseAddServiceToLoadBalancerResponse(rsp)
}

// DeleteLoadBalancerServiceWithResponse request returning *DeleteLoadBalancerServiceResponse
func (c *ClientWithResponses) DeleteLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string) (*DeleteLoadBalancerServiceResponse, error) {
	rsp, err := c.DeleteLoadBalancerService(ctx, id, serviceId)
	if err != nil {
		return nil, err
	}
	return ParseDeleteLoadBalancerServiceResponse(rsp)
}

// GetLoadBalancerServiceWithResponse request returning *GetLoadBalancerServiceResponse
func (c *ClientWithResponses) GetLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string) (*GetLoadBalancerServiceResponse, error) {
	rsp, err := c.GetLoadBalancerService(ctx, id, serviceId)
	if err != nil {
		return nil, err
	}
	return ParseGetLoadBalancerServiceResponse(rsp)
}

// UpdateLoadBalancerServiceWithBodyWithResponse request with arbitrary body returning *UpdateLoadBalancerServiceResponse
func (c *ClientWithResponses) UpdateLoadBalancerServiceWithBodyWithResponse(ctx context.Context, id string, serviceId string, contentType string, body io.Reader) (*UpdateLoadBalancerServiceResponse, error) {
	rsp, err := c.UpdateLoadBalancerServiceWithBody(ctx, id, serviceId, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseUpdateLoadBalancerServiceResponse(rsp)
}

func (c *ClientWithResponses) UpdateLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string, body UpdateLoadBalancerServiceJSONRequestBody) (*UpdateLoadBalancerServiceResponse, error) {
	rsp, err := c.UpdateLoadBalancerService(ctx, id, serviceId, body)
	if err != nil {
		return nil, err
	}
	return ParseUpdateLoadBalancerServiceResponse(rsp)
}

// GetOperationWithResponse request returning *GetOperationResponse
func (c *ClientWithResponses) GetOperationWithResponse(ctx context.Context, id string) (*GetOperationResponse, error) {
	rsp, err := c.GetOperation(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseGetOperationResponse(rsp)
}

// PingWithResponse request returning *PingResponse
func (c *ClientWithResponses) PingWithResponse(ctx context.Context) (*PingResponse, error) {
	rsp, err := c.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return ParsePingResponse(rsp)
}

// ListSecurityGroupsWithResponse request returning *ListSecurityGroupsResponse
func (c *ClientWithResponses) ListSecurityGroupsWithResponse(ctx context.Context) (*ListSecurityGroupsResponse, error) {
	rsp, err := c.ListSecurityGroups(ctx)
	if err != nil {
		return nil, err
	}
	return ParseListSecurityGroupsResponse(rsp)
}

// CreateSecurityGroupWithBodyWithResponse request with arbitrary body returning *CreateSecurityGroupResponse
func (c *ClientWithResponses) CreateSecurityGroupWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader) (*CreateSecurityGroupResponse, error) {
	rsp, err := c.CreateSecurityGroupWithBody(ctx, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseCreateSecurityGroupResponse(rsp)
}

func (c *ClientWithResponses) CreateSecurityGroupWithResponse(ctx context.Context, body CreateSecurityGroupJSONRequestBody) (*CreateSecurityGroupResponse, error) {
	rsp, err := c.CreateSecurityGroup(ctx, body)
	if err != nil {
		return nil, err
	}
	return ParseCreateSecurityGroupResponse(rsp)
}

// DeleteSecurityGroupWithResponse request returning *DeleteSecurityGroupResponse
func (c *ClientWithResponses) DeleteSecurityGroupWithResponse(ctx context.Context, id string) (*DeleteSecurityGroupResponse, error) {
	rsp, err := c.DeleteSecurityGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseDeleteSecurityGroupResponse(rsp)
}

// GetSecurityGroupWithResponse request returning *GetSecurityGroupResponse
func (c *ClientWithResponses) GetSecurityGroupWithResponse(ctx context.Context, id string) (*GetSecurityGroupResponse, error) {
	rsp, err := c.GetSecurityGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseGetSecurityGroupResponse(rsp)
}

// AddRuleToSecurityGroupWithBodyWithResponse request with arbitrary body returning *AddRuleToSecurityGroupResponse
func (c *ClientWithResponses) AddRuleToSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader) (*AddRuleToSecurityGroupResponse, error) {
	rsp, err := c.AddRuleToSecurityGroupWithBody(ctx, id, contentType, body)
	if err != nil {
		return nil, err
	}
	return ParseAddRuleToSecurityGroupResponse(rsp)
}

func (c *ClientWithResponses) AddRuleToSecurityGroupWithResponse(ctx context.Context, id string, body AddRuleToSecurityGroupJSONRequestBody) (*AddRuleToSecurityGroupResponse, error) {
	rsp, err := c.AddRuleToSecurityGroup(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return ParseAddRuleToSecurityGroupResponse(rsp)
}

// DeleteRuleFromSecurityGroupWithResponse request returning *DeleteRuleFromSecurityGroupResponse
func (c *ClientWithResponses) DeleteRuleFromSecurityGroupWithResponse(ctx context.Context, id string, ruleId string) (*DeleteRuleFromSecurityGroupResponse, error) {
	rsp, err := c.DeleteRuleFromSecurityGroup(ctx, id, ruleId)
	if err != nil {
		return nil, err
	}
	return ParseDeleteRuleFromSecurityGroupResponse(rsp)
}

// ListSnapshotsWithResponse request returning *ListSnapshotsResponse
func (c *ClientWithResponses) ListSnapshotsWithResponse(ctx context.Context) (*ListSnapshotsResponse, error) {
	rsp, err := c.ListSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	return ParseListSnapshotsResponse(rsp)
}

// DeleteSnapshotWithResponse request returning *DeleteSnapshotResponse
func (c *ClientWithResponses) DeleteSnapshotWithResponse(ctx context.Context, id string) (*DeleteSnapshotResponse, error) {
	rsp, err := c.DeleteSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseDeleteSnapshotResponse(rsp)
}

// GetSnapshotWithResponse request returning *GetSnapshotResponse
func (c *ClientWithResponses) GetSnapshotWithResponse(ctx context.Context, id string) (*GetSnapshotResponse, error) {
	rsp, err := c.GetSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseGetSnapshotResponse(rsp)
}

// GetExportSnapshotWithResponse request returning *GetExportSnapshotResponse
func (c *ClientWithResponses) GetExportSnapshotWithResponse(ctx context.Context, id string) (*GetExportSnapshotResponse, error) {
	rsp, err := c.GetExportSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseGetExportSnapshotResponse(rsp)
}

// ExportSnapshotWithResponse request returning *ExportSnapshotResponse
func (c *ClientWithResponses) ExportSnapshotWithResponse(ctx context.Context, id string) (*ExportSnapshotResponse, error) {
	rsp, err := c.ExportSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseExportSnapshotResponse(rsp)
}

// GetTemplateWithResponse request returning *GetTemplateResponse
func (c *ClientWithResponses) GetTemplateWithResponse(ctx context.Context, id string) (*GetTemplateResponse, error) {
	rsp, err := c.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}
	return ParseGetTemplateResponse(rsp)
}

// VersionWithResponse request returning *VersionResponse
func (c *ClientWithResponses) VersionWithResponse(ctx context.Context) (*VersionResponse, error) {
	rsp, err := c.Version(ctx)
	if err != nil {
		return nil, err
	}
	return ParseVersionResponse(rsp)
}

// ListZonesWithResponse request returning *ListZonesResponse
func (c *ClientWithResponses) ListZonesWithResponse(ctx context.Context) (*ListZonesResponse, error) {
	rsp, err := c.ListZones(ctx)
	if err != nil {
		return nil, err
	}
	return ParseListZonesResponse(rsp)
}

// ParseListCdnConfigurationsResponse parses an HTTP response from a ListCdnConfigurationsWithResponse call
func ParseListCdnConfigurationsResponse(rsp *http.Response) (*ListCdnConfigurationsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ListCdnConfigurationsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			CdnConfigurations *[]CdnConfiguration `json:"cdn-configurations,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateCdnConfigurationResponse parses an HTTP response from a CreateCdnConfigurationWithResponse call
func ParseCreateCdnConfigurationResponse(rsp *http.Response) (*CreateCdnConfigurationResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &CreateCdnConfigurationResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteCdnConfigurationResponse parses an HTTP response from a DeleteCdnConfigurationWithResponse call
func ParseDeleteCdnConfigurationResponse(rsp *http.Response) (*DeleteCdnConfigurationResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &DeleteCdnConfigurationResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateInstanceResponse parses an HTTP response from a CreateInstanceWithResponse call
func ParseCreateInstanceResponse(rsp *http.Response) (*CreateInstanceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &CreateInstanceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListInstanceTypesResponse parses an HTTP response from a ListInstanceTypesWithResponse call
func ParseListInstanceTypesResponse(rsp *http.Response) (*ListInstanceTypesResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ListInstanceTypesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			InstanceTypes *[]InstanceType `json:"instance-types,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetInstanceTypeResponse parses an HTTP response from a GetInstanceTypeWithResponse call
func ParseGetInstanceTypeResponse(rsp *http.Response) (*GetInstanceTypeResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetInstanceTypeResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest InstanceType
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateSnapshotResponse parses an HTTP response from a CreateSnapshotWithResponse call
func ParseCreateSnapshotResponse(rsp *http.Response) (*CreateSnapshotResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &CreateSnapshotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListLoadBalancersResponse parses an HTTP response from a ListLoadBalancersWithResponse call
func ParseListLoadBalancersResponse(rsp *http.Response) (*ListLoadBalancersResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ListLoadBalancersResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			LoadBalancers *[]LoadBalancer `json:"load-balancers,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateLoadBalancerResponse parses an HTTP response from a CreateLoadBalancerWithResponse call
func ParseCreateLoadBalancerResponse(rsp *http.Response) (*CreateLoadBalancerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &CreateLoadBalancerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteLoadBalancerResponse parses an HTTP response from a DeleteLoadBalancerWithResponse call
func ParseDeleteLoadBalancerResponse(rsp *http.Response) (*DeleteLoadBalancerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &DeleteLoadBalancerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetLoadBalancerResponse parses an HTTP response from a GetLoadBalancerWithResponse call
func ParseGetLoadBalancerResponse(rsp *http.Response) (*GetLoadBalancerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetLoadBalancerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest LoadBalancer
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateLoadBalancerResponse parses an HTTP response from a UpdateLoadBalancerWithResponse call
func ParseUpdateLoadBalancerResponse(rsp *http.Response) (*UpdateLoadBalancerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &UpdateLoadBalancerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseAddServiceToLoadBalancerResponse parses an HTTP response from a AddServiceToLoadBalancerWithResponse call
func ParseAddServiceToLoadBalancerResponse(rsp *http.Response) (*AddServiceToLoadBalancerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &AddServiceToLoadBalancerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteLoadBalancerServiceResponse parses an HTTP response from a DeleteLoadBalancerServiceWithResponse call
func ParseDeleteLoadBalancerServiceResponse(rsp *http.Response) (*DeleteLoadBalancerServiceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &DeleteLoadBalancerServiceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetLoadBalancerServiceResponse parses an HTTP response from a GetLoadBalancerServiceWithResponse call
func ParseGetLoadBalancerServiceResponse(rsp *http.Response) (*GetLoadBalancerServiceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetLoadBalancerServiceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest LoadBalancerService
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateLoadBalancerServiceResponse parses an HTTP response from a UpdateLoadBalancerServiceWithResponse call
func ParseUpdateLoadBalancerServiceResponse(rsp *http.Response) (*UpdateLoadBalancerServiceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &UpdateLoadBalancerServiceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetOperationResponse parses an HTTP response from a GetOperationWithResponse call
func ParseGetOperationResponse(rsp *http.Response) (*GetOperationResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetOperationResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePingResponse parses an HTTP response from a PingWithResponse call
func ParsePingResponse(rsp *http.Response) (*PingResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &PingResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListSecurityGroupsResponse parses an HTTP response from a ListSecurityGroupsWithResponse call
func ParseListSecurityGroupsResponse(rsp *http.Response) (*ListSecurityGroupsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ListSecurityGroupsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateSecurityGroupResponse parses an HTTP response from a CreateSecurityGroupWithResponse call
func ParseCreateSecurityGroupResponse(rsp *http.Response) (*CreateSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &CreateSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteSecurityGroupResponse parses an HTTP response from a DeleteSecurityGroupWithResponse call
func ParseDeleteSecurityGroupResponse(rsp *http.Response) (*DeleteSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &DeleteSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetSecurityGroupResponse parses an HTTP response from a GetSecurityGroupWithResponse call
func ParseGetSecurityGroupResponse(rsp *http.Response) (*GetSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest SecurityGroup
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseAddRuleToSecurityGroupResponse parses an HTTP response from a AddRuleToSecurityGroupWithResponse call
func ParseAddRuleToSecurityGroupResponse(rsp *http.Response) (*AddRuleToSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &AddRuleToSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteRuleFromSecurityGroupResponse parses an HTTP response from a DeleteRuleFromSecurityGroupWithResponse call
func ParseDeleteRuleFromSecurityGroupResponse(rsp *http.Response) (*DeleteRuleFromSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &DeleteRuleFromSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListSnapshotsResponse parses an HTTP response from a ListSnapshotsWithResponse call
func ParseListSnapshotsResponse(rsp *http.Response) (*ListSnapshotsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ListSnapshotsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Snapshots *[]Snapshot `json:"snapshots,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteSnapshotResponse parses an HTTP response from a DeleteSnapshotWithResponse call
func ParseDeleteSnapshotResponse(rsp *http.Response) (*DeleteSnapshotResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &DeleteSnapshotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetSnapshotResponse parses an HTTP response from a GetSnapshotWithResponse call
func ParseGetSnapshotResponse(rsp *http.Response) (*GetSnapshotResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetSnapshotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Snapshot
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetExportSnapshotResponse parses an HTTP response from a GetExportSnapshotWithResponse call
func ParseGetExportSnapshotResponse(rsp *http.Response) (*GetExportSnapshotResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetExportSnapshotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest SnapshotExport
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseExportSnapshotResponse parses an HTTP response from a ExportSnapshotWithResponse call
func ParseExportSnapshotResponse(rsp *http.Response) (*ExportSnapshotResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ExportSnapshotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetTemplateResponse parses an HTTP response from a GetTemplateWithResponse call
func ParseGetTemplateResponse(rsp *http.Response) (*GetTemplateResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &GetTemplateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Template
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseVersionResponse parses an HTTP response from a VersionWithResponse call
func ParseVersionResponse(rsp *http.Response) (*VersionResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &VersionResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListZonesResponse parses an HTTP response from a ListZonesWithResponse call
func ParseListZonesResponse(rsp *http.Response) (*ListZonesResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &ListZonesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Zones *[]Zone `json:"zones,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}
