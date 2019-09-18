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
	"fmt"
	"strings"
	"time"

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

func getVpaObserver(vpaClientSet *vpa_clientset.Clientset) *observer {
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
		config, err := framework.LoadConfig()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		checkpoint := vpa_types.VerticalPodAutoscalerCheckpoint{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: ns,
			},
			Spec: vpa_types.VerticalPodAutoscalerCheckpointSpec{
				VPAObjectName: "some-vpa",
			},
		}

		vpaClientSet := vpa_clientset.NewForConfigOrDie(config)
		_, err = vpaClientSet.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns).Create(&checkpoint)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		time.Sleep(15 * time.Minute)

		list, err := vpaClientSet.AutoscalingV1().VerticalPodAutoscalerCheckpoints(ns).List(metav1.ListOptions{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(list.Items).To(gomega.BeEmpty())
	})
})

var _ = RecommenderE2eDescribe("VPA CRD object", func() {
	f := framework.NewDefaultFramework("vertical-pod-autoscaling")

	var (
		vpaCRD       *vpa_types.VerticalPodAutoscaler
		vpaClientSet *vpa_clientset.Clientset
	)

	ginkgo.BeforeEach(func() {
		ginkgo.By("Setting up a hamster deployment")
		c := f.ClientSet
		ns := f.Namespace.Name

		cpuQuantity := ParseQuantityOrDie("100m")
		memoryQuantity := ParseQuantityOrDie("100Mi")

		d := NewHamsterDeploymentWithResources(f, cpuQuantity, memoryQuantity)
		_, err := c.AppsV1().Deployments(ns).Create(d)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		err = framework.WaitForDeploymentComplete(c, d)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("Setting up a VPA CRD")
		config, err := framework.LoadConfig()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		vpaCRD = NewVPA(f, "hamster-vpa", hamsterTargetRef)

		vpaClientSet = vpa_clientset.NewForConfigOrDie(config)
		vpaClient := vpaClientSet.AutoscalingV1()
		_, err = vpaClient.VerticalPodAutoscalers(ns).Create(vpaCRD)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
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

func deleteRecommender(c clientset.Interface) error {
	namespace := "kube-system"
	listOptions := metav1.ListOptions{}
	podList, err := c.CoreV1().Pods(namespace).List(listOptions)
	if err != nil {
		fmt.Println("Could not list pods.", err)
		return err
	}
	fmt.Print("Pods list items:", len(podList.Items))
	for _, pod := range podList.Items {
		if strings.HasPrefix(pod.Name, "vpa-recommender") {
			fmt.Print("Deleting pod.", namespace, pod.Name)
			err := c.CoreV1().Pods(namespace).Delete(pod.Name, &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("vpa recommender not found")
}
