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
	"fmt"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/builder"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/loop"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	fakek8s "k8s.io/autoscaler/cluster-autoscaler/utils/fake"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

// Benchmark evaluates the performance of the Cluster Autoscaler's primary control loop (RunOnce).
//
// The benchmark simulates a cluster ecosystem by mocking two critical layers:
// - Kubeclient (fakeClient): Acts as the Kubernetes API "source of truth," managing Pod and Node
//   objects, simulating node registration, and mimicking scheduler behavior.
// - CloudProvider (testprovider): Acts as the infrastructure abstraction, managing NodeGroups
//   and providing the capacity templates used for scale-up simulations.
//
// This tool is designed for comparative analysis between patches to detect performance
// regressions in the core autoscaling logic.
//
// Current Simplifications:
// - Homogeneous Templates: Uses uniform Node and Pod templates.
// - Discrete Iterations: Advances a virtual clock to simulate the asynchronous stabilization
//   of the cluster (e.g., node provisioning and pod placement).

type benchmarkConfig struct {
	nodeAvailableCPU    int64
	nodeAvailableMemory int64
	nodeGroupMaxSize    int
	podsPerNode         int
	maxCoresTotal       int64
	maxMemoryTotal      int64
}

func defaultConfig() benchmarkConfig {
	return benchmarkConfig{
		nodeAvailableCPU:    10000,
		nodeAvailableMemory: 10000,
		nodeGroupMaxSize:    10000,
		podsPerNode:         100,
		maxCoresTotal:       10000 * 10000,
		maxMemoryTotal:      10000 * 10000 * 1024 * 1024 * 1024,
	}
}

type benchmarkScenario struct {
	name             string
	config           benchmarkConfig
	init             func(c *fake.Clientset, p *testprovider.CloudProvider, cfg benchmarkConfig) ([]*apiv1.Node, error)
	podManipulator   func(c *fake.Clientset, cfg benchmarkConfig) error
	stopCondition    func(readyNodes int) bool
	customizeOptions func(*config.AutoscalingOptions)
}

func runBenchmark(b *testing.B, sc benchmarkScenario) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		runner := newBenchmarkRunner(b, sc)
		for {
			b.StartTimer()
			runner.stepAutoscaler()
			b.StopTimer()

			if runner.stepSimulation() {
				break
			}
		}
	}
}

type benchmarkRunner struct {
	b           *testing.B
	sc          benchmarkScenario
	fakeClient  *fake.Clientset
	autoscaler  core.Autoscaler
	currentTime time.Time
}

func (r *benchmarkRunner) stepAutoscaler() {
	err := r.autoscaler.RunOnce(r.currentTime)
	if err != nil {
		r.b.Fatalf("RunOnce failed: %v", err)
	}
}

func (r *benchmarkRunner) stepSimulation() bool {
	// Simulate nodes setup.
	// Make any new nodes instantly ready and schedulable.
	readyNodes := r.makeNodesReady()

	if r.sc.stopCondition(readyNodes) {
		return true
	}

	if err := r.sc.podManipulator(r.fakeClient, r.sc.config); err != nil {
		r.b.Fatalf("failed to manipulate pods: %v", err)
	}

	// Simulate scheduler.
	r.trySchedulePods()

	// Allow caches to populate.
	time.Sleep(100 * time.Millisecond)

	// Advance time to the next autoscaler RunOnce loop.
	r.currentTime = r.currentTime.Add(5 * time.Second)

	return false
}

func newBenchmarkRunner(b *testing.B, sc benchmarkScenario) *benchmarkRunner {
	fakeClient := fake.NewClientset()
	informerFactory := informers.NewSharedInformerFactory(fakeClient, 0)
	k8s := fakek8s.NewKubernetes(fakeClient, informerFactory)
	provider := testprovider.NewCloudProvider(k8s)
	provider.SetResourceLimit(cloudprovider.ResourceNameCores, 0, sc.config.maxCoresTotal)
	provider.SetResourceLimit(cloudprovider.ResourceNameMemory, 0, sc.config.maxMemoryTotal)

	// Setup Nodes.
	_, err := sc.init(fakeClient, provider, sc.config)
	if err != nil {
		b.Fatalf("Failed to init nodes: %v", err)
	}

	// Setup Autoscaler.
	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Minute,
			ScaleDownUtilizationThreshold: 0.5,
			MaxNodeProvisionTime:          10 * time.Second,
		},
		EstimatorName:                  estimator.BinpackingEstimatorName,
		ExpanderNames:                  "random",
		EnforceNodeGroupMinSize:        true,
		ScaleDownEnabled:               true,
		MaxNodesTotal:                  sc.config.nodeGroupMaxSize,
		MaxCoresTotal:                  sc.config.maxCoresTotal,
		MaxMemoryTotal:                 sc.config.maxMemoryTotal,
		OkTotalUnreadyCount:            1,
		MaxBinpackingTime:              1 * time.Second,
		MaxNodeGroupBinpackingDuration: 1 * time.Second,
		ScaleDownSimulationTimeout:     1 * time.Second,
		MaxScaleDownParallelism:        10,
	}

	if sc.customizeOptions != nil {
		sc.customizeOptions(&options)
	}

	debuggingSnapshotter := debuggingsnapshot.NewDebuggingSnapshotter(false)
	autoscaler, _, err := builder.New(options).
		WithKubeClient(fakeClient).
		WithInformerFactory(informerFactory).
		WithCloudProvider(provider).
		WithListerRegistry(k8s.ListerRegistry()).
		WithPodObserver(&loop.UnschedulablePodObserver{}).
		Build(context.Background(), debuggingSnapshotter)
	if err != nil {
		b.Fatalf("Failed to build autoscaler: %v", err)
	}

	return &benchmarkRunner{
		b:           b,
		sc:          sc,
		fakeClient:  fakeClient,
		autoscaler:  autoscaler,
		currentTime: time.Now(),
	}
}

func (r *benchmarkRunner) makeNodesReady() int {
	nodeList, err := r.fakeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		r.b.Fatalf("failed to list nodes: %v", err)
	}
	for _, n := range nodeList.Items {
		if !kube_util.IsNodeReadyAndSchedulable(&n) {
			updated := n.DeepCopy()
			SetNodeReadyState(updated, true, r.currentTime)
			RemoveNodeNotReadyTaint(updated)
			if _, err := r.fakeClient.CoreV1().Nodes().Update(context.Background(), updated, metav1.UpdateOptions{}); err != nil {
				r.b.Fatalf("failed to update nodes: %v", err)
			}
		}
	}
	return len(nodeList.Items)
}

// TODO: Speed up pods scheduling.
// CPU profile shows big contribution of trySchedulePods method. Reduce it.
// This might be completely "artificial" method.
func (r *benchmarkRunner) trySchedulePods() {
	allPods, err := r.fakeClient.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		r.b.Fatalf("failed to list pods: %v", err)
	}

	var unschedulablePods []*apiv1.Pod
	podsPerNode := make(map[string]int)
	for i := range allPods.Items {
		if allPods.Items[i].Spec.NodeName == "" {
			unschedulablePods = append(unschedulablePods, &allPods.Items[i])
		} else {
			podsPerNode[allPods.Items[i].Spec.NodeName]++
		}
	}

	allNodes, err := r.fakeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		r.b.Fatalf("failed to list nodes: %v", err)
	}
	var readyNodes []*apiv1.Node
	for i := range allNodes.Items {
		if kube_util.IsNodeReadyAndSchedulable(&allNodes.Items[i]) {
			readyNodes = append(readyNodes, &allNodes.Items[i])
		}
	}

	for _, node := range readyNodes {
		for podsPerNode[node.Name] < r.sc.config.podsPerNode && len(unschedulablePods) > 0 {
			var pod *apiv1.Pod
			pod, unschedulablePods = unschedulablePods[0], unschedulablePods[1:]
			podCopy := pod.DeepCopy()
			podCopy.Spec.NodeName = node.Name
			_, err := r.fakeClient.CoreV1().Pods("default").Update(context.Background(), podCopy, metav1.UpdateOptions{})
			if err != nil {
				r.b.Fatalf("Failed to schedule pod: %v", err)
			}
			podsPerNode[node.Name]++
		}
	}
}

func addPods(perCall int) func(c *fake.Clientset, cfg benchmarkConfig) error {
	added := 0
	return func(c *fake.Clientset, cfg benchmarkConfig) error {
		for i := 0; i < perCall; i++ {
			name := fmt.Sprintf("pod-%d", added)
			cpu := cfg.nodeAvailableCPU / int64(cfg.podsPerNode)
			mem := cfg.nodeAvailableMemory / int64(cfg.podsPerNode)
			pod := BuildTestPod(name, cpu, mem, MarkUnschedulable())
			added++

			_, err := c.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func initNodes(n int) func(c *fake.Clientset, p *testprovider.CloudProvider, cfg benchmarkConfig) ([]*apiv1.Node, error) {
	return func(c *fake.Clientset, p *testprovider.CloudProvider, cfg benchmarkConfig) ([]*apiv1.Node, error) {
		nodes := make([]*apiv1.Node, n)
		for j := 0; j < n; j++ {
			name := fmt.Sprintf("n-%d", j)
			node := BuildTestNode(name, cfg.nodeAvailableCPU, cfg.nodeAvailableMemory)
			SetNodeReadyState(node, true, time.Now())
			nodes[j] = node
			_, err := c.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}
		}

		nTemplate := BuildTestNode("n-template", cfg.nodeAvailableCPU, cfg.nodeAvailableMemory)
		SetNodeReadyState(nTemplate, true, time.Now())
		tni := framework.NewTestNodeInfo(nTemplate)

		p.AddNodeGroup("ng1",
			testprovider.WithTemplate(tni),
			testprovider.WithMaxSize(cfg.nodeGroupMaxSize),
		)
		for _, node := range nodes {
			p.AddNode("ng1", node)
		}

		return nodes, nil
	}
}

func stopAtNodesGreaterOrEqual(n int) func(readyNodes int) bool {
	return func(readyNodes int) bool {
		return readyNodes >= n
	}
}

func stopAtNodesLessOrEqual(n int) func(readyNodes int) bool {
	return func(readyNodes int) bool {
		return readyNodes <= n
	}
}

func BenchmarkRunOnceScaleUp500Nodes(b *testing.B) {
	sc := benchmarkScenario{
		name:           "ScaleUp500Nodes",
		config:         defaultConfig(),
		init:           initNodes(1),
		stopCondition:  stopAtNodesGreaterOrEqual(500),
		podManipulator: addPods(30), // Assuming 100 pods can fit on a node (see defaultConfig).
	}

	runBenchmark(b, sc)
}

func initNodesWithPods(n int) func(c *fake.Clientset, p *testprovider.CloudProvider, cfg benchmarkConfig) ([]*apiv1.Node, error) {
	return func(c *fake.Clientset, p *testprovider.CloudProvider, cfg benchmarkConfig) ([]*apiv1.Node, error) {
		nodes := make([]*apiv1.Node, n)
		for j := range n {
			name := fmt.Sprintf("n-%d", j)
			node := BuildTestNode(name, cfg.nodeAvailableCPU, cfg.nodeAvailableMemory)
			SetNodeReadyState(node, true, time.Now())
			nodes[j] = node
			_, err := c.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}

			// Add pods
			for k := range cfg.podsPerNode {
				podName := fmt.Sprintf("pod-%s-%d", name, k)
				cpu := cfg.nodeAvailableCPU / int64(cfg.podsPerNode)
				mem := cfg.nodeAvailableMemory / int64(cfg.podsPerNode)
				pod := BuildTestPod(podName, cpu, mem)
				pod.Spec.NodeName = name
				_, err := c.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
				if err != nil {
					return nil, err
				}
			}
		}

		nTemplate := BuildTestNode("n-template", cfg.nodeAvailableCPU, cfg.nodeAvailableMemory)
		SetNodeReadyState(nTemplate, true, time.Now())
		tni := framework.NewTestNodeInfo(nTemplate)

		p.AddNodeGroup("ng1",
			testprovider.WithTemplate(tni),
			testprovider.WithMaxSize(cfg.nodeGroupMaxSize),
		)
		for _, node := range nodes {
			p.AddNode("ng1", node)
		}

		return nodes, nil
	}
}

func removePods(perCall int) func(c *fake.Clientset, _ benchmarkConfig) error {
	return func(c *fake.Clientset, _ benchmarkConfig) error {
		pods, err := c.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		toRemove := min(perCall, len(pods.Items))
		for i := range toRemove {
			err := c.CoreV1().Pods("default").Delete(context.Background(), pods.Items[i].Name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func BenchmarkRunOnceScaleDown(b *testing.B) {
	sc := benchmarkScenario{
		name:           "ScaleDown10Nodes",
		config:         defaultConfig(),
		init:           initNodesWithPods(500),
		stopCondition:  stopAtNodesLessOrEqual(1),
		podManipulator: removePods(30),
		customizeOptions: func(opts *config.AutoscalingOptions) {
			opts.NodeGroupDefaults.ScaleDownUnneededTime = 0
			opts.NodeGroupDefaults.ScaleDownUnreadyTime = 0
			// Avoid long waits for unneeded nodes
		},
	}

	runBenchmark(b, sc)
}

// TODO: Add DRA scenario.
