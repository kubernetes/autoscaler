/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package common

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	npconsts "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/nodepools/consts"
	kubeletapis "k8s.io/kubelet/pkg/apis"
	"k8s.io/kubernetes/pkg/apis/scheduling"
	"math/rand"
	"regexp"
	"strings"
)

// BuildCSINodePod builds a template of the CSI Node Driver pod
func BuildCSINodePod() *apiv1.Pod {
	priority := scheduling.SystemCriticalPriority
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("csi-oci-node-%d", rand.Int63()),
			Namespace: "kube-system",
			Labels: map[string]string{
				"app": "csi-oci-node",
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Image: "iad.ocir.io/oracle/cloud-provider-oci:latest",
				},
			},
			Priority: &priority,
		},
		Status: apiv1.PodStatus{
			Phase: apiv1.PodRunning,
			Conditions: []apiv1.PodCondition{
				{
					Type:   apiv1.PodReady,
					Status: apiv1.ConditionTrue,
				},
			},
		},
	}
}

// BuildProxymuxClientPod builds a template of the Proxymux Client pod
func BuildProxymuxClientPod() *apiv1.Pod {
	priority := scheduling.SystemCriticalPriority
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("proxymux-client-ds-%d", rand.Int63()),
			Namespace: "kube-system",
			Labels: map[string]string{
				"oke-app": "proxymux-client-ds",
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Image: "proxymux-client",
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(int64(50), resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(int64(64), resource.BinarySI),
						},
					},
				},
			},
			Priority: &priority,
		},
		Status: apiv1.PodStatus{
			Phase: apiv1.PodRunning,
			Conditions: []apiv1.PodCondition{
				{
					Type:   apiv1.PodReady,
					Status: apiv1.ConditionTrue,
				},
			},
		},
	}
}

// BuildFlannelPod builds a template of the Flannel pod
func BuildFlannelPod() *apiv1.Pod {
	priority := scheduling.SystemCriticalPriority
	return &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("kube-flannel-ds-%d", rand.Int63()),
			Namespace: "kube-system",
			Labels: map[string]string{
				"app":  "flannel",
				"tier": "node",
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Image: "kube-flannel",
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(int64(100), resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(int64(50), resource.BinarySI),
						},
					},
				},
			},
			Priority: &priority,
		},
		Status: apiv1.PodStatus{
			Phase: apiv1.PodRunning,
			Conditions: []apiv1.PodCondition{
				{
					Type:   apiv1.PodReady,
					Status: apiv1.ConditionTrue,
				},
			},
		},
	}
}

// BuildGenericLabels defines all the default labels that nodes should have
func BuildGenericLabels(ocid string, nodeName, shape, availabilityDomain string) map[string]string {
	result := make(map[string]string)
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[apiv1.LabelArchStable] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS
	result[apiv1.LabelOSStable] = cloudprovider.DefaultOS

	parts := strings.Split(ocid, ".")
	if len(parts) >= 5 {
		// backward compatibility with older pod labels
		result[apiv1.LabelZoneRegion] = parts[3]
		result[apiv1.LabelZoneRegionStable] = parts[3]

		compiledArmRegexp := regexp.MustCompile("\\.A[0-9]+\\.") // Matches node shapes with the pattern '.A<number>.'
		if compiledArmRegexp.MatchString(shape) {
			result[kubeletapis.LabelArch] = npconsts.ArmArch
			result[apiv1.LabelArchStable] = npconsts.ArmArch
		}
	}

	result[apiv1.LabelZoneFailureDomain] = availabilityDomain
	result[apiv1.LabelZoneFailureDomainStable] = availabilityDomain

	result[apiv1.LabelHostname] = nodeName

	result[apiv1.LabelInstanceType] = shape
	result[apiv1.LabelInstanceTypeStable] = shape

	return result
}
