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

package scaleup_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure/test/pkg/environment"
)

var (
	env           *environment.Environment
	resourceGroup string

	// Helm deployment flags (CI path). When image repo+tag are set, CAS is deployed via Helm.
	// When empty, CAS is assumed to already be running (local dev path via skaffold).
	clusterName           string
	clientID              string
	casNamespace          string
	casServiceAccountName string
	casImageRepository    string
	casImageTag           string
)

func helmValuesForScaleUp() map[string]interface{} {
	return map[string]interface{}{
		"extraArgs": map[string]interface{}{
			"scale-down-delay-after-add": "30m",
			"scale-down-unneeded-time":   "30m",
		},
	}
}

func helmValuesForFastScaleDown() map[string]interface{} {
	return map[string]interface{}{
		"extraArgs": map[string]interface{}{
			"scale-down-delay-after-add":       "10s",
			"scale-down-unneeded-time":         "10s",
			"scale-down-candidates-pool-ratio": "1.0",
			"unremovable-node-recheck-timeout": "10s",
			"skip-nodes-with-system-pods":      "false",
			"skip-nodes-with-local-storage":    "false",
		},
	}
}

func init() {
	flag.StringVar(&resourceGroup, "resource-group", "", "resource group containing cluster-autoscaler-managed resources (the MC_ node resource group)")
	flag.StringVar(&clusterName, "cluster-name", "", "Cluster API Cluster name (CI only)")
	flag.StringVar(&clientID, "client-id", "", "Azure client ID for CAS workload identity (CI only)")
	flag.StringVar(&casNamespace, "cas-namespace", "default", "Namespace for CAS Helm release (CI only)")
	flag.StringVar(&casServiceAccountName, "cas-serviceaccount-name", "cluster-autoscaler", "CAS ServiceAccount name (CI only)")
	flag.StringVar(&casImageRepository, "cas-image-repository", "", "CAS image repository (CI only, triggers Helm deploy)")
	flag.StringVar(&casImageTag, "cas-image-tag", "", "CAS image tag (CI only, triggers Helm deploy)")
}

func TestScaleUp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scale Up Suite")
}

var _ = BeforeSuite(func() {
	var helm *environment.HelmConfig
	if casImageRepository != "" && casImageTag != "" {
		helm = &environment.HelmConfig{
			// From suites/scaleup/, 5 levels up reaches cluster-autoscaler/ where charts/ lives.
			ChartPath:             "../../../../../charts/cluster-autoscaler",
			ClusterName:           clusterName,
			ClientID:              clientID,
			CASNamespace:          casNamespace,
			CASServiceAccountName: casServiceAccountName,
			CASImageRepository:    casImageRepository,
			CASImageTag:           casImageTag,
		}
	}
	env = environment.NewEnvironment(resourceGroup, helm)
	env.EnsureHelmRelease(helmValuesForScaleUp())
})

var _ = Describe("Azure Provider", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		Eventually(env.AllVMSSStable, "10m", "30s").Should(Succeed())
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "azure-e2e-"},
		}
		Expect(env.K8s.Create(env.Ctx, namespace)).To(Succeed())
	})

	AfterEach(func() {
		Expect(env.K8s.Delete(env.Ctx, namespace)).To(Succeed())
		Eventually(func() bool {
			err := env.K8s.Get(env.Ctx, client.ObjectKeyFromObject(namespace), &corev1.Namespace{})
			return apierrors.IsNotFound(err)
		}, "1m", "5s").Should(BeTrue(), "Namespace "+namespace.Name+" still exists")
	})

	It("scales up AKS node pools when pending Pods exist", func() {
		nodes := &corev1.NodeList{}
		Expect(env.K8s.List(env.Ctx, nodes)).To(Succeed())
		nodeCountBefore := len(nodes.Items)

		By("Creating 100 Pods")
		deploy := &appsv1.Deployment{
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
				Replicas: ptr.To[int32](100),
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
								Image: "registry.k8s.io/hpa-example",
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
		Expect(env.K8s.Create(env.Ctx, deploy)).To(Succeed())

		By("Waiting for more Ready Nodes to exist")
		Eventually(func() (int, error) {
			readyCount := 0
			nodes := &corev1.NodeList{}
			if err := env.K8s.List(env.Ctx, nodes); err != nil {
				return 0, err
			}
			for _, node := range nodes.Items {
				for _, cond := range node.Status.Conditions {
					if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
						readyCount++
						break
					}
				}
			}
			return readyCount, nil
		}, "10m", "10s").Should(BeNumerically(">", nodeCountBefore))

		Eventually(env.AllVMSSStable, "20m", "30s").Should(Succeed())

		By("Reconfiguring Cluster Autoscaler for fast scale down")
		env.EnsureHelmRelease(helmValuesForFastScaleDown())

		By("Deleting 100 Pods")
		Expect(env.K8s.Delete(env.Ctx, deploy)).To(Succeed())

		By("Waiting for the original number of Nodes to be Ready")
		Eventually(func(g Gomega) {
			nodes := &corev1.NodeList{}
			g.Expect(env.K8s.List(env.Ctx, nodes)).To(Succeed())
			g.Expect(nodes.Items).To(SatisfyAll(
				HaveLen(nodeCountBefore),
				ContainElements(Satisfy(func(node corev1.Node) bool {
					for _, cond := range node.Status.Conditions {
						if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
							return true
						}
					}
					return false
				})),
			))
		}, "20m", "10s").Should(Succeed())
	})
})
