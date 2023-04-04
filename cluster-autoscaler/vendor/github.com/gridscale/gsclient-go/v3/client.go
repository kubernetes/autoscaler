package gsclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"time"
)

const (
	requestBase                   = "/requests/"
	apiServerBase                 = "/objects/servers"
	apiStorageBase                = "/objects/storages"
	apiNetworkBase                = "/objects/networks"
	apiIPBase                     = "/objects/ips"
	apiSshkeyBase                 = "/objects/sshkeys"
	apiTemplateBase               = "/objects/templates"
	apiLoadBalancerBase           = "/objects/loadbalancers"
	apiPaaSBase                   = "/objects/paas"
	apiISOBase                    = "/objects/isoimages"
	apiObjectStorageBase          = "/objects/objectstorages"
	apiFirewallBase               = "/objects/firewalls"
	apiMarketplaceApplicationBase = "/objects/marketplace/applications"
	apiLocationBase               = "/objects/locations"
	apiEventBase                  = "/objects/events"
	apiLabelBase                  = "/objects/labels"
	apiDeletedBase                = "/objects/deleted"
	apiSSLCertificateBase         = "/objects/certificates"
	apiProjectLevelUsage          = "/projects/usage"
	apiContractLevelUsage         = "/contracts/usage"
	apiBackupLocationBase         = "/objects/backup_locations"
)

// Client struct of a gridscale golang client.
type Client struct {
	cfg *Config
}

// NewClient creates new gridscale golang client.
func NewClient(c *Config) *Client {
	client := &Client{
		cfg: c,
	}
	return client
}

// HttpClient returns http client.
func (c *Client) HttpClient() *http.Client {
	return c.cfg.httpClient
}

// Synchronous returns if the client is sync or not.
func (c *Client) Synchronous() bool {
	return c.cfg.sync
}

// DelayInterval returns request delay interval.
func (c *Client) DelayInterval() time.Duration {
	return c.cfg.delayInterval
}

// MaxNumberOfRetries returns max number of retries.
func (c *Client) MaxNumberOfRetries() int {
	return c.cfg.maxNumberOfRetries
}

// APIURL returns api URL.
func (c *Client) APIURL() string {
	return c.cfg.apiURL
}

// UserAgent returns user agent.
func (c *Client) UserAgent() string {
	return c.cfg.userAgent
}

// UserUUID returns user UUID.
func (c *Client) UserUUID() string {
	return c.cfg.userUUID
}

// APIToken returns api token.
func (c *Client) APIToken() string {
	return c.cfg.apiToken
}

// WithHTTPHeaders adds custom HTTP headers to Client.
func (c *Client) WithHTTPHeaders(headers map[string]string) {
	c.cfg.httpHeaders = headers
}

// waitForRequestCompleted allows to wait for a request to complete.
func (c *Client) waitForRequestCompleted(ctx context.Context, id string) error {
	if !isValidUUID(id) {
		return errors.New("'id' is invalid")
	}
	return retryWithContext(ctx, func() (bool, error) {
		r := gsRequest{
			uri:                 path.Join(requestBase, id),
			method:              "GET",
			skipCheckingRequest: true,
		}
		var response RequestStatus
		err := r.execute(ctx, *c, &response)
		if err != nil {
			return false, err
		}
		if response[id].Status == requestDoneStatus {
			return false, nil
		} else if response[id].Status == requestFailStatus {
			errMessage := fmt.Sprintf("request %s failed with error %s", id, response[id].Message)
			return false, errors.New(errMessage)
		}
		return true, nil
	}, c.DelayInterval())
}
