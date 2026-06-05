/*
Copyright 2025 The Kubernetes Authors.

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

package controller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	fakeK8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/clock"
	buffersfake "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned/fake"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/updater"
)

func TestEnsureWatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := scheme.Scheme
	gvk := schema.GroupVersionKind{Group: "custom.io", Version: "v1", Kind: "CustomType"}
	gvr := schema.GroupVersionResource{Group: "custom.io", Version: "v1", Resource: "customtypes"}

	fakeDynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(s, map[schema.GroupVersionResource]string{
		gvr: "CustomTypeList",
	})

	// Create a mock RESTMapper
	mockMapper := &mockRESTMapper{
		mappings: map[schema.GroupKind]*schema.GroupVersionResource{
			{Group: "custom.io", Kind: "CustomType"}: &gvr,
		},
	}

	fakeCBClient, _ := cbclient.NewCapacityBufferClientFromClients(nil, nil, fakeDynamicClient, nil, mockMapper)

	rbacUpdater := &mockRBACUpdater{}

	bc := &bufferController{
		client:                 fakeCBClient,
		rbacUpdater:            rbacUpdater,
		dynamicClient:          fakeDynamicClient,
		dynamicInformerFactory: dynamicinformer.NewDynamicSharedInformerFactory(fakeDynamicClient, 0),
		stopCh:                 ctx.Done(),
		clock:                  clock.RealClock{},
	}

	gvk = schema.GroupVersionKind{Group: "custom.io", Version: "v1", Kind: "CustomType"}
	mapping := &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "custom.io", Version: "v1", Resource: "customtypes"},
		GroupVersionKind: gvk,
	}

	// First call should establish watch
	bc.ensureWatch(mapping)

	// Since establishWatchWithRetry runs in a goroutine, we need to wait
	assert.Eventually(t, func() bool {
		return len(rbacUpdater.updatedMappings) == 1
	}, 5*time.Second, 100*time.Millisecond)

	assert.Equal(t, gvk, rbacUpdater.updatedMappings[0].GroupVersionKind)

	// Second call should be a no-op (cached)
	bc.ensureWatch(mapping)
	// Give it a bit of time to make sure no extra call was made
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, 1, len(rbacUpdater.updatedMappings))
	_, loaded := bc.watchedGVKs.Load(gvk)
	assert.True(t, loaded)
}

func TestMarkBuffersAsNonEventDriven(t *testing.T) {
	gvk := schema.GroupVersionKind{Group: "custom.io", Version: "v1", Kind: "CustomType"}
	gvr := schema.GroupVersionResource{Group: "custom.io", Version: "v1", Resource: "customtypes"}

	buffer := &v1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-buffer",
			Namespace: "default",
		},
		Spec: v1.CapacityBufferSpec{
			ScalableRef: &v1.ScalableRef{
				APIGroup: gvk.Group,
				Kind:     gvk.Kind,
				Name:     "target",
			},
		},
	}

	fakeBuffersClient := buffersfake.NewSimpleClientset(buffer)
	fakeK8sClient := fakeK8s.NewSimpleClientset()
	fakeDynamicClient := fake.NewSimpleDynamicClient(scheme.Scheme)
	fakeCBClient, err := cbclient.NewCapacityBufferClientFromClients(fakeBuffersClient, fakeK8sClient, fakeDynamicClient, nil, nil)
	assert.NoError(t, err)
	
	statusUpdater := updater.NewStatusUpdater(fakeCBClient)
	
	bc := &bufferController{
		client:  fakeCBClient,
		updater: *statusUpdater,
		clock:   clock.RealClock{},
	}

	bc.markBuffersAsNonEventDriven(gvk, gvr)

	// Verify status via client
	updated, err := fakeBuffersClient.AutoscalingV1beta1().CapacityBuffers("default").Get(context.TODO(), "test-buffer", metav1.GetOptions{})
	assert.NoError(t, err)
	
	found := false
	for _, cond := range updated.Status.Conditions {
		if cond.Type == EventDrivenReconciliationCondition {
			assert.Equal(t, metav1.ConditionFalse, cond.Status)
			assert.Equal(t, DynamicWatchFailedReason, cond.Reason)
			assert.Contains(t, cond.Message, "Failed to establish dynamic watch")
			found = true
			break
		}
	}
	assert.True(t, found)
}

type mockRESTMapper struct {
	mappings map[schema.GroupKind]*schema.GroupVersionResource
}

func (m *mockRESTMapper) KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	return schema.GroupVersionKind{}, nil
}

func (m *mockRESTMapper) KindsFor(resource schema.GroupVersionResource) ([]schema.GroupVersionKind, error) {
	return nil, nil
}

func (m *mockRESTMapper) ResourceFor(input schema.GroupVersionResource) (schema.GroupVersionResource, error) {
	return schema.GroupVersionResource{}, nil
}

func (m *mockRESTMapper) ResourcesFor(input schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	return nil, nil
}

func (m *mockRESTMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	gvr, ok := m.mappings[gk]
	if !ok {
		return nil, nil
	}
	return &meta.RESTMapping{
		Resource:         *gvr,
		GroupVersionKind: gk.WithVersion(gvr.Version),
	}, nil
}

func (m *mockRESTMapper) RESTMappings(gk schema.GroupKind, versions ...string) ([]*meta.RESTMapping, error) {
	return nil, nil
}

func (m *mockRESTMapper) ResourceSingularizer(resource string) (singular string, err error) {
	return "", nil
}
