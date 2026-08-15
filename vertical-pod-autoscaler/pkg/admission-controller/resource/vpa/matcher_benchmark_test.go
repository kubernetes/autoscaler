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

package vpa

import (
	"context"
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

type fixedSelectorFetcher struct {
	selector labels.Selector
}

func (f *fixedSelectorFetcher) Fetch(_ context.Context, _ *vpa_types.VerticalPodAutoscaler) (labels.Selector, error) {
	return f.selector, nil
}

// setupMatcherBenchmark populates a cache with vpaCount VPAs in the pod's namespace, each
// targeting a distinct workload. Exactly one VPA targets the pod's owner; the matcher still
// scans all of them.
func setupMatcherBenchmark(b *testing.B, vpaCount int) (Matcher, *corev1.Pod) {
	b.Helper()

	owner := appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("workload-%d", vpaCount-1),
			Namespace: "default",
		},
	}
	pod := test.Pod().WithName("bench-pod").
		WithLabels(map[string]string{"app": "test"}).
		AddContainer(test.Container().WithName("bench-container").Get()).
		WithCreator(&owner.ObjectMeta, &owner.TypeMeta).
		Get()

	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
	})

	for i := range vpaCount {
		vpa := test.VerticalPodAutoscaler().
			WithContainer("bench-container").
			WithName(fmt.Sprintf("vpa-%d", i)).
			WithUpdateMode(vpa_types.UpdateModeRecreate).
			WithTargetRef(&autoscalingv1.CrossVersionObjectReference{
				Kind:       "StatefulSet",
				Name:       fmt.Sprintf("workload-%d", i),
				APIVersion: "apps/v1",
			}).
			Get()
		if err := indexer.Add(vpa); err != nil {
			b.Fatal(err)
		}
	}

	selectorFetcher := &fixedSelectorFetcher{selector: parseLabelSelector("app = test")}
	matcher := NewMatcher(vpa_lister.NewVerticalPodAutoscalerLister(indexer), selectorFetcher, controllerfetcher.FakeControllerFetcher{})
	return matcher, pod
}

func BenchmarkGetMatchingVPA(b *testing.B) {
	ctx := context.Background()

	for _, vpaCount := range []int{1, 10, 100, 1000, 10000} {
		matcher, pod := setupMatcherBenchmark(b, vpaCount)
		expectedVPAName := fmt.Sprintf("vpa-%d", vpaCount-1)

		b.Run(fmt.Sprintf("vpas=%d", vpaCount), func(b *testing.B) {
			b.ReportAllocs()
			got := matcher.GetMatchingVPA(ctx, pod)
			if got == nil || got.Name != expectedVPAName {
				b.Fatal("expected a matching VPA")
			}
			for b.Loop() {
				matcher.GetMatchingVPA(ctx, pod)
			}
		})
	}
}
