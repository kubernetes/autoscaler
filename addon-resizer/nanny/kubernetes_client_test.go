/*
Copyright 2026 The Kubernetes Authors.

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

package nanny

import (
	"testing"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

func newTestPodLister(pods ...*core.Pod) v1lister.PodNamespaceLister {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, pod := range pods {
		store.Add(pod)
	}
	return v1lister.NewPodLister(store).Pods("ns")
}

func TestContainerResourcesFindsNamedContainer(t *testing.T) {
	want := core.ResourceRequirements{
		Requests: core.ResourceList{"cpu": resource.MustParse("100m")},
	}
	pod := &core.Pod{
		ObjectMeta: v1.ObjectMeta{Name: "mypod", Namespace: "ns"},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{Name: "sidecar"},
				{Name: "target", Resources: want},
			},
		},
	}
	k := &kubernetesClient{podLister: newTestPodLister(pod), pod: "mypod", container: "target"}

	got, err := k.ContainerResources()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Requests["cpu"] != want.Requests["cpu"] {
		t.Errorf("got %v, want %v", got.Requests["cpu"], want.Requests["cpu"])
	}
}

func TestContainerResourcesMissingContainer(t *testing.T) {
	pod := &core.Pod{
		ObjectMeta: v1.ObjectMeta{Name: "mypod", Namespace: "ns"},
		Spec:       core.PodSpec{Containers: []core.Container{{Name: "other"}}},
	}
	k := &kubernetesClient{podLister: newTestPodLister(pod), pod: "mypod", container: "target", deployment: "mydep", namespace: "ns"}

	if _, err := k.ContainerResources(); err == nil {
		t.Fatal("expected an error for a missing container, got nil")
	}
}

func TestContainerResourcesMissingPod(t *testing.T) {
	k := &kubernetesClient{podLister: newTestPodLister(), pod: "mypod", container: "target"}

	if _, err := k.ContainerResources(); err == nil {
		t.Fatal("expected an error for a missing pod, got nil")
	}
}
