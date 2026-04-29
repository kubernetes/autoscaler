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

package clusterapi

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"k8s.io/klog/v2"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

const (
	debugFormat = "%s (min: %d, max: %d, replicas: %d)"

	// The default for the maximum number of pods is inspired by the Kubernetes
	// best practices documentation for large clusters.
	// see https://kubernetes.io/docs/setup/best-practices/cluster-large/
	defaultMaxPods = 110
)

type nodegroup struct {
	machineController *machineController
	scalableResource  *unstructuredScalableResource
}

var _ cloudprovider.NodeGroup = (*nodegroup)(nil)

func (ng *nodegroup) MinSize() int {
	return ng.scalableResource.MinSize()
}

func (ng *nodegroup) MaxSize() int {
	return ng.scalableResource.MaxSize()
}

// TargetSize returns the current target size of the node group. It is
// possible that the number of nodes in Kubernetes is different at the
// moment but should be equal to Size() once everything stabilizes
// (new nodes finish startup and registration or removed nodes are
// deleted completely). Implementation required.
func (ng *nodegroup) TargetSize() (int, error) {
	replicas, found, err := unstructured.NestedInt64(ng.scalableResource.unstructured.Object, "spec", "replicas")
	if err != nil {
		return 0, errors.Wrap(err, "error getting replica count")
	}
	if !found {
		return 0, fmt.Errorf("unable to find replicas")
	}
	return int(replicas), nil
}

// IncreaseSize increases the size of the node group. To delete a node
// you need to explicitly name it and use DeleteNode. This function
// should wait until node group size is updated. Implementation
// required.
func (ng *nodegroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}

	size, err := ng.scalableResource.Replicas()
	if err != nil {
		return err
	}

	return ng.scalableResource.SetSize(size + delta)
}

// AtomicIncreaseSize is not implemented.
func (ng *nodegroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group. Error is returned
// either on failure or if the given node doesn't belong to this node
// group. This function should wait until node group size is updated.
// Implementation required.
func (ng *nodegroup) DeleteNodes(nodes []*corev1.Node) error {
	ng.machineController.accessLock.Lock()
	defer ng.machineController.accessLock.Unlock()

	replicas, err := ng.scalableResource.Replicas()
	if err != nil {
		return err
	}

	// if we are at minSize already we wail early.
	if replicas <= ng.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}

	// Step 1: Verify all nodes belong to this node group.
	for _, node := range nodes {
		actualNodeGroup, err := ng.machineController.nodeGroupForNode(node)
		if err != nil {
			return nil
		}

		if actualNodeGroup == nil {
			return fmt.Errorf("no node group found for node %q", node.Spec.ProviderID)
		}

		if actualNodeGroup.Id() != ng.Id() {
			return fmt.Errorf("node %q doesn't belong to node group %q", node.Spec.ProviderID, ng.Id())
		}
	}

	// Step 2: if deleting len(nodes) would make the replica count
	// < minSize, then the request to delete that many nodes is bogus
	// and we fail fast.
	if replicas-len(nodes) < ng.MinSize() {
		return fmt.Errorf("unable to delete %d machines in %q, machine replicas are %q, minSize is %q ", len(nodes), ng.Id(), replicas, ng.MinSize())
	}

	// Step 3: annotate the corresponding machine that it is a
	// suitable candidate for deletion and drop the replica count
	// by 1. Fail fast on any error.
	for _, node := range nodes {
		machine, err := ng.machineController.findMachineByProviderID(normalizedProviderString(node.Spec.ProviderID))
		if err != nil {
			return err
		}
		if machine == nil {
			return fmt.Errorf("unknown machine for node %q", node.Spec.ProviderID)
		}

		machine = machine.DeepCopy()

		if !machine.GetDeletionTimestamp().IsZero() {
			// The machine for this node is already being deleted
			continue
		}

		nodeGroup, err := ng.machineController.nodeGroupForNode(node)
		if err != nil {
			return err
		}

		if err := nodeGroup.scalableResource.MarkMachineForDeletion(machine); err != nil {
			return err
		}

		if err := ng.scalableResource.SetSize(replicas - 1); err != nil {
			_ = nodeGroup.scalableResource.UnmarkMachineForDeletion(machine)
			return err
		}

		replicas--
	}

	return nil
}

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
func (ng *nodegroup) ForceDeleteNodes(nodes []*corev1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group.
// This function doesn't permit to delete any existing node and can be
// used only to reduce the request for new nodes that have not been
// yet fulfilled. Delta should be negative. It is assumed that cloud
// nodegroup will not delete the existing nodes when there is an option
// to just decrease the target. Implementation required.
func (ng *nodegroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}

	size, err := ng.scalableResource.Replicas()
	if err != nil {
		return err
	}

	nodes, err := ng.Nodes()
	if err != nil {
		return err
	}

	// we want to filter out machines that are not nodes (eg failed or pending)
	// so that the node group size can be set properly. this affects situations
	// where an instance is created, but cannot become a node due to cloud provider
	// issues such as quota limitations, and thus the autoscaler needs to properly
	// set the size of the node group. without this adjustment, the core autoscaler
	// will become confused about the state of nodes and instances in the clusterapi
	// provider.
	actualNodes := 0
	for _, node := range nodes {
		if isProviderIDNormalized(normalizedProviderID(node.Id)) {
			actualNodes += 1
		}
	}

	if size+delta < actualNodes {
		return fmt.Errorf("node group %s: attempt to delete existing nodes currentReplicas:%d delta:%d existingNodes: %d",
			ng.scalableResource.Name(), size, delta, actualNodes)
	}

	klog.V(4).Infof("%s: DecreaseTargetSize: scaling down: currentReplicas: %d, delta: %d, existingNodes: %d", ng.scalableResource.Name(), size, delta, len(nodes))
	return ng.scalableResource.SetSize(size + delta)
}

// Id returns an unique identifier of the node group.
func (ng *nodegroup) Id() string {
	return ng.scalableResource.ID()
}

// Debug returns a string containing all information regarding this node group.
func (ng *nodegroup) Debug() string {
	replicas, err := ng.scalableResource.Replicas()
	if err != nil {
		return fmt.Sprintf("%s (min: %d, max: %d, replicas: %v)", ng.Id(), ng.MinSize(), ng.MaxSize(), err)
	}
	return fmt.Sprintf(debugFormat, ng.Id(), ng.MinSize(), ng.MaxSize(), replicas)
}

// Nodes returns a list of all nodes that belong to this node group.
// This includes instances that might have not become a kubernetes node yet.
func (ng *nodegroup) Nodes() ([]cloudprovider.Instance, error) {
	providerIDs, err := ng.scalableResource.ProviderIDs()
	if err != nil {
		return nil, err
	}

	// Nodes do not have normalized IDs, so do not normalize the ID here.
	// The IDs returned here are used to check if a node is registered or not and
	// must match the ID on the Node object itself.
	// https://github.com/kubernetes/autoscaler/blob/a973259f1852303ba38a3a61eeee8489cf4e1b13/cluster-autoscaler/clusterstate/clusterstate.go#L967-L985
	instances := make([]cloudprovider.Instance, len(providerIDs))
	for i, providerID := range providerIDs {
		providerIDNormalized := normalizedProviderID(providerID)

		// Add instance Status to report instance state to cluster-autoscaler.
		// This helps cluster-autoscaler make better scaling decisions.
		//
		// Machine can be Failed for a variety of reasons, here we are looking for a specific errors:
		// - Failed to provision
		// - Failed to delete.
		// Other reasons for a machine to be in a Failed state are not forwarding Status to the autoscaler.
		var status *cloudprovider.InstanceStatus
		switch {
		case isFailedMachineProviderID(providerIDNormalized):
			klog.V(4).Infof("Machine failed in node group %s (%s)", ng.Id(), providerID)

			machine, err := ng.machineController.findMachineByProviderID(providerIDNormalized)
			if err != nil {
				return nil, err
			}

			if machine != nil {
				if !machine.GetDeletionTimestamp().IsZero() {
					klog.V(4).Infof("Machine failed in node group %s (%s) is being deleted", ng.Id(), providerID)
					status = &cloudprovider.InstanceStatus{
						State: cloudprovider.InstanceDeleting,
						ErrorInfo: &cloudprovider.InstanceErrorInfo{
							ErrorClass:   cloudprovider.OtherErrorClass,
							ErrorCode:    "DeletingFailed",
							ErrorMessage: "Machine deletion failed",
						},
					}
				} else {
					_, nodeFound, err := unstructured.NestedFieldCopy(machine.UnstructuredContent(), "status", "nodeRef")
					if err != nil {
						return nil, err
					}

					// Machine failed without a nodeRef, this indicates that the machine failed to provision.
					// This is a special case where the machine is in a Failed state and the node is not created.
					// Machine controller will not reconcile this machine, so we need to report this to the autoscaler.
					if !nodeFound {
						klog.V(4).Infof("Machine failed in node group %s (%s) was being created", ng.Id(), providerID)
						status = &cloudprovider.InstanceStatus{
							State: cloudprovider.InstanceCreating,
							ErrorInfo: &cloudprovider.InstanceErrorInfo{
								ErrorClass:   cloudprovider.OtherErrorClass,
								ErrorCode:    "ProvisioningFailed",
								ErrorMessage: "Machine provisioning failed",
							},
						}
					}
				}
			}

		case isPendingMachineProviderID(providerIDNormalized):
			klog.V(4).Infof("Machine pending in node group %s (%s)", ng.Id(), providerID)
			status = &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			}

		case isDeletingMachineProviderID(providerIDNormalized):
			klog.V(4).Infof("Machine deleting in node group %s (%s)", ng.Id(), providerID)
			status = &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			}

		default:
			klog.V(4).Infof("Machine running in node group %s (%s)", ng.Id(), providerID)
			status = &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			}
		}

		instances[i] = cloudprovider.Instance{
			Id:     providerID,
			Status: status,
		}
	}

	return instances, nil
}

// TemplateNodeInfo returns a schedulercache.NodeInfo structure of an
// empty (as if just started) node. This will be used in scale-up
// simulations to predict what would a new node look like if a node
// group was expanded. The returned NodeInfo is expected to have a
// fully populated Node object, with all of the labels, capacity and
// allocatable information as well as all pods that are started on the
// node by default, using manifest (most likely only kube-proxy).
// Implementation optional.
func (ng *nodegroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	if !ng.scalableResource.CanScaleFromZero() {
		return nil, cloudprovider.ErrNotImplemented
	}

	capacity, err := ng.scalableResource.InstanceCapacity()
	if err != nil {
		return nil, err
	}

	nodeName := fmt.Sprintf("%s-asg-%d", ng.scalableResource.Name(), rand.Int63())
	node := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nodeName,
			Labels: map[string]string{},
		},
	}

	nsi := ng.scalableResource.InstanceSystemInfo()
	if nsi != nil {
		node.Status.NodeInfo = *nsi
	}

	node.Status.Capacity = capacity
	node.Status.Allocatable = capacity
	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	node.Spec.Taints = ng.scalableResource.Taints()

	node.Labels, err = ng.buildTemplateLabels(nodeName, nsi)
	if err != nil {
		return nil, err
	}

	resourceSlices, err := ng.scalableResource.InstanceResourceSlices(nodeName)
	if err != nil {
		return nil, err
	}
	csiNode := ng.scalableResource.InstanceCSINode()

	nodeInfo := framework.NewNodeInfo(&node, resourceSlices, framework.NewPodInfo(cloudprovider.BuildKubeProxy(ng.scalableResource.Name()), nil))
	if csiNode != nil {
		nodeInfo.SetCSINode(csiNode)
	}
	return nodeInfo, nil
}

func (ng *nodegroup) buildTemplateLabels(nodeName string, nsi *corev1.NodeSystemInfo) (map[string]string, error) {
	nsiLabels := make(map[string]string)
	if nsi != nil {
		nsiLabels[corev1.LabelArchStable] = nsi.Architecture
		nsiLabels[corev1.LabelOSStable] = nsi.OperatingSystem
	}

	// The order of priority is:
	// - Labels set in existing nodes for not-autoscale-from-zero cases
	// - Labels set in the labels capacity annotation of machine template, machine set, and machine deployment.
	// - Values in the status.nodeSystemInfo of MachineTemplates
	// - Generic/default labels set in the environment of the cluster autoscaler
	labels := cloudprovider.JoinStringMaps(buildGenericLabels(nodeName), nsiLabels, ng.scalableResource.Labels())

	nodes, err := ng.Nodes()
	if err != nil {
		return nil, err
	}

	if len(nodes) > 0 {
		node, err := ng.machineController.findNodeByProviderID(normalizedProviderString(nodes[0].Id))
		if err != nil {
			return nil, err
		}

		if node != nil {
			labels = cloudprovider.JoinStringMaps(labels, extractNodeLabels(node))
		}
	}
	return labels, nil
}

// Exist checks if the node group really exists on the cloud nodegroup
// side. Allows to tell the theoretical node group from the real one.
// Implementation required.
func (ng *nodegroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud nodegroup side.
// Implementation optional.
func (ng *nodegroup) Create() (cloudprovider.NodeGroup, error) {
	if ng.Exist() {
		return nil, cloudprovider.ErrAlreadyExist
	}
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud nodegroup side. This will
// be executed only for autoprovisioned node groups, once their size
// drops to 0. Implementation optional.
func (ng *nodegroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
// An autoprovisioned group was created by CA and can be deleted when
// scaled to 0.
func (ng *nodegroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (ng *nodegroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	options := ng.scalableResource.autoscalingOptions
	if options == nil || len(options) == 0 {
		return &defaults, nil
	}

	if opt, ok := getFloat64Option(options, ng.Id(), config.DefaultScaleDownUtilizationThresholdKey); ok {
		defaults.ScaleDownUtilizationThreshold = opt
	}
	if opt, ok := getFloat64Option(options, ng.Id(), config.DefaultScaleDownGpuUtilizationThresholdKey); ok {
		defaults.ScaleDownGpuUtilizationThreshold = opt
	}
	if opt, ok := getDurationOption(options, ng.Id(), config.DefaultScaleDownUnneededTimeKey); ok {
		defaults.ScaleDownUnneededTime = opt
	}
	if opt, ok := getDurationOption(options, ng.Id(), config.DefaultScaleDownUnreadyTimeKey); ok {
		defaults.ScaleDownUnreadyTime = opt
	}
	if opt, ok := getDurationOption(options, ng.Id(), config.DefaultMaxNodeProvisionTimeKey); ok {
		defaults.MaxNodeProvisionTime = opt
	}
	if opt, ok := getDurationOption(options, ng.Id(), config.DefaultMaxNodeStartupTimeKey); ok {
		defaults.MaxNodeStartupTime = opt
	}

	return &defaults, nil
}

func (ng *nodegroup) IsMachineDeploymentAndRollingOut() (bool, error) {
	if ng.scalableResource.Kind() != machineDeploymentKind {
		// Not a MachineDeployment.
		return false, nil
	}

	machineSets, err := ng.machineController.listMachineSetsForMachineDeployment(ng.scalableResource.unstructured)
	if err != nil {
		return false, err
	}

	if len(machineSets) == 0 {
		// No MachineSets => MD is not rolling out.
		return false, nil
	}

	// Find the latest revision, the MachineSet with the latest revision is the MachineSet that
	// matches the MachineDeployment spec.
	var latestMSRevisionInt int64
	for _, ms := range machineSets {
		msRevision, ok := ms.GetAnnotations()[machineDeploymentRevisionAnnotation]
		if !ok {
			continue
		}

		msRevisionInt, err := strconv.ParseInt(msRevision, 10, 64)
		if err != nil {
			return false, errors.Wrapf(err, "failed to parse current revision on MachineSet %s", klog.KObj(ms))
		}
		latestMSRevisionInt = max(latestMSRevisionInt, msRevisionInt)
	}
	maxMSRevision := strconv.FormatInt(latestMSRevisionInt, 10)

	for _, ms := range machineSets {
		if ms.GetAnnotations()[machineDeploymentRevisionAnnotation] == maxMSRevision {
			// Ignore the MachineSet with the latest revision
			continue
		}

		// Check if any of the old MachineSets still have replicas
		replicas, found, err := unstructured.NestedInt64(ms.UnstructuredContent(), "spec", "replicas")
		if err != nil {
			return false, errors.Wrapf(err, "failed to find spec replicas on MachineSet %s", klog.KObj(ms))
		}
		if found && replicas > 0 {
			// Found old MachineSets that still has replicas => MD is still rolling out.
			return true, nil
		}
		replicas, found, err = unstructured.NestedInt64(ms.UnstructuredContent(), "status", "replicas")
		if err != nil {
			return false, errors.Wrapf(err, "failed to find status replicas on MachineSet %s", klog.KObj(ms))
		}
		if found && replicas > 0 {
			// Found old MachineSets that still has replicas => MD is still rolling out.
			return true, nil
		}
	}

	// Didn't find any old MachineSets that still have replicas => MD is not rolling out.
	return false, nil
}

func newNodeGroupFromScalableResource(controller *machineController, unstructuredScalableResource *unstructured.Unstructured) (*nodegroup, error) {
	// Ensure that the resulting node group would be allowed based on the autodiscovery specs if defined
	if !controller.allowedByAutoDiscoverySpecs(unstructuredScalableResource) {
		return nil, nil
	}

	scalableResource, err := newUnstructuredScalableResource(controller, unstructuredScalableResource)
	if err != nil {
		return nil, err
	}

	replicas, found, err := unstructured.NestedInt64(unstructuredScalableResource.UnstructuredContent(), "spec", "replicas")
	if err != nil {
		return nil, err
	}

	// Ensure that if the nodegroup has 0 replicas it is capable
	// of scaling before adding it.
	if found && replicas == 0 && !scalableResource.CanScaleFromZero() {
		return nil, nil
	}

	// Ensure the node group would have the capacity to scale
	// allow MinSize = 0
	// allow MaxSize = MinSize
	// don't allow MaxSize < MinSize
	// don't allow MaxSize = MinSize = 0
	if scalableResource.MaxSize()-scalableResource.MinSize() < 0 || scalableResource.MaxSize() == 0 {
		klog.V(4).Infof("nodegroup %s has no scaling capacity, skipping", scalableResource.Name())
		return nil, nil
	}

	return &nodegroup{
		machineController: controller,
		scalableResource:  scalableResource,
	}, nil
}

func buildGenericLabels(nodeName string) map[string]string {
	// TODO revisit this function and add an explanation about what these
	// labels are used for, or remove them if not necessary
	m := make(map[string]string)
	m[corev1.LabelArchStable] = GetDefaultScaleFromZeroArchitecture().Name()

	m[corev1.LabelOSStable] = cloudprovider.DefaultOS

	m[corev1.LabelHostname] = nodeName
	return m
}

// extract a predefined list of labels from the existing node
func extractNodeLabels(node *corev1.Node) map[string]string {
	m := make(map[string]string)
	if node.Labels == nil {
		return m
	}

	setLabelIfNotEmpty(m, node.Labels, corev1.LabelArchStable)

	setLabelIfNotEmpty(m, node.Labels, corev1.LabelOSStable)

	setLabelIfNotEmpty(m, node.Labels, corev1.LabelInstanceType)
	setLabelIfNotEmpty(m, node.Labels, corev1.LabelInstanceTypeStable)

	setLabelIfNotEmpty(m, node.Labels, corev1.LabelZoneRegion)
	setLabelIfNotEmpty(m, node.Labels, corev1.LabelZoneRegionStable)

	setLabelIfNotEmpty(m, node.Labels, corev1.LabelZoneFailureDomain)

	return m
}

func setLabelIfNotEmpty(to, from map[string]string, key string) {
	if value := from[key]; value != "" {
		to[key] = value
	}
}

func getFloat64Option(options map[string]string, templateName, name string) (float64, bool) {
	raw, ok := options[name]
	if !ok {
		return 0, false
	}

	option, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		klog.Warningf("failed to convert autoscaling_options option %q (value %q) for scalable resource %q to float: %v", name, raw, templateName, err)
		return 0, false
	}

	return option, true
}

func getDurationOption(options map[string]string, templateName, name string) (time.Duration, bool) {
	raw, ok := options[name]
	if !ok {
		return 0, false
	}

	option, err := time.ParseDuration(raw)
	if err != nil {
		klog.Warningf("failed to convert autoscaling_options option %q (value %q) for scalable resource %q to duration: %v", name, raw, templateName, err)
		return 0, false
	}

	return option, true
}
