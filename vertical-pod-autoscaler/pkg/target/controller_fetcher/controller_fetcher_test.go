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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
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

const (
	testNamespace             = "test-namespace"
	testDeployment            = "test-deployment"
	testReplicaSet            = "test-rs"
	testStatefulSet           = "test-statefulset"
	testDaemonSet             = "test-daemonset"
	testCronJob               = "test-cronjob"
	testJob                   = "test-job"
	testReplicationController = "test-rc"
)

var (
	wellKnownControllers = []wellKnownController{daemonSet, deployment, replicaSet, statefulSet, replicationController, job, cronJob}
	trueVar              = true
)

func simpleControllerFetcher() *controllerFetcher {
	f := controllerFetcher{}
	f.informersMap = make(map[wellKnownController]cache.SharedIndexInformer)
	f.scaleSubresourceCacheStorage = newControllerCacheStorage(time.Second, time.Minute, 0.1)
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

	// return not found if if tries to find the scale subresource on bah
	scaleNamespacer.AddReactor("get", "bah", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		groupResource := schema.GroupResource{}
		error := apierrors.NewNotFound(groupResource, "Foo")
		return true, nil, error
	})

	// resource that can scale
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

func addController(t *testing.T, controller *controllerFetcher, obj runtime.Object) {
	kind := wellKnownController(obj.GetObjectKind().GroupVersionKind().Kind)
	_, ok := controller.informersMap[kind]
	if ok {
		err := controller.informersMap[kind].GetStore().Add(obj)
		assert.NoError(t, err)
	}
}

func TestControllerFetcher(t *testing.T) {
	type testCase struct {
		name          string
		key           *ControllerKeyWithAPIVersion
		objects       []runtime.Object
		expectedKey   *ControllerKeyWithAPIVersion
		expectedError error
	}
	for _, tc := range []testCase{
		{
			name:          "nils",
			key:           nil,
			expectedKey:   nil,
			expectedError: nil,
		},
		{
			name: "deployment doesn't exist",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDeployment, Kind: "Deployment", Namespace: testNamespace}},
			expectedKey:   nil,
			expectedError: fmt.Errorf("Deployment %s/%s does not exist", testNamespace, testDeployment),
		},
		{
			name: "deployment no parent",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDeployment, Kind: "Deployment", Namespace: testNamespace}},
			objects: []runtime.Object{&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDeployment,
					Namespace: testNamespace,
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDeployment, Kind: "Deployment", Namespace: testNamespace}}, // Deployment has no parent
			expectedError: nil,
		},
		{
			name: "deployment with parent",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testReplicaSet, Kind: "ReplicaSet", Namespace: testNamespace}},
			objects: []runtime.Object{&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDeployment,
					Namespace: testNamespace,
				},
			}, &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicaSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testReplicaSet,
					Namespace: testNamespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							Controller: &trueVar,
							Kind:       "Deployment",
							Name:       testDeployment,
						},
					},
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDeployment, Kind: "Deployment", Namespace: testNamespace}}, // Deployment has no parent
			expectedError: nil,
		},
		{
			name: "StatefulSet",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testStatefulSet, Kind: "StatefulSet", Namespace: testNamespace}},
			objects: []runtime.Object{&appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testStatefulSet,
					Namespace: testNamespace,
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testStatefulSet, Kind: "StatefulSet", Namespace: testNamespace}}, // StatefulSet has no parent
			expectedError: nil,
		},
		{
			name: "DaemonSet",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDaemonSet, Kind: "DaemonSet", Namespace: testNamespace}},
			objects: []runtime.Object{&appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "DaemonSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDaemonSet,
					Namespace: testNamespace,
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDaemonSet, Kind: "DaemonSet", Namespace: testNamespace}}, // DaemonSet has no parent
			expectedError: nil,
		},
		{
			name: "CronJob",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testJob, Kind: "Job", Namespace: testNamespace}},
			objects: []runtime.Object{&batchv1.Job{
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testJob,
					Namespace: testNamespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							Controller: &trueVar,
							Kind:       "CronJob",
							Name:       testCronJob,
						},
					},
				},
			}, &batchv1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCronJob,
					Namespace: testNamespace,
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testCronJob, Kind: "CronJob", Namespace: testNamespace}}, // CronJob has no parent
			expectedError: nil,
		},
		{
			name: "CronJob no parent",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testCronJob, Kind: "CronJob", Namespace: testNamespace}},
			objects: []runtime.Object{&batchv1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCronJob,
					Namespace: testNamespace,
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testCronJob, Kind: "CronJob", Namespace: testNamespace}}, // CronJob has no parent
			expectedError: nil,
		},
		{
			name: "rc no parent",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testReplicationController, Kind: "ReplicationController", Namespace: testNamespace}},
			objects: []runtime.Object{&corev1.ReplicationController{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicationController",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testReplicationController,
					Namespace: testNamespace,
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testReplicationController, Kind: "ReplicationController", Namespace: testNamespace}}, // ReplicationController has no parent
			expectedError: nil,
		},
		{
			name: "deployment cycle in ownership",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDeployment, Kind: "Deployment", Namespace: testNamespace}},
			objects: []runtime.Object{&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDeployment,
					Namespace: testNamespace,
					// Deployment points to itself
					OwnerReferences: []metav1.OwnerReference{
						{
							Controller: &trueVar,
							Kind:       "Deployment",
							Name:       testDeployment,
						},
					},
				},
			}},
			expectedKey:   nil,
			expectedError: fmt.Errorf("Cycle detected in ownership chain"),
		},
		{
			name: "deployment, parent with no scale subresource",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDeployment, Kind: "Deployment", Namespace: testNamespace}},
			objects: []runtime.Object{&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDeployment,
					Namespace: testNamespace,
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
				Name: testDeployment, Kind: "Deployment", Namespace: testNamespace}}, // Parent does not support scale subresource so should return itself"
			expectedError: nil,
		},
		{
			name: "deployment, parent not well known with scale subresource",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDeployment, Kind: "Deployment", Namespace: testNamespace}},
			objects: []runtime.Object{&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDeployment,
					Namespace: testNamespace,
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
				Name: "iCanScale", Kind: "Scale", Namespace: testNamespace}, ApiVersion: "Foo/Foo"}, // Parent supports scale subresource"
			expectedError: nil,
		},
		{
			name: "pod, parent is node",
			key: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDeployment, Kind: "Deployment", Namespace: testNamespace}},
			objects: []runtime.Object{&appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      testDeployment,
					Namespace: testNamespace,
					// Parent is a node
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Controller: &trueVar,
							Kind:       "Node",
							Name:       "node",
						},
					},
				},
			}},
			expectedKey:   nil,
			expectedError: fmt.Errorf("Unhandled targetRef v1 / Node / node, last error node is not a valid owner"),
		},
		{
			name: "custom resource with no scale subresource",
			key: &ControllerKeyWithAPIVersion{
				ApiVersion: "Foo/Foo", ControllerKey: ControllerKey{
					Name: "bah", Kind: "Foo", Namespace: testNamespace},
			},
			objects:       []runtime.Object{},
			expectedKey:   nil, // Pod owner does not support scale subresource so should return nil"
			expectedError: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := simpleControllerFetcher()
			for _, obj := range tc.objects {
				addController(t, f, obj)
			}
			topMostWellKnownOrScalableController, err := f.FindTopMostWellKnownOrScalable(context.Background(), tc.key)
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
