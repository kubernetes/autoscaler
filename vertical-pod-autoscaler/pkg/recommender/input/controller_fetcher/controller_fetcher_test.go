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
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/restmapper"
	scalefake "k8s.io/client-go/scale/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

var wellKnownControllers = []wellKnownController{daemonSet, deployment, replicaSet, statefulSet, replicationController, job, cronJob}
var trueVar = true

func simpleControllerFetcher() *controllerFetcher {
	f := controllerFetcher{}
	f.informersMap = make(map[wellKnownController]cache.SharedIndexInformer)
	versioned := map[string][]metav1.APIResource{
		"Foo": {{Kind: "Foo", Name: "bah", Group: "foo"}, {Kind: "Scale", Name: "iCanScale", Group: "foo"}},
	}
	fakeMapper := []*restmapper.APIGroupResources{
		{
			Group: metav1.APIGroup{
				Name:     "Foo",
				Versions: []metav1.GroupVersionForDiscovery{{GroupVersion: "Foo", Version: "Foo"}},
			},
			VersionedResources: versioned,
		},
	}
	mapper := restmapper.NewDiscoveryRESTMapper(fakeMapper)
	f.mapper = mapper

	scaleNamespacer := &scalefake.FakeScaleClient{}
	f.scaleNamespacer = scaleNamespacer

	//return not found if if tries to find the scale subresouce on bah
	scaleNamespacer.AddReactor("get", "bah", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		groupResource := schema.GroupResource{}
		error := apierrors.NewNotFound(groupResource, "Foo")
		return true, nil, error
	})

	//resource that can scale
	scaleNamespacer.AddReactor("get", "iCanScale", func(action core.Action) (handled bool, ret runtime.Object, err error) {

		ret = &autoscalingv1.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Scaler",
				Namespace: "foo",
			},
			Spec: autoscalingv1.ScaleSpec{
				Replicas: 5,
			},
			Status: autoscalingv1.ScaleStatus{
				Replicas: 5,
			},
		}
		return true, ret, nil
	})

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
	_, ok := controller.informersMap[kind]
	if ok {
		controller.informersMap[kind].GetStore().Add(obj)
	}
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
				Name: "test-job", Kind: "Job", Namespace: "test-namespace"}},
			objects: []runtime.Object{&batchv1.Job{
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job",
					Namespace: "test-namespace",
					OwnerReferences: []metav1.OwnerReference{
						{
							Controller: &trueVar,
							Kind:       "CronJob",
							Name:       "test-cronjob",
						},
					},
				},
			}, &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cronjob",
					Namespace: "test-namespace",
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-cronjob", Kind: "CronJob", Namespace: "test-namespace"}}, // CronJob has no parent
			expectedError: nil,
		},
		{
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-cronjob", Kind: "CronJob", Namespace: "test-namespace"}},
			objects: []runtime.Object{&batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cronjob",
					Namespace: "test-namespace",
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-cronjob", Kind: "CronJob", Namespace: "test-namespace"}}, // CronJob has no parent
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
					// Parent that does not support scale subresource and is not well known
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "Foo/Foo",
							Controller: &trueVar,
							Kind:       "Foo",
							Name:       "bah",
						},
					},
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "test-deployment", Kind: "Deployment", Namespace: "test-namesapce"}}, // Parent does not support scale subresource so should return itself"
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
					// Parent that support scale subresource and is not well known
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "Foo/Foo",
							Controller: &trueVar,
							Kind:       "Scale",
							Name:       "iCanScale",
						},
					},
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "iCanScale", Kind: "Scale", Namespace: "test-namesapce"}, ApiVersion: "Foo/Foo"}, // Parent supports scale subresource"
			expectedError: nil,
		},
	} {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			f := simpleControllerFetcher()
			for _, obj := range tc.objects {
				addController(f, obj)
			}
			topMostWellKnownOrScalableController, err := f.FindTopMostWellKnownOrScalable(tc.key)
			if tc.expectedKey == nil {
				assert.Nil(t, topMostWellKnownOrScalableController)
			} else {
				assert.Equal(t, tc.expectedKey, topMostWellKnownOrScalableController)
			}
			if tc.expectedError == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.expectedError, err)
			}
		})
	}
}
