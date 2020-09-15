package paperspace

import (
	"fmt"
)

type AutoscalingGroup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Min         int       `json:"min"`
	Max         int       `json:"max"`
	Current     int       `json:"current"`
	ClusterID   string    `json:"clusterId"`
	MachineType string    `json:"machineType"`
	TemplateID  string    `json:"templateId"`
	ScriptID    string    `json:"startupScriptId"`
	NetworkID   string    `json:"networkId"`
	Nodes       []Machine `json:"nodes"`
}

type AutoscalingGroupCreateParams struct {
	RequestParams

	Name        string `json:"name"`
	Min         int    `json:"min"`
	Max         int    `json:"max"`
	ClusterID   string `json:"clusterId"`
	MachineType string `json:"machineType"`
	TemplateID  string `json:"templateId"`
	ScriptID    string `json:"startupScriptId,omitempty"`
	NetworkID   string `json:"networkId"`
}

type AutoscalingGroupDeleteParams struct {
	RequestParams
}

type AutoscalingGroupGetParams struct {
	RequestParams

	IncludeNodes bool `json:"includeNodes,omitempty"`
}

type AutoscalingGroupListParams struct {
	RequestParams

	Filter       Filter `json:"filter,omitempty"`
	IncludeNodes bool   `json:"includeNodes,omitempty"`
}

type AutoscalingGroupUpdateAttributeParams struct {
	Name        string `json:"name,omitempty"`
	Min         *int   `json:"min,omitempty"`
	Max         *int   `json:"max,omitempty"`
	Current     *int   `json:"current,omitempty"`
	MachineType string `json:"machineType,omitempty"`
	TemplateID  string `json:"templateId,omitempty"`
	ScriptID    string `json:"startupScriptId,omitempty"`
	NetworkID   string `json:"networkId,omitempty"`
}

type AutoscalingGroupUpdateParams struct {
	RequestParams

	Attributes AutoscalingGroupUpdateAttributeParams `json:"attributes,omitempty"`
}

func (c Client) CreateAutoscalingGroup(params AutoscalingGroupCreateParams) (AutoscalingGroup, error) {
	autoscalingGroup := AutoscalingGroup{}

	url := fmt.Sprintf("/autoscalingGroups")
	_, err := c.Request("POST", url, params, &autoscalingGroup, params.RequestParams)

	return autoscalingGroup, err
}

func (c Client) GetAutoscalingGroup(id string, params AutoscalingGroupGetParams) (AutoscalingGroup, error) {
	autoscalingGroup := AutoscalingGroup{}

	url := fmt.Sprintf("/autoscalingGroups/%s", id)
	_, err := c.Request("GET", url, params, &autoscalingGroup, params.RequestParams)

	return autoscalingGroup, err
}

func (c Client) GetAutoscalingGroups(params AutoscalingGroupListParams) ([]AutoscalingGroup, error) {
	var autoscalingGroups []AutoscalingGroup

	url := fmt.Sprintf("/autoscalingGroups")
	_, err := c.Request("GET", url, params, &autoscalingGroups, params.RequestParams)

	return autoscalingGroups, err
}

func (c Client) UpdateAutoscalingGroup(id string, params AutoscalingGroupUpdateParams) error {
	url := fmt.Sprintf("/autoscalingGroups/%s", id)
	_, err := c.Request("PATCH", url, params, nil, params.RequestParams)

	return err
}

func (c Client) DeleteAutoscalingGroup(id string, params AutoscalingGroupDeleteParams) error {
	url := fmt.Sprintf("/autoscalingGroups/%s", id)
	_, err := c.Request("DELETE", url, nil, nil, params.RequestParams)

	return err
}
