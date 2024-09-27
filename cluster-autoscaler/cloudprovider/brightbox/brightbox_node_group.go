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

package brightbox

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	brightbox "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/gobrightbox"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/gobrightbox/status"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/brightbox/k8ssdk"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// Allocatable Resources reserves
	// Reserve 4% of memory
	memoryReservePercent = 4
	// with a minimum of 160MB
	minimumMemoryReserve = 167772160
	// Reserve 5GB of disk space
	minimumDiskReserve = 5368709120
)

var (
	checkInterval = time.Second * 1
	checkTimeout  = time.Second * 30
)

type brightboxNodeGroup struct {
	id            string
	minSize       int
	maxSize       int
	serverOptions *brightbox.ServerOptions
	*k8ssdk.Cloud
}

// MaxSize returns maximum size of the node group.
func (ng *brightboxNodeGroup) MaxSize() int {
	klog.V(4).Info("MaxSize")
	return ng.maxSize
}

// MinSize returns minimum size of the node group.
func (ng *brightboxNodeGroup) MinSize() int {
	klog.V(4).Info("MinSize")
	return ng.minSize
}

// TargetSize returns the current target size of the node group. It
// is possible that the number of nodes in Kubernetes is different at
// the moment but should be equal to Size() once everything stabilizes
// (new nodes finish startup and registration or removed nodes are deleted
// completely). Implementation required.
func (ng *brightboxNodeGroup) TargetSize() (int, error) {
	klog.V(4).Info("TargetSize")
	group, err := ng.GetServerGroup(ng.Id())
	if err != nil {
		return 0, err
	}
	return len(group.Servers), nil
}

// CurrentSize returns the current actual size of the node group.
func (ng *brightboxNodeGroup) CurrentSize() (int, error) {
	klog.V(4).Info("CurrentSize")
	// The implementation is currently synchronous, so
	// CurrentSize and TargetSize will be identical at all times
	return ng.TargetSize()
}

// IncreaseSize increases the size of the node group. To delete a node
// you need to explicitly name it and use DeleteNode. This function should
// wait until node group size is updated. Implementation required.
func (ng *brightboxNodeGroup) IncreaseSize(delta int) error {
	klog.V(4).Infof("IncreaseSize: %v", delta)
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := ng.TargetSize()
	if err != nil {
		return err
	}
	desiredSize := size + delta
	if desiredSize > ng.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", desiredSize, ng.MaxSize())
	}
	err = ng.createServers(delta)
	if err != nil {
		return err
	}
	return wait.Poll(
		checkInterval,
		checkTimeout,
		func() (bool, error) {
			size, err := ng.TargetSize()
			return err == nil && size >= desiredSize, err
		},
	)
}

// AtomicIncreaseSize is not implemented.
func (ng *brightboxNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group. Error is returned
// either on failure or if the given node doesn't belong to this
// node group. This function should wait until node group size is
// updated. Implementation required.
func (ng *brightboxNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(4).Info("DeleteNodes")
	klog.V(4).Infof("Nodes: %+v", nodes)
	for _, node := range nodes {
		size, err := ng.CurrentSize()
		if err != nil {
			return err
		}
		if size <= ng.MinSize() {
			return fmt.Errorf("min size reached, no further nodes will be deleted")
		}
		serverID := k8ssdk.MapProviderIDToServerID(node.Spec.ProviderID)
		err = ng.deleteServerFromGroup(serverID)
		if err != nil {
			return err
		}
	}
	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This
// function doesn't permit to delete any existing node and can be used
// only to reduce the request for new nodes that have not been yet
// fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes
// when there is an option to just decrease the target. Implementation
// required.
func (ng *brightboxNodeGroup) DecreaseTargetSize(delta int) error {
	klog.V(4).Infof("DecreaseTargetSize: %v", delta)
	if delta >= 0 {
		return fmt.Errorf("decrease size must be negative")
	}
	size, err := ng.TargetSize()
	if err != nil {
		return err
	}
	nodesize, err := ng.CurrentSize()
	if err != nil {
		return err
	}
	// Group size is synchronous at present, so this always fails
	if size+delta < nodesize {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, nodesize)
	}
	return fmt.Errorf("shouldn't have got here")
}

// Id returns an unique identifier of the node group.
func (ng *brightboxNodeGroup) Id() string {
	klog.V(4).Info("Id")
	return ng.id
}

// Debug returns a string containing all information regarding this
// node group.
func (ng *brightboxNodeGroup) Debug() string {
	klog.V(4).Info("Debug")
	return fmt.Sprintf("brightboxNodeGroup %+v", *ng)
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id
// field set.  Other fields are optional.
func (ng *brightboxNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	klog.V(4).Info("Nodes")
	group, err := ng.GetServerGroup(ng.Id())
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Found %d servers in group", len(group.Servers))
	nodes := make([]cloudprovider.Instance, len(group.Servers))
	for i, server := range group.Servers {
		cpStatus := cloudprovider.InstanceStatus{}
		switch server.Status {
		case status.Active:
			cpStatus.State = cloudprovider.InstanceRunning
		case status.Creating:
			cpStatus.State = cloudprovider.InstanceCreating
		case status.Deleting:
			cpStatus.State = cloudprovider.InstanceDeleting
		default:
			errorInfo := cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    server.Status,
				ErrorMessage: server.Status,
			}
			cpStatus.ErrorInfo = &errorInfo
		}
		nodes[i] = cloudprovider.Instance{
			Id:     k8ssdk.MapServerIDToProviderID(server.Id),
			Status: &cpStatus,
		}
	}
	klog.V(4).Infof("Created %d nodes", len(nodes))
	return nodes, nil
}

// Exist checks if the node group really exists on the cloud provider
// side. Allows to tell the theoretical node group from the real
// one. Implementation required.
func (ng *brightboxNodeGroup) Exist() bool {
	klog.V(4).Info("Exist")
	_, err := ng.GetServerGroup(ng.Id())
	return err == nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (ng *brightboxNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	klog.V(4).Info("TemplateNodeInfo")
	klog.V(4).Infof("Looking for server type %q", ng.serverOptions.ServerType)
	serverType, err := ng.findServerType()
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("ServerType %+v", serverType)
	// AllowedPodNumber is the kubelet default. The way to obtain that default programmatically
	// has been lost in a twisty maze of endless indirection.
	resources := &schedulerframework.Resource{
		MilliCPU:         int64(serverType.Cores * 1000),
		Memory:           int64(serverType.Ram * 1024 * 1024),
		EphemeralStorage: int64(serverType.DiskSize * 1024 * 1024),
		AllowedPodNumber: 110,
	}
	node := apiv1.Node{
		Status: apiv1.NodeStatus{
			Capacity:    resourceList(resources),
			Allocatable: resourceList(applyFudgeFactor(resources)),
			Conditions:  cloudprovider.BuildReadyConditions(),
		},
	}
	nodeInfo := framework.NewNodeInfo(&node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(ng.Id())})
	return nodeInfo, nil
}

// ResourceList returns a resource list of this resource.
func resourceList(r *schedulerframework.Resource) v1.ResourceList {
	result := v1.ResourceList{
		v1.ResourceCPU:              *resource.NewMilliQuantity(r.MilliCPU, resource.DecimalSI),
		v1.ResourceMemory:           *resource.NewQuantity(r.Memory, resource.BinarySI),
		v1.ResourcePods:             *resource.NewQuantity(int64(r.AllowedPodNumber), resource.BinarySI),
		v1.ResourceEphemeralStorage: *resource.NewQuantity(r.EphemeralStorage, resource.BinarySI),
	}
	for rName, rQuant := range r.ScalarResources {
		if v1helper.IsHugePageResourceName(rName) {
			result[rName] = *resource.NewQuantity(rQuant, resource.BinarySI)
		} else {
			result[rName] = *resource.NewQuantity(rQuant, resource.DecimalSI)
		}
	}
	return result
}

// Create creates the node group on the cloud provider
// side. Implementation optional.
func (ng *brightboxNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	klog.V(4).Info("Create")
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once
// their size drops to 0.  Implementation optional.
func (ng *brightboxNodeGroup) Delete() error {
	klog.V(4).Info("Delete")
	return cloudprovider.ErrNotImplemented
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (ng *brightboxNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An
// autoprovisioned group was created by CA and can be deleted when scaled
// to 0.
func (ng *brightboxNodeGroup) Autoprovisioned() bool {
	klog.V(4).Info("Autoprovisioned")
	return false
}

//private

func (ng *brightboxNodeGroup) findServerType() (*brightbox.ServerType, error) {
	handle := ng.serverOptions.ServerType
	if strings.HasPrefix(handle, "typ-") {
		return ng.GetServerType(handle)
	}
	servertypes, err := ng.GetServerTypes()
	if err != nil {
		return nil, err
	}
	for _, servertype := range servertypes {
		if servertype.Handle == handle {
			return &servertype, nil
		}
	}
	return nil, fmt.Errorf("ServerType with handle '%s' doesn't exist", handle)
}

func applyFudgeFactor(capacity *schedulerframework.Resource) *schedulerframework.Resource {
	allocatable := capacity.Clone()
	allocatable.Memory = max(0, capacity.Memory-max(capacity.Memory*memoryReservePercent/100, minimumMemoryReserve))
	allocatable.EphemeralStorage = max(0, capacity.EphemeralStorage-minimumDiskReserve)
	return allocatable
}

func makeNodeGroupFromAPIDetails(
	name string,
	mapData map[string]string,
	minSize int,
	maxSize int,
	cloudclient *k8ssdk.Cloud,
) (*brightboxNodeGroup, error) {
	klog.V(4).Info("makeNodeGroupFromApiDetails")
	if mapData["server_group"] == "" {
		return nil, cloudprovider.ErrIllegalConfiguration
	}
	ng := brightboxNodeGroup{
		id:      mapData["server_group"],
		minSize: minSize,
		maxSize: maxSize,
		Cloud:   cloudclient,
	}
	imageID := mapData["image"]
	if !(len(imageID) == 9 && strings.HasPrefix(imageID, "img-")) {
		image, err := ng.GetImageByName(imageID)
		if err != nil || image == nil {
			return nil, cloudprovider.ErrIllegalConfiguration
		}
		imageID = image.Id
	}
	userData := mapData["user_data"]
	options := &brightbox.ServerOptions{
		Image:        imageID,
		Name:         &name,
		ServerType:   mapData["type"],
		Zone:         mapData["zone"],
		UserData:     &userData,
		ServerGroups: mergeServerGroups(mapData),
	}
	ng.serverOptions = options
	klog.V(4).Info(ng.Debug())
	return &ng, nil
}

func mergeServerGroups(data map[string]string) []string {
	uniqueMap := map[string]bool{}
	addFromSplit(uniqueMap, data["server_group"])
	addFromSplit(uniqueMap, data["default_group"])
	addFromSplit(uniqueMap, data["additional_groups"])
	result := make([]string, 0, len(uniqueMap))
	for key := range uniqueMap {
		result = append(result, key)
	}
	return result
}

func addFromSplit(uniqueMap map[string]bool, source string) {
	for _, element := range strings.Split(source, ",") {
		if element != "" {
			uniqueMap[element] = true
		}
	}
}

func (ng *brightboxNodeGroup) createServers(amount int) error {
	klog.V(4).Infof("createServers: %d", amount)
	for i := 1; i <= amount; i++ {
		_, err := ng.CreateServer(ng.serverOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete the server and wait for the group details to be updated
func (ng *brightboxNodeGroup) deleteServerFromGroup(serverID string) error {
	klog.V(4).Infof("deleteServerFromGroup: %q", serverID)
	serverIDNotInGroup := func() (bool, error) {
		return ng.isMissing(serverID)
	}
	missing, err := serverIDNotInGroup()
	if err != nil {
		return err
	} else if missing {
		return fmt.Errorf("%s belongs to a different group than %s", serverID, ng.Id())
	}
	err = ng.DestroyServer(serverID)
	if err != nil {
		return err
	}
	return wait.Poll(
		checkInterval,
		checkTimeout,
		serverIDNotInGroup,
	)
}

func serverNotFoundError(id string) error {
	klog.V(4).Infof("serverNotFoundError: created for %q", id)
	return fmt.Errorf("Server %s not found", id)
}

func (ng *brightboxNodeGroup) isMissing(serverID string) (bool, error) {
	klog.V(4).Infof("isMissing: %q from %q", serverID, ng.Id())
	server, err := ng.GetServer(
		context.Background(),
		serverID,
		serverNotFoundError(serverID),
	)
	if err != nil {
		return false, err
	}
	if server.DeletedAt != nil {
		klog.V(4).Info("server deleted")
		return true, nil
	}
	for _, group := range server.ServerGroups {
		if group.Id == ng.Id() {
			return false, nil
		}
	}
	return true, nil
}
