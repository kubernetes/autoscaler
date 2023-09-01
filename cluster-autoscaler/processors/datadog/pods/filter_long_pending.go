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

/*
  Ensures pods pending for a long time are retried at a slower pace:
  * We don't want those long pending to slow down scale-up for new pods.
  * If one of those is causing a runaway infinite upscale, we want to give
    autoscaler some slack time to recover from cooldown, and reap unused nodes.
  * For autoscaling efficiency, we prefer skipping or keeping all pods from
    a given workload (all pods sharing a common controller ref).

  We quarantine (remove from next upscale attempt) for a progressively
  longer duration (up to a couple of scaledown cooldown duration + jitter)
  all pods from workloads that didn't had new pods recently (all that
  workload's pods are "old"), and were already considered several times
  for upscale.
*/

package pods

import (
	"math/rand"
	"os"
	"strconv"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	klog "k8s.io/klog/v2"
)

var (
	now              = time.Now // unit tests
	retriesIncrement = 2 * time.Minute

	// we won't quarantine a workload's pods until they all went
	// through (were considered for upscales) minAttempts times
	minAttempts = getEnv("QUARANTINE_MIN_ATTEMPTS", 2)

	// how many multiple of coolDownDelay will "long pending"
	// pods be delayed (at most).
	delayFactor = getEnv("QUARANTINE_DELAY_FACTOR", 3)
)

type pendingTracker struct {
	// last time we saw a new pod for that workload (or most recent
	// pod's creation time, whichever is the most recent)
	lastSeen time.Time
	// deadline we won't attempt upscale for that workload's pods until reached
	nextTry time.Time
	// how many times that pods went through unfiltered upscale attempts
	attempts int
}

type filterOutLongPending struct {
	seen         map[types.UID]*pendingTracker
	deadlineFunc func(int, time.Duration) time.Time
}

// NewFilterOutLongPending returns a processor slowing down retries on long pending pods
func NewFilterOutLongPending() *filterOutLongPending {
	rand.Seed(time.Now().UnixNano())
	return &filterOutLongPending{
		seen:         make(map[types.UID]*pendingTracker),
		deadlineFunc: buildDeadline,
	}
}

// CleanUp tears down the FilterOutLongPending processor
func (p *filterOutLongPending) CleanUp() {}

// Process slows down upscales retries on long pending pods
func (p *filterOutLongPending) Process(ctx *context.AutoscalingContext, pending []*apiv1.Pod) ([]*apiv1.Pod, error) {
	coolDownDelay := ctx.AutoscalingOptions.ScaleDownDelayAfterAdd
	allowedPods := make([]*apiv1.Pod, 0)
	refsToPods := make(map[types.UID][]*apiv1.Pod)

	for _, pod := range pending {
		refID := pod.GetUID()
		controllerRef := metav1.GetControllerOf(pod)
		if controllerRef != nil {
			refID = controllerRef.UID
		}

		// ephemeral ownerrefs -> pods mapping
		if _, ok := refsToPods[refID]; !ok {
			refsToPods[refID] = make([]*apiv1.Pod, 0)
		}
		refsToPods[refID] = append(refsToPods[refID], pod)

		// track upscale attempts for a given ownerref
		if _, ok := p.seen[refID]; !ok {
			p.seen[refID] = &pendingTracker{
				lastSeen: now(),
			}
		}

		// reset counters whenever a workload has new pods
		if p.seen[refID].lastSeen.Before(pod.GetCreationTimestamp().Time) {
			p.seen[refID] = &pendingTracker{
				lastSeen: pod.GetCreationTimestamp().Time,
			}
		}
	}

	for ref, pods := range refsToPods {
		// skip (don't attempt upscales for) quarantined pods
		if p.seen[ref].nextTry.After(now()) {
			logSkipped(ref, p.seen[ref], pods)
			continue
		}

		// let autoscaler attempt upscales for non-quarantined pods
		allowedPods = append(allowedPods, pods...)
		p.seen[ref].attempts++

		// don't quarantine pods until they had several upscale opportunities
		if p.seen[ref].attempts < minAttempts {
			continue
		}

		// don't quarantine workloads having recent pods
		if p.seen[ref].lastSeen.Add(longPendingCutoff).After(now()) {
			continue
		}

		// pending pods from that workload already had several upscale attempts
		// opportunities, and aren't recent: will delay after current loop.
		p.seen[ref].nextTry = p.deadlineFunc(p.seen[ref].attempts, coolDownDelay)
	}

	// reap stale entries
	for ref := range p.seen {
		if _, ok := refsToPods[ref]; !ok {
			delete(p.seen, ref)
		}
	}

	return allowedPods, nil
}

func logSkipped(ref types.UID, tracker *pendingTracker, pods []*apiv1.Pod) {
	klog.Warningf(
		"ignoring %d long pending pod(s) (eg. %s/%s) "+
			"from workload uid %q until %q (attempted %d times)",
		len(pods), pods[0].GetNamespace(), pods[0].GetName(),
		ref, tracker.nextTry, tracker.attempts,
	)
}

func buildDeadline(attempts int, coolDownDelay time.Duration) time.Time {
	increment := time.Duration(attempts-minAttempts) * retriesIncrement
	delay := minDuration(increment, time.Duration(delayFactor)*coolDownDelay.Abs())
	delay -= jitterDuration(coolDownDelay.Abs())
	return now().Add(delay.Abs())
}

func minDuration(a, b time.Duration) time.Duration {
	if a <= b {
		return a
	}
	return b
}

func jitterDuration(duration time.Duration) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(duration.Abs() + 1)))
	return jitter - duration/2
}

func getEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
