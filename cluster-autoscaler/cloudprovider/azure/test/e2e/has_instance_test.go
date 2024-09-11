//go:build e2e

/*
Copyright 2024 The Kubernetes Authors.

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

package e2e_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"k8s.io/apimachinery/pkg/types"
	api "k8s.io/autoscaler/cluster-autoscaler/clusterstate/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("cloudprovider.HasInstance(v1.Node)", func() {
	var (
		namespace *corev1.Namespace
	)

	BeforeEach(func() {
		Eventually(allVMSSStable, "10m", "30s").Should(Succeed())

		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "azure-e2e-",
			},
		}
		Expect(k8s.Create(ctx, namespace)).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8s.Delete(ctx, namespace)).To(Succeed())
		Eventually(func() bool {
			err := k8s.Get(ctx, client.ObjectKeyFromObject(namespace), &corev1.Namespace{})
			return apierrors.IsNotFound(err)
		}, "1m", "5s").Should(BeTrue(), "Namespace "+namespace.Name+" still exists")
	})

	It("should validate cluster state does not report a node as BeingDeleted if a node has ToBeDeletedByClusterAutoscaler taint and the vm is still there", func() {
		ensureHelmValues(map[string]interface{}{
			"extraArgs": map[string]interface{}{
				"scale-down-delay-after-add":       "10s",
				"scale-down-unneeded-time":         "10s",
				"scale-down-unready-time":          "10s",
				"scale-down-candidates-pool-ratio": "1.0",
				"unremovable-node-recheck-timeout": "10s",
				"skip-nodes-with-system-pods":      "false",
				"skip-nodes-with-local-storage":    "false",
				"max-graceful-termination-sec":     "300",
			},
		})

		By("tainting all existing nodes to ensure workload gets scheduled on a new node")
		ExpectTaintedSystempool(ctx, k8s)
		By("schedule a workload to have one user pool node stick around")
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "inflate-pause-image",
				Namespace: namespace.Name,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"run": "inflate-pause-image",
					},
				},
				Replicas: ptr.To[int32](1), // one pod is all it takes :)
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"run": "inflate-pause-image",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "inflate-pause-image",
								Image: "mcr.microsoft.com/oss/kubernetes/pause:3.6",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										corev1.ResourceCPU: resource.MustParse("500m"),
									},
									Requests: corev1.ResourceList{
										corev1.ResourceCPU: resource.MustParse("200m"),
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(k8s.Create(ctx, deployment)).To(Succeed())
		ExpectDeploymentToBeReady(ctx, k8s, deployment, 1, "10m", "10s")
		var newNode *corev1.Node
		Eventually(func() bool {
			podList := &corev1.PodList{}
			Expect(k8s.List(ctx, podList, client.InNamespace(namespace.Name))).To(Succeed())

			for _, pod := range podList.Items {
				if pod.Spec.NodeName != "" {
					node := &corev1.Node{}
					Expect(k8s.Get(ctx, client.ObjectKey{Name: pod.Spec.NodeName}, node)).To(Succeed())
					newNode = node
					return true
				}
			}
			return false
		}, "5m", "10s").Should(BeTrue(), "Workload should be scheduled on a new node")

		By("getting Cluster Autoscaler status configmap for post scale down attempt comparisons")
		ExpectStatusConfigmapExists(ctx, k8s, 5*time.Minute, 5*time.Second)
		casStatusBeforeScaleDown, err := GetStructuredStatus(ctx, k8s)
		Expect(err).ShouldNot(HaveOccurred())

		By("ensuring we don't have scale down candidates before we delete the deployment")

		By("scaling down the workload")
		Expect(k8s.Delete(ctx, deployment)).To(Succeed())

		By("verifying the node is marked with ToBeDeletedByClusterAutoscaler taint and the vm exists still")
		ExpectNodeEventuallyHasTaint(ctx, newNode, "ToBeDeletedByClusterAutoscaler", "1m", "1s")

		By("getting the latest status")
		latestStatus, err := GetStructuredStatus(ctx, k8s)
		Expect(err).ToNot(HaveOccurred())

		By("expecting 1 scale down candidate to be present for us to be able to trust the configmap state for assertions below")
		Expect(latestStatus.ClusterWide.ScaleDown.Candidates).To(Equal(1))

		By("expecting cluster autoscaler status to not report this node as BeingDeleted")
		Expect(latestStatus.ClusterWide.Health.NodeCounts.Registered.BeingDeleted).To(
			Equal(casStatusBeforeScaleDown.ClusterWide.Health.NodeCounts.Registered.BeingDeleted),
		)

		By("checking if node counts for Ready or Unready nodes have decreased by 1")
		Expect(latestStatus.ClusterWide.Health.NodeCounts.Registered.Ready).To(Equal(casStatusBeforeScaleDown.ClusterWide.Health.NodeCounts.Registered.Ready))
		Expect(latestStatus.ClusterWide.Health.NodeCounts.Registered.Unready.Total).To(Equal(casStatusBeforeScaleDown.ClusterWide.Health.NodeCounts.Registered.Unready.Total))
	})
})

func ExpectTaintedSystempool(ctx context.Context, k8sClient client.Client) {
	taint := corev1.Taint{
		Key:    "CriticalAddonsOnly",
		Value:  "true",
		Effect: corev1.TaintEffectNoSchedule,
	}
	nodeList := &corev1.NodeList{}
	Expect(k8s.List(ctx, nodeList)).To(Succeed())
	for _, node := range nodeList.Items {
		if node.Labels["kubernetes.azure.com/mode"] == "system" {
			updateNode := node.DeepCopy()
			updateNode.Spec.Taints = append(updateNode.Spec.Taints, taint)
			err := k8s.Update(ctx, updateNode)
			Expect(err).To(BeNil())
		}
	}
}

func ExpectNodeEventuallyHasTaint(ctx context.Context, node *corev1.Node, key, timeout, interval string) {
	Eventually(func() bool {
		Expect(k8s.Get(ctx, client.ObjectKey{Name: node.Name}, node)).To(Succeed())
		for _, taint := range node.Spec.Taints {
			if taint.Key == key {
				return true
			}
		}
		return false
	}, timeout, interval).Should(BeTrue(), fmt.Sprintf("Node should have %s taint", key))
}

func ExpectStatusConfigmapExists(ctx context.Context, k8sClient client.Client, timeout, interval time.Duration) {
	Eventually(func() (*corev1.ConfigMap, error) {
		return GetStatusConfigmap(ctx, k8sClient)
	}, timeout, interval).ShouldNot(BeNil(), "ConfigMap 'cluster-autoscaler-status' should exist in the 'default' namespace")
}

func GetStatusConfigmap(ctx context.Context, k8sClient client.Client) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := k8sClient.Get(ctx, types.NamespacedName{
		Namespace: "default",
		Name:      "cluster-autoscaler-status",
	}, configMap)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

// This will work for 1.31+, but a separate effort will have to happen to have coverage in older versions
func GetStructuredStatus(ctx context.Context, k8sClient client.Client) (*api.ClusterAutoscalerStatus, error) {
	configMap, err := GetStatusConfigmap(ctx, k8sClient)
	if err != nil {
		return nil, err
	}
	data := configMap.Data
	Status := &api.ClusterAutoscalerStatus{}
	if err := yaml.Unmarshal([]byte(data["status"]), &Status); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal cluster autoscaler status. data status: \n%s", data["status"])
	}
	return Status, nil
}

func ExpectDeploymentToBeReady(ctx context.Context, k8sClient client.Client, deployment *appsv1.Deployment, expectedReplicas int32, timeout string, interval string) {
	Eventually(func() int32 {
		fetchedDeployment := &appsv1.Deployment{}
		err := k8sClient.Get(ctx, client.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}, fetchedDeployment)
		Expect(err).ToNot(HaveOccurred())
		return fetchedDeployment.Status.ReadyReplicas
	}, timeout, interval).Should(Equal(expectedReplicas))
}
