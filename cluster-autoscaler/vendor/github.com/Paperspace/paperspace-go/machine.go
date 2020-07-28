package paperspace

import (
	"fmt"
	"time"
)

type MachineState string

const (
	MachineStateOff          MachineState = "off"
	MachineStateProvisioning MachineState = "provisioning"
	MachineStateRunning      MachineState = "running"
)

type Machine struct {
	ID                     string       `json:"id"`
	Name                   string       `json:"name"`
	OS                     string       `json:"os"`
	RAM                    int64        `json:"ram,string"`
	CPUs                   int          `json:"cpus"`
	GPU                    string       `json:"gpu"`
	State                  MachineState `json:"state"`
	Region                 string       `json:"region"`
	StorageTotal           int64        `json:"storageTotal,string"`
	StorageUsed            int64        `json:"storageUsed,string"`
	UsageRate              string       `json:"usageRate"`
	ShutdownTimeoutInHours int          `json:"shutdownTimeoutInHours"`
	ShutdownTimeoutForces  bool         `json:"shutdownTimeoutForces"`
	AutoSnapshotFrequency  int          `json:"autoSnapshotFrequency"`
	AutoSnapshotSaveCount  int          `json:"autoSnapshotSaveCount"`
	AgentType              string       `json:"agentType"`
	NetworkID              string       `json:"networkId"`
	PrivateIpAddress       string       `json:"privateIpAddress"`
	PublicIpAddress        string       `json:"publicIpAddress"`
	DtCreated              time.Time    `json:"dtCreated"`
	DtDeleted              time.Time    `json:"dtDeleted"`
	UserID                 string       `json:"userId"`
	TeamID                 string       `json:"teamId"`
	ScriptID               string       `json:"scriptId"`
	DtLastRun              string       `json:"dtLastRun"`
	IsManaged              bool         `json:"isManaged"`
}

type MachineCreateParams struct {
	RequestParams

	Name                   string `json:"name"`
	Region                 string `json:"region"`
	MachineType            string `json:"machineType"`
	Size                   int    `json:"size"`
	BillingType            string `json:"billingType"`
	TemplateID             string `json:"templateId"`
	UserID                 string `json:"userId,omitempty"`
	TeamID                 string `json:"teamId,omitempty"`
	ScriptID               string `json:"scriptId,omitempty"`
	NetworkID              string `json:"networkId,omitempty"`
	ShutdownTimeoutInHours bool   `json:"shutdownTimeoutInHours,omitempty"`
	AssignPublicIP         bool   `json:"assignPublicIP,omitempty"`
	IsManaged              bool   `json:"isManaged,omitempty"`
}

type MachineDeleteParams struct {
	RequestParams
}

type MachineGetParams struct {
	RequestParams
}

type MachineListParams struct {
	RequestParams

	Filter map[string]string `json:"filter,omitempty"`
}

type MachineUpdateAttributeParams struct {
	RequestParams

	Name string `json:"name,omitempty" yaml:"name"`
}

type MachineUpdateParams struct {
	RequestParams

	ID                     string `json:"machineId"`
	Name                   string `json:"machineName,omitempty"`
	ShutdownTimeoutInHours int    `json:"shutdownTimeoutInHours,omitempty"`
	ShutdownTimeoutForces  bool   `json:"shutdownTimeoutForces,omitempty"`
	AutoSnapshotFrequency  string `json:"autoSnapshotFrequency,omitempty"`
	AutoSnapshotSaveCount  int    `json:"autoSnapshotSaveCount,omitempty"`
	PerformAutoSnapshot    bool   `json:"performAutoSnapshot,omitempty"`
	DynamicPublicIP        bool   `json:"dynamicPublicIp,omitempty"`
}

func NewMachineListParams() MachineListParams {
	machineListParams := MachineListParams{
		Filter: make(map[string]string),
	}

	return machineListParams
}

func (c Client) CreateMachine(params MachineCreateParams) (Machine, error) {
	machine := Machine{}

	url := fmt.Sprintf("/machines/createSingleMachinePublic")
	_, err := c.Request("POST", url, params, &machine, params.RequestParams)

	return machine, err
}

func (c Client) GetMachine(id string, params MachineGetParams) (Machine, error) {
	machine := Machine{}

	url := fmt.Sprintf("/machines/getMachinePublic?machineId=%s", id)
	_, err := c.Request("GET", url, nil, &machine, params.RequestParams)

	return machine, err
}

func (c Client) GetMachines(params MachineListParams) ([]Machine, error) {
	var machines []Machine

	url := fmt.Sprintf("/machines/getMachines")
	_, err := c.Request("GET", url, params, &machines, params.RequestParams)

	return machines, err
}

func (c Client) UpdateMachine(params MachineUpdateParams) (Machine, error) {
	machine := Machine{}

	url := fmt.Sprintf("/machines/updateMachine")
	_, err := c.Request("POST", url, params, &machine, params.RequestParams)

	return machine, err
}

func (c Client) DeleteMachine(id string, params MachineDeleteParams) error {
	url := fmt.Sprintf("/machines/%s/destroyMachine", id)
	_, err := c.Request("POST", url, nil, nil, params.RequestParams)

	return err
}
