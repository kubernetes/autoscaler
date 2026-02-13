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
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/builder"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
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
//
// Because of these simplifications, absolute timing numbers from this benchmark should NOT
// be interpreted as expected production latency. They are strictly relative metrics for
// comparing code versions.

const (
	nodeCPU   = 10000
	nodeMem   = 10000
	maxNGSize = 10000
	maxCores  = nodeCPU * maxNGSize
	maxMem    = nodeMem * maxNGSize

	ngName = "ng1"
)

var runOnceCpuProfile = flag.String("profile-cpu", "", "If set, the benchmark writes a CPU profile to this file, covering the RunOnce execution during the first iteration.")

type scenario struct {
	// setup initializes the cluster state before the benchmarked RunOnce call.
	setup func(*integration.FakeSet) error
	// verify checks the cluster state after the RunOnce call to ensure correctness.
	verify func(*integration.FakeSet) error
	// config allows overriding default autoscaling options for this scenario.
	config func(*config.AutoscalingOptions)
}

// run executes the benchmark for a given scenario. It handles environment stabilization,
// profiling, and repeated execution of the RunOnce loop.
func run(b *testing.B, s scenario) {
	b.StopTimer()

	if !flag.Parsed() {
		flag.Parse()
	}

	var f *os.File
	if *runOnceCpuProfile != "" {
		var err error
		f, err = os.Create(*runOnceCpuProfile)
		if err != nil {
			b.Fatalf("Failed to create cpu profile file: %v", err)
		}
		defer f.Close()
	}

	for i := 0; i < b.N; i++ {
		cluster := newCluster()
		if err := s.setup(cluster); err != nil {
			b.Fatalf("setup failed: %v", err)
		}
		autoscaler := newAutoscaler(b, s, cluster)

		if f != nil && i == 0 {
			if err := pprof.StartCPUProfile(f); err != nil {
				b.Fatalf("Failed to start cpu profile: %v", err)
			}
		}

		b.StartTimer()
		err := autoscaler.RunOnce(time.Now().Add(10 * time.Second))
		b.StopTimer()

		if f != nil && i == 0 {
			pprof.StopCPUProfile()
		}

		if err != nil {
			b.Fatalf("RunOnce failed: %v", err)
		}

		if s.verify != nil {
			if err := s.verify(cluster); err != nil {
				b.Fatalf("verify failed: %v", err)
			}
		}
	}
}

// newCluster initializes a fake cluster with predefined resource limits.
func newCluster() *integration.FakeSet {
	cluster := integration.NewFakeSet()
	cluster.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameCores, 0, maxCores)
	cluster.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameMemory, 0, maxMem)
	return cluster
}

// newAutoscaler constructs a core.Autoscaler instance configured for the given scenario.
func newAutoscaler(b *testing.B, s scenario, cluster *integration.FakeSet) core.Autoscaler {
	opts := defaultCAOptions()
	if s.config != nil {
		s.config(&opts)
	}

	ds := debuggingsnapshot.NewDebuggingSnapshotter(false)
	mgr := integration.MustCreateControllerRuntimeMgr(b)
	a, _, err := builder.New(opts).
		WithDebuggingSnapshotter(ds).
		WithManager(mgr).
		WithKubeClient(cluster.KubeClient).
		WithInformerFactory(cluster.InformerFactory).
		WithCloudProvider(cluster.CloudProvider).
		WithPodObserver(cluster.PodObserver).Build(context.Background())
	if err != nil {
		b.Fatalf("Failed to build: %v", err)
	}
	return a
}

// defaultCAOptions returns the standard autoscaling configuration used as a baseline for benchmarks.
func defaultCAOptions() config.AutoscalingOptions {
	return config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         1 * time.Minute,
			ScaleDownUnreadyTime:          1 * time.Minute,
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodeProvisionTime:          10 * time.Minute,
		},
		EstimatorName:                  estimator.BinpackingEstimatorName,
		ExpanderNames:                  expander.LeastWasteExpanderName,
		MaxBinpackingTime:              60 * time.Second,
		MaxNodeGroupBinpackingDuration: 60 * time.Second,
		MaxCoresTotal:                  maxCores,
		MaxMemoryTotal:                 maxMem,
		MaxNodesTotal:                  maxNGSize,
		// In the benchmark we do not care about parallelism.
		// We set it to 1 to make the benchmark results more stable and reproducible.
		PredicateParallelism: 1,
	}
}

// setupScaleUp prepares a scenario that triggers a scale-up for the specified number of nodes
// by creating a corresponding number of unschedulable pods.
// Each node is designed to fit 50 pods.
//
// For simplicity, this scenario uses a single node group where both pods and nodes are homogeneous.
// The pods are created individually without a controller (e.g., Deployment or ReplicaSet),
// ensuring they do not follow the optimized scale-up path for pod groups.
func setupScaleUp(nodes int) func(*integration.FakeSet) error {
	return func(cluster *integration.FakeSet) error {
		nTemplate := BuildTestNode("n-template", nodeCPU, nodeMem)
		SetNodeReadyState(nTemplate, true, time.Now())

		cluster.CloudProvider.AddNodeGroup(ngName,
			testprovider.WithTemplate(framework.NewNodeInfo(nTemplate, nil)),
			testprovider.WithMaxSize(maxNGSize),
		)

		if _, err := cluster.KubeClient.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "default"},
		}, metav1.CreateOptions{}); err != nil {
			return err
		}

		const podsPerNode = 50
		for i := range nodes * podsPerNode {
			podName := fmt.Sprintf("pod-%d", i)
			cpu := int64(nodeCPU / podsPerNode)
			mem := int64(nodeMem / podsPerNode)
			pod := BuildTestPod(podName, cpu, mem, MarkUnschedulable())
			if _, err := cluster.KubeClient.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{}); err != nil {
				return err
			}
		}
		return nil
	}
}

// setupScaleDown60Percent prepares a scenario where the workload is reduced such
// that it fits on 40% of the existing nodes, triggering a 60% scale-down.
// Each node starts with 40 pods, each consuming 1% of the node's resources,
// resulting in 40% utilization per node.
func setupScaleDown60Percent(nodesCount int) func(*integration.FakeSet) error {
	return func(cluster *integration.FakeSet) error {
		nTemplate := BuildTestNode("n-template", nodeCPU, nodeMem)
		SetNodeReadyState(nTemplate, true, time.Now())

		cluster.CloudProvider.AddNodeGroup(ngName,
			testprovider.WithTemplate(framework.NewNodeInfo(nTemplate, nil)),
			testprovider.WithMaxSize(maxNGSize),
			testprovider.WithMinSize(0),
		)

		if _, err := cluster.KubeClient.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "default"},
		}, metav1.CreateOptions{}); err != nil {
			return err
		}

		ng := cluster.CloudProvider.GetNodeGroup(ngName)
		if err := ng.IncreaseSize(nodesCount); err != nil {
			return err
		}

		ngID := ng.Id()
		// Create 40 pods per node.
		// Each pod uses 1% of a node's resources, leading to 40% utilization.
		for i := range nodesCount * 40 {
			podName := fmt.Sprintf("pod-%d", i)
			nodeName := fmt.Sprintf("%s-node-%d", ngID, i%nodesCount)
			cpu := int64(nodeCPU / 100)
			mem := int64(nodeMem / 100)
			pod := BuildTestPod(podName, cpu, mem)
			pod.Spec.NodeName = nodeName
			if pod.Annotations == nil {
				pod.Annotations = make(map[string]string)
			}
			pod.Annotations["cluster-autoscaler.kubernetes.io/safe-to-evict"] = "true"
			if _, err := cluster.KubeClient.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{}); err != nil {
				return err
			}
		}

		return nil
	}
}

// verifyTargetSize returns a verification function that checks if the node group's target size
// matches the expected value.
func verifyTargetSize(expectedTargetSize int) func(*integration.FakeSet) error {
	return func(cluster *integration.FakeSet) error {
		ng := cluster.CloudProvider.GetNodeGroup(ngName)
		if ng == nil {
			return fmt.Errorf("nodegroup %s not found", ngName)
		}
		targetSize, err := ng.TargetSize()
		if err != nil {
			return err
		}
		if targetSize != expectedTargetSize {
			return fmt.Errorf("expected target size %d, got %d", expectedTargetSize, targetSize)
		}
		return nil
	}
}

// verifyToBeDeleted returns a verification function that checks if the number of nodes
// marked with the ToBeDeleted taint matches the expected count.
func verifyToBeDeleted(expectedDeletedSize int) func(*integration.FakeSet) error {
	return func(cluster *integration.FakeSet) error {
		nodes, err := cluster.KubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		toBeDeleted := 0
		for _, n := range nodes.Items {
			if taints.HasToBeDeletedTaint(&n) {
				toBeDeleted++
			}
		}

		if toBeDeleted == expectedDeletedSize {
			return nil
		}

		return fmt.Errorf("invalid number of deleted nodes, got: %v, expected: %v", toBeDeleted, expectedDeletedSize)
	}
}

func BenchmarkRunOnceScaleUp(b *testing.B) {
	s := scenario{
		setup:  setupScaleUp(200),
		verify: verifyTargetSize(200),
		config: func(opts *config.AutoscalingOptions) {
			opts.MaxNodesPerScaleUp = maxNGSize
			opts.ScaleUpFromZero = true
		},
	}
	run(b, s)
}

func BenchmarkRunOnceScaleDown(b *testing.B) {
	s := scenario{
		setup:  setupScaleDown60Percent(400),
		verify: verifyToBeDeleted(240),
		config: func(opts *config.AutoscalingOptions) {
			opts.NodeGroupDefaults.ScaleDownUnneededTime = 0
			opts.MaxScaleDownParallelism = 1000
			opts.MaxDrainParallelism = 1000
			opts.ScaleDownDelayAfterAdd = 0
			opts.ScaleDownEnabled = true
			opts.ScaleDownNonEmptyCandidatesCount = 1000
			opts.ScaleDownUnreadyEnabled = true
			opts.ScaleDownSimulationTimeout = 60 * time.Second
		},
	}
	run(b, s)
}
