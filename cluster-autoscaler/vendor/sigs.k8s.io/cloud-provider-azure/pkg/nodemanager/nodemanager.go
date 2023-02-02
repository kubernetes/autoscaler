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

package nodemanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	clientretry "k8s.io/client-go/util/retry"
	cloudprovider "k8s.io/cloud-provider"
	cloudproviderapi "k8s.io/cloud-provider/api"
	cloudnodeutil "k8s.io/cloud-provider/node/helpers"
	nodeutil "k8s.io/component-helpers/node/util"
	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

// NodeProvider defines the interfaces for node provider.
type NodeProvider interface {
	// NodeAddresses returns the addresses of the specified instance.
	NodeAddresses(ctx context.Context, name types.NodeName) ([]v1.NodeAddress, error)
	// InstanceID returns the cloud provider ID of the specified instance.
	InstanceID(ctx context.Context, name types.NodeName) (string, error)
	// InstanceType returns the type of the specified instance.
	InstanceType(ctx context.Context, name types.NodeName) (string, error)
	// GetZone returns the Zone containing the current failure zone and locality region that the program is running in
	GetZone(ctx context.Context, name types.NodeName) (cloudprovider.Zone, error)
	// GetPlatformSubFaultDomain returns the PlatformSubFaultDomain from IMDS if set.
	GetPlatformSubFaultDomain() (string, error)
}

// labelReconcileInfo lists Node labels to reconcile, and how to reconcile them.
// primaryKey and secondaryKey are keys of labels to reconcile.
// - If both keys exist, but their values don't match. Use the value from the
// primaryKey as the source of truth to reconcile.
// - If ensureSecondaryExists is true, and the secondaryKey does not
// exist, secondaryKey will be added with the value of the primaryKey.
var labelReconcileInfo = []struct {
	primaryKey            string
	secondaryKey          string
	ensureSecondaryExists bool
}{}

// UpdateNodeSpecBackoff is the back configure for node update.
var UpdateNodeSpecBackoff = wait.Backoff{
	Steps:    20,
	Duration: 50 * time.Millisecond,
	Jitter:   1.0,
}

var updateNetworkConditionBackoff = wait.Backoff{
	Steps:    5, // Maximum number of retries.
	Duration: 100 * time.Millisecond,
	Jitter:   1.0,
}

// CloudNodeController reconciles node information.
type CloudNodeController struct {
	nodeName      string
	waitForRoutes bool
	nodeProvider  NodeProvider
	nodeInformer  coreinformers.NodeInformer
	kubeClient    clientset.Interface
	recorder      record.EventRecorder

	nodeStatusUpdateFrequency time.Duration
}

// NewCloudNodeController creates a CloudNodeController object
func NewCloudNodeController(
	nodeName string,
	nodeInformer coreinformers.NodeInformer,
	kubeClient clientset.Interface,
	nodeProvider NodeProvider,
	nodeStatusUpdateFrequency time.Duration,
	waitForRoutes bool) *CloudNodeController {

	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "cloud-node-controller"})
	eventBroadcaster.StartLogging(klog.Infof)
	if kubeClient != nil {
		klog.V(0).Infof("Sending events to api server.")
		eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	} else {
		klog.V(0).Infof("No api server defined - no events will be sent to API server.")
	}

	cnc := &CloudNodeController{
		nodeName:                  nodeName,
		nodeInformer:              nodeInformer,
		kubeClient:                kubeClient,
		recorder:                  recorder,
		nodeProvider:              nodeProvider,
		waitForRoutes:             waitForRoutes,
		nodeStatusUpdateFrequency: nodeStatusUpdateFrequency,
	}

	// Use shared informer to listen to add/update of nodes. Note that any nodes
	// that exist before node controller starts will show up in the update method
	_, _ = cnc.nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { cnc.AddCloudNode(context.TODO(), obj) },
		UpdateFunc: func(oldObj, newObj interface{}) { cnc.UpdateCloudNode(context.TODO(), oldObj, newObj) },
	})

	return cnc
}

// Run controller updates newly registered nodes with information
// from the cloud provider. This call is blocking so should be called
// via a goroutine
func (cnc *CloudNodeController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	// The following loops run communicate with the APIServer with a worst case complexity
	// of O(num_nodes) per cycle. These functions are justified here because these events fire
	// very infrequently. DO NOT MODIFY this to perform frequent operations.

	// Start a loop to periodically update the node addresses obtained from the cloud
	wait.Until(func() { cnc.UpdateNodeStatus(context.TODO()) }, cnc.nodeStatusUpdateFrequency, stopCh)
}

// UpdateNodeStatus updates the node status, such as node addresses
func (cnc *CloudNodeController) UpdateNodeStatus(ctx context.Context) {
	node, err := cnc.nodeInformer.Lister().Get(cnc.nodeName)
	if err != nil {
		// If node not found, just ignore it.
		if apierrors.IsNotFound(err) {
			return
		}

		klog.Errorf("Error getting node %q from informer, err: %v", cnc.nodeName, err)
		return
	}

	err = cnc.updateNodeAddress(ctx, node)
	if err != nil {
		klog.Errorf("Error reconciling node address for node %q, err: %v", node.Name, err)
	}

	err = cnc.reconcileNodeLabels(node)
	if err != nil {
		klog.Errorf("Error reconciling node labels for node %q, err: %v", node.Name, err)
	}
}

// reconcileNodeLabels reconciles node labels transitioning from beta to GA
func (cnc *CloudNodeController) reconcileNodeLabels(node *v1.Node) error {
	if node.Labels == nil {
		// Nothing to reconcile.
		return nil
	}

	labelsToUpdate := map[string]string{}
	for _, r := range labelReconcileInfo {
		primaryValue, primaryExists := node.Labels[r.primaryKey]
		secondaryValue, secondaryExists := node.Labels[r.secondaryKey]

		if !primaryExists {
			// The primary label key does not exist. This should not happen
			// within our supported version skew range, when no external
			// components/factors modifying the node object. Ignore this case.
			continue
		}
		if secondaryExists && primaryValue != secondaryValue {
			// Secondary label exists, but not consistent with the primary
			// label. Need to reconcile.
			labelsToUpdate[r.secondaryKey] = primaryValue

		} else if !secondaryExists && r.ensureSecondaryExists {
			// Apply secondary label based on primary label.
			labelsToUpdate[r.secondaryKey] = primaryValue
		}
	}

	if len(labelsToUpdate) == 0 {
		return nil
	}

	if !cloudnodeutil.AddOrUpdateLabelsOnNode(cnc.kubeClient, labelsToUpdate, node) {
		return fmt.Errorf("failed update labels for node %+v", node)
	}

	return nil
}

// UpdateNodeAddress updates the nodeAddress of a single node
func (cnc *CloudNodeController) updateNodeAddress(ctx context.Context, node *v1.Node) error {
	// Do not process nodes that are still tainted
	cloudTaint := GetCloudTaint(node.Spec.Taints)
	if cloudTaint != nil {
		klog.V(5).Infof("This node %s is still tainted. Will not process.", node.Name)
		return nil
	}

	// Node that isn't present according to the cloud provider shouldn't have its address updated
	exists, err := cnc.ensureNodeExistsByProviderID(ctx, node)
	if err != nil {
		// Continue to update node address when not sure the node is not exists
		klog.Warningf("ensureNodeExistsByProviderID (node %s) reported an error (%v), continue to update its address", node.Name, err)
	} else if !exists {
		klog.V(4).Infof("The node %s is no longer present according to the cloud provider, do not process.", node.Name)
		return nil
	}

	nodeAddresses, err := cnc.getNodeAddressesByName(ctx, node)
	if err != nil {
		return fmt.Errorf("getting node addresses for node %q: %w", node.Name, err)
	}

	if len(nodeAddresses) == 0 {
		klog.V(5).Infof("Skipping node address update for node %q since cloud provider did not return any", node.Name)
		return nil
	}

	// Check if a hostname address exists in the cloud provided addresses
	hostnameExists := false
	for i := range nodeAddresses {
		if nodeAddresses[i].Type == v1.NodeHostName {
			hostnameExists = true
			break
		}
	}
	// If hostname was not present in cloud provided addresses, use the hostname
	// from the existing node (populated by kubelet)
	if !hostnameExists {
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeHostName {
				nodeAddresses = append(nodeAddresses, addr)
			}
		}
	}
	// If nodeIP was suggested by user, ensure that
	// it can be found in the cloud as well (consistent with the behaviour in kubelet)
	if nodeIP, ok := ensureNodeProvidedIPExists(node, nodeAddresses); ok {
		if nodeIP == nil {
			return fmt.Errorf("specified Node IP %s not found in cloudprovider for node %q", nodeAddresses, node.Name)
		}
	}
	if !nodeAddressesChangeDetected(node.Status.Addresses, nodeAddresses) {
		return nil
	}

	newNode := node.DeepCopy()
	newNode.Status.Addresses = nodeAddresses
	_, _, err = PatchNodeStatus(cnc.kubeClient.CoreV1(), types.NodeName(node.Name), node, newNode)
	if err != nil {
		return fmt.Errorf("patching node with cloud ip addresses: %w", err)
	}

	return nil
}

// nodeModifier is used to carry changes to node objects across multiple attempts to update them
// in a retry-if-conflict loop.
type nodeModifier func(*v1.Node)

// UpdateCloudNode handles node update event.
func (cnc *CloudNodeController) UpdateCloudNode(ctx context.Context, _, newObj interface{}) {
	node, ok := newObj.(*v1.Node)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	// Skip other nodes other than cnc.nodeName.
	if !strings.EqualFold(cnc.nodeName, node.Name) {
		return
	}

	cloudTaint := GetCloudTaint(node.Spec.Taints)
	if cloudTaint == nil {
		// The node has already been initialized so nothing to do.
		return
	}

	cnc.initializeNode(ctx, node)
}

// AddCloudNode handles initializing new nodes registered with the cloud taint.
func (cnc *CloudNodeController) AddCloudNode(ctx context.Context, obj interface{}) {
	node := obj.(*v1.Node)

	// Skip other nodes other than cnc.nodeName.
	if !strings.EqualFold(cnc.nodeName, node.Name) {
		return
	}

	cloudTaint := GetCloudTaint(node.Spec.Taints)
	if cloudTaint == nil {
		klog.V(2).Infof("This node %s is registered without the cloud taint. Will not process.", node.Name)
		return
	}

	cnc.initializeNode(ctx, node)
}

// This processes nodes that were added into the cluster, and cloud initialize them if appropriate
func (cnc *CloudNodeController) initializeNode(ctx context.Context, node *v1.Node) {
	klog.Infof("Initializing node %s with cloud provider", node.Name)
	curNode, err := cnc.kubeClient.CoreV1().Nodes().Get(ctx, node.Name, metav1.GetOptions{})
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to get node %s: %w", node.Name, err))
		return
	}

	cloudTaint := GetCloudTaint(curNode.Spec.Taints)
	if cloudTaint == nil {
		// Node object received from event had the cloud taint but was outdated,
		// the node has actually already been initialized.
		return
	}

	if cnc.waitForRoutes {
		// Set node condition node NodeNetworkUnavailable=true so that Pods won't
		// be scheduled to this node until routes have been created.
		err = cnc.updateNetworkingCondition(node, false)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("failed to patch condition for node %s: %w", node.Name, err))
			return
		}
	}

	var nodeModifiers []nodeModifier
	err = clientretry.OnError(UpdateNodeSpecBackoff, func(err error) bool {
		return err != nil && strings.HasPrefix(err.Error(), "failed to set node provider id")
	}, func() error {
		nodeModifiers, err = cnc.getNodeModifiersFromCloudProvider(ctx, curNode)
		return err
	})
	if err != nil {
		// Instead of just logging the error, panic and node manager can restart
		utilruntime.Must(fmt.Errorf("failed to initialize node %s at cloudprovider: %w", node.Name, err))
		return
	}

	nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
		n.Spec.Taints = excludeCloudTaint(n.Spec.Taints)
	})

	err = clientretry.RetryOnConflict(UpdateNodeSpecBackoff, func() error {
		curNode, err := cnc.kubeClient.CoreV1().Nodes().Get(ctx, node.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for _, modify := range nodeModifiers {
			modify(curNode)
		}

		_, err = cnc.kubeClient.CoreV1().Nodes().Update(ctx, curNode, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

		// After adding, call UpdateNodeAddress to set the CloudProvider provided IPAddresses
		// So that users do not see any significant delay in IP addresses being filled into the node
		err = cnc.updateNodeAddress(ctx, curNode)
		if err != nil {
			return err
		}

		klog.Infof("Successfully initialized node %s with cloud provider", node.Name)
		return nil
	})
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
}

// getNodeModifiersFromCloudProvider returns a slice of nodeModifiers that update
// a node object with provider-specific information.
// All of the returned functions are idempotent, because they are used in a retry-if-conflict
// loop, meaning they could get called multiple times.
func (cnc *CloudNodeController) getNodeModifiersFromCloudProvider(ctx context.Context, node *v1.Node) ([]nodeModifier, error) {
	var nodeModifiers []nodeModifier

	if node.Spec.ProviderID == "" {
		providerID, err := cnc.nodeProvider.InstanceID(ctx, types.NodeName(node.Name))
		if err == nil {
			nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
				if n.Spec.ProviderID == "" {
					n.Spec.ProviderID = providerID
				}
			})
		} else {
			// if we are not able to get node provider id,
			// we return error here and retry in the caller initializeNode()
			return nil, fmt.Errorf("failed to set node provider id: %w", err)
		}
	}

	nodeAddresses, err := cnc.getNodeAddressesByName(ctx, node)
	if err != nil {
		return nil, err
	}

	// If user provided an IP address, ensure that IP address is found
	// in the cloud provider before removing the taint on the node
	if nodeIP, ok := ensureNodeProvidedIPExists(node, nodeAddresses); ok {
		if nodeIP == nil {
			return nil, errors.New("failed to find kubelet node IP from cloud provider")
		}
	}

	if instanceType, err := cnc.getInstanceTypeByName(ctx, node); err != nil {
		return nil, err
	} else if instanceType != "" {
		klog.V(2).Infof("Adding node label from cloud provider: %s=%s", v1.LabelInstanceTypeStable, instanceType)
		nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
			if n.Labels == nil {
				n.Labels = map[string]string{}
			}
			n.Labels[v1.LabelInstanceTypeStable] = instanceType
		})
	}
	zone, err := cnc.getZoneByName(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone from cloud provider: %w", err)
	}
	if zone.FailureDomain != "" {
		klog.V(2).Infof("Adding node label from cloud provider: %s=%s", v1.LabelZoneFailureDomainStable, zone.FailureDomain)
		nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
			if n.Labels == nil {
				n.Labels = map[string]string{}
			}
			n.Labels[v1.LabelZoneFailureDomainStable] = zone.FailureDomain
		})
	}
	if zone.Region != "" {
		klog.V(2).Infof("Adding node label from cloud provider: %s=%s", v1.LabelZoneRegionStable, zone.Region)
		nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
			if n.Labels == nil {
				n.Labels = map[string]string{}
			}
			n.Labels[v1.LabelZoneRegionStable] = zone.Region
		})
	}
	platformSubFaultDomain, err := cnc.getPlatformSubFaultDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to get platformSubFaultDomain: %w", err)
	}
	if platformSubFaultDomain != "" {
		klog.V(2).Infof("Adding node label from cloud provider: %s=%s", consts.LabelPlatformSubFaultDomain, platformSubFaultDomain)
		nodeModifiers = append(nodeModifiers, func(n *v1.Node) {
			if n.Labels == nil {
				n.Labels = map[string]string{}
			}
			n.Labels[consts.LabelPlatformSubFaultDomain] = platformSubFaultDomain
		})
	}

	return nodeModifiers, nil
}

func GetCloudTaint(taints []v1.Taint) *v1.Taint {
	for _, taint := range taints {
		if taint.Key == cloudproviderapi.TaintExternalCloudProvider {
			return &taint
		}
	}
	return nil
}

func excludeCloudTaint(taints []v1.Taint) []v1.Taint {
	newTaints := []v1.Taint{}
	for _, taint := range taints {
		if taint.Key == cloudproviderapi.TaintExternalCloudProvider {
			continue
		}
		newTaints = append(newTaints, taint)
	}
	return newTaints
}

// ensureNodeExistsByProviderID checks if the instance exists by the provider id,
// If provider id in spec is empty it calls instanceId with node name to get provider id
func (cnc *CloudNodeController) ensureNodeExistsByProviderID(ctx context.Context, node *v1.Node) (bool, error) {
	providerID := node.Spec.ProviderID
	if providerID == "" {
		var err error
		providerID, err = cnc.nodeProvider.InstanceID(ctx, types.NodeName(node.Name))
		if err != nil {
			if errors.Is(err, cloudprovider.InstanceNotFound) {
				return false, nil
			}
			return false, err
		}

		if providerID == "" {
			klog.Warningf("Cannot find valid providerID for node name %q, assuming non existence", node.Name)
			return false, nil
		}
	}

	return true, nil
}

func (cnc *CloudNodeController) getNodeAddressesByName(ctx context.Context, node *v1.Node) ([]v1.NodeAddress, error) {
	nodeAddresses, err := cnc.nodeProvider.NodeAddresses(ctx, types.NodeName(node.Name))
	if err != nil {
		return nil, fmt.Errorf("error fetching node by name %s: %w", node.Name, err)
	}
	return nodeAddresses, nil
}

func nodeAddressesChangeDetected(addressSet1, addressSet2 []v1.NodeAddress) bool {
	if len(addressSet1) != len(addressSet2) {
		return true
	}
	addressMap1 := map[v1.NodeAddressType]string{}

	for i := range addressSet1 {
		addressMap1[addressSet1[i].Type] = addressSet1[i].Address
	}

	for _, v := range addressSet2 {
		if addressMap1[v.Type] != v.Address {
			return true
		}
	}
	return false
}

func ensureNodeProvidedIPExists(node *v1.Node, nodeAddresses []v1.NodeAddress) (*v1.NodeAddress, bool) {
	var nodeIP *v1.NodeAddress
	nodeIPExists := false
	if providedIP, ok := node.ObjectMeta.Annotations[cloudproviderapi.AnnotationAlphaProvidedIPAddr]; ok {
		nodeIPExists = true
		for i := range nodeAddresses {
			if nodeAddresses[i].Address == providedIP {
				nodeIP = &nodeAddresses[i]
				break
			}
		}
	}
	return nodeIP, nodeIPExists
}

func (cnc *CloudNodeController) getInstanceTypeByName(ctx context.Context, node *v1.Node) (string, error) {
	instanceType, err := cnc.nodeProvider.InstanceType(ctx, types.NodeName(node.Name))
	if err != nil {
		return "", fmt.Errorf("InstanceType: Error fetching by NodeName %s: %w", node.Name, err)
	}
	return instanceType, err
}

// getZoneByName will attempt to get the zone of node using its providerID
// then it's name. If both attempts fail, an error is returned
func (cnc *CloudNodeController) getZoneByName(ctx context.Context, node *v1.Node) (cloudprovider.Zone, error) {
	zone, err := cnc.nodeProvider.GetZone(ctx, types.NodeName(node.Name))
	if err != nil {
		return cloudprovider.Zone{}, fmt.Errorf("Zone: Error fetching by NodeName %s: %w", node.Name, err)
	}

	return zone, nil
}

func (cnc *CloudNodeController) getPlatformSubFaultDomain() (string, error) {
	subFD, err := cnc.nodeProvider.GetPlatformSubFaultDomain()
	if err != nil {
		return "", fmt.Errorf("cnc.getPlatformSubfaultDomain: %w", err)
	}
	return subFD, nil
}

func (cnc *CloudNodeController) updateNetworkingCondition(node *v1.Node, networkReady bool) error {
	_, condition := nodeutil.GetNodeCondition(&(node.Status), v1.NodeNetworkUnavailable)
	if networkReady && condition != nil && condition.Status == v1.ConditionFalse {
		klog.V(4).Infof("set node %v with NodeNetworkUnavailable=false was canceled because it is already set", node.Name)
		return nil
	}

	if !networkReady && condition != nil && condition.Status == v1.ConditionTrue {
		klog.V(4).Infof("set node %v with NodeNetworkUnavailable=true was canceled because it is already set", node.Name)
		return nil
	}

	klog.V(2).Infof("Patching node status %v with %v previous condition was:%+v", node.Name, networkReady, condition)

	// either condition is not there, or has a value != to what we need
	// start setting it
	err := clientretry.RetryOnConflict(updateNetworkConditionBackoff, func() error {
		var err error
		// Patch could also fail, even though the chance is very slim. So we still do
		// patch in the retry loop.
		currentTime := metav1.Now()
		if networkReady {
			err = nodeutil.SetNodeCondition(cnc.kubeClient, types.NodeName(node.Name), v1.NodeCondition{
				Type:               v1.NodeNetworkUnavailable,
				Status:             v1.ConditionFalse,
				Reason:             "NodeInitialization",
				Message:            "Don't need to wait for cloud routes",
				LastTransitionTime: currentTime,
			})
		} else {
			err = nodeutil.SetNodeCondition(cnc.kubeClient, types.NodeName(node.Name), v1.NodeCondition{
				Type:               v1.NodeNetworkUnavailable,
				Status:             v1.ConditionTrue,
				Reason:             "NodeInitialization",
				Message:            "Waiting for cloud routes",
				LastTransitionTime: currentTime,
			})
		}
		if err != nil {
			klog.V(4).Infof("Error updating node %s, retrying: %v", types.NodeName(node.Name), err)
		}
		return err
	})

	if err != nil {
		klog.Errorf("Error updating node %s: %v", node.Name, err)
	}

	return err
}

// PatchNodeStatus patches node status.
func PatchNodeStatus(c v1core.CoreV1Interface, nodeName types.NodeName, oldNode *v1.Node, newNode *v1.Node) (*v1.Node, []byte, error) {
	patchBytes, err := preparePatchBytesforNodeStatus(nodeName, oldNode, newNode)
	if err != nil {
		return nil, nil, err
	}

	updatedNode, err := c.Nodes().Patch(context.TODO(), string(nodeName), types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}, "status")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to patch status %q for node %q: %w", patchBytes, nodeName, err)
	}
	return updatedNode, patchBytes, nil
}

func preparePatchBytesforNodeStatus(nodeName types.NodeName, oldNode *v1.Node, newNode *v1.Node) ([]byte, error) {
	oldData, err := json.Marshal(oldNode)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal oldData for node %q: %w", nodeName, err)
	}

	// NodeStatus.Addresses is incorrectly annotated as patchStrategy=merge, which
	// will cause strategicpatch.CreateTwoWayMergePatch to create an incorrect patch
	// if it changed.
	manuallyPatchAddresses := (len(oldNode.Status.Addresses) > 0) && !equality.Semantic.DeepEqual(oldNode.Status.Addresses, newNode.Status.Addresses)

	// Reset spec to make sure only patch for Status or ObjectMeta is generated.
	// Note that we don't reset ObjectMeta here, because:
	// 1. This aligns with Nodes().UpdateStatus().
	// 2. Some component does use this to update node annotations.
	diffNode := newNode.DeepCopy()
	diffNode.Spec = oldNode.Spec
	if manuallyPatchAddresses {
		diffNode.Status.Addresses = oldNode.Status.Addresses
	}
	newData, err := json.Marshal(diffNode)
	if err != nil {
		return nil, fmt.Errorf("failed to Marshal newData for node %q: %w", nodeName, err)
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, v1.Node{})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateTwoWayMergePatch for node %q: %w", nodeName, err)
	}
	if manuallyPatchAddresses {
		patchBytes, err = fixupPatchForNodeStatusAddresses(patchBytes, newNode.Status.Addresses)
		if err != nil {
			return nil, fmt.Errorf("failed to fix up NodeAddresses in patch for node %q: %w", nodeName, err)
		}
	}

	return patchBytes, nil
}

// fixupPatchForNodeStatusAddresses adds a replace-strategy patch for Status.Addresses to
// the existing patch
func fixupPatchForNodeStatusAddresses(patchBytes []byte, addresses []v1.NodeAddress) ([]byte, error) {
	// Given patchBytes='{"status": {"conditions": [ ... ], "phase": ...}}' and
	// addresses=[{"type": "InternalIP", "address": "10.0.0.1"}], we need to generate:
	//
	//   {
	//     "status": {
	//       "conditions": [ ... ],
	//       "phase": ...,
	//       "addresses": [
	//         {
	//           "type": "InternalIP",
	//           "address": "10.0.0.1"
	//         },
	//         {
	//           "$patch": "replace"
	//         }
	//       ]
	//     }
	//   }

	var patchMap map[string]interface{}
	if err := json.Unmarshal(patchBytes, &patchMap); err != nil {
		return nil, err
	}

	addrBytes, err := json.Marshal(addresses)
	if err != nil {
		return nil, err
	}
	var addrArray []interface{}
	if err := json.Unmarshal(addrBytes, &addrArray); err != nil {
		return nil, err
	}
	addrArray = append(addrArray, map[string]interface{}{"$patch": "replace"})

	status := patchMap["status"]
	if status == nil {
		status = map[string]interface{}{}
		patchMap["status"] = status
	}
	statusMap, ok := status.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data in patch")
	}
	statusMap["addresses"] = addrArray

	return json.Marshal(patchMap)
}
