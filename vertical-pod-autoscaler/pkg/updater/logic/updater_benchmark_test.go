/*
Copyright The Kubernetes Authors.

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

package logic

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/priority"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/restriction"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// countingSelectorFetcher returns the same selector for every VPA and counts the
// calls. Fetching is free here but may be an API call per VPA in production, so
// benchmarks report the count alongside the timings.
type countingSelectorFetcher struct {
	selector labels.Selector
	fetches  int
}

func (f *countingSelectorFetcher) Fetch(_ context.Context, _ *vpa_types.VerticalPodAutoscaler) (labels.Selector, error) {
	f.fetches++
	return f.selector, nil
}

// countingEvictionRestriction never allows eviction, so RunOnce evicts nothing and never
// reaches the rate limiters. It counts CanEvict calls, which RunOnce makes exactly once per
// controlled pod.
type countingEvictionRestriction struct {
	canEvictCalls int
}

func (*countingEvictionRestriction) Evict(_ *corev1.Pod, _ *vpa_types.VerticalPodAutoscaler, _ record.EventRecorder) error {
	return nil
}

func (r *countingEvictionRestriction) CanEvict(_ *corev1.Pod) bool {
	r.canEvictCalls++
	return false
}

// countingInPlaceRestriction defers every pod. Under InPlaceOrRecreate, the only in-place mode
// the fixture uses, filterNonInPlaceUpdatablePods drops deferred pods, so RunOnce calls
// CanInPlaceUpdate exactly once per controlled pod and never reaches the in-place update loop.
// Under InPlace mode deferred pods would be kept and checked a second time there.
type countingInPlaceRestriction struct {
	canInPlaceCalls int
}

func (*countingInPlaceRestriction) InPlaceUpdate(_ *corev1.Pod, _ *vpa_types.VerticalPodAutoscaler, _ record.EventRecorder) error {
	return nil
}

func (r *countingInPlaceRestriction) CanInPlaceUpdate(_ *corev1.Pod, _ *vpa_types.VerticalPodAutoscaler, _ map[types.UID]*vpa_types.RecommendedPodResources) utils.InPlaceDecision {
	r.canInPlaceCalls++
	return utils.InPlaceDeferred
}

func (*countingInPlaceRestriction) CanUnboost(_ *corev1.Pod, _ *vpa_types.VerticalPodAutoscaler) bool {
	return false
}

// setupUpdaterBenchmark builds an updater over vpaCount VPAs and podCount pods in one
// namespace. Every VPA targets a distinct Deployment and pods are spread evenly across
// those Deployments, so every pod is controlled by exactly one VPA.
func setupUpdaterBenchmark(b *testing.B, vpaCount, podCount int) (*updater, *countingEvictionRestriction, *countingInPlaceRestriction, *countingSelectorFetcher) {
	b.Helper()

	podLabels := map[string]string{"app": "bench"}
	vpaIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, vpa_api_util.VPAIndexers())
	podIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	workloads := make([]appsv1.Deployment, vpaCount)
	updateModes := []vpa_types.UpdateMode{vpa_types.UpdateModeRecreate, vpa_types.UpdateModeInPlaceOrRecreate}
	for i := range vpaCount {
		workloads[i] = appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("workload-%d", i),
				Namespace: "default",
			},
		}
		vpa := test.VerticalPodAutoscaler().
			WithName(fmt.Sprintf("vpa-%d", i)).
			WithContainer("bench-container").
			WithUpdateMode(updateModes[i%len(updateModes)]).
			WithTargetRef(&autoscalingv1.CrossVersionObjectReference{
				Kind:       workloads[i].Kind,
				Name:       workloads[i].Name,
				APIVersion: workloads[i].APIVersion,
			}).
			Get()
		if err := vpaIndexer.Add(vpa); err != nil {
			b.Fatal(err)
		}
	}

	for i := range podCount {
		owner := &workloads[i%vpaCount]
		name := fmt.Sprintf("pod-%d", i)
		pod := test.Pod().
			WithName(name).
			WithUID(types.UID(name)).
			WithLabels(podLabels).
			AddContainer(test.Container().WithName("bench-container").Get()).
			WithCreator(&owner.ObjectMeta, &owner.TypeMeta).
			Get()
		if err := podIndexer.Add(pod); err != nil {
			b.Fatal(err)
		}
	}

	eviction := &countingEvictionRestriction{}
	inplace := &countingInPlaceRestriction{}
	fetcher := &countingSelectorFetcher{selector: parseLabelSelector("app = bench")}
	return &updater{
		vpaLister: vpa_lister.NewVerticalPodAutoscalerLister(vpaIndexer),
		podLister: listersv1.NewPodLister(podIndexer),
		restrictionFactory: &restriction.FakePodsRestrictionFactory{
			Eviction: eviction,
			InPlace:  inplace,
		},
		evictionRateLimiter:          rate.NewLimiter(rate.Inf, 0),
		inPlaceRateLimiter:           rate.NewLimiter(rate.Inf, 0),
		evictionAdmission:            priority.NewDefaultPodEvictionAdmission(),
		recommendationProcessor:      &test.FakeRecommendationProcessor{},
		selectorFetcher:              fetcher,
		controllerFetcher:            controllerfetcher.FakeControllerFetcher{},
		useAdmissionControllerStatus: false,
		priorityProcessor:            priority.NewProcessor(),
	}, eviction, inplace, fetcher
}

// BenchmarkRunOnce measures RunOnce with its API-backed collaborators faked out: pods are
// owned directly by their Deployment so the controller fetcher does no lookups, creator maps
// are empty, and every pod is denied so priority calculation sees no pods. What remains is
// listing VPAs, fetching selectors and matching each live pod to its controlling VPA, the
// last of which is expected to dominate as the number of VPAs grows.
func BenchmarkRunOnce(b *testing.B) {
	ctx := context.Background()

	for _, tc := range []struct {
		vpas int
		pods int
	}{
		{vpas: 1, pods: 1000},
		{vpas: 10, pods: 1000},
		{vpas: 100, pods: 1000},
		{vpas: 1000, pods: 1000},
		{vpas: 10000, pods: 1000},
		{vpas: 1, pods: 10000},
		{vpas: 10, pods: 10000},
		{vpas: 100, pods: 10000},
		{vpas: 1000, pods: 10000},
		{vpas: 10000, pods: 10000},
	} {
		b.Run(fmt.Sprintf("vpas=%d/pods=%d", tc.vpas, tc.pods), func(b *testing.B) {
			u, eviction, inplace, fetcher := setupUpdaterBenchmark(b, tc.vpas, tc.pods)
			b.ReportAllocs()
			u.RunOnce(ctx)
			if eviction.canEvictCalls+inplace.canInPlaceCalls != tc.pods {
				b.Fatalf("expected all %d pods to be matched to a VPA, got %d (eviction: %d, inPlace: %d)", tc.pods, eviction.canEvictCalls+inplace.canInPlaceCalls, eviction.canEvictCalls, inplace.canInPlaceCalls)
			}
			fetcher.fetches = 0
			for b.Loop() {
				u.RunOnce(ctx)
			}
			b.ReportMetric(float64(fetcher.fetches)/float64(b.N), "fetches/op")
		})
	}
}

// BenchmarkBoostWorkerVPALookup measures looking up a single pod's controlling VPA,
// which the startup boost worker does for every pod it processes.
func BenchmarkBoostWorkerVPALookup(b *testing.B) {
	ctx := context.Background()

	for _, vpaCount := range []int{1, 10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("vpas=%d", vpaCount), func(b *testing.B) {
			u, _, _, fetcher := setupUpdaterBenchmark(b, vpaCount, 1)
			pods, err := u.podLister.List(labels.Everything())
			if err != nil || len(pods) != 1 {
				b.Fatalf("expected exactly one pod, got %d (%v)", len(pods), err)
			}
			pod := pods[0]

			b.ReportAllocs()
			// A fixture whose pod matches no VPA would silently benchmark the cheap no-match path.
			got, err := u.getControllingVPAForPod(ctx, pod)
			if err != nil || got == nil {
				b.Fatalf("expected the pod to be controlled by vpa-0, got %v (err: %v)", got, err)
			}
			if got.Vpa.Name != "vpa-0" {
				b.Fatalf("expected the pod to be controlled by vpa-0, got %s", got.Vpa.Name)
			}
			fetcher.fetches = 0
			for b.Loop() {
				_, _ = u.getControllingVPAForPod(ctx, pod)
			}
			b.ReportMetric(float64(fetcher.fetches)/float64(b.N), "fetches/op")
		})
	}
}
