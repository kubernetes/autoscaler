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

package builder

import (
	"context"
	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/loop"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"testing"
	"testing/synctest"
	"time"
)

func TestAutoscalerBuilderNoError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		options := config.AutoscalingOptions{
			CloudProviderName: "gce",
			EstimatorName:     estimator.BinpackingEstimatorName,
			ExpanderNames:     expander.LeastWasteExpanderName,
		}

		debuggingSnapshotter := debuggingsnapshot.NewDebuggingSnapshotter(false)
		kubeClient := fake.NewClientset()

		mgr, err := manager.New(&rest.Config{}, manager.Options{
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
			HealthProbeBindAddress: "0",
		})

		autoscaler, trigger, err := New(options).
			WithDebuggingSnapshotter(debuggingSnapshotter).
			WithManager(mgr).
			WithKubeClient(kubeClient).
			WithInformerFactory(informers.NewSharedInformerFactory(kubeClient, 0)).
			WithCloudProvider(test.NewCloudProvider()).
			WithPodObserver(&loop.UnschedulablePodObserver{}).
			Build(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, autoscaler)
		assert.NotNil(t, trigger)

		cancel()

		// Synctest drain: Background goroutines (like MetricAsyncRecorder) often use uninterruptible time.Sleep loops.
		// In a synctest bubble, these are "durable" sleeps. We must advance the virtual clock to allow these goroutines to wake up, observe the
		// closed context channel, and terminate gracefully.
		time.Sleep(1 * time.Second)
	})
}
