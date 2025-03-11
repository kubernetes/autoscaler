package hcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/http/httpguts"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud/internal/instrumentation"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud/schema"
)

// Endpoint is the base URL of the API.
const Endpoint = "https://api.hetzner.cloud/v1"

// UserAgent is the value for the library part of the User-Agent header
// that is sent with each request.
const UserAgent = "hcloud-go/" + Version

// A BackoffFunc returns the duration to wait before performing the
// next retry. The retries argument specifies how many retries have
// already been performed. When called for the first time, retries is 0.
type BackoffFunc func(retries int) time.Duration

// ConstantBackoff returns a BackoffFunc which backs off for
// constant duration d.
func ConstantBackoff(d time.Duration) BackoffFunc {
	return func(_ int) time.Duration {
		return d
	}
}

// ExponentialBackoff returns a BackoffFunc which implements an exponential
// backoff.
// It uses the formula:
//
//	b^retries * d
func ExponentialBackoff(b float64, d time.Duration) BackoffFunc {
	return func(retries int) time.Duration {
		return time.Duration(math.Pow(b, float64(retries))) * d
	}
}

// Client is a client for the Hetzner Cloud API.
type Client struct {
	endpoint                string
	token                   string
	tokenValid              bool
	backoffFunc             BackoffFunc
	pollBackoffFunc         BackoffFunc
	httpClient              *http.Client
	applicationName         string
	applicationVersion      string
	userAgent               string
	debugWriter             io.Writer
	instrumentationRegistry prometheus.Registerer

	Action           ActionClient
	Certificate      CertificateClient
	Datacenter       DatacenterClient
	Firewall         FirewallClient
	FloatingIP       FloatingIPClient
	Image            ImageClient
	ISO              ISOClient
	LoadBalancer     LoadBalancerClient
	LoadBalancerType LoadBalancerTypeClient
	Location         LocationClient
	Network          NetworkClient
	Pricing          PricingClient
	Server           ServerClient
	ServerType       ServerTypeClient
	SSHKey           SSHKeyClient
	Volume           VolumeClient
	PlacementGroup   PlacementGroupClient
	RDNS             RDNSClient
	PrimaryIP        PrimaryIPClient
}

// A ClientOption is used to configure a Client.
type ClientOption func(*Client)

// WithEndpoint configures a Client to use the specified API endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(client *Client) {
		client.endpoint = strings.TrimRight(endpoint, "/")
	}
}

// WithToken configures a Client to use the specified token for authentication.
func WithToken(token string) ClientOption {
	return func(client *Client) {
		client.token = token
		client.tokenValid = httpguts.ValidHeaderFieldValue(token)
	}
}

// WithPollInterval configures a Client to use the specified interval when
// polling from the API.
//
// Deprecated: Setting the poll interval is deprecated, you can now configure
// [WithPollBackoffFunc] with a [ConstantBackoff] to get the same results. To
// migrate your code, replace your usage like this:
//
//	// before
//	hcloud.WithPollInterval(2 * time.Second)
//	// now
//	hcloud.WithPollBackoffFunc(hcloud.ConstantBackoff(2 * time.Second))
func WithPollInterval(pollInterval time.Duration) ClientOption {
	return WithPollBackoffFunc(ConstantBackoff(pollInterval))
}

// WithPollBackoffFunc configures a Client to use the specified backoff
// function when polling from the API.
func WithPollBackoffFunc(f BackoffFunc) ClientOption {
	return func(client *Client) {
		client.pollBackoffFunc = f
	}
}

// WithBackoffFunc configures a Client to use the specified backoff function.
// The backoff function is used for retrying HTTP requests.
func WithBackoffFunc(f BackoffFunc) ClientOption {
	return func(client *Client) {
		client.backoffFunc = f
	}
}

// WithApplication configures a Client with the given application name and
// application version. The version may be blank. Programs are encouraged
// to at least set an application name.
func WithApplication(name, version string) ClientOption {
	return func(client *Client) {
		client.applicationName = name
		client.applicationVersion = version
	}
}

// WithDebugWriter configures a Client to print debug information to the given
// writer. To, for example, print debug information on stderr, set it to os.Stderr.
func WithDebugWriter(debugWriter io.Writer) ClientOption {
	return func(client *Client) {
		client.debugWriter = debugWriter
	}
}

// WithHTTPClient configures a Client to perform HTTP requests with httpClient.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// WithInstrumentation configures a Client to collect metrics about the performed HTTP requests.
func WithInstrumentation(registry prometheus.Registerer) ClientOption {
	return func(client *Client) {
		client.instrumentationRegistry = registry
	}
}

// NewClient creates a new client.
func NewClient(options ...ClientOption) *Client {
	client := &Client{
		endpoint:        Endpoint,
		tokenValid:      true,
		httpClient:      &http.Client{},
		backoffFunc:     ExponentialBackoff(2, 500*time.Millisecond),
		pollBackoffFunc: ConstantBackoff(500 * time.Millisecond),
	}

	for _, option := range options {
		option(client)
	}

	client.buildUserAgent()
	if client.instrumentationRegistry != nil {
		i := instrumentation.New("api", client.instrumentationRegistry)
		client.httpClient.Transport = i.InstrumentedRoundTripper()
	}

	client.Action = ActionClient{action: &ResourceActionClient{client: client}}
	client.Datacenter = DatacenterClient{client: client}
	client.FloatingIP = FloatingIPClient{client: client, Action: &ResourceActionClient{client: client, resource: "floating_ips"}}
	client.Image = ImageClient{client: client, Action: &ResourceActionClient{client: client, resource: "images"}}
	client.ISO = ISOClient{client: client}
	client.Location = LocationClient{client: client}
	client.Network = NetworkClient{client: client, Action: &ResourceActionClient{client: client, resource: "networks"}}
	client.Pricing = PricingClient{client: client}
	client.Server = ServerClient{client: client, Action: &ResourceActionClient{client: client, resource: "servers"}}
	client.ServerType = ServerTypeClient{client: client}
	client.SSHKey = SSHKeyClient{client: client}
	client.Volume = VolumeClient{client: client, Action: &ResourceActionClient{client: client, resource: "volumes"}}
	client.LoadBalancer = LoadBalancerClient{client: client, Action: &ResourceActionClient{client: client, resource: "load_balancers"}}
	client.LoadBalancerType = LoadBalancerTypeClient{client: client}
	client.Certificate = CertificateClient{client: client, Action: &ResourceActionClient{client: client, resource: "certificates"}}
	client.Firewall = FirewallClient{client: client, Action: &ResourceActionClient{client: client, resource: "firewalls"}}
	client.PlacementGroup = PlacementGroupClient{client: client}
	client.RDNS = RDNSClient{client: client}
	client.PrimaryIP = PrimaryIPClient{client: client, Action: &ResourceActionClient{client: client, resource: "primary_ips"}}

	return client
}

// NewRequest creates an HTTP request against the API. The returned request
// is assigned with ctx and has all necessary headers set (auth, user agent, etc.).
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	url := c.endpoint + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	if !c.tokenValid {
		return nil, errors.New("Authorization token contains invalid characters")
	} else if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req = req.WithContext(ctx)
	return req, nil
}

// Do performs an HTTP request against the API.
// v can be nil, an io.Writer to write the response body to or a pointer to
// a struct to json.Unmarshal the response to.
func (c *Client) Do(r *http.Request, v interface{}) (*Response, error) {
	var retries int
	var body []byte
	var err error
	if r.ContentLength > 0 {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			r.Body.Close()
			return nil, err
		}
		r.Body.Close()
	}
	for {
		if r.ContentLength > 0 {
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		if c.debugWriter != nil {
			dumpReq, err := dumpRequest(r)
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(c.debugWriter, "--- Request:\n%s\n\n", dumpReq)
		}

		resp, err := c.httpClient.Do(r)
		if err != nil {
			return nil, err
		}
		response := &Response{Response: resp}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return response, err
		}
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(body))

		if c.debugWriter != nil {
			dumpResp, err := httputil.DumpResponse(resp, true)
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(c.debugWriter, "--- Response:\n%s\n\n", dumpResp)
		}

		if err = response.readMeta(body); err != nil {
			return response, fmt.Errorf("hcloud: error reading response meta data: %s", err)
		}

		if response.StatusCode >= 400 && response.StatusCode <= 599 {
			err = errorFromResponse(response, body)
			if err == nil {
				err = fmt.Errorf("hcloud: server responded with status code %d", resp.StatusCode)
			} else if IsError(err, ErrorCodeConflict) {
				c.backoff(retries)
				retries++
				continue
			}
			return response, err
		}
		if v != nil {
			if w, ok := v.(io.Writer); ok {
				_, err = io.Copy(w, bytes.NewReader(body))
			} else {
				err = json.Unmarshal(body, v)
			}
		}

		return response, err
	}
}

func (c *Client) backoff(retries int) {
	time.Sleep(c.backoffFunc(retries))
}

func (c *Client) all(f func(int) (*Response, error)) error {
	var (
		page = 1
	)
	for {
		resp, err := f(page)
		if err != nil {
			return err
		}
		if resp.Meta.Pagination == nil || resp.Meta.Pagination.NextPage == 0 {
			return nil
		}
		page = resp.Meta.Pagination.NextPage
	}
}

func (c *Client) buildUserAgent() {
	switch {
	case c.applicationName != "" && c.applicationVersion != "":
		c.userAgent = c.applicationName + "/" + c.applicationVersion + " " + UserAgent
	case c.applicationName != "" && c.applicationVersion == "":
		c.userAgent = c.applicationName + " " + UserAgent
	default:
		c.userAgent = UserAgent
	}
}

func dumpRequest(r *http.Request) ([]byte, error) {
	// Duplicate the request, so we can redact the auth header
	rDuplicate := r.Clone(context.Background())
	rDuplicate.Header.Set("Authorization", "REDACTED")

	// To get the request body we need to read it before the request was actually sent.
	// See https://github.com/golang/go/issues/29792
	dumpReq, err := httputil.DumpRequestOut(rDuplicate, true)
	if err != nil {
		return nil, err
	}

	// Set original request body to the duplicate created by DumpRequestOut. The request body is not duplicated
	// by .Clone() and instead just referenced, so it would be completely read otherwise.
	r.Body = rDuplicate.Body

	return dumpReq, nil
}

func errorFromResponse(resp *Response, body []byte) error {
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		return nil
	}

	var respBody schema.ErrorResponse
	if err := json.Unmarshal(body, &respBody); err != nil {
		return nil
	}
	if respBody.Error.Code == "" && respBody.Error.Message == "" {
		return nil
	}

	hcErr := ErrorFromSchema(respBody.Error)
	hcErr.response = resp
	return hcErr
}

const (
	headerCorrelationID = "X-Correlation-Id"
)

// Response represents a response from the API. It embeds http.Response.
type Response struct {
	*http.Response
	Meta Meta
}

func (r *Response) readMeta(body []byte) error {
	if h := r.Header.Get("RateLimit-Limit"); h != "" {
		r.Meta.Ratelimit.Limit, _ = strconv.Atoi(h)
	}
	if h := r.Header.Get("RateLimit-Remaining"); h != "" {
		r.Meta.Ratelimit.Remaining, _ = strconv.Atoi(h)
	}
	if h := r.Header.Get("RateLimit-Reset"); h != "" {
		if ts, err := strconv.ParseInt(h, 10, 64); err == nil {
			r.Meta.Ratelimit.Reset = time.Unix(ts, 0)
		}
	}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var s schema.MetaResponse
		if err := json.Unmarshal(body, &s); err != nil {
			return err
		}
		if s.Meta.Pagination != nil {
			p := PaginationFromSchema(*s.Meta.Pagination)
			r.Meta.Pagination = &p
		}
	}

	return nil
}

// internalCorrelationID returns the unique ID of the request as set by the API. This ID can help with support requests,
// as it allows the people working on identify this request in particular.
func (r *Response) internalCorrelationID() string {
	return r.Header.Get(headerCorrelationID)
}

// Meta represents meta information included in an API response.
type Meta struct {
	Pagination *Pagination
	Ratelimit  Ratelimit
}

// Pagination represents pagination meta information.
type Pagination struct {
	Page         int
	PerPage      int
	PreviousPage int
	NextPage     int
	LastPage     int
	TotalEntries int
}

// Ratelimit represents ratelimit information.
type Ratelimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// ListOpts specifies options for listing resources.
type ListOpts struct {
	Page          int    // Page (starting at 1)
	PerPage       int    // Items per page (0 means default)
	LabelSelector string // Label selector for filtering by labels
}

// Values returns the ListOpts as URL values.
func (l ListOpts) Values() url.Values {
	vals := url.Values{}
	if l.Page > 0 {
		vals.Add("page", strconv.Itoa(l.Page))
	}
	if l.PerPage > 0 {
		vals.Add("per_page", strconv.Itoa(l.PerPage))
	}
	if len(l.LabelSelector) > 0 {
		vals.Add("label_selector", l.LabelSelector)
	}
	return vals
}
