/*
Copyright 2024 The Kubernetes Authors.

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

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	claim1 = &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "claim-1", UID: "claim-1"}}
	claim2 = &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "claim-2", UID: "claim-2"}}
	claim3 = &resourceapi.ResourceClaim{ObjectMeta: metav1.ObjectMeta{Name: "claim-3", UID: "claim-3"}}

	localSlice1  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-1", UID: "local-slice-1"}, Spec: resourceapi.ResourceSliceSpec{NodeName: "n1"}}
	localSlice2  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-2", UID: "local-slice-2"}, Spec: resourceapi.ResourceSliceSpec{NodeName: "n1"}}
	localSlice3  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-3", UID: "local-slice-3"}, Spec: resourceapi.ResourceSliceSpec{NodeName: "n2"}}
	localSlice4  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-4", UID: "local-slice-4"}, Spec: resourceapi.ResourceSliceSpec{NodeName: "n2"}}
	globalSlice1 = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "global-slice-1", UID: "global-slice-1"}, Spec: resourceapi.ResourceSliceSpec{AllNodes: true}}
	globalSlice2 = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "global-slice-2", UID: "global-slice-2"}, Spec: resourceapi.ResourceSliceSpec{AllNodes: true}}
	globalSlice3 = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "global-slice-3", UID: "global-slice-3"}, Spec: resourceapi.ResourceSliceSpec{NodeSelector: &apiv1.NodeSelector{}}}

	class1 = &resourceapi.DeviceClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1", UID: "class-1"}}
	class2 = &resourceapi.DeviceClass{ObjectMeta: metav1.ObjectMeta{Name: "class-2", UID: "class-2"}}
	class3 = &resourceapi.DeviceClass{ObjectMeta: metav1.ObjectMeta{Name: "class-3", UID: "class-3"}}
)

func TestProviderSnapshot(t *testing.T) {
	for _, tc := range []struct {
		testName            string
		claims              []*resourceapi.ResourceClaim
		triggerClaimsError  bool
		slices              []*resourceapi.ResourceSlice
		triggerSlicesError  bool
		classes             []*resourceapi.DeviceClass
		triggerClassesError bool
		wantSnapshot        drasnapshot.Snapshot
		wantErr             error
	}{
		{
			testName:           "claim lister error results in an error",
			triggerClaimsError: true,
			wantErr:            cmpopts.AnyError,
		},
		{
			testName:           "slices lister error results in an error",
			triggerSlicesError: true,
			wantErr:            cmpopts.AnyError,
		},
		{
			testName:            "classes lister error results in an error",
			triggerClassesError: true,
			wantErr:             cmpopts.AnyError,
		},
		{
			testName: "claims are correctly snapshot by id",
			claims:   []*resourceapi.ResourceClaim{claim1, claim2, claim3},
			wantSnapshot: drasnapshot.NewSnapshot(
				map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
					drasnapshot.GetClaimId(claim1): claim1,
					drasnapshot.GetClaimId(claim2): claim2,
					drasnapshot.GetClaimId(claim3): claim3,
				}, nil, nil, nil),
		},
		{
			testName: "slices are correctly divided and snapshot",
			slices:   []*resourceapi.ResourceSlice{localSlice1, localSlice2, localSlice3, localSlice4, globalSlice1, globalSlice2, globalSlice3},
			wantSnapshot: drasnapshot.NewSnapshot(nil,
				map[string][]*resourceapi.ResourceSlice{
					"n1": {localSlice1, localSlice2},
					"n2": {localSlice3, localSlice4},
				},
				[]*resourceapi.ResourceSlice{globalSlice1, globalSlice2, globalSlice3}, nil),
		},
		{
			testName: "classes are correctly snapshot by name",
			classes:  []*resourceapi.DeviceClass{class1, class2, class3},
			wantSnapshot: drasnapshot.NewSnapshot(nil, nil, nil,
				map[string]*resourceapi.DeviceClass{"class-1": class1, "class-2": class2, "class-3": class3}),
		},
		{
			testName: "everything is correctly snapshot together",
			claims:   []*resourceapi.ResourceClaim{claim1, claim2, claim3},
			slices:   []*resourceapi.ResourceSlice{localSlice1, localSlice2, localSlice3, localSlice4, globalSlice1, globalSlice2, globalSlice3},
			classes:  []*resourceapi.DeviceClass{class1, class2, class3},
			wantSnapshot: drasnapshot.NewSnapshot(
				map[drasnapshot.ResourceClaimId]*resourceapi.ResourceClaim{
					drasnapshot.GetClaimId(claim1): claim1,
					drasnapshot.GetClaimId(claim2): claim2,
					drasnapshot.GetClaimId(claim3): claim3,
				},
				map[string][]*resourceapi.ResourceSlice{
					"n1": {localSlice1, localSlice2},
					"n2": {localSlice3, localSlice4},
				},
				[]*resourceapi.ResourceSlice{globalSlice1, globalSlice2, globalSlice3},
				map[string]*resourceapi.DeviceClass{"class-1": class1, "class-2": class2, "class-3": class3}),
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			claimLister := &fakeLister[*resourceapi.ResourceClaim]{objects: tc.claims, triggerErr: tc.triggerClaimsError}
			sliceLister := &fakeLister[*resourceapi.ResourceSlice]{objects: tc.slices, triggerErr: tc.triggerSlicesError}
			classLister := &fakeLister[*resourceapi.DeviceClass]{objects: tc.classes, triggerErr: tc.triggerClassesError}
			provider := NewProvider(claimLister, sliceLister, classLister)
			snapshot, err := provider.Snapshot()
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Provider.Snapshot(): unexpected error (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tc.wantSnapshot, snapshot, cmp.AllowUnexported(drasnapshot.Snapshot{}), cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("Provider.Snapshot(): snapshot differs from expected (-want +got): %s", diff)
			}
		})
	}
}

// TestNewProviderFromInformers verifies that the interface translation listers created in NewProviderFromInformers correctly return
// all objects in the cluster.
func TestNewProviderFromInformers(t *testing.T) {
	for _, tc := range []struct {
		testName string
		claims   []*resourceapi.ResourceClaim
		slices   []*resourceapi.ResourceSlice
		classes  []*resourceapi.DeviceClass
	}{
		{
			testName: "no objects in informers",
		},
		{
			testName: "ResourceClaims present in informers",
			claims:   []*resourceapi.ResourceClaim{claim1, claim2, claim3},
		},
		{
			testName: "ResourceSlices present in informers",
			slices:   []*resourceapi.ResourceSlice{localSlice1, localSlice2, localSlice3},
		},
		{
			testName: "DeviceClasses present in informers",
			classes:  []*resourceapi.DeviceClass{class1, class2, class3},
		},
		{
			testName: "all objects present in informers together",
			claims:   []*resourceapi.ResourceClaim{claim1, claim2, claim3},
			slices:   []*resourceapi.ResourceSlice{localSlice1, localSlice2, localSlice3},
			classes:  []*resourceapi.DeviceClass{class1, class2, class3},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			var objects []runtime.Object
			for _, claim := range tc.claims {
				objects = append(objects, claim)
			}
			for _, slice := range tc.slices {
				objects = append(objects, slice)
			}
			for _, class := range tc.classes {
				objects = append(objects, class)
			}
			client := fake.NewSimpleClientset(objects...)
			informerFactory := informers.NewSharedInformerFactory(client, 0)
			provider := NewProviderFromInformers(informerFactory)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			informerFactory.Start(ctx.Done())
			informerFactory.WaitForCacheSync(ctx.Done())

			allClaims, err := provider.resourceClaims.ListAll()
			if err != nil {
				t.Fatalf("provider.resourceClaims.ListAll(): got unexpected error %v", err)
			}
			if diff := cmp.Diff(tc.claims, allClaims, test.IgnoreObjectOrder[*resourceapi.ResourceClaim]()); diff != "" {
				t.Errorf("provider.resourceClaims.ListAll(): result differs from expected (-want +got): %s", diff)
			}

			allSlices, err := provider.resourceSlices.ListAll()
			if err != nil {
				t.Fatalf("provider.resourceSlices.ListAll(): got unexpected error %v", err)
			}
			if diff := cmp.Diff(tc.slices, allSlices, test.IgnoreObjectOrder[*resourceapi.ResourceSlice]()); diff != "" {
				t.Errorf("provider.resourceSlices.ListAll(): result differs from expected (-want +got): %s", diff)
			}

			allClasses, err := provider.deviceClasses.ListAll()
			if err != nil {
				t.Fatalf("provider.deviceClasses.ListAll(): got unexpected error %v", err)
			}
			if diff := cmp.Diff(tc.classes, allClasses, test.IgnoreObjectOrder[*resourceapi.DeviceClass]()); diff != "" {
				t.Errorf("provider.deviceClasses.ListAll(): result differs from expected (-want +got): %s", diff)
			}
		})
	}
}

type fakeLister[T any] struct {
	objects    []T
	triggerErr bool
}

func (l *fakeLister[T]) ListAll() ([]T, error) {
	var err error
	if l.triggerErr {
		err = fmt.Errorf("fake test error")
	}
	return l.objects, err
}
