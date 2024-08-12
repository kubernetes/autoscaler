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
	"strings"

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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
)

// Context: Why this suite?
// Without an implementation of HasInstance, nodes with ToBeDeletedByClusterAutoscaler that still
// have a VM will be counted as deleted rather than unready due to cluster state checking if a
// node is being deleted based on the existence of that taint.
// Inside the scale-up loop, we call GetUpcomingNodes, which returns how many nodes will be added to each of the node groups.
// https://github.com/kubernetes/autoscaler/blob/cluster-autoscaler-release-1.30/cluster-autoscaler/clusterstate/clusterstate.go#L987
//
//	newNodes := ar.CurrentTarget - (len(readiness.Ready) + len(readiness.Unready) + len(readiness.LongUnregistered))
//
// We will falsely report the count of newNodes here, which leads to not creating new nodes in the scale-up loop.
// We want to validate that for the Azure provider, HasInstance solves the case where we have a VM that has not been deleted yet,
// counting toward our node count.
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

	It("should return if a node has ToBeDeletedByClusterAutoscaler and the vm is still there", func() {
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
		By("schedule workload to go on the node")
		// Create a deployment that doesn't tolerate the taints of the other nodes
		// This will ensure cluster autoscaler will not scale down the one node we create an agentpool put above until we remove this deployment
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "php-apache",
				Namespace: namespace.Name,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"run": "php-apache",
					},
				},
				Replicas: ptr.To[int32](1), // one pod is all it takes :)
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"run": "php-apache",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "php-apache",
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

		By("verifying the new node's ProviderID matches the expected resource ID and the resource exists before placing a delete lock on it")
		providerIDParts := strings.Split(strings.TrimPrefix(newNode.Spec.ProviderID, "azure:///"), "/")
		if len(providerIDParts) < 9 {
			Fail("ProviderID format is incorrect")
		}
		vmssName := providerIDParts[7]
		instanceID := providerIDParts[9]
		_, err := vmssVMsClient.Get(ctx, nodeResourceGroup, vmssName, instanceID, nil)
		Expect(err).To(BeNil())
		By("Getting Cluster Autoscaler status configmap for post scale down attempt comparisons")
		Expect(err).NotTo(HaveOccurred())

		casStatusBeforeScaleDown, err := GetStructuredStatus(ctx, k8s)
		Expect(err).ShouldNot(HaveOccurred())

		By("scaling down the workload")
		Expect(k8s.Delete(ctx, deployment)).To(Succeed())

		// The AzureCache keeps the vm in the cache for at most 1 minute after its deletion.
		// We can set this to one minute for that reason
		By("verifying the node is marked with ToBeDeletedByClusterAutoscaler taint and the vm exists still")
		Eventually(func() bool {
			Expect(k8s.Get(ctx, client.ObjectKey{Name: newNode.Name}, newNode)).To(Succeed())
			for _, taint := range newNode.Spec.Taints {
				if taint.Key == "ToBeDeletedByClusterAutoscaler" {
					return true
				}
			}
			return false
		}, "1m", "1s").Should(BeTrue(), "Node should have ToBeDeletedByClusterAutoscaler taint")
		_, err = vmssVMsClient.Get(ctx, nodeResourceGroup, vmssName, instanceID, nil)
		Expect(err).To(BeNil())
		By("Expecting cluster autoscaler status to not report this node as BeingDeleted")
		newStatus, err := GetStructuredStatus(ctx, k8s)
		Expect(err).ToNot(HaveOccurred())
		// We should not be reporting this node as being deleted
		// even though it has the ToBeDeleted CAS Taint with HasInstance implemented
		Expect(newStatus.ClusterWide.Health.NodeCounts.Registered.BeingDeleted).To(
			Equal(casStatusBeforeScaleDown.ClusterWide.Health.NodeCounts.Registered.BeingDeleted),
		)
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
	// Taint all the systempool nodes
	for _, node := range nodeList.Items {
		if node.Labels["kubernetes.azure.com/mode"] == "system" {
			updateNode := node.DeepCopy()
			updateNode.Spec.Taints = append(updateNode.Spec.Taints, taint)
			err := k8s.Update(ctx, updateNode)
			Expect(err).To(BeNil())
		}
	}
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

func CreateNodepool(ctx context.Context, npClient *armcontainerservice.AgentPoolsClient, rg, clusterName, nodepoolName string, agentpool armcontainerservice.AgentPool) (*armcontainerservice.AgentPool, error) {
	poller, err := npClient.BeginCreateOrUpdate(ctx, rg, clusterName, nodepoolName, agentpool, nil)
	if err != nil {
		return nil, err
	}
	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &res.AgentPool, nil
}

func DeleteNodepool(ctx context.Context, npClient *armcontainerservice.AgentPoolsClient, rg, clusterName, nodepoolName string) error {
	poller, err := npClient.BeginDelete(ctx, rg, clusterName, nodepoolName, nil)
	Expect(err).ToNot(HaveOccurred())
	_, err = poller.PollUntilDone(ctx, nil)
	return err
}

func ExpectDeploymentToBeReady(ctx context.Context, k8sClient client.Client, deployment *appsv1.Deployment, expectedReplicas int32, timeout string, interval string) {
	Eventually(func() int32 {
		fetchedDeployment := &appsv1.Deployment{}
		err := k8sClient.Get(ctx, client.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}, fetchedDeployment)
		Expect(err).ToNot(HaveOccurred())
		return fetchedDeployment.Status.ReadyReplicas
	}, timeout, interval).Should(Equal(expectedReplicas))
}
