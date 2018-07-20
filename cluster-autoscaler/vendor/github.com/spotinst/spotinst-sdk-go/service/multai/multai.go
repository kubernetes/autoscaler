package multai

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

// A Protocol represents the type of an application protocol.
type Protocol int

const (
	// ProtocolHTTP represents the Hypertext Transfer Protocol (HTTP) protocol.
	ProtocolHTTP Protocol = iota

	// ProtocolHTTPS represents the Hypertext Transfer Protocol (HTTP) within
	// a connection encrypted by Transport Layer Security, or its predecessor,
	// Secure Sockets Layer.
	ProtocolHTTPS

	// ProtocolHTTP2 represents the Hypertext Transfer Protocol (HTTP) protocol
	// version 2.
	ProtocolHTTP2
)

var Protocol_name = map[Protocol]string{
	ProtocolHTTP:  "HTTP",
	ProtocolHTTPS: "HTTPS",
	ProtocolHTTP2: "HTTP2",
}

var Protocol_value = map[string]Protocol{
	"HTTP":  ProtocolHTTP,
	"HTTPS": ProtocolHTTPS,
	"HTTP2": ProtocolHTTP2,
}

func (p Protocol) String() string {
	return Protocol_name[p]
}

// A ReadinessStatus represents the readiness status of a target.
type ReadinessStatus int

const (
	// StatusReady represents a ready state.
	StatusReady ReadinessStatus = iota

	// StatusMaintenance represents a maintenance state.
	StatusMaintenance
)

var ReadinessStatus_name = map[ReadinessStatus]string{
	StatusReady:       "READY",
	StatusMaintenance: "MAINTENANCE",
}

var ReadinessStatus_value = map[string]ReadinessStatus{
	"READY":       StatusReady,
	"MAINTENANCE": StatusMaintenance,
}

func (s ReadinessStatus) String() string {
	return ReadinessStatus_name[s]
}

// A HealthinessStatus represents the healthiness status of a target.
type HealthinessStatus int

const (
	// StatusUnknown represents an unknown state.
	StatusUnknown HealthinessStatus = iota

	// StatusHealthy represents a healthy state.
	StatusHealthy

	// StatusUnhealthy represents an unhealthy state.
	StatusUnhealthy
)

var HealthinessStatus_name = map[HealthinessStatus]string{
	StatusUnknown:   "UNKNOWN",
	StatusHealthy:   "HEALTHY",
	StatusUnhealthy: "UNHEALTHY",
}

var HealthinessStatus_value = map[string]HealthinessStatus{
	"UNKNOWN":   StatusUnknown,
	"HEALTHY":   StatusHealthy,
	"UNHEALTHY": StatusUnhealthy,
}

func (s HealthinessStatus) String() string {
	return HealthinessStatus_name[s]
}

// A Strategy represents the load balancing methods used to determine which
// application server to send a request to.
type Strategy int

const (
	// StrategyRandom represents a random load balancing method where
	// a request is passed to the server with the least number of
	// active connections.
	StrategyRandom Strategy = iota

	// StrategyRoundRobin represents a random load balancing method where
	// a request is passed to the server in round-robin fashion.
	StrategyRoundRobin

	// StrategyLeastConn represents a random load balancing method where
	// a request is passed to the server with the least number of
	// active connections.
	StrategyLeastConn

	// StrategyIPHash represents a IP hash load balancing method where
	// a request is passed to the server based on the result of hashing
	// the request IP address.
	StrategyIPHash
)

var Strategy_name = map[Strategy]string{
	StrategyRandom:     "RANDOM",
	StrategyRoundRobin: "ROUNDROBIN",
	StrategyLeastConn:  "LEASTCONN",
	StrategyIPHash:     "IPHASH",
}

var Strategy_value = map[string]Strategy{
	"RANDOM":     StrategyRandom,
	"ROUNDROBIN": StrategyRoundRobin,
	"LEASTCONN":  StrategyLeastConn,
	"IPHASH":     StrategyIPHash,
}

func (s Strategy) String() string {
	return Strategy_name[s]
}

type LoadBalancer struct {
	ID              *string    `json:"id,omitempty"`
	Name            *string    `json:"name,omitempty"`
	DNSRRType       *string    `json:"dnsRrType,omitempty"`
	DNSRRName       *string    `json:"dnsRrName,omitempty"`
	DNSCNAMEAliases []string   `json:"dnsCnameAliases,omitempty"`
	Timeouts        *Timeouts  `json:"timeouts,omitempty"`
	Tags            []*Tag     `json:"tags,omitempty"`
	CreatedAt       *time.Time `json:"createdAt,omitempty"`
	UpdatedAt       *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type Timeouts struct {
	Idle     *int `json:"idle"`
	Draining *int `json:"draining"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListLoadBalancersInput struct {
	DeploymentID *string `json:"deploymentId,omitempty"`
}

type ListLoadBalancersOutput struct {
	Balancers []*LoadBalancer `json:"balancers,omitempty"`
}

type CreateLoadBalancerInput struct {
	Balancer *LoadBalancer `json:"balancer,omitempty"`
}

type CreateLoadBalancerOutput struct {
	Balancer *LoadBalancer `json:"balancer,omitempty"`
}

type ReadLoadBalancerInput struct {
	BalancerID *string `json:"balancerId,omitempty"`
}

type ReadLoadBalancerOutput struct {
	Balancer *LoadBalancer `json:"balancer,omitempty"`
}

type UpdateLoadBalancerInput struct {
	Balancer *LoadBalancer `json:"balancer,omitempty"`
}

type UpdateLoadBalancerOutput struct{}

type DeleteLoadBalancerInput struct {
	BalancerID *string `json:"balancerId,omitempty"`
}

type DeleteLoadBalancerOutput struct{}

func balancerFromJSON(in []byte) (*LoadBalancer, error) {
	b := new(LoadBalancer)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func balancersFromJSON(in []byte) ([]*LoadBalancer, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*LoadBalancer, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := balancerFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func balancersFromHttpResponse(resp *http.Response) ([]*LoadBalancer, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return balancersFromJSON(body)
}

func (s *ServiceOp) ListLoadBalancers(ctx context.Context, input *ListLoadBalancersInput) (*ListLoadBalancersOutput, error) {
	r := client.NewRequest(http.MethodGet, "/loadBalancer/balancer")

	if input.DeploymentID != nil {
		r.Params.Set("deploymentId", spotinst.StringValue(input.DeploymentID))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bs, err := balancersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListLoadBalancersOutput{Balancers: bs}, nil
}

func (s *ServiceOp) CreateLoadBalancer(ctx context.Context, input *CreateLoadBalancerInput) (*CreateLoadBalancerOutput, error) {
	r := client.NewRequest(http.MethodPost, "/loadBalancer/balancer")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bs, err := balancersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateLoadBalancerOutput)
	if len(bs) > 0 {
		output.Balancer = bs[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadLoadBalancer(ctx context.Context, input *ReadLoadBalancerInput) (*ReadLoadBalancerOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/balancer/{balancerId}", uritemplates.Values{
		"balancerId": spotinst.StringValue(input.BalancerID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bs, err := balancersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadLoadBalancerOutput)
	if len(bs) > 0 {
		output.Balancer = bs[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateLoadBalancer(ctx context.Context, input *UpdateLoadBalancerInput) (*UpdateLoadBalancerOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/balancer/{balancerId}", uritemplates.Values{
		"balancerId": spotinst.StringValue(input.Balancer.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Balancer.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &UpdateLoadBalancerOutput{}, nil
}

func (s *ServiceOp) DeleteLoadBalancer(ctx context.Context, input *DeleteLoadBalancerInput) (*DeleteLoadBalancerOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/balancer/{balancerId}", uritemplates.Values{
		"balancerId": spotinst.StringValue(input.BalancerID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteLoadBalancerOutput{}, nil
}

// region LoadBalancer

func (o *LoadBalancer) MarshalJSON() ([]byte, error) {
	type noMethod LoadBalancer
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *LoadBalancer) SetId(v *string) *LoadBalancer {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *LoadBalancer) SetName(v *string) *LoadBalancer {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *LoadBalancer) SetDNSRRType(v *string) *LoadBalancer {
	if o.DNSRRType = v; o.DNSRRType == nil {
		o.nullFields = append(o.nullFields, "DNSRRType")
	}
	return o
}

func (o *LoadBalancer) SetDNSRRName(v *string) *LoadBalancer {
	if o.DNSRRName = v; o.DNSRRName == nil {
		o.nullFields = append(o.nullFields, "DNSRRName")
	}
	return o
}

func (o *LoadBalancer) SetDNSCNAMEAliases(v []string) *LoadBalancer {
	if o.DNSCNAMEAliases = v; o.DNSCNAMEAliases == nil {
		o.nullFields = append(o.nullFields, "DNSCNAMEAliases")
	}
	return o
}

func (o *LoadBalancer) SetTimeouts(v *Timeouts) *LoadBalancer {
	if o.Timeouts = v; o.Timeouts == nil {
		o.nullFields = append(o.nullFields, "Timeouts")
	}
	return o
}

func (o *LoadBalancer) SetTags(v []*Tag) *LoadBalancer {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// endregion

type Listener struct {
	ID         *string    `json:"id,omitempty"`
	BalancerID *string    `json:"balancerId,omitempty"`
	Protocol   *string    `json:"protocol,omitempty"`
	Port       *int       `json:"port,omitempty"`
	TLSConfig  *TLSConfig `json:"tlsConfig,omitempty"`
	Tags       []*Tag     `json:"tags,omitempty"`
	CreatedAt  *time.Time `json:"createdAt,omitempty"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListListenersInput struct {
	BalancerID *string `json:"balancerId,omitempty"`
}

type ListListenersOutput struct {
	Listeners []*Listener `json:"listeners,omitempty"`
}

type CreateListenerInput struct {
	Listener *Listener `json:"listener,omitempty"`
}

type CreateListenerOutput struct {
	Listener *Listener `json:"listener,omitempty"`
}

type ReadListenerInput struct {
	ListenerID *string `json:"listenerId,omitempty"`
}

type ReadListenerOutput struct {
	Listener *Listener `json:"listener,omitempty"`
}

type UpdateListenerInput struct {
	Listener *Listener `json:"listener,omitempty"`
}

type UpdateListenerOutput struct{}

type DeleteListenerInput struct {
	ListenerID *string `json:"listenerId,omitempty"`
}

type DeleteListenerOutput struct{}

func listenerFromJSON(in []byte) (*Listener, error) {
	b := new(Listener)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func listenersFromJSON(in []byte) ([]*Listener, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Listener, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rl := range rw.Response.Items {
		l, err := listenerFromJSON(rl)
		if err != nil {
			return nil, err
		}
		out[i] = l
	}
	return out, nil
}

func listenersFromHttpResponse(resp *http.Response) ([]*Listener, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return listenersFromJSON(body)
}

func (s *ServiceOp) ListListeners(ctx context.Context, input *ListListenersInput) (*ListListenersOutput, error) {
	r := client.NewRequest(http.MethodGet, "/loadBalancer/listener")

	if input.BalancerID != nil {
		r.Params.Set("balancerId", spotinst.StringValue(input.BalancerID))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ls, err := listenersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListListenersOutput{Listeners: ls}, nil
}

func (s *ServiceOp) CreateListener(ctx context.Context, input *CreateListenerInput) (*CreateListenerOutput, error) {
	r := client.NewRequest(http.MethodPost, "/loadBalancer/listener")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ls, err := listenersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateListenerOutput)
	if len(ls) > 0 {
		output.Listener = ls[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadListener(ctx context.Context, input *ReadListenerInput) (*ReadListenerOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/listener/{listenerId}", uritemplates.Values{
		"listenerId": spotinst.StringValue(input.ListenerID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ls, err := listenersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadListenerOutput)
	if len(ls) > 0 {
		output.Listener = ls[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateListener(ctx context.Context, input *UpdateListenerInput) (*UpdateListenerOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/listener/{listenerId}", uritemplates.Values{
		"listenerId": spotinst.StringValue(input.Listener.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Listener.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &UpdateListenerOutput{}, nil
}

func (s *ServiceOp) DeleteListener(ctx context.Context, input *DeleteListenerInput) (*DeleteListenerOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/listener/{listenerId}", uritemplates.Values{
		"listenerId": spotinst.StringValue(input.ListenerID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteListenerOutput{}, nil
}

// region Listener

func (o *Listener) MarshalJSON() ([]byte, error) {
	type noMethod Listener
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Listener) SetId(v *string) *Listener {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Listener) SetBalancerId(v *string) *Listener {
	if o.BalancerID = v; o.BalancerID == nil {
		o.nullFields = append(o.nullFields, "BalancerID")
	}
	return o
}

func (o *Listener) SetProtocol(v *string) *Listener {
	if o.Protocol = v; o.Protocol == nil {
		o.nullFields = append(o.nullFields, "Protocol")
	}
	return o
}

func (o *Listener) SetPort(v *int) *Listener {
	if o.Port = v; o.Port == nil {
		o.nullFields = append(o.nullFields, "Port")
	}
	return o
}

func (o *Listener) SetTLSConfig(v *TLSConfig) *Listener {
	if o.TLSConfig = v; o.TLSConfig == nil {
		o.nullFields = append(o.nullFields, "TLSConfig")
	}
	return o
}

func (o *Listener) SetTags(v []*Tag) *Listener {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// endregion

// region TLSConfig

func (o *TLSConfig) MarshalJSON() ([]byte, error) {
	type noMethod TLSConfig
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *TLSConfig) SetCertificateIDs(v []string) *TLSConfig {
	if o.CertificateIDs = v; o.CertificateIDs == nil {
		o.nullFields = append(o.nullFields, "CertificateIDs")
	}
	return o
}

func (o *TLSConfig) SetMinVersion(v *string) *TLSConfig {
	if o.MinVersion = v; o.MinVersion == nil {
		o.nullFields = append(o.nullFields, "MinVersion")
	}
	return o
}

func (o *TLSConfig) SetMaxVersion(v *string) *TLSConfig {
	if o.MaxVersion = v; o.MaxVersion == nil {
		o.nullFields = append(o.nullFields, "MaxVersion")
	}
	return o
}

func (o *TLSConfig) SetSessionTicketsDisabled(v *bool) *TLSConfig {
	if o.SessionTicketsDisabled = v; o.SessionTicketsDisabled == nil {
		o.nullFields = append(o.nullFields, "SessionTicketsDisabled")
	}
	return o
}

func (o *TLSConfig) SetPreferServerCipherSuites(v *bool) *TLSConfig {
	if o.PreferServerCipherSuites = v; o.PreferServerCipherSuites == nil {
		o.nullFields = append(o.nullFields, "PreferServerCipherSuites")
	}
	return o
}

func (o *TLSConfig) SetCipherSuites(v []string) *TLSConfig {
	if o.CipherSuites = v; o.CipherSuites == nil {
		o.nullFields = append(o.nullFields, "CipherSuites")
	}
	return o
}

func (o *TLSConfig) SetInsecureSkipVerify(v *bool) *TLSConfig {
	if o.InsecureSkipVerify = v; o.InsecureSkipVerify == nil {
		o.nullFields = append(o.nullFields, "InsecureSkipVerify")
	}
	return o
}

// endregion

type RoutingRule struct {
	ID            *string    `json:"id,omitempty"`
	BalancerID    *string    `json:"balancerId,omitempty"`
	ListenerID    *string    `json:"listenerId,omitempty"`
	MiddlewareIDs []string   `json:"middlewareIds,omitempty"`
	TargetSetIDs  []string   `json:"targetSetIds,omitempty"`
	Priority      *int       `json:"priority,omitempty"`
	Strategy      *string    `json:"strategy,omitempty"`
	Route         *string    `json:"route,omitempty"`
	Tags          []*Tag     `json:"tags,omitempty"`
	CreatedAt     *time.Time `json:"createdAt,omitempty"`
	UpdatedAt     *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListRoutingRulesInput struct {
	BalancerID *string `json:"balancerId,omitempty"`
}

type ListRoutingRulesOutput struct {
	RoutingRules []*RoutingRule `json:"routingRules,omitempty"`
}

type CreateRoutingRuleInput struct {
	RoutingRule *RoutingRule `json:"routingRule,omitempty"`
}

type CreateRoutingRuleOutput struct {
	RoutingRule *RoutingRule `json:"routingRule,omitempty"`
}

type ReadRoutingRuleInput struct {
	RoutingRuleID *string `json:"routingRuleId,omitempty"`
}

type ReadRoutingRuleOutput struct {
	RoutingRule *RoutingRule `json:"routingRule,omitempty"`
}

type UpdateRoutingRuleInput struct {
	RoutingRule *RoutingRule `json:"routingRule,omitempty"`
}

type UpdateRoutingRuleOutput struct{}

type DeleteRoutingRuleInput struct {
	RoutingRuleID *string `json:"routingRuleId,omitempty"`
}

type DeleteRoutingRuleOutput struct{}

func routingRuleFromJSON(in []byte) (*RoutingRule, error) {
	rr := new(RoutingRule)
	if err := json.Unmarshal(in, rr); err != nil {
		return nil, err
	}
	return rr, nil
}

func routingRulesFromJSON(in []byte) ([]*RoutingRule, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*RoutingRule, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rr := range rw.Response.Items {
		r, err := routingRuleFromJSON(rr)
		if err != nil {
			return nil, err
		}
		out[i] = r
	}
	return out, nil
}

func routingRulesFromHttpResponse(resp *http.Response) ([]*RoutingRule, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return routingRulesFromJSON(body)
}

func (s *ServiceOp) ListRoutingRules(ctx context.Context, input *ListRoutingRulesInput) (*ListRoutingRulesOutput, error) {
	r := client.NewRequest(http.MethodGet, "/loadBalancer/routingRule")

	if input.BalancerID != nil {
		r.Params.Set("balancerId", spotinst.StringValue(input.BalancerID))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rr, err := routingRulesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListRoutingRulesOutput{RoutingRules: rr}, nil
}

func (s *ServiceOp) CreateRoutingRule(ctx context.Context, input *CreateRoutingRuleInput) (*CreateRoutingRuleOutput, error) {
	r := client.NewRequest(http.MethodPost, "/loadBalancer/routingRule")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rr, err := routingRulesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateRoutingRuleOutput)
	if len(rr) > 0 {
		output.RoutingRule = rr[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadRoutingRule(ctx context.Context, input *ReadRoutingRuleInput) (*ReadRoutingRuleOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/routingRule/{routingRuleId}", uritemplates.Values{
		"routingRuleId": spotinst.StringValue(input.RoutingRuleID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rr, err := routingRulesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadRoutingRuleOutput)
	if len(rr) > 0 {
		output.RoutingRule = rr[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateRoutingRule(ctx context.Context, input *UpdateRoutingRuleInput) (*UpdateRoutingRuleOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/routingRule/{routingRuleId}", uritemplates.Values{
		"routingRuleId": spotinst.StringValue(input.RoutingRule.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.RoutingRule.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &UpdateRoutingRuleOutput{}, nil
}

func (s *ServiceOp) DeleteRoutingRule(ctx context.Context, input *DeleteRoutingRuleInput) (*DeleteRoutingRuleOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/routingRule/{routingRuleId}", uritemplates.Values{
		"routingRuleId": spotinst.StringValue(input.RoutingRuleID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteRoutingRuleOutput{}, nil
}

// region RoutingRule

func (o *RoutingRule) MarshalJSON() ([]byte, error) {
	type noMethod RoutingRule
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *RoutingRule) SetId(v *string) *RoutingRule {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *RoutingRule) SetBalancerId(v *string) *RoutingRule {
	if o.BalancerID = v; o.BalancerID == nil {
		o.nullFields = append(o.nullFields, "BalancerID")
	}
	return o
}

func (o *RoutingRule) SetListenerId(v *string) *RoutingRule {
	if o.ListenerID = v; o.ListenerID == nil {
		o.nullFields = append(o.nullFields, "ListenerID")
	}
	return o
}

func (o *RoutingRule) SetMiddlewareIDs(v []string) *RoutingRule {
	if o.MiddlewareIDs = v; o.MiddlewareIDs == nil {
		o.nullFields = append(o.nullFields, "MiddlewareIDs")
	}
	return o
}

func (o *RoutingRule) SetTargetSetIDs(v []string) *RoutingRule {
	if o.TargetSetIDs = v; o.TargetSetIDs == nil {
		o.nullFields = append(o.nullFields, "TargetSetIDs")
	}
	return o
}

func (o *RoutingRule) SetPriority(v *int) *RoutingRule {
	if o.Priority = v; o.Priority == nil {
		o.nullFields = append(o.nullFields, "Priority")
	}
	return o
}

func (o *RoutingRule) SetStrategy(v *string) *RoutingRule {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

func (o *RoutingRule) SetRoute(v *string) *RoutingRule {
	if o.Route = v; o.Route == nil {
		o.nullFields = append(o.nullFields, "Route")
	}
	return o
}

func (o *RoutingRule) SetTags(v []*Tag) *RoutingRule {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// endregion

type Middleware struct {
	ID         *string         `json:"id,omitempty"`
	BalancerID *string         `json:"balancerId,omitempty"`
	Type       *string         `json:"type,omitempty"`
	Priority   *int            `json:"priority,omitempty"`
	Spec       json.RawMessage `json:"spec,omitempty"`
	Tags       []*Tag          `json:"tags,omitempty"`
	CreatedAt  *time.Time      `json:"createdAt,omitempty"`
	UpdatedAt  *time.Time      `json:"updatedAt,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListMiddlewaresInput struct {
	BalancerID *string `json:"balancerId,omitempty"`
}

type ListMiddlewaresOutput struct {
	Middlewares []*Middleware `json:"middlewares,omitempty"`
}

type CreateMiddlewareInput struct {
	Middleware *Middleware `json:"middleware,omitempty"`
}

type CreateMiddlewareOutput struct {
	Middleware *Middleware `json:"middleware,omitempty"`
}

type ReadMiddlewareInput struct {
	MiddlewareID *string `json:"middlewareId,omitempty"`
}

type ReadMiddlewareOutput struct {
	Middleware *Middleware `json:"middleware,omitempty"`
}

type UpdateMiddlewareInput struct {
	Middleware *Middleware `json:"middleware,omitempty"`
}

type UpdateMiddlewareOutput struct{}

type DeleteMiddlewareInput struct {
	MiddlewareID *string `json:"middlewareId,omitempty"`
}

type DeleteMiddlewareOutput struct{}

func middlewareFromJSON(in []byte) (*Middleware, error) {
	m := new(Middleware)
	if err := json.Unmarshal(in, m); err != nil {
		return nil, err
	}
	return m, nil
}

func middlewaresFromJSON(in []byte) ([]*Middleware, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Middleware, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rm := range rw.Response.Items {
		m, err := middlewareFromJSON(rm)
		if err != nil {
			return nil, err
		}
		out[i] = m
	}
	return out, nil
}

func middlewaresFromHttpResponse(resp *http.Response) ([]*Middleware, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return middlewaresFromJSON(body)
}

func (s *ServiceOp) ListMiddlewares(ctx context.Context, input *ListMiddlewaresInput) (*ListMiddlewaresOutput, error) {
	r := client.NewRequest(http.MethodGet, "/loadBalancer/middleware")

	if input.BalancerID != nil {
		r.Params.Set("balancerId", spotinst.StringValue(input.BalancerID))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ms, err := middlewaresFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListMiddlewaresOutput{Middlewares: ms}, nil
}

func (s *ServiceOp) CreateMiddleware(ctx context.Context, input *CreateMiddlewareInput) (*CreateMiddlewareOutput, error) {
	r := client.NewRequest(http.MethodPost, "/loadBalancer/middleware")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ms, err := middlewaresFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateMiddlewareOutput)
	if len(ms) > 0 {
		output.Middleware = ms[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadMiddleware(ctx context.Context, input *ReadMiddlewareInput) (*ReadMiddlewareOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/middleware/{middlewareId}", uritemplates.Values{
		"middlewareId": spotinst.StringValue(input.MiddlewareID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ms, err := middlewaresFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadMiddlewareOutput)
	if len(ms) > 0 {
		output.Middleware = ms[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateMiddleware(ctx context.Context, input *UpdateMiddlewareInput) (*UpdateMiddlewareOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/middleware/{middlewareId}", uritemplates.Values{
		"middlewareId": spotinst.StringValue(input.Middleware.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Middleware.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &UpdateMiddlewareOutput{}, nil
}

func (s *ServiceOp) DeleteMiddleware(ctx context.Context, input *DeleteMiddlewareInput) (*DeleteMiddlewareOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/middleware/{middlewareId}", uritemplates.Values{
		"middlewareId": spotinst.StringValue(input.MiddlewareID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteMiddlewareOutput{}, nil
}

// region Middleware

func (o *Middleware) MarshalJSON() ([]byte, error) {
	type noMethod Middleware
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Middleware) SetId(v *string) *Middleware {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Middleware) SetBalancerId(v *string) *Middleware {
	if o.BalancerID = v; o.BalancerID == nil {
		o.nullFields = append(o.nullFields, "BalancerID")
	}
	return o
}

func (o *Middleware) SetType(v *string) *Middleware {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *Middleware) SetPriority(v *int) *Middleware {
	if o.Priority = v; o.Priority == nil {
		o.nullFields = append(o.nullFields, "Priority")
	}
	return o
}

func (o *Middleware) SetSpec(v json.RawMessage) *Middleware {
	if o.Spec = v; o.Spec == nil {
		o.nullFields = append(o.nullFields, "Spec")
	}
	return o
}

func (o *Middleware) SetTags(v []*Tag) *Middleware {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// endregion

type TargetSet struct {
	ID           *string               `json:"id,omitempty"`
	BalancerID   *string               `json:"balancerId,omitempty"`
	DeploymentID *string               `json:"deploymentId,omitempty"`
	Name         *string               `json:"name,omitempty"`
	Protocol     *string               `json:"protocol,omitempty"`
	Port         *int                  `json:"port,omitempty"`
	Weight       *int                  `json:"weight,omitempty"`
	HealthCheck  *TargetSetHealthCheck `json:"healthCheck,omitempty"`
	Tags         []*Tag                `json:"tags,omitempty"`
	CreatedAt    *time.Time            `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time            `json:"updatedAt,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type TargetSetHealthCheck struct {
	Path                    *string `json:"path,omitempty"`
	Port                    *int    `json:"port,omitempty"`
	Protocol                *string `json:"protocol,omitempty"`
	Timeout                 *int    `json:"timeout,omitempty"`
	Interval                *int    `json:"interval,omitempty"`
	HealthyThresholdCount   *int    `json:"healthyThresholdCount,omitempty"`
	UnhealthyThresholdCount *int    `json:"unhealthyThresholdCount,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListTargetSetsInput struct {
	BalancerID *string `json:"balancerId,omitempty"`
}

type ListTargetSetsOutput struct {
	TargetSets []*TargetSet `json:"targetSets,omitempty"`
}

type CreateTargetSetInput struct {
	TargetSet *TargetSet `json:"targetSet,omitempty"`
}

type CreateTargetSetOutput struct {
	TargetSet *TargetSet `json:"targetSet,omitempty"`
}

type ReadTargetSetInput struct {
	TargetSetID *string `json:"targetSetId,omitempty"`
}

type ReadTargetSetOutput struct {
	TargetSet *TargetSet `json:"targetSet,omitempty"`
}

type UpdateTargetSetInput struct {
	TargetSet *TargetSet `json:"targetSet,omitempty"`
}

type UpdateTargetSetOutput struct{}

type DeleteTargetSetInput struct {
	TargetSetID *string `json:"targetSetId,omitempty"`
}

type DeleteTargetSetOutput struct{}

func targetSetFromJSON(in []byte) (*TargetSet, error) {
	ts := new(TargetSet)
	if err := json.Unmarshal(in, ts); err != nil {
		return nil, err
	}
	return ts, nil
}

func targetSetsFromJSON(in []byte) ([]*TargetSet, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*TargetSet, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rts := range rw.Response.Items {
		ts, err := targetSetFromJSON(rts)
		if err != nil {
			return nil, err
		}
		out[i] = ts
	}
	return out, nil
}

func targetSetsFromHttpResponse(resp *http.Response) ([]*TargetSet, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return targetSetsFromJSON(body)
}

func (s *ServiceOp) ListTargetSets(ctx context.Context, input *ListTargetSetsInput) (*ListTargetSetsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/loadBalancer/targetSet")

	if input.BalancerID != nil {
		r.Params.Set("balancerId", spotinst.StringValue(input.BalancerID))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ts, err := targetSetsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListTargetSetsOutput{TargetSets: ts}, nil
}

func (s *ServiceOp) CreateTargetSet(ctx context.Context, input *CreateTargetSetInput) (*CreateTargetSetOutput, error) {
	r := client.NewRequest(http.MethodPost, "/loadBalancer/targetSet")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ts, err := targetSetsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateTargetSetOutput)
	if len(ts) > 0 {
		output.TargetSet = ts[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadTargetSet(ctx context.Context, input *ReadTargetSetInput) (*ReadTargetSetOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/targetSet/{targetSetId}", uritemplates.Values{
		"targetSetId": spotinst.StringValue(input.TargetSetID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ts, err := targetSetsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadTargetSetOutput)
	if len(ts) > 0 {
		output.TargetSet = ts[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateTargetSet(ctx context.Context, input *UpdateTargetSetInput) (*UpdateTargetSetOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/targetSet/{targetSetId}", uritemplates.Values{
		"targetSetId": spotinst.StringValue(input.TargetSet.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.TargetSet.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &UpdateTargetSetOutput{}, nil
}

func (s *ServiceOp) DeleteTargetSet(ctx context.Context, input *DeleteTargetSetInput) (*DeleteTargetSetOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/targetSet/{targetSetId}", uritemplates.Values{
		"targetSetId": spotinst.StringValue(input.TargetSetID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteTargetSetOutput{}, nil
}

// region TargetSet

func (o *TargetSet) MarshalJSON() ([]byte, error) {
	type noMethod TargetSet
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *TargetSet) SetId(v *string) *TargetSet {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *TargetSet) SetBalancerId(v *string) *TargetSet {
	if o.BalancerID = v; o.BalancerID == nil {
		o.nullFields = append(o.nullFields, "BalancerID")
	}
	return o
}

func (o *TargetSet) SetDeploymentId(v *string) *TargetSet {
	if o.DeploymentID = v; o.DeploymentID == nil {
		o.nullFields = append(o.nullFields, "DeploymentID")
	}
	return o
}

func (o *TargetSet) SetName(v *string) *TargetSet {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *TargetSet) SetProtocol(v *string) *TargetSet {
	if o.Protocol = v; o.Protocol == nil {
		o.nullFields = append(o.nullFields, "Protocol")
	}
	return o
}

func (o *TargetSet) SetPort(v *int) *TargetSet {
	if o.Port = v; o.Port == nil {
		o.nullFields = append(o.nullFields, "Port")
	}
	return o
}

func (o *TargetSet) SetWeight(v *int) *TargetSet {
	if o.Weight = v; o.Weight == nil {
		o.nullFields = append(o.nullFields, "Weight")
	}
	return o
}

func (o *TargetSet) SetHealthCheck(v *TargetSetHealthCheck) *TargetSet {
	if o.HealthCheck = v; o.HealthCheck == nil {
		o.nullFields = append(o.nullFields, "HealthCheck")
	}
	return o
}

func (o *TargetSet) SetTags(v []*Tag) *TargetSet {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// endregion

// region TargetSetHealthCheck

func (o *TargetSetHealthCheck) MarshalJSON() ([]byte, error) {
	type noMethod TargetSetHealthCheck
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *TargetSetHealthCheck) SetPath(v *string) *TargetSetHealthCheck {
	if o.Path = v; o.Path == nil {
		o.nullFields = append(o.nullFields, "Path")
	}
	return o
}

func (o *TargetSetHealthCheck) SetPort(v *int) *TargetSetHealthCheck {
	if o.Port = v; o.Port == nil {
		o.nullFields = append(o.nullFields, "Port")
	}
	return o
}

func (o *TargetSetHealthCheck) SetProtocol(v *string) *TargetSetHealthCheck {
	if o.Protocol = v; o.Protocol == nil {
		o.nullFields = append(o.nullFields, "Protocol")
	}
	return o
}

func (o *TargetSetHealthCheck) SetTimeout(v *int) *TargetSetHealthCheck {
	if o.Timeout = v; o.Timeout == nil {
		o.nullFields = append(o.nullFields, "Timeout")
	}
	return o
}

func (o *TargetSetHealthCheck) SetInterval(v *int) *TargetSetHealthCheck {
	if o.Interval = v; o.Interval == nil {
		o.nullFields = append(o.nullFields, "Interval")
	}
	return o
}

func (o *TargetSetHealthCheck) SetHealthyThresholdCount(v *int) *TargetSetHealthCheck {
	if o.HealthyThresholdCount = v; o.HealthyThresholdCount == nil {
		o.nullFields = append(o.nullFields, "HealthyThresholdCount")
	}
	return o
}

func (o *TargetSetHealthCheck) SetUnhealthyThresholdCount(v *int) *TargetSetHealthCheck {
	if o.UnhealthyThresholdCount = v; o.UnhealthyThresholdCount == nil {
		o.nullFields = append(o.nullFields, "UnhealthyThresholdCount")
	}
	return o
}

// endregion

type Target struct {
	ID          *string    `json:"id,omitempty"`
	BalancerID  *string    `json:"balancerId,omitempty"`
	TargetSetID *string    `json:"targetSetId,omitempty"`
	Name        *string    `json:"name,omitempty"`
	Host        *string    `json:"host,omitempty"`
	Port        *int       `json:"port,omitempty"`
	Weight      *int       `json:"weight,omitempty"`
	Status      *Status    `json:"status,omitempty"`
	Tags        []*Tag     `json:"tags,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type Status struct {
	Readiness   *string `json:"readiness"`
	Healthiness *string `json:"healthiness"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListTargetsInput struct {
	BalancerID  *string `json:"balancerId,omitempty"`
	TargetSetID *string `json:"targetSetId,omitempty"`
}

type ListTargetsOutput struct {
	Targets []*Target `json:"targets,omitempty"`
}

type CreateTargetInput struct {
	TargetSetID *string `json:"targetSetId,omitempty"`
	Target      *Target `json:"target,omitempty"`
}

type CreateTargetOutput struct {
	Target *Target `json:"target,omitempty"`
}

type ReadTargetInput struct {
	TargetSetID *string `json:"targetSetId,omitempty"`
	TargetID    *string `json:"targetId,omitempty"`
}

type ReadTargetOutput struct {
	Target *Target `json:"target,omitempty"`
}

type UpdateTargetInput struct {
	TargetSetID *string `json:"targetSetId,omitempty"`
	Target      *Target `json:"target,omitempty"`
}

type UpdateTargetOutput struct{}

type DeleteTargetInput struct {
	TargetSetID *string `json:"targetSetId,omitempty"`
	TargetID    *string `json:"targetId,omitempty"`
}

type DeleteTargetOutput struct{}

func targetFromJSON(in []byte) (*Target, error) {
	t := new(Target)
	if err := json.Unmarshal(in, t); err != nil {
		return nil, err
	}
	return t, nil
}

func targetsFromJSON(in []byte) ([]*Target, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Target, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rt := range rw.Response.Items {
		t, err := targetFromJSON(rt)
		if err != nil {
			return nil, err
		}
		out[i] = t
	}
	return out, nil
}

func targetsFromHttpResponse(resp *http.Response) ([]*Target, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return targetsFromJSON(body)
}

func (s *ServiceOp) ListTargets(ctx context.Context, input *ListTargetsInput) (*ListTargetsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/loadBalancer/target")

	if input.BalancerID != nil {
		r.Params.Set("balancerId", spotinst.StringValue(input.BalancerID))
	}

	if input.TargetSetID != nil {
		r.Params.Set("targetSetId", spotinst.StringValue(input.TargetSetID))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ts, err := targetsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListTargetsOutput{Targets: ts}, nil
}

func (s *ServiceOp) CreateTarget(ctx context.Context, input *CreateTargetInput) (*CreateTargetOutput, error) {
	r := client.NewRequest(http.MethodPost, "/loadBalancer/target")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ts, err := targetsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateTargetOutput)
	if len(ts) > 0 {
		output.Target = ts[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadTarget(ctx context.Context, input *ReadTargetInput) (*ReadTargetOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/target/{targetId}", uritemplates.Values{
		"targetId": spotinst.StringValue(input.TargetID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ts, err := targetsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadTargetOutput)
	if len(ts) > 0 {
		output.Target = ts[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateTarget(ctx context.Context, input *UpdateTargetInput) (*UpdateTargetOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/target/{targetId}", uritemplates.Values{
		"targetId": spotinst.StringValue(input.Target.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Target.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &UpdateTargetOutput{}, nil
}

func (s *ServiceOp) DeleteTarget(ctx context.Context, input *DeleteTargetInput) (*DeleteTargetOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/target/{targetId}", uritemplates.Values{
		"targetId": spotinst.StringValue(input.TargetID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteTargetOutput{}, nil
}

// region Target

func (o *Target) MarshalJSON() ([]byte, error) {
	type noMethod Target
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Target) SetId(v *string) *Target {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Target) SetBalancerId(v *string) *Target {
	if o.BalancerID = v; o.BalancerID == nil {
		o.nullFields = append(o.nullFields, "BalancerID")
	}
	return o
}

func (o *Target) SetTargetSetId(v *string) *Target {
	if o.TargetSetID = v; o.TargetSetID == nil {
		o.nullFields = append(o.nullFields, "TargetSetID")
	}
	return o
}

func (o *Target) SetName(v *string) *Target {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Target) SetHost(v *string) *Target {
	if o.Host = v; o.Host == nil {
		o.nullFields = append(o.nullFields, "Host")
	}
	return o
}

func (o *Target) SetPort(v *int) *Target {
	if o.Port = v; o.Port == nil {
		o.nullFields = append(o.nullFields, "Port")
	}
	return o
}

func (o *Target) SetWeight(v *int) *Target {
	if o.Weight = v; o.Weight == nil {
		o.nullFields = append(o.nullFields, "Weight")
	}
	return o
}

func (o *Target) SetStatus(v *Status) *Target {
	if o.Status = v; o.Status == nil {
		o.nullFields = append(o.nullFields, "Status")
	}
	return o
}

func (o *Target) SetTags(v []*Tag) *Target {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// endregion

type Runtime struct {
	ID             *string    `json:"id,omitempty"`
	DeploymentID   *string    `json:"deploymentId,omitempty"`
	IPAddr         *string    `json:"ip,omitempty"`
	Version        *string    `json:"version,omitempty"`
	Status         *Status    `json:"status,omitempty"`
	LastReportedAt *time.Time `json:"lastReported,omitempty"`
	Leader         *bool      `json:"isLeader,omitempty"`
	Tags           []*Tag     `json:"tags,omitempty"`
	CreatedAt      *time.Time `json:"createdAt,omitempty"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListRuntimesInput struct {
	DeploymentID *string `json:"deploymentId,omitempty"`
}

type ListRuntimesOutput struct {
	Runtimes []*Runtime `json:"runtimes,omitempty"`
}

type ReadRuntimeInput struct {
	RuntimeID *string `json:"runtimeId,omitempty"`
}

type ReadRuntimeOutput struct {
	Runtime *Runtime `json:"runtime,omitempty"`
}

func runtimeFromJSON(in []byte) (*Runtime, error) {
	rt := new(Runtime)
	if err := json.Unmarshal(in, rt); err != nil {
		return nil, err
	}
	return rt, nil
}

func runtimesFromJSON(in []byte) ([]*Runtime, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Runtime, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rrt := range rw.Response.Items {
		rt, err := runtimeFromJSON(rrt)
		if err != nil {
			return nil, err
		}
		out[i] = rt
	}
	return out, nil
}

func runtimesFromHttpResponse(resp *http.Response) ([]*Runtime, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return runtimesFromJSON(body)
}

func (s *ServiceOp) ListRuntimes(ctx context.Context, input *ListRuntimesInput) (*ListRuntimesOutput, error) {
	r := client.NewRequest(http.MethodGet, "/loadBalancer/runtime")

	if input.DeploymentID != nil {
		r.Params.Set("deploymentId", spotinst.StringValue(input.DeploymentID))
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rts, err := runtimesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListRuntimesOutput{Runtimes: rts}, nil
}

func (s *ServiceOp) ReadRuntime(ctx context.Context, input *ReadRuntimeInput) (*ReadRuntimeOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/runtime/{runtimeId}", uritemplates.Values{
		"runtimeId": spotinst.StringValue(input.RuntimeID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rt, err := runtimesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadRuntimeOutput)
	if len(rt) > 0 {
		output.Runtime = rt[0]
	}

	return output, nil
}

// region Runtime

func (o *Runtime) MarshalJSON() ([]byte, error) {
	type noMethod Runtime
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Runtime) SetId(v *string) *Runtime {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Runtime) SetDeploymentId(v *string) *Runtime {
	if o.DeploymentID = v; o.DeploymentID == nil {
		o.nullFields = append(o.nullFields, "DeploymentID")
	}
	return o
}

func (o *Runtime) SetStatus(v *Status) *Runtime {
	if o.Status = v; o.Status == nil {
		o.nullFields = append(o.nullFields, "Status")
	}
	return o
}

func (o *Runtime) SetTags(v []*Tag) *Runtime {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// endregion

type Deployment struct {
	ID        *string    `json:"id,omitempty"`
	Name      *string    `json:"name,omitempty"`
	Tags      []*Tag     `json:"tags,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListDeploymentsInput struct{}

type ListDeploymentsOutput struct {
	Deployments []*Deployment `json:"deployments,omitempty"`
}

type CreateDeploymentInput struct {
	Deployment *Deployment `json:"deployment,omitempty"`
}

type CreateDeploymentOutput struct {
	Deployment *Deployment `json:"deployment,omitempty"`
}

type ReadDeploymentInput struct {
	DeploymentID *string `json:"deploymentId,omitempty"`
}

type ReadDeploymentOutput struct {
	Deployment *Deployment `json:"deployment,omitempty"`
}

type UpdateDeploymentInput struct {
	Deployment *Deployment `json:"deployment,omitempty"`
}

type UpdateDeploymentOutput struct{}

type DeleteDeploymentInput struct {
	DeploymentID *string `json:"deployment,omitempty"`
}

type DeleteDeploymentOutput struct{}

func deploymentFromJSON(in []byte) (*Deployment, error) {
	b := new(Deployment)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func deploymentsFromJSON(in []byte) ([]*Deployment, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Deployment, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rp := range rw.Response.Items {
		p, err := deploymentFromJSON(rp)
		if err != nil {
			return nil, err
		}
		out[i] = p
	}
	return out, nil
}

func deploymentsFromHttpResponse(resp *http.Response) ([]*Deployment, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return deploymentsFromJSON(body)
}

func (s *ServiceOp) ListDeployments(ctx context.Context, input *ListDeploymentsInput) (*ListDeploymentsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/loadBalancer/deployment")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ds, err := deploymentsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListDeploymentsOutput{Deployments: ds}, nil
}

func (s *ServiceOp) CreateDeployment(ctx context.Context, input *CreateDeploymentInput) (*CreateDeploymentOutput, error) {
	r := client.NewRequest(http.MethodPost, "/loadBalancer/deployment")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ds, err := deploymentsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateDeploymentOutput)
	if len(ds) > 0 {
		output.Deployment = ds[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadDeployment(ctx context.Context, input *ReadDeploymentInput) (*ReadDeploymentOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/deployment/{deploymentId}", uritemplates.Values{
		"deploymentId": spotinst.StringValue(input.DeploymentID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ds, err := deploymentsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadDeploymentOutput)
	if len(ds) > 0 {
		output.Deployment = ds[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateDeployment(ctx context.Context, input *UpdateDeploymentInput) (*UpdateDeploymentOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/deployment/{deploymentId}", uritemplates.Values{
		"deploymentId": spotinst.StringValue(input.Deployment.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Deployment.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &UpdateDeploymentOutput{}, nil
}

func (s *ServiceOp) DeleteDeployment(ctx context.Context, input *DeleteDeploymentInput) (*DeleteDeploymentOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/deployment/{deploymentId}", uritemplates.Values{
		"deploymentId": spotinst.StringValue(input.DeploymentID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteDeploymentOutput{}, nil
}

// region Deployment

func (o *Deployment) MarshalJSON() ([]byte, error) {
	type noMethod Deployment
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Deployment) SetId(v *string) *Deployment {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Deployment) SetName(v *string) *Deployment {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Deployment) SetTags(v []*Tag) *Deployment {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

// endregion

type Certificate struct {
	ID           *string    `json:"id,omitempty"`
	Name         *string    `json:"name,omitempty"`
	CertPEMBlock *string    `json:"certificatePemBlock,omitempty"`
	KeyPEMBlock  *string    `json:"keyPemBlock,omitempty"`
	Tags         []*Tag     `json:"tags,omitempty"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListCertificatesInput struct{}

type ListCertificatesOutput struct {
	Certificates []*Certificate `json:"certificates,omitempty"`
}

type CreateCertificateInput struct {
	Certificate *Certificate `json:"certificate,omitempty"`
}

type CreateCertificateOutput struct {
	Certificate *Certificate `json:"certificate,omitempty"`
}

type ReadCertificateInput struct {
	CertificateID *string `json:"certificateId,omitempty"`
}

type ReadCertificateOutput struct {
	Certificate *Certificate `json:"certificate,omitempty"`
}

type UpdateCertificateInput struct {
	Certificate *Certificate `json:"certificate,omitempty"`
}

type UpdateCertificateOutput struct{}

type DeleteCertificateInput struct {
	CertificateID *string `json:"certificateId,omitempty"`
}

type DeleteCertificateOutput struct{}

func certificateFromJSON(in []byte) (*Certificate, error) {
	b := new(Certificate)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func certificatesFromJSON(in []byte) ([]*Certificate, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Certificate, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rp := range rw.Response.Items {
		p, err := certificateFromJSON(rp)
		if err != nil {
			return nil, err
		}
		out[i] = p
	}
	return out, nil
}

func certificatesFromHttpResponse(resp *http.Response) ([]*Certificate, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return certificatesFromJSON(body)
}

func (s *ServiceOp) ListCertificates(ctx context.Context, input *ListCertificatesInput) (*ListCertificatesOutput, error) {
	r := client.NewRequest(http.MethodGet, "/loadBalancer/certificate")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cs, err := certificatesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListCertificatesOutput{Certificates: cs}, nil
}

func (s *ServiceOp) CreateCertificate(ctx context.Context, input *CreateCertificateInput) (*CreateCertificateOutput, error) {
	r := client.NewRequest(http.MethodPost, "/loadBalancer/certificate")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cs, err := certificatesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateCertificateOutput)
	if len(cs) > 0 {
		output.Certificate = cs[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadCertificate(ctx context.Context, input *ReadCertificateInput) (*ReadCertificateOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/certificate/{certificateId}", uritemplates.Values{
		"certificateId": spotinst.StringValue(input.CertificateID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cs, err := certificatesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadCertificateOutput)
	if len(cs) > 0 {
		output.Certificate = cs[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateCertificate(ctx context.Context, input *UpdateCertificateInput) (*UpdateCertificateOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/certificate/{certificateId}", uritemplates.Values{
		"certificateId": spotinst.StringValue(input.Certificate.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Certificate.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &UpdateCertificateOutput{}, nil
}

func (s *ServiceOp) DeleteCertificate(ctx context.Context, input *DeleteCertificateInput) (*DeleteCertificateOutput, error) {
	path, err := uritemplates.Expand("/loadBalancer/certificate/{certificateId}", uritemplates.Values{
		"certificateId": spotinst.StringValue(input.CertificateID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteCertificateOutput{}, nil
}
