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

package pressure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math"
	"math/rand/v2"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
)

const (
	minBackoff        = 5 * time.Second
	latencyWindowSize = 500
)

// sweep fetches every eligible node once, within SweepTimeout. Nodes whose request the previous
// sweep's deadline cancelled go first; the rest follow a cursor that resumes after the last node the
// previous sweep completed, so a deadline cannot starve the same nodes every sweep. Nodes in backoff
// are skipped.
func (o *KubeletObserver) sweep(ctx context.Context) {
	start := o.clock.Now()
	plan := o.planSweep(start)

	sctx, cancel := context.WithTimeout(ctx, o.cfg.SweepTimeout)
	defer cancel()
	work := make(chan string)
	var done sync.WaitGroup
	var succeeded atomic.Int64
	for range plan.concurrency {
		done.Go(func() {
			for node := range work {
				if o.fetchOne(sctx, node) {
					succeeded.Add(1)
				}
			}
		})
	}
	sent := 0
feed:
	for _, n := range plan.fetch {
		select {
		case work <- n:
			sent++
		case <-sctx.Done():
			break feed
		}
	}
	close(work)
	done.Wait()

	undersampled := o.finishSweep(plan)
	klog.V(3).InfoS("Pressure sweep", "eligibleNodes", len(plan.all), "attempted", sent, "succeeded", succeeded.Load(),
		"concurrency", plan.concurrency, "duration", o.clock.Since(start).Round(time.Millisecond), "undersampled", undersampled,
		"inBackoff", len(plan.inBackoff))
}

// sweepPlan is the node order and worker count for one sweep.
type sweepPlan struct {
	// all is every eligible node, sorted.
	all []string
	// rotated is all, starting at the cursor, without the nodes carried from the previous sweep.
	rotated []string
	// fetch is the carried nodes followed by rotated, without the nodes in backoff.
	fetch       []string
	inBackoff   map[string]bool
	concurrency int
}

// planSweep orders the eligible nodes for one sweep and forgets scheduling state for nodes that are no
// longer eligible, so that the per-node maps stay bounded by the eligible node count.
func (o *KubeletObserver) planSweep(now time.Time) sweepPlan {
	o.mu.Lock()
	eligible := map[string]bool{}
	for _, t := range o.targets {
		eligible[t.Node] = true
	}
	o.mu.Unlock()
	plan := sweepPlan{all: slices.Sorted(maps.Keys(eligible)), inBackoff: map[string]bool{}}

	o.schedMu.Lock()
	defer o.schedMu.Unlock()
	maps.DeleteFunc(o.lastOK, func(n string, _ time.Time) bool { return !eligible[n] })
	maps.DeleteFunc(o.backoff, func(n string, _ backoffState) bool { return !eligible[n] })
	maps.DeleteFunc(o.nodeFlag, func(n string, _ string) bool { return !eligible[n] })

	carried := map[string]bool{}
	for _, n := range o.carry {
		if eligible[n] && !carried[n] {
			carried[n] = true
			plan.fetch = append(plan.fetch, n)
		}
	}
	o.carry = nil
	o.done = map[string]bool{}
	if len(plan.all) > 0 {
		o.cursor %= len(plan.all)
	}
	for _, n := range slices.Concat(plan.all[o.cursor:], plan.all[:o.cursor]) {
		if !carried[n] {
			plan.rotated = append(plan.rotated, n)
		}
	}
	for _, n := range plan.rotated {
		if b, ok := o.backoff[n]; ok && now.Before(b.until) {
			plan.inBackoff[n] = true
			continue
		}
		plan.fetch = append(plan.fetch, n)
	}
	plan.concurrency = o.deriveConcurrencyLocked(len(plan.fetch))
	return plan
}

// finishSweep advances the cursor past the leading rotated nodes that completed or were skipped, sets
// the node gauges and returns the number of undersampled nodes.
func (o *KubeletObserver) finishSweep(plan sweepPlan) int {
	now := o.clock.Now()
	o.schedMu.Lock()
	advance := 0
	for _, n := range plan.rotated {
		if !plan.inBackoff[n] && !o.done[n] {
			break
		}
		advance++
	}
	if len(plan.all) > 0 {
		o.cursor = (o.cursor + advance) % len(plan.all)
	}
	states := map[string]int{}
	for _, n := range plan.all {
		if last, ok := o.lastOK[n]; !ok || now.Sub(last) > 2*o.cfg.FetchInterval {
			states["undersampled"]++
		}
		switch {
		case o.backoff[n].until.After(now):
			states["failed"]++
		case o.nodeFlag[n] != "":
			states[o.nodeFlag[n]]++
		case now.Sub(o.lastOK[n]) <= 2*o.cfg.FetchInterval:
			states["observed"]++
		default:
			// Already counted as undersampled.
		}
	}
	o.schedMu.Unlock()
	metrics_recommender.SetPressureNodes(states)

	byClass := map[string]int{}
	o.mu.Lock()
	for _, st := range o.state {
		if st.lastClass != "" {
			byClass[st.lastClass]++
		}
	}
	o.mu.Unlock()
	metrics_recommender.SetPressureContainers(byClass)
	return states["undersampled"]
}

// deriveConcurrencyLocked sizes the worker pool so a sweep of n nodes fits the sweep deadline at the
// observed p95 fetch latency: ceil(n * p95 / SweepTimeout), clamped to [1, MaxConcurrent].
func (o *KubeletObserver) deriveConcurrencyLocked(n int) int {
	p95 := 100 * time.Millisecond
	if len(o.latencies) >= 20 {
		l := slices.Clone(o.latencies)
		slices.Sort(l)
		p95 = l[len(l)*95/100]
	}
	c := int(math.Ceil(float64(n) * float64(p95) / float64(o.cfg.SweepTimeout)))
	return min(max(c, 1), o.cfg.MaxConcurrent)
}

// fetchOne fetches and processes one node's Summary. It holds no lock during I/O or processing.
func (o *KubeletObserver) fetchOne(sweepCtx context.Context, node string) bool {
	klog.V(5).InfoS("Pressure fetch", "node", node, "transport", o.cfg.Transport)
	start := o.clock.Now()
	rctx, cancel := context.WithTimeout(sweepCtx, o.cfg.RequestTimeout)
	defer cancel()
	s, err := o.fetchSummary(rctx, node)
	latency := o.clock.Since(start)

	result := "success"
	switch {
	case err == nil:
	case apierrors.IsTooManyRequests(err):
		result = "throttled"
	case rctx.Err() != nil:
		result = "timeout"
	default:
		result = "error"
	}
	metrics_recommender.ObservePressureFetch(result, latency.Seconds())

	if err != nil {
		o.recordFailure(node, err, sweepCtx.Err() != nil)
		return false
	}
	o.recordSuccess(node, latency, o.nodeFlagFor(s))
	o.process(s)
	return true
}

func (o *KubeletObserver) fetchSummary(ctx context.Context, node string) (summary, error) {
	var s summary
	body, err := o.openSummary(ctx, node)
	if err != nil {
		return s, err
	}
	defer func() { _ = body.Close() }()
	// Decode while reading instead of buffering the whole response; one byte past the limit marks an
	// oversized response.
	limited := &io.LimitedReader{R: body, N: o.cfg.MaxResponseBytes + 1}
	err = json.NewDecoder(limited).Decode(&s)
	if limited.N <= 0 {
		return s, fmt.Errorf("response exceeds %d bytes", o.cfg.MaxResponseBytes)
	}
	return s, err
}

// recordFailure carries a node the sweep deadline cancelled to the next sweep, without backoff, and
// backs off any other failing node exponentially with jitter, honouring Retry-After.
func (o *KubeletObserver) recordFailure(node string, err error, cancelledBySweep bool) {
	o.schedMu.Lock()
	defer o.schedMu.Unlock()
	if cancelledBySweep {
		o.carry = append(o.carry, node)
		klog.V(4).InfoS("Pressure fetch cancelled by sweep deadline, carried to next sweep", "node", node)
		return
	}
	o.done[node] = true
	d := max(o.backoff[node].delay*2, minBackoff)
	if secs, ok := apierrors.SuggestsClientDelay(err); ok {
		d = max(d, time.Duration(secs)*time.Second)
	}
	d = time.Duration(float64(min(d, o.cfg.MaxBackoff)) * (0.8 + 0.4*rand.Float64()))
	o.backoff[node] = backoffState{until: o.clock.Now().Add(d), delay: d}
	klog.V(3).InfoS("Pressure fetch failed", "node", node, "error", err, "backoff", d.Round(time.Second))
}

func (o *KubeletObserver) recordSuccess(node string, latency time.Duration, flag string) {
	now := o.clock.Now()
	o.schedMu.Lock()
	defer o.schedMu.Unlock()
	o.done[node] = true
	delete(o.backoff, node)
	if prev, ok := o.lastOK[node]; ok {
		metrics_recommender.ObservePressureNodeRevisit(now.Sub(prev).Seconds())
	}
	o.lastOK[node] = now
	if flag == "" {
		delete(o.nodeFlag, node)
	} else {
		o.nodeFlag[node] = flag
	}
	o.latencies = append(o.latencies, latency)
	if len(o.latencies) > latencyWindowSize {
		o.latencies = o.latencies[len(o.latencies)-latencyWindowSize:]
	}
}

// nodeFlagFor returns missing_psi or stale when any container in the Summary lacks PSI or carries a
// sample outside the freshness bounds, and "" otherwise.
func (o *KubeletObserver) nodeFlagFor(s summary) string {
	now := o.clock.Now()
	for _, p := range s.Pods {
		for _, c := range p.Containers {
			if c.Memory == nil || c.Memory.PSI == nil {
				return "missing_psi"
			}
			if !o.sampleFresh(c.Memory.Time, now) {
				return "stale"
			}
		}
	}
	return ""
}
