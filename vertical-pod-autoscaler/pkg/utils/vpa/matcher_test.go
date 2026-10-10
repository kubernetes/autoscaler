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

package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/ptr"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

// annotationSelectorFetcher returns the selector stored in the VPA's "selector"
// annotation, or an error if the annotation is "error". It records fetch counts.
type annotationSelectorFetcher struct {
	fetches map[string]int
}

func (f *annotationSelectorFetcher) Fetch(_ context.Context, vpa *vpa_types.VerticalPodAutoscaler) (labels.Selector, error) {
	f.fetches[vpa.Name]++
	selector := vpa.Annotations["selector"]
	if selector == "error" {
		return nil, errors.New("cannot fetch selector")
	}
	return parseLabelSelector(selector), nil
}

func newTestVPAIndexer(t *testing.T, vpas ...*vpa_types.VerticalPodAutoscaler) cache.Indexer {
	t.Helper()
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{TargetRefIndex: TargetRefIndexFunc})
	for _, vpa := range vpas {
		require.NoError(t, indexer.Add(vpa))
	}
	return indexer
}

func TestControllingVPAFinder(t *testing.T) {
	ctx := context.Background()

	newPod := func(namespace, ownerName string) *corev1.Pod {
		pod := test.Pod().WithName("pod-" + ownerName).WithLabels(map[string]string{"app": "testingApp"}).Get()
		pod.Namespace = namespace
		pod.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: "apps/v1",
				Kind:       "StatefulSet",
				Name:       ownerName,
				Controller: ptr.To(true),
			},
		}
		return pod
	}
	newVPA := func(name, namespace, targetName, selector string, created time.Time, updateMode vpa_types.UpdateMode) *vpa_types.VerticalPodAutoscaler {
		return test.VerticalPodAutoscaler().
			WithName(name).
			WithNamespace(namespace).
			WithContainer(containerName).
			WithCreationTimestamp(created).
			WithUpdateMode(updateMode).
			WithAnnotations(map[string]string{"selector": selector}).
			WithTargetRef(&autoscalingv1.CrossVersionObjectReference{
				Kind:       "StatefulSet",
				Name:       targetName,
				APIVersion: "apps/v1",
			}).
			Get()
	}

	vpaStsA := newVPA("vpa-sts-a", "ns1", "sts-a", "app = testingApp", time.Unix(10, 0), vpa_types.UpdateModeRecreate)
	vpaStsAOlder := newVPA("vpa-sts-a-older", "ns1", "sts-a", "app = testingApp", time.Unix(5, 0), vpa_types.UpdateModeRecreate)
	vpaStsAOldestOff := newVPA("vpa-sts-a-oldest-off", "ns1", "sts-a", "app = testingApp", time.Unix(1, 0), vpa_types.UpdateModeOff)
	vpaStsANonMatching := newVPA("vpa-sts-a-other", "ns1", "sts-a", "app = other", time.Unix(1, 0), vpa_types.UpdateModeRecreate)
	vpaStsAUnfetchable := newVPA("vpa-sts-a-unfetchable", "ns1", "sts-a", "error", time.Unix(1, 0), vpa_types.UpdateModeRecreate)
	vpaStsB := newVPA("vpa-sts-b", "ns1", "sts-b", "app = testingApp", time.Unix(1, 0), vpa_types.UpdateModeRecreate)
	vpaStsAOtherNamespace := newVPA("vpa-sts-a", "ns2", "sts-a", "app = testingApp", time.Unix(1, 0), vpa_types.UpdateModeRecreate)
	vpaNoTargetRef := newVPA("no-target-ref", "ns1", "sts-c", "app = testingApp", time.Unix(1, 0), vpa_types.UpdateModeRecreate)
	vpaNoTargetRef.Spec.TargetRef = nil

	indexer := newTestVPAIndexer(t,
		vpaStsA, vpaStsAOlder, vpaStsAOldestOff, vpaStsANonMatching, vpaStsAUnfetchable,
		vpaStsB, vpaStsAOtherNamespace, vpaNoTargetRef)
	notOff := func(vpa *vpa_types.VerticalPodAutoscaler) bool {
		return GetUpdateMode(vpa) != vpa_types.UpdateModeOff
	}

	for _, tc := range []struct {
		name          string
		pod           *corev1.Pod
		filter        VPAFilter
		ctrlFetcher   controllerfetcher.ControllerFetcher
		want          *vpa_types.VerticalPodAutoscaler
		wantSelector  string
		wantNoFetches []string
	}{
		{
			name:         "strongest matching VPA for target",
			pod:          newPod("ns1", "sts-a"),
			want:         vpaStsAOldestOff,
			wantSelector: "app=testingApp",
		},
		{
			name:         "filter excludes strongest VPA",
			pod:          newPod("ns1", "sts-a"),
			filter:       notOff,
			want:         vpaStsAOlder,
			wantSelector: "app=testingApp",
			// Filtered-out VPAs must not have their selector fetched.
			wantNoFetches: []string{"vpa-sts-a-oldest-off"},
		},
		{
			name:         "other target in same namespace",
			pod:          newPod("ns1", "sts-b"),
			want:         vpaStsB,
			wantSelector: "app=testingApp",
		},
		{
			name:         "same target name in other namespace",
			pod:          newPod("ns2", "sts-a"),
			want:         vpaStsAOtherNamespace,
			wantSelector: "app=testingApp",
		},
		{
			name: "no VPA for target",
			pod:  newPod("ns1", "sts-c"),
			want: nil,
		},
		{
			name: "pod without controller",
			pod:  test.Pod().WithName("orphan").Get(),
			want: nil,
		},
		{
			// For some Pods (which are *not* under VPA), controllerFetcher.FindTopMostWellKnownOrScalable returns nil,
			// e.g. when the Pod owner is a custom resource which doesn't implement the /scale subresource.
			name:        "parent controller not resolvable",
			pod:         newPod("ns1", "sts-a"),
			ctrlFetcher: &controllerfetcher.NilControllerFetcher{},
			want:        nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrlFetcher := tc.ctrlFetcher
			if ctrlFetcher == nil {
				ctrlFetcher = &controllerfetcher.FakeControllerFetcher{}
			}
			selectorFetcher := &annotationSelectorFetcher{fetches: map[string]int{}}
			finder := NewControllingVPAFinder(indexer, selectorFetcher, ctrlFetcher, tc.filter)

			gotVPA, gotSelector := finder.GetControllingVPAForPod(ctx, tc.pod)
			if tc.want == nil {
				assert.Nil(t, gotVPA)
				assert.Nil(t, gotSelector)
				return
			}
			assert.Equal(t, tc.want, gotVPA)
			require.NotNil(t, gotSelector)
			assert.Equal(t, tc.wantSelector, gotSelector.String())
			for _, name := range tc.wantNoFetches {
				assert.Zero(t, selectorFetcher.fetches[name], "selector of %s should not have been fetched", name)
			}
		})
	}
}
