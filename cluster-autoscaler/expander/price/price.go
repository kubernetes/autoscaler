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
*/

package price

import (
	"fmt"
	"math"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	klog "k8s.io/klog/v2"
)

// *************
// The detailed description of what is going on in this expander can be found
// here:
// https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/proposals/pricing.md
// **********

type priceBased struct {
	cloudProvider         cloudprovider.CloudProvider
	preferredNodeProvider PreferredNodeProvider
	nodeUnfitness         NodeUnfitness
}

var (
	// defaultPreferredNode is the node that is preferred if PreferredNodeProvider fails.
	// 4 cpu, 16gb ram.
	defaultPreferredNode = buildNode(4*1000, 4*4*units.GiB)

	// priceStabilizationPod is the pod cost to stabilize node_cost/pod_cost ratio a bit.
	// 0.5 cpu, 500 mb ram
	priceStabilizationPod = buildPod("stabilize", 500, 500*units.MiB)

	// Penalty given to node groups that are yet to be created.
	// TODO: make it a flag
	// TODO: investigate what a proper value should be
	notExistCoeficient = 2.0

	// This value will be used as unfitness for node groups using GPU. This serves
	// 2 purposes:
	// - It makes nodes with GPU extremely unattractive to expander, so it will never
	//   use nodes with expensive GPUs for pods that don't require it.
	// - By overriding unfitness for node groups with GPU we ignore preferred cluster
	//   shape when comparing such node groups. Node unfitness logic is meant to
	//   minimize per-node cost (resources consumed by kubelet, kube-proxy, etc) and
	//   resource fragmentation, while avoiding putting a significant fraction of all
	//   pods on a single node for availability reasons.
	//   Those goals don't apply well to nodes with GPUs that are generally dedicated
	//   for specific workload and need to be optimized for GPU utilization, not CPU
	//   utilization.
	gpuUnfitnessOverride = 1000.0
)

// NewStrategy returns an expansion strategy that picks nodes based on price and preferred node type.
func NewStrategy(cloudProvider cloudprovider.CloudProvider,
	preferredNodeProvider PreferredNodeProvider,
	nodeUnfitness NodeUnfitness,
) expander.Strategy {
	return &priceBased{
		cloudProvider:         cloudProvider,
		preferredNodeProvider: preferredNodeProvider,
		nodeUnfitness:         nodeUnfitness,
	}
}

// BestOption selects option based on cost and preferred node type.
func (p *priceBased) BestOption(expansionOptions []expander.Option, nodeInfos map[string]*schedulerframework.NodeInfo) *expander.Option {
	var bestOption *expander.Option
	bestOptionScore := 0.0
	now := time.Now()
	then := now.Add(time.Hour)

	preferredNode, err := p.preferredNodeProvider.Node()
	if err != nil {
		klog.Errorf("Failed to get preferred node, switching to default: %v", err)
		preferredNode = defaultPreferredNode
	}

	pricingModel, err := p.cloudProvider.Pricing()
	if err != nil {
		klog.Errorf("Failed to get pricing model from cloud provider: %v", err)
	}

	stabilizationPrice, err := pricingModel.PodPrice(priceStabilizationPod, now, then)
	if err != nil {
		klog.Errorf("Failed to get price for stabilization pod: %v", err)
		// continuing without stabilization.
	}

nextoption:
	for _, option := range expansionOptions {
		nodeInfo, found := nodeInfos[option.NodeGroup.Id()]
		if !found {
			klog.Warningf("No node info for %s", option.NodeGroup.Id())
			continue
		}
		nodePrice, err := pricingModel.NodePrice(nodeInfo.Node(), now, then)
		if err != nil {
			klog.Warningf("Failed to calculate node price for %s: %v", option.NodeGroup.Id(), err)
			continue
		}
		totalNodePrice := nodePrice * float64(option.NodeCount)
		totalPodPrice := 0.0
		for _, pod := range option.Pods {
			podPrice, err := pricingModel.PodPrice(pod, now, then)
			if err != nil {
				klog.Warningf("Failed to calculate pod price for %s/%s: %v", pod.Namespace, pod.Name, err)
				continue nextoption
			}
			totalPodPrice += podPrice
		}
		// Total pod price is 0 when the pods have no requests. The pods must have some other
		// requirements that prevent them from scheduling like AntiAffinity, HostPort or the
		// pods quota on all nodes has been already used. We use stabilizationPrice in the formula
		// below so this should not be a problem.

		// How well the money is spent.
		priceSubScore := (totalNodePrice + stabilizationPrice) / (totalPodPrice + stabilizationPrice)
		// How well the node matches generic cluster needs
		nodeUnfitness := p.nodeUnfitness(preferredNode, nodeInfo.Node())

		// TODO: normalize node count against preferred node.
		supressedUnfitness := (nodeUnfitness-1.0)*(1.0-math.Tanh(float64(option.NodeCount-1)/15.0)) + 1.0

		// Set constant, very high unfitness to make them unattractive for pods that doesn't need GPU and
		// avoid optimizing them for CPU utilization.
		if gpu.NodeHasGpu(p.cloudProvider.GPULabel(), nodeInfo.Node()) {
			klog.V(4).Infof("Price expander overriding unfitness for node group with GPU %s", option.NodeGroup.Id())
			supressedUnfitness = gpuUnfitnessOverride
		}

		optionScore := supressedUnfitness * priceSubScore

		if !option.NodeGroup.Exist() {
			optionScore *= notExistCoeficient
		}

		debug := fmt.Sprintf("all_nodes_price=%f pods_price=%f stabilized_ratio=%f unfitness=%f suppressed=%f final_score=%f",
			totalNodePrice,
			totalPodPrice,
			priceSubScore,
			nodeUnfitness,
			supressedUnfitness,
			optionScore,
		)

		klog.V(5).Infof("Price expander for %s: %s", option.NodeGroup.Id(), debug)

		if bestOption == nil || bestOptionScore > optionScore {
			bestOption = &expander.Option{
				NodeGroup: option.NodeGroup,
				NodeCount: option.NodeCount,
				Debug:     fmt.Sprintf("%s | price-expander: %s", option.Debug, debug),
				Pods:      option.Pods,
			}
			bestOptionScore = optionScore
		}
	}
	return bestOption
}

// buildPod creates a pod with specified resources.
func buildPod(name string, millicpu int64, mem int64) *apiv1.Pod {
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			SelfLink:  fmt.Sprintf("/api/v1/namespaces/default/pods/%s", name),
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(millicpu, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(mem, resource.DecimalSI),
						},
					},
				},
			},
		},
	}
}
