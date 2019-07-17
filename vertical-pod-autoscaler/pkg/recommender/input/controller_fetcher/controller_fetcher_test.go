/*
Copyright 2019 The Kubernetes Authors.

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

package controllerfetcher

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

var trueVar = true

func simpleControllerFetcher() *controllerFetcher {
	f := controllerFetcher{}
	f.informersMap = make(map[wellKnownController]cache.SharedIndexInformer)

	for _, kind := range wellKnownControllers {
		f.informersMap[kind] = cache.NewSharedIndexInformer(
			&cache.ListWatch{},
			nil,
			time.Duration(-1),
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	}
	return &f
}

func addController(controller *controllerFetcher, obj runtime.Object) {
	kind := wellKnownController(obj.GetObjectKind().GroupVersionKind().Kind)
	controller.informersMap[kind].GetStore().Add(obj)
}

func TestControllerFetcher(t *testing.T) {
	type testCase struct {
		apiVersion    string
		key           *ControllerKeyWithAPIVersion
		objects       []runtime.Object
		expectedKey   *ControllerKeyWithAPIVersion
		expectedError error
	}
	for i, tc := range []testCase{
		{
			key:           nil,
			expectedKey:   nil,
			expectedError: nil,
		},
		{
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-deployment", Kind: "Deployment", Namespace: "test-namesapce"}},
			expectedKey:   nil,
			expectedError: fmt.Errorf("Deployment test-namesapce/test-deployment does not exist"),
		},
		{
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-deployment", Kind: "Deployment", Namespace: "test-namesapce"}},
			objects: []runtime.Object{&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: "test-namesapce",
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-deployment", Kind: "Deployment", Namespace: "test-namesapce"}}, // Deployment has no parrent
			expectedError: nil,
		},
		{
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-rs", Kind: "ReplicaSet", Namespace: "test-namesapce"}},
			objects: []runtime.Object{&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: "test-namesapce",
				},
			}, &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicaSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-rs",
					Namespace: "test-namesapce",
					OwnerReferences: []metav1.OwnerReference{
						{
							Controller: &trueVar,
							Kind:       "Deployment",
							Name:       "test-deployment",
						},
					},
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-deployment", Kind: "Deployment", Namespace: "test-namesapce"}}, // Deployment has no parent
			expectedError: nil,
		},
		{
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-statefulset", Kind: "StatefulSet", Namespace: "test-namesapce"}},
			objects: []runtime.Object{&appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-statefulset",
					Namespace: "test-namesapce",
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-statefulset", Kind: "StatefulSet", Namespace: "test-namesapce"}}, // StatefulSet has no parent
			expectedError: nil,
		},
		{
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-daemonset", Kind: "DaemonSet", Namespace: "test-namesapce"}},
			objects: []runtime.Object{&appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "DaemonSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-daemonset",
					Namespace: "test-namesapce",
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-daemonset", Kind: "DaemonSet", Namespace: "test-namesapce"}}, // DaemonSet has no parent
			expectedError: nil,
		},
		{
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-job", Kind: "Job", Namespace: "test-namesapce"}},
			objects: []runtime.Object{&batchv1.Job{
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job",
					Namespace: "test-namesapce",
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-job", Kind: "Job", Namespace: "test-namesapce"}}, // Job has no parent
			expectedError: nil,
		},
		{
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-rc", Kind: "ReplicationController", Namespace: "test-namesapce"}},
			objects: []runtime.Object{&corev1.ReplicationController{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicationController",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-rc",
					Namespace: "test-namesapce",
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-rc", Kind: "ReplicationController", Namespace: "test-namesapce"}}, // ReplicationController has no parent
			expectedError: nil,
		},
		{
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-deployment", Kind: "Deployment", Namespace: "test-namesapce"}},
			objects: []runtime.Object{&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: "test-namesapce",
					// Deployment points to itself
					OwnerReferences: []metav1.OwnerReference{
						{
							Controller: &trueVar,
							Kind:       "Deployment",
							Name:       "test-deployment",
						},
					},
				},
			}},
			expectedKey:   nil,
			expectedError: fmt.Errorf("Cycle detected in ownership chain"),
		},
	} {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			f := simpleControllerFetcher()
			for _, obj := range tc.objects {
				addController(f, obj)
			}
			topLevelController, err := f.FindTopLevel(tc.key)
			if tc.expectedKey == nil {
				assert.Nil(t, topLevelController)
			} else {
				assert.Equal(t, tc.expectedKey, topLevelController)
			}
			if tc.expectedError == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.expectedError, err)
			}
		})
	}
}
