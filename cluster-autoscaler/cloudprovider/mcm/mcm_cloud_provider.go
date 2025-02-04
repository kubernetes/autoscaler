/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This file was copied and modified from the kubernetes/autoscaler project
https://github.com/kubernetes/autoscaler/blob/cluster-autoscaler-release-1.1/cluster-autoscaler/cloudprovider/aws/aws_cloud_provider.go

Modifications Copyright (c) 2017 SAP SE or an SAP affiliate company. All rights reserved.
*/

package mcm

import (
	"context"
	"fmt"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/eligibility"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/machineutils"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// ProviderName is the cloud provider name for MCM
	ProviderName = "mcm"

	// GPULabel is the label added to nodes with GPU resource.
	// TODO: Align on a GPU Label for Gardener.
	GPULabel = "gardener.cloud/accelerator"

	// ScaleDownUtilizationThresholdAnnotation is the annotation key for the value of NodeGroupAutoscalingOptions.ScaleDownUtilizationThreshold
	ScaleDownUtilizationThresholdAnnotation = "autoscaler.gardener.cloud/scale-down-utilization-threshold"

	// ScaleDownGpuUtilizationThresholdAnnotation is the annotation key for the value of NodeGroupAutoscalingOptions.ScaleDownGpuUtilizationThreshold
	ScaleDownGpuUtilizationThresholdAnnotation = "autoscaler.gardener.cloud/scale-down-gpu-utilization-threshold"

	// ScaleDownUnneededTimeAnnotation is the annotation key for the value of NodeGroupAutoscalingOptions.ScaleDownUnneededTime
	ScaleDownUnneededTimeAnnotation = "autoscaler.gardener.cloud/scale-down-unneeded-time"

	// ScaleDownUnreadyTimeAnnotation is the annotation key for the value of NodeGroupAutoscalingOptions.ScaleDownUnreadyTime
	ScaleDownUnreadyTimeAnnotation = "autoscaler.gardener.cloud/scale-down-unready-time"

	// MaxNodeProvisionTimeAnnotation is the annotation key for the value of NodeGroupAutoscalingOptions.MaxNodeProvisionTime
	MaxNodeProvisionTimeAnnotation = "autoscaler.gardener.cloud/max-node-provision-time"
)

// MCMCloudProvider implements the cloud provider interface for machine-controller-manager
// Reference: https://github.com/gardener/machine-controller-manager
type mcmCloudProvider struct {
	mcmManager      *McmManager
	resourceLimiter *cloudprovider.ResourceLimiter
}

var _ cloudprovider.CloudProvider = (*mcmCloudProvider)(nil)

// BuildMcmCloudProvider builds CloudProvider implementation for machine-controller-manager.
func BuildMcmCloudProvider(mcmManager *McmManager, resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	if mcmManager.discoveryOpts.StaticDiscoverySpecified() {
		return buildStaticallyDiscoveringProvider(mcmManager, resourceLimiter)
	}
	return nil, fmt.Errorf("Failed to build an mcm cloud provider: Either node group specs or node group auto discovery spec must be specified")
}

// BuildMCM builds the MCM provider and MCMmanager.
func BuildMCM(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var mcmManager *McmManager
	var err error
	mcmManager, err = CreateMcmManager(do)

	if err != nil {
		klog.Fatalf("Failed to create MCM Manager: %v", err)
	}
	provider, err := BuildMcmCloudProvider(mcmManager, rl)
	if err != nil {
		klog.Fatalf("Failed to create MCM cloud provider: %v", err)
	}
	return provider
}

func buildStaticallyDiscoveringProvider(mcmManager *McmManager, resourceLimiter *cloudprovider.ResourceLimiter) (*mcmCloudProvider, error) {
	mcm := &mcmCloudProvider{
		mcmManager:      mcmManager,
		resourceLimiter: resourceLimiter,
	}
	return mcm, nil
}

// Cleanup stops the go routine that is handling the current view of the nodeGroup in the form of a cache
func (mcm *mcmCloudProvider) Cleanup() error {
	mcm.mcmManager.Cleanup()
	return nil
}

func (mcm *mcmCloudProvider) Name() string {
	return "machine-controller-manager"
}

// NodeGroups returns all node groups configured for this cloud provider.
func (mcm *mcmCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0, len(mcm.mcmManager.nodeGroups))
	for _, ng := range mcm.mcmManager.nodeGroups {
		if ng.maxSize == 0 {
			continue
		}
		result = append(result, ng)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (mcm *mcmCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	if len(node.Spec.ProviderID) == 0 {
		klog.Warningf("Node %v has no providerId", node.Name)
		return nil, nil
	}

	mInfo, err := mcm.mcmManager.getMachineInfo(node)
	if err != nil {
		return nil, err
	}

	if mInfo == nil {
		klog.V(4).Infof("Skipped node %v, it's either been removed or it's not managed by this controller", node.Spec.ProviderID)
		return nil, nil
	}

	ng, err := mcm.mcmManager.getNodeGroup(mInfo.Key)
	if err != nil {
		return nil, err
	}

	key := types.NamespacedName{Namespace: ng.Namespace, Name: ng.Name}
	_, isManaged := mcm.mcmManager.nodeGroups[key]
	if !isManaged {
		klog.V(4).Infof("Skipped node %v, it's not managed by this controller", node.Spec.ProviderID)
		return nil, nil
	}

	return ng, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (mcm *mcmCloudProvider) HasInstance(*apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (mcm *mcmCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (mcm *mcmCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (mcm *mcmCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (mcm *mcmCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return mcm.resourceLimiter, nil
}

func (mcm *mcmCloudProvider) checkMCMAvailableReplicas() error {
	namespace := mcm.mcmManager.namespace
	deployment, err := mcm.mcmManager.deploymentLister.Deployments(namespace).Get("machine-controller-manager")
	if err != nil {
		return fmt.Errorf("failed to get machine-controller-manager deployment: %v", err.Error())
	}

	if deployment.Status.AvailableReplicas == 0 {
		return fmt.Errorf("machine-controller-manager is offline. Cluster autoscaler operations would be suspended.")
	}

	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (mcm *mcmCloudProvider) Refresh() error {
	err := mcm.checkMCMAvailableReplicas()
	if err != nil {
		return err
	}
	return mcm.mcmManager.Refresh()
}

// GPULabel returns the label added to nodes with GPU resource.
func (mcm *mcmCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (mcm *mcmCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (mcm *mcmCloudProvider) GetNodeGpuConfig(*apiv1.Node) *cloudprovider.GpuConfig {
	return nil
}

// nodeGroup implements NodeGroup interface.
type nodeGroup struct {
	types.NamespacedName

	mcmManager *McmManager

	scalingMutex sync.Mutex
	minSize      int
	maxSize      int
}

var _ cloudprovider.NodeGroup = (*nodeGroup)(nil)

// MaxSize returns maximum size of the node group.
func (ngImpl *nodeGroup) MaxSize() int {
	return ngImpl.maxSize
}

// MinSize returns minimum size of the node group.
func (ngImpl *nodeGroup) MinSize() int {
	return ngImpl.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (ngImpl *nodeGroup) TargetSize() (int, error) {
	size, err := ngImpl.mcmManager.GetMachineDeploymentSize(ngImpl.Name)
	return int(size), err
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
// TODO: Implement this to check if machine-deployment really exists.
func (ngImpl *nodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (ngImpl *nodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (ngImpl *nodeGroup) Autoprovisioned() bool {
	return false
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (ngImpl *nodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// IncreaseSize of the Machinedeployment.
func (ngImpl *nodeGroup) IncreaseSize(delta int) error {
	klog.V(0).Infof("Received request to increase size of machine deployment %s by %d", ngImpl.Name, delta)
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	release := ngImpl.AcquireScalingMutex(fmt.Sprintf("IncreaseSize by #%d", delta))
	defer release()
	size, err := ngImpl.mcmManager.GetMachineDeploymentSize(ngImpl.Name)
	if err != nil {
		return err
	}
	targetSize := int(size) + delta
	if targetSize > ngImpl.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", targetSize, ngImpl.MaxSize())
	}
	return ngImpl.mcmManager.retry(func(ctx context.Context) (bool, error) {
		return ngImpl.mcmManager.SetMachineDeploymentSize(ctx, ngImpl, int64(targetSize))
	}, "MachineDeployment", "update", ngImpl.Name)
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (ngImpl *nodeGroup) DecreaseTargetSize(delta int) error {
	klog.V(0).Infof("Received request to decrease target size of machine deployment %s by %d", ngImpl.Name, delta)
	if delta >= 0 {
		return fmt.Errorf("size decrease size must be negative")
	}
	release := ngImpl.AcquireScalingMutex(fmt.Sprintf("DecreaseTargetSize by #%d", delta))
	defer release()
	size, err := ngImpl.mcmManager.GetMachineDeploymentSize(ngImpl.Name)
	if err != nil {
		return err
	}
	decreaseAmount := int(size) + delta
	if decreaseAmount < ngImpl.minSize {
		klog.Warningf("Cannot go below min size= %d for ngImpl %s, requested target size= %d . Setting target size to min size", ngImpl.minSize, ngImpl.Name, size+int64(delta))
		decreaseAmount = ngImpl.minSize
	}
	return ngImpl.mcmManager.retry(func(ctx context.Context) (bool, error) {
		return ngImpl.mcmManager.SetMachineDeploymentSize(ctx, ngImpl, int64(decreaseAmount))
	}, "MachineDeployment", "update", ngImpl.Name)
}

// Refresh cordons the Nodes corresponding to the machines that have been marked for deletion in the TriggerDeletionByMCM annotation on the MachineDeployment
func (ngImpl *nodeGroup) Refresh() error {
	mcd, err := ngImpl.mcmManager.GetMachineDeploymentObject(ngImpl.Name)
	if err != nil {
		return err
	}
	toBeDeletedMachineNames := getMachineNamesTriggeredForDeletion(mcd)
	if len(toBeDeletedMachineNames) == 0 {
		return nil
	}
	machinesOfNodeGroup, err := ngImpl.mcmManager.getMachinesForMachineDeployment(ngImpl.Name)
	if err != nil {
		klog.Warningf("NodeGroup.Refresh() of %q failed to get machines for MachineDeployment due to: %v", ngImpl.Name, err)
		return nil
	}
	toBeDeletedMachines := filterMachinesMatchingNames(machinesOfNodeGroup, sets.New(toBeDeletedMachineNames...))
	if len(toBeDeletedMachines) == 0 {
		klog.Warningf("NodeGroup.Refresh() of %q could not find Machine objects for toBeDeletedMachineNames %q", ngImpl.Name, toBeDeletedMachineNames)
		return nil
	}
	toBeDeletedNodeNames := getNodeNamesFromMachines(toBeDeletedMachines)
	if len(toBeDeletedNodeNames) == 0 {
		klog.Warningf("NodeGroup.Refresh() of %q could not find toBeDeletedNodeNames for toBeDeletedMachineNames %q of MachineDeployment", ngImpl.Name, toBeDeletedMachineNames)
		return nil
	}
	err = ngImpl.mcmManager.cordonNodes(toBeDeletedNodeNames)
	if err != nil {
		// we do not return error since we don't want this to block CA operation. This is best-effort.
		klog.Warningf("NodeGroup.Refresh() of %q ran into error cordoning nodes: %v", ngImpl.Name, err)
	}
	return nil
}

// belongs checks if the given node belongs to this NodeGroup and also returns its MachineInfo for its corresponding Machine
func (ngImpl *nodeGroup) belongs(node *apiv1.Node) (belongs bool, mInfo *machineInfo, err error) {
	mInfo, err = ngImpl.mcmManager.getMachineInfo(node)
	if err != nil || mInfo == nil {
		return
	}
	targetNg, err := ngImpl.mcmManager.getNodeGroup(mInfo.Key)
	if err != nil {
		return
	}
	if targetNg == nil {
		err = fmt.Errorf("%s doesn't belong to a known MachinDeployment", node.Name)
		return
	}
	if targetNg.Id() == ngImpl.Id() {
		belongs = true
	}
	return
}

// DeleteNodes deletes the nodes from the group. It is expected that this method will not be called
// for nodes which are not part of ANY machine deployment.
func (ngImpl *nodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(0).Infof("for NodeGroup %q, Received request to delete nodes:- %v", ngImpl.Name, getNodeNames(nodes))
	size, err := ngImpl.mcmManager.GetMachineDeploymentSize(ngImpl.Name)
	if err != nil {
		return err
	}
	if int(size) <= ngImpl.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	var toBeDeletedMachineInfos = make([]machineInfo, 0, len(nodes))
	for _, node := range nodes {
		belongs, mInfo, err := ngImpl.belongs(node)
		if err != nil {
			return err
		} else if !belongs {
			return fmt.Errorf("%s belongs to a different MachineDeployment than %q", node.Name, ngImpl.Name)
		}
		if mInfo.FailedOrTerminating {
			klog.V(3).Infof("for NodeGroup %q, Machine %q is already marked as terminating - skipping deletion", ngImpl.Name, mInfo.Key.Name)
			continue
		}
		if eligibility.HasNoScaleDownAnnotation(node) {
			klog.V(4).Infof("for NodeGroup %q, Node %q corresponding to Machine %q is marked with ScaleDownDisabledAnnotation %q - skipping deletion", ngImpl.Name, node.Name, mInfo.Key.Name, eligibility.ScaleDownDisabledKey)
			continue
		}
		toBeDeletedMachineInfos = append(toBeDeletedMachineInfos, *mInfo)
	}
	return ngImpl.deleteMachines(toBeDeletedMachineInfos)
}

// deleteMachines annotates the corresponding MachineDeployment with machine names of toBeDeletedMachineInfos, reduces the desired replicas of the corresponding MachineDeployment and cordons corresponding nodes belonging to toBeDeletedMachineInfos
func (ngImpl *nodeGroup) deleteMachines(toBeDeletedMachineInfos []machineInfo) error {
	numDeletionCandidates := len(toBeDeletedMachineInfos)
	if numDeletionCandidates == 0 {
		return nil
	}

	release := ngImpl.AcquireScalingMutex(fmt.Sprintf("deleteMachines for %s", toBeDeletedMachineInfos))
	defer release()

	// get the machine deployment and return if rolling update is not finished
	md, err := ngImpl.mcmManager.GetMachineDeploymentObject(ngImpl.Name)
	if err != nil {
		return err
	}
	if !isRollingUpdateFinished(md) {
		return fmt.Errorf("MachineDeployment %s is under rolling update , cannot reduce replica count", ngImpl.Name)
	}

	// Trying to update the machineDeployment till the deadline
	err = ngImpl.mcmManager.retry(func(ctx context.Context) (bool, error) {
		return ngImpl.mcmManager.scaleDownMachineDeployment(ctx, ngImpl.Name, toBeDeletedMachineInfos)
	}, "MachineDeployment", "update", ngImpl.Name)
	if err != nil {
		klog.Errorf("Unable to scale down MachineDeployment %s by %d and delete machines %q due to: %v", ngImpl.Name, numDeletionCandidates, toBeDeletedMachineInfos, err)
		return fmt.Errorf("for NodeGroup %q, cannot scale down due to: %w", ngImpl.Name, err)
	}
	return nil
}

// AcquireScalingMutex acquires the scalingMutex associated with this NodeGroup and returns a function that releases the scalingMutex that is expected to be deferred by the caller.
func (ngImpl *nodeGroup) AcquireScalingMutex(operation string) (releaseFn func()) {
	ngImpl.scalingMutex.Lock()
	klog.V(3).Infof("%s has acquired scalingMutex of NodeGroup %q", operation, ngImpl.Name)
	releaseFn = func() {
		klog.V(3).Infof("%s is releasing scalingMutex of NodeGroup %q", operation, ngImpl.Name)
		ngImpl.scalingMutex.Unlock()
	}
	return
}

func getNodeNames(nodes []*apiv1.Node) []string {
	nodeNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}
	return nodeNames
}

func getNodeNamesFromMachines(machines []*v1alpha1.Machine) []string {
	var nodeNames []string
	for _, m := range machines {
		nodeName := m.Labels["node"]
		if nodeName != "" {
			nodeNames = append(nodeNames, nodeName)
		}
	}
	return nodeNames
}

// Id returns NodeGroup name
func (ngImpl *nodeGroup) Id() string {
	return ngImpl.Name
}

// Debug returns a debug string for the Asg.
func (ngImpl *nodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", ngImpl.Id(), ngImpl.MinSize(), ngImpl.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (ngImpl *nodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	instances, err := ngImpl.mcmManager.GetInstancesForMachineDeployment(ngImpl.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get the cloudprovider.Instance for machines backed by the MachineDeployment %q, error: %v", ngImpl.Name, err)
	}
	erroneousInstanceInfos := make([]string, 0, len(instances))
	for _, instance := range instances {
		if instance.Status != nil && instance.Status.ErrorInfo != nil {
			erroneousInstanceInfos = append(erroneousInstanceInfos, fmt.Sprintf("[Instance: %s, ErrorClass: %s, ErrorCode: %s]", instance.Id, instance.Status.ErrorInfo.ErrorClass.String(), instance.Status.ErrorInfo.ErrorCode))
		}
	}
	if len(erroneousInstanceInfos) != 0 {
		klog.V(0).Infof("Found erroneous instances:- %v", erroneousInstanceInfos)
	}
	return instances, nil
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
// Implementation optional.
func (ngImpl *nodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	options := defaults
	mcdAnnotations, err := ngImpl.mcmManager.GetMachineDeploymentAnnotations(ngImpl.Name)
	if err != nil {
		return nil, err
	}

	if _, ok := mcdAnnotations[ScaleDownUtilizationThresholdAnnotation]; ok {
		if floatVal, err := strconv.ParseFloat(mcdAnnotations[ScaleDownUtilizationThresholdAnnotation], 64); err == nil {
			options.ScaleDownUtilizationThreshold = floatVal
		}
	}
	if _, ok := mcdAnnotations[ScaleDownGpuUtilizationThresholdAnnotation]; ok {
		if floatVal, err := strconv.ParseFloat(mcdAnnotations[ScaleDownGpuUtilizationThresholdAnnotation], 64); err == nil {
			options.ScaleDownGpuUtilizationThreshold = floatVal
		}
	}
	if _, ok := mcdAnnotations[ScaleDownUnneededTimeAnnotation]; ok {
		if durationVal, err := time.ParseDuration(mcdAnnotations[ScaleDownUnneededTimeAnnotation]); err == nil {
			options.ScaleDownUnneededTime = durationVal
		}
	}
	if _, ok := mcdAnnotations[ScaleDownUnreadyTimeAnnotation]; ok {
		if durationVal, err := time.ParseDuration(mcdAnnotations[ScaleDownUnreadyTimeAnnotation]); err == nil {
			options.ScaleDownUnreadyTime = durationVal
		}
	}
	if _, ok := mcdAnnotations[MaxNodeProvisionTimeAnnotation]; ok {
		if durationVal, err := time.ParseDuration(mcdAnnotations[MaxNodeProvisionTimeAnnotation]); err == nil {
			options.MaxNodeProvisionTime = durationVal
		}
	}
	return &options, nil
}

// TemplateNodeInfo returns a node template for this node group.
func (ngImpl *nodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {

	nodeTemplate, err := ngImpl.mcmManager.GetMachineDeploymentNodeTemplate(ngImpl.Name)
	if err != nil {
		return nil, err
	}

	node, err := ngImpl.mcmManager.buildNodeFromTemplate(ngImpl.Name, nodeTemplate)
	if err != nil {
		return nil, err
	}

	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(ngImpl.Name))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

// AtomicIncreaseSize is not implemented.
func (ngImpl *nodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// getMachineNamesTriggeredForDeletion returns the set of machine names contained within the machineutils.TriggerDeletionByMCM annotation on the given MachineDeployment
// TODO: Move to using MCM annotations.GetMachineNamesTriggeredForDeletion after MCM release.
func getMachineNamesTriggeredForDeletion(mcd *v1alpha1.MachineDeployment) []string {
	if mcd.Annotations == nil || mcd.Annotations[machineutils.TriggerDeletionByMCM] == "" {
		return nil
	}
	return strings.Split(mcd.Annotations[machineutils.TriggerDeletionByMCM], ",")
}

// TODO: Move to using MCM annotations.CreateMachinesTriggeredForDeletionAnnotValue after MCM release
func createMachinesTriggeredForDeletionAnnotValue(machineNames []string) string {
	slices.Sort(machineNames)
	return strings.Join(machineNames, ",")
}
