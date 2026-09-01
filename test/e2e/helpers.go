//go:build e2e

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

package e2e

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

const (
	defaultEventTimeout     = 2 * time.Minute
	defaultPodTimeout       = 2 * time.Minute
	defaultNodeTimeout      = 2 * time.Minute
	defaultScaleDownTimeout = 4 * time.Minute
	kwokTaintKey            = "kwok-provider"
	nodegroupLabelKey       = "kwok-nodegroup"
)

// NewTestPod creates a test pod configuration targeted at a specific nodegroup.
func NewTestPod(name, namespace, nodegroup, cpu, memory string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "fake-container",
					Image: "fake-image",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse(cpu),
							corev1.ResourceMemory: resource.MustParse(memory),
						},
					},
				},
			},
			NodeSelector: map[string]string{
				nodegroupLabelKey: nodegroup,
			},
			TerminationGracePeriodSeconds: new(int64),
			Tolerations: []corev1.Toleration{
				{
					Key:      kwokTaintKey,
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
	}
}

// NewTestPodWithPriority creates a test pod with a specific priority class name.
func NewTestPodWithPriority(name, namespace, nodegroup, cpu, memory, priorityClassName string) *corev1.Pod {
	pod := NewTestPod(name, namespace, nodegroup, cpu, memory)
	pod.Spec.PriorityClassName = priorityClassName
	return pod
}

// NewTestPodWithAntiAffinity creates a pod that specifies pod anti-affinity.
func NewTestPodWithAntiAffinity(name, namespace, nodegroup, matchLabelKey, matchLabelValue, cpu, memory string) *corev1.Pod {
	pod := NewTestPod(name, namespace, nodegroup, cpu, memory)
	pod.Labels[matchLabelKey] = matchLabelValue
	pod.Spec.Affinity = &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      matchLabelKey,
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{matchLabelValue},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		},
	}
	return pod
}

// NewTestDeployment creates a deployment that targets a specific nodegroup.
func NewTestDeployment(name, namespace, nodegroup string, replicas int32, cpu, memory string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "fake-container",
							Image: "fake-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(cpu),
									corev1.ResourceMemory: resource.MustParse(memory),
								},
							},
						},
					},
					NodeSelector: map[string]string{
						nodegroupLabelKey: nodegroup,
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      kwokTaintKey,
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
	}
}

// NewTestPDB creates a PodDisruptionBudget.
func NewTestPDB(name, namespace, matchLabelKey, matchLabelValue string, minAvailable, maxUnavailable *intstr.IntOrString) *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable:   minAvailable,
			MaxUnavailable: maxUnavailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					matchLabelKey: matchLabelValue,
				},
			},
		},
	}
}

// NewPriorityClass creates a PriorityClass object.
func NewPriorityClass(name string, value int32, globalDefault bool) *schedulingv1.PriorityClass {
	return &schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Value:         value,
		GlobalDefault: globalDefault,
	}
}

// CleanUpNodeGroup deletes all fake nodes for the given nodegroup and waits until count is 0.
func CleanUpNodeGroup(ctx context.Context, client klient.Client, nodegroup string) error {
	nodeList := &corev1.NodeList{}
	err := client.Resources().List(ctx, nodeList)
	if err != nil {
		return err
	}
	for _, node := range nodeList.Items {
		if node.Labels[nodegroupLabelKey] == nodegroup {
			n := node
			_ = client.Resources().Delete(ctx, &n)
		}
	}
	return WaitForNodeCount(ctx, client, nodegroup, 0, 30*time.Second)
}

// TeardownPodAndNodeGroup deletes the test pods, waits for them to be deleted,
// waits for Cluster Autoscaler to scale down the node naturally (to keep CA state synchronized),
// and performs forced node cleanup if CA didn't scale down in time.
func TeardownPodAndNodeGroup(ctx context.Context, client klient.Client, pods []*corev1.Pod, nodegroup string) {
	for _, pod := range pods {
		if pod != nil {
			_ = client.Resources().Delete(ctx, pod)
			_ = WaitForPodDeleted(ctx, client, pod, defaultPodTimeout)
		}
	}
	// Allow CA to scale down the empty node naturally
	_ = WaitForNodeCount(ctx, client, nodegroup, 0, 35*time.Second)
	_ = CleanUpNodeGroup(ctx, client, nodegroup)
}

// WaitForTriggeredScaleUp waits for a TriggeredScaleUp event for the given pod.
func WaitForTriggeredScaleUp(ctx context.Context, client klient.Client, namespace, podName string, timeout time.Duration) error {
	return wait.For(func(ctx context.Context) (done bool, err error) {
		events := &corev1.EventList{}
		err = client.Resources(namespace).List(ctx, events)
		if err != nil {
			return false, err
		}
		for _, event := range events.Items {
			if event.InvolvedObject.Name == podName && event.Reason == "TriggeredScaleUp" {
				return true, nil
			}
		}
		return false, nil
	}, wait.WithTimeout(timeout), wait.WithContext(ctx))
}

// WaitForPodScheduled waits until the pod is assigned to a node.
func WaitForPodScheduled(ctx context.Context, client klient.Client, pod *corev1.Pod, timeout time.Duration) error {
	return wait.For(conditions.New(client.Resources()).ResourceMatch(pod, func(object k8s.Object) bool {
		p := object.(*corev1.Pod)
		return p.Spec.NodeName != ""
	}), wait.WithTimeout(timeout), wait.WithContext(ctx))
}

// WaitForPodDeleted waits until the pod is deleted.
func WaitForPodDeleted(ctx context.Context, client klient.Client, pod *corev1.Pod, timeout time.Duration) error {
	return wait.For(conditions.New(client.Resources()).ResourceDeleted(pod), wait.WithTimeout(timeout), wait.WithContext(ctx))
}

// WaitForNodeCount waits until the number of nodes with the specified nodegroup matches expected count.
func WaitForNodeCount(ctx context.Context, client klient.Client, nodegroup string, expectedCount int, timeout time.Duration) error {
	return wait.For(func(ctx context.Context) (done bool, err error) {
		nodeList := &corev1.NodeList{}
		err = client.Resources().List(ctx, nodeList)
		if err != nil {
			return false, err
		}
		count := 0
		for _, node := range nodeList.Items {
			if node.Labels[nodegroupLabelKey] == nodegroup {
				count++
			}
		}
		return count == expectedCount, nil
	}, wait.WithTimeout(timeout), wait.WithContext(ctx))
}

// WaitForNodesAtLeast waits until the number of nodes in a nodegroup is at least expectedCount.
func WaitForNodesAtLeast(ctx context.Context, client klient.Client, nodegroup string, expectedCount int, timeout time.Duration) error {
	return wait.For(func(ctx context.Context) (done bool, err error) {
		nodeList := &corev1.NodeList{}
		err = client.Resources().List(ctx, nodeList)
		if err != nil {
			return false, err
		}
		count := 0
		for _, node := range nodeList.Items {
			if node.Labels[nodegroupLabelKey] == nodegroup {
				count++
			}
		}
		return count >= expectedCount, nil
	}, wait.WithTimeout(timeout), wait.WithContext(ctx))
}

// WaitForNodesReady waits until at least expectedCount nodes in a nodegroup have Ready condition True.
func WaitForNodesReady(ctx context.Context, client klient.Client, nodegroup string, expectedCount int, timeout time.Duration) error {
	return wait.For(func(ctx context.Context) (done bool, err error) {
		nodeList := &corev1.NodeList{}
		err = client.Resources().List(ctx, nodeList)
		if err != nil {
			return false, err
		}
		readyCount := 0
		for _, node := range nodeList.Items {
			if node.Labels[nodegroupLabelKey] == nodegroup {
				for _, condition := range node.Status.Conditions {
					if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
						readyCount++
						break
					}
				}
			}
		}
		return readyCount >= expectedCount, nil
	}, wait.WithTimeout(timeout), wait.WithContext(ctx))
}

// CountNodegroupNodes returns the number of nodes currently matching the nodegroup.
func CountNodegroupNodes(ctx context.Context, client klient.Client, nodegroup string) (int, error) {
	nodeList := &corev1.NodeList{}
	err := client.Resources().List(ctx, nodeList)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, node := range nodeList.Items {
		if node.Labels[nodegroupLabelKey] == nodegroup {
			count++
		}
	}
	return count, nil
}
