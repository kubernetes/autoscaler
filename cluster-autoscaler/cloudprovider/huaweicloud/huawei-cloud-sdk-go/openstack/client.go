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

package openstack

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/auth"
	akskAuth "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/auth/aksk"
	tokenAuth "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/auth/token"
	tokens2 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/openstack/identity/v2/tokens"
	tokens3 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/openstack/identity/v3/tokens"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/openstack/utils"

	"encoding/json"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/openstack/identity/v3/endpoints"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/openstack/identity/v3/services"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/pagination"
)

const (
	// v2 represents Keystone v2.
	// It should never increase beyond 2.0.
	v2 = "v2.0"

	// v3 represents Keystone v3.
	// The version can be anything from v3 to v3.x.
	v3 = "v3"
)

// MyRoundTripper Rewrite RoundTrip to achieve reauth limit 3 times
type MyRoundTripper struct {
	// http.RoundTripper interface.
	rt http.RoundTripper

	// numReauthAttempts, http client Reauth times.
	numReauthAttempts int
}

//Initialize httpclient according to the config parameter.
func newHTTPClient(conf *huaweicloudsdk.Config) http.Client {

	hc := new(http.Client)

	if conf.Timeout > 0 {
		hc.Timeout = conf.Timeout
	}

	if conf.HttpTransport != nil {
		hc.Transport = conf.HttpTransport
	} else {
		hc.Transport = &MyRoundTripper{
			rt: http.DefaultTransport,
		}
	}

	return *hc

}

//RoundTrip Implement the RoundTrip interface function.The reauth default setting is three times.
func (mrt *MyRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := mrt.rt.RoundTrip(request)
	if response == nil {
		return nil, err
	}

	if response.StatusCode == http.StatusUnauthorized {
		if mrt.numReauthAttempts == 3 {
			return response, fmt.Errorf("Tried to re-authenticate 3 times with no success.")
		}
		mrt.numReauthAttempts++
	}

	return response, err
}

/*
AuthenticatedClientWithOptions Initialize the provider client based on the incoming config configuration，and returns
a Provider Client instance that's ready to request SDK service API.

Example of Creating a Service Client with options

	conf := gophercloud.NewConfig()
	ao, err := openstack.AuthOptionsFromEnv()
	provider, err := openstack.AuthenticatedClientWithOptions(ao,conf)
	client, err := openstack.NewNetworkV2(client, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
*/
func AuthenticatedClientWithOptions(options auth.AuthOptionsProvider, conf *huaweicloudsdk.Config) (*huaweicloudsdk.ProviderClient, error) {
	client, err := NewClient(options.GetIdentityEndpoint(), options.GetDomainId(), options.GetProjectId(), conf)
	if err != nil {
		return nil, err
	}

	err = Authenticate(client, options)
	if err != nil {
		return nil, err
	}
	return client, nil
}

/*
AuthenticatedClient logs in to an OpenStack cloud found at the identity endpoint
specified by the options, acquires a token, and returns a Provider Client
instance that's ready to operate.

If the full path to a versioned identity endpoint was specified  (example:
http://example.com:5000/v3), that path will be used as the endpoint to query.

If a versionless endpoint was specified (example: http://example.com:5000/),
the endpoint will be queried to determine which versions of the identity service
are available, then chooses the most recent or most supported version.

Example:

	ao, err := openstack.AuthOptionsFromEnv()
	provider, err := openstack.AuthenticatedClient(ao)
	client, err := openstack.NewNetworkV2(client, EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
*/
func AuthenticatedClient(options auth.AuthOptionsProvider) (*huaweicloudsdk.ProviderClient, error) {
	client, err := NewClient(options.GetIdentityEndpoint(), options.GetDomainId(), options.GetProjectId(), huaweicloudsdk.NewConfig())
	if err != nil {
		return nil, err
	}

	err = Authenticate(client, options)
	if err != nil {
		return nil, err
	}
	return client, nil
}

/*
NewClient prepares an unauthenticated ProviderClient instance.
Most users will probably prefer using the AuthenticatedClient function
instead.

This is useful if you wish to explicitly control the version of the identity
service that's used for authentication explicitly, for example.

A basic example of using this would be:

	ao, err := openstack.AuthOptionsFromEnv()
	provider, err := openstack.NewClient(ao.IdentityEndpoint)
	client, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
*/
func NewClient(endpoint, domainID, projectID string, conf *huaweicloudsdk.Config) (*huaweicloudsdk.ProviderClient, error) {
	if endpoint == "" {
		message := fmt.Sprintf(huaweicloudsdk.CEMissingInputMessage, "IdentityEndpoint")
		err := huaweicloudsdk.NewSystemCommonError(huaweicloudsdk.CEMissingInputCode, message)
		return nil, err
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	//if domainID == "" {
	//	message := fmt.Sprintf(gophercloud.CEMissingInputMessage, "domainID")
	//	err := gophercloud.NewSystemCommonError(gophercloud.CEMissingInputCode, message)
	//	return nil, err
	//}
	//
	//if projectID == "" {
	//	message := fmt.Sprintf(gophercloud.CEMissingInputMessage, "projectID")
	//	err := gophercloud.NewSystemCommonError(gophercloud.CEMissingInputCode, message)
	//	return nil, err
	//}

	u.RawQuery, u.Fragment = "", ""

	var base string
	versionRe := regexp.MustCompile("v[0-9.]+/?")
	if version := versionRe.FindString(u.Path); version != "" {
		base = strings.Replace(u.String(), version, "", -1)
	} else {
		base = u.String()
	}

	endpoint = huaweicloudsdk.NormalizeURL(endpoint)
	base = huaweicloudsdk.NormalizeURL(base)

	p := new(huaweicloudsdk.ProviderClient)
	p.IdentityBase = base
	p.IdentityEndpoint = endpoint
	p.DomainID = domainID
	p.ProjectID = projectID
	p.Conf = conf
	p.UseTokenLock()
	p.HTTPClient = newHTTPClient(conf) //自定义httpclient，限制reauth为3次

	return p, nil
}

// Authenticate or re-authenticate against the most recent identity service
// supported at the provided endpoint.
func Authenticate(client *huaweicloudsdk.ProviderClient, options auth.AuthOptionsProvider) error {
	versions := []*utils.Version{
		{ID: v2, Priority: 20, Suffix: "/v2.0/"},
		{ID: v3, Priority: 30, Suffix: "/v3/"},
	}

	chosen, endpoint, err := utils.ChooseVersion(client, versions)
	if err != nil {
		return err
	}

	authOptions, isTokenAuthOptions := options.(tokenAuth.TokenOptions)

	if isTokenAuthOptions {
		switch chosen.ID {
		case v2:
			return tokenAuthV2(client, endpoint, authOptions, huaweicloudsdk.EndpointOpts{})
		case v3:
			return tokenAuthV3(client, endpoint, &authOptions, huaweicloudsdk.EndpointOpts{})
		default:
			// The switch statement must be out of date from the versions list.
			return fmt.Errorf("Unrecognized identity version: %s", chosen.ID)
		}
	} else {
		akskOptions, isAKSKOptions := options.(akskAuth.AKSKOptions)

		if isAKSKOptions {
			return akskAuthV3(client, endpoint, akskOptions, huaweicloudsdk.EndpointOpts{})
		}
		TokenIdOptions, isTokenIdOptions := options.(tokenAuth.TokenIdOptions)

		if isTokenIdOptions {
			return tokenIDAuthV3(client, endpoint, TokenIdOptions, huaweicloudsdk.EndpointOpts{})
		}
		return fmt.Errorf("Unrecognized auth options provider: %s", reflect.TypeOf(options))
	}

}

func getEntryByServiceId(entries []tokens3.CatalogEntry, serviceId string) *tokens3.CatalogEntry {
	if entries == nil {
		return nil
	}

	for idx := range entries {
		if entries[idx].ID == serviceId {
			return &entries[idx]
		}
	}

	return nil
}

func tokenIDAuthV3(client *huaweicloudsdk.ProviderClient, endpoint string, tokenIdOptions tokenAuth.TokenIdOptions, eo huaweicloudsdk.EndpointOpts) error {
	// Override the generated service endpoint with the one returned by the version endpoint.

	client.TokenID = tokenIdOptions.AuthToken

	v3Client, err := NewIdentityV3(client, eo)
	if err != nil {
		return err
	}

	if endpoint != "" {
		v3Client.Endpoint = endpoint
	}

	var entries = make([]tokens3.CatalogEntry, 0, 1)
	serviceListErr := services.List(v3Client, services.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		serviceLst, err := services.ExtractServices(page)
		if err != nil {
			return false, err
		}

		for _, svc := range serviceLst {
			entry := tokens3.CatalogEntry{
				Type: svc.Type,
				Name: svc.Name,
				ID:   svc.ID,
			}
			entries = append(entries, entry)
		}

		return true, nil
	})

	if serviceListErr != nil {
		return serviceListErr
	}

	endpointListErr := endpoints.List(v3Client, endpoints.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		endpointList, err := endpoints.ExtractEndpoints(page)
		if err != nil {
			return false, err
		}

		for _, endpoint := range endpointList {
			entry := getEntryByServiceId(entries, endpoint.ServiceID)

			if entry != nil {
				entry.Endpoints = append(entry.Endpoints, tokens3.Endpoint{
					URL:       strings.Replace(endpoint.URL, "$(tenant_id)s", tokenIdOptions.ProjectID, -1),
					Region:    endpoint.Region,
					Interface: string(endpoint.Availability),
					ID:        endpoint.ID,
				})
			}
		}

		client.EndpointLocator = func(opts huaweicloudsdk.EndpointOpts) (string, error) {
			return V3TokenIdExtractEndpointURL(&tokens3.ServiceCatalog{
				Entries: entries,
			}, opts, tokenIdOptions)
		}

		return true, nil
	})

	if endpointListErr != nil {
		return endpointListErr
	}

	if client.EndpointLocator == nil {
		return huaweicloudsdk.NewSystemCommonError(huaweicloudsdk.CENoEndPointInCatalogCode, huaweicloudsdk.CENoEndPointInCatalogMessage)
	}
	return nil
}

func akskAuthV3(client *huaweicloudsdk.ProviderClient, endpoint string, options akskAuth.AKSKOptions, eo huaweicloudsdk.EndpointOpts) error {
	v3Client, err := NewIdentityV3(client, eo)
	if err != nil {
		return err
	}

	if endpoint != "" {
		v3Client.Endpoint = endpoint
	}

	v3Client.AKSKOptions = options

	var entries = make([]tokens3.CatalogEntry, 0, 1)
	serviceListErr := services.List(v3Client, services.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		serviceLst, err := services.ExtractServices(page)
		if err != nil {
			return false, err
		}

		for _, svc := range serviceLst {
			entry := tokens3.CatalogEntry{
				Type: svc.Type,
				Name: svc.Name,
				ID:   svc.ID,
			}
			entries = append(entries, entry)
		}

		return true, nil
	})

	if serviceListErr != nil {
		return serviceListErr
	}

	endpointListErr := endpoints.List(v3Client, endpoints.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		endpointList, err := endpoints.ExtractEndpoints(page)
		if err != nil {
			return false, err
		}

		for _, endpoint := range endpointList {
			entry := getEntryByServiceId(entries, endpoint.ServiceID)

			if entry != nil {
				entry.Endpoints = append(entry.Endpoints, tokens3.Endpoint{
					URL:       strings.Replace(endpoint.URL, "$(tenant_id)s", options.ProjectID, -1),
					Region:    endpoint.Region,
					Interface: string(endpoint.Availability),
					ID:        endpoint.ID,
				})
			}
		}

		client.EndpointLocator = func(opts huaweicloudsdk.EndpointOpts) (string, error) {
			return GetEndpointURLForAKSKAuth(&tokens3.ServiceCatalog{
				Entries: entries,
			}, opts, options)
		}

		return true, nil
	})

	if endpointListErr != nil {
		return endpointListErr
	}

	if client.EndpointLocator == nil {
		return huaweicloudsdk.NewSystemCommonError(huaweicloudsdk.CENoEndPointInCatalogCode, huaweicloudsdk.CENoEndPointInCatalogMessage)
	}
	return nil

}

// AuthenticateV2 explicitly authenticates against the identity v2 endpoint.
func AuthenticateV2(client *huaweicloudsdk.ProviderClient, options tokenAuth.TokenOptions, eo huaweicloudsdk.EndpointOpts) error {
	return tokenAuthV2(client, "", options, eo)
}

func tokenAuthV2(client *huaweicloudsdk.ProviderClient, endpoint string, options tokenAuth.TokenOptions, eo huaweicloudsdk.EndpointOpts) error {
	v2Client, err := NewIdentityV2(client, eo)
	if err != nil {
		return err
	}

	if endpoint != "" {
		v2Client.Endpoint = endpoint
	}

	v2Opts := tokens2.AuthOptions{
		IdentityEndpoint: options.IdentityEndpoint,
		Username:         options.Username,
		Password:         options.Password,
		TenantID:         options.TenantID,
		TenantName:       options.TenantName,
		AllowReauth:      options.AllowReauth,
		TokenID:          options.TokenID,
	}

	result := tokens2.Create(v2Client, v2Opts)

	token, err := result.ExtractToken()
	if err != nil {
		return err
	}

	catalog, err := result.ExtractServiceCatalog()
	if err != nil {
		return err
	}

	if options.AllowReauth {
		// here we're creating a throw-away client (tac). it's a copy of the user's provider client, but
		// with the token and reauth func zeroed out. combined with setting `AllowReauth` to `false`,
		// this should retry authentication only once
		tac := *client
		tac.ReauthFunc = nil
		tac.TokenID = ""
		tao := options
		tao.AllowReauth = false
		client.ReauthFunc = func() error {
			err := tokenAuthV2(&tac, endpoint, tao, eo)
			if err != nil {
				return err
			}
			client.TokenID = tac.TokenID
			return nil
		}
	}
	client.TokenID = token.ID
	client.EndpointLocator = func(opts huaweicloudsdk.EndpointOpts) (string, error) {
		return V2EndpointURL(catalog, opts)
	}

	return nil
}

// AuthenticateV3 explicitly authenticates against the identity v3 service.
func AuthenticateV3(client *huaweicloudsdk.ProviderClient, options tokens3.AuthOptionsBuilder, eo huaweicloudsdk.EndpointOpts) error {
	return tokenAuthV3(client, "", options, eo)
}

func tokenAuthV3(client *huaweicloudsdk.ProviderClient, endpoint string, opts tokens3.AuthOptionsBuilder, eo huaweicloudsdk.EndpointOpts) error {
	// Override the generated service endpoint with the one returned by the version endpoint.
	v3Client, err := NewIdentityV3(client, eo)
	if err != nil {
		return err
	}

	if endpoint != "" {
		v3Client.Endpoint = endpoint
	}

	result := tokens3.Create(v3Client, opts)

	token, err := result.ExtractToken()
	if err != nil {
		return err
	}

	catalog, err := result.ExtractServiceCatalog()
	if err != nil {
		return err
	}

	client.TokenID = token.ID

	if opts.CanReauth() {
		// here we're creating a throw-away client (tac). it's a copy of the user's provider client, but
		// with the token and reauth func zeroed out. combined with setting `AllowReauth` to `false`,
		// this should retry authentication only once
		tac := *client
		tac.ReauthFunc = nil
		tac.TokenID = ""
		var tao tokens3.AuthOptionsBuilder
		switch ot := opts.(type) {
		case *tokenAuth.TokenOptions:
			o := *ot
			o.AllowReauth = false
			tao = &o
		case *tokens3.TokenOptions:
			o := *ot
			o.AllowReauth = false
			tao = &o
		default:
			tao = opts
		}
		client.ReauthFunc = func() error {
			err := tokenAuthV3(&tac, endpoint, tao, eo)
			if err != nil {
				return err
			}
			client.TokenID = tac.TokenID
			return nil
		}
	}
	client.EndpointLocator = func(endpointOpts huaweicloudsdk.EndpointOpts) (string, error) {
		return V3ExtractEndpointURL(catalog, endpointOpts, opts)
	}

	return nil
}

// NewIdentityV2 creates a ServiceClient that may be used to interact with the
// v2 identity service.
func NewIdentityV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	endpoint := client.IdentityBase + "v2.0/"
	clientType := "identity"
	var err error
	if !reflect.DeepEqual(eo, huaweicloudsdk.EndpointOpts{}) {
		eo.ApplyDefaults(clientType)
		endpoint, err = client.EndpointLocator(eo)
		if err != nil {
			return nil, err
		}
	}

	return &huaweicloudsdk.ServiceClient{
		ProviderClient: client,
		Endpoint:       endpoint,
		Type:           clientType,
	}, nil
}

// NewIdentityV3 creates a ServiceClient that may be used to access the v3
// identity service.
func NewIdentityV3(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	endpoint := client.IdentityBase + "v3/"
	clientType := "identity"
	var err error
	if !reflect.DeepEqual(eo, huaweicloudsdk.EndpointOpts{}) {
		eo.ApplyDefaults(clientType)
		endpoint, err = client.EndpointLocator(eo)
		if err != nil {
			return nil, err
		}
	}

	// Ensure endpoint still has a suffix of v3.
	// This is because EndpointLocator might have found a versionless
	// endpoint and requests will fail unless targeted at /v3.
	if !strings.HasSuffix(endpoint, "v3/") {
		endpoint = endpoint + "v3/"
	}

	return &huaweicloudsdk.ServiceClient{
		ProviderClient: client,
		Endpoint:       endpoint,
		Type:           clientType,
	}, nil
}

func initClientOpts(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts, clientType string) (*huaweicloudsdk.ServiceClient, error) {
	sc := new(huaweicloudsdk.ServiceClient)
	eo.ApplyDefaults(clientType)
	url, err := client.EndpointLocator(eo)
	if err != nil {
		return sc, err
	}
	sc.ProviderClient = client
	sc.Endpoint = url
	sc.Type = clientType
	return sc, nil
}

// NewObjectStorageV1 creates a ServiceClient that may be used with the v1
// object storage package.
func NewObjectStorageV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "object-store")
}

// NewComputeV2 creates a ServiceClient that may be used with the v2 compute
// package.
func NewComputeV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "compute")
}

// NewNetworkV2 creates a ServiceClient that may be used with the v2 network
// package.
func NewNetworkV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "network")
	sc.ResourceBase = sc.Endpoint + "v2.0/"
	return sc, err
}

// NewBlockStorageV1 creates a ServiceClient that may be used to access the v1
// block storage service.
func NewBlockStorageV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "volume")
}

// NewBlockStorageV2 creates a ServiceClient that may be used to access the v2
// block storage service.
func NewBlockStorageV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "volumev2")
}

// NewBlockStorageV3 creates a ServiceClient that may be used to access the v3 block storage service.
func NewBlockStorageV3(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "volumev3")
}

// NewSharedFileSystemV2 creates a ServiceClient that may be used to access the v2 shared file system service.
func NewSharedFileSystemV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "sharev2")
}

// NewCDNV1 creates a ServiceClient that may be used to access the OpenStack v1
// CDN service.
func NewCDNV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "cdn")
}

// NewOrchestrationV1 creates a ServiceClient that may be used to access the v1
// orchestration service.
func NewOrchestrationV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "orchestration")
}

// NewDBV1 creates a ServiceClient that may be used to access the v1 DB service.
func NewDBV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "database")
}

// NewDNSV2 creates a ServiceClient that may be used to access the v2 DNS
// service.
func NewDNSV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "dns")
	sc.ResourceBase = sc.Endpoint + "v2/"
	return sc, err
}

// NewImageServiceV2 creates a ServiceClient that may be used to access the v2
// image service.
func NewImageServiceV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "image")
	sc.ResourceBase = sc.Endpoint + "v2/"
	return sc, err
}

// NewLoadBalancerV2 creates a ServiceClient that may be used to access the v2
// load balancer service.
func NewLoadBalancerV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "load-balancer")
	sc.ResourceBase = sc.Endpoint + "v2.0/"
	return sc, err
}

// NewECSV1 creates a ServiceClient that may be used to access the v1
// ecs service.
func NewECSV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "ecs")
}

// NewECSV1_1 creates a ServiceClient that may be used to access the v1.1
// ecs service.
func NewECSV1_1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "ecsv1.1")
}

// NewECSV2 creates a ServiceClient that may be used to access the v2
// ecs service.
func NewECSV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "ecsv2")
}

// NewIMSV1 ...
func NewIMSV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "image")
	sc.ResourceBase = sc.Endpoint + "v1/"
	return sc, err
}

// NewIMSV2 creates a ServiceClient that may be used to access the v2
// image service.
func NewIMSV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "image")
	sc.ResourceBase = sc.Endpoint + "v2/"
	return sc, err
}

// NewBSSV1 creates a ServiceClient that may be used to access the v1.0
// BSS service.
func NewBSSV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "bssv1")
	return sc, err
}

// NewBSSIntlV1 creates a ServiceClient that may be used to access the v1.0
// BSS-INTLV1 service.
func NewBSSIntlV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "bss-intlv1")
	return sc, err
}

// NewVPCV1 creates a ServiceClient that may be used with the v1 network
// package.
func NewVPCV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "vpc")
	return sc, err
}

// NewCESV1 creates a ServiceClient that may be used with the v1 cloud eye service
// package.
func NewCESV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	type details struct {
		Details string `json:"details"`
		Code    string `json:"code"`
	}
	type CESError struct {
		Message string  `json:"message"`
		Code    int     `json:"code"`
		Details details `json:"details"`
		Element string  `json:"element"`
	}

	sc, err := initClientOpts(client, eo, "cesv1")
	sc.HandleError = func(httpStatus int, responseContent string) error {
		var cesErr CESError
		var code string
		message := responseContent
		marshalErr := json.Unmarshal([]byte(responseContent), &cesErr)

		if marshalErr == nil && cesErr.Details.Code != "" {
			code = cesErr.Details.Code
			message = cesErr.Details.Details
		} else {
			code = huaweicloudsdk.MatchErrorCode(httpStatus, message)
		}
		return &huaweicloudsdk.UnifiedError{
			ErrCode:    code,
			ErrMessage: message,
		}

	}
	return sc, err
}

// NewVPCV2 creates a ServiceClient that may be used with the v2.0 vpc
// package.
func NewVPCV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "vpcv2.0")
	return sc, err
}

// NewASV1 creates a ServiceClient that may be used with the v1 as
// package.
func NewASV1(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "asv1")
	return sc, err
}

// NewASV2 creates a ServiceClient that may be used with the v2 as
// package.
func NewASV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "asv2")
	return sc, err
}

// NewFGSV2 creates a ServiceClient that may be used with the v2 as
// package.
func NewFGSV2(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "fgsv2")
	return sc, err
}

// NewRDSV3 creates a ServiceClient that may be used with the v3 rds
// package.
func NewRDSV3(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "rdsv3")
	return sc, err
}

// NewCCEV3 creates a ServiceClient that may be used with the v3 CCE
// package.
func NewCCEV3(client *huaweicloudsdk.ProviderClient, eo huaweicloudsdk.EndpointOpts) (*huaweicloudsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "ccev3")
	return sc, err
}
