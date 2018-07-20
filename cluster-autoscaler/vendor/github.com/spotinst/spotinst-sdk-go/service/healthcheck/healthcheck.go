package healthcheck

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

type HealthCheck struct {
	ID         *string `json:"id,omitempty"`
	Name       *string `json:"name,omitempty"`
	ResourceID *string `json:"resourceId,omitempty"`
	Check      *Check  `json:"check,omitempty"`
	ProxyAddr  *string `json:"proxyAddress,omitempty"`
	ProxyPort  *int    `json:"proxyPort,omitempty"`

	// forceSendFields is a list of field names (e.g. "Keys") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	forceSendFields []string `json:"-"`

	// nullFields is a list of field names (e.g. "Keys") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	nullFields []string `json:"-"`
}

type Check struct {
	Protocol  *string `json:"protocol,omitempty"`
	Endpoint  *string `json:"endpoint,omitempty"`
	Port      *int    `json:"port,omitempty"`
	Interval  *int    `json:"interval,omitempty"`
	Timeout   *int    `json:"timeout,omitempty"`
	Healthy   *int    `json:"healthyThreshold,omitempty"`
	Unhealthy *int    `json:"unhealthyThreshold,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListHealthChecksInput struct{}

type ListHealthChecksOutput struct {
	HealthChecks []*HealthCheck `json:"healthChecks,omitempty"`
}

type CreateHealthCheckInput struct {
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
}

type CreateHealthCheckOutput struct {
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
}

type ReadHealthCheckInput struct {
	HealthCheckID *string `json:"healthCheckId,omitempty"`
}

type ReadHealthCheckOutput struct {
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
}

type UpdateHealthCheckInput struct {
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
}

type UpdateHealthCheckOutput struct {
	HealthCheck *HealthCheck `json:"healthCheck,omitempty"`
}

type DeleteHealthCheckInput struct {
	HealthCheckID *string `json:"healthCheckId,omitempty"`
}

type DeleteHealthCheckOutput struct{}

func healthCheckFromJSON(in []byte) (*HealthCheck, error) {
	b := new(HealthCheck)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func healthChecksFromJSON(in []byte) ([]*HealthCheck, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*HealthCheck, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := healthCheckFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func healthChecksFromHttpResponse(resp *http.Response) ([]*HealthCheck, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return healthChecksFromJSON(body)
}

func (s *ServiceOp) List(ctx context.Context, input *ListHealthChecksInput) (*ListHealthChecksOutput, error) {
	r := client.NewRequest(http.MethodGet, "/healthCheck")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	hcs, err := healthChecksFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListHealthChecksOutput{HealthChecks: hcs}, nil
}

func (s *ServiceOp) Create(ctx context.Context, input *CreateHealthCheckInput) (*CreateHealthCheckOutput, error) {
	r := client.NewRequest(http.MethodPost, "/healthCheck")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	hcs, err := healthChecksFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateHealthCheckOutput)
	if len(hcs) > 0 {
		output.HealthCheck = hcs[0]
	}

	return output, nil
}

func (s *ServiceOp) Read(ctx context.Context, input *ReadHealthCheckInput) (*ReadHealthCheckOutput, error) {
	path, err := uritemplates.Expand("/healthCheck/{healthCheckId}", uritemplates.Values{
		"healthCheckId": spotinst.StringValue(input.HealthCheckID),
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

	hcs, err := healthChecksFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadHealthCheckOutput)
	if len(hcs) > 0 {
		output.HealthCheck = hcs[0]
	}

	return output, nil
}

func (s *ServiceOp) Update(ctx context.Context, input *UpdateHealthCheckInput) (*UpdateHealthCheckOutput, error) {
	path, err := uritemplates.Expand("/healthCheck/{healthCheckId}", uritemplates.Values{
		"healthCheckId": spotinst.StringValue(input.HealthCheck.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.HealthCheck.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	hcs, err := healthChecksFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateHealthCheckOutput)
	if len(hcs) > 0 {
		output.HealthCheck = hcs[0]
	}

	return output, nil
}

func (s *ServiceOp) Delete(ctx context.Context, input *DeleteHealthCheckInput) (*DeleteHealthCheckOutput, error) {
	path, err := uritemplates.Expand("/healthCheck/{healthCheckId}", uritemplates.Values{
		"healthCheckId": spotinst.StringValue(input.HealthCheckID),
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

	return &DeleteHealthCheckOutput{}, nil
}

// region HealthCheck

func (o *HealthCheck) MarshalJSON() ([]byte, error) {
	type noMethod HealthCheck
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *HealthCheck) SetId(v *string) *HealthCheck {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *HealthCheck) SetName(v *string) *HealthCheck {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *HealthCheck) SetResourceId(v *string) *HealthCheck {
	if o.ResourceID = v; o.ResourceID == nil {
		o.nullFields = append(o.nullFields, "ResourceID")
	}
	return o
}

func (o *HealthCheck) SetCheck(v *Check) *HealthCheck {
	if o.Check = v; o.Check == nil {
		o.nullFields = append(o.nullFields, "Check")
	}
	return o
}

func (o *HealthCheck) SetProxyAddr(v *string) *HealthCheck {
	if o.ProxyAddr = v; o.ProxyAddr == nil {
		o.nullFields = append(o.nullFields, "ProxyAddr")
	}
	return o
}

func (o *HealthCheck) SetProxyPort(v *int) *HealthCheck {
	if o.ProxyPort = v; o.ProxyPort == nil {
		o.nullFields = append(o.nullFields, "ProxyPort")
	}
	return o
}

// endregion

// region Check

func (o *Check) MarshalJSON() ([]byte, error) {
	type noMethod Check
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Check) SetProtocol(v *string) *Check {
	if o.Protocol = v; o.Protocol == nil {
		o.nullFields = append(o.nullFields, "Protocol")
	}
	return o
}

func (o *Check) SetEndpoint(v *string) *Check {
	if o.Endpoint = v; o.Endpoint == nil {
		o.nullFields = append(o.nullFields, "Endpoint")
	}
	return o
}

func (o *Check) SetPort(v *int) *Check {
	if o.Port = v; o.Port == nil {
		o.nullFields = append(o.nullFields, "Port")
	}
	return o
}

func (o *Check) SetInterval(v *int) *Check {
	if o.Interval = v; o.Interval == nil {
		o.nullFields = append(o.nullFields, "Interval")
	}
	return o
}

func (o *Check) SetTimeout(v *int) *Check {
	if o.Timeout = v; o.Timeout == nil {
		o.nullFields = append(o.nullFields, "Timeout")
	}
	return o
}

func (o *Check) SetHealthy(v *int) *Check {
	if o.Healthy = v; o.Healthy == nil {
		o.nullFields = append(o.nullFields, "Healthy")
	}
	return o
}

func (o *Check) SetUnhealthy(v *int) *Check {
	if o.Unhealthy = v; o.Unhealthy == nil {
		o.nullFields = append(o.nullFields, "Unhealthy")
	}
	return o
}

// endregion
