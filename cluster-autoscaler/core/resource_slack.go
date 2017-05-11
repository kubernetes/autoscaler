package core

import (
	"fmt"
	"math"

	"k8s.io/autoscaler/cluster-autoscaler/core/placeholderpod"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"

	"k8s.io/apimachinery/pkg/api/resource"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// CreateAndSchedulePlaceholderPods creates placeholder pods and try to virtually assign pods to nodes so that CA keeps more nodes than actually necessary
func CreateAndSchedulePlaceholderPods(readyNodes []*apiv1.Node, allScheduled []*apiv1.Pod, autoscalingContext *AutoscalingContext, predicateChecker *simulator.PredicateChecker) ([]*apiv1.Pod, []*apiv1.Pod, error) {
	toGB := func(value float64) float64 {
		return value / 1024 / 1024 / 1024
	}

	// TODO Could be simplified more
	// Ty our best to make all the words match to the original design at https://github.com/kubernetes/contrib/issues/2251#issuecomment-272893268
	clusterSize := int64(len(readyNodes))
	clusterSizeInFloat := float64(clusterSize)
	// Foundamental statistics used for creating placeholder pods matching the reality
	// Use Quantity where possible not to introduce bugs due to numeric conversion among different formats
	clusterCpuSum := &resource.Quantity{}
	clusterMemorySum := &resource.Quantity{}
	minNodeMilliCPU := int64(math.MaxInt64)
	minNodeMemory := int64(math.MaxInt64)
	for _, n := range readyNodes {
		if cpu := n.Status.Capacity.Cpu(); cpu != nil {
			clusterCpuSum.Add(*cpu)
			v := cpu.MilliValue()
			if v < minNodeMilliCPU {
				minNodeMilliCPU = v
			}
		}
		if mem := n.Status.Capacity.Memory(); mem != nil {
			clusterMemorySum.Add(*mem)
			v := mem.Value()
			if v < minNodeMemory {
				minNodeMemory = v
			}
		}
	}
	// Resource occupied by each node
	avgNodeMilliCPU := clusterCpuSum.MilliValue() / clusterSize
	avgNodeMemory := float64(clusterMemorySum.Value()) / clusterSizeInFloat
	// Resource occupied by the cluster
	clusterMilliCPU := avgNodeMilliCPU * clusterSize
	clusterMemory := avgNodeMemory * clusterSizeInFloat
	// Resource should be occupied by placeholder pods
	minExtraCapacityRate := autoscalingContext.MinExtraCapacityRate
	clusterExtraMilliCPU := int64(float64(clusterMilliCPU) * minExtraCapacityRate)
	clusterExtraMemory := clusterMemory * minExtraCapacityRate
	// Mix different types of placeholder pods to ensure both:
	// (1) providing extra capacity even for the largest pod in the cluster and
	// (2) not wasting cluster capacity by making all the placeholder pods too big
	//
	// * Note that (1) could be made configurable if necessary. How about capping the max size of a placeholder pod via
	// flags --placeholder-pod-max-cpu and --placeholder-pod-max-memory so that CA won't waste money by keeping
	// considerably large nodes for the latest pod.
	//
	// In a nutshell, we prefer choosing a node group with a largest machine type for scaling up so that
	// every pod can have a chance to be scheduled faster. Choosing a too small machine type for scaling up result in
	// large pods to not fit within extra capacity, which breaks the original purpose of the resource slack feature(--min-extra-capacity-rate)
	//
	// General rules
	// * Every place holder pod should be small enough to fit within at least one of nodes
	//   For example, a 30GB placeholder pod which doesn't fit to any small node which results in more nodes doesn't make sense.
	// * Exactly one place holder pod should be sized the same as the largest(in terms of resource requests) pod
	//   so that CA can make extra space even for the largest pod.
	//   Beware that this results in CA to choose a node group for the larger machine type as a scaling target.
	// * All the other placeholder pods should be small enough not to unnecessarily occupy capacity to add unnecessary nodes
	//   For example, a 5GB placeholder pod doesn't fit to the smallest node doesn't make sense
	maxPodRequestedMilliCPU := int64(0)
	maxPodRequestedMemory := int64(0)
	for _, p := range allScheduled {
		maxContainerMilliCPU := int64(0)
		maxContainerMemory := int64(0)
		for _, c := range p.Spec.Containers {
			cpuRes := c.Resources.Requests[apiv1.ResourceCPU]
			cpu := (&cpuRes).MilliValue()
			if cpu > maxContainerMilliCPU {
				maxContainerMilliCPU = cpu
			}
			memRes := c.Resources.Requests[apiv1.ResourceMemory]
			mem := (&memRes).Value()
			if mem > maxContainerMemory {
				maxContainerMemory = mem
			}
		}
		if maxContainerMilliCPU > maxPodRequestedMilliCPU {
			maxPodRequestedMilliCPU = maxContainerMilliCPU
		}
		if maxContainerMemory > maxPodRequestedMemory {
			maxPodRequestedMemory = maxContainerMemory
		}
	}

	// Resource required for all the remaining placeholder pods
	remainingClusterExtraMilliCPU := clusterExtraMilliCPU - maxPodRequestedMilliCPU
	remainingClusterExtraMemory := clusterExtraMemory - float64(maxPodRequestedMemory)
	// Could be computed via e.g. ceil(minNodeMilliCPU / avgPodMilliCPU) but then it would result in less deterministic behavior of CA
	// because avgPodMilliCPU would change greatly over time
	granularity := int64(5)
	// Resource required for each of other placeholder pods
	// This is intended to be the amount of CPU requested by a typical pod running in the cluster.
	// The higher this value is, the more cluster capacity is wasted.
	// ppMilliCPU is the amount of CPU requested for a placeholder pod should be smaller than the CPU of the minimum node
	// It should be less than or equal to maxPodRequestedMilliCPU not to instruct CA to add unnecessarily large nodes
	ppMilliCPU := minNodeMilliCPU / granularity
	if ppMilliCPU > maxPodRequestedMilliCPU {
		ppMilliCPU = maxPodRequestedMilliCPU
	}
	// ppMemory is the amount of memory requested for a placeholder pod should be smaller than the memory of the minimum node
	// It should be less than or equal to maxPodRequestedMemory not to instruct CA to add unnecessarily large nodes
	ppMemory := minNodeMemory / granularity
	if ppMemory > maxPodRequestedMemory {
		ppMemory = maxPodRequestedMemory
	}
	// Don't round but ceil to ensure the rate specified via --min-extra-capacity-rate is met
	extraPlaceholderPodsCount := int64(math.Ceil(float64(remainingClusterExtraMilliCPU) / float64(ppMilliCPU)))
	if extraPlaceholderPodsCount < 0 {
		extraPlaceholderPodsCount = 0
	}

	glog.V(4).Infof(
		"Creating placeholder pods based on the following observations: "+
			"cluster size = %v, cluster cpu sum = %s cores, cluster memory sum = %s, avg milli cpu per node = %v, avg memory per node = %.1fGB, "+
			"cluster total milli cpu = %v, cluster total memory = %.1fGB, cluster extra milli cpu = %v, cluster extra memory = %.1fGB, "+
			"max pod milli cpu = %d, max pod memory = %.1fGB, remaining cluster extra milli cpu = %d, remaining cluster extra memory = %.1fGB, "+
			"min node milli cpu = %d, min node memory = %.1fGB, "+
			"milli cpu per extra placeholder pod = %d, memory per extra placeholder pod = %.1fGB, extra placeholder pods count = %v",
		clusterSize,
		clusterCpuSum.String(),
		clusterMemorySum.String(),
		avgNodeMilliCPU,
		toGB(avgNodeMemory),
		clusterMilliCPU,
		toGB(clusterMemory),
		clusterExtraMilliCPU,
		toGB(clusterExtraMemory),
		maxPodRequestedMilliCPU,
		toGB(float64(maxPodRequestedMemory)),
		remainingClusterExtraMilliCPU,
		toGB(remainingClusterExtraMemory),
		minNodeMilliCPU,
		toGB(float64(minNodeMemory)),
		ppMilliCPU,
		toGB(float64(ppMemory)),
		extraPlaceholderPodsCount,
	)

	nodeStickiness := placeholderpod.NodeStickiness{
		NodeSelector:                 map[string]string{},
		PodAntiAffinityRequiredTerms: []apiv1.PodAffinityTerm{},
		NodeAffinityRequiredTerms:    nil,
	}

	largestCPU := placeholderpod.New(
		int64(maxPodRequestedMilliCPU),
		int64(0),
		nodeStickiness,
	)

	largestMemory := placeholderpod.New(
		int64(0),
		int64(maxPodRequestedMemory),
		nodeStickiness,
	)

	standard := placeholderpod.New(
		int64(ppMilliCPU),
		int64(ppMemory),
		nodeStickiness,
	)

	// TODO Create at least one placeholder pod for every occurrence of a set of nodeSelector + affinity + tolerations
	// so that even pods demanding such specific requirements on nodes can have extra capacity
	// TODO Pods dedicated to master nodes should be ignored as master nodes shouldn't be auto-scaled
	placeholderPods := placeholderpod.CreateReplicaSets(
		placeholderpod.ReplicaSet{Name: "largestcpu", Count: 1, PodSpec: largestCPU},
		placeholderpod.ReplicaSet{Name: "largestmemory", Count: 1, PodSpec: largestMemory},
		placeholderpod.ReplicaSet{Name: "standard", Count: extraPlaceholderPodsCount, PodSpec: standard},
	)

	glog.V(4).Infof("Created %d placeholder pods", len(placeholderPods))

	return schedulePlaceholderPods(placeholderPods, readyNodes, allScheduled, predicateChecker)
}

// schedulePlaceholderPods try to fit placeholder pods to actual nodes and virtually assign pods
// to the fitted nodes so that CA keeps more nodes than actually necessary
func schedulePlaceholderPods(placeholderPods []*apiv1.Pod, nodes []*apiv1.Node, scheduledPods []*apiv1.Pod, predicateChecker *simulator.PredicateChecker) ([]*apiv1.Pod, []*apiv1.Pod, error) {
	scheduledPlaceholderPods := []*apiv1.Pod{}
	unscheduledPlaceholderPods := []*apiv1.Pod{}
	nodeNameToNodeInfo := createNodeNameToInfoMap(scheduledPods, nodes)

	for _, pod := range placeholderPods {
		if nodeName, err := predicateChecker.FitsAny(pod, nodeNameToNodeInfo); err == nil {
			glog.Infof("Pod %s is virtually scheduled on %s", pod.Name, nodeName)

			scheduledPlaceholderPod := *pod
			scheduledPlaceholderPod.Spec.NodeName = nodeName

			scheduledPlaceholderPods = append(scheduledPlaceholderPods, &scheduledPlaceholderPod)

			oldNodeInfo := nodeNameToNodeInfo[nodeName]
			newPods := oldNodeInfo.Pods()
			newPods = append(newPods, &scheduledPlaceholderPod)
			newNodeInfo := schedulercache.NewNodeInfo(newPods...)
			if err := newNodeInfo.SetNode(oldNodeInfo.Node()); err != nil {
				return []*apiv1.Pod{}, []*apiv1.Pod{}, fmt.Errorf("failed to schedule placeholder pods: %v", err)
			}

			nodeNameToNodeInfo[nodeName] = newNodeInfo
		} else {
			unscheduledPlaceholderPods = append(unscheduledPlaceholderPods, pod)
		}
	}

	return scheduledPlaceholderPods, unscheduledPlaceholderPods, nil
}
