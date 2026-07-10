/*
Copyright The Kubernetes Authors.

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

package state

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"sync"
	"sync/atomic"
	"time"

	"github.com/awslabs/operatorpkg/serrors"
	"github.com/samber/lo"
	"go.uber.org/multierr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/scheduling"
	nodeclaimutils "sigs.k8s.io/karpenter/pkg/utils/nodeclaim"
	podutils "sigs.k8s.io/karpenter/pkg/utils/pod"
	"sigs.k8s.io/karpenter/pkg/utils/resources"
)

// Cluster maintains cluster state that is often needed but expensive to compute.
type Cluster struct {
	kubeClient    client.Client
	cloudProvider cloudprovider.CloudProvider
	clock         clock.Clock
	hasSynced     atomic.Bool

	mu                        sync.RWMutex
	nodes                     map[string]*StateNode           // provider id -> cached node
	bindings                  map[types.NamespacedName]string // pod namespaced named -> node name
	nodeNameToProviderID      map[string]string               // node name -> provider id
	nodeClaimNameToProviderID map[string]string               // node claim name -> provider id
	nodePoolResources         map[string]corev1.ResourceList  // node pool name -> resource list
	daemonSetPods             sync.Map                        // daemonSet -> existing pod

	NodePoolState *NodePoolState

	podAcks                         sync.Map // pod namespaced name -> time when Karpenter first saw the pod as pending
	podsSchedulingAttempted         sync.Map // pod namespaced name -> time when Karpenter tried to schedule a pod
	podsSchedulableTimes            sync.Map // pod namespaced name -> time when it was first marked as able to fit to a node
	podHealthyNodePoolScheduledTime sync.Map // pod namespaced name -> time when pod scheduled to a nodePool that has NodeRegistrationHealthy=true, is marked as able to fit to a node
	podToNodeClaim                  sync.Map // pod namespaced name -> nodeClaim name

	clusterStateMu sync.RWMutex // Separate mutex as this is called in some places that mu is held
	// A monotonically increasing timestamp representing the time state of the
	// cluster with respect to consolidation. This increases when something has
	// changed about the cluster that might make consolidation possible. By recording
	// the state, interested disruption methods can check to see if this has changed to
	// optimize and not try to disrupt if nothing about the cluster has changed.
	clusterState time.Time

	unsyncedTimeMu      sync.Mutex
	unsyncedStartTime   time.Time
	lastUnsyncedLogTime time.Time
	antiAffinityPods    sync.Map // pod namespaced name -> *corev1.Pod of pods that have required anti affinities

	// bufferPodCounts tracks how many virtual CapacityBuffer pods the provisioner's
	// scheduling simulation placed on each node (keyed by providerID). This is
	// rebuilt wholesale after every provisioning pass — it is NOT incremental.
	//
	// Used by:
	//   - disruption/emptiness.go: prevents empty-consolidation of nodes that host
	//     buffer capacity (HasBufferPods check in ShouldDisrupt).
	//
	// NOT used by consolidation — consolidation naturally accounts for buffer pods
	// because SimulateScheduling calls GetPendingPods which injects virtual pods
	// into the pending set. Any replacement must fit both real and virtual pods.
	bufferPodCountsMu sync.RWMutex
	bufferPodCounts   map[string]int
}

func NewCluster(clk clock.Clock, client client.Client, cloudProvider cloudprovider.CloudProvider) *Cluster {
	return &Cluster{
		clock:                     clk,
		kubeClient:                client,
		cloudProvider:             cloudProvider,
		nodes:                     map[string]*StateNode{},
		bindings:                  map[types.NamespacedName]string{},
		daemonSetPods:             sync.Map{},
		nodeNameToProviderID:      map[string]string{},
		nodeClaimNameToProviderID: map[string]string{},
		nodePoolResources:         map[string]corev1.ResourceList{},

		NodePoolState: NewNodePoolState(),

		bufferPodCounts: map[string]int{},

		podAcks:                         sync.Map{},
		podsSchedulableTimes:            sync.Map{},
		podsSchedulingAttempted:         sync.Map{},
		podHealthyNodePoolScheduledTime: sync.Map{},
		podToNodeClaim:                  sync.Map{},
	}
}

// Synced validates that the NodeClaims and the Nodes that are stored in the apiserver
// have the same representation in the cluster state. This is to ensure that our view
// of the cluster is as close to correct as it can be when we begin to perform operations
// utilizing the cluster state as our source of truth
//
//nolint:gocyclo
func (c *Cluster) Synced(ctx context.Context) (synced bool) {
	// Set the metric depending on the result of the Synced() call
	defer func() {
		c.unsyncedTimeMu.Lock()
		defer c.unsyncedTimeMu.Unlock()

		if synced {
			c.unsyncedStartTime = time.Time{}
			c.lastUnsyncedLogTime = time.Time{}
			ClusterStateUnsyncedTimeSeconds.Set(0, nil)
		} else {
			if c.unsyncedStartTime.IsZero() {
				c.unsyncedStartTime = c.clock.Now()
			}
			// We want to log every 10s when the cluster hasn't synced for 30s which is long enough for us to think there is an issue
			if c.clock.Since(c.unsyncedStartTime) > time.Second*30 && c.clock.Since(c.lastUnsyncedLogTime) > time.Second*10 {
				c.lastUnsyncedLogTime = c.clock.Now()
				log.FromContext(ctx).Error(
					fmt.Errorf("waiting on cluster sync"),
					"cluster is waiting on sync for extended duration",
					"duration", c.clock.Since(c.unsyncedStartTime).Truncate(time.Second))
			}
			ClusterStateUnsyncedTimeSeconds.Set(c.clock.Since(c.unsyncedStartTime).Seconds(), nil)
		}
	}()
	// Set the metric to whatever the result of the Synced() call is
	defer func() {
		ClusterStateSynced.Set(lo.Ternary[float64](synced, 1, 0), nil)
	}()

	// If the cluster state has already synced once, then we assume that objects are kept internally consistent
	// with each other to avoid having to continually re-check that we have fully captured the same view
	// of cluster state that controller-runtime has
	if c.hasSynced.Load() {
		c.mu.RLock()
		defer c.mu.RUnlock()

		for _, providerID := range c.nodeClaimNameToProviderID {
			// Check to see if any node claim doesn't have a provider ID. If it doesn't, then the nodeclaim hasn't been
			// launched, and we need to wait to see what the resolved values are before continuing.
			if providerID == "" {
				return false
			}
		}
		return true
	}

	// If we haven't synced before, then we need to make sure that our internal cache is fully hydrated
	// before we start doing operations against the state
	// Because we get so many NodeClaims from this response, we are not DeepCopying the cached data here
	// DO NOT MUTATE NodeClaims in this function as this will affect the underlying cached NodeClaim
	nodeClaims, err := nodeclaimutils.ListManaged(ctx, c.kubeClient, c.cloudProvider, client.UnsafeDisableDeepCopy)
	if err != nil {
		log.FromContext(ctx).Error(err, "failed checking cluster state sync")
		return false
	}
	nodeList := &corev1.NodeList{}
	// Because we get so many Nodes from this response, we are not DeepCopying the cached data here
	// DO NOT MUTATE Nodes in this function as this will affect the underlying cached Node
	if err := c.kubeClient.List(ctx, nodeList, client.UnsafeDisableDeepCopy); err != nil {
		log.FromContext(ctx).Error(err, "failed checking cluster state sync")
		return false
	}

	c.mu.RLock()
	stateNodeClaimNames := sets.New[string]()
	for name, providerID := range c.nodeClaimNameToProviderID {
		// Check to see if any node claim doesn't have a provider ID. If it doesn't, then the nodeclaim hasn't been
		// launched, and we need to wait to see what the resolved values are before continuing.
		if providerID == "" {
			c.mu.RUnlock()
			return false
		}
		stateNodeClaimNames.Insert(name)
	}
	stateNodeNames := sets.New(lo.Keys(c.nodeNameToProviderID)...)
	c.mu.RUnlock()

	nodeClaimNames := sets.New[string]()
	for _, nodeClaim := range nodeClaims {
		nodeClaimNames.Insert(nodeClaim.Name)
	}
	nodeNames := sets.New[string]()
	for _, node := range nodeList.Items {
		nodeNames.Insert(node.Name)
	}
	// The names tracked in-memory should at least have all the data that is in the api-server
	// This doesn't ensure that the two states are exactly aligned (we could still not be tracking a node
	// that exists in the cluster state but not in the apiserver) but it ensures that we have a state
	// representation for every node/nodeClaim that exists on the apiserver
	synced = stateNodeClaimNames.IsSuperset(nodeClaimNames) && stateNodeNames.IsSuperset(nodeNames)
	if synced {
		c.hasSynced.Store(true)
	}
	return synced
}

// ForPodsWithAntiAffinity calls the supplied function once for each pod with required anti affinity terms that is
// currently bound to a node. The pod returned may not be up-to-date with respect to status, however since the
// anti-affinity terms can't be modified, they will be correct.
func (c *Cluster) ForPodsWithAntiAffinity(fn func(p *corev1.Pod, n *corev1.Node) bool) {
	c.antiAffinityPods.Range(func(key, value any) bool {
		pod := value.(*corev1.Pod)
		c.mu.RLock()
		defer c.mu.RUnlock()
		nodeName, ok := c.bindings[client.ObjectKeyFromObject(pod)]
		if !ok {
			return true
		}
		node, ok := c.nodes[c.nodeNameToProviderID[nodeName]]
		if !ok || node.Node == nil {
			// if we receive the node deletion event before the pod deletion event, this can happen
			return true
		}
		return fn(pod, node.Node)
	})
}

// Nodes returns an iterator which iterates over the state nodes in the cluster under a read-lock. It is not safe to
// store the state.StateNode object and it should only be accessed while the iterator is active.
func (c *Cluster) Nodes() iter.Seq[*StateNode] {
	return func(yield func(*StateNode) bool) {
		c.mu.RLock()
		defer c.mu.RUnlock()
		for _, node := range c.nodes {
			if !yield(node) {
				return
			}
		}
	}
}

// DeepCopyNodes creates a DeepCopy of all state nodes.
// NOTE: This is very inefficient so this should only be used when DeepCopying is absolutely necessary
func (c *Cluster) DeepCopyNodes() StateNodes {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return lo.Map(lo.Values(c.nodes), func(n *StateNode, _ int) *StateNode {
		return n.DeepCopy()
	})
}

// IsNodeNominated returns true if the given node was expected to have a pod bound to it during a recent scheduling
// batch
func (c *Cluster) IsNodeNominated(providerID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if n, ok := c.nodes[providerID]; ok {
		return n.Nominated(c.clock)
	}
	return false
}

// NominateNodeForPod records that a node was the target of a pending pod during a scheduling batch
func (c *Cluster) NominateNodeForPod(ctx context.Context, providerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.nodes[providerID]; ok {
		n.Nominate(ctx, c.clock) // extends nomination window if already nominated
	}
}

// UpdateBufferPodCounts replaces the entire buffer pod count mapping with the
// new counts. Called by the provisioner after each scheduling pass to reflect
// which nodes are hosting virtual buffer pods. Nodes not in the map are cleared.
//
// The mapping is derived from Results.ExistingNodes — each ExistingNode.Pods
// entry that carries the fake-pod annotation is counted. When buffer capacity
// is consumed (real pods take the space), virtual pods move to other nodes or
// new NodeClaims, and this map updates accordingly on the next pass.
func (c *Cluster) UpdateBufferPodCounts(counts map[string]int) {
	c.bufferPodCountsMu.Lock()
	defer c.bufferPodCountsMu.Unlock()
	c.bufferPodCounts = counts
}

// HasBufferPods returns true if the node with the given providerID has at least
// one virtual buffer pod placed on it during the last provisioning pass.
func (c *Cluster) HasBufferPods(providerID string) bool {
	return c.BufferPodCount(providerID) > 0
}

// BufferPodCount returns the number of virtual buffer pods on the node.
func (c *Cluster) BufferPodCount(providerID string) int {
	c.bufferPodCountsMu.RLock()
	defer c.bufferPodCountsMu.RUnlock()
	return c.bufferPodCounts[providerID]
}

// UnmarkForDeletion removes the marking on the node as a node the controller intends to delete
func (c *Cluster) UnmarkForDeletion(providerIDs ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, id := range providerIDs {
		if n, ok := c.nodes[id]; ok {
			oldNode := n.ShallowCopy()
			n.markedForDeletion = false
			c.updateNodePoolResources(oldNode, n)
			if n.NodeClaim != nil && n.NodeClaim.DeletionTimestamp.IsZero() {
				c.NodePoolState.MarkNodeClaimActive(n.NodeClaim.Labels[v1.NodePoolLabelKey], n.NodeClaim.Name)
			}
		}
	}
}

// MarkForDeletion marks the node as pending deletion in the internal cluster state
func (c *Cluster) MarkForDeletion(providerIDs ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, id := range providerIDs {
		if n, ok := c.nodes[id]; ok {
			oldNode := n.ShallowCopy()
			n.markedForDeletion = true
			c.updateNodePoolResources(oldNode, n)
			if n.NodeClaim != nil {
				c.NodePoolState.MarkNodeClaimDeleting(n.NodeClaim.Labels[v1.NodePoolLabelKey], n.NodeClaim.Name)
			}
		}
	}
}

func (c *Cluster) UpdateNodeClaim(nodeClaim *v1.NodeClaim) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If the nodeclaim has a providerID, create a StateNode for it, and populate the data.
	// We only need to do this for a nodeclaim with a providerID as nodeclaims without provider IDs haven't
	// been launched yet.
	if nodeClaim.Status.ProviderID != "" {
		n := c.newStateFromNodeClaim(nodeClaim, c.nodes[nodeClaim.Status.ProviderID])
		c.nodes[nodeClaim.Status.ProviderID] = n
	}

	// Update nodepool state with NodeClaim
	markedForDel := false
	if n, ok := c.nodes[nodeClaim.Status.ProviderID]; ok {
		markedForDel = n.MarkedForDeletion()
	}
	c.NodePoolState.UpdateNodeClaim(nodeClaim, markedForDel)

	// If the nodeclaim hasn't launched yet, we want to add it into cluster state to ensure
	// that we're not racing with the internal cache for the cluster, assuming the node doesn't exist.
	c.nodeClaimNameToProviderID[nodeClaim.Name] = nodeClaim.Status.ProviderID
	ClusterStateNodesCount.Set(float64(len(c.nodes)), nil)
}

func (c *Cluster) DeleteNodeClaim(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cleanupNodeClaim(name)
	ClusterStateNodesCount.Set(float64(len(c.nodes)), nil)
}

func (c *Cluster) UpdateNode(ctx context.Context, node *corev1.Node) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	managed := node.Labels[v1.NodePoolLabelKey] != ""
	initialized := node.Labels[v1.NodeInitializedLabelKey] != ""
	if node.Spec.ProviderID == "" {
		// If we know that we own this node, we shouldn't allow the providerID to be empty
		if managed {
			return nil
		}
		node.Spec.ProviderID = node.Name
	}
	// If we have a managed node with no instance type label that hasn't been initialized,
	// we need to wait until the instance type label gets propagated on it
	if managed && node.Labels[corev1.LabelInstanceTypeStable] == "" && !initialized {
		return nil
	}
	n, err := c.newStateFromNode(ctx, node, c.nodes[node.Spec.ProviderID])
	if err != nil {
		return err
	}
	c.nodes[node.Spec.ProviderID] = n
	c.nodeNameToProviderID[node.Name] = node.Spec.ProviderID
	ClusterStateNodesCount.Set(float64(len(c.nodes)), nil)
	return nil
}

func (c *Cluster) DeleteNode(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleanupNode(name)
	ClusterStateNodesCount.Set(float64(len(c.nodes)), nil)
}

func (c *Cluster) UpdatePod(ctx context.Context, pod *corev1.Pod) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	if podutils.IsTerminal(pod) {
		c.updateNodeUsageFromPodCompletion(client.ObjectKeyFromObject(pod))
	} else {
		err = c.updateNodeUsageFromPod(ctx, pod)
	}
	c.updatePodAntiAffinities(pod)
	return err
}

func (c *Cluster) NodeClaimExists(nodeClaimName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.nodeClaimNameToProviderID[nodeClaimName]
	return ok
}

// AckPods marks the pod as acknowledged for scheduling from the provisioner. This is only done once per-pod.
func (c *Cluster) AckPods(pods ...*corev1.Pod) {
	now := c.clock.Now()
	for _, pod := range pods {
		// store the value as now only if it doesn't exist.
		c.podAcks.LoadOrStore(types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, now)
	}
}

// PodAckTime will return the time the pod was first seen in our cache.
func (c *Cluster) PodAckTime(podKey types.NamespacedName) time.Time {
	if ackTime, ok := c.podAcks.Load(podKey); ok {
		return ackTime.(time.Time)
	}
	return time.Time{}
}

// MarkPodSchedulingDecisions keeps track of when we first tried to schedule a pod to a node.
// It updates podHealthyNodePoolScheduledTime for pods scheduled against nodePool that have
// NodeRegistrationHealthy=true. This also marks when the pod is first seen as schedulable for pod metrics.
// We'll only emit a metric for a pod if we haven't done it before.
// nolint:gocyclo
func (c *Cluster) MarkPodSchedulingDecisions(ctx context.Context, podErrors map[*corev1.Pod]error, npPods map[string][]*corev1.Pod, ncPods map[string][]*corev1.Pod) {
	now := c.clock.Now()
	for pod := range podErrors {
		nn := client.ObjectKeyFromObject(pod)
		// delete podsSchedulableTimes and podHealthyNodePoolScheduledTime for pods that have pod errors
		c.podsSchedulableTimes.Delete(nn)
		_, alreadyExists := c.podsSchedulingAttempted.LoadOrStore(nn, now)
		// If we already attempted this, we don't need to emit another metric.
		if !alreadyExists {
			// We should have ACK'd the pod.
			if ackTime := c.PodAckTime(nn); !ackTime.IsZero() {
				PodSchedulingDecisionSeconds.Observe(c.clock.Since(ackTime).Seconds(), nil)
			}
		}
		c.podHealthyNodePoolScheduledTime.Delete(nn)
		c.podToNodeClaim.Delete(nn)
	}
	for nodePoolName, pods := range npPods {
		nodePool := &v1.NodePool{}
		if nodePoolName != "" {
			// Swallow errors if we can't get the nodepool
			_ = c.kubeClient.Get(ctx, types.NamespacedName{Name: nodePoolName}, nodePool)
		}
		for _, p := range pods {
			nn := client.ObjectKeyFromObject(p)
			// Skip pods that are already bound to a node (e.g. pods from deleting nodes
			// included in the scheduling simulation for capacity planning). Storing a new
			// timestamp for already-bound pods would cause negative metric values since
			// their PodScheduled LastTransitionTime is in the past.
			if podutils.IsScheduled(p) {
				continue
			}
			c.podsSchedulableTimes.LoadOrStore(nn, now)
			_, alreadyExists := c.podsSchedulingAttempted.LoadOrStore(nn, now)
			// If we already attempted this, we don't need to emit another metric.
			if !alreadyExists {
				// We should have ACK'd the pod.
				if ackTime := c.PodAckTime(nn); !ackTime.IsZero() {
					PodSchedulingDecisionSeconds.Observe(c.clock.Since(ackTime).Seconds(), nil)
				}
			}
			// If the pod is scheduled to a nodePool and if the nodePool has NodeRegistrationHealthy=true
			// then mark the time when we thought it can schedule to now.
			if nodePoolName != "" && nodePool.StatusConditions().IsTrue(v1.ConditionTypeNodeRegistrationHealthy) {
				c.podHealthyNodePoolScheduledTime.LoadOrStore(nn, c.clock.Now())
			} else {
				// If the pod was scheduled to a healthy nodePool earlier but is now getting scheduled to an
				// unhealthy one then we need to delete its entry from the map because it will not schedule successfully
				c.podHealthyNodePoolScheduledTime.Delete(nn)
			}
		}
	}
	c.UpdatePodToNodeClaimMapping(ncPods)
}

func (c *Cluster) UpdatePodToNodeClaimMapping(ncPods map[string][]*corev1.Pod) {
	for ncName, pods := range ncPods {
		for _, p := range pods {
			c.podToNodeClaim.Store(client.ObjectKeyFromObject(p), ncName)
		}
	}
}

// PodSchedulingDecisionTime returns when Karpenter first decided if a pod could schedule a pod in scheduling simulations.
// This returns 0, false if Karpenter never made a decision on the pod.
func (c *Cluster) PodSchedulingDecisionTime(podKey types.NamespacedName) time.Time {
	if val, found := c.podsSchedulingAttempted.Load(podKey); found {
		return val.(time.Time)
	}
	return time.Time{}
}

// PodSchedulingSuccessTime returns when Karpenter first thought it could schedule a pod in its scheduling simulation.
// This returns 0, false if the pod was never considered in scheduling as a pending pod.
func (c *Cluster) PodSchedulingSuccessTime(podKey types.NamespacedName) time.Time {
	if val, found := c.podsSchedulableTimes.Load(podKey); found {
		return val.(time.Time)
	}
	return time.Time{}
}

// PodNodeClaimMapping returns the nodeClaim against which the pod is simulated to get scheduled
func (c *Cluster) PodNodeClaimMapping(podKey types.NamespacedName) string {
	if val, found := c.podToNodeClaim.Load(podKey); found {
		return val.(string)
	}
	return ""
}

// PodSchedulingSuccessTimeRegistrationHealthyCheck returns when Karpenter first thought it could schedule a pod in its scheduling simulation.
// This returns 0, false if the pod was never considered in scheduling as a pending pod.
func (c *Cluster) PodSchedulingSuccessTimeRegistrationHealthyCheck(podKey types.NamespacedName) time.Time {
	if val, found := c.podHealthyNodePoolScheduledTime.Load(podKey); found {
		return val.(time.Time)
	}
	return time.Time{}
}

func (c *Cluster) DeletePod(podKey types.NamespacedName) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.antiAffinityPods.Delete(podKey)
	c.updateNodeUsageFromPodCompletion(podKey)
	c.ClearPodSchedulingMappings(podKey)
	c.MarkUnconsolidated()
}

func (c *Cluster) ClearPodSchedulingMappings(podKey types.NamespacedName) {
	c.podAcks.Delete(podKey)
	c.podsSchedulableTimes.Delete(podKey)
	c.podsSchedulingAttempted.Delete(podKey)
	c.podHealthyNodePoolScheduledTime.Delete(podKey)
	c.podToNodeClaim.Delete(podKey)
}

// MarkUnconsolidated marks the cluster state as being unconsolidated.  This should be called in any situation where
// something in the cluster has changed such that the cluster may have moved from a non-consolidatable to a consolidatable
// state.
func (c *Cluster) MarkUnconsolidated() time.Time {
	newState := c.clock.Now()
	c.clusterStateMu.Lock()
	c.clusterState = newState
	c.clusterStateMu.Unlock()
	return newState
}

// ConsolidationState returns a timestamp of the last time that the cluster state with respect to consolidation changed.
// If nothing changes, this timestamp resets after five minutes to force watchers that use this to defer work to
// occasionally revalidate that nothing external (e.g. an instance type becoming available) has changed that now makes
// it possible for them to operate. Time was chosen as the type here as it allows comparisons using the built-in
// monotonic clock.
func (c *Cluster) ConsolidationState() time.Time {
	c.clusterStateMu.RLock()
	state := c.clusterState
	c.clusterStateMu.RUnlock()

	// time.Time uses a monotonic clock for these comparisons
	if c.clock.Since(state) < time.Minute*5 {
		return state
	}

	// This ensures that at least once every 5 minutes we consider consolidating our cluster in case something else has
	// changed (e.g. instance type availability) that we can't detect which would allow consolidation to occur.
	return c.MarkUnconsolidated()
}

func (c *Cluster) NodePoolResourcesFor(nodePoolName string) corev1.ResourceList {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return maps.Clone(c.nodePoolResources[nodePoolName])
}

// Reset the cluster state for unit testing
func (c *Cluster) Reset() {
	c.unsyncedTimeMu.Lock()
	defer c.unsyncedTimeMu.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clusterState = time.Time{}
	c.unsyncedStartTime = time.Time{}
	c.lastUnsyncedLogTime = time.Time{}
	c.hasSynced.Store(false)
	c.nodes = map[string]*StateNode{}
	c.nodeNameToProviderID = map[string]string{}
	c.nodeClaimNameToProviderID = map[string]string{}
	c.NodePoolState.Reset()
	c.nodePoolResources = map[string]corev1.ResourceList{}
	c.bindings = map[types.NamespacedName]string{}
	c.antiAffinityPods = sync.Map{}
	c.daemonSetPods = sync.Map{}
	c.podAcks = sync.Map{}
	c.podsSchedulingAttempted = sync.Map{}
	c.podsSchedulableTimes = sync.Map{}
	c.bufferPodCounts = map[string]int{}
}

// sets the cluster to be synced or unsynced for unit testing
func (c *Cluster) SetSynced(state bool) {
	c.unsyncedTimeMu.Lock()
	defer c.unsyncedTimeMu.Unlock()
	c.hasSynced.Store(state)
}

func (c *Cluster) GetDaemonSetPod(daemonset *appsv1.DaemonSet) *corev1.Pod {
	if pod, ok := c.daemonSetPods.Load(client.ObjectKeyFromObject(daemonset)); ok {
		return pod.(*corev1.Pod).DeepCopy()
	}

	return nil
}

func (c *Cluster) UpdateDaemonSet(ctx context.Context, daemonset *appsv1.DaemonSet) error {
	pods := &corev1.PodList{}
	// Scope down this call to only select the pods in this namespace that specifically match the DaemonSet
	// Because we get so many pods from this response, we are not DeepCopying the cached data here
	// DO NOT MUTATE pods in this function as this will affect the underlying cached pod
	if err := c.kubeClient.List(ctx, pods, client.InNamespace(daemonset.Namespace), client.UnsafeDisableDeepCopy); err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return nil
	}
	var pod *corev1.Pod
	for _, p := range pods.Items {
		if metav1.IsControlledBy(&p, daemonset) && (pod == nil || p.CreationTimestamp.After(pod.CreationTimestamp.Time)) {
			pod = &p
		}
	}
	if pod != nil {
		c.daemonSetPods.Store(client.ObjectKeyFromObject(daemonset), pod.DeepCopy())
	}
	return nil
}

func (c *Cluster) DeleteDaemonSet(key types.NamespacedName) {
	c.daemonSetPods.Delete(key)
}

// WARNING
// Everything under this section of code assumes that you have already held a lock when you are calling into these functions
// and explicitly modifying the cluster state. If you do not hold the cluster state lock before calling any of these helpers
// you will hit race conditions and data corruption

func (c *Cluster) newStateFromNodeClaim(nodeClaim *v1.NodeClaim, oldNode *StateNode) *StateNode {
	if oldNode == nil {
		oldNode = NewNode()
	}
	n := &StateNode{
		Node:              oldNode.Node,
		NodeClaim:         nodeClaim,
		daemonSetRequests: oldNode.daemonSetRequests,
		daemonSetLimits:   oldNode.daemonSetLimits,
		podRequests:       oldNode.podRequests,
		podLimits:         oldNode.podLimits,
		hostPortUsage:     oldNode.hostPortUsage,
		volumeUsage:       oldNode.volumeUsage,
		markedForDeletion: oldNode.markedForDeletion,
		nominatedUntil:    oldNode.nominatedUntil,
	}
	// Cleanup the old nodeClaim with its old providerID if its providerID changes
	// This can happen since nodes don't get created with providerIDs. Rather, CCM picks up the
	// created node and injects the providerID into the spec.providerID
	if id, ok := c.nodeClaimNameToProviderID[nodeClaim.Name]; ok && id != nodeClaim.Status.ProviderID {
		c.cleanupNodeClaim(nodeClaim.Name)
	}
	c.updateNodePoolResources(oldNode, n)
	c.triggerConsolidationOnChange(oldNode, n)
	return n
}

func (c *Cluster) cleanupNodeClaim(name string) {
	if id := c.nodeClaimNameToProviderID[name]; id != "" {
		if c.nodes[id].Node == nil {
			c.updateNodePoolResources(c.nodes[id], nil)
			delete(c.nodes, id)
		} else {
			oldNode := c.nodes[id].ShallowCopy()
			c.nodes[id].NodeClaim = nil
			c.updateNodePoolResources(oldNode, c.nodes[id])
		}
		c.MarkUnconsolidated()
	}
	// Delete the node claim from the nodeClaimNameToProviderID in the case that the provider ID hasn't resolved
	// yet. This ensures that if a nodeClaim is created and then deleted before it was able to launch that
	// this is cleaned up.
	delete(c.nodeClaimNameToProviderID, name)

	// Delete the NodeClaim that is tracked in NodePoolState
	c.NodePoolState.Cleanup(name)
}

func (c *Cluster) newStateFromNode(ctx context.Context, node *corev1.Node, oldNode *StateNode) (*StateNode, error) {
	if oldNode == nil {
		oldNode = NewNode()
	}
	n := &StateNode{
		Node:              node,
		NodeClaim:         oldNode.NodeClaim,
		daemonSetRequests: map[types.NamespacedName]corev1.ResourceList{},
		daemonSetLimits:   map[types.NamespacedName]corev1.ResourceList{},
		podRequests:       map[types.NamespacedName]corev1.ResourceList{},
		podLimits:         map[types.NamespacedName]corev1.ResourceList{},
		hostPortUsage:     scheduling.NewHostPortUsage(),
		volumeUsage:       scheduling.NewVolumeUsage(),
		markedForDeletion: oldNode.markedForDeletion,
		nominatedUntil:    oldNode.nominatedUntil,
	}
	if err := multierr.Combine(
		c.populateResourceRequests(ctx, n),
		c.populateVolumeLimits(ctx, n),
	); err != nil {
		return nil, err
	}
	// Cleanup the old node with its old providerID if its providerID changes
	// This can happen since nodes don't get created with providerIDs. Rather, CCM picks up the
	// created node and injects the providerID into the spec.providerID
	if id, ok := c.nodeNameToProviderID[node.Name]; ok && id != node.Spec.ProviderID {
		c.cleanupNode(node.Name)
	}
	c.updateNodePoolResources(oldNode, n)
	c.triggerConsolidationOnChange(oldNode, n)
	return n, nil
}

func (c *Cluster) cleanupNode(name string) {
	if id := c.nodeNameToProviderID[name]; id != "" {
		if c.nodes[id].NodeClaim == nil {
			c.updateNodePoolResources(c.nodes[id], nil)
			delete(c.nodes, id)
		} else {
			oldNode := c.nodes[id].ShallowCopy()
			c.nodes[id].Node = nil
			c.updateNodePoolResources(oldNode, c.nodes[id])
		}
		delete(c.nodeNameToProviderID, name)
		c.MarkUnconsolidated()
	}
}

// nolint:gocyclo
func (c *Cluster) updateNodePoolResources(oldNode, newNode *StateNode) {
	var oldNodePoolName, newNodePoolName string
	var oldResources, newResources corev1.ResourceList
	if oldNode != nil && (oldNode.Node != nil || oldNode.NodeClaim != nil) {
		oldNodePoolName = oldNode.Labels()[v1.NodePoolLabelKey]
		oldResources = lo.Ternary(oldNode.MarkedForDeletion(), corev1.ResourceList{}, oldNode.Capacity())
	}
	if newNode != nil && (newNode.Node != nil || newNode.NodeClaim != nil) {
		newNodePoolName = newNode.Labels()[v1.NodePoolLabelKey]
		newResources = lo.Ternary(newNode.MarkedForDeletion(), corev1.ResourceList{}, newNode.Capacity())
	}
	if len(oldResources) != 0 {
		oldResources[resources.Node] = resource.MustParse("1")
	}
	if len(newResources) != 0 {
		newResources[resources.Node] = resource.MustParse("1")
	}
	if _, ok := c.nodePoolResources[newNodePoolName]; !ok && newNodePoolName != "" {
		c.nodePoolResources[newNodePoolName] = corev1.ResourceList{}
	}
	if oldNodePoolName != "" {
		for resourceName, quantity := range oldResources {
			current := c.nodePoolResources[oldNodePoolName][resourceName]
			current.Sub(quantity)
			c.nodePoolResources[oldNodePoolName][resourceName] = current
		}
	}
	if newNodePoolName != "" {
		for resourceName, quantity := range newResources {
			current := c.nodePoolResources[newNodePoolName][resourceName]
			current.Add(quantity)
			c.nodePoolResources[newNodePoolName][resourceName] = current
		}
	}
	// Garbage collect any NodePool keys that no longer have any resources assigned to them.
	// We do this when there are no longer any NodeClaims that map to this NodePool
	// so that we don't leak NodePool keys in our nodePoolResources map
	for _, name := range []string{oldNodePoolName, newNodePoolName} {
		allZero := true
		for _, v := range c.nodePoolResources[name] {
			if !v.IsZero() {
				allZero = false
				break
			}
		}
		if allZero {
			delete(c.nodePoolResources, name)
		}
	}
}

func (c *Cluster) populateVolumeLimits(ctx context.Context, n *StateNode) error {
	var csiNode storagev1.CSINode
	if err := c.kubeClient.Get(ctx, client.ObjectKey{Name: n.Node.Name}, &csiNode); err != nil {
		return client.IgnoreNotFound(serrors.Wrap(fmt.Errorf("getting CSINode to determine volume limit, %w", err), "CSINode", klog.KRef("", n.Node.Name)))
	}
	for _, driver := range csiNode.Spec.Drivers {
		if driver.Allocatable == nil {
			continue
		}
		n.volumeUsage.AddLimit(driver.Name, int(lo.FromPtr(driver.Allocatable.Count)))
	}
	return nil
}

func (c *Cluster) populateResourceRequests(ctx context.Context, n *StateNode) error {
	var pods corev1.PodList
	if err := c.kubeClient.List(ctx, &pods, client.MatchingFields{"spec.nodeName": n.Node.Name}); err != nil {
		return fmt.Errorf("listing pods, %w", err)
	}
	for i := range pods.Items {
		pod := &pods.Items[i]
		if podutils.IsTerminal(pod) {
			continue
		}
		if err := n.updateForPod(ctx, c.kubeClient, pod); err != nil {
			return err
		}
		c.cleanupOldBindings(pod)
		c.bindings[client.ObjectKeyFromObject(pod)] = pod.Spec.NodeName
	}
	return nil
}

// updateNodeUsageFromPod is called every time a reconcile event occurs for the pod. If the pods binding has changed
// (unbound to bound), we need to update the resource requests on the node.
func (c *Cluster) updateNodeUsageFromPod(ctx context.Context, pod *corev1.Pod) error {
	// nothing to do if the pod isn't bound, checking early allows avoiding unnecessary locking
	if pod.Spec.NodeName == "" {
		return nil
	}

	n, ok := c.nodes[c.nodeNameToProviderID[pod.Spec.NodeName]]
	if !ok {
		// the node must exist for us to update the resource requests on the node
		return errors.NewNotFound(schema.GroupResource{Resource: "Node"}, pod.Spec.NodeName)
	}
	if err := n.updateForPod(ctx, c.kubeClient, pod); err != nil {
		return err
	}
	c.cleanupOldBindings(pod)
	c.bindings[client.ObjectKeyFromObject(pod)] = pod.Spec.NodeName
	return nil
}

func (c *Cluster) updateNodeUsageFromPodCompletion(podKey types.NamespacedName) {
	nodeName, bindingKnown := c.bindings[podKey]
	if !bindingKnown {
		// we didn't think the pod was bound, so we weren't tracking it and don't need to do anything
		return
	}

	delete(c.bindings, podKey)
	n, ok := c.nodes[c.nodeNameToProviderID[nodeName]]
	if !ok {
		// we weren't tracking the node yet, so nothing to do
		return
	}
	n.cleanupForPod(podKey)
}

func (c *Cluster) cleanupOldBindings(pod *corev1.Pod) {
	if oldNodeName, bindingKnown := c.bindings[client.ObjectKeyFromObject(pod)]; bindingKnown {
		if oldNodeName == pod.Spec.NodeName {
			// we are already tracking the pod binding, so nothing to update
			return
		}
		// the pod has switched nodes, this can occur if a pod name was re-used, and it was deleted/re-created rapidly,
		// binding to a different node the second time
		if oldNode, ok := c.nodes[c.nodeNameToProviderID[oldNodeName]]; ok {
			// we were tracking the old node, so we need to reduce its capacity by the amount of the pod that left
			oldNode.cleanupForPod(client.ObjectKeyFromObject(pod))
			delete(c.bindings, client.ObjectKeyFromObject(pod))
		}
	}
	// new pod binding has occurred
	c.MarkUnconsolidated()
}

func (c *Cluster) updatePodAntiAffinities(pod *corev1.Pod) {
	// We intentionally don't track inverse anti-affinity preferences. We're not
	// required to enforce them so it just adds complexity for very little
	// value. The problem with them comes from the relaxation process, the pod
	// we are relaxing is not the pod with the anti-affinity term.
	if podKey := client.ObjectKeyFromObject(pod); podutils.HasRequiredPodAntiAffinity(pod) {
		c.antiAffinityPods.Store(podKey, pod)
	} else {
		c.antiAffinityPods.Delete(podKey)
	}
}

func (c *Cluster) triggerConsolidationOnChange(old, new *StateNode) {
	if old == nil || new == nil {
		c.MarkUnconsolidated()
		return
	}
	// If either the old node or new node are mocked
	if (old.Node == nil && old.NodeClaim == nil) || (new.Node == nil && new.NodeClaim == nil) {
		c.MarkUnconsolidated()
		return
	}
	if old.Initialized() != new.Initialized() {
		c.MarkUnconsolidated()
		return
	}
	if old.MarkedForDeletion() != new.MarkedForDeletion() {
		c.MarkUnconsolidated()
		return
	}
}

// HasSynced returns whether the cluster state has been synchronized at least once.
func (c *Cluster) HasSynced() bool {
	return c.hasSynced.Load()
}
