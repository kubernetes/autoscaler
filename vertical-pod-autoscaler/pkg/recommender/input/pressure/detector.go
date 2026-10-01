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
	"time"

	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	metrics_recommender "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
)

// maxFutureSkew is how far ahead of the local clock a sample may be timestamped.
const maxFutureSkew = 5 * time.Second

// summary is the part of the kubelet Summary API (k8s.io/kubelet/pkg/apis/stats/v1alpha1) the detector
// reads. It is declared here because VPA does not depend on the kubelet module.
type summary struct {
	Pods []podStats `json:"pods"`
}

type podStats struct {
	PodRef struct {
		UID string `json:"uid"`
	} `json:"podRef"`
	Containers []containerStats `json:"containers"`
}

type containerStats struct {
	Name      string       `json:"name"`
	StartTime time.Time    `json:"startTime"`
	Memory    *memoryStats `json:"memory"`
}

type memoryStats struct {
	Time            time.Time `json:"time"`
	WorkingSetBytes *uint64   `json:"workingSetBytes"`
	PSI             *psiStats `json:"psi"`
}

type psiStats struct {
	Some struct {
		Total uint64 `json:"total"`
	} `json:"some"`
}

// containerState is the detector state for one container.
type containerState struct {
	lastTime      time.Time
	lastTotal     uint64
	startTime     time.Time
	limit         int64
	qualifying    int
	cooldownUntil time.Time
	// lastClass is the container's pressure_containers state in the latest window, or "".
	lastClass string
}

func newBaseline(sampleTime time.Time, total uint64, startTime time.Time, limit int64) *containerState {
	return &containerState{lastTime: sampleTime, lastTotal: total, startTime: startTime, limit: limit}
}

func (o *KubeletObserver) sampleFresh(sampleTime, now time.Time) bool {
	age := now.Sub(sampleTime)
	return age <= o.cfg.MaxSampleAge && age >= -maxFutureSkew
}

// process applies the detection rule to every eligible container in a node's Summary: two consecutive
// windows with a stall rate above StallRatio and a working set at or above WorkingSetFraction of the
// limit, outside startup grace, cooldown and any outstanding resize.
func (o *KubeletObserver) process(s summary) {
	o.mu.Lock()
	defer o.mu.Unlock()
	now := o.clock.Now()
	for _, p := range s.Pods {
		for _, c := range p.Containers {
			key := targetKey{podUID: p.PodRef.UID, container: c.Name}
			t, ok := o.targets[key]
			if !ok || c.Memory == nil || c.Memory.PSI == nil || c.Memory.WorkingSetBytes == nil || t.LimitBytes <= 0 {
				continue
			}
			total := c.Memory.PSI.Some.Total
			// Check freshness before anything is stored: a future-dated sample used as a baseline would make
			// every later valid sample look older and be ignored.
			if !o.sampleFresh(c.Memory.Time, now) {
				klog.V(3).InfoS("Pressure sample outside freshness bounds, discarding", "container", t.ContainerID, "age", now.Sub(c.Memory.Time).Round(time.Second))
				metrics_recommender.RecordPressureEvent("stale")
				delete(o.state, key)
				continue
			}
			st := o.state[key]
			// A new container or a resized limit starts a new baseline: windows measured against the old
			// limit say nothing about the new one.
			identityChanged := st != nil && (!st.startTime.Equal(c.StartTime) || st.limit != t.LimitBytes)
			if identityChanged {
				metrics_recommender.RecordPressureEvent("identity_changed")
			}
			if st == nil || identityChanged || total < st.lastTotal || c.Memory.Time.Sub(st.lastTime) > 2*o.cfg.FetchInterval {
				o.state[key] = newBaseline(c.Memory.Time, total, c.StartTime, t.LimitBytes)
				continue
			}
			if !c.Memory.Time.After(st.lastTime) {
				continue
			}
			dt := c.Memory.Time.Sub(st.lastTime)
			rate := float64(total-st.lastTotal) * float64(o.cfg.PSITotalUnit) / float64(dt)
			// A stall rate cannot exceed 1. Summary timestamps have one-second resolution, so allow 1s/dt
			// of slack; anything above that is a unit mismatch.
			if rate > 1+float64(time.Second)/float64(dt) {
				klog.V(3).InfoS("Pressure rate out of range, discarding and resetting baseline", "container", t.ContainerID, "stallRate", rate)
				metrics_recommender.RecordPressureEvent("rate_out_of_range")
				o.state[key] = newBaseline(c.Memory.Time, total, c.StartTime, t.LimitBytes)
				continue
			}
			ws := float64(*c.Memory.WorkingSetBytes)
			frac := ws / float64(t.LimitBytes)
			st.lastTime, st.lastTotal = c.Memory.Time, total
			qualifies := rate > o.cfg.StallRatio && frac >= o.cfg.WorkingSetFraction
			if qualifies {
				st.qualifying++
			} else {
				st.qualifying = 0
			}
			klog.V(5).InfoS("Pressure sample", "container", t.ContainerID, "stallRate", rate, "wsFraction", frac, "qualifying", st.qualifying, "resizeState", t.ResizeState)
			switch {
			case !qualifies:
				st.lastClass = ""
			case t.ResizeState != "":
				st.lastClass = t.ResizeState
			case now.Before(st.cooldownUntil):
				st.lastClass = "cooldown"
			default:
				st.lastClass = "injectable"
			}
			if t.ResizeState != "" {
				if qualifies {
					metrics_recommender.RecordPressureEvent("resize_pending")
				}
				st.qualifying = 0
				continue
			}
			if st.qualifying < 2 || now.Before(st.cooldownUntil) || now.Sub(c.StartTime) <= o.cfg.StartupGrace {
				continue
			}
			select {
			case o.ch <- Info{ContainerID: t.ContainerID, PodUID: t.PodUID, SampleTime: c.Memory.Time, WorkingSet: model.MemoryAmountFromBytes(ws), StallRate: rate}:
				klog.V(3).InfoS("Pressure event emitted", "container", t.ContainerID, "stallRate", rate, "wsFraction", frac)
				st.cooldownUntil = now.Add(o.cfg.Cooldown)
			default:
				klog.V(3).InfoS("Pressure queue full, dropping observation", "container", t.ContainerID)
				metrics_recommender.RecordPressureEvent("queue_full")
			}
			st.qualifying = 0
		}
	}
}
