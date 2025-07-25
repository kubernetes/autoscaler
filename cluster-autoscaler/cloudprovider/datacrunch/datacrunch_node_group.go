/*
Copyright 2019 The Kubernetes Authors.

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

package datacrunch

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"os"
	"strings"
	"sync"
	texttmpl "text/template"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	datacrunchclient "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/datacrunch/datacrunch-go"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

const (
	// ResourceGPU is the resource name for GPU
	ResourceGPU apiv1.ResourceName = "nvidia.com/gpu"
)

var (
	preStartupScriptTemplate = `
#!/bin/bash

# 1. get access token
TOKEN_RESPONSE=$(curl https://api.datacrunch.io/v1/oauth2/token \
    --request POST \
    --header 'Content-Type: application/json' \
    --data '{
    "grant_type": "client_credentials",
    "client_id": "{{ .DATACRUNCH_CLIENT_ID }}",
    "client_secret": "{{ .DATACRUNCH_CLIENT_SECRET }}"
}'
)

ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token')

# 2. if script should be deleted
if [ "{{ .DELETE_SCRIPT }}" = "true" ]; then
    # a. Get all scripts and find the one with the matching name
    REAL_SCRIPT_ID=$(curl -s https://api.datacrunch.io/v1/scripts \
        --header "Authorization: Bearer $ACCESS_TOKEN" | \
        jq -r --arg NAME "{{ .SCRIPT_NAME }}" '.[] | select(.name == $NAME) | .id')

    # b. delete the script
    if [ -n "$REAL_SCRIPT_ID" ]; then
        echo "Deleting script with id: $REAL_SCRIPT_ID (name: {{ .SCRIPT_NAME }})"
        curl -s -X DELETE https://api.datacrunch.io/v1/scripts \
        --header "Authorization: Bearer $ACCESS_TOKEN" \
        --header 'Content-Type: application/json' \
        --data '{"scripts": ["'$REAL_SCRIPT_ID'"]}'
    else
        echo "Script with name {{ .SCRIPT_NAME }} not found, skipping deletion."
    fi
fi

# 3.Get instance ID based on $HOSTNAME
INSTANCE_ID=$(curl -s https://api.datacrunch.io/v1/instances \
    --header "Authorization: Bearer $ACCESS_TOKEN" | \
    jq -r --arg HOSTNAME "$HOSTNAME" '.[] | select(.hostname == $HOSTNAME) | .id')

if [ -n "$INSTANCE_ID" ]; then
    echo "Instance ID for hostname $HOSTNAME is $INSTANCE_ID"
else
    echo "No instance found for hostname $HOSTNAME"
fi
	`
)

// datacrunchNodeGroup implements cloudprovider.NodeGroup interface. datacrunchNodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type datacrunchNodeGroup struct {
	id           string
	manager      *datacrunchManager
	minSize      int
	maxSize      int
	targetSize   int
	region       string
	instanceType string

	clusterUpdateMutex *sync.Mutex
}

type datacrunchNodeGroupSpec struct {
	name         string
	minSize      int
	maxSize      int
	region       string
	instanceType string
}

// MaxSize returns maximum size of the node group.
func (n *datacrunchNodeGroup) MaxSize() int {
	return n.maxSize
}

// MinSize returns minimum size of the node group.
func (n *datacrunchNodeGroup) MinSize() int {
	return n.minSize
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (n *datacrunchNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// TargetSize returns the current target size of the node group. It is possible
// that the number of nodes in Kubernetes is different at the moment but should
// be equal to Size() once everything stabilizes (new nodes finish startup and
// registration or removed nodes are deleted completely). Implementation
// required.
func (n *datacrunchNodeGroup) TargetSize() (int, error) {
	return n.targetSize, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *datacrunchNodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	desiredTargetSize := n.targetSize + delta
	if desiredTargetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d", n.targetSize, desiredTargetSize, n.MaxSize())
	}

	actualDelta := delta

	klog.V(4).Infof("Scaling Instance Pool %s to %d", n.id, desiredTargetSize)

	n.clusterUpdateMutex.Lock()
	defer n.clusterUpdateMutex.Unlock()

	instanceOption := n.manager.clusterConfig.NodeConfigs[n.id].InstanceOption

	var available bool
	var err error

	switch instanceOption {
	case InstanceOptionPreferSpot:
		available, err = serverTypeAvailable(n.manager, n.instanceType, n.region, true)
		if err != nil {
			klog.V(4).Infof("Failed to check if server type %s is available in region %s with isSpot %t, trying to check with isSpot %t", n.instanceType, n.region, true, false)
		}
		if !available {
			available, err = serverTypeAvailable(n.manager, n.instanceType, n.region, false)
			if err != nil {
				klog.V(4).Infof("Failed to check if server type %s is available in region %s with isSpot %t, trying to check with isSpot %t", n.instanceType, n.region, false, true)
			}
		}
	case InstanceOptionPreferOnDemand:
		available, err = serverTypeAvailable(n.manager, n.instanceType, n.region, false)
		if err != nil {
			klog.V(4).Infof("Failed to check if server type %s is available in region %s with isSpot %t, trying to check with isSpot %t", n.instanceType, n.region, false, true)
		}
		if !available {
			available, err = serverTypeAvailable(n.manager, n.instanceType, n.region, true)
			if err != nil {
				klog.V(4).Infof("Failed to check if server type %s is available in region %s with isSpot %t, trying to check with isSpot %t", n.instanceType, n.region, true, false)
			}
		}
	case InstanceOptionSpotOnly:
		available, err = serverTypeAvailable(n.manager, n.instanceType, n.region, true)
	case InstanceOptionOnDemandOnly:
		available, err = serverTypeAvailable(n.manager, n.instanceType, n.region, false)
	}

	if err != nil {
		return fmt.Errorf("failed to check if server type %s is available in region %s with instance option %s: %v", n.instanceType, n.region, instanceOption, err)
	}

	if !available {
		return fmt.Errorf("server type %s not available in region %s with instance option %s", n.instanceType, n.region, instanceOption)
	}

	defer func() {
		// create new servers cache
		if _, err := n.manager.cachedServers.servers(); err != nil {
			klog.Errorf("failed to update servers cache: %v", err)
		}

		// Update target size
		n.resetTargetSize(actualDelta)
	}()

	// There is no "Server Group" in Datacrunch Cloud, we need to create every
	// server manually. This operation might fail for some of the servers
	// because of quotas, rate limiting or server type availability. We need to
	// collect the errors and inform cluster-autoscaler about this, so it can
	// try other node groups if configured.
	waitGroup := sync.WaitGroup{}
	errsCh := make(chan error, delta)
	for i := 0; i < delta; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := createServer(n)
			if err != nil {
				actualDelta--
				errsCh <- err
			}
		}()
	}
	waitGroup.Wait()
	close(errsCh)

	errs := make([]error, 0, delta)
	for err = range errsCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to create all servers: %w", errors.Join(errs...))
	}

	return nil
}

// AtomicIncreaseSize is not implemented.
func (n *datacrunchNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group (and also increasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *datacrunchNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	n.clusterUpdateMutex.Lock()
	defer n.clusterUpdateMutex.Unlock()

	delta := len(nodes)

	targetSize := n.targetSize - delta
	if targetSize < n.MinSize() {
		return fmt.Errorf("size decrease is too large. current: %d desired: %d min: %d", n.targetSize, targetSize, n.MinSize())
	}

	actualDelta := delta

	defer func() {
		// create new servers cache
		if _, err := n.manager.cachedServers.servers(); err != nil {
			klog.Errorf("failed to update servers cache: %v", err)
		}

		n.resetTargetSize(-actualDelta)
	}()

	waitGroup := sync.WaitGroup{}
	errsCh := make(chan error, len(nodes))
	for _, node := range nodes {
		waitGroup.Add(1)
		go func(node *apiv1.Node) {
			klog.Infof("Evicting server %s", node.Name)

			err := n.manager.deleteByNode(node)
			if err != nil {
				actualDelta--
				errsCh <- fmt.Errorf("failed to delete server for node %q: %w", node.Name, err)
			}

			waitGroup.Done()
		}(node)
	}
	waitGroup.Wait()
	close(errsCh)

	errs := make([]error, 0, len(nodes))
	for err := range errsCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to delete all nodes: %w", errors.Join(errs...))
	}

	return nil
}

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
func (n *datacrunchNodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *datacrunchNodeGroup) DecreaseTargetSize(delta int) error {
	n.targetSize = n.targetSize + delta
	return nil
}

// Id returns an unique identifier of the node group.
func (n *datacrunchNodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *datacrunchNodeGroup) Debug() string {
	return fmt.Sprintf("cluster ID: %s (min:%d max:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.  It is
// required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *datacrunchNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	servers, err := n.manager.cachedServers.getServersByNodeGroupName(n.id)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers for datacrunch: %v", err)
	}

	instances := make([]cloudprovider.Instance, 0, len(servers))
	for _, vm := range servers {
		instances = append(instances, toInstance(vm))
	}

	return instances, nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The
// returned NodeInfo is expected to have a fully populated Node object, with
// all of the labels, capacity and allocatable information as well as all pods
// that are started on the node by default, using manifest (most likely only
// kube-proxy). Implementation optional.
func (n *datacrunchNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	resourceList, err := getMachineTypeResourceList(n)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource list for node group %s error: %v", n.id, err)
	}

	nodeName := newNodeName(n)

	node := apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
			Labels: map[string]string{
				apiv1.LabelHostname: nodeName,
			},
		},
		Status: apiv1.NodeStatus{
			Capacity:   resourceList,
			Conditions: cloudprovider.BuildReadyConditions(),
		},
	}
	node.Status.Allocatable = node.Status.Capacity
	node.Status.Conditions = cloudprovider.BuildReadyConditions()

	nodeGroupLabels, err := buildNodeGroupLabels(n)
	if err != nil {
		return nil, err
	}
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, nodeGroupLabels)

	for _, taint := range n.manager.clusterConfig.NodeConfigs[n.id].Taints {
		node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
			Key:    taint.Key,
			Value:  taint.Value,
			Effect: taint.Effect,
		})
	}

	nodeInfo := framework.NewNodeInfo(&node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(n.id)})
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side.
// Allows to tell the theoretical node group from the real one. Implementation
// required.
func (n *datacrunchNodeGroup) Exist() bool {
	_, exists := n.manager.nodeGroups[n.id]
	return exists
}

// Create creates the node group on the cloud provider side. Implementation
// optional.
func (n *datacrunchNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	n.manager.nodeGroups[n.id] = n

	return n, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.  This will be
// executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *datacrunchNodeGroup) Delete() error {
	// We do not use actual node groups but all nodes within the Datacrunch project are labeled with a group
	return nil
}

// Autoprovisioned returns true if the node group is autoprovisioned. An
// autoprovisioned group was created by CA and can be deleted when scaled to 0.
func (n *datacrunchNodeGroup) Autoprovisioned() bool {
	// All groups are auto provisioned
	return false
}

func toInstance(vm *datacrunchclient.Instance) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     toProviderID(vm.ID),
		Status: toInstanceStatus(vm),
	}
}

func toProviderID(nodeID string) string {
	return fmt.Sprintf("%s%s", providerIDPrefix, nodeID)
}

func toInstanceStatus(vm *datacrunchclient.Instance) *cloudprovider.InstanceStatus {
	if vm.Status == "" {
		return nil
	}

	st := &cloudprovider.InstanceStatus{}
	switch vm.Status {
	case "provisioning":
		st.State = cloudprovider.InstanceCreating
	case "running":
		st.State = cloudprovider.InstanceRunning
	case "offline", "discontinued":
		st.State = cloudprovider.InstanceDeleting
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "unknown-status",
			ErrorMessage: "unknown status",
		}
	}

	return st
}

func newNodeName(n *datacrunchNodeGroup) string {
	return fmt.Sprintf("%s-%x", n.id, rand.Int63())
}

func buildNodeGroupLabels(n *datacrunchNodeGroup) (map[string]string, error) {
	klog.V(4).Infof("Build node group label for %s", n.id)

	labels := map[string]string{
		apiv1.LabelInstanceType:   n.instanceType,
		apiv1.LabelTopologyRegion: n.region,
		nodeGroupLabel:            n.id,
	}

	maps.Copy(labels, n.manager.clusterConfig.NodeConfigs[n.id].Labels)

	klog.V(4).Infof("%s nodegroup labels: %s", n.id, labels)

	return labels, nil
}

func getMachineTypeResourceList(n *datacrunchNodeGroup) (apiv1.ResourceList, error) {
	typeInfo, err := n.manager.cachedServerType.getServerType(n.instanceType)
	if err != nil || typeInfo == nil {
		return nil, fmt.Errorf("failed to get machine type %s info error: %v", n.instanceType, err)
	}

	diskSizeGB := n.manager.clusterConfig.NodeConfigs[n.id].DiskSizeGB

	numGPUs := typeInfo.GPU.NumberOfGPUs

	// Override the number of GPUs if specified. Useful for MIG mode.
	if n.manager.clusterConfig.NodeConfigs[n.id].OverrideNumGPUs != nil {
		numGPUs = *n.manager.clusterConfig.NodeConfigs[n.id].OverrideNumGPUs
	}

	return apiv1.ResourceList{
		apiv1.ResourcePods:             *resource.NewQuantity(defaultPodAmountsLimit, resource.DecimalSI),
		apiv1.ResourceCPU:              *resource.NewQuantity(int64(typeInfo.CPU.NumberOfCores), resource.DecimalSI),
		ResourceGPU:                    *resource.NewQuantity(int64(numGPUs), resource.DecimalSI),
		apiv1.ResourceMemory:           *resource.NewQuantity(int64(typeInfo.Memory.SizeInGigabytes*1024*1024*1024), resource.DecimalSI),
		apiv1.ResourceEphemeralStorage: *resource.NewQuantity(int64(diskSizeGB*1024*1024*1024), resource.DecimalSI),
	}, nil
}

func serverTypeAvailable(manager *datacrunchManager, instanceType string, region string, isSpot bool) (bool, error) {
	return manager.cachedServerType.GetInstanceTypeAvailabilityCached(instanceType, region, isSpot)
}

func processPreScriptTemplate(preScriptTemplate string, data interface{}) (string, error) {
	if preScriptTemplate == "" {
		return "", nil
	}

	tmpl, err := texttmpl.New("pre-script").Parse(preScriptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse pre-script template: %v", err)
	}

	var script bytes.Buffer
	if err := tmpl.Execute(&script, data); err != nil {
		return "", fmt.Errorf("failed to execute pre-script template: %v", err)
	}

	return script.String(), nil
}

func buildPreScript(n *datacrunchNodeGroup, scriptName string) (string, error) {
	// Get credentials from environment
	clientID := os.Getenv("DATACRUNCH_CLIENT_ID")
	clientSecret := os.Getenv("DATACRUNCH_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("DATACRUNCH_CLIENT_ID and DATACRUNCH_CLIENT_SECRET must be set")
	}

	deleteScriptsAfterBoot := (strings.ToLower(os.Getenv("DATACRUNCH_DELETE_SCRIPTS_AFTER_BOOT")) == "true")

	// Prepare template data
	templateData := map[string]string{
		"DATACRUNCH_CLIENT_ID":     clientID,
		"DATACRUNCH_CLIENT_SECRET": clientSecret,
		"SCRIPT_NAME":              scriptName,
		"DELETE_SCRIPT":            fmt.Sprintf("%t", deleteScriptsAfterBoot),
	}

	// Process the template
	return processPreScriptTemplate(preStartupScriptTemplate, templateData)
}

func deployInstance(client *datacrunchclient.Client, deployReq datacrunchclient.DeployInstanceRequest, instanceOption InstanceOption, pricingOption *PricingOption) error {
	// initial deploy request
	switch instanceOption {
	case InstanceOptionSpotOnly:
		deployReq.IsSpot = true
	case InstanceOptionOnDemandOnly:
		deployReq.IsSpot = false
	case InstanceOptionPreferSpot:
		deployReq.IsSpot = true
	case InstanceOptionPreferOnDemand:
		deployReq.IsSpot = false
	}

	// if not spot, set pricing based on pricing option. Default is dynamic.
	if !deployReq.IsSpot && pricingOption != nil {
		switch *pricingOption {
		case PricingOptionDynamic:
			deployReq.Pricing = "DYNAMIC_PRICE"
		case PricingOptionFixed:
			deployReq.Pricing = "FIXED_PRICE"
		default:
			klog.Warningf("Unknown pricing option: %s", *pricingOption)
		}
	}

	klog.V(4).Infof("Trying to deploy instance %+v", deployReq)
	_, err := client.DeployInstance(deployReq)

	if err != nil {
		if strings.Contains(err.Error(), "Not enough resources to deploy") && (instanceOption == InstanceOptionPreferSpot || instanceOption == InstanceOptionPreferOnDemand) {
			if instanceOption == InstanceOptionPreferSpot {
				klog.V(4).Infof("Got error: %v, not enough resources to deploy instance %+v, trying to deploy instance with on_demand instead", err, deployReq)
				return deployInstance(client, deployReq, InstanceOptionOnDemandOnly, pricingOption)
			}

			klog.V(4).Infof("Got error: %v, not enough resources to deploy instance %+v, trying to deploy instance with spot instead", err, deployReq)
			return deployInstance(client, deployReq, InstanceOptionSpotOnly, pricingOption)
		}

		return fmt.Errorf("could not create instance type %s in region %s: %v", deployReq.InstanceType, deployReq.LocationCode, err)
	}
	return nil
}

func createServer(n *datacrunchNodeGroup) error {
	typeInfo, err := n.manager.cachedServerType.getServerType(n.instanceType)
	if err != nil {
		return err
	}

	diskSizeGB := n.manager.clusterConfig.NodeConfigs[n.id].DiskSizeGB
	image := n.manager.clusterConfig.NodeConfigs[n.id].ImageType
	sshKeyIDs := n.manager.clusterConfig.NodeConfigs[n.id].SSHKeyIDs
	instanceOption := n.manager.clusterConfig.NodeConfigs[n.id].InstanceOption
	// generate node name
	nodeName := newNodeName(n)
	startupScriptName := fmt.Sprintf("autoscaler-startup-script-%s", nodeName)

	var startupScriptID string
	startupScript := os.Getenv("DATACRUNCH_STARTUP_SCRIPT")
	startupScriptFile := os.Getenv("DATACRUNCH_STARTUP_SCRIPT_FILE")
	if startupScript == "" && startupScriptFile != "" {
		startupScriptBytes, err := os.ReadFile(startupScriptFile)
		if err != nil {
			return fmt.Errorf("failed to read startup script file: %v", err)
		}
		startupScript = string(startupScriptBytes)
	}

	if n.manager.clusterConfig.NodeConfigs[n.id].StartupScriptBase64 != "" {
		startupScriptBytes, err := base64.StdEncoding.DecodeString(n.manager.clusterConfig.NodeConfigs[n.id].StartupScriptBase64)
		if err != nil {
			return fmt.Errorf("failed to decode startup script: %v", err)
		}
		startupScript = string(startupScriptBytes)

	}

	if startupScript != "" {
		// Build pre-script from template
		preScript, err := buildPreScript(n, startupScriptName)
		if err != nil {
			return fmt.Errorf("failed to build pre-script: %v", err)
		}

		// Combine pre-script with user script
		finalScript := preScript + "\n\n# User startup script starts here\n" + startupScript
		klog.V(4).Infof("Combined pre-script with user startup script")

		startupScriptID, err = n.manager.client.UploadStartupScript(startupScriptName, finalScript)
		if err != nil {
			return fmt.Errorf("failed to upload startup script: %v", err)
		}
		klog.V(4).Infof("Uploaded startup script defined in cluster config with ID: %s", startupScriptID)
	}

	deployReq := datacrunchclient.DeployInstanceRequest{
		InstanceType: typeInfo.InstanceType,
		Image:        image,
		Hostname:     nodeName,
		Description:  n.id,
		LocationCode: strings.ToUpper(n.region),
		OSVolume: &datacrunchclient.OSVolume{
			Name: nodeName, // also use node name as volume name so we can delete the volume later during scale down
			Size: diskSizeGB,
		},
		StartupScriptID: startupScriptID,
		SSHKeyIDs:       sshKeyIDs,
	}

	// get pricing option
	pricingOption := n.manager.clusterConfig.NodeConfigs[n.id].PricingOption

	// deploy instance
	err = deployInstance(n.manager.client, deployReq, instanceOption, pricingOption)
	if err != nil {
		return fmt.Errorf("could not create instance type %s in region %s: %v", n.instanceType, n.region, err)
	}

	return nil
}

func (n *datacrunchNodeGroup) resetTargetSize(expectedDelta int) {
	servers, err := n.manager.allServers(n.id)
	if err != nil {
		klog.Warningf("failed to set node pool %s size, using delta %d error: %v", n.id, expectedDelta, err)
		n.targetSize = n.targetSize + expectedDelta
	} else {
		klog.Infof("Set node group %s size from %d to %d, expected delta %d", n.id, n.targetSize, len(servers), expectedDelta)
		n.targetSize = len(servers)
	}
}
