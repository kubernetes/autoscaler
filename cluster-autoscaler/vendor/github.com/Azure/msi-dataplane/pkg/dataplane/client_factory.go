package dataplane

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/go-logr/logr"

	"github.com/Azure/msi-dataplane/pkg/dataplane/internal/client"
)

const (
	// TODO - Make module name configurable
	moduleName = "managedidentitydataplane.APIClient"
	// TODO - Tie the module version to update automatically with new releases
	moduleVersion = "v0.0.1"
)

// ClientFactory creates clients for managed identity credentials.
type ClientFactory interface {
	// NewClient creates a client that can operate on credentials for one managed identity.
	// identityURL is the x-ms-identity-url header provided from ARM, including any path,
	// query parameters, etc.
	NewClient(identityURL string) (Client, error)
}

type clientOpts struct {
	logger *logr.Logger
}

type ClientFactoryOption func(*clientOpts)

// WithLogger sets a custom logger for the reloadingCredential.
// This can be useful for debugging or logging purposes.
func WithClientLogger(logger *logr.Logger) ClientFactoryOption {
	return func(c *clientOpts) {
		c.logger = logger
	}
}

// NewClientFactory creates a new MSI data plane client factory. The credentials and audience presented
// are for the first-party credential. As the server to be contacted for each identity varies, a factory
// is returned that can create clients on-demand.
func NewClientFactory(cred azcore.TokenCredential, audience string, opts *azcore.ClientOptions, clientFactoryOpts ...ClientFactoryOption) ClientFactory {
	defaultLogger := logr.FromSlogHandler(slog.NewTextHandler(os.Stdout, nil))
	cfOpts := &clientOpts{
		logger: &defaultLogger,
	}
	for _, opt := range clientFactoryOpts {
		opt(cfOpts)
	}
	return &clientFactory{
		cred:       cred,
		audience:   audience,
		cfOpts:     cfOpts,
		clientOpts: opts,
	}
}

type clientFactory struct {
	cred       azcore.TokenCredential
	audience   string
	cfOpts     *clientOpts
	clientOpts *azcore.ClientOptions
}

var _ ClientFactory = (*clientFactory)(nil)

type httpRequestDoerFunc func(*policy.Request) (*http.Response, error)

var _ policy.Policy = (httpRequestDoerFunc)(nil)

func (f httpRequestDoerFunc) Do(req *policy.Request) (*http.Response, error) { return f(req) }

func (c *clientFactory) NewClient(identityURL string) (Client, error) {
	parsedURL, err := url.ParseRequestURI(identityURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing identity URL: %w", err)
	}
	server := url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
		Path:   parsedURL.Path,
	}

	azCoreClient, err := azcore.NewClient(moduleName, moduleVersion, runtime.PipelineOptions{
		PerCall: []policy.Policy{
			httpRequestDoerFunc(func(req *policy.Request) (*http.Response, error) {
				// x-ms-identity-url header from ARM contains query parameters we need to keep
				query := req.Raw().URL.Query()
				for key, values := range parsedURL.Query() {
					for _, value := range values {
						query.Add(key, value)
					}
				}
				req.Raw().URL.RawQuery = query.Encode()
				return req.Next()
			}),
			newAuthenticatorPolicy(c.cred, c.audience),
		},
	}, c.clientOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating azcore client: %w", err)
	}
	return &clientAdapter{
		hostPath: server.String(),
		delegate: client.NewManagedIdentityDataPlaneAPIClient(azCoreClient),
	}, nil
}
