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
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	mpa_clientset "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/test"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"

	ginkgo "github.com/onsi/ginkgo/v2"
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

func (*observer) OnAdd(obj interface{}, isInInitialList bool) {}
func (*observer) OnDelete(obj interface{})                    {}

func (o *observer) OnUpdate(oldObj, newObj interface{}) {
	get := func(mpa *mpa_types.MultidimPodAutoscaler) (result resourceRecommendation, found bool) {
		if mpa.Status.Recommendation == nil || len(mpa.Status.Recommendation.ContainerRecommendations) == 0 {
			found = false
			result = resourceRecommendation{}
		} else {
			found = true
			result = getResourceRecommendation(&mpa.Status.Recommendation.ContainerRecommendations[0], apiv1.ResourceCPU)
		}
		return
	}
	oldMPA, _ := oldObj.(*mpa_types.MultidimPodAutoscaler)
	NewMPA, _ := newObj.(*mpa_types.MultidimPodAutoscaler)
	oldRecommendation, oldFound := get(oldMPA)
	newRecommendation, newFound := get(NewMPA)
	result := recommendationChange{
		oldMissing: !oldFound,
		newMissing: !newFound,
		diff:       newRecommendation.sub(&oldRecommendation),
	}
	go func() { o.channel <- result }()
}

func getMpaObserver(mpaClientSet mpa_clientset.Interface) *observer {
	mpaListWatch := cache.NewListWatchFromClient(mpaClientSet.AutoscalingV1alpha1().RESTClient(), "multidimpodautoscalers", apiv1.NamespaceAll, fields.Everything())
	mpaObserver := observer{channel: make(chan recommendationChange)}
	_, controller := cache.NewIndexerInformer(mpaListWatch,
		&mpa_types.MultidimPodAutoscaler{},
		1*time.Hour,
		&mpaObserver,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	go controller.Run(make(chan struct{}))
	if !cache.WaitForCacheSync(make(chan struct{}), controller.HasSynced) {
		klog.Fatalf("Failed to sync MPA cache during initialization")
	} else {
		klog.InfoS("Initial MPA synced successfully")
	}
	return &mpaObserver
}

var _ = RecommenderE2eDescribe("Checkpoints", func() {
	f := framework.NewDefaultFramework("multidimensional-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	ginkgo.It("with missing MPA objects are garbage collected", func() {
		ns := f.Namespace.Name
		mpaClientSet := getMpaClientSet(f)

		checkpoint := mpa_types.MultidimPodAutoscalerCheckpoint{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: ns,
			},
			Spec: mpa_types.MultidimPodAutoscalerCheckpointSpec{
				MPAObjectName: "some-mpa",
			},
		}

		_, err := mpaClientSet.AutoscalingV1alpha1().MultidimPodAutoscalerCheckpoints(ns).Create(context.TODO(), &checkpoint, metav1.CreateOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		klog.InfoS("Sleeping for up to 15 minutes...")

		maxRetries := 90
		retryDelay := 10 * time.Second
		for i := 0; i < maxRetries; i++ {
			list, err := mpaClientSet.AutoscalingV1alpha1().MultidimPodAutoscalerCheckpoints(ns).List(context.TODO(), metav1.ListOptions{})
			if err == nil && len(list.Items) == 0 {
				break
			}
			klog.InfoS("Still waiting...")
			time.Sleep(retryDelay)
		}

		list, err := mpaClientSet.AutoscalingV1alpha1().MultidimPodAutoscalerCheckpoints(ns).List(context.TODO(), metav1.ListOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(list.Items).To(gomega.BeEmpty())
	})
})

var _ = RecommenderE2eDescribe("MPA CRD object", func() {
	f := framework.NewDefaultFramework("multidimensional-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	ginkgo.It("serves recommendation for CronJob", func() {
		ginkgo.By("Setting up hamster CronJob")
		SetupHamsterCronJob(f, "*/5 * * * *", "100m", "100Mi", defaultHamsterReplicas)

		mpaClientSet := getMpaClientSet(f)

		ginkgo.By("Setting up MPA")
		targetRef := &autoscaling.CrossVersionObjectReference{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Name:       "hamster-cronjob",
		}

		containerName := GetHamsterContainerNameByIndex(0)
		mpaCRD := test.MultidimPodAutoscaler().
			WithName("hamster-mpa").
			WithNamespace(f.Namespace.Name).
			WithScaleTargetRef(targetRef).
			WithContainer(containerName).
			Get()

		InstallMPA(f, mpaCRD)

		ginkgo.By("Waiting for recommendation to be filled")
		_, err := WaitForRecommendationPresent(mpaClientSet, mpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

var _ = RecommenderE2eDescribe("MPA CRD object", func() {
	f := framework.NewDefaultFramework("multidimensional-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	var (
		mpaCRD       *mpa_types.MultidimPodAutoscaler
		mpaClientSet mpa_clientset.Interface
	)

	ginkgo.BeforeEach(func() {
		ginkgo.By("Setting up a hamster deployment")
		_ = SetupHamsterDeployment(
			f,       /* framework */
			"100m",  /* cpu */
			"100Mi", /* memory */
			1,       /* number of replicas */
		)

		ginkgo.By("Setting up a MPA CRD")
		containerName := GetHamsterContainerNameByIndex(0)
		mpaCRD = test.MultidimPodAutoscaler().
			WithName("hamster-mpa").
			WithNamespace(f.Namespace.Name).
			WithScaleTargetRef(hamsterTargetRef).
			WithContainer(containerName).
			Get()

		InstallMPA(f, mpaCRD)

		mpaClientSet = getMpaClientSet(f)
	})

	ginkgo.It("serves recommendation", func() {
		ginkgo.By("Waiting for recommendation to be filled")
		_, err := WaitForRecommendationPresent(mpaClientSet, mpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("doesn't drop lower/upper after recommender's restart", func() {

		o := getMpaObserver(mpaClientSet)

		ginkgo.By("Waiting for recommendation to be filled")
		_, err := WaitForRecommendationPresent(mpaClientSet, mpaCRD)
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
		ginkgo.By("Accumulating diffs after restart, sleeping for 5 minutes...")
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

var _ = RecommenderE2eDescribe("MPA CRD object", func() {
	f := framework.NewDefaultFramework("multidimensional-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	var (
		mpaClientSet mpa_clientset.Interface
	)

	ginkgo.BeforeEach(func() {
		ginkgo.By("Setting up a hamster deployment")
		_ = SetupHamsterDeployment(
			f,       /* framework */
			"100m",  /* cpu */
			"100Mi", /* memory */
			1,       /* number of replicas */
		)

		mpaClientSet = getMpaClientSet(f)
	})

	ginkgo.It("respects min allowed recommendation", func() {
		const minMilliCpu = 10000
		ginkgo.By("Setting up a MPA CRD")
		containerName := GetHamsterContainerNameByIndex(0)
		mpaCRD2 := test.MultidimPodAutoscaler().
			WithName("hamster-mpa").
			WithNamespace(f.Namespace.Name).
			WithScaleTargetRef(hamsterTargetRef).
			WithContainer(containerName).
			WithMinAllowed(containerName, "10000", "").
			Get()

		InstallMPA(f, mpaCRD2)
		mpaCRD := mpaCRD2

		ginkgo.By("Waiting for recommendation to be filled")
		mpa, err := WaitForRecommendationPresent(mpaClientSet, mpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(mpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(1))
		cpu := getMilliCpu(mpa.Status.Recommendation.ContainerRecommendations[0].Target)
		gomega.Expect(cpu).Should(gomega.BeNumerically(">=", minMilliCpu),
			fmt.Sprintf("target cpu recommendation should be greater than or equal to %dm", minMilliCpu))
		cpuUncapped := getMilliCpu(mpa.Status.Recommendation.ContainerRecommendations[0].UncappedTarget)
		gomega.Expect(cpuUncapped).Should(gomega.BeNumerically("<", minMilliCpu),
			fmt.Sprintf("uncapped target cpu recommendation should be less than %dm", minMilliCpu))
	})

	ginkgo.It("respects max allowed recommendation", func() {
		const maxMilliCpu = 1
		ginkgo.By("Setting up a MPA CRD")
		containerName := GetHamsterContainerNameByIndex(0)
		mpaCRD := test.MultidimPodAutoscaler().
			WithName("hamster-mpa").
			WithNamespace(f.Namespace.Name).
			WithScaleTargetRef(hamsterTargetRef).
			WithContainer(containerName).
			WithMaxAllowed(containerName, "1m", "").
			Get()

		InstallMPA(f, mpaCRD)

		ginkgo.By("Waiting for recommendation to be filled")
		mpa, err := WaitForUncappedCPURecommendationAbove(mpaClientSet, mpaCRD, maxMilliCpu)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), fmt.Sprintf(
			"Timed out waiting for uncapped cpu recommendation above %d mCPU", maxMilliCpu))
		gomega.Expect(mpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(1))
		cpu := getMilliCpu(mpa.Status.Recommendation.ContainerRecommendations[0].Target)
		gomega.Expect(cpu).Should(gomega.BeNumerically("<=", maxMilliCpu),
			fmt.Sprintf("target cpu recommendation should be less than or equal to %dm", maxMilliCpu))
	})
})

func getMilliCpu(resources apiv1.ResourceList) int64 {
	cpu := resources[apiv1.ResourceCPU]
	return cpu.MilliValue()
}

var _ = RecommenderE2eDescribe("MPA CRD object", func() {
	f := framework.NewDefaultFramework("multidimensional-pod-autoscaling")
	f.NamespacePodSecurityEnforceLevel = podsecurity.LevelBaseline

	var mpaClientSet mpa_clientset.Interface

	ginkgo.BeforeEach(func() {
		mpaClientSet = getMpaClientSet(f)
	})

	ginkgo.It("with no containers opted out all containers get recommendations", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := NewNHamstersDeployment(f, 2 /*number of containers*/)
		_ = startDeploymentPods(f, d)

		ginkgo.By("Setting up MPA CRD")
		container1Name := GetHamsterContainerNameByIndex(0)
		container2Name := GetHamsterContainerNameByIndex(1)
		mpaCRD := test.MultidimPodAutoscaler().
			WithName("hamster-mpa").
			WithNamespace(f.Namespace.Name).
			WithScaleTargetRef(hamsterTargetRef).
			WithContainer(container1Name).
			WithContainer(container2Name).
			Get()

		InstallMPA(f, mpaCRD)

		ginkgo.By("Waiting for recommendation to be filled for both containers")
		mpa, err := WaitForRecommendationPresent(mpaClientSet, mpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(mpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(2))
	})

	ginkgo.It("only containers not-opted-out get recommendations", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := NewNHamstersDeployment(f, 2 /*number of containers*/)
		_ = startDeploymentPods(f, d)

		ginkgo.By("Setting up MPA CRD")
		container1Name := GetHamsterContainerNameByIndex(0)
		container2Name := GetHamsterContainerNameByIndex(1)
		mpaCRD := test.MultidimPodAutoscaler().
			WithName("hamster-mpa").
			WithNamespace(f.Namespace.Name).
			WithScaleTargetRef(hamsterTargetRef).
			WithContainer(container1Name).
			WithScalingMode(container1Name, vpa_types.ContainerScalingModeOff).
			WithContainer(container2Name).
			Get()

		InstallMPA(f, mpaCRD)

		ginkgo.By("Waiting for recommendation to be filled for just one container")
		mpa, err := WaitForRecommendationPresent(mpaClientSet, mpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		errMsg := fmt.Sprintf("%s container has recommendations turned off. We expect expect only recommendations for %s",
			GetHamsterContainerNameByIndex(0),
			GetHamsterContainerNameByIndex(1))
		gomega.Expect(mpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(1), errMsg)
		gomega.Expect(mpa.Status.Recommendation.ContainerRecommendations[0].ContainerName).To(gomega.Equal(GetHamsterContainerNameByIndex(1)), errMsg)
	})
})

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
		if strings.HasPrefix(pod.Name, "mpa-recommender") {
			fmt.Println("Deleting pod.", namespace, pod.Name)
			err := c.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("mpa recommender not found")
}
