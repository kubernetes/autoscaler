/*
Copyright 2018 The Kubernetes Authors.

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

package autoscaling

import (
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/kubernetes/test/e2e/framework"
)

const (
	recommenderComponent         = "recommender"
	updateComponent              = "updater"
	admissionControllerComponent = "admission-controller"
	fullVpaSuite                 = "full-vpa"
	actuationSuite               = "actuation"
	pollInterval                 = 10 * time.Second
	pollTimeout                  = 15 * time.Minute
	// VpaEvictionTimeout is a timeout for VPA to restart a pod if there are no
	// mechanisms blocking it (for example PDB).
	VpaEvictionTimeout = 3 * time.Minute

	defaultHamsterReplicas = int32(3)
)

var hamsterLabels = map[string]string{"app": "hamster"}

// SIGDescribe adds sig-autoscaling tag to test description.
func SIGDescribe(text string, body func()) bool {
	return ginkgo.Describe(fmt.Sprintf("[sig-autoscaling] %v", text), body)
}

// E2eDescribe describes a VPA e2e test.
func E2eDescribe(scenario, name string, body func()) bool {
	return SIGDescribe(fmt.Sprintf("[VPA] [%s] %s", scenario, name), body)
}

// RecommenderE2eDescribe describes a VPA recommender e2e test.
func RecommenderE2eDescribe(name string, body func()) bool {
	return E2eDescribe(recommenderComponent, name, body)
}

// UpdaterE2eDescribe describes a VPA updater e2e test.
func UpdaterE2eDescribe(name string, body func()) bool {
	return E2eDescribe(updateComponent, name, body)
}

// AdmissionControllerE2eDescribe describes a VPA admission controller e2e test.
func AdmissionControllerE2eDescribe(name string, body func()) bool {
	return E2eDescribe(admissionControllerComponent, name, body)
}

// FullVpaE2eDescribe describes a VPA full stack e2e test.
func FullVpaE2eDescribe(name string, body func()) bool {
	return E2eDescribe(fullVpaSuite, name, body)
}

// ActuationSuiteE2eDescribe describes a VPA actuation e2e test.
func ActuationSuiteE2eDescribe(name string, body func()) bool {
	return E2eDescribe(actuationSuite, name, body)
}

// SetupHamsterDeployment creates and installs a simple hamster deployment
// for e2e test purposes, then makes sure the deployment is running.
func SetupHamsterDeployment(f *framework.Framework, cpu, memory string, replicas int32) *appsv1.Deployment {
	cpuQuantity := ParseQuantityOrDie(cpu)
	memoryQuantity := ParseQuantityOrDie(memory)

	d := NewHamsterDeploymentWithResources(f, cpuQuantity, memoryQuantity)
	d.Spec.Replicas = &replicas
	d, err := f.ClientSet.AppsV1().Deployments(f.Namespace.Name).Create(d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = framework.WaitForDeploymentComplete(f.ClientSet, d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return d
}

// NewHamsterDeployment creates a simple hamster deployment for e2e test
// purposes.
func NewHamsterDeployment(f *framework.Framework) *appsv1.Deployment {
	d := framework.NewDeployment("hamster-deployment", defaultHamsterReplicas, hamsterLabels, "hamster", "k8s.gcr.io/ubuntu-slim:0.1", appsv1.RollingUpdateDeploymentStrategyType)
	d.ObjectMeta.Namespace = f.Namespace.Name
	d.Spec.Template.Spec.Containers[0].Command = []string{"/bin/sh"}
	d.Spec.Template.Spec.Containers[0].Args = []string{"-c", "/usr/bin/yes >/dev/null"}
	return d
}

// NewHamsterDeploymentWithResources creates a simple hamster deployment with specific
// resource requests for e2e test purposes.
func NewHamsterDeploymentWithResources(f *framework.Framework, cpuQuantity, memoryQuantity resource.Quantity) *appsv1.Deployment {
	d := NewHamsterDeployment(f)
	d.Spec.Template.Spec.Containers[0].Resources.Requests = apiv1.ResourceList{
		apiv1.ResourceCPU:    cpuQuantity,
		apiv1.ResourceMemory: memoryQuantity,
	}
	return d
}

// GetHamsterPods returns running hamster pods (matched by hamsterLabels)
func GetHamsterPods(f *framework.Framework) (*apiv1.PodList, error) {
	label := labels.SelectorFromSet(labels.Set(hamsterLabels))
	selector := fields.ParseSelectorOrDie("status.phase!=" + string(apiv1.PodSucceeded) +
		",status.phase!=" + string(apiv1.PodFailed))
	options := metav1.ListOptions{LabelSelector: label.String(), FieldSelector: selector.String()}
	return f.ClientSet.CoreV1().Pods(f.Namespace.Name).List(options)
}

// SetupVPA creates and installs a simple hamster VPA for e2e test purposes.
func SetupVPA(f *framework.Framework, cpu string, mode vpa_types.UpdateMode) {
	vpaCRD := NewVPA(f, "hamster-vpa", &metav1.LabelSelector{
		MatchLabels: hamsterLabels,
	})
	vpaCRD.Spec.UpdatePolicy.UpdateMode = &mode

	cpuQuantity := ParseQuantityOrDie(cpu)
	resourceList := apiv1.ResourceList{apiv1.ResourceCPU: cpuQuantity}

	vpaCRD.Status.Recommendation = &vpa_types.RecommendedPodResources{
		ContainerRecommendations: []vpa_types.RecommendedContainerResources{{
			ContainerName: "hamster",
			Target:        resourceList,
			LowerBound:    resourceList,
			UpperBound:    resourceList,
		}},
	}
	InstallVPA(f, vpaCRD)
}

// NewVPA creates a VPA object for e2e test purposes.
func NewVPA(f *framework.Framework, name string, selector *metav1.LabelSelector) *vpa_types.VerticalPodAutoscaler {
	updateMode := vpa_types.UpdateModeAuto
	vpa := vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: f.Namespace.Name,
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Selector: selector,
			UpdatePolicy: &vpa_types.PodUpdatePolicy{
				UpdateMode: &updateMode,
			},
			ResourcePolicy: &vpa_types.PodResourcePolicy{
				ContainerPolicies: []vpa_types.ContainerResourcePolicy{},
			},
		},
	}
	return &vpa
}

// InstallVPA installs a VPA object in the test cluster.
func InstallVPA(f *framework.Framework, vpa *vpa_types.VerticalPodAutoscaler) {
	ns := f.Namespace.Name
	config, err := framework.LoadConfig()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	vpaClientSet := vpa_clientset.NewForConfigOrDie(config)
	vpaClient := vpaClientSet.AutoscalingV1beta1()
	_, err = vpaClient.VerticalPodAutoscalers(ns).Create(vpa)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

// ParseQuantityOrDie parses quantity from string and dies with an error if
// unparsable.
func ParseQuantityOrDie(text string) resource.Quantity {
	quantity, err := resource.ParseQuantity(text)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return quantity
}

// PodSet is a simplified representation of PodList mapping names to UIDs.
type PodSet map[string]types.UID

// MakePodSet converts PodList to podset for easier comparison of pod collections.
func MakePodSet(pods *apiv1.PodList) PodSet {
	result := make(PodSet)
	if pods == nil {
		return result
	}
	for _, p := range pods.Items {
		result[p.Name] = p.UID
	}
	return result
}

// WaitForPodsRestarted waits until some pods from the list are restarted.
func WaitForPodsRestarted(f *framework.Framework, podList *apiv1.PodList) error {
	initialPodSet := MakePodSet(podList)

	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		currentPodList, err := GetHamsterPods(f)
		if err != nil {
			return false, err
		}
		currentPodSet := MakePodSet(currentPodList)
		return WerePodsSuccessfullyRestarted(currentPodSet, initialPodSet), nil
	})

	if err != nil {
		return fmt.Errorf("Waiting for set of pods changed: %v", err)
	}
	return nil
}

// WaitForPodsEvicted waits until some pods from the list are evicted.
func WaitForPodsEvicted(f *framework.Framework, podList *apiv1.PodList) error {
	initialPodSet := MakePodSet(podList)

	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		currentPodList, err := GetHamsterPods(f)
		if err != nil {
			return false, err
		}
		currentPodSet := MakePodSet(currentPodList)
		return GetEvictedPodsCount(currentPodSet, initialPodSet) > 0, nil
	})

	if err != nil {
		return fmt.Errorf("Waiting for set of pods changed: %v", err)
	}
	return nil
}

// WerePodsSuccessfullyRestarted returns true if some pods from initialPodSet have been
// successfully restarted comparing to currentPodSet (pods were evicted and
// are running).
func WerePodsSuccessfullyRestarted(currentPodSet PodSet, initialPodSet PodSet) bool {
	if len(currentPodSet) < len(initialPodSet) {
		// If we have less pods running than in the beginning, there is a restart
		// in progress - a pod was evicted but not yet recreated.
		framework.Logf("Restart in progress")
		return false
	}
	evictedCount := GetEvictedPodsCount(currentPodSet, initialPodSet)
	framework.Logf("%v of initial pods were already evicted", evictedCount)
	return evictedCount > 0
}

// GetEvictedPodsCount returns the count of pods from initialPodSet that have
// been evicted comparing to currentPodSet.
func GetEvictedPodsCount(currentPodSet PodSet, initialPodSet PodSet) int {
	diffs := 0
	for name, initialUID := range initialPodSet {
		currentUID, inCurrent := currentPodSet[name]
		if !inCurrent {
			diffs += 1
		} else if initialUID != currentUID {
			diffs += 1
		}
	}
	return diffs
}

// CheckNoPodsEvicted waits for long enough period for VPA to start evicting
// pods and checks that no pods were restarted.
func CheckNoPodsEvicted(f *framework.Framework, initialPodSet PodSet) {
	time.Sleep(VpaEvictionTimeout)
	currentPodList, err := GetHamsterPods(f)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	restarted := GetEvictedPodsCount(MakePodSet(currentPodList), initialPodSet)
	gomega.Expect(restarted).To(gomega.Equal(0))
}
