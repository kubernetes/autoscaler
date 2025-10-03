/*
Copyright 2019 The Kubernetes Authors.

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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/vertical-pod-autoscaler/e2e/utils"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	framework_deployment "k8s.io/kubernetes/test/e2e/framework/deployment"
)

const (
	updateComponent              = "updater"
	admissionControllerComponent = "admission-controller"
	fullVpaSuite                 = "full-vpa"
	actuationSuite               = "actuation"
	cronJobsWaitTimeout          = 15 * time.Minute
	// VpaEvictionTimeout is a timeout for VPA to restart a pod if there are no
	// mechanisms blocking it (for example PDB).
	VpaEvictionTimeout = 3 * time.Minute
	// VpaInPlaceTimeout is a timeout for the VPA to finish in-place resizing a
	// pod, if there are no mechanisms blocking it.
	VpaInPlaceTimeout = 2 * time.Minute

	// VpaNamespace is the default namespace that holds the all the VPA components.
	VpaNamespace = "kube-system"
)

// UpdaterE2eDescribe describes a VPA updater e2e test.
func UpdaterE2eDescribe(name string, args ...interface{}) bool {
	return utils.SIGDescribe(updateComponent, name, args...)
}

// AdmissionControllerE2eDescribe describes a VPA admission controller e2e test.
func AdmissionControllerE2eDescribe(name string, args ...interface{}) bool {
	return utils.SIGDescribe(admissionControllerComponent, name, args...)
}

// FullVpaE2eDescribe describes a VPA full stack e2e test.
func FullVpaE2eDescribe(name string, args ...interface{}) bool {
	return utils.SIGDescribe(fullVpaSuite, name, args...)
}

// ActuationSuiteE2eDescribe describes a VPA actuation e2e test.
func ActuationSuiteE2eDescribe(name string, args ...interface{}) bool {
	return utils.SIGDescribe(actuationSuite, name, args...)
}

// SetupHamsterDeployment creates and installs a simple hamster deployment
// for e2e test purposes, then makes sure the deployment is running.
func SetupHamsterDeployment(f *framework.Framework, cpu, memory string, replicas int32) *appsv1.Deployment {
	cpuQuantity := ParseQuantityOrDie(cpu)
	memoryQuantity := ParseQuantityOrDie(memory)

	d := NewHamsterDeploymentWithResources(f, cpuQuantity, memoryQuantity)
	d.Spec.Replicas = &replicas
	d, err := f.ClientSet.AppsV1().Deployments(f.Namespace.Name).Create(context.TODO(), d, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error when starting deployment creation")
	err = framework_deployment.WaitForDeploymentComplete(f.ClientSet, d)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error waiting for deployment creation to finish")
	return d
}

// NewHamsterDeployment creates a simple hamster deployment for e2e test purposes.
func NewHamsterDeployment(f *framework.Framework) *appsv1.Deployment {
	return utils.NewNHamstersDeployment(f, 1)
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

// NewHamsterDeploymentWithGuaranteedResources creates a simple hamster deployment with specific
// resource requests for e2e test purposes. Since the container in the pod specifies resource limits
// but not resource requests K8s will set requests equal to limits and the pod will have guaranteed
// QoS class.
func NewHamsterDeploymentWithGuaranteedResources(f *framework.Framework, cpuQuantity, memoryQuantity resource.Quantity) *appsv1.Deployment {
	d := NewHamsterDeployment(f)
	d.Spec.Template.Spec.Containers[0].Resources.Limits = apiv1.ResourceList{
		apiv1.ResourceCPU:    cpuQuantity,
		apiv1.ResourceMemory: memoryQuantity,
	}
	return d
}

// NewHamsterDeploymentWithResourcesAndLimits creates a simple hamster deployment with specific
// resource requests and limits for e2e test purposes.
func NewHamsterDeploymentWithResourcesAndLimits(f *framework.Framework, cpuQuantityRequest, memoryQuantityRequest, cpuQuantityLimit, memoryQuantityLimit resource.Quantity) *appsv1.Deployment {
	d := NewHamsterDeploymentWithResources(f, cpuQuantityRequest, memoryQuantityRequest)
	d.Spec.Template.Spec.Containers[0].Resources.Limits = apiv1.ResourceList{
		apiv1.ResourceCPU:    cpuQuantityLimit,
		apiv1.ResourceMemory: memoryQuantityLimit,
	}
	return d
}

func getPodSelectorExcludingDonePodsOrDie() string {
	stringSelector := "status.phase!=" + string(apiv1.PodSucceeded) +
		",status.phase!=" + string(apiv1.PodFailed)
	selector := fields.ParseSelectorOrDie(stringSelector)
	return selector.String()
}

// GetHamsterPods returns running hamster pods (matched by utils.HamsterLabels)
func GetHamsterPods(f *framework.Framework) (*apiv1.PodList, error) {
	label := labels.SelectorFromSet(labels.Set(utils.HamsterLabels))
	options := metav1.ListOptions{LabelSelector: label.String(), FieldSelector: getPodSelectorExcludingDonePodsOrDie()}
	return f.ClientSet.CoreV1().Pods(f.Namespace.Name).List(context.TODO(), options)
}

// NewTestCronJob returns a CronJob for test purposes.
func NewTestCronJob(name, schedule string, replicas int32) *batchv1.CronJob {
	backoffLimit := utils.DefaultHamsterBackoffLimit
	sj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "CronJob",
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          schedule,
			ConcurrencyPolicy: batchv1.AllowConcurrent,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Parallelism:  &replicas,
					Completions:  &replicas,
					BackoffLimit: &backoffLimit,
					Template: apiv1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"job": name},
						},
						Spec: apiv1.PodSpec{
							RestartPolicy: apiv1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}

	return sj
}

func waitForActiveJobs(c clientset.Interface, ns, cronJobName string, active int) error {
	return wait.PollUntilContextTimeout(context.Background(), framework.Poll, cronJobsWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
		curr, err := getCronJob(c, ns, cronJobName)
		if err != nil {
			return false, err
		}
		return len(curr.Status.Active) >= active, nil
	})
}

func createCronJob(c clientset.Interface, ns string, cronJob *batchv1.CronJob) (*batchv1.CronJob, error) {
	return c.BatchV1().CronJobs(ns).Create(context.TODO(), cronJob, metav1.CreateOptions{})
}

func getCronJob(c clientset.Interface, ns, name string) (*batchv1.CronJob, error) {
	return c.BatchV1().CronJobs(ns).Get(context.TODO(), name, metav1.GetOptions{})
}

// SetupHamsterCronJob creates and sets up a new CronJob
func SetupHamsterCronJob(f *framework.Framework, schedule, cpu, memory string, replicas int32) {
	cronJob := NewTestCronJob("hamster-cronjob", schedule, replicas)
	cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers = []apiv1.Container{SetupHamsterContainer(cpu, memory)}
	for label, value := range utils.HamsterLabels {
		cronJob.Spec.JobTemplate.Spec.Template.Labels[label] = value
	}
	cronJob, err := createCronJob(f.ClientSet, f.Namespace.Name, cronJob)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = waitForActiveJobs(f.ClientSet, f.Namespace.Name, cronJob.Name, 1)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

// SetupHamsterContainer returns container with given amount of cpu and memory
func SetupHamsterContainer(cpu, memory string) apiv1.Container {
	cpuQuantity := ParseQuantityOrDie(cpu)
	memoryQuantity := ParseQuantityOrDie(memory)

	return apiv1.Container{
		Name:  "hamster",
		Image: "registry.k8s.io/ubuntu-slim:0.14",
		Resources: apiv1.ResourceRequirements{
			Requests: apiv1.ResourceList{
				apiv1.ResourceCPU:    cpuQuantity,
				apiv1.ResourceMemory: memoryQuantity,
			},
		},
		Command: []string{"/bin/sh"},
		Args:    []string{"-c", "while true; do sleep 10 ; done"},
	}
}

// InstallRawVPA installs a VPA object passed in as raw json in the test cluster.
func InstallRawVPA(f *framework.Framework, obj interface{}) error {
	vpaClientSet := utils.GetVpaClientSet(f)
	err := vpaClientSet.AutoscalingV1().RESTClient().Post().
		Namespace(f.Namespace.Name).
		Resource("verticalpodautoscalers").
		Body(obj).
		Do(context.TODO())
	return err.Error()
}

// AnnotatePod adds annotation for an existing pod.
func AnnotatePod(f *framework.Framework, podName, annotationName, annotationValue string) {
	bytes, err := json.Marshal([]utils.PatchRecord{{
		Op:    "add",
		Path:  fmt.Sprintf("/metadata/annotations/%v", annotationName),
		Value: annotationValue,
	}})
	pod, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).Patch(context.TODO(), podName, types.JSONPatchType, bytes, metav1.PatchOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to patch pod.")
	gomega.Expect(pod.Annotations[annotationName]).To(gomega.Equal(annotationValue))
}

// ParseQuantityOrDie parses quantity from string and dies with an error if
// unparsable.
func ParseQuantityOrDie(text string) resource.Quantity {
	quantity, err := resource.ParseQuantity(text)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error parsing quantity: %s", text)
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

	return wait.PollUntilContextTimeout(context.Background(), utils.PollInterval, utils.PollTimeout, true, func(ctx context.Context) (done bool, err error) {
		currentPodList, err := GetHamsterPods(f)
		if err != nil {
			return false, err
		}
		currentPodSet := MakePodSet(currentPodList)
		return WerePodsSuccessfullyRestarted(currentPodSet, initialPodSet), nil
	})
}

// WaitForPodsEvicted waits until some pods from the list are evicted.
func WaitForPodsEvicted(f *framework.Framework, podList *apiv1.PodList) error {
	initialPodSet := MakePodSet(podList)

	return wait.PollUntilContextTimeout(context.Background(), utils.PollInterval, utils.PollTimeout, true, func(ctx context.Context) (done bool, err error) {
		currentPodList, err := GetHamsterPods(f)
		if err != nil {
			return false, err
		}
		currentPodSet := MakePodSet(currentPodList)
		return GetEvictedPodsCount(currentPodSet, initialPodSet) > 0, nil
	})
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
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error when listing hamster pods to check number of pod evictions")
	restarted := GetEvictedPodsCount(MakePodSet(currentPodList), initialPodSet)
	gomega.Expect(restarted).To(gomega.Equal(0), "there should be no pod evictions")
}

// WaitForUncappedCPURecommendationAbove pools VPA object until uncapped recommendation is above specified value.
// Returns polled VPA object. On timeout returns error.
func WaitForUncappedCPURecommendationAbove(c vpa_clientset.Interface, vpa *vpa_types.VerticalPodAutoscaler, minMilliCPU int64) (*vpa_types.VerticalPodAutoscaler, error) {
	return utils.WaitForVPAMatch(c, vpa, func(vpa *vpa_types.VerticalPodAutoscaler) bool {
		if vpa.Status.Recommendation == nil || len(vpa.Status.Recommendation.ContainerRecommendations) == 0 {
			return false
		}
		uncappedCpu := vpa.Status.Recommendation.ContainerRecommendations[0].UncappedTarget[apiv1.ResourceCPU]
		return uncappedCpu.MilliValue() > minMilliCPU
	})
}

func installLimitRange(f *framework.Framework, minCpuLimit, minMemoryLimit, maxCpuLimit, maxMemoryLimit *resource.Quantity, lrType apiv1.LimitType) {
	lr := &apiv1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.Namespace.Name,
			Name:      "hamster-lr",
		},
		Spec: apiv1.LimitRangeSpec{
			Limits: []apiv1.LimitRangeItem{},
		},
	}

	if maxMemoryLimit != nil || maxCpuLimit != nil {
		lrItem := apiv1.LimitRangeItem{
			Type: lrType,
			Max:  apiv1.ResourceList{},
		}
		if maxCpuLimit != nil {
			lrItem.Max[apiv1.ResourceCPU] = *maxCpuLimit
		}
		if maxMemoryLimit != nil {
			lrItem.Max[apiv1.ResourceMemory] = *maxMemoryLimit
		}
		lr.Spec.Limits = append(lr.Spec.Limits, lrItem)
	}

	if minMemoryLimit != nil || minCpuLimit != nil {
		lrItem := apiv1.LimitRangeItem{
			Type: lrType,
			Min:  apiv1.ResourceList{},
		}
		if minCpuLimit != nil {
			lrItem.Min[apiv1.ResourceCPU] = *minCpuLimit
		}
		if minMemoryLimit != nil {
			lrItem.Min[apiv1.ResourceMemory] = *minMemoryLimit
		}
		lr.Spec.Limits = append(lr.Spec.Limits, lrItem)
	}
	_, err := f.ClientSet.CoreV1().LimitRanges(f.Namespace.Name).Create(context.TODO(), lr, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error when creating limit range")
}

// InstallLimitRangeWithMax installs a LimitRange with a maximum limit for CPU and memory.
func InstallLimitRangeWithMax(f *framework.Framework, maxCpuLimit, maxMemoryLimit string, lrType apiv1.LimitType) {
	ginkgo.By(fmt.Sprintf("Setting up LimitRange with max limits - CPU: %v, memory: %v", maxCpuLimit, maxMemoryLimit))
	maxCpuLimitQuantity := ParseQuantityOrDie(maxCpuLimit)
	maxMemoryLimitQuantity := ParseQuantityOrDie(maxMemoryLimit)
	installLimitRange(f, nil, nil, &maxCpuLimitQuantity, &maxMemoryLimitQuantity, lrType)
}

// InstallLimitRangeWithMin installs a LimitRange with a minimum limit for CPU and memory.
func InstallLimitRangeWithMin(f *framework.Framework, minCpuLimit, minMemoryLimit string, lrType apiv1.LimitType) {
	ginkgo.By(fmt.Sprintf("Setting up LimitRange with min limits - CPU: %v, memory: %v", minCpuLimit, minMemoryLimit))
	minCpuLimitQuantity := ParseQuantityOrDie(minCpuLimit)
	minMemoryLimitQuantity := ParseQuantityOrDie(minMemoryLimit)
	installLimitRange(f, &minCpuLimitQuantity, &minMemoryLimitQuantity, nil, nil, lrType)
}

// WaitForPodsUpdatedWithoutEviction waits for pods to be updated without any evictions taking place over the polling
// interval.
// TODO: Use events to track in-place resizes instead of polling when ready: https://github.com/kubernetes/kubernetes/issues/127172
func WaitForPodsUpdatedWithoutEviction(f *framework.Framework, initialPods *apiv1.PodList) error {
	framework.Logf("waiting for at least one pod to be updated without eviction")
	err := wait.PollUntilContextTimeout(context.TODO(), utils.PollInterval, VpaInPlaceTimeout, false, func(context.Context) (bool, error) {
		podList, err := GetHamsterPods(f)
		if err != nil {
			return false, err
		}
		resourcesHaveDiffered := false
		podMissing := false
		for _, initialPod := range initialPods.Items {
			found := false
			for _, pod := range podList.Items {
				if initialPod.Name == pod.Name {
					found = true
					for num, container := range pod.Status.ContainerStatuses {
						for resourceName, resourceLimit := range container.Resources.Limits {
							initialResourceLimit := initialPod.Status.ContainerStatuses[num].Resources.Limits[resourceName]
							if !resourceLimit.Equal(initialResourceLimit) {
								framework.Logf("%s/%s: %s limit status(%v) differs from initial limit spec(%v)", pod.Name, container.Name, resourceName, resourceLimit.String(), initialResourceLimit.String())
								resourcesHaveDiffered = true
							}
						}
						for resourceName, resourceRequest := range container.Resources.Requests {
							initialResourceRequest := initialPod.Status.ContainerStatuses[num].Resources.Requests[resourceName]
							if !resourceRequest.Equal(initialResourceRequest) {
								framework.Logf("%s/%s: %s request status(%v) differs from initial request spec(%v)", pod.Name, container.Name, resourceName, resourceRequest.String(), initialResourceRequest.String())
								resourcesHaveDiffered = true
							}
						}
					}
				}
			}
			if !found {
				podMissing = true
			}
		}
		if podMissing {
			return true, fmt.Errorf("a pod was erroneously evicted")
		}
		if resourcesHaveDiffered {
			framework.Logf("after checking %d pods, resources have started to differ for at least one of them", len(podList.Items))
			return true, nil
		}
		return false, nil
	})
	framework.Logf("finished waiting for at least one pod to be updated without eviction")
	return err
}

func anyContainsSubstring(arr []string, substr string) bool {
	for _, s := range arr {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
