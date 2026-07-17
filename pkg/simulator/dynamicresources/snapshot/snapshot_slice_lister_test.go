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

package snapshot

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	schedulerinterface "k8s.io/kube-scheduler/framework"
)

func TestSnapshotSliceListerList(t *testing.T) {
	var (
		n1Name       = "n1"
		n2Name       = "n2"
		trueValue    = true
		localSlice1  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-1", UID: "local-slice-1"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &n1Name}}
		localSlice2  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-2", UID: "local-slice-2"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &n1Name}}
		localSlice3  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-3", UID: "local-slice-3"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &n2Name}}
		localSlice4  = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "local-slice-4", UID: "local-slice-4"}, Spec: resourceapi.ResourceSliceSpec{NodeName: &n2Name}}
		globalSlice1 = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "global-slice-1", UID: "global-slice-1"}, Spec: resourceapi.ResourceSliceSpec{AllNodes: &trueValue}}
		globalSlice2 = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "global-slice-2", UID: "global-slice-2"}, Spec: resourceapi.ResourceSliceSpec{AllNodes: &trueValue}}
		globalSlice3 = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: "global-slice-3", UID: "global-slice-3"}, Spec: resourceapi.ResourceSliceSpec{NodeSelector: &apiv1.NodeSelector{}}}
	)

	for _, tc := range []struct {
		testName     string
		localSlices  map[string][]*resourceapi.ResourceSlice
		globalSlices []*resourceapi.ResourceSlice
		wantSlices   []*resourceapi.ResourceSlice
	}{
		{
			testName:   "no slices in snapshot",
			wantSlices: []*resourceapi.ResourceSlice{},
		},
		{
			testName: "local slices in snapshot",
			localSlices: map[string][]*resourceapi.ResourceSlice{
				"n1": {localSlice1, localSlice2},
				"n2": {localSlice3, localSlice4},
			},
			wantSlices: []*resourceapi.ResourceSlice{localSlice1, localSlice2, localSlice3, localSlice4},
		},
		{
			testName:     "global slices in snapshot",
			globalSlices: []*resourceapi.ResourceSlice{globalSlice1, globalSlice2, globalSlice3},
			wantSlices:   []*resourceapi.ResourceSlice{globalSlice1, globalSlice2, globalSlice3},
		},
		{
			testName: "global and local slices in snapshot",
			localSlices: map[string][]*resourceapi.ResourceSlice{
				"n1": {localSlice1, localSlice2},
				"n2": {localSlice3, localSlice4},
			},
			globalSlices: []*resourceapi.ResourceSlice{globalSlice1, globalSlice2, globalSlice3},
			wantSlices:   []*resourceapi.ResourceSlice{localSlice1, localSlice2, localSlice3, localSlice4, globalSlice1, globalSlice2, globalSlice3},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			snapshot := NewSnapshot(nil, tc.localSlices, tc.globalSlices, nil)
			var resourceSliceLister schedulerinterface.ResourceSliceLister = snapshot.ResourceSlices()
			slices, err := resourceSliceLister.ListWithDeviceTaintRules()
			if err != nil {
				t.Fatalf("snapshotSliceLister.List(): got unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantSlices, slices, cmpopts.EquateEmpty(), test.IgnoreObjectOrder[*resourceapi.ResourceSlice]()); diff != "" {
				t.Errorf("snapshotSliceLister.List(): unexpected output (-want +got): %s", diff)
			}
		})
	}
}
