// This file is part of gobizfly

package gobizfly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	defaultAuthType         = "token"
	version                 = "0.0.1"
	ua                      = "bizfly-client-go/" + version
	defaultAPIURL           = "https://manage.bizflycloud.vn/api"
	mediaType               = "application/json; charset=utf-8"
	accountName             = "account"
	loadBalancerServiceName = "load_balancer"
	serverServiceName       = "cloud_server"
	autoScalingServiceName  = "auto_scaling"
	cloudwatcherServiceName = "alert"
	authServiceName         = "auth"
	kubernetesServiceName   = "kubernetes_engine"
	containerRegistryName   = "container_registry"
	cdnName                 = "cdn"
	dnsName                 = "dns"
)

var (
	// ErrNotFound for resource not found status
	ErrNotFound = errors.New("Resource not found")
	// ErrPermissionDenied for permission denied
	ErrPermissionDenied = errors.New("You are not allowed to do this action")
	// ErrCommon for common error
	ErrCommon = errors.New("Error")
)

// Client represents BizFly API client.
type Client struct {
	AutoScaling       AutoScalingService
	CloudWatcher      CloudWatcherService
	Token             TokenService
	LoadBalancer      LoadBalancerService
	Listener          ListenerService
	Pool              PoolService
	Member            MemberService
	HealthMonitor     HealthMonitorService
	KubernetesEngine  KubernetesEngineService
	ContainerRegistry ContainerRegistryService
	CDN               CDNService
	DNS               DNSService
	Account           AccountService
	VPC               VPCService

	Snapshot SnapshotService

	Volume VolumeService
	Server ServerService

	Service  ServiceInterface
	Firewall FirewallService
	SSHKey   SSHKeyService

	httpClient    *http.Client
	apiURL        *url.URL
	userAgent     string
	keystoneToken string
	authMethod    string
	authType      string
	username      string
	password      string
	projectName   string
	appCredID     string
	appCredSecret string
	// TODO: this will be removed in near future
	tenantName string
	tenantID   string
	regionName string
	services   []*Service
}

// Option set Client specific attributes
type Option func(c *Client) error

// WithAPIUrl sets the API url option for BizFly client.
func WithAPIUrl(u string) Option {
	return func(c *Client) error {
		var err error
		c.apiURL, err = url.Parse(u)
		return err
	}
}

// WithHTTPClient sets the client option for BizFly client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			return errors.New("client is nil")
		}

		c.httpClient = client

		return nil
	}
}

// WithRegionName sets the client region for BizFly client.
func WithRegionName(region string) Option {
	return func(c *Client) error {
		c.regionName = region
		return nil
	}
}

// WithTenantName sets the tenant name option for BizFly client.
//
// Deprecated: X-Tenant-Name header required will be removed in API server.
func WithTenantName(tenant string) Option {
	return func(c *Client) error {
		c.tenantName = tenant
		return nil
	}
}

// WithTenantID sets the tenant id name option for BizFly client
//
// Deprecated: X-Tenant-Id header required will be removed in API server.
func WithTenantID(id string) Option {
	return func(c *Client) error {
		c.tenantID = id
		return nil
	}
}

// NewClient creates new BizFly client.
func NewClient(options ...Option) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
		userAgent:  ua,
	}

	err := WithAPIUrl(defaultAPIURL)(c)
	if err != nil {
		return nil, err
	}

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}

	c.AutoScaling = &autoscalingService{client: c}
	c.CloudWatcher = &cloudwatcherService{client: c}
	c.Snapshot = &snapshot{client: c}
	c.Token = &token{client: c}
	c.LoadBalancer = &loadbalancer{client: c}
	c.Listener = &listener{client: c}
	c.Pool = &pool{client: c}
	c.HealthMonitor = &healthmonitor{client: c}
	c.Member = &member{client: c}
	c.Volume = &volume{client: c}
	c.Server = &server{client: c}
	c.Service = &service{client: c}
	c.Firewall = &firewall{client: c}
	c.SSHKey = &sshkey{client: c}
	c.KubernetesEngine = &kubernetesEngineService{client: c}
	c.ContainerRegistry = &containerRegistry{client: c}
	c.CDN = &cdnService{client: c}
	c.DNS = &dnsService{client: c}
	c.Account = &accountService{client: c}
	c.VPC = &vpcService{client: c}
	return c, nil
}

func (c *Client) GetServiceUrl(serviceName string) string {
	for _, service := range c.services {
		if service.CanonicalName == serviceName && service.Region == c.regionName {
			return service.ServiceUrl
		}
	}
	return defaultAPIURL
}

// NewRequest creates an API request.
func (c *Client) NewRequest(ctx context.Context, method, serviceName string, urlStr string, body interface{}) (*http.Request, error) {
	serviceUrl := c.GetServiceUrl(serviceName)
	url := serviceUrl + urlStr

	buf := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", c.userAgent)
	req.Header.Add("X-Tenant-Name", c.tenantName)
	req.Header.Add("X-Tenant-Id", c.tenantID)

	if c.authType == "" {
		c.authType = defaultAuthType
	}

	if c.keystoneToken != "" {
		req.Header.Add("X-Auth-Token", c.keystoneToken)
	}

	req.Header.Add("X-Auth-Type", c.authType)
	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

// Do sends API request.
func (c *Client) Do(ctx context.Context, req *http.Request) (resp *http.Response, err error) {

	resp, err = c.do(ctx, req)
	if err != nil {
		return
	}

	// If 401, get new token and retry one time.
	if resp.StatusCode == http.StatusUnauthorized {
		tok, tokErr := c.Token.Refresh(ctx)
		if tokErr != nil {
			buf, _ := ioutil.ReadAll(resp.Body)
			err = fmt.Errorf("%s : %w", string(buf), tokErr)
			return
		}
		c.SetKeystoneToken(tok.KeystoneToken)
		req.Header.Set("X-Auth-Token", c.keystoneToken)
		resp, err = c.do(ctx, req)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()
		buf, _ := ioutil.ReadAll(resp.Body)
		err = errorFromStatus(resp.StatusCode, string(buf))

	}
	return
}

// SetKeystoneToken sets keystone token value, which will be used for authentication.
func (c *Client) SetKeystoneToken(s string) {
	c.keystoneToken = s
}

// ListOptions specifies the optional parameters for List method.
type ListOptions struct {
	Page  int `json:"page,omitempty"`
	Limit int `json:"limit,omitempty"`
}

func errorFromStatus(code int, msg string) error {
	switch code {
	case http.StatusNotFound:
		return fmt.Errorf("%s: %w", msg, ErrNotFound)
	case http.StatusForbidden:
		return fmt.Errorf("%s: %w", msg, ErrPermissionDenied)
	default:
		return fmt.Errorf("%s: %w", msg, ErrCommon)
	}
}
