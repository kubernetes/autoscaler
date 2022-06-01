// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var _ CloudWatcherService = (*cloudwatcherService)(nil)

type cloudwatcherService struct {
	client *Client
}

// CloudWatcherService is interface wrap others resource's interfaces
type CloudWatcherService interface {
	Agents() *agents
	Alarms() *alarms
	Histories() *histories
	Receivers() *receivers
	Secrets() *secrets
}

func (cws *cloudwatcherService) Agents() *agents {
	return &agents{client: cws.client}
}

func (cws *cloudwatcherService) Alarms() *alarms {
	return &alarms{client: cws.client}
}

func (cws *cloudwatcherService) Receivers() *receivers {
	return &receivers{client: cws.client}
}

func (cws *cloudwatcherService) Histories() *histories {
	return &histories{client: cws.client}
}

func (cws *cloudwatcherService) Secrets() *secrets {
	return &secrets{client: cws.client}
}

const (
	agentsResourcePath    = "/agents"
	alarmsResourcePath    = "/alarms"
	getVerificationPath   = "/resend"
	historiesResourcePath = "/histories"
	receiversResourcePath = "/receivers"
	secretsResourcePath   = "/secrets"
)

// Comparison is represents comparison payload in alarms
type Comparison struct {
	CompareType string  `json:"compare_type"`
	Measurement string  `json:"measurement"`
	RangeTime   int     `json:"range_time"`
	Value       float64 `json:"value"`
}

// AlarmInstancesMonitors is represents instances payload - which servers will be monitored
type AlarmInstancesMonitors struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AlarmVolumesMonitor is represents volumes payload - which volumes will be monitored
type AlarmVolumesMonitor struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Device string `json:"device,omitempty"`
}

// HTTPHeaders is is represents http headers - which using call to http_url
type HTTPHeaders struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AlarmLoadBalancersMonitor is represents load balancer payload - which load balancer will be monitored
type AlarmLoadBalancersMonitor struct {
	LoadBalancerID   string `json:"load_balancer_id"`
	LoadBalancerName string `json:"load_balancer_name"`
	TargetID         string `json:"target_id"`
	TargetName       string `json:"target_name,omitempty"`
	TargetType       string `json:"target_type"`
}

// AlarmReceiversUse is represents receiver's payload will be use create alarms
type AlarmReceiversUse struct {
	AutoscaleClusterName string `json:"autoscale_cluster_name,omitempty"`
	EmailAddress         string `json:"email_address,omitempty"`
	Name                 string `json:"name"`
	ReceiverID           string `json:"receiver_id"`
	SlackChannelName     string `json:"slack_channel_name,omitempty"`
	SMSInterval          int    `json:"sms_interval,omitempty"`
	SMSNumber            string `json:"sms_number,omitempty"`
	TelegramChatID       string `json:"telegram_chat_id,omitempty"`
	WebhookURL           string `json:"webhook_url,omitempty"`
}

type agents struct {
	client *Client
}

type alarms struct {
	client *Client
}

type receivers struct {
	client *Client
}

type histories struct {
	client *Client
}

type secrets struct {
	client *Client
}

// AlarmCreateRequest represents create new alarm request payload.
type AlarmCreateRequest struct {
	AlertInterval    int                          `json:"alert_interval"`
	ClusterID        string                       `json:"cluster_id,omitempty"`
	ClusterName      string                       `json:"cluster_name,omitempty"`
	Comparison       *Comparison                  `json:"comparison,omitempty"`
	Hostname         string                       `json:"hostname,omitempty"`
	HTTPExpectedCode int                          `json:"http_expected_code,omitempty"`
	HTTPHeaders      *[]HTTPHeaders               `json:"http_headers,omitempty"`
	HTTPURL          string                       `json:"http_url,omitempty"`
	Instances        *[]AlarmInstancesMonitors    `json:"instances,omitempty"`
	LoadBalancers    []*AlarmLoadBalancersMonitor `json:"load_balancers,omitempty"`
	Name             string                       `json:"name"`
	Receivers        []AlarmReceiversUse          `json:"receivers"`
	ResourceType     string                       `json:"resource_type"`
	Volumes          *[]AlarmVolumesMonitor       `json:"volumes,omitempty"`
}

// ReceiverCreateRequest contains receiver information.
type ReceiverCreateRequest struct {
	AutoScale      *AutoScalingWebhook `json:"autoscale,omitempty"`
	EmailAddress   string              `json:"email_address,omitempty"`
	Name           string              `json:"name"`
	Slack          *Slack              `json:"slack,omitempty"`
	SMSNumber      string              `json:"sms_number,omitempty"`
	TelegramChatID string              `json:"telegram_chat_id,omitempty"`
	WebhookURL     string              `json:"webhook_url,omitempty"`
}

// SecretsCreateRequest contains receiver information.
type SecretsCreateRequest struct {
	Name string `json:"name,omitempty"`
}

// AlarmUpdateRequest represents update alarm request payload.
type AlarmUpdateRequest struct {
	AlertInterval    int                          `json:"alert_interval,omitempty"`
	ClusterID        string                       `json:"cluster_id,omitempty"`
	ClusterName      string                       `json:"cluster_name,omitempty"`
	Comparison       *Comparison                  `json:"comparison,omitempty"`
	Enable           bool                         `json:"enable"`
	Hostname         string                       `json:"hostname,omitempty"`
	HTTPExpectedCode int                          `json:"http_expected_code,omitempty"`
	HTTPHeaders      *[]HTTPHeaders               `json:"http_headers,omitempty"`
	HTTPURL          string                       `json:"http_url,omitempty"`
	Instances        *[]AlarmInstancesMonitors    `json:"instances,omitempty"`
	LoadBalancers    []*AlarmLoadBalancersMonitor `json:"load_balancers,omitempty"`
	Name             string                       `json:"name,omitempty"`
	Receivers        *[]AlarmReceiversUse         `json:"receivers,omitempty"`
	ResourceType     string                       `json:"resource_type,omitempty"`
	Volumes          *[]AlarmVolumesMonitor       `json:"volumes,omitempty"`
}

// ResponseRequest represents api's response.
type ResponseRequest struct {
	Created string `json:"_created,omitempty"`
	Deleted bool   `json:"_deleted,omitempty"`
	ID      string `json:"_id,omitempty"`
	Status  string `json:"_status,omitempty"`
}

// Agents contains agent information.
type Agents struct {
	Created   string `json:"_created"`
	Hostname  string `json:"hostname"`
	ID        string `json:"_id"`
	Name      string `json:"name"`
	Online    bool   `json:"online"`
	ProjectID string `json:"project_id"`
	Runtime   string `json:"runtime"`
	UserID    string `json:"user_id"`
}

// Alarms contains alarm information.
type Alarms struct {
	AlertInterval    int                          `json:"alert_interval"`
	ClusterID        string                       `json:"cluster_id,omitempty"`
	ClusterName      string                       `json:"cluster_name,omitempty"`
	Comparison       *Comparison                  `json:"comparison,omitempty"`
	Created          string                       `json:"_created"`
	Creator          string                       `json:"creator"`
	Enable           bool                         `json:"enable"`
	Hostname         string                       `json:"hostname,omitempty"`
	HTTPExpectedCode int                          `json:"http_expected_code,omitempty"`
	HTTPHeaders      []HTTPHeaders                `json:"http_headers,omitempty"`
	HTTPURL          string                       `json:"http_url,omitempty"`
	ID               string                       `json:"_id"`
	Instances        []AlarmInstancesMonitors     `json:"instances,omitempty"`
	LoadBalancers    []*AlarmLoadBalancersMonitor `json:"load_balancers,omitempty"`
	Name             string                       `json:"name"`
	ProjectID        string                       `json:"project_id"`
	Receivers        []AlarmReceiversUse          `json:"receivers"`
	ResourceType     string                       `json:"resource_type"`
	UserID           string                       `json:"user_id"`
	Volumes          []AlarmVolumesMonitor        `json:"volumes,omitempty"`
}

// AlarmsInHistories contains alarm information in a history
type AlarmsInHistories struct {
	AlertInterval    int                          `json:"alert_interval"`
	ClusterID        string                       `json:"cluster_id,omitempty"`
	ClusterName      string                       `json:"cluster_name,omitempty"`
	Comparison       *Comparison                  `json:"comparison,omitempty"`
	Enable           bool                         `json:"enable"`
	Hostname         string                       `json:"hostname,omitempty"`
	HTTPExpectedCode int                          `json:"http_expected_code,omitempty"`
	HTTPHeaders      *[]HTTPHeaders               `json:"http_headers,omitempty"`
	HTTPURL          string                       `json:"http_url,omitempty"`
	ID               string                       `json:"_id"`
	Instances        *[]AlarmInstancesMonitors    `json:"instances,omitempty"`
	LoadBalancers    *[]AlarmLoadBalancersMonitor `json:"load_balancers,omitempty"`
	Name             string                       `json:"name"`
	ProjectID        string                       `json:"project_id"`
	Receivers        *[]struct {
		ReceiverID string   `json:"receiver_id"`
		Methods    []string `json:"methods"`
	} `json:"receivers"`
	ResourceType string                 `json:"resource_type"`
	UserID       string                 `json:"user_id"`
	Volumes      *[]AlarmVolumesMonitor `json:"volumes,omitempty"`
}

// Slack is represents slack payload - which will be use create a receiver
type Slack struct {
	SlackChannelName string `json:"channel_name"`
	WebhookURL       string `json:"webhook_url"`
}

// Receivers contains receiver information.
type Receivers struct {
	AutoScale              *AutoScalingWebhook `json:"autoscale,omitempty"`
	Created                string              `json:"_created"`
	Creator                string              `json:"creator"`
	EmailAddress           string              `json:"email_address,omitempty"`
	Name                   string              `json:"name"`
	ProjectID              string              `json:"project_id,omitempty"`
	ReceiverID             string              `json:"_id"`
	Slack                  *Slack              `json:"slack,omitempty"`
	SMSNumber              string              `json:"sms_number,omitempty"`
	TelegramChatID         string              `json:"telegram_chat_id,omitempty"`
	UserID                 string              `json:"user_id,omitempty"`
	VerifiedEmailDddress   bool                `json:"verified_email_address,omitempty"`
	VerifiedSMSNumber      bool                `json:"verified_sms_number,omitempty"`
	VerifiedTelegramChatID bool                `json:"verified_telegram_chat_id,omitempty"`
	VerifiedWebhookURL     bool                `json:"verified_webhook_url,omitempty"`
	WebhookURL             string              `json:"webhook_url,omitempty"`
}

// Histories contains history information.
type Histories struct {
	HistoryID   string            `json:"_id"`
	Name        string            `json:"name"`
	ProjectID   string            `json:"project_id,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	Resource    interface{}       `json:"resource,omitempty"`
	State       string            `json:"state,omitempty"`
	Measurement string            `json:"measurement,omitempty"`
	AlarmID     string            `json:"alarm_id"`
	Alarm       AlarmsInHistories `json:"alarm,omitempty"`
	Created     string            `json:"_created,omitempty"`
}

// SecretCreateRequest represents create new secret request payload.
type SecretCreateRequest struct {
	Name string `json:"name,omitempty"`
}

// Secrets contains secrets information.
type Secrets struct {
	Created   string `json:"_created,omitempty"`
	ID        string `json:"_id"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id,omitempty"`
	Secret    string `json:"secret,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

// ===========================================================================
// Start block interaction for agents
// ===========================================================================
func (a *agents) resourcePath() string {
	return strings.Join([]string{agentsResourcePath}, "/")
}

func (a *agents) itemPath(id string) string {
	return strings.Join([]string{agentsResourcePath, id}, "/")
}

// List agents
func (a *agents) List(ctx context.Context, filters *string) ([]*Agents, error) {
	req, err := a.client.NewRequest(ctx, http.MethodGet, cloudwatcherServiceName, a.resourcePath(), nil)
	if err != nil {
		return nil, err
	}

	if filters != nil {
		q := req.URL.Query()
		q.Add("where", *filters)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Agents []*Agents `json:"_items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Agents, nil
}

// Get an agent
func (a *agents) Get(ctx context.Context, id string) (*Agents, error) {
	req, err := a.client.NewRequest(ctx, http.MethodGet, cloudwatcherServiceName, a.itemPath(id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	agent := &Agents{}
	if err := json.NewDecoder(resp.Body).Decode(agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// Delete an agent
func (a *agents) Delete(ctx context.Context, id string) error {
	req, err := a.client.NewRequest(ctx, http.MethodDelete, cloudwatcherServiceName, a.itemPath(id), nil)
	if err != nil {
		return err
	}
	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	return resp.Body.Close()
}

// ===========================================================================
// Start block interaction for alarms
// ===========================================================================
func (a *alarms) resourcePath() string {
	return strings.Join([]string{alarmsResourcePath}, "/")
}

func (a *alarms) itemPath(id string) string {
	return strings.Join([]string{alarmsResourcePath, id}, "/")
}

// List alarms
func (a *alarms) List(ctx context.Context, filters *string) ([]*Alarms, error) {
	req, err := a.client.NewRequest(ctx, http.MethodGet, cloudwatcherServiceName, a.resourcePath(), nil)
	if err != nil {
		return nil, err
	}

	if filters != nil {
		q := req.URL.Query()
		q.Add("where", *filters)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Alarms []*Alarms `json:"_items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Alarms, nil
}

// Create an alarm
func (a *alarms) Create(ctx context.Context, acr *AlarmCreateRequest) (*ResponseRequest, error) {
	req, err := a.client.NewRequest(ctx, http.MethodPost, cloudwatcherServiceName, a.resourcePath(), &acr)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respData = &ResponseRequest{}
	if err := json.NewDecoder(resp.Body).Decode(respData); err != nil {
		return nil, err
	}
	return respData, nil
}

// Get an alarm
func (a *alarms) Get(ctx context.Context, id string) (*Alarms, error) {
	req, err := a.client.NewRequest(ctx, http.MethodGet, cloudwatcherServiceName, a.itemPath(id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	alarm := &Alarms{}
	if err := json.NewDecoder(resp.Body).Decode(alarm); err != nil {
		return nil, err
	}
	// hardcode in here
	for _, loadbalancer := range alarm.LoadBalancers {
		if loadbalancer.TargetType == "frontend" {
			frontend, err := a.client.Listener.Get(ctx, loadbalancer.TargetID)
			if err != nil {
				loadbalancer.TargetName = ""
			}
			loadbalancer.TargetName = frontend.Name
		} else {
			backend, err := a.client.Pool.Get(ctx, loadbalancer.TargetID)
			if err != nil {
				loadbalancer.TargetName = ""
			}
			loadbalancer.TargetName = backend.Name
		}
	}
	return alarm, nil
}

// Update an alarm
func (a *alarms) Update(ctx context.Context, id string, aur *AlarmUpdateRequest) (*ResponseRequest, error) {
	req, err := a.client.NewRequest(ctx, http.MethodPatch, cloudwatcherServiceName, a.itemPath(id), &aur)
	if err != nil {
		return nil, err
	}
	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData := &ResponseRequest{}
	if err := json.NewDecoder(resp.Body).Decode(respData); err != nil {
		return nil, err
	}

	return respData, nil
}

// Delete an alarm
func (a *alarms) Delete(ctx context.Context, id string) error {
	req, err := a.client.NewRequest(ctx, http.MethodDelete, cloudwatcherServiceName, a.itemPath(id), nil)
	if err != nil {
		return err
	}
	resp, err := a.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	return resp.Body.Close()
}

// ===========================================================================
// Start block interaction for receivers
// ===========================================================================
func (r *receivers) resourcePath() string {
	return strings.Join([]string{receiversResourcePath}, "/")
}

func (r *receivers) itemPath(id string) string {
	return strings.Join([]string{receiversResourcePath, id}, "/")
}

func (r *receivers) verificationPath() string {
	return strings.Join([]string{getVerificationPath}, "/")
}

// List receivers
func (r *receivers) List(ctx context.Context, filters *string) ([]*Receivers, error) {
	req, err := r.client.NewRequest(ctx, http.MethodGet, cloudwatcherServiceName, r.resourcePath(), nil)
	if err != nil {
		return nil, err
	}

	if filters != nil {
		q := req.URL.Query()
		q.Add("where", *filters)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := r.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Receivers []*Receivers `json:"_items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Receivers, nil
}

// Create a receiver
func (r *receivers) Create(ctx context.Context, rcr *ReceiverCreateRequest) (*ResponseRequest, error) {
	req, err := r.client.NewRequest(ctx, http.MethodPost, cloudwatcherServiceName, r.resourcePath(), &rcr)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respData = &ResponseRequest{}
	if err := json.NewDecoder(resp.Body).Decode(respData); err != nil {
		return nil, err
	}
	return respData, nil
}

// Get a receiver
func (r *receivers) Get(ctx context.Context, id string) (*Receivers, error) {
	req, err := r.client.NewRequest(ctx, http.MethodGet, cloudwatcherServiceName, r.itemPath(id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	receiver := &Receivers{}
	if err := json.NewDecoder(resp.Body).Decode(receiver); err != nil {
		return nil, err
	}
	return receiver, nil
}

// Update receiver
func (r *receivers) Update(ctx context.Context, id string, rur *ReceiverCreateRequest) (*ResponseRequest, error) {
	req, err := r.client.NewRequest(ctx, http.MethodPut, cloudwatcherServiceName, r.itemPath(id), &rur)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData := &ResponseRequest{}
	if err := json.NewDecoder(resp.Body).Decode(respData); err != nil {
		return nil, err
	}

	return respData, nil
}

// Delete receiver
func (r *receivers) Delete(ctx context.Context, id string) error {
	req, err := r.client.NewRequest(ctx, http.MethodDelete, cloudwatcherServiceName, r.itemPath(id), nil)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	return resp.Body.Close()
}

// ResendVerificationLink is use get a link verification
func (r *receivers) ResendVerificationLink(ctx context.Context, id string, rType string) error {
	req, err := r.client.NewRequest(ctx, http.MethodGet, cloudwatcherServiceName, r.verificationPath(), nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("rec_id", id)
	q.Add("rec_type", rType)
	req.URL.RawQuery = q.Encode()

	resp, err := r.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	return resp.Body.Close()
}

// ===========================================================================
// Start block interaction for histories
// ===========================================================================
func (h *histories) resourcePath() string {
	return strings.Join([]string{historiesResourcePath}, "/")
}

// List histories
func (h *histories) List(ctx context.Context, filters *string) ([]*Histories, error) {
	req, err := h.client.NewRequest(ctx, http.MethodGet, cloudwatcherServiceName, h.resourcePath(), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if filters != nil {
		q.Add("where", *filters)
	}
	q.Add("sort", "-_created")
	req.URL.RawQuery = q.Encode()

	resp, err := h.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Histories []*Histories `json:"_items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Histories, nil
}

// ===========================================================================
// Start block interaction for secrets
// ===========================================================================
func (s *secrets) resourcePath() string {
	return strings.Join([]string{secretsResourcePath}, "/")
}

func (s *secrets) itemPath(id string) string {
	return strings.Join([]string{secretsResourcePath, id}, "/")
}

// List secrets
func (s *secrets) List(ctx context.Context, filters *string) ([]*Secrets, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, cloudwatcherServiceName, s.resourcePath(), nil)
	if err != nil {
		return nil, err
	}

	if filters != nil {
		q := req.URL.Query()
		q.Add("where", *filters)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Secrets []*Secrets `json:"_items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Secrets, nil
}

// Create a secret
func (s *secrets) Create(ctx context.Context, scr *SecretsCreateRequest) (*ResponseRequest, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, cloudwatcherServiceName, s.resourcePath(), &scr)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respData = &ResponseRequest{}
	if err := json.NewDecoder(resp.Body).Decode(respData); err != nil {
		return nil, err
	}
	return respData, nil
}

// Get a secret
func (s *secrets) Get(ctx context.Context, id string) (*Secrets, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, cloudwatcherServiceName, s.itemPath(id), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	secret := &Secrets{}
	if err := json.NewDecoder(resp.Body).Decode(secret); err != nil {
		return nil, err
	}
	return secret, nil
}

// Delete secret
func (s *secrets) Delete(ctx context.Context, id string) error {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, cloudwatcherServiceName, s.itemPath(id), nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	return resp.Body.Close()
}
