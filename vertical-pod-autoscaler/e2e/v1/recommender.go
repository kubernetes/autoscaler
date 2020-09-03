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
	"fmt"
	"strings"
	"time"

	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

type resourceRecommendation struct {
	target, lower, upper int64
}

func (r *resourceRecommendation) sub(other *resourceRecommendation) resourceRecommendation {
	return resourceRecommendation{
		target: r.target - other.target,
		lower:  r.lower - other.lower,
		upper:  r.upper - other.upper,
	}

}

func getResourceRecommendation(containerRecommendation *vpa_types.RecommendedContainerResources, r apiv1.ResourceName) resourceRecommendation {
	getOrZero := func(resourceList apiv1.ResourceList) int64 {
		value, found := resourceList[r]
		if found {
			return value.Value()
		}
		return 0
	}
	return resourceRecommendation{
		target: getOrZero(containerRecommendation.Target),
		lower:  getOrZero(containerRecommendation.LowerBound),
		upper:  getOrZero(containerRecommendation.UpperBound),
	}
}

type recommendationChange struct {
	oldMissing, newMissing bool
	diff                   resourceRecommendation
}

type observer struct {
	channel chan recommendationChange
}

func (*observer) OnAdd(obj interface{})    {}
func (*observer) OnDelete(obj interface{}) {}

func (o *observer) OnUpdate(oldObj, newObj interface{}) {
	get := func(vpa *vpa_types.VerticalPodAutoscaler) (result resourceRecommendation, found bool) {
		if vpa.Status.Recommendation == nil || len(vpa.Status.Recommendation.ContainerRecommendations) == 0 {
			found = false
			result = resourceRecommendation{}
		} else {
			found = true
			result = getResourceRecommendation(&vpa.Status.Recommendation.ContainerRecommendations[0], apiv1.ResourceCPU)
		}
		return
	}
	oldVPA, _ := oldObj.(*vpa_types.VerticalPodAutoscaler)
	NewVPA, _ := newObj.(*vpa_types.VerticalPodAutoscaler)
	oldRecommendation, oldFound := get(oldVPA)
	newRecommendation, newFound := get(NewVPA)
	result := recommendationChange{
		oldMissing: !oldFound,
		newMissing: !newFound,
		diff:       newRecommendation.sub(&oldRecommendation),
	}
	go func() { o.channel <- result }()
}

func getVpaObserver(vpaClientSet vpa_clientset.Interface) *observer {
	vpaListWatch := cache.NewListWatchFromClient(vpaClientSet.AutoscalingV1().RESTClient(), "verticalpodautoscalers", apiv1.NamespaceAll, fields.Everything())
	vpaObserver := observer{channel: make(chan recommendationChange)}
	_, controller := cache.NewIndexerInformer(vpaListWatch,
		&vpa_types.VerticalPodAutoscaler{},
		1*time.Hour,
		&vpaObserver,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	go controller.Run(make(chan struct{}))
	if !cache.WaitForCacheSync(make(chan struct{}), controller.HasSynced) {
		klog.Fatalf("Failed to sync VPA cache during initialization")
	} else {
		klog.Info("Initial VPA synced successfully")
	}
	return &vpaObserver
}

var _ = RecommenderE2eDescribe("Checkpoints", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("with missing VPA objects are garbage collected", func() {
		ns := f.Namespace.Name
		vpaClientSet := getVpaClientSet(f)

		checkpoint := vpa_types.VerticalPodAutoscalerCheckpoint{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: ns,
			},
			Spec: vpa_types.VerticalPodAutoscalerCheckpointSpec{
				VPAObjectName: "some-vpa",
			},
		}

		_, err := vpaClientSet.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns).Create(context.TODO(), &checkpoint, metav1.CreateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		time.Sleep(15 * time.Minute)

		list, err := vpaClientSet.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns).List(context.TODO(), metav1.ListOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(list.Items).To(gomega.BeEmpty())
	})
})

var _ = RecommenderE2eDescribe("VPA CRD object", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	ginkgo.It("serves recommendation for CronJob", func() {
		ginkgo.By("Setting up hamster CronJob")
		SetupHamsterCronJob(f, "*/5 * * * *", "100m", "100Mi", defaultHamsterReplicas)

		vpaClientSet := getVpaClientSet(f)

		ginkgo.By("Setting up VPA")
		vpaCRD := NewVPA(f, "hamster-cronjob-vpa", &autoscaling.CrossVersionObjectReference{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Name:       "hamster-cronjob",
		})

		InstallVPA(f, vpaCRD)

		ginkgo.By("Waiting for recommendation to be filled")
		_, err := WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

var _ = RecommenderE2eDescribe("VPA CRD object", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	var (
		vpaCRD       *vpa_types.VerticalPodAutoscaler
		vpaClientSet vpa_clientset.Interface
	)

	ginkgo.BeforeEach(func() {
		ginkgo.By("Setting up a hamster deployment")
		_ = SetupHamsterDeployment(
			f,       /* framework */
			"100m",  /* cpu */
			"100Mi", /* memeory */
			1,       /* number of replicas */
		)

		ginkgo.By("Setting up a VPA CRD")
		vpaCRD = NewVPA(f, "hamster-vpa", hamsterTargetRef)
		InstallVPA(f, vpaCRD)

		vpaClientSet = getVpaClientSet(f)
	})

	ginkgo.It("serves recommendation", func() {
		ginkgo.By("Waiting for recommendation to be filled")
		_, err := WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("doesn't drop lower/upper after recommender's restart", func() {

		o := getVpaObserver(vpaClientSet)

		ginkgo.By("Waiting for recommendation to be filled")
		_, err := WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.By("Drain diffs")
	out:
		for {
			select {
			case recommendationDiff := <-o.channel:
				fmt.Println("Dropping recommendation diff", recommendationDiff)
			default:
				break out
			}
		}
		ginkgo.By("Deleting recommender")
		gomega.Expect(deleteRecommender(f.ClientSet)).To(gomega.BeNil())
		ginkgo.By("Accumulating diffs after restart")
		time.Sleep(5 * time.Minute)
		changeDetected := false
	finish:
		for {
			select {
			case recommendationDiff := <-o.channel:
				fmt.Println("checking recommendation diff", recommendationDiff)
				changeDetected = true
				gomega.Expect(recommendationDiff.oldMissing).To(gomega.Equal(false))
				gomega.Expect(recommendationDiff.newMissing).To(gomega.Equal(false))
				gomega.Expect(recommendationDiff.diff.lower).Should(gomega.BeNumerically(">=", 0))
				gomega.Expect(recommendationDiff.diff.upper).Should(gomega.BeNumerically("<=", 0))
			default:
				break finish
			}
		}
		gomega.Expect(changeDetected).To(gomega.Equal(true))
	})
})

var _ = RecommenderE2eDescribe("VPA CRD object", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	var (
		vpaClientSet vpa_clientset.Interface
	)

	ginkgo.BeforeEach(func() {
		ginkgo.By("Setting up a hamster deployment")
		_ = SetupHamsterDeployment(
			f,       /* framework */
			"100m",  /* cpu */
			"100Mi", /* memeory */
			1,       /* number of replicas */
		)

		vpaClientSet = getVpaClientSet(f)
	})

	ginkgo.It("respects min allowed recommendation", func() {
		const minMilliCpu = 10000
		ginkgo.By("Setting up a VPA CRD")
		minAllowed := apiv1.ResourceList{
			apiv1.ResourceCPU: ParseQuantityOrDie(fmt.Sprintf("%dm", minMilliCpu)),
		}
		vpaCRD := createVpaCRDWithMinMaxAllowed(f, minAllowed, nil)

		ginkgo.By("Waiting for recommendation to be filled")
		vpa, err := WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(1))
		cpu := getMilliCpu(vpa.Status.Recommendation.ContainerRecommendations[0].Target)
		gomega.Expect(cpu).Should(gomega.BeNumerically(">=", minMilliCpu),
			fmt.Sprintf("target cpu recommendation should be greater than or equal to %dm", minMilliCpu))
		cpuUncapped := getMilliCpu(vpa.Status.Recommendation.ContainerRecommendations[0].UncappedTarget)
		gomega.Expect(cpuUncapped).Should(gomega.BeNumerically("<", minMilliCpu),
			fmt.Sprintf("uncapped target cpu recommendation should be less than %dm", minMilliCpu))
	})

	ginkgo.It("respects max allowed recommendation", func() {
		const maxMilliCpu = 1
		ginkgo.By("Setting up a VPA CRD")
		maxAllowed := apiv1.ResourceList{
			apiv1.ResourceCPU: ParseQuantityOrDie(fmt.Sprintf("%dm", maxMilliCpu)),
		}
		vpaCRD := createVpaCRDWithMinMaxAllowed(f, nil, maxAllowed)

		ginkgo.By("Waiting for recommendation to be filled")
		vpa, err := WaitForUncappedCPURecommendationAbove(vpaClientSet, vpaCRD, maxMilliCpu)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), fmt.Sprintf(
			"Timed out waiting for uncapped cpu recommendation above %d mCPU", maxMilliCpu))
		gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(1))
		cpu := getMilliCpu(vpa.Status.Recommendation.ContainerRecommendations[0].Target)
		gomega.Expect(cpu).Should(gomega.BeNumerically("<=", maxMilliCpu),
			fmt.Sprintf("target cpu recommendation should be less than or equal to %dm", maxMilliCpu))
	})
})

func getMilliCpu(resources apiv1.ResourceList) int64 {
	cpu := resources[apiv1.ResourceCPU]
	return cpu.MilliValue()
}

// createVpaCRDWithMinMaxAllowed creates vpa object with min and max resources allowed.
func createVpaCRDWithMinMaxAllowed(f *framework.Framework, minAllowed, maxAllowed apiv1.ResourceList) *vpa_types.VerticalPodAutoscaler {
	vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
	containerResourcePolicies := []vpa_types.ContainerResourcePolicy{
		{
			ContainerName: GetHamsterContainerNameByIndex(0),
			MinAllowed:    minAllowed,
			MaxAllowed:    maxAllowed,
		},
	}
	vpaCRD.Spec.ResourcePolicy = &vpa_types.PodResourcePolicy{
		ContainerPolicies: containerResourcePolicies,
	}
	InstallVPA(f, vpaCRD)
	return vpaCRD
}

var _ = RecommenderE2eDescribe("VPA CRD object", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	var vpaClientSet vpa_clientset.Interface

	ginkgo.BeforeEach(func() {
		vpaClientSet = getVpaClientSet(f)
	})

	ginkgo.It("with no containers opted out all containers get recommendations", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := NewNHamstersDeployment(f, 2 /*number of containers*/)
		_ = startDeploymentPods(f, d)
		ginkgo.By("Setting up VPA CRD")
		vpaCRD := createVpaCRDWithContainerScalingModes(f, vpa_types.ContainerScalingModeAuto, vpa_types.ContainerScalingModeAuto)

		ginkgo.By("Waiting for recommendation to be filled for both containers")
		vpa, err := WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(2))
	})

	ginkgo.It("only containers not-opted-out get recommendations", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := NewNHamstersDeployment(f, 2 /*number of containers*/)
		_ = startDeploymentPods(f, d)
		vpaCRD := createVpaCRDWithContainerScalingModes(f, vpa_types.ContainerScalingModeOff, vpa_types.ContainerScalingModeAuto)

		ginkgo.By("Waiting for recommendation to be filled for just one container")
		vpa, err := WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		errMsg := fmt.Sprintf("%s container has recommendations turned off. We expect expect only recommendations for %s",
			GetHamsterContainerNameByIndex(0),
			GetHamsterContainerNameByIndex(1))
		gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(1), errMsg)
		gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations[0].ContainerName).To(gomega.Equal(GetHamsterContainerNameByIndex(1)), errMsg)
	})
})

// createVpaCRDWithContainerScalingModes creates vpa object with containers policies
// having assigned given scaling modes respectively.
func createVpaCRDWithContainerScalingModes(f *framework.Framework, modes ...vpa_types.ContainerScalingMode) *vpa_types.VerticalPodAutoscaler {
	vpaCRD := NewVPA(f, "hamster-vpa", hamsterTargetRef)
	containerResourcePolicies := make([]vpa_types.ContainerResourcePolicy, len(modes), len(modes))
	for i := range modes {
		containerResourcePolicies[i] = vpa_types.ContainerResourcePolicy{
			ContainerName: GetHamsterContainerNameByIndex(i),
			Mode:          &modes[i],
		}
	}
	vpaCRD.Spec.ResourcePolicy = &vpa_types.PodResourcePolicy{
		ContainerPolicies: containerResourcePolicies,
	}
	InstallVPA(f, vpaCRD)
	return vpaCRD
}

func deleteRecommender(c clientset.Interface) error {
	namespace := "kube-system"
	listOptions := metav1.ListOptions{}
	podList, err := c.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err != nil {
		fmt.Println("Could not list pods.", err)
		return err
	}
	fmt.Println("Pods list items:", len(podList.Items))
	for _, pod := range podList.Items {
		if strings.HasPrefix(pod.Name, "vpa-recommender") {
			fmt.Println("Deleting pod.", namespace, pod.Name)
			err := c.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("vpa recommender not found")
}
