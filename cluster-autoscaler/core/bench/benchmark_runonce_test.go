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

package bench

import (
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/config"
)

// Benchmark evaluates the performance of the Cluster Autoscaler's primary control loop (RunOnce).
//
// It's intended for:
// 1.  Comparative Analysis: Detect performance regressions or improvements in core logic.
// 2.  Regression Testing: Ensure key scalability metrics (time complexity) remain stable.
// 3.  Profiling: Provide a noise-free environment for CPU profiling of the RunOnce loop.
//
// To achieve stable and reproducible results, this benchmark introduces several synthetic
// conditions that differ from a production environment:
//
// -   Fake Client & Provider: API latency, network conditions, and rate limits are completely absent.
// -   Synthetic Workloads: Pods and Nodes are homogeneous or algorithmically generated, which
//     may not fully represent the complexity of real-world cluster states.
// -   Garbage Collection is DISABLED during the timed RunOnce execution. This eliminates
//     memory management noise but means results do not reflect GC overhead or pause times.
// -   klog is SILENCED to remove I/O and locking overhead. Real-world logging costs are ignored.
// -   Event Recording is a NO-OP. The cost of generating and sending events is excluded.
//
// Because of these simplifications, absolute timing numbers from this benchmark should NOT
// be interpreted as expected production latency. They are strictly relative metrics for
// comparing code versions.

func BenchmarkRunOnceScaleUp(b *testing.B) {
	s := Scenario{
		Setup:  setupScaleUp(200),
		Verify: verifyTargetSize(200),
		Config: func(opts *config.AutoscalingOptions) {
			opts.MaxNodesPerScaleUp = maxNGSize
			opts.ScaleUpFromZero = true
		},
	}
	s.Run(b)
}

func BenchmarkRunOnceScaleDown(b *testing.B) {
	s := Scenario{
		Setup:  setupScaleDown60Percent(400),
		Verify: verifyToBeDeleted(240),
		Config: func(opts *config.AutoscalingOptions) {
			opts.NodeGroupDefaults.ScaleDownUnneededTime = 0
			opts.MaxScaleDownParallelism = 1000
			opts.MaxDrainParallelism = 1000
			opts.ScaleDownDelayAfterAdd = 0
			opts.ScaleDownNonEmptyCandidatesCount = 1000
			opts.ScaleDownUnreadyEnabled = true
			opts.ScaleDownSimulationTimeout = 60 * time.Second
		},
	}
	s.Run(b)
}
