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

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/autoscaler/vertical-pod-autoscaler/e2e/utils"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"
	"k8s.io/kubernetes/test/e2e/framework"
	podsecurity "k8s.io/pod-security-admission/api"
)

func init() {
	// Dynamically register feature gates from the VPA's versioned feature gate configuration
	// This ensures consistency with the main VPA feature gate definitions
	if err := utilfeature.DefaultMutableFeatureGate.Add(features.MutableFeatureGate.GetAll()); err != nil {
		panic(fmt.Sprintf("Failed to add VPA feature gates: %v", err))
	}
}

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

func (*observer) OnAdd(obj any, isInInitialList bool) {}
func (*observer) OnDelete(obj any)                    {}

func (o *observer) OnUpdate(oldObj, newObj any) {
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

func getVpaObserver(vpaClientSet vpa_clientset.Interface, namespace string) *observer {
	vpaListWatch := cache.NewListWatchFromClient(vpaClientSet.AutoscalingV1().RESTClient(), "verticalpodautoscalers", namespace, fields.Everything())
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
		klog.InfoS("Initial VPA synced successfully")
	}
	return &vpaObserver
}

var _ = utils.RecommenderE2eDescribe("Checkpoints", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	f.It("with missing VPA objects are garbage collected", framework.WithSlow(), func() {
		ns := f.Namespace.Name
		vpaClientSet := utils.GetVpaClientSet(f)

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

		klog.InfoS("Polling for up to 15 minutes...")

		var list *vpa_types.VerticalPodAutoscalerCheckpointList
		err = wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, 15*time.Minute, false, func(ctx context.Context) (done bool, err error) {
			list, err = vpaClientSet.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				klog.ErrorS(err, "Error listing VPA checkpoints")
				return false, err
			}
			if len(list.Items) > 0 {
				return false, nil
			}
			klog.InfoS("No VPA checkpoints found")
			return true, nil

		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

var _ = utils.RecommenderE2eDescribe("VPA CRD object", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	f.It("serves recommendation for CronJob", framework.WithSlow(), func() {
		ginkgo.By("Setting up hamster CronJob")
		SetupHamsterCronJob(f, "*/5 * * * *", "100m", "100Mi", utils.DefaultHamsterReplicas)

		vpaClientSet := utils.GetVpaClientSet(f)

		ginkgo.By("Setting up VPA")
		targetRef := &autoscaling.CrossVersionObjectReference{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Name:       "hamster-cronjob",
		}

		containerName := utils.GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(targetRef).
			WithContainer(containerName).
			Get()

		utils.InstallVPA(f, vpaCRD)

		ginkgo.By("Waiting for recommendation to be filled")
		_, err := utils.WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})
})

var _ = utils.RecommenderE2eDescribe("VPA CRD object", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	var (
		vpaCRD       *vpa_types.VerticalPodAutoscaler
		vpaClientSet vpa_clientset.Interface
	)

	ginkgo.BeforeEach(func() {
		ginkgo.By("Setting up a hamster deployment")
		_ = SetupHamsterDeployment(
			f,       /* framework */
			"100m",  /* cpu */
			"100Mi", /* memory */
			1,       /* number of replicas */
		)

		ginkgo.By("Setting up a VPA CRD")
		containerName := utils.GetHamsterContainerNameByIndex(0)
		vpaCRD = test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(test.HamsterTargetRef).
			WithContainer(containerName).
			Get()

		utils.InstallVPA(f, vpaCRD)

		vpaClientSet = utils.GetVpaClientSet(f)
	})

	ginkgo.It("serves recommendation", func() {
		ginkgo.By("Waiting for recommendation to be filled")
		_, err := utils.WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	// FIXME todo(adrianmoisey): This test seems to be flaky after running in parallel, unsure why, see if it's possible to fix
	f.It("doesn't drop lower/upper after recommender's restart", framework.WithSerial(), framework.WithSlow(), func() {

		o := getVpaObserver(vpaClientSet, f.Namespace.Name)

		ginkgo.By("Waiting for recommendation to be filled")
		_, err := utils.WaitForRecommendationPresent(vpaClientSet, vpaCRD)
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

var _ = utils.RecommenderE2eDescribe("VPA CRD object", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	var (
		vpaClientSet vpa_clientset.Interface
	)

	ginkgo.BeforeEach(func() {
		ginkgo.By("Setting up a hamster deployment")
		_ = SetupHamsterDeployment(
			f,       /* framework */
			"100m",  /* cpu */
			"100Mi", /* memory */
			1,       /* number of replicas */
		)

		vpaClientSet = utils.GetVpaClientSet(f)
	})

	ginkgo.It("respects min allowed recommendation", func() {
		const minMilliCpu = 10000
		ginkgo.By("Setting up a VPA CRD")
		containerName := utils.GetHamsterContainerNameByIndex(0)
		vpaCRD2 := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(test.HamsterTargetRef).
			WithContainer(containerName).
			WithMinAllowed(containerName, "10000", "").
			Get()

		utils.InstallVPA(f, vpaCRD2)
		vpaCRD := vpaCRD2

		ginkgo.By("Waiting for recommendation to be filled")
		vpa, err := utils.WaitForRecommendationPresent(vpaClientSet, vpaCRD)
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
		containerName := utils.GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(test.HamsterTargetRef).
			WithContainer(containerName).
			WithMaxAllowed(containerName, "1m", "").
			Get()

		utils.InstallVPA(f, vpaCRD)

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

var _ = utils.RecommenderE2eDescribe("VPA CRD object", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")
	f.NamespacePodSecurityLevel = podsecurity.LevelBaseline

	var vpaClientSet vpa_clientset.Interface

	ginkgo.BeforeEach(func() {
		vpaClientSet = utils.GetVpaClientSet(f)
	})

	ginkgo.It("with no containers opted out all containers get recommendations", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := utils.NewNHamstersDeployment(f, 2 /*number of containers*/)
		_ = utils.StartDeploymentPods(f, d)

		ginkgo.By("Setting up VPA CRD")
		container1Name := utils.GetHamsterContainerNameByIndex(0)
		container2Name := utils.GetHamsterContainerNameByIndex(1)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(test.HamsterTargetRef).
			WithContainer(container1Name).
			WithContainer(container2Name).
			Get()

		utils.InstallVPA(f, vpaCRD)

		ginkgo.By("Waiting for recommendation to be filled for both containers")
		vpa, err := utils.WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(2))
	})

	ginkgo.It("only containers not-opted-out get recommendations", func() {
		ginkgo.By("Setting up a hamster deployment")
		d := utils.NewNHamstersDeployment(f, 2 /*number of containers*/)
		_ = utils.StartDeploymentPods(f, d)

		ginkgo.By("Setting up VPA CRD")
		container1Name := utils.GetHamsterContainerNameByIndex(0)
		container2Name := utils.GetHamsterContainerNameByIndex(1)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(test.HamsterTargetRef).
			WithContainer(container1Name).
			WithScalingMode(container1Name, vpa_types.ContainerScalingModeOff).
			WithContainer(container2Name).
			Get()

		utils.InstallVPA(f, vpaCRD)

		ginkgo.By("Waiting for recommendation to be filled for just one container")
		vpa, err := utils.WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		errMsg := fmt.Sprintf("%s container has recommendations turned off. We expect expect only recommendations for %s",
			utils.GetHamsterContainerNameByIndex(0),
			utils.GetHamsterContainerNameByIndex(1))
		gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(1), errMsg)
		gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations[0].ContainerName).To(gomega.Equal(utils.GetHamsterContainerNameByIndex(1)), errMsg)
	})
	f.It("have memory requests growing with OOMs more than the default", framework.WithFeatureGate(features.PerVPAConfig), func() {
		const replicas = 1
		const defaultOOMBumpUpRatio = model.DefaultOOMBumpUpRatio
		const oomBumpUpRatio = 3

		ns := f.Namespace.Name
		vpaClientSet = utils.GetVpaClientSet(f)

		ginkgo.By("Setting up a hamster deployment")
		runOomingReplicationController(
			f.ClientSet,
			ns,
			"hamster",
			replicas)

		ginkgo.By("Setting up a VPA CRD")
		targetRef := &autoscaling.CrossVersionObjectReference{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "hamster",
		}
		containerName := utils.GetHamsterContainerNameByIndex(0)
		vpaCRD := test.VerticalPodAutoscaler().
			WithName("hamster-vpa").
			WithNamespace(f.Namespace.Name).
			WithTargetRef(targetRef).
			WithContainer(containerName).
			WithOOMBumpUpRatio(resource.NewQuantity(oomBumpUpRatio, resource.DecimalSI)).
			Get()
		utils.InstallVPA(f, vpaCRD)

		ginkgo.By("Waiting for recommendation to be filled")
		vpa, err := utils.WaitForRecommendationPresent(vpaClientSet, vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(vpa.Status.Recommendation.ContainerRecommendations).Should(gomega.HaveLen(1))

		currentMemory := vpa.Status.Recommendation.ContainerRecommendations[0].Target.Memory().Value()
		oomReplicationControllerRequestLimit := int64(1024 * 1024 * 1024)                          // from runOomingReplicationController
		defaultBumpMemory := float64(oomReplicationControllerRequestLimit) * defaultOOMBumpUpRatio // DefaultOOMBumpUpRatio
		customBumpMemory := float64(oomReplicationControllerRequestLimit) * oomBumpUpRatio         // Custom ratio from VPA config

		// Sanity check: verify that our custom bump ratio is indeed higher than the default
		gomega.Expect(customBumpMemory).Should(gomega.BeNumerically(">", defaultBumpMemory),
			fmt.Sprintf("Custom OOM bump ratio (%fx) should be greater than default (%fx)", float64(oomBumpUpRatio), defaultOOMBumpUpRatio))

		// Verify that the actual recommendation is at least the custom bump ratio
		gomega.Expect(currentMemory).Should(gomega.BeNumerically(">=", int64(customBumpMemory)),
			fmt.Sprintf("Memory recommendation should be at least custom bump up ratio (%dx). Got: %d bytes (%.2fx), Expected: >= %d bytes (%dx)",
				oomBumpUpRatio,
				currentMemory,
				float64(currentMemory)/float64(oomReplicationControllerRequestLimit),
				int64(customBumpMemory),
				oomBumpUpRatio))
	})
})

func deleteRecommender(c clientset.Interface) error {
	namespace := utils.VpaNamespace
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
