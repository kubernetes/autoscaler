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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/recommendation"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/vpa"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limitrange"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// staticSelectorFetcher returns the same selector for every VPA, standing in
// for the informer-backed target selector fetcher.
type staticSelectorFetcher struct {
	selector labels.Selector
}

func (f *staticSelectorFetcher) Fetch(_ context.Context, _ *vpa_types.VerticalPodAutoscaler) (labels.Selector, error) {
	return f.selector, nil
}

// newBenchmarkAdmissionServer returns an AdmissionServer wired with the same
// handler chain as production (see pkg/admission-controller/main.go), backed
// by a lister with vpaCount VPA objects, together with a serialized
// AdmissionReview request for a pod with containerCount containers that
// matches exactly one of the VPAs.
func newBenchmarkAdmissionServer(b *testing.B, containerCount, vpaCount int) (*AdmissionServer, []byte) {
	b.Helper()

	// A StatefulSet directly controls its pods, so the FakeControllerFetcher
	// (which resolves a controller to itself) behaves like the real
	// controller fetcher for this owner hierarchy.
	sts := appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bench-sts",
			Namespace: "default",
		},
	}
	targetRef := &autoscalingv1.CrossVersionObjectReference{
		Kind:       sts.Kind,
		Name:       sts.Name,
		APIVersion: sts.APIVersion,
	}

	podLabels := map[string]string{"app": "bench"}
	podBuilder := test.Pod().WithName("bench-pod").WithLabels(podLabels).WithCreator(&sts.ObjectMeta, &sts.TypeMeta)
	vpaBuilder := test.VerticalPodAutoscaler().WithName("bench-vpa").WithTargetRef(targetRef)
	for i := range containerCount {
		containerName := fmt.Sprintf("container-%d", i)
		podBuilder = podBuilder.AddContainer(test.Container().WithName(containerName).
			WithCPURequest(resource.MustParse("100m")).
			WithMemRequest(resource.MustParse("100Mi")).
			WithCPULimit(resource.MustParse("200m")).
			WithMemLimit(resource.MustParse("200Mi")).
			Get())
		vpaBuilder = vpaBuilder.WithContainer(containerName).
			AppendRecommendation(test.Recommendation().WithContainer(containerName).WithTarget("150m", "150Mi").GetContainerResources())
	}

	// One matching VPA plus vpaCount-1 VPAs targeting other controllers, all
	// of which the matcher has to consider.
	vpas := []*vpa_types.VerticalPodAutoscaler{vpaBuilder.Get()}
	for i := 1; i < vpaCount; i++ {
		vpas = append(vpas, test.VerticalPodAutoscaler().
			WithName(fmt.Sprintf("other-vpa-%d", i)).
			WithContainer("container-0").
			WithTargetRef(&autoscalingv1.CrossVersionObjectReference{
				Kind:       "StatefulSet",
				Name:       fmt.Sprintf("other-sts-%d", i),
				APIVersion: "apps/v1",
			}).
			Get())
	}
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, vpaObject := range vpas {
		if err := indexer.Add(vpaObject); err != nil {
			b.Fatalf("failed to add VPA to indexer: %v", err)
		}
	}
	vpaLister := vpa_lister.NewVerticalPodAutoscalerLister(indexer)

	vpaMatcher := vpa.NewMatcher(vpaLister, &staticSelectorFetcher{selector: labels.SelectorFromSet(podLabels)}, controllerfetcher.FakeControllerFetcher{})
	limitRangeCalculator := limitrange.NewNoopLimitsCalculator()
	recommendationProvider := recommendation.NewProvider(limitRangeCalculator, vpa_api_util.NewCappingRecommendationProcessor(limitRangeCalculator))
	calculators := []patch.Calculator{patch.NewResourceUpdatesCalculator(recommendationProvider, resource.QuantityValue{}), patch.NewObservedContainersCalculator()}
	server := NewAdmissionServer(pod.NewDefaultPreProcessor(), vpa.NewDefaultPreProcessor(), limitRangeCalculator, vpaMatcher, calculators)

	podJSON, err := json.Marshal(podBuilder.Get())
	if err != nil {
		b.Fatalf("failed to marshal pod: %v", err)
	}
	review := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Request: &admissionv1.AdmissionRequest{
			UID:         types.UID("bench-uid"),
			Resource:    metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			RequestKind: &metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			Namespace:   "default",
			Operation:   admissionv1.Create,
			Object:      runtime.RawExtension{Raw: podJSON},
		},
	}
	body, err := json.Marshal(review)
	if err != nil {
		b.Fatalf("failed to marshal admission review: %v", err)
	}
	return server, body
}

func serveBenchmarkRequest(server *AdmissionServer, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	server.Serve(recorder, request)
	return recorder
}

// BenchmarkAdmissionServerServe measures the per-request cost of the pod
// admission webhook: reading the request, unmarshalling the AdmissionReview,
// matching the pod to a VPA, calculating resource patches and marshalling the
// response.
func BenchmarkAdmissionServerServe(b *testing.B) {
	for _, tc := range []struct {
		containerCount int
		vpaCount       int
	}{
		{containerCount: 1, vpaCount: 1},
		{containerCount: 1, vpaCount: 100},
		{containerCount: 10, vpaCount: 1},
		{containerCount: 10, vpaCount: 100},
	} {
		b.Run(fmt.Sprintf("containers=%d/vpas=%d", tc.containerCount, tc.vpaCount), func(b *testing.B) {
			server, body := newBenchmarkAdmissionServer(b, tc.containerCount, tc.vpaCount)

			// Sanity-check that the request produces patches, so the
			// benchmark doesn't silently measure a no-op path.
			recorder := serveBenchmarkRequest(server, body)
			if recorder.Code != http.StatusOK {
				b.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
			}
			review := admissionv1.AdmissionReview{}
			if err := json.Unmarshal(recorder.Body.Bytes(), &review); err != nil {
				b.Fatalf("failed to unmarshal admission response: %v", err)
			}
			if review.Response == nil || len(review.Response.Patch) == 0 {
				b.Fatalf("expected an admission response with patches, got %s", recorder.Body.String())
			}

			b.ReportAllocs()
			for b.Loop() {
				serveBenchmarkRequest(server, body)
			}
		})
	}
}
