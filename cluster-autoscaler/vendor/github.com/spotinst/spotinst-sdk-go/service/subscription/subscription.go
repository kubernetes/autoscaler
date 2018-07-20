package subscription

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

type Subscription struct {
	ID         *string                `json:"id,omitempty"`
	ResourceID *string                `json:"resourceId,omitempty"`
	EventType  *string                `json:"eventType,omitempty"`
	Protocol   *string                `json:"protocol,omitempty"`
	Endpoint   *string                `json:"endpoint,omitempty"`
	Format     map[string]interface{} `json:"eventFormat,omitempty"`

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

type ListSubscriptionsInput struct{}

type ListSubscriptionsOutput struct {
	Subscriptions []*Subscription `json:"subscriptions,omitempty"`
}

type CreateSubscriptionInput struct {
	Subscription *Subscription `json:"subscription,omitempty"`
}

type CreateSubscriptionOutput struct {
	Subscription *Subscription `json:"subscription,omitempty"`
}

type ReadSubscriptionInput struct {
	SubscriptionID *string `json:"subscriptionId,omitempty"`
}

type ReadSubscriptionOutput struct {
	Subscription *Subscription `json:"subscription,omitempty"`
}

type UpdateSubscriptionInput struct {
	Subscription *Subscription `json:"subscription,omitempty"`
}

type UpdateSubscriptionOutput struct {
	Subscription *Subscription `json:"subscription,omitempty"`
}

type DeleteSubscriptionInput struct {
	SubscriptionID *string `json:"subscriptionId,omitempty"`
}

type DeleteSubscriptionOutput struct{}

func subscriptionFromJSON(in []byte) (*Subscription, error) {
	b := new(Subscription)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func subscriptionsFromJSON(in []byte) ([]*Subscription, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Subscription, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := subscriptionFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func subscriptionsFromHttpResponse(resp *http.Response) ([]*Subscription, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return subscriptionsFromJSON(body)
}

func (s *ServiceOp) List(ctx context.Context, input *ListSubscriptionsInput) (*ListSubscriptionsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/events/subscription")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := subscriptionsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListSubscriptionsOutput{Subscriptions: gs}, nil
}

func (s *ServiceOp) Create(ctx context.Context, input *CreateSubscriptionInput) (*CreateSubscriptionOutput, error) {
	r := client.NewRequest(http.MethodPost, "/events/subscription")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ss, err := subscriptionsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateSubscriptionOutput)
	if len(ss) > 0 {
		output.Subscription = ss[0]
	}

	return output, nil
}

func (s *ServiceOp) Read(ctx context.Context, input *ReadSubscriptionInput) (*ReadSubscriptionOutput, error) {
	path, err := uritemplates.Expand("/events/subscription/{subscriptionId}", uritemplates.Values{
		"subscriptionId": spotinst.StringValue(input.SubscriptionID),
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

	ss, err := subscriptionsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadSubscriptionOutput)
	if len(ss) > 0 {
		output.Subscription = ss[0]
	}

	return output, nil
}

func (s *ServiceOp) Update(ctx context.Context, input *UpdateSubscriptionInput) (*UpdateSubscriptionOutput, error) {
	path, err := uritemplates.Expand("/events/subscription/{subscriptionId}", uritemplates.Values{
		"subscriptionId": spotinst.StringValue(input.Subscription.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Subscription.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ss, err := subscriptionsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateSubscriptionOutput)
	if len(ss) > 0 {
		output.Subscription = ss[0]
	}

	return output, nil
}

func (s *ServiceOp) Delete(ctx context.Context, input *DeleteSubscriptionInput) (*DeleteSubscriptionOutput, error) {
	path, err := uritemplates.Expand("/events/subscription/{subscriptionId}", uritemplates.Values{
		"subscriptionId": spotinst.StringValue(input.SubscriptionID),
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

	return &DeleteSubscriptionOutput{}, nil
}

// region Subscription

func (o *Subscription) MarshalJSON() ([]byte, error) {
	type noMethod Subscription
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Subscription) SetId(v *string) *Subscription {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Subscription) SetResourceId(v *string) *Subscription {
	if o.ResourceID = v; o.ResourceID == nil {
		o.nullFields = append(o.nullFields, "ResourceID")
	}
	return o
}

func (o *Subscription) SetEventType(v *string) *Subscription {
	if o.EventType = v; o.EventType == nil {
		o.nullFields = append(o.nullFields, "EventType")
	}
	return o
}

func (o *Subscription) SetProtocol(v *string) *Subscription {
	if o.Protocol = v; o.Protocol == nil {
		o.nullFields = append(o.nullFields, "Protocol")
	}
	return o
}

func (o *Subscription) SetEndpoint(v *string) *Subscription {
	if o.Endpoint = v; o.Endpoint == nil {
		o.nullFields = append(o.nullFields, "Endpoint")
	}
	return o
}

func (o *Subscription) SetFormat(v map[string]interface{}) *Subscription {
	if o.Format = v; o.Format == nil {
		o.nullFields = append(o.nullFields, "Format")
	}
	return o
}

// endregion
