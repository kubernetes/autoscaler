package podsharding

import (
	"k8s.io/api/core/v1"
)

// NewOssPodSharder returns a PodSharder that shards pods based on the podShardingLabels specified.
func NewOssPodSharder(podShardingLabels map[string]string) PodSharder {
	computeFunctions := []FeatureShardComputeFunction{
		{
			"label",
			computeLabelNameShard(podShardingLabels),
		},
	}
	return NewCompositePodSharder(computeFunctions)
}

// computeLabelNameShard returns a function with nodeGroupDescriptor labels updated to contain the podShardingLabels if they are present in the pod.
func computeLabelNameShard(podShardingLabels map[string]string) func(*v1.Pod, *NodeGroupDescriptor) {
	return func(pod *v1.Pod, nodeGroupDescriptor *NodeGroupDescriptor) {
		if len(podShardingLabels) == 0 {
			return
		}

		for labelKey, labelValue := range podShardingLabels {
			if podLabelValue, ok := pod.Spec.NodeSelector[labelKey]; ok && podLabelValue == labelValue {
				nodeGroupDescriptor.Labels[labelKey] = labelValue
			}
		}
	}
}
