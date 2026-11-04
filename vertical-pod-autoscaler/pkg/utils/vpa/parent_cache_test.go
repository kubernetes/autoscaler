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
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"

	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
)

func TestFindParentControllerForPodDuringCacheLag(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var discoveryRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		discoveryRequests.Add(1)
		http.Error(w, "unexpected discovery request", http.StatusNotFound)
	}))
	defer server.Close()

	cron := &batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{
		Namespace: "example", Name: "periodic", UID: "cron-uid",
	}}
	client := fake.NewClientset(cron)
	var ownerReads atomic.Int32
	client.PrependReactor("get", "*", func(clienttesting.Action) (bool, runtime.Object, error) {
		ownerReads.Add(1)
		return false, nil, nil
	})
	jobEvents := watch.NewRaceFreeFake()
	watchStarted := make(chan struct{})
	var watches atomic.Int32
	client.PrependWatchReactor("jobs", func(clienttesting.Action) (bool, watch.Interface, error) {
		if watches.Add(1) == 1 {
			close(watchStarted)
		}
		return true, jobEvents, nil
	})
	factory := informers.NewSharedInformerFactory(client, 0)
	defer func() {
		cancel()
		factory.Shutdown()
		jobEvents.Stop()
	}()
	fetcher := controllerfetcher.NewControllerFetcher(&rest.Config{Host: server.URL}, client, factory,
		time.Second, time.Minute, 0, ctx.Done())
	jobInformer := factory.Batch().V1().Jobs().Informer()
	factory.Start(ctx.Done())
	for resource, synced := range factory.WaitForCacheSync(ctx.Done()) {
		require.True(t, synced, "informer %v did not synchronize", resource)
	}
	select {
	case <-watchStarted:
	case <-ctx.Done():
		t.Fatal("Job watch did not start")
	}

	job, err := client.BatchV1().Jobs(cron.Namespace).Create(ctx, &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Namespace: cron.Namespace, Name: "periodic-1", UID: "job-uid",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "batch/v1", Kind: "CronJob",
				Name: cron.Name, UID: cron.UID, Controller: ptr.To(true)}}},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	authoritative, err := client.BatchV1().Jobs(cron.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, job.UID, authoritative.UID)
	_, exists, err := jobInformer.GetStore().GetByKey(cron.Namespace + "/" + job.Name)
	require.NoError(t, err)
	require.False(t, exists, "Job watch must still be held")
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: cron.Namespace, Name: "periodic-1-pod",
		OwnerReferences: []metav1.OwnerReference{{APIVersion: "batch/v1", Kind: "Job",
			Name: job.Name, UID: job.UID, Controller: ptr.To(true)}}}}

	ownerReads.Store(0)
	parent, err := FindParentControllerForPod(ctx, pod, fetcher)
	require.NoError(t, err, "API owner exists while its informer event is delayed")
	require.NotNil(t, parent)
	require.Equal(t, controllerfetcher.ControllerKey{Namespace: cron.Namespace, Kind: "CronJob", Name: cron.Name}, parent.ControllerKey)
	require.Equal(t, "batch/v1", parent.ApiVersion)
	require.EqualValues(t, 1, ownerReads.Load(), "cache miss must use exactly one API read")
	_, exists, err = jobInformer.GetStore().GetByKey(cron.Namespace + "/" + job.Name)
	require.NoError(t, err)
	require.False(t, exists, "fallback must not populate the informer store")

	jobEvents.Add(job.DeepCopy())
	require.NoError(t, wait.PollUntilContextTimeout(ctx, time.Millisecond, 2*time.Second, true,
		func(context.Context) (bool, error) {
			_, exists, err := jobInformer.GetStore().GetByKey(cron.Namespace + "/" + job.Name)
			return exists, err
		}))
	ownerReads.Store(0)
	warmParent, err := FindParentControllerForPod(ctx, pod, fetcher)
	require.NoError(t, err)
	require.Equal(t, parent, warmParent)
	require.Zero(t, ownerReads.Load(), "cache hits must not read from the API")
	require.Zero(t, discoveryRequests.Load())
}
