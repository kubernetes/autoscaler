// This file is part of gobizfly

package gobizfly

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	statusActive = "Active"
	statusError  = "Error"
)

const (
	autoscalingGroupResourcePath     = "/groups"
	eventsResourcePath               = "/events"
	launchConfigurationsResourcePath = "/launch_configs"
	nodesResourcePath                = "nodes"
	policiesResourcePath             = "policies"
	quotasResourcePath               = "/quotas"
	schedulesResourcePath            = "cron_triggers"
	suggestionResourcePath           = "/suggestion"
	tasksResourcePath                = "/task"
	usingResourcePath                = "/using_resource"
	webhooksResourcePath             = "webhooks"
)

var (
	actionTypeSupportedWebhooks = []string{
		"CLUSTER SCALE IN",
		"CLUSTER SCALE OUT",
	}
	networkPlan = []string{
		"free_datatransfer",
		"free_bandwidth",
	}
)

var _ AutoScalingService = (*autoscalingService)(nil)

type autoscalingService struct {
	client *Client
}

/*
AutoScalingService is interface wrap others resource's interfaces. Includes:
1. AutoScalingGroups: Provides function interact with an autoscaling group such as:
    Create, Update, Delete
2. Events: Provides function to list events of an autoscaling group
3. LaunchConfigurations: Provides function to interact with launch configurations
4. Nodes: Provides function to interact with members of autoscaling group
5. Policies: Provides function to interact with autoscaling policies of autoscaling group
6. Schedules: Provides function to interact with schedule for auto scaling
7. Tasks: Provides function to get information of task
8. Webhooks: Provides fucntion to list webhook triggers scale of autoscaling group
*/
type AutoScalingService interface {
	AutoScalingGroups() *autoScalingGroup
	Common() *common
	Events() *event
	LaunchConfigurations() *launchConfiguration
	Nodes() *node
	Policies() *policy
	Schedules() *schedule
	Tasks() *task
	Webhooks() *webhook
}

func (as *autoscalingService) AutoScalingGroups() *autoScalingGroup {
	return &autoScalingGroup{client: as.client}
}

func (as *autoscalingService) LaunchConfigurations() *launchConfiguration {
	return &launchConfiguration{client: as.client}
}

func (as *autoscalingService) Webhooks() *webhook {
	return &webhook{client: as.client}
}

func (as *autoscalingService) Events() *event {
	return &event{client: as.client}
}

func (as *autoscalingService) Nodes() *node {
	return &node{client: as.client}
}

func (as *autoscalingService) Policies() *policy {
	return &policy{client: as.client}
}

func (as *autoscalingService) Schedules() *schedule {
	return &schedule{client: as.client}
}

func (as *autoscalingService) Tasks() *task {
	return &task{client: as.client}
}

func (as *autoscalingService) Common() *common {
	return &common{client: as.client}
}

type autoScalingGroup struct {
	client *Client
}

type launchConfiguration struct {
	client *Client
}

type webhook struct {
	client *Client
}

type event struct {
	client *Client
}

type policy struct {
	client *Client
}

type node struct {
	client *Client
}

type schedule struct {
	client *Client
}

type task struct {
	client *Client
}

type common struct {
	client *Client
}

// Webhook - informaion of cluster's receiver
type Webhook struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ASWebhookIDs - information about cluster's receivers
type ASWebhookIDs struct {
	ScaleIn  Webhook `json:"scale_in"`
	ScaleOut Webhook `json:"scale_out"`
}

// ASAlarm - alarm to triggers scale
type ASAlarm struct {
	AutoScaling  []string `json:"auto_scaling"`
	LoadBalancer []string `json:"load_balancer"`
}

// ASAlarms - alarms to do trigger scale
type ASAlarms struct {
	ScaleIn  ASAlarm `json:"scale_in"`
	ScaleOut ASAlarm `json:"scale_out"`
}

// taskResult - Struct of data was returned by workers
type taskResult struct {
	Action   string      `json:"action,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Message  string      `json:"message,omitempty"`
	Progress int         `json:"progress,omitempty"`
	Success  bool        `json:"success,omitempty"`
}

// ASTask is information of doing task
type ASTask struct {
	Ready  bool       `json:"ready"`
	Result taskResult `json:"result"`
}

// ASMetadata - Medata of an auto scaling group
type ASMetadata struct {
	Alarms           ASAlarms     `json:"alarms"`
	DeletionPolicy   []string     `json:"deletion_policy"`
	ScaleInReceiver  string       `json:"scale_in_receiver"`
	ScaleOutReceiver string       `json:"scale_out_receiver"`
	WebhookIDs       ASWebhookIDs `json:"webhook_ids"`
}

// ScalePolicyInformation - information about a scale policy
type ScalePolicyInformation struct {
	BestEffort bool    `json:"best_effort"`
	CoolDown   int     `json:"cooldown"`
	Event      string  `json:"event,omitempty"`
	ID         *string `json:"id,omitempty"`
	MetricType string  `json:"metric"`
	RangeTime  int     `json:"range_time"`
	ScaleSize  int     `json:"number"`
	Threshold  float64 `json:"threshold"`
	Type       *string `json:"type,omitempty"`
}

// DeleteionPolicyHooks represents a hooks was triggered when delete node in auto scaling group
type DeleteionPolicyHooks struct {
	URL     string                  `json:"url"`
	Method  *string                 `json:"method"`
	Params  *map[string]interface{} `json:"params"`
	Headers *map[string]interface{} `json:"headers"`
	Verify  *bool                   `json:"verify"`
}

// LoadBalancerScalingPolicy represents payload load balancers in LoadBalancersPolicyCreateRequest/Update
type LoadBalancerScalingPolicy struct {
	ID         string `json:"load_balancer_id"`
	Name       string `json:"load_balancer_name,omitempty"`
	TargetID   string `json:"target_id"`
	TargetName string `json:"target_name,omitempty"`
	TargetType string `json:"target_type"`
}

// LoadBalancersPolicyCreateRequest represents payload in create a load balancer policy
type LoadBalancersPolicyCreateRequest struct {
	CoolDown      int                       `json:"cooldown,omitempty"`
	Event         string                    `json:"event,omitempty"`
	LoadBalancers LoadBalancerScalingPolicy `json:"load_balancers,omitempty"`
	MetricType    string                    `json:"metric,omitempty"`
	RangeTime     int                       `json:"range_time,omitempty"`
	ScaleSize     int                       `json:"number,omitempty"`
	Threshold     int                       `json:"threshold,omitempty"`
}

// LoadBalancersPolicyUpdateRequest represents payload in create a load balancer policy
type LoadBalancersPolicyUpdateRequest struct {
	CoolDown      int                       `json:"cooldown,omitempty"`
	Event         string                    `json:"event,omitempty"`
	LoadBalancers LoadBalancerScalingPolicy `json:"load_balancers,omitempty"`
	MetricType    string                    `json:"metric,omitempty"`
	RangeTime     int                       `json:"range_time,omitempty"`
	ScaleSize     int                       `json:"number,omitempty"`
	Threshold     int                       `json:"threshold,omitempty"`
}

// HooksPolicyCreateRequest represents payload use create a deletion policy
type HooksPolicyCreateRequest struct {
	Criteria              string               `json:"criteria,omitempty"`
	DestroyAfterDeletion  bool                 `json:"destroy_after_deletion,omitempty"`
	GracePeriod           int                  `json:"grace_period,omitempty"`
	Hooks                 DeleteionPolicyHooks `json:"hooks,omitempty"`
	ReduceDesiredCapacity bool                 `json:"reduce_desired_capacity,omitempty"`
	Type                  string               `json:"policy_type"`
}

// PolicyCreateRequest represents payload use create a  policy
type PolicyCreateRequest struct {
	CoolDown   int    `json:"cooldown,omitempty"`
	Event      string `json:"event,omitempty"`
	MetricType string `json:"metric,omitempty"`
	RangeTime  int    `json:"range_time,omitempty"`
	ScaleSize  int    `json:"number,omitempty"`
	Threshold  int    `json:"threshold,omitempty"`
}

// PolicyUpdateRequest represents payload use update a  policy
type PolicyUpdateRequest struct {
	CoolDown   int    `json:"cooldown,omitempty"`
	Event      string `json:"event,omitempty"`
	MetricType string `json:"metric,omitempty"`
	RangeTime  int    `json:"range_time,omitempty"`
	ScaleSize  int    `json:"number,omitempty"`
	Threshold  int    `json:"threshold,omitempty"`
}

// TaskResponses is responses
type TaskResponses struct {
	Message string `json:"message"`
	TaskID  string `json:"task_id,omitempty"`
}

// DeletionInformation - represents a deletion policy
type DeletionInformation struct {
	ID    string               `json:"id"`
	Hooks DeleteionPolicyHooks `json:"hooks"`
}

// LoadBalancerPolicyInformation - information of load balancer will be use for auto scaling group
type LoadBalancerPolicyInformation struct {
	LoadBalancerID   string `json:"id"`
	LoadBalancerName string `json:"name,omitempty"`
	ServerGroupID    string `json:"server_group_id"`
	ServerGroupName  string `json:"server_group_name,omitempty"`
	ServerGroupPort  int    `json:"server_group_port"`
}

// AutoScalingPolicies - information of policies using for auto scaling group
type AutoScalingPolicies struct {
	ScaleInPolicyInformations      []ScalePolicyInformation      `json:"scale_in_info,omitempty"`
	ScaleOutPolicyInformations     []ScalePolicyInformation      `json:"scale_out_info,omitempty"`
	LoadBalancerPolicyInformations LoadBalancerPolicyInformation `json:"load_balancer_info,omitempty"`
	DeletionPolicyInformations     DeletionInformation           `json:"deletion_info,omitempty"`
	DoingTasks                     []string                      `json:"doing_task,omitempty"`
}

// AutoScalingGroupCreateRequest - payload use to create auto scaling group
type AutoScalingGroupCreateRequest struct {
	DeletionPolicyInformations     *DeletionInformation           `json:"deletion_info,omitempty"`
	DesiredCapacity                int                            `json:"desired_capacity"`
	LoadBalancerPolicyInformations *LoadBalancerPolicyInformation `json:"load_balancer_info,omitempty"`
	MaxSize                        int                            `json:"max_size"`
	MinSize                        int                            `json:"min_size"`
	Name                           string                         `json:"name"`
	ProfileID                      string                         `json:"profile_id"`
	ScaleInPolicyInformations      *[]ScalePolicyInformation      `json:"scale_in_info,omitempty"`
	ScaleOutPolicyInformations     *[]ScalePolicyInformation      `json:"scale_out_info,omitempty"`
}

// AutoScalingGroupUpdateRequest - payload use to update auto scaling group
type AutoScalingGroupUpdateRequest struct {
	DesiredCapacity                int                            `json:"desired_capacity"`
	LoadBalancerPolicyInformations *LoadBalancerPolicyInformation `json:"load_balancer_info,omitempty"`
	MaxSize                        int                            `json:"max_size"`
	MinSize                        int                            `json:"min_size"`
	Name                           string                         `json:"name"`
	ProfileID                      string                         `json:"profile_id"`
	ProfileOnly                    bool                           `json:"profile_only"`
}

// AutoScalingGroup - is represents an auto scaling group
type AutoScalingGroup struct {
	Created                        string                        `json:"created_at"`
	Data                           map[string]interface{}        `json:"data"`
	DeletionPolicyInformations     DeletionInformation           `json:"deletion_info,omitempty"`
	DesiredCapacity                int                           `json:"desired_capacity"`
	ID                             string                        `json:"id"`
	LoadBalancerPolicyInformations LoadBalancerPolicyInformation `json:"load_balancer_info,omitempty"`
	MaxSize                        int                           `json:"max_size"`
	Metadata                       ASMetadata                    `json:"metadata"`
	MinSize                        int                           `json:"min_size"`
	Name                           string                        `json:"name"`
	NodeIDs                        []string                      `json:"node_ids"`
	ProfileID                      string                        `json:"profile_id"`
	ProfileName                    string                        `json:"profile_name"`
	ScaleInPolicyInformations      []ScalePolicyInformation      `json:"scale_in_info,omitempty"`
	ScaleOutPolicyInformations     []ScalePolicyInformation      `json:"scale_out_info,omitempty"`
	Status                         string                        `json:"status"`
	TaskID                         string                        `json:"task_id,omitempty"`
	Timeout                        int                           `json:"timeout"`
	Updated                        string                        `json:"updated_at"`
}

// AutoScalingDataDisk is represents for a data disk being created with servers
type AutoScalingDataDisk struct {
	DeleteOnTermination bool   `json:"delete_on_termination"`
	Size                int    `json:"size"`
	Type                string `json:"type"`
}

// AutoScalingOperatingSystem is represents for operating system being use to create servers
type AutoScalingOperatingSystem struct {
	CreateFrom string `json:"type,omitempty"`
	Error      string `json:"error,omitempty"`
	ID         string `json:"id,omitempty"`
	OSName     string `json:"os_name,omitempty"`
}

// AutoScalingNetworks - is represents for relationship between network and firewalls
type AutoScalingNetworks struct {
	ID             string    `json:"id"`
	SecurityGroups []*string `json:"security_groups,omitempty"`
}

// LaunchConfiguration - is represents a launch configurations
type LaunchConfiguration struct {
	AvailabilityZone string                     `json:"availability_zone"`
	Created          string                     `json:"created_at,omitempty"`
	DataDisks        []*AutoScalingDataDisk     `json:"datadisks,omitempty"`
	Flavor           string                     `json:"flavor"`
	ID               string                     `json:"id,omitempty"`
	Metadata         map[string]interface{}     `json:"metadata"`
	Name             string                     `json:"name"`
	NetworkPlan      string                     `json:"network_plan,omitempty"`
	Networks         []*AutoScalingNetworks     `json:"networks,omitempty"`
	OperatingSystem  AutoScalingOperatingSystem `json:"os"`
	ProfileType      string                     `json:"profile_type,omitempty"`
	RootDisk         *AutoScalingDataDisk       `json:"rootdisk"`
	SecurityGroups   []*string                  `json:"security_groups,omitempty"`
	SSHKey           string                     `json:"key_name,omitempty"`
	Status           string                     `json:"status,omitempty"`
	Type             string                     `json:"type,omitempty"`
	UserData         string                     `json:"user_data,omitempty"`
}

// AutoScalingWebhook is represents for a Webhook to trigger scale
type AutoScalingWebhook struct {
	ActionID    string `json:"action_id"`
	ActionType  string `json:"action_type"`
	ClusterID   string `json:"cluster_id"`
	ClusterName string `json:"cluster_name"`
}

// AutoScalingEvent is represents for a event of auto scaling group
type AutoScalingEvent struct {
	ActionName string `json:"action"`
	ActionType string `json:"type"`
	Metadata   struct {
		Action struct {
			Outputs map[string]interface{} `json:"outputs"`
		} `json:"action"`
	} `json:"meta_data"`
	ClusterID    string `json:"cluster_id"`
	ID           string `json:"id"`
	Level        string `json:"level"`
	ObjectType   string `json:"otype"`
	StatusReason string `json:"status_reason"`
	Timestamp    string `json:"timestamp"`
}

// AutoScalingNode is represents a cloud server in auto scaling group
type AutoScalingNode struct {
	Status       string                 `json:"status"`
	Name         string                 `json:"name"`
	ProfileID    string                 `json:"profile_id"`
	ProfileName  string                 `json:"profile_name"`
	PhysicalID   string                 `json:"physical_id"`
	StatusReason string                 `json:"status_reason"`
	ID           string                 `json:"id"`
	Addresses    map[string]interface{} `json:"addresses"`
}

// AutoScalingNodesDelete is represents a list cloud server being deleted
type AutoScalingNodesDelete struct {
	Nodes []string `json:"nodes"`
}

// usingResource - list snapshot, ssh key using to create launch configurations
type usingResource struct {
	SSHKeys   []string `json:"ssh_keys"`
	Snapshots []string `json:"snapshots"`
}

// usingResource - list snapshot, ssh key using to create launch configurations
type autoscalingQuotas struct {
	Availability map[string]int `json:"can_create,omitempty"`
	Limited      map[string]int `json:"limited,omitempty"`
	Valid        bool           `json:"valid"`
}

// AutoScalingSize - size of auto scaling group
type AutoScalingSize struct {
	DesiredCapacity int `json:"desired_capacity"`
	MaxSize         int `json:"max_size"`
	MinSize         int `json:"min_size"`
}

// AutoScalingSchdeuleValid - represents for a validation time of cron triggers
type AutoScalingSchdeuleValid struct {
	From string  `json:"_from"`
	To   *string `json:"_to,omitempty"`
}

// AutoScalingSchdeuleInputs - represents for a input of cron triggers
type AutoScalingSchdeuleInputs struct {
	CronPattern string          `json:"cron_pattern"`
	Inputs      AutoScalingSize `json:"inputs"`
}

// AutoScalingSchdeuleSizing - represents for phase time of cron triggers
type AutoScalingSchdeuleSizing struct {
	From AutoScalingSchdeuleInputs `json:"_from"`
	To   AutoScalingSchdeuleInputs `json:"_to,omitempty"`
	Type string                    `json:"_type"`
}

// AutoScalingSchdeuleCreateRequest - payload use create a scheduler (cron trigger)
type AutoScalingSchdeuleCreateRequest struct {
	Name   string                    `json:"name"`
	Sizing AutoScalingSchdeuleSizing `json:"sizing"`
	Valid  AutoScalingSchdeuleValid  `json:"valid"`
}

// AutoScalingSchdeule - cron triggers to do time-based scale
type AutoScalingSchdeule struct {
	ClusterID string                    `json:"cluster_id"`
	Created   string                    `json:"created_at"`
	ID        string                    `json:"_id"`
	Name      string                    `json:"name"`
	Sizing    AutoScalingSchdeuleSizing `json:"sizing"`
	Status    string                    `json:"status"`
	TaskID    string                    `json:"task_id"`
	Valid     AutoScalingSchdeuleValid  `json:"valid"`
}

// Auto Scaling Group path
func (asg *autoScalingGroup) resourcePath() string {
	return autoscalingGroupResourcePath
}

func (asg *autoScalingGroup) itemPath(id string) string {
	return strings.Join([]string{autoscalingGroupResourcePath, id}, "/")
}

// Launch Configurations path
func (lc *launchConfiguration) resourcePath() string {
	return launchConfigurationsResourcePath
}

func (lc *launchConfiguration) itemPath(id string) string {
	return strings.Join([]string{launchConfigurationsResourcePath, id}, "/")
}

// Webhook path
func (wh *webhook) resourcePath(clusterID string) string {
	return strings.Join([]string{autoscalingGroupResourcePath, clusterID, webhooksResourcePath}, "/")
}

// Events path
func (e *event) resourcePath(clusterID string, page, total int) string {
	return eventsResourcePath
}

// Policy path
func (p *policy) resourcePath(clusterID string) string {
	return strings.Join([]string{autoscalingGroupResourcePath, clusterID, policiesResourcePath}, "/")
}

func (p *policy) itemPath(clusterID, policyID string) string {
	return strings.Join([]string{autoscalingGroupResourcePath, clusterID, policiesResourcePath, policyID}, "/")
}

// Node path
func (n *node) resourcePath(clusterID string) string {
	return strings.Join([]string{autoscalingGroupResourcePath, clusterID, nodesResourcePath}, "/")
}

// Schedule path
func (s *schedule) resourcePath(clusterID string) string {
	return strings.Join([]string{autoscalingGroupResourcePath, clusterID, schedulesResourcePath}, "/")
}

func (s *schedule) itemPath(clusterID, scheduleID string) string {
	return strings.Join([]string{autoscalingGroupResourcePath, clusterID, schedulesResourcePath, scheduleID}, "/")
}

// Task path
func (t *task) resourcePath(taskID string) string {
	return strings.Join([]string{tasksResourcePath, taskID, "status"}, "/")
}

// Using Resource path
func (c *common) usingResourcePath() string {
	return usingResourcePath
}

func getQuotasResourcePath() string {
	return quotasResourcePath
}

func getSuggestionResourcePath() string {
	return suggestionResourcePath
}

// List
func (asg *autoScalingGroup) List(ctx context.Context, all bool) ([]*AutoScalingGroup, error) {
	req, err := asg.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, asg.resourcePath(), nil)
	if err != nil {
		return nil, err
	}

	if all {
		q := req.URL.Query()
		q.Add("all", "true")
		req.URL.RawQuery = q.Encode()
	}

	resp, err := asg.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		AutoScalingGroups []*AutoScalingGroup `json:"clusters"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.AutoScalingGroups, nil
}

func (lc *launchConfiguration) List(ctx context.Context, all bool) ([]*LaunchConfiguration, error) {
	req, err := lc.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, lc.resourcePath(), nil)
	if err != nil {
		return nil, err
	}

	if all {
		q := req.URL.Query()
		q.Add("all", "true")
		req.URL.RawQuery = q.Encode()
	}

	resp, err := lc.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		LaunchConfigurations []*LaunchConfiguration `json:"profiles"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	for _, LaunchConfiguration := range data.LaunchConfigurations {
		// Force ProfileType = Type
		LaunchConfiguration.ProfileType = LaunchConfiguration.Type

		if LaunchConfiguration.OperatingSystem.Error != "" {
			LaunchConfiguration.Status = statusError
		} else {
			LaunchConfiguration.Status = statusActive
		}
	}
	return data.LaunchConfigurations, nil
}

func (wh *webhook) List(ctx context.Context, clusterID string) ([]*AutoScalingWebhook, error) {
	if clusterID == "" {
		return nil, errors.New("Auto Scaling Group ID is required")
	}

	req, err := wh.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, wh.resourcePath(clusterID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := wh.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data []*AutoScalingWebhook

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (e *event) List(ctx context.Context, clusterID string, page, total int) ([]*AutoScalingEvent, error) {
	if clusterID == "" {
		return nil, errors.New("Auto Scaling Group ID is required")
	}

	req, err := e.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, e.resourcePath(clusterID, page, total), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("cluster_id", clusterID)
	q.Add("page", strconv.Itoa(page))
	q.Add("total", strconv.Itoa(total))
	req.URL.RawQuery = q.Encode()

	resp, err := e.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		AutoScalingEvents []*AutoScalingEvent `json:"events"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.AutoScalingEvents, nil
}

func (p *policy) List(ctx context.Context, clusterID string) (*AutoScalingPolicies, error) {
	if clusterID == "" {
		return nil, errors.New("Auto Scaling Group ID is required")
	}

	req, err := p.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, p.resourcePath(clusterID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data = &AutoScalingPolicies{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (n *node) List(ctx context.Context, clusterID string) ([]*AutoScalingNode, error) {
	if clusterID == "" {
		return nil, errors.New("Auto Scaling Group ID is required")
	}

	req, err := n.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, n.resourcePath(clusterID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := n.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		AutoScalingNodes []*AutoScalingNode `json:"nodes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.AutoScalingNodes, nil
}

func (s *schedule) List(ctx context.Context, clusterID string) ([]*AutoScalingSchdeule, error) {
	if clusterID == "" {
		return nil, errors.New("Auto Scaling Group ID is required")
	}

	req, err := s.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, s.resourcePath(clusterID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		AutoScalingSchdeules []*AutoScalingSchdeule `json:"cron_triggers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.AutoScalingSchdeules, nil
}

// Get
func (asg *autoScalingGroup) Get(ctx context.Context, clusterID string) (*AutoScalingGroup, error) {
	req, err := asg.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, asg.itemPath(clusterID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := asg.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := &AutoScalingGroup{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (lc *launchConfiguration) Get(ctx context.Context, profileID string) (*LaunchConfiguration, error) {
	req, err := lc.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, lc.itemPath(profileID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := lc.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := &LaunchConfiguration{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Force ProfileType = Type
	data.ProfileType = data.Type

	if data.OperatingSystem.Error != "" {
		data.Status = statusError
	} else {
		data.Status = statusActive
	}
	return data, nil
}

func (wh *webhook) Get(ctx context.Context, clusterID string, ActionType string) (*AutoScalingWebhook, error) {
	if clusterID == "" {
		return nil, errors.New("Auto Scaling Group ID is required")
	}
	if _, ok := SliceContains(actionTypeSupportedWebhooks, ActionType); !ok {
		return nil, errors.New("UNSUPPORTED action type")
	}
	req, err := wh.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, wh.resourcePath(clusterID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := wh.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data []*AutoScalingWebhook

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	for _, webhook := range data {
		if webhook.ActionType == ActionType {
			return webhook, nil
		}
	}
	return nil, nil
}

func (t *task) Get(ctx context.Context, taskID string) (*ASTask, error) {
	req, err := t.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, t.resourcePath(taskID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data = &ASTask{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *schedule) Get(ctx context.Context, clusterID, scheduleID string) (*AutoScalingSchdeule, error) {
	req, err := s.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, s.itemPath(clusterID, scheduleID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data = &AutoScalingSchdeule{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *policy) Get(ctx context.Context, clusterID, PolicyID string) (*ScalePolicyInformation, error) {
	req, err := p.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, p.itemPath(clusterID, PolicyID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &ScalePolicyInformation{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

//  Delete
func (asg *autoScalingGroup) Delete(ctx context.Context, clusterID string) error {
	req, err := asg.client.NewRequest(ctx, http.MethodDelete, autoScalingServiceName, asg.itemPath(clusterID), nil)
	if err != nil {
		return err
	}
	resp, err := asg.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	return resp.Body.Close()
}

func (lc *launchConfiguration) Delete(ctx context.Context, profileID string) error {
	req, err := lc.client.NewRequest(ctx, http.MethodDelete, autoScalingServiceName, lc.itemPath(profileID), nil)
	if err != nil {
		return err
	}
	resp, err := lc.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	return resp.Body.Close()
}

func (p *policy) Delete(ctx context.Context, clusterID, PolicyID string) error {
	req, err := p.client.NewRequest(ctx, http.MethodDelete, autoScalingServiceName, p.itemPath(clusterID, PolicyID), nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	return resp.Body.Close()
}

func (n *node) Delete(ctx context.Context, clusterID string, asnd *AutoScalingNodesDelete) error {
	req, err := n.client.NewRequest(ctx, http.MethodDelete, autoScalingServiceName, n.resourcePath(clusterID), &asnd)
	if err != nil {
		return err
	}

	resp, err := n.client.Do(ctx, req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, resp.Body)

	return resp.Body.Close()
}

func (s *schedule) Delete(ctx context.Context, clusterID, scheduleID string) error {
	req, err := s.client.NewRequest(ctx, http.MethodDelete, autoScalingServiceName, s.itemPath(clusterID, scheduleID), nil)
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

// Create
func (asg *autoScalingGroup) Create(ctx context.Context, ascr *AutoScalingGroupCreateRequest) (*AutoScalingGroup, error) {
	if valid, _ := isValidQuotas(ctx, asg.client, "", ascr.ProfileID, ascr.DesiredCapacity, ascr.MaxSize); !valid {
		return nil, errors.New("Not enough quotas to create new auto scaling group")
	}
	req, err := asg.client.NewRequest(ctx, http.MethodPost, autoScalingServiceName, asg.resourcePath(), &ascr)
	if err != nil {
		return nil, err
	}

	resp, err := asg.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &AutoScalingGroup{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (lc *launchConfiguration) Create(ctx context.Context, lcr *LaunchConfiguration) (*LaunchConfiguration, error) {
	if _, ok := SliceContains(networkPlan, lcr.NetworkPlan); !ok {
		return nil, errors.New("UNSUPPORTED network plan")
	}

	req, err := lc.client.NewRequest(ctx, http.MethodPost, autoScalingServiceName, lc.resourcePath(), &lcr)
	if err != nil {
		return nil, err
	}

	resp, err := lc.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &LaunchConfiguration{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (p *policy) Create(ctx context.Context, clusterID string, pcr *PolicyCreateRequest) (*TaskResponses, error) {
	req, err := p.client.NewRequest(ctx, http.MethodPost, autoScalingServiceName, p.resourcePath(clusterID), &pcr)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &TaskResponses{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (p *policy) CreateHooks(ctx context.Context, clusterID string, hpcr *HooksPolicyCreateRequest) (*TaskResponses, error) {
	req, err := p.client.NewRequest(ctx, http.MethodPost, autoScalingServiceName, p.resourcePath(clusterID), &hpcr)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &TaskResponses{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (p *policy) CreateLoadBalancers(ctx context.Context, clusterID string, lbpcr *LoadBalancersPolicyCreateRequest) (*TaskResponses, error) {
	req, err := p.client.NewRequest(ctx, http.MethodPost, autoScalingServiceName, p.resourcePath(clusterID), &lbpcr)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &TaskResponses{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *schedule) Create(ctx context.Context, clusterID string, asscr *AutoScalingSchdeuleCreateRequest) (*TaskResponses, error) {
	req, err := s.client.NewRequest(ctx, http.MethodPost, autoScalingServiceName, s.resourcePath(clusterID), &asscr)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &TaskResponses{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

// Update
func (asg *autoScalingGroup) Update(ctx context.Context, clusterID string, asur *AutoScalingGroupUpdateRequest) (*AutoScalingGroup, error) {
	if valid, _ := isValidQuotas(ctx, asg.client, clusterID, asur.ProfileID, asur.DesiredCapacity, asur.MaxSize); !valid {
		return nil, errors.New("Not enough quotas to update new auto scaling group")
	}

	req, err := asg.client.NewRequest(ctx, http.MethodPut, autoScalingServiceName, asg.itemPath(clusterID), &asur)
	if err != nil {
		return nil, err
	}

	resp, err := asg.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &AutoScalingGroup{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (p *policy) UpdateLoadBalancers(ctx context.Context, clusterID, PolicyID string, lbpur *LoadBalancersPolicyUpdateRequest) (*TaskResponses, error) {
	req, err := p.client.NewRequest(ctx, http.MethodPut, autoScalingServiceName, p.itemPath(clusterID, PolicyID), &lbpur)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &TaskResponses{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (p *policy) Update(ctx context.Context, clusterID, PolicyID string, pur *PolicyUpdateRequest) (*TaskResponses, error) {
	req, err := p.client.NewRequest(ctx, http.MethodPut, autoScalingServiceName, p.itemPath(clusterID, PolicyID), &pur)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &TaskResponses{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

// Common
func (c *common) AutoScalingUsingResource(ctx context.Context) (*usingResource, error) {
	req, err := c.client.NewRequest(ctx, http.MethodGet, autoScalingServiceName, c.usingResourcePath(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := &usingResource{}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *common) AutoScalingIsValidQuotas(ctx context.Context, clusterID, ProfileID string, desiredCapacity, maxSize int) (bool, error) {
	return isValidQuotas(ctx, c.client, clusterID, ProfileID, desiredCapacity, maxSize)
}

func isValidQuotas(ctx context.Context, client *Client, clusterID, ProfileID string, desiredCapacity, maxSize int) (bool, error) {
	payload := map[string]interface{}{
		"desired_capacity": desiredCapacity,
		"max_size":         maxSize,
		"profile_id":       ProfileID,
	}

	if clusterID != "" {
		payload["cluster_id"] = clusterID
	}

	req, err := client.NewRequest(ctx, http.MethodPost, autoScalingServiceName, getQuotasResourcePath(), &payload)
	if err != nil {
		return false, err
	}

	resp, err := client.Do(ctx, req)
	if err != nil {
		return false, err
	}

	var data struct {
		Quotas autoscalingQuotas `json:"message"`
	}

	// data := &map[string]interface{}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, err
	}
	return data.Quotas.Valid, nil
}

func (c *common) AutoScalingGetSuggestion(ctx context.Context, ProfileID string, desiredCapacity, maxSize int) (interface{}, error) {
	return getSuggestion(ctx, c.client, ProfileID, desiredCapacity, maxSize)
}

// getSuggestion do get suggestion
func getSuggestion(ctx context.Context, client *Client, ProfileID string, desiredCapacity, maxSize int) (interface{}, error) {
	payload := map[string]interface{}{
		"desired_capacity": desiredCapacity,
		"max_size":         maxSize,
		"profile_id":       ProfileID,
	}

	req, err := client.NewRequest(ctx, http.MethodPost, autoScalingServiceName, getSuggestionResourcePath(), &payload)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
