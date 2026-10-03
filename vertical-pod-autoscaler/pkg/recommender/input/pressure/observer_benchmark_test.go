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
	"fmt"
	"runtime"
	"testing"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// BenchmarkObserverState reports the retained heap of the target set plus per-container detector
// state, at 110 containers per node (550,000 containers is 5,000 nodes).
func BenchmarkObserverState(b *testing.B) {
	for _, n := range []int{55_000, 550_000} {
		b.Run(fmt.Sprintf("containers=%d", n), func(b *testing.B) {
			for range b.N {
				runtime.GC()
				var before, after runtime.MemStats
				runtime.ReadMemStats(&before)
				o := NewObserver(nil, DefaultConfig())
				targets := make([]Target, 0, n)
				for i := range n {
					targets = append(targets, Target{
						ContainerID: model.ContainerID{PodID: model.PodID{Namespace: fmt.Sprint("ns-", i%50), PodName: fmt.Sprintf("app-%08d-abcde", i)}, ContainerName: "main"},
						PodUID:      fmt.Sprintf("%08d-1111-2222-3333-444455556666", i),
						Node:        fmt.Sprintf("node-%d", i/110),
						LimitBytes:  512 << 20,
					})
				}
				o.SetTargets(targets)
				now := time.Now()
				for k := range o.targets {
					o.state[k] = &containerState{lastTime: now, lastTotal: 123456789, startTime: now}
				}
				runtime.GC()
				runtime.ReadMemStats(&after)
				b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc)/float64(n), "heap-B/container")
				runtime.KeepAlive(o)
			}
		})
	}
}
