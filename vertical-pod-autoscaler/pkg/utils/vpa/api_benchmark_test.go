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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
)

// BenchmarkFindParentControllerForPod measures resolving a pod's controller ownerReference
// into the key handed to the ControllerFetcher. The fetcher is faked out, so this is the
// fixed per-pod cost the updater and admission controller pay before any controller lookup.
func BenchmarkFindParentControllerForPod(b *testing.B) {
	ctx := context.Background()
	const ownerName = "workload-7c9d8f6b5"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ownerName + "-x2k9q",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "apps/v1",
				Kind:       "ReplicaSet",
				Name:       ownerName,
				Controller: ptr.To(true),
			}},
		},
	}
	var fetcher controllerfetcher.ControllerFetcher = controllerfetcher.FakeControllerFetcher{}

	b.ReportAllocs()
	got, err := FindParentControllerForPod(ctx, pod, fetcher)
	if err != nil || got.Name != ownerName {
		b.Fatalf("expected the pod's parent controller to be %s, got %v (err: %v)", ownerName, got, err)
	}
	for b.Loop() {
		_, _ = FindParentControllerForPod(ctx, pod, fetcher)
	}
}
