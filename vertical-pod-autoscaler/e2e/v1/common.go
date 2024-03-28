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
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	framework_deployment "k8s.io/kubernetes/test/e2e/framework/deployment"
)

const (
	recommenderComponent         = "recommender"
	updateComponent              = "updater"
	admissionControllerComponent = "admission-controller"
	fullVpaSuite                 = "full-vpa"
	actuationSuite               = "actuation"
	pollInterval                 = 10 * time.Second
	pollTimeout                  = 15 * time.Minute
	cronJobsWaitTimeout          = 15 * time.Minute
	// VpaEvictionTimeout is a timeout for VPA to restart a pod if there are no
	// mechanisms blocking it (for example PDB).
	VpaEvictionTimeout = 3 * time.Minute

	defaultHamsterReplicas     = int32(3)
	defaultHamsterBackoffLimit = int32(10)
)

var hamsterTargetRef = &autoscaling.CrossVersionObjectReference{
	APIVersion: "apps/v1",
	Kind:       "Deployment",
	Name:       "hamster-deployment",
}

var hamsterLabels = map[string]string{"app": "hamster"}

// SIGDescribe adds sig-autoscaling tag to test description.
func SIGDescribe(text string, body func()) bool {
	return ginkgo.Describe(fmt.Sprintf("[sig-autoscaling] %v", text), body)
}

// E2eDescribe describes a VPA e2e test.
func E2eDescribe(scenario, name string, body func()) bool {
	return SIGDescribe(fmt.Sprintf("[VPA] [%s] [v1] %s", scenario, name), body)
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
	return NewNHamstersDeployment(f, 1)
}

// NewNHamstersDeployment creates a simple hamster deployment with n containers
// for e2e test purposes.
func NewNHamstersDeployment(f *framework.Framework, n int) *appsv1.Deployment {
	if n < 1 {
		panic("container count should be greater than 0")
	}
	d := framework_deployment.NewDeployment(
		"hamster-deployment",                       /*deploymentName*/
		defaultHamsterReplicas,                     /*replicas*/
		hamsterLabels,                              /*podLabels*/
		GetHamsterContainerNameByIndex(0),          /*imageName*/
		"registry.k8s.io/ubuntu-slim:0.1",          /*image*/
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

// NewHamsterDeploymentWithResources creates a simple hamster deployment with specific
// resource requests for e2e test purposes.
func NewHamsterDeploymentWithResources(f *framework.Framework, cpuQuantity, memoryQuantity resource.Quantity) *appsv1.Deployment {
	d := NewHamsterDeployment(f)
	d.Spec.Template.Spec.Containers[0].Resources.Requests = apiv1.ResourceList{
		apiv1.ResourceCPU:    cpuQuantity,
		apiv1.ResourceMemory: memoryQuantity,
	}
	// TODO(jkyros): It seems to behave differently if we have limits?
	/*
		cpuQuantity.Add(resource.MustParse("100m"))
		memoryQuantity.Add(resource.MustParse("100Mi"))

		d.Spec.Template.Spec.Containers[0].Resources.Limits = apiv1.ResourceList{
			apiv1.ResourceCPU:    cpuQuantity,
			apiv1.ResourceMemory: memoryQuantity,
		}*/
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

// GetHamsterPods returns running hamster pods (matched by hamsterLabels)
func GetHamsterPods(f *framework.Framework) (*apiv1.PodList, error) {
	label := labels.SelectorFromSet(labels.Set(hamsterLabels))
	options := metav1.ListOptions{LabelSelector: label.String(), FieldSelector: getPodSelectorExcludingDonePodsOrDie()}
	return f.ClientSet.CoreV1().Pods(f.Namespace.Name).List(context.TODO(), options)
}

// NewTestCronJob returns a CronJob for test purposes.
func NewTestCronJob(name, schedule string, replicas int32) *batchv1.CronJob {
	backoffLimit := defaultHamsterBackoffLimit
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
	return wait.Poll(framework.Poll, cronJobsWaitTimeout, func() (bool, error) {
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
	for label, value := range hamsterLabels {
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
		Image: "registry.k8s.io/ubuntu-slim:0.1",
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

type patchRecord struct {
	Op    string      `json:"op,inline"`
	Path  string      `json:"path,inline"`
	Value interface{} `json:"value"`
}

func getVpaClientSet(f *framework.Framework) vpa_clientset.Interface {
	config, err := framework.LoadConfig()
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error loading framework")
	return vpa_clientset.NewForConfigOrDie(config)
}

// InstallVPA installs a VPA object in the test cluster.
func InstallVPA(f *framework.Framework, vpa *vpa_types.VerticalPodAutoscaler) {
	vpaClientSet := getVpaClientSet(f)
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

// InstallRawVPA installs a VPA object passed in as raw json in the test cluster.
func InstallRawVPA(f *framework.Framework, obj interface{}) error {
	vpaClientSet := getVpaClientSet(f)
	err := vpaClientSet.AutoscalingV1().RESTClient().Post().
		Namespace(f.Namespace.Name).
		Resource("verticalpodautoscalers").
		Body(obj).
		Do(context.TODO())
	return err.Error()
}

// PatchVpaRecommendation installs a new recommendation for VPA object.
func PatchVpaRecommendation(f *framework.Framework, vpa *vpa_types.VerticalPodAutoscaler,
	recommendation *vpa_types.RecommendedPodResources) {
	newStatus := vpa.Status.DeepCopy()
	newStatus.Recommendation = recommendation
	bytes, err := json.Marshal([]patchRecord{{
		Op:    "replace",
		Path:  "/status",
		Value: *newStatus,
	}})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	_, err = getVpaClientSet(f).AutoscalingV1().VerticalPodAutoscalers(f.Namespace.Name).Patch(context.TODO(), vpa.Name, types.JSONPatchType, bytes, metav1.PatchOptions{}, "status")
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to patch VPA.")
}

// AnnotatePod adds annotation for an existing pod.
func AnnotatePod(f *framework.Framework, podName, annotationName, annotationValue string) {
	bytes, err := json.Marshal([]patchRecord{{
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

	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		currentPodList, err := GetHamsterPods(f)
		if err != nil {
			return false, err
		}
		currentPodSet := MakePodSet(currentPodList)
		return WerePodsSuccessfullyRestarted(currentPodSet, initialPodSet), nil
	})

	if err != nil {
		return fmt.Errorf("waiting for set of pods changed: %v", err)
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
		return fmt.Errorf("waiting for set of pods changed: %v", err)
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
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "unexpected error when listing hamster pods to check number of pod evictions")
	restarted := GetEvictedPodsCount(MakePodSet(currentPodList), initialPodSet)
	gomega.Expect(restarted).To(gomega.Equal(0), "there should be no pod evictions")
}

// WaitForVPAMatch pools VPA object until match function returns true. Returns
// polled vpa object. On timeout returns error.
func WaitForVPAMatch(c vpa_clientset.Interface, vpa *vpa_types.VerticalPodAutoscaler, match func(vpa *vpa_types.VerticalPodAutoscaler) bool) (*vpa_types.VerticalPodAutoscaler, error) {
	var polledVpa *vpa_types.VerticalPodAutoscaler
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		var err error
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

// WaitForRecommendationPresent pools VPA object until recommendations are not empty. Returns
// polled vpa object. On timeout returns error.
func WaitForRecommendationPresent(c vpa_clientset.Interface, vpa *vpa_types.VerticalPodAutoscaler) (*vpa_types.VerticalPodAutoscaler, error) {
	return WaitForVPAMatch(c, vpa, func(vpa *vpa_types.VerticalPodAutoscaler) bool {
		return vpa.Status.Recommendation != nil && len(vpa.Status.Recommendation.ContainerRecommendations) != 0
	})
}

// WaitForUncappedCPURecommendationAbove pools VPA object until uncapped recommendation is above specified value.
// Returns polled VPA object. On timeout returns error.
func WaitForUncappedCPURecommendationAbove(c vpa_clientset.Interface, vpa *vpa_types.VerticalPodAutoscaler, minMilliCPU int64) (*vpa_types.VerticalPodAutoscaler, error) {
	return WaitForVPAMatch(c, vpa, func(vpa *vpa_types.VerticalPodAutoscaler) bool {
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
func WaitForPodsUpdatedWithoutEviction(f *framework.Framework, initialPods, podList *apiv1.PodList) error {
	// TODO(jkyros): This needs to be:
	// 1. Make sure we wait for each of the containers to get an update queued
	// 2. Make sure each of the containers actually finish the update
	// 3. Once everyone has gone through 1 cycle, we don't care anymore, we can move on (it will keep scaling obviously)
	framework.Logf("waiting for update to start and resources to differ")
	var resourcesHaveDiffered bool
	err := wait.PollUntilContextTimeout(context.TODO(), pollInterval, pollTimeout, false, func(context.Context) (bool, error) {
		// TODO(jkyros): make sure we don't update too many pods at once
		podList, err := GetHamsterPods(f)
		if err != nil {
			return false, err
		}
		resourcesAreSynced := true
		podMissing := false
		// Go through the list of initial pods
		for _, initialPod := range initialPods.Items {
			found := false
			// Go through the list of pods we have now
			for _, pod := range podList.Items {
				// If we still have our initial pod, good
				if initialPod.Name == pod.Name {
					found = true

					// Check to see if we have our container resources updated
					for num, container := range pod.Spec.Containers {
						// If our current spec differs from initial, we know we were told to update
						if !resourcesHaveDiffered {
							for resourceName, resourceLimit := range container.Resources.Limits {
								initialResourceLimit := initialPod.Spec.Containers[num].Resources.Limits[resourceName]
								if !initialResourceLimit.Equal(resourceLimit) {
									framework.Logf("E: %s/%s: %s limit (%v) differs from initial (%v), change has started ", pod.Name, container.Name, resourceName, resourceLimit.String(), initialResourceLimit.String())
									//fmt.Printf("UPD: L:%s: %s/%s %v differs from initial %v\n", resourceName, pod.Name, container.Name, resourceLimit, pod.Status.ContainerStatuses[num].Resources.Limits[resourceName])
									resourcesHaveDiffered = true

								}

							}
							for resourceName, resourceRequest := range container.Resources.Requests {
								initialResourceRequest := initialPod.Spec.Containers[num].Resources.Requests[resourceName]
								if !initialResourceRequest.Equal(resourceRequest) {
									framework.Logf("%s/%s: %s request (%v) differs from initial (%v), change has started ", pod.Name, container.Name, resourceName, resourceRequest.String(), initialResourceRequest.String())
									resourcesHaveDiffered = true

								}
							}
						}

						if len(pod.Status.ContainerStatuses) > num {
							if pod.Status.ContainerStatuses[num].Resources != nil {
								for resourceName, resourceLimit := range container.Resources.Limits {
									statusResourceLimit := pod.Status.ContainerStatuses[num].Resources.Limits[resourceName]
									if !statusResourceLimit.Equal(resourceLimit) {
										framework.Logf("%s/%s: %s limit status (%v) differs from limit spec (%v), still in progress", pod.Name, container.Name, resourceName, resourceLimit.String(), statusResourceLimit.String())

										resourcesAreSynced = false

									}

								}
								for resourceName, resourceRequest := range container.Resources.Requests {
									statusResourceRequest := pod.Status.ContainerStatuses[num].Resources.Requests[resourceName]
									if !pod.Status.ContainerStatuses[num].Resources.Requests[resourceName].Equal(resourceRequest) {
										framework.Logf("%s/%s: %s request status (%v) differs from request spec(%v), still in progress ", pod.Name, container.Name, resourceName, resourceRequest.String(), statusResourceRequest.String())
										resourcesAreSynced = false

									}
								}

							} else {
								framework.Logf("SOMEHOW ITS EMPTY\n")
							}
						}

					}
				}

			}
			if !found {
				//framework.Logf("pod %s was evicted and should not have been\n", initialPod.Name)
				podMissing = true
			}

		}
		if podMissing {
			return false, fmt.Errorf("a pod was erroneously evicted")
		}
		if len(podList.Items) > 0 && resourcesAreSynced {
			if !resourcesHaveDiffered {
				return false, nil
			}
			framework.Logf("after checking %d pods, were are in sync\n", len(podList.Items))
			return true, nil
		}
		return false, nil
	})
	return err
}
