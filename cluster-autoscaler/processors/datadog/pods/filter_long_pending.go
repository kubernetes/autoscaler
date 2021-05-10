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
  * We don't want those long pending to slow down scale-up for fresh pods.
  * If one of those is a pod causing a runaway infinite upscale, we want to give
    autoscaler some slack time to recover from cooldown, and reap unused nodes.
*/

package pods

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/context"

	apiv1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

const maxDistinctWorkloadsHavingPendingPods = 50

var now = time.Now // unit tests

type pendingTracker struct {
	firstSeen time.Time
	lastTried time.Time
}

type filterOutLongPending struct {
	seen map[types.UID]*pendingTracker
}

// NewFilterOutLongPending returns a processor slowing down retries on long pending pods
func NewFilterOutLongPending() *filterOutLongPending {
	return &filterOutLongPending{
		seen: make(map[types.UID]*pendingTracker),
	}
}

// CleanUp tears down the FilterOutLongPending processor
func (p *filterOutLongPending) CleanUp() {}

// Process slows down upscales retries on long pending pods
func (p *filterOutLongPending) Process(ctx *context.AutoscalingContext, pending []*apiv1.Pod) ([]*apiv1.Pod, error) {
	longPendingBackoff := ctx.AutoscalingOptions.ScaleDownDelayAfterAdd * 2

	currentPods := make(map[types.UID]struct{}, len(pending))
	allowedPods := make([]*apiv1.Pod, 0)

	distinctWorkloadsCount := countDistinctOwnerReferences(pending)
	if distinctWorkloadsCount > maxDistinctWorkloadsHavingPendingPods {
		klog.Warningf("detected pending pods from too many (%d) distinct workloads:"+
			" disabling backoff on long pending pods", distinctWorkloadsCount)
		return pending, nil
	}

	for _, pod := range pending {
		currentPods[pod.UID] = struct{}{}
		if _, found := p.seen[pod.UID]; !found {
			p.seen[pod.UID] = &pendingTracker{
				firstSeen: now(),
				lastTried: now(),
			}
		}

		if p.seen[pod.UID].firstSeen.Add(longPendingCutoff).Before(now()) {
			deadline := p.seen[pod.UID].lastTried.Add(longPendingBackoff)
			if deadline.After(now()) {
				klog.Warningf("ignoring long pending pod %s/%s until %s",
					pod.GetNamespace(), pod.GetName(), deadline)
				continue
			}
		}
		p.seen[pod.UID].lastTried = now()

		allowedPods = append(allowedPods, pod)
	}

	for uid := range p.seen {
		if _, found := currentPods[uid]; !found {
			delete(p.seen, uid)
		}
	}

	return allowedPods, nil
}
