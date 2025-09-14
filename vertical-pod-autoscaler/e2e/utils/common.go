/*
Copyright 2025 The Kubernetes Authors.

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

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	framework_deployment "k8s.io/kubernetes/test/e2e/framework/deployment"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
)

const (
	recommenderComponent = "recommender"

	// RecommenderDeploymentName is VPA recommender deployment name
	RecommenderDeploymentName = "vpa-recommender"
	// RecommenderNamespace is namespace to deploy VPA recommender
	RecommenderNamespace = "kube-system"
	// PollInterval is interval for polling
	PollInterval = 10 * time.Second
	// PollTimeout is timeout for polling
	PollTimeout = 15 * time.Minute

	// DefaultHamsterReplicas is replicas of hamster deployment
	DefaultHamsterReplicas = int32(3)
	// DefaultHamsterBackoffLimit is BackoffLimit of hamster app
	DefaultHamsterBackoffLimit = int32(10)
)

// HamsterTargetRef is CrossVersionObjectReference of hamster app
var HamsterTargetRef = &autoscaling.CrossVersionObjectReference{
	APIVersion: "apps/v1",
	Kind:       "Deployment",
	Name:       "hamster-deployment",
}

// RecommenderLabels are labels of VPA recommender
var RecommenderLabels = map[string]string{"app": "vpa-recommender"}

// HamsterLabels are labels of hamster app
var HamsterLabels = map[string]string{"app": "hamster"}

// SIGDescribe adds sig-autoscaling tag to test description.
// Takes args that are passed to ginkgo.Describe.
func SIGDescribe(scenario, name string, args ...interface{}) bool {
	full := fmt.Sprintf("[sig-autoscaling] [VPA] [%s] [v1] %s", scenario, name)
	return ginkgo.Describe(full, args...)
}

// RecommenderE2eDescribe describes a VPA recommender e2e test.
func RecommenderE2eDescribe(name string, args ...interface{}) bool {
	return SIGDescribe(recommenderComponent, name, args...)
}

// GetHamsterContainerNameByIndex returns name of i-th hamster container.
func GetHamsterContainerNameByIndex(i int) string {
	switch {
	case i < 0:
		panic("negative index")
	case i == 0:
		return "hamster"
	default:
		return fmt.Sprintf("hamster%d", i+1)
	}
}

// GetVpaClientSet return a VpaClientSet
func GetVpaClientSet(f *framework.Framework) vpa_clientset.Interface {
	config, err := framework.LoadConfig()
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error loading framework")
	return vpa_clientset.NewForConfigOrDie(config)
}

// InstallVPA installs a VPA object in the test cluster.
func InstallVPA(f *framework.Framework, vpa *vpa_types.VerticalPodAutoscaler) {
	vpaClientSet := GetVpaClientSet(f)
	_, err := vpaClientSet.AutoscalingV1().VerticalPodAutoscalers(f.Namespace.Name).Create(context.TODO(), vpa, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error creating VPA")
	// apiserver ignore status in vpa create, so need to update status
	if !isStatusEmpty(&vpa.Status) {
		if vpa.Status.Recommendation != nil {
			PatchVpaRecommendation(f, vpa, vpa.Status.Recommendation)
		}
	}
}

func isStatusEmpty(status *vpa_types.VerticalPodAutoscalerStatus) bool {
	if status == nil {
		return true
	}

	if len(status.Conditions) == 0 && status.Recommendation == nil {
		return true
	}
	return false
}

// PatchRecord used for patch action
type PatchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

// PatchVpaRecommendation installs a new recommendation for VPA object.
func PatchVpaRecommendation(f *framework.Framework, vpa *vpa_types.VerticalPodAutoscaler,
	recommendation *vpa_types.RecommendedPodResources) {
	newStatus := vpa.Status.DeepCopy()
	newStatus.Recommendation = recommendation
	bytes, err := json.Marshal([]PatchRecord{{
		Op:    "replace",
		Path:  "/status",
		Value: *newStatus,
	}})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	_, err = GetVpaClientSet(f).AutoscalingV1().VerticalPodAutoscalers(f.Namespace.Name).Patch(context.TODO(), vpa.Name, types.JSONPatchType, bytes, metav1.PatchOptions{}, "status")
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to patch VPA.")
}

// NewVPADeployment creates a VPA deployment with n containers
// for e2e test purposes.
func NewVPADeployment(f *framework.Framework, flags []string) *appsv1.Deployment {
	d := framework_deployment.NewDeployment(
		RecommenderDeploymentName,        /*deploymentName*/
		1,                                /*replicas*/
		RecommenderLabels,                /*podLabels*/
		"recommender",                    /*imageName*/
		"localhost:5001/vpa-recommender", /*image*/
		appsv1.RollingUpdateDeploymentStrategyType, /*strategyType*/
	)
	d.ObjectMeta.Namespace = f.Namespace.Name
	d.Spec.Template.Spec.Containers[0].ImagePullPolicy = apiv1.PullNever // Image must be loaded first
	d.Spec.Template.Spec.ServiceAccountName = "vpa-recommender"
	d.Spec.Template.Spec.Containers[0].Command = []string{"/recommender"}
	d.Spec.Template.Spec.Containers[0].Args = flags

	runAsNonRoot := true
	var runAsUser int64 = 65534 // nobody
	d.Spec.Template.Spec.SecurityContext = &apiv1.PodSecurityContext{
		RunAsNonRoot: &runAsNonRoot,
		RunAsUser:    &runAsUser,
	}

	// Same as deploy/recommender-deployment.yaml
	d.Spec.Template.Spec.Containers[0].Resources = apiv1.ResourceRequirements{
		Limits: apiv1.ResourceList{
			apiv1.ResourceCPU:    resource.MustParse("200m"),
			apiv1.ResourceMemory: resource.MustParse("1000Mi"),
		},
		Requests: apiv1.ResourceList{
			apiv1.ResourceCPU:    resource.MustParse("50m"),
			apiv1.ResourceMemory: resource.MustParse("500Mi"),
		},
	}

	d.Spec.Template.Spec.Containers[0].Ports = []apiv1.ContainerPort{{
		Name:          "prometheus",
		ContainerPort: 8942,
	}}

	d.Spec.Template.Spec.Containers[0].LivenessProbe = &apiv1.Probe{
		ProbeHandler: apiv1.ProbeHandler{
			HTTPGet: &apiv1.HTTPGetAction{
				Path:   "/health-check",
				Port:   intstr.FromString("prometheus"),
				Scheme: apiv1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		FailureThreshold:    3,
	}
	d.Spec.Template.Spec.Containers[0].ReadinessProbe = &apiv1.Probe{
		ProbeHandler: apiv1.ProbeHandler{
			HTTPGet: &apiv1.HTTPGetAction{
				Path:   "/health-check",
				Port:   intstr.FromString("prometheus"),
				Scheme: apiv1.URISchemeHTTP,
			},
		},
		PeriodSeconds:    10,
		FailureThreshold: 3,
	}

	return d
}

// NewNHamstersDeployment creates a simple hamster deployment with n containers
// for e2e test purposes.
func NewNHamstersDeployment(f *framework.Framework, n int) *appsv1.Deployment {
	if n < 1 {
		panic("container count should be greater than 0")
	}
	d := framework_deployment.NewDeployment(
		"hamster-deployment",                       /*deploymentName*/
		DefaultHamsterReplicas,                     /*replicas*/
		HamsterLabels,                              /*podLabels*/
		GetHamsterContainerNameByIndex(0),          /*imageName*/
		"registry.k8s.io/ubuntu-slim:0.14",         /*image*/
		appsv1.RollingUpdateDeploymentStrategyType, /*strategyType*/
	)
	d.ObjectMeta.Namespace = f.Namespace.Name
	d.Spec.Template.Spec.Containers[0].Command = []string{"/bin/sh"}
	d.Spec.Template.Spec.Containers[0].Args = []string{"-c", "/usr/bin/yes >/dev/null"}
	for i := 1; i < n; i++ {
		d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, d.Spec.Template.Spec.Containers[0])
		d.Spec.Template.Spec.Containers[i].Name = GetHamsterContainerNameByIndex(i)
	}
	return d
}

// StartDeploymentPods start and wait for a deployment to complete
func StartDeploymentPods(f *framework.Framework, deployment *appsv1.Deployment) *apiv1.PodList {
	// Apiserver watch can lag depending on cached object count and apiserver resource usage.
	// We assume that watch can lag up to 5 seconds.
	const apiserverWatchLag = 5 * time.Second
	// In admission controller e2e tests a recommendation is created before deployment.
	// Creating deployment with size greater than 0 would create a race between information
	// about pods and information about deployment getting to the admission controller.
	// Any pods that get processed by AC before it receives information about the deployment
	// don't receive recommendation.
	// To avoid this create deployment with size 0, then scale it up to the desired size.
	desiredPodCount := *deployment.Spec.Replicas
	zero := int32(0)
	deployment.Spec.Replicas = &zero
	c, ns := f.ClientSet, f.Namespace.Name
	deployment, err := c.AppsV1().Deployments(ns).Create(context.TODO(), deployment, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "when creating deployment with size 0")

	err = framework_deployment.WaitForDeploymentComplete(c, deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "when waiting for empty deployment to create")
	// If admission controller receives pod before controller it will not apply recommendation and test will fail.
	// Wait after creating deployment to ensure VPA knows about it, then scale up.
	// Normally watch lag is not a problem in terms of correctness:
	// - Mode "Auto": created pod without assigned resources will be handled by the eviction loop.
	// - Mode "Initial": calculating recommendations takes more than potential ectd lag.
	// - Mode "Off": pods are not handled by the admission controller.
	// In e2e admission controller tests we want to focus on scenarios without considering watch lag.
	// TODO(#2631): Remove sleep when issue is fixed.
	time.Sleep(apiserverWatchLag)

	scale := autoscaling.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.ObjectMeta.Name,
			Namespace: deployment.ObjectMeta.Namespace,
		},
		Spec: autoscaling.ScaleSpec{
			Replicas: desiredPodCount,
		},
	}
	afterScale, err := c.AppsV1().Deployments(ns).UpdateScale(context.TODO(), deployment.Name, &scale, metav1.UpdateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(afterScale.Spec.Replicas).To(gomega.Equal(desiredPodCount), fmt.Sprintf("expected %d replicas after scaling", desiredPodCount))

	// After scaling deployment we need to retrieve current version with updated replicas count.
	deployment, err = c.AppsV1().Deployments(ns).Get(context.TODO(), deployment.Name, metav1.GetOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "when getting scaled deployment")
	err = framework_deployment.WaitForDeploymentComplete(c, deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "when waiting for deployment to resize")

	podList, err := framework_deployment.GetPodsForDeployment(context.TODO(), c, deployment)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "when listing pods after deployment resize")
	return podList
}

// WaitForRecommendationPresent pools VPA object until recommendations are not empty. Returns
// polled vpa object. On timeout returns error.
func WaitForRecommendationPresent(c vpa_clientset.Interface, vpa *vpa_types.VerticalPodAutoscaler) (*vpa_types.VerticalPodAutoscaler, error) {
	return WaitForVPAMatch(c, vpa, func(vpa *vpa_types.VerticalPodAutoscaler) bool {
		return vpa.Status.Recommendation != nil && len(vpa.Status.Recommendation.ContainerRecommendations) != 0
	})
}

// WaitForVPAMatch pools VPA object until match function returns true. Returns
// polled vpa object. On timeout returns error.
func WaitForVPAMatch(c vpa_clientset.Interface, vpa *vpa_types.VerticalPodAutoscaler, match func(vpa *vpa_types.VerticalPodAutoscaler) bool) (*vpa_types.VerticalPodAutoscaler, error) {
	var polledVpa *vpa_types.VerticalPodAutoscaler
	err := wait.PollUntilContextTimeout(context.Background(), PollInterval, PollTimeout, true, func(ctx context.Context) (done bool, err error) {
		polledVpa, err = c.AutoscalingV1().VerticalPodAutoscalers(vpa.Namespace).Get(context.TODO(), vpa.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if match(polledVpa) {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error waiting for recommendation present in %v: %v", vpa.Name, err)
	}
	return polledVpa, nil
}
