package placeholderpod

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
)

// CreateReplicaSets create one or more replica set(s) of placeholder pods
func CreateReplicaSets(sets ...ReplicaSet) []*apiv1.Pod {
	pods := []*apiv1.Pod{}
	for _, set := range sets {
		for i := int64(0); i < set.Count; i++ {
			podSpec := set.PodSpec
			nodeStickiness := podSpec.NodeStickiness
			milliCPU := podSpec.MilliCPU
			memory := podSpec.Memory
			req := apiv1.ResourceList{}
			req[apiv1.ResourceCPU] = *resource.NewMilliQuantity(milliCPU, resource.DecimalSI)
			req[apiv1.ResourceMemory] = *resource.NewQuantity(memory, resource.DecimalSI)
			var name string
			if set.Count > 1 {
				name = fmt.Sprintf("%s-%d", set.Name, i)
			} else {
				name = set.Name
			}
			pod := &apiv1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:              fmt.Sprintf("placeholder-%s-%dm-%dMB", name, milliCPU, memory/1024/1024),
					CreationTimestamp: metav1.NewTime(time.Now()),
					// Without a namespace, this causes an error like:
					// > POST https://10.3.0.1:443/api/v1/namespaces/default/events 422 Unprocessable Entity in 1 milliseconds
					// due to an invalid pod:
					// > 'Event "placeholder-12-2800m-5626449100.14bd31ca98a9fa02" is invalid: involvedObject.namespace: Required value: required for kind Pod' (will not retry!)
					Namespace: "default",
				},
				Spec: apiv1.PodSpec{
					Affinity: &apiv1.Affinity{
						NodeAffinity: &apiv1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: nodeStickiness.NodeAffinityRequiredTerms,
						},
						PodAntiAffinity: &apiv1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: nodeStickiness.PodAntiAffinityRequiredTerms,
						},
					},
					NodeSelector: nodeStickiness.NodeSelector,
					Containers: []apiv1.Container{
						{
							Resources: apiv1.ResourceRequirements{
								Requests: req,
							},
						},
					},
				},
				Status: apiv1.PodStatus{},
			}
			pods = append(pods, pod)
		}
	}
	return pods
}
