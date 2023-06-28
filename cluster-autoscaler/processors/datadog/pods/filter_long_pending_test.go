/*
Copyright 2021 The Kubernetes Authors.

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

package pods

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	cacontext "k8s.io/autoscaler/cluster-autoscaler/context"
)

var testScaleDownDelay = time.Minute

var testCtx = &cacontext.AutoscalingContext{
	AutoscalingOptions: config.AutoscalingOptions{
		ScaleDownDelayAfterAdd: testScaleDownDelay,
	},
}

func TestProcess(t *testing.T) {
	// Case 1: Survive empty args (no pending pods)
	p := NewFilterOutLongPending()
	allowedPods, _ := p.Process(testCtx, []*apiv1.Pod{})
	assert.Empty(t, allowedPods)

	// Case 1: Survive standalone pods not having controlleref
	p = NewFilterOutLongPending()
	pendingPods := []*apiv1.Pod{{}}
	allowedPods, _ = p.Process(testCtx, pendingPods)
	assert.ElementsMatch(t, allowedPods, pendingPods)

	// Case 2: Filter out old pods having many attempts
	now = func() time.Time { return time.Now().Add(-time.Hour) }
	p = NewFilterOutLongPending()
	pendingPods = buildPendingPods(1, "foo", time.Now().Add(-time.Hour))
	for i := 1; i < 2*minAttempts; i++ {
		allowedPods, _ = p.Process(testCtx, pendingPods)
		now = func() time.Time { return time.Now() }
	}
	assert.Empty(t, allowedPods)

	// Case 3: Old pods with only few attempts are kept
	now = func() time.Time { return time.Now().Add(-time.Hour) }
	p = NewFilterOutLongPending()
	pendingPods = buildPendingPods(7, "foo", time.Now().Add(-time.Hour))
	for i := 1; i < minAttempts; i++ {
		allowedPods, _ = p.Process(testCtx, pendingPods)
		now = func() time.Time { return time.Now() }
	}
	assert.ElementsMatch(t, pendingPods, allowedPods)

	// Case 4: Recent pods having many attempts are kept
	p = NewFilterOutLongPending()
	pendingPods = buildPendingPods(7, "foo", now())
	for i := 1; i < 2*minAttempts; i++ {
		allowedPods, _ = p.Process(testCtx, pendingPods)
	}
	assert.ElementsMatch(t, pendingPods, allowedPods)

	// Case 5: New pod reset counters for all pods sharing its ownerref
	now = func() time.Time { return time.Now().Add(-time.Hour) }
	p = NewFilterOutLongPending()
	pendingPods = buildPendingPods(1, "foo", time.Now().Add(-time.Hour))
	newPod := buildPendingPods(1, "foo", time.Now())
	pendingPods = append(pendingPods, newPod...)
	for i := 1; i < 2*minAttempts; i++ {
		allowedPods, _ = p.Process(testCtx, pendingPods)
		now = func() time.Time { return time.Now() }
	}
	assert.ElementsMatch(t, pendingPods, allowedPods)

	// Case 6: New pod doesn't reset counters for other ownerrefs
	now = func() time.Time { return time.Now().Add(-time.Hour) }
	p = NewFilterOutLongPending()
	pendingPods = buildPendingPods(1, "foo", time.Now().Add(-time.Hour))
	newPod = buildPendingPods(1, "bar", time.Now())
	pendingPods = append(pendingPods, newPod...)
	for i := 1; i < 2*minAttempts; i++ {
		allowedPods, _ = p.Process(testCtx, pendingPods)
		now = func() time.Time { return time.Now() }
	}
	assert.ElementsMatch(t, newPod, allowedPods)

	// Case 7: Forget about pods that went away
	p = NewFilterOutLongPending()
	oldPendingPods := buildPendingPods(1, "foo", now())
	newPendingPods := buildPendingPods(1, "bar", now())
	allowedPods, _ = p.Process(testCtx, oldPendingPods)
	allowedPods, _ = p.Process(testCtx, newPendingPods)
	assert.ElementsMatch(t, newPendingPods, allowedPods)
	assert.Equal(t, len(newPendingPods), len(p.seen))

}

func TestBuildDeadline(t *testing.T) {
	tests := []struct {
		name          string
		attempts      int
		coolDownDelay time.Duration
	}{

		{
			name:          "no attempts",
			attempts:      0,
			coolDownDelay: time.Hour,
		},

		{
			name:          "100 attempts",
			attempts:      100,
			coolDownDelay: time.Hour,
		},

		{
			name:     "nul delay",
			attempts: 10,
		},

		{
			name:          "negative delay",
			attempts:      10,
			coolDownDelay: -time.Hour,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			now = func() time.Time { return time.Time{} }
			maxExpected := now().Add(3 * test.coolDownDelay.Abs())
			result := buildDeadline(test.attempts, test.coolDownDelay)
			if result.Before(now()) || result.After(maxExpected) {
				t.Errorf("buildDeadline(%v, %v) = %v, should be in [%v-%v] range",
					test.attempts, test.coolDownDelay, result, now(), maxExpected)
			}
		})
	}
}

func TestMinDuration(t *testing.T) {
	tests := []struct {
		name      string
		durationA time.Duration
		durationB time.Duration
		expected  time.Duration
	}{
		{
			name:      "Both durations are equal",
			durationA: time.Second,
			durationB: time.Second,
			expected:  time.Second,
		},
		{
			name:      "Duration A is smaller than Duration B",
			durationA: time.Second,
			durationB: 2 * time.Second,
			expected:  time.Second,
		},
		{
			name:      "Duration B is smaller than Duration A",
			durationA: 2 * time.Second,
			durationB: time.Second,
			expected:  time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := minDuration(test.durationA, test.durationB)

			if result != test.expected {
				t.Errorf("minDuration(%v, %v) = %v, expected %v",
					test.durationA, test.durationB, result, test.expected)
			}
		})
	}
}

func TestJitterDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		/* this function should return a duration within [-duration/2 - 1.5 * duration],
		   and more importantly, shouldn't blow up on negative or zero arguments (since
		   rand.Int63n panic on <=0).
		*/
		{
			name:     "Zero duration",
			duration: 0,
		},
		{
			name:     "Positive duration",
			duration: time.Minute,
		},
		{
			name:     "Negative duration",
			duration: -time.Minute,
		},
		{
			name:     "Large positive duration",
			duration: 2 * time.Hour,
		},
		{
			name:     "Large negative duration",
			duration: -2 * time.Hour,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := jitterDuration(test.duration)

			maxValue := test.duration.Abs() + test.duration.Abs()/2
			if result < -test.duration/2 || result > maxValue {
				t.Errorf("jitterDuration(%v) = %v, expected duration within [%s-%s]",
					test.duration, result, -test.duration/2, maxValue)
			}
		})
	}
}

func buildPendingPods(count int, setName string, creationTime time.Time) []*apiv1.Pod {
	var result []*apiv1.Pod
	trueish := true
	for i := 0; i < count; i++ {
		result = append(result, &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				CreationTimestamp: metav1.NewTime(creationTime),
				UID:               types.UID(fmt.Sprintf("%s-%d", setName, i)),
				OwnerReferences: []metav1.OwnerReference{{
					UID:        types.UID(fmt.Sprintf("%s-%d", setName, i)),
					Name:       fmt.Sprintf("%s-%d", setName, i),
					Controller: &trueish,
				}},
			},
		})
	}
	return result
}
