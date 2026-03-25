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

	resourceapi "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	schedulerinterface "k8s.io/kube-scheduler/framework"
)

var (
	class1 = &resourceapi.DeviceClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1", UID: "class-1"}}
	class2 = &resourceapi.DeviceClass{ObjectMeta: metav1.ObjectMeta{Name: "class-2", UID: "class-2"}}
	class3 = &resourceapi.DeviceClass{ObjectMeta: metav1.ObjectMeta{Name: "class-3", UID: "class-3"}}
)

func TestSnapshotClassListerList(t *testing.T) {
	for _, tc := range []struct {
		testName    string
		classes     map[string]*resourceapi.DeviceClass
		wantClasses []*resourceapi.DeviceClass
	}{
		{
			testName:    "no classes in snapshot",
			wantClasses: []*resourceapi.DeviceClass{},
		},
		{
			testName:    "classes present in snapshot",
			classes:     map[string]*resourceapi.DeviceClass{"class-1": class1, "class-2": class2, "class-3": class3},
			wantClasses: []*resourceapi.DeviceClass{class1, class2, class3},
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			snapshot := NewSnapshot(nil, nil, nil, tc.classes)
			var deviceClassLister schedulerinterface.DeviceClassLister = snapshot.DeviceClasses()
			classes, err := deviceClassLister.List()
			if err != nil {
				t.Fatalf("snapshotClassLister.List(): got unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantClasses, classes, cmpopts.EquateEmpty(), test.IgnoreObjectOrder[*resourceapi.DeviceClass]()); diff != "" {
				t.Errorf("snapshotClassLister.List(): unexpected output (-want +got): %s", diff)
			}
		})
	}
}

func TestSnapshotClassListerGet(t *testing.T) {
	for _, tc := range []struct {
		testName  string
		classes   map[string]*resourceapi.DeviceClass
		className string
		wantClass *resourceapi.DeviceClass
		wantErr   error
	}{
		{
			testName:  "class present in snapshot",
			className: "class-2",
			wantClass: class2,
		},
		{
			testName:  "class not present in snapshot",
			className: "class-1337",
			wantErr:   cmpopts.AnyError,
		},
	} {
		t.Run(tc.testName, func(t *testing.T) {
			classes := map[string]*resourceapi.DeviceClass{"class-1": class1, "class-2": class2, "class-3": class3}
			snapshot := NewSnapshot(nil, nil, nil, classes)
			var deviceClassLister schedulerinterface.DeviceClassLister = snapshot.DeviceClasses()
			class, err := deviceClassLister.Get(tc.className)
			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("snapshotClassLister.Get(): unexpected error (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tc.wantClass, class); diff != "" {
				t.Errorf("snapshotClassLister.Get(): unexpected output (-want +got): %s", diff)
			}
		})
	}
}
