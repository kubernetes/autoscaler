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
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/autoscaler/cluster-autoscaler/builder"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
	k8s_testing "k8s.io/client-go/testing"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
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

const (
	nodeCPU   = 10000
	nodeMem   = 10000
	maxNGSize = 10000
	maxCores  = nodeCPU * maxNGSize
	maxMem    = nodeMem * maxNGSize

	ngName = "ng1"
)

var (
	runOnceCpuProfile = flag.String("profile-cpu", "", "If set, the benchmark writes a CPU profile to this file, covering the RunOnce execution during the first iteration.")
	withGC            = flag.Bool("gc", false, "If set to false, the benchmark disables garbage collection to stabilize the runtime.")
)

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
func (s scenario) run(b *testing.B) {
	b.StopTimer()

	if !flag.Parsed() {
		flag.Parse()
	}

	// Silence klog during benchmark to avoid output noise and focus on performance.
	// It is recommended to occasionally re-enable logging during development to verify that
	// the benchmarked code isn't producing unexpected log messages.
	klog.LogToStderr(false)
	klog.SetOutput(io.Discard)
	ctrl.SetLogger(klog.Background())

	if !*withGC {
		// Disable automatic Garbage Collection during the timed portion of the benchmark
		// to minimize variance and ensure that CPU profiles focus on the RunOnce logic.
		// This approach prioritizes identifying performance regressions in the core
		// logic over measuring absolute throughput in a production-like GC environment.
		oldGC := debug.SetGCPercent(-1)
		defer debug.SetGCPercent(oldGC)
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
		clusterFakes := newClusterFakes()
		if err := s.setup(clusterFakes); err != nil {
			b.Fatalf("setup failed: %v", err)
		}
		autoscaler := newAutoscaler(b, s, clusterFakes)

		// Manually trigger GC before the timed section to ensure a clean state
		// for each iteration.
		runtime.GC()

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
			if err := s.verify(clusterFakes); err != nil {
				b.Fatalf("verify failed: %v", err)
			}
		}
	}
}

// newClusterFakes initializes a fake cluster with predefined resource limits.
func newClusterFakes() *integration.FakeSet {
	clusterFakes := integration.NewFakeSet()
	clusterFakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameCores, 0, maxCores)
	clusterFakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameMemory, 0, maxMem)
	return clusterFakes
}

// newAutoscaler constructs a core.Autoscaler instance configured for the given scenario.
func newAutoscaler(b *testing.B, s scenario, clusterFakes *integration.FakeSet) core.Autoscaler {
	opts := defaultCAOptions()
	if s.config != nil {
		s.config(&opts)
	}

	ds := debuggingsnapshot.NewDebuggingSnapshotter(false)
	mgr := integration.MustCreateControllerRuntimeMgr(b)

	ftkc := &fastTaintingKubeClient{taintedNodes: make(map[string]bool)}
	ftkc.registerReactors(clusterFakes.KubeClient)

	kubeClients := ca_context.NewAutoscalingKubeClients(context.Background(), opts, clusterFakes.KubeClient, clusterFakes.InformerFactory)
	kubeClients.Recorder = &noOpRecorder{}

	wrappedCloudProvider := &fastScaleUpCloudProvider{
		CloudProvider: clusterFakes.CloudProvider,
	}

	a, _, err := builder.New(opts).
		WithDebuggingSnapshotter(ds).
		WithManager(mgr).
		WithKubeClient(clusterFakes.KubeClient).
		WithAutoscalingKubeClients(kubeClients).
		WithInformerFactory(clusterFakes.InformerFactory).
		WithCloudProvider(wrappedCloudProvider).
		WithPodObserver(clusterFakes.PodObserver).Build(context.Background())
	if err != nil {
		b.Fatalf("Failed to build: %v", err)
	}
	return a
}

// noOpRecorder is a dummy implementation of record.EventRecorder that discards all events.
// Benchmark workloads generate a lot of events in a short period
// which results in event drops and noise in the logs.
type noOpRecorder struct{}

func (n *noOpRecorder) Event(_ apimachineryruntime.Object, _, _, _ string)                    {}
func (n *noOpRecorder) Eventf(_ apimachineryruntime.Object, _, _, _ string, _ ...interface{}) {}
func (n *noOpRecorder) AnnotatedEventf(_ apimachineryruntime.Object, _ map[string]string, _, _, _ string, _ ...interface{}) {
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
		// In a homogeneous benchmark environment, any node is a valid fit.
		// Higher parallelism causes a race condition where multiple workers perform
		// redundant Filter checks before the first success can trigger cancellation.
		// This introduces lock contention on CycleState and massive variance (±20%)
		// depending on Go scheduler non-determinism. We set it to 1 to ensure
		// deterministic, sequential evaluation and stable profiling results.
		PredicateParallelism: 1,
	}
}

// fastScaleUpNodeGroup does not simulate a real scale up by creating new Node objects.
// Instead, it only increases the artificial targetSize counter in the fake cloud provider.
// This is used in benchmarks to eliminate the noise introduced by node object creation
// and management in the fake cloud provider, which is not relevant for evaluating
// the autoscaler's scale-up logic and target size calculations.
type fastScaleUpNodeGroup struct {
	*testprovider.NodeGroup
}

func (f *fastScaleUpNodeGroup) IncreaseSize(delta int) error {
	return f.DecreaseTargetSize(-delta)
}

// fastScaleUpCloudProvider is a wrapper around the fake cloud provider that uses
// fastScaleUpNodeGroup to bypass node creation during scale-up, reducing CPU profile noise.
type fastScaleUpCloudProvider struct {
	*testprovider.CloudProvider
}

func (f *fastScaleUpCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	groups := f.CloudProvider.NodeGroups()
	result := make([]cloudprovider.NodeGroup, len(groups))
	for i, g := range groups {
		ng := g.(*testprovider.NodeGroup)
		fg := &fastScaleUpNodeGroup{NodeGroup: ng}
		result[i] = fg
	}
	return result
}

func (f *fastScaleUpCloudProvider) GetNodeGroup(id string) *fastScaleUpNodeGroup {
	g := f.CloudProvider.GetNodeGroup(id)
	if g == nil {
		return nil
	}
	ng := g.(*testprovider.NodeGroup)
	fg := &fastScaleUpNodeGroup{NodeGroup: ng}
	return fg
}

// fastTaintingKubeClient tracks nodes that were artificially tainted in-memory to bypass
// the overhead of full API server round-trips and complex object tracking in the fake client.
type fastTaintingKubeClient struct {
	mu           sync.Mutex
	taintedNodes map[string]bool
}

func (c *fastTaintingKubeClient) registerReactors(client *fake.Clientset) {
	fake := &client.Fake
	fake.PrependReactor("update", "nodes", func(action k8s_testing.Action) (handled bool, ret apimachineryruntime.Object, err error) {
		node := action.(k8s_testing.UpdateAction).GetObject().(*corev1.Node)
		if taints.HasToBeDeletedTaint(node) {
			c.mu.Lock()
			c.taintedNodes[node.Name] = true
			c.mu.Unlock()
		}
		return true, node, nil
	})

	fake.PrependReactor("get", "nodes", func(action k8s_testing.Action) (handled bool, ret apimachineryruntime.Object, err error) {
		name := action.(k8s_testing.GetAction).GetName()
		c.mu.Lock()
		isTainted := c.taintedNodes[name]
		c.mu.Unlock()

		if !isTainted {
			return false, nil, nil
		}

		obj, err := client.Tracker().Get(action.GetResource(), action.GetNamespace(), name)
		if err != nil {
			return true, nil, err
		}
		node := obj.(*corev1.Node).DeepCopy()
		if !taints.HasToBeDeletedTaint(node) {
			node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
				Key:    taints.ToBeDeletedTaint,
				Value:  fmt.Sprint(time.Now().Unix()),
				Effect: corev1.TaintEffectNoSchedule,
			})
		}
		return true, node, nil
	})

	fake.PrependReactor("list", "nodes", func(action k8s_testing.Action) (handled bool, ret apimachineryruntime.Object, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		gvk := schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"}
		obj, err := client.Tracker().List(gvr, gvk, ns)
		if err != nil {
			return true, nil, err
		}
		list := obj.(*corev1.NodeList).DeepCopy()
		c.mu.Lock()
		defer c.mu.Unlock()
		for i := range list.Items {
			node := &list.Items[i]
			if c.taintedNodes[node.Name] {
				if !taints.HasToBeDeletedTaint(node) {
					node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
						Key:    taints.ToBeDeletedTaint,
						Value:  fmt.Sprint(time.Now().Unix()),
						Effect: corev1.TaintEffectNoSchedule,
					})
				}
			}
		}
		return true, list, nil
	})
}

// setupScaleUp prepares a scenario that triggers a scale-up for the specified number of nodes
// by creating a corresponding number of unschedulable pods.
// Each node is designed to fit 50 pods.
//
// For simplicity, this scenario uses a single node group where both pods and nodes are homogeneous.
// The pods are created individually without a controller (e.g., Deployment or ReplicaSet),
// ensuring they do not follow the optimized scale-up path for pod groups.
func setupScaleUp(nodes int) func(*integration.FakeSet) error {
	return func(clusterFakes *integration.FakeSet) error {
		nTemplate := BuildTestNode("n-template", nodeCPU, nodeMem)
		SetNodeReadyState(nTemplate, true, time.Now())

		clusterFakes.CloudProvider.AddNodeGroup(ngName,
			testprovider.WithTemplate(framework.NewNodeInfo(nTemplate, nil)),
			testprovider.WithNGSize(0, maxNGSize),
		)

		const podsPerNode = 50
		for i := range nodes * podsPerNode {
			podName := fmt.Sprintf("pod-%d", i)
			cpu := int64(nodeCPU / podsPerNode)
			mem := int64(nodeMem / podsPerNode)
			pod := BuildTestPod(podName, cpu, mem, MarkUnschedulable())
			clusterFakes.K8s.AddPod(pod)
		}
		return nil
	}
}

// setupScaleDown60Percent prepares a scenario where the workload is reduced such
// that it fits on 40% of the existing nodes, triggering a 60% scale-down.
// Each node starts with 40 pods, each consuming 1% of the node's resources,
// resulting in 40% utilization per node.
func setupScaleDown60Percent(nodesCount int) func(*integration.FakeSet) error {
	return func(clusterFakes *integration.FakeSet) error {
		nTemplate := BuildTestNode("n-template", nodeCPU, nodeMem)
		SetNodeReadyState(nTemplate, true, time.Now())

		clusterFakes.CloudProvider.AddNodeGroup(ngName,
			testprovider.WithTemplate(framework.NewNodeInfo(nTemplate, nil)),
			testprovider.WithNGSize(0, maxNGSize),
		)

		ng := clusterFakes.CloudProvider.GetNodeGroup(ngName)
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
			clusterFakes.K8s.AddPod(pod)
		}

		return nil
	}
}

// verifyTargetSize returns a verification function that checks if the node group's target size
// matches the expected value.
func verifyTargetSize(expectedTargetSize int) func(*integration.FakeSet) error {
	return func(clusterFakes *integration.FakeSet) error {
		ng := clusterFakes.CloudProvider.GetNodeGroup(ngName)
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
	return func(clusterFakes *integration.FakeSet) error {
		nodes := clusterFakes.K8s.Nodes()
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
	s.run(b)
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
			opts.ScaleDownNonEmptyCandidatesCount = 1000
			opts.ScaleDownUnreadyEnabled = true
			opts.ScaleDownSimulationTimeout = 60 * time.Second
		},
	}
	s.run(b)
}
