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

package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/builder"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/loop"
	fakek8s "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes/fake"
	"testing"
	"testing/synctest"
	"time"
)

// TestContext acts as the bridge between the test logic and the simulation.
type TestContext struct {
	t          *testing.T
	Autoscaler core.Autoscaler
}

// Fakes is the struct used at test phase to make assertions.
type Fakes struct {
	FakeK8s *fakek8s.Kubernetes
}

// RunOnceAfter advances the virtual clock by the specified duration and then
// executes a single Cluster Autoscaler cycle.
func (c *TestContext) RunOnceAfter(d time.Duration) error {
	c.t.Helper()

	// Ensure any pending work is done  before changing the time.
	synctest.Wait()

	time.Sleep(d)
	err := c.Autoscaler.RunOnce(time.Now())

	// Let side-effects of the RunOnce finish.
	synctest.Wait()
	return err
}

// MustRunOnceAfter is a helper that calls RunOnceAfter and
// immediately fails the test if an error occurs.
// Use this for "happy path" simulation steps.
func (c *TestContext) MustRunOnceAfter(d time.Duration) {
	c.t.Helper()
	err := c.RunOnceAfter(d)
	assert.NoError(c.t, err)
}

// RunTest encapsulates the setup, execution, and teardown.
func RunTest(t *testing.T, config *Config, scenario func(*TestContext)) {
	t.Helper()

	debuggingSnapshotter := debuggingsnapshot.NewDebuggingSnapshotter(false)

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer tearDown(cancel)

		autoscaler, _, err := builder.NewAutoscaler(*config.AutoscalingOptions).
			WithKubeClient(config.KubeClient).
			WithCloudProvider(config.CloudProvider).
			WithPodObserver(&loop.UnschedulablePodObserver{}).
			Build(ctx, debuggingSnapshotter)

		assert.NoError(t, err)
		assert.NotNil(t, autoscaler)

		testCtx := &TestContext{
			t:          t,
			Autoscaler: autoscaler,
		}
		scenario(testCtx)
	})
}

func tearDown(cancel context.CancelFunc) {
	cancel()
	// Synctest drain: Background goroutines (like MetricAsyncRecorder) often use uninterruptible time.Sleep loops.
	// In a synctest bubble, these are "durable" sleeps. We must advance the virtual clock to allow these goroutines to wake up, observe the
	// closed context channel, and terminate gracefully.
	time.Sleep(1 * time.Second)
}
