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

package cluster

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
)

// Cluster-level constants for the benchmark environment.
const (
	BenchmarkNamespace    = "benchmark"
	KwokNodeName          = "kwok-node"
	VPANamespace          = "kube-system"
	ReplicasPerReplicaSet = 2
	PauseImage            = "registry.k8s.io/pause:3.10.1"
)

// Profiles defines the number of ReplicaSets (each with ReplicasPerReplicaSet pods)
// for each benchmark size.
var Profiles = map[string]int{
	"small":   25,   // 25 VPAs, 25 ReplicaSets, 50 pods
	"medium":  100,  // 100 VPAs, 100 ReplicaSets, 200 pods
	"large":   250,  // 250 VPAs, 250 ReplicaSets, 500 pods
	"xlarge":  500,  // 500 VPAs, 500 ReplicaSets, 1000 pods
	"xxlarge": 1000, // 1000 VPAs, 1000 ReplicaSets, 2000 pods
}

var retryBackoff = wait.Backoff{
	Steps:    5,
	Duration: 100 * time.Millisecond,
	Factor:   2.0,
	Jitter:   0.1,
}

func withRetry(fn func() error) error {
	return retry.OnError(retryBackoff, func(err error) bool {
		return errors.IsConflict(err) || errors.IsServerTimeout(err) || errors.IsTooManyRequests(err) || errors.IsServiceUnavailable(err)
	}, fn)
}

// ScaleDeployment sets the replica count on the named deployment.
func ScaleDeployment(ctx context.Context, kubeClient kubernetes.Interface, namespace, name string, replicas int32) error {
	return withRetry(func() error {
		scale, err := kubeClient.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		scale.Spec.Replicas = replicas
		_, err = kubeClient.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
		return err
	})
}

// WaitForVPAPodReady polls until a running, ready pod exists for the given component label.
func WaitForVPAPodReady(ctx context.Context, kubeClient kubernetes.Interface, appLabel string) error {
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, 120*time.Second, true, func(ctx context.Context) (bool, error) {
		pods, _ := kubeClient.CoreV1().Pods(VPANamespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app.kubernetes.io/component=%s", appLabel),
		})
		for _, p := range pods.Items {
			if p.Status.Phase == corev1.PodRunning {
				allReady := true
				for _, c := range p.Status.ContainerStatuses {
					if !c.Ready {
						allReady = false
						break
					}
				}
				if allReady {
					return true, nil
				}
			}
		}
		return false, nil
	})
}

// DeleteAllVPACheckpoints removes all VPA checkpoint objects across all namespaces.
func DeleteAllVPACheckpoints(ctx context.Context, vpaClient vpa_clientset.Interface) {
	nsList, _ := vpaClient.AutoscalingV1().VerticalPodAutoscalerCheckpoints("").List(ctx, metav1.ListOptions{})
	for _, cp := range nsList.Items {
		vpaClient.AutoscalingV1().VerticalPodAutoscalerCheckpoints(cp.Namespace).Delete(ctx, cp.Name, metav1.DeleteOptions{})
	}
	klog.Infof("> Deleted %d VPA checkpoints", len(nsList.Items))
}

// MakeReplicaSet builds a ReplicaSet spec for the benchmark namespace with KWOK-compatible pods.
func MakeReplicaSet(name string) *appsv1.ReplicaSet {
	replicas := int32(ReplicasPerReplicaSet)
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: BenchmarkNamespace},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec: corev1.PodSpec{
					NodeName: KwokNodeName,
					Tolerations: []corev1.Toleration{{
						Key:      "kwok.x-k8s.io/node",
						Operator: corev1.TolerationOpEqual,
						Value:    "fake",
						Effect:   corev1.TaintEffectNoSchedule,
					}},
					Containers: []corev1.Container{{
						Name:  "app",
						Image: PauseImage,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("10m"),
								corev1.ResourceMemory: resource.MustParse("10Mi"),
							},
						},
					}},
				},
			},
		},
	}
}

// MakeVPA builds a VPA spec targeting the named ReplicaSet with Recreate update mode.
func MakeVPA(name string) *vpa_types.VerticalPodAutoscaler {
	mode := vpa_types.UpdateModeRecreate
	return &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: BenchmarkNamespace},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			TargetRef: &autoscalingv1.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "ReplicaSet",
				Name:       name,
			},
			UpdatePolicy: &vpa_types.PodUpdatePolicy{UpdateMode: &mode},
		},
	}
}

// CleanupBenchmarkResources deletes all VPAs, ReplicaSets, and pods in the benchmark namespace.
func CleanupBenchmarkResources(ctx context.Context, kubeClient kubernetes.Interface, vpaClient vpa_clientset.Interface) {
	vpaClient.AutoscalingV1().VerticalPodAutoscalers(BenchmarkNamespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
	kubeClient.AppsV1().ReplicaSets(BenchmarkNamespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{})
	wait.PollUntilContextTimeout(ctx, time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		pods, _ := kubeClient.CoreV1().Pods(BenchmarkNamespace).List(ctx, metav1.ListOptions{})
		return len(pods.Items) == 0, nil
	})
}

// CreateInParallel runs createFn concurrently for bench-0..bench-(count-1).
func CreateInParallel(ctx context.Context, count int, createFn func(ctx context.Context, name string)) {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(50)
	for i := range count {
		g.Go(func() error {
			name := fmt.Sprintf("bench-%d", i)
			createFn(gctx, name)
			return nil
		})
	}
	g.Wait()
}

// CreateReplicaSet creates a benchmark ReplicaSet, retrying on transient errors.
func CreateReplicaSet(ctx context.Context, kubeClient kubernetes.Interface, name string) {
	rs := MakeReplicaSet(name)
	err := withRetry(func() error {
		_, err := kubeClient.AppsV1().ReplicaSets(BenchmarkNamespace).Create(ctx, rs, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return err
	})
	if err != nil {
		klog.Warningf("Error creating ReplicaSet %s: %v", name, err)
	}
}

// CreateVPA creates a VPA targeting the named ReplicaSet, retrying on transient errors.
func CreateVPA(ctx context.Context, vpaClient vpa_clientset.Interface, name string) {
	vpa := MakeVPA(name)
	err := withRetry(func() error {
		_, err := vpaClient.AutoscalingV1().VerticalPodAutoscalers(BenchmarkNamespace).Create(ctx, vpa, metav1.CreateOptions{})
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return err
	})
	if err != nil {
		klog.Warningf("Error creating VPA %s: %v", name, err)
	}
}

// CreateNoiseReplicaSets creates unmanaged ReplicaSets to simulate background pod noise.
func CreateNoiseReplicaSets(ctx context.Context, kubeClient kubernetes.Interface, noiseCount int) {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(50)
	for i := range noiseCount {
		g.Go(func() error {
			name := fmt.Sprintf("noise-%d", i)
			rs := MakeReplicaSet(name)
			err := withRetry(func() error {
				_, err := kubeClient.AppsV1().ReplicaSets(BenchmarkNamespace).Create(gctx, rs, metav1.CreateOptions{})
				if errors.IsAlreadyExists(err) {
					return nil
				}
				return err
			})
			if err != nil {
				klog.Warningf("Error creating noise ReplicaSet %s: %v", name, err)
			}
			return nil
		})
	}
	g.Wait()
}
