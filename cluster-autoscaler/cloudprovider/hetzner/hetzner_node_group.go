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

package hetzner

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"strings"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

// hetznerNodeGroup implements cloudprovider.NodeGroup interface. hetznerNodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type hetznerNodeGroup struct {
	id           string
	manager      *hetznerManager
	minSize      int
	maxSize      int
	targetSize   int
	region       string
	instanceType string

	clusterUpdateMutex *sync.Mutex
}

type hetznerNodeGroupSpec struct {
	name         string
	minSize      int
	maxSize      int
	region       string
	instanceType string
}

// MaxSize returns maximum size of the node group.
func (n *hetznerNodeGroup) MaxSize() int {
	return n.maxSize
}

// MinSize returns minimum size of the node group.
func (n *hetznerNodeGroup) MinSize() int {
	return n.minSize
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (n *hetznerNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// TargetSize returns the current target size of the node group. It is possible
// that the number of nodes in Kubernetes is different at the moment but should
// be equal to Size() once everything stabilizes (new nodes finish startup and
// registration or removed nodes are deleted completely). Implementation
// required.
func (n *hetznerNodeGroup) TargetSize() (int, error) {
	return n.targetSize, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *hetznerNodeGroup) IncreaseSize(delta int) error {
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

	available, err := serverTypeAvailable(n.manager, n.instanceType, n.region)
	if err != nil {
		return fmt.Errorf("failed to check if type %s is available in region %s error: %v", n.instanceType, n.region, err)
	}
	if !available {
		return fmt.Errorf("server type %s not available in region %s", n.instanceType, n.region)
	}

	defer func() {
		// create new servers cache
		if _, err := n.manager.cachedServers.servers(); err != nil {
			klog.Errorf("failed to update servers cache: %v", err)
		}

		// Update target size
		n.resetTargetSize(actualDelta)
	}()

	// There is no "Server Group" in Hetzner Cloud, we need to create every
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
func (n *hetznerNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group (and also increasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *hetznerNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
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

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *hetznerNodeGroup) DecreaseTargetSize(delta int) error {
	n.targetSize = n.targetSize + delta
	return nil
}

// Id returns an unique identifier of the node group.
func (n *hetznerNodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *hetznerNodeGroup) Debug() string {
	return fmt.Sprintf("cluster ID: %s (min:%d max:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.  It is
// required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *hetznerNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	servers, err := n.manager.cachedServers.getServersByNodeGroupName(n.id)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers for hcloud: %v", err)
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
func (n *hetznerNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	resourceList, err := getMachineTypeResourceList(n.manager, n.instanceType)
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

	if n.manager.clusterConfig.IsUsingNewFormat {
		for _, taint := range n.manager.clusterConfig.NodeConfigs[n.id].Taints {
			node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: taint.Effect,
			})
		}
	}

	nodeInfo := framework.NewNodeInfo(&node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(n.id)})
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side.
// Allows to tell the theoretical node group from the real one. Implementation
// required.
func (n *hetznerNodeGroup) Exist() bool {
	_, exists := n.manager.nodeGroups[n.id]
	return exists
}

// Create creates the node group on the cloud provider side. Implementation
// optional.
func (n *hetznerNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	n.manager.nodeGroups[n.id] = n

	return n, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.  This will be
// executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *hetznerNodeGroup) Delete() error {
	// We do not use actual node groups but all nodes within the Hcloud project are labeled with a group
	return nil
}

// Autoprovisioned returns true if the node group is autoprovisioned. An
// autoprovisioned group was created by CA and can be deleted when scaled to 0.
func (n *hetznerNodeGroup) Autoprovisioned() bool {
	// All groups are auto provisioned
	return false
}

func toInstance(vm *hcloud.Server) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     toProviderID(vm.ID),
		Status: toInstanceStatus(vm.Status),
	}
}

func toProviderID(nodeID int64) string {
	return fmt.Sprintf("%s%d", providerIDPrefix, nodeID)
}

func toInstanceStatus(status hcloud.ServerStatus) *cloudprovider.InstanceStatus {
	if status == "" {
		return nil
	}

	st := &cloudprovider.InstanceStatus{}
	switch status {
	case hcloud.ServerStatusInitializing:
	case hcloud.ServerStatusStarting:
		st.State = cloudprovider.InstanceCreating
	case hcloud.ServerStatusRunning:
		st.State = cloudprovider.InstanceRunning
	case hcloud.ServerStatusOff:
	case hcloud.ServerStatusDeleting:
	case hcloud.ServerStatusStopping:
		st.State = cloudprovider.InstanceDeleting
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "no-code-hcloud",
			ErrorMessage: "error",
		}
	}

	return st
}

func newNodeName(n *hetznerNodeGroup) string {
	return fmt.Sprintf("%s-%x", n.id, rand.Int63())
}

func buildNodeGroupLabels(n *hetznerNodeGroup) (map[string]string, error) {
	archLabel, err := instanceTypeArch(n.manager, n.instanceType)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Build node group label for %s", n.id)

	labels := map[string]string{
		apiv1.LabelInstanceType:      n.instanceType,
		apiv1.LabelTopologyRegion:    n.region,
		apiv1.LabelArchStable:        archLabel,
		"csi.hetzner.cloud/location": n.region,
		nodeGroupLabel:               n.id,
	}

	if n.manager.clusterConfig.IsUsingNewFormat {
		maps.Copy(labels, n.manager.clusterConfig.NodeConfigs[n.id].Labels)
	}

	klog.V(4).Infof("%s nodegroup labels: %s", n.id, labels)

	return labels, nil
}

func getMachineTypeResourceList(m *hetznerManager, instanceType string) (apiv1.ResourceList, error) {
	typeInfo, err := m.cachedServerType.getServerType(instanceType)
	if err != nil || typeInfo == nil {
		return nil, fmt.Errorf("failed to get machine type %s info error: %v", instanceType, err)
	}

	return apiv1.ResourceList{
		// TODO somehow determine the actual pods that will be running
		apiv1.ResourcePods:             *resource.NewQuantity(defaultPodAmountsLimit, resource.DecimalSI),
		apiv1.ResourceCPU:              *resource.NewQuantity(int64(typeInfo.Cores), resource.DecimalSI),
		apiv1.ResourceMemory:           *resource.NewQuantity(int64(typeInfo.Memory*1024*1024*1024), resource.DecimalSI),
		apiv1.ResourceEphemeralStorage: *resource.NewQuantity(int64(typeInfo.Disk*1024*1024*1024), resource.DecimalSI),
	}, nil
}

func serverTypeAvailable(manager *hetznerManager, instanceType string, region string) (bool, error) {
	serverType, err := manager.cachedServerType.getServerType(instanceType)
	if err != nil {
		return false, err
	}

	for _, price := range serverType.Pricings {
		if price.Location.Name == region {
			return true, nil
		}
	}

	return false, nil
}

func instanceTypeArch(manager *hetznerManager, instanceType string) (string, error) {
	serverType, err := manager.cachedServerType.getServerType(instanceType)
	if err != nil {
		return "", err
	}

	switch serverType.Architecture {
	case hcloud.ArchitectureARM:
		return "arm64", nil
	case hcloud.ArchitectureX86:
		return "amd64", nil
	default:
		return "amd64", nil
	}
}

func createServer(n *hetznerNodeGroup) error {
	ctx, cancel := context.WithTimeout(n.manager.apiCallContext, n.manager.createTimeout)
	defer cancel()

	serverType, err := n.manager.cachedServerType.getServerType(n.instanceType)
	if err != nil {
		return err
	}

	image, err := findImage(n, serverType)
	if err != nil {
		return err
	}

	cloudInit := n.manager.clusterConfig.LegacyConfig.CloudInit

	if n.manager.clusterConfig.IsUsingNewFormat {
		cloudInit = n.manager.clusterConfig.NodeConfigs[n.id].CloudInit
	}

	StartAfterCreate := true
	opts := hcloud.ServerCreateOpts{
		Name:             newNodeName(n),
		UserData:         cloudInit,
		Location:         &hcloud.Location{Name: n.region},
		ServerType:       serverType,
		Image:            image,
		StartAfterCreate: &StartAfterCreate,
		Labels: map[string]string{
			nodeGroupLabel: n.id,
		},
		PublicNet: &hcloud.ServerCreatePublicNet{
			EnableIPv4: n.manager.publicIPv4,
			EnableIPv6: n.manager.publicIPv6,
		},
	}
	if n.manager.sshKey != nil {
		opts.SSHKeys = []*hcloud.SSHKey{n.manager.sshKey}
	}
	if n.manager.network != nil {
		opts.Networks = []*hcloud.Network{n.manager.network}
	}
	if n.manager.firewall != nil {
		serverCreateFirewall := &hcloud.ServerCreateFirewall{Firewall: *n.manager.firewall}
		opts.Firewalls = []*hcloud.ServerCreateFirewall{serverCreateFirewall}
	}

	serverCreateResult, _, err := n.manager.client.Server.Create(ctx, opts)
	if err != nil {
		return fmt.Errorf("could not create server type %s in region %s: %v", n.instanceType, n.region, err)
	}

	server := serverCreateResult.Server

	actions := append(serverCreateResult.NextActions, serverCreateResult.Action)

	// Delete the server if any action (most importantly create_server & start_server) fails
	err = n.manager.client.Action.WaitFor(ctx, actions...)
	if err != nil {
		_ = n.manager.deleteServer(server)
		return fmt.Errorf("failed to start server %s error: %v", server.Name, err)
	}

	return nil
}

// findImage searches for an image ID corresponding to the supplied
// HCLOUD_IMAGE env variable. This value can either be an image ID itself (an
// int), a name (e.g. "ubuntu-20.04"), or a label selector associated with an
// image snapshot. In the latter case it will use the most recent snapshot.
// It also verifies that the returned image has a compatible architecture with
// server.
func findImage(n *hetznerNodeGroup, serverType *hcloud.ServerType) (*hcloud.Image, error) {
	// Select correct image based on server type architecture
	imageName := n.manager.clusterConfig.LegacyConfig.ImageName
	if n.manager.clusterConfig.IsUsingNewFormat {
		if serverType.Architecture == hcloud.ArchitectureARM {
			imageName = n.manager.clusterConfig.ImagesForArch.Arm64
		}

		if serverType.Architecture == hcloud.ArchitectureX86 {
			imageName = n.manager.clusterConfig.ImagesForArch.Amd64
		}
	}

	image, _, err := n.manager.client.Image.GetForArchitecture(context.TODO(), imageName, serverType.Architecture)
	if err != nil {
		// Keep looking for label if image was not found by id or name
		if !strings.HasPrefix(err.Error(), "image not found") {
			return nil, err
		}
	}

	if image != nil {
		return image, nil
	}

	// Look for snapshot with label
	images, err := n.manager.client.Image.AllWithOpts(context.TODO(), hcloud.ImageListOpts{
		Type:         []hcloud.ImageType{hcloud.ImageTypeSnapshot},
		Status:       []hcloud.ImageStatus{hcloud.ImageStatusAvailable},
		Sort:         []string{"created:desc"},
		Architecture: []hcloud.Architecture{serverType.Architecture},
		ListOpts: hcloud.ListOpts{
			LabelSelector: imageName,
		},
	})

	if err != nil || len(images) == 0 {
		return nil, fmt.Errorf("unable to find image %s with architecture %s: %v", imageName, serverType.Architecture, err)
	}

	return images[0], nil
}

func (n *hetznerNodeGroup) resetTargetSize(expectedDelta int) {
	servers, err := n.manager.allServers(n.id)
	if err != nil {
		klog.Warningf("failed to set node pool %s size, using delta %d error: %v", n.id, expectedDelta, err)
		n.targetSize = n.targetSize + expectedDelta
	} else {
		klog.Infof("Set node group %s size from %d to %d, expected delta %d", n.id, n.targetSize, len(servers), expectedDelta)
		n.targetSize = len(servers)
	}
}
