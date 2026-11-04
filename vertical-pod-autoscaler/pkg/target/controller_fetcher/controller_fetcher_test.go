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
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
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
	f.kubeClient = fake.NewClientset()
	f.informersMap = make(map[wellKnownController]cache.SharedIndexInformer)
	f.scaleSubresourceCacheStorage = newControllerCacheStorage(time.Second, time.Minute, 0.1)
	versioned := map[string][]metav1.APIResource{
		"Foo": {
			{Kind: "NotScalingResource", Name: "iCanNotScale", Group: "foo"},
			{Kind: "ScalingResource", Name: "iCanScale", Group: "foo"},
		},
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

	// return not found if it tries to find the scale subresource on iCanNotScale
	scaleNamespacer.AddReactor("get", "iCanNotScale", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		groupResource := schema.GroupResource{}
		err = apierrors.NewNotFound(groupResource, "Foo")
		return true, nil, err
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
			expectedError: apierrors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "deployments"}, testDeployment),
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
			expectedError: errors.New("cycle detected in ownership chain"),
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
							Kind:       "NotScalingResource",
							Name:       "iCanNotScale",
						},
					},
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: testDeployment, Kind: "Deployment", Namespace: testNamespace}}, // Parent does not support scale subresource so should return itself
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
							Kind:       "ScalingResource",
							Name:       "iCanScale",
						},
					},
				},
			}},
			expectedKey: &ControllerKeyWithAPIVersion{ControllerKey: ControllerKey{
				Name: "iCanScale", Kind: "ScalingResource", Namespace: testNamespace}, ApiVersion: "Foo/Foo"}, // Parent supports scale subresource so it is returned
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
			expectedError: fmt.Errorf("unhandled targetRef v1 / Node / node, last error %w", ErrNodeInvalidOwner),
		},
		{
			name: "custom resource with no scale subresource",
			key: &ControllerKeyWithAPIVersion{
				ApiVersion: "Foo/Foo", ControllerKey: ControllerKey{
					Name: "iCanNotScale", Kind: "NotScalingResource", Namespace: testNamespace},
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

func TestWellKnownControllerCacheLookup(t *testing.T) {
	meta := metav1.ObjectMeta{Name: "controller", Namespace: testNamespace, UID: "owner-uid"}
	for _, tc := range []struct {
		kind     wellKnownController
		resource string
		object   runtime.Object
	}{
		{daemonSet, "daemonsets", &appsv1.DaemonSet{ObjectMeta: meta}},
		{deployment, "deployments", &appsv1.Deployment{ObjectMeta: meta}},
		{statefulSet, "statefulsets", &appsv1.StatefulSet{ObjectMeta: meta}},
		{replicaSet, "replicasets", &appsv1.ReplicaSet{ObjectMeta: meta}},
		{job, "jobs", &batchv1.Job{ObjectMeta: meta}},
		{cronJob, "cronjobs", &batchv1.CronJob{ObjectMeta: meta}},
		{replicationController, "replicationcontrollers", &corev1.ReplicationController{ObjectMeta: meta}},
	} {
		for _, cached := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/cached=%t", tc.kind, cached), func(t *testing.T) {
				client := fake.NewClientset(tc.object)
				f := simpleControllerFetcher()
				f.kubeClient = client
				store := f.informersMap[tc.kind].GetStore()
				if cached {
					require.NoError(t, store.Add(tc.object))
				}
				key := &ControllerKeyWithAPIVersion{
					ControllerKey: ControllerKey{Namespace: meta.Namespace, Name: meta.Name, Kind: string(tc.kind)},
					UID:           meta.UID,
				}
				parent, err := f.FindTopMostWellKnownOrScalable(context.Background(), key)
				require.NoError(t, err)
				require.Equal(t, key, parent)
				if cached {
					require.Empty(t, client.Actions())
				} else {
					actions := client.Actions()
					require.Len(t, actions, 1)
					require.True(t, actions[0].Matches("get", tc.resource))
					require.Equal(t, meta.Namespace, actions[0].GetNamespace())
					require.Equal(t, meta.Name, actions[0].(core.GetAction).GetName())
					require.Empty(t, store.List(), "fallback must not populate the informer")
				}
			})
		}
	}
}

func TestJobControllerOwnershipChain(t *testing.T) {
	for _, tc := range []struct {
		name       string
		cachedJob  bool
		cachedCron bool
		mismatch   wellKnownController
		deleted    wellKnownController
	}{
		{name: "both owners cached", cachedJob: true, cachedCron: true},
		{name: "Job cache lags", cachedCron: true},
		{name: "CronJob cache lags", cachedJob: true},
		{name: "both caches lag"},
		{name: "replaced Job in API", cachedCron: true, mismatch: job},
		{name: "replaced CronJob in API", cachedJob: true, mismatch: cronJob},
		{name: "replaced Job in cache", cachedJob: true, cachedCron: true, mismatch: job},
		{name: "replaced CronJob in cache", cachedJob: true, cachedCron: true, mismatch: cronJob},
		{name: "deleted Job", cachedCron: true, deleted: job},
		{name: "deleted CronJob", cachedJob: true, deleted: cronJob},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cron := &batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{
				Namespace: testNamespace, Name: testCronJob, UID: "cron-uid",
			}}
			jobObj := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{
				Namespace: testNamespace, Name: testJob, UID: "job-uid",
				OwnerReferences: []metav1.OwnerReference{{APIVersion: "batch/v1", Kind: "CronJob",
					Name: cron.Name, UID: cron.UID, Controller: &trueVar}},
			}}
			key := &ControllerKeyWithAPIVersion{
				ControllerKey: ControllerKey{Namespace: testNamespace, Kind: "Job", Name: testJob},
				ApiVersion:    "batch/v1", UID: jobObj.UID,
			}
			client := fake.NewClientset(jobObj, cron)
			f := simpleControllerFetcher()
			f.kubeClient = client
			if tc.mismatch == job {
				jobObj.UID = "replacement-uid"
			}
			if tc.mismatch == cronJob {
				cron.UID = "replacement-uid"
			}
			for kind, obj := range map[wellKnownController]runtime.Object{job: jobObj, cronJob: cron} {
				cached := kind == job && tc.cachedJob || kind == cronJob && tc.cachedCron
				if cached {
					require.NoError(t, f.informersMap[kind].GetStore().Add(obj))
				} else {
					// The API can contain a same-name replacement while an old
					// owner reference still identifies the deleted object's UID.
					meta, err := apimeta.Accessor(obj)
					require.NoError(t, err)
					resource := "jobs"
					if kind == cronJob {
						resource = "cronjobs"
					}
					require.NoError(t, client.Tracker().Update(batchv1.SchemeGroupVersion.WithResource(resource), obj, meta.GetNamespace()))
				}
			}
			if tc.deleted != "" {
				resource, name := "jobs", testJob
				if tc.deleted == cronJob {
					resource, name = "cronjobs", testCronJob
				}
				require.NoError(t, client.Tracker().Delete(batchv1.SchemeGroupVersion.WithResource(resource), testNamespace, name))
			}
			parent, err := f.FindTopMostWellKnownOrScalable(context.Background(), key)
			if tc.deleted != "" {
				require.True(t, apierrors.IsNotFound(err), "expected deletion error, got %v", err)
				require.Nil(t, parent)
				require.Len(t, client.Actions(), 1)
				return
			}
			if tc.mismatch != "" {
				require.ErrorContains(t, err, "has UID \"replacement-uid\", expected")
				require.Nil(t, parent)
				if tc.mismatch == job && tc.cachedJob || tc.mismatch == cronJob && tc.cachedCron {
					require.Empty(t, client.Actions(), "UID mismatches must not trigger API recovery")
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, &ControllerKeyWithAPIVersion{
				ControllerKey: ControllerKey{Namespace: testNamespace, Kind: "CronJob", Name: testCronJob},
				ApiVersion:    "batch/v1", UID: cron.UID,
			}, parent)
			wantGets := 0
			if !tc.cachedJob {
				wantGets++
			}
			if !tc.cachedCron {
				wantGets++
			}
			require.Len(t, client.Actions(), wantGets)
		})
	}
}

func TestScaleControllerOwnerUID(t *testing.T) {
	for _, uid := range []types.UID{"", "scale-uid", "deleted-uid"} {
		t.Run(string(uid), func(t *testing.T) {
			f := simpleControllerFetcher()
			f.scaleNamespacer.(*scalefake.FakeScaleClient).PrependReactor("get", "iCanScale", func(core.Action) (bool, runtime.Object, error) {
				return true, &autoscalingv1.Scale{ObjectMeta: metav1.ObjectMeta{UID: "scale-uid"}}, nil
			})
			key := &ControllerKeyWithAPIVersion{
				ControllerKey: ControllerKey{Namespace: testNamespace, Kind: "ScalingResource", Name: "iCanScale"},
				ApiVersion:    "Foo/Foo", UID: uid,
			}
			parent, err := f.FindTopMostWellKnownOrScalable(context.Background(), key)
			if uid == "deleted-uid" {
				require.ErrorContains(t, err, "has UID \"scale-uid\", expected \"deleted-uid\"")
				require.Nil(t, parent)
				return
			}
			require.NoError(t, err)
			require.Equal(t, key, parent)
		})
	}
}

func TestWellKnownControllerAPIErrors(t *testing.T) {
	for _, kind := range []wellKnownController{job, cronJob} {
		resource := "jobs"
		if kind == cronJob {
			resource = "cronjobs"
		}
		gr := schema.GroupResource{Group: "batch", Resource: resource}
		for _, apiErr := range []error{
			apierrors.NewNotFound(gr, "deleted-owner"),
			apierrors.NewForbidden(gr, "owner", errors.New("denied")),
			apierrors.NewInternalError(errors.New("API unavailable")),
			apierrors.NewTimeoutError("request timed out", 1),
			context.Canceled,
		} {
			t.Run(string(kind)+"/"+apiErr.Error(), func(t *testing.T) {
				client := fake.NewClientset()
				client.PrependReactor("get", resource, func(core.Action) (bool, runtime.Object, error) {
					return true, nil, apiErr
				})
				f := simpleControllerFetcher()
				f.kubeClient = client
				parent, err := f.FindTopMostWellKnownOrScalable(context.Background(), &ControllerKeyWithAPIVersion{
					ControllerKey: ControllerKey{Namespace: testNamespace, Kind: string(kind), Name: "owner"},
					ApiVersion:    "batch/v1",
				})
				require.ErrorIs(t, err, apiErr)
				require.Nil(t, parent)
				require.Len(t, client.Actions(), 1, "API errors must not be retried or suppressed")
			})
		}
	}
}
