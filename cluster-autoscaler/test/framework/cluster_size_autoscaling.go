/*
Copyright 2020 The Kubernetes Authors.

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

package framework

import (
	"context"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
	e2erc "k8s.io/kubernetes/test/e2e/framework/rc"
)

const (
	defaultTimeout         = 3 * time.Minute
	manualResizeTimeout    = 6 * time.Minute
	scaleUpTimeout         = 5 * time.Minute
	scaleUpTriggerTimeout  = 2 * time.Minute
	scaleDownTimeout       = 20 * time.Minute
	podTimeout             = 2 * time.Minute
	rcCreationRetryTimeout = 4 * time.Minute
	rcCreationRetryDelay   = 20 * time.Second
	makeSchedulableTimeout = 10 * time.Minute
	makeSchedulableDelay   = 20 * time.Second
	freshStatusLimit       = 20 * time.Second

	disabledTaint             = "DisabledForAutoscalingTest"
	criticalAddonsOnlyTaint   = "CriticalAddonsOnly"
	newNodesForScaledownTests = 2

	caNoScaleUpStatus      = "NoActivity"
	caOngoingScaleUpStatus = "InProgress"
	timestampFormat        = "2006-01-02 15:04:05 -0700 MST"

	expendablePriorityClassName = "expendable-priority"
	highPriorityClassName       = "high-priority"
)

var (
	kubeconfig        = flag.String(clientcmd.RecommendedConfigPathFlag, "", "Path to kubeconfig containing embedded authinfo.")
	nodeInstanceGroup = flag.String("node-instance-group", "", "Name of the managed instance group for nodes")
)

func ClusterAutoscalerSuite(t *testing.T, provider Provider) {
	spec.Run(t, "Cluster size autoscaling [Slow]", func(t *testing.T, when spec.G, it spec.S) {
		var (
			nodeCount        int
			memAllocatableMb int
			originalSizes    map[string]int
			f                = &Framework{
				T:        t,
				Provider: provider,
			}
		)

		it.Before(func() {
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			if kubeconfig != nil {
				loadingRules.ExplicitPath = *kubeconfig
			}

			f.ClientConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

			config, err := f.ClientConfig.ClientConfig()
			require.NoError(t, err)

			f.ClientSet, err = clientset.NewForConfig(config)
			require.NoError(t, err)

			namespace := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "autoscaling-",
				},
			}
			ns, err := f.ClientSet.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
			require.NoError(t, err)
			f.Namespace = ns

			provider.FrameworkBeforeEach(f)

			originalSizes = make(map[string]int)
			sum := 0
			for _, mig := range strings.Split(*nodeInstanceGroup, ",") {
				size, err := provider.GroupSize(mig)
				require.NoError(t, err)
				t.Logf("Initial size of %s: %d", mig, size)
				originalSizes[mig] = size
				sum += size
			}
			// Give instances time to spin up
			require.NoError(t, e2enode.WaitForReadyNodes(f.ClientSet, sum, scaleUpTimeout))

			nodes, err := e2enode.GetReadySchedulableNodes(f.ClientSet)
			require.NoError(t, err)
			nodeCount = len(nodes.Items)
			t.Logf("Initial number of schedulable nodes: %v", nodeCount)
			require.NotEqual(t, nodeCount, 0)
			mem := nodes.Items[0].Status.Allocatable[v1.ResourceMemory]
			memAllocatableMb = int((&mem).Value() / 1024 / 1024)
			t.Logf("Using node %q to determine that we have %dMi of memory on each node", nodes.Items[0].Name, memAllocatableMb)

			require.Equal(t, nodeCount, sum)

			require.NoError(t, provider.EnableAutoscaler("default-pool", 3, 5))
		})

		it.After(func() {
			provider.FrameworkAfterEach(f)
			require.NoError(t, provider.DisableAutoscaler("default-pool"))

			err := f.ClientSet.CoreV1().Namespaces().Delete(context.TODO(), f.Namespace.Name, metav1.DeleteOptions{})
			require.NoError(t, err)

			t.Log("Restoring initial size of the cluster")
			setMigSizes(f, originalSizes)
			expectedNodes := 0
			for _, size := range originalSizes {
				expectedNodes += size
			}
			require.NoError(t, e2enode.WaitForReadyNodes(f.ClientSet, expectedNodes, scaleDownTimeout))
			nodes, err := f.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			require.NoError(t, err)

			s := time.Now()
		makeSchedulableLoop:
			for start := time.Now(); time.Since(start) < makeSchedulableTimeout; time.Sleep(makeSchedulableDelay) {
				for i := range nodes.Items {
					err = makeNodeSchedulable(f, &nodes.Items[i], true)
					switch err.(type) {
					case CriticalAddonsOnlyError:
						continue makeSchedulableLoop
					default:
						require.NoError(t, err)
					}
				}
				break
			}
			f.T.Logf("Made nodes schedulable again in %v", time.Since(s).String())
		})

		it("shouldn't increase cluster size if pending pod is too large [Feature:ClusterSizeAutoscalingScaleUp]", func() {
			t.Log("Creating unschedulable pod")
			ReserveMemory(f, "memory-reservation", 1, int(1.1*float64(memAllocatableMb)), false, defaultTimeout)
			defer e2erc.DeleteRCAndWaitForGC(f.ClientSet, f.Namespace.Name, "memory-reservation")

			t.Log("Waiting for scale up hoping it won't happen")
			// Verify that the appropriate event was generated
			eventFound := false
		EventsLoop:
			for start := time.Now(); time.Since(start) < scaleUpTimeout; time.Sleep(20 * time.Second) {
				t.Log("Waiting for NotTriggerScaleUp event")
				events, err := f.ClientSet.CoreV1().Events(f.Namespace.Name).List(context.TODO(), metav1.ListOptions{})
				require.NoError(t, err)

				for _, e := range events.Items {
					if e.InvolvedObject.Kind == "Pod" && e.Reason == "NotTriggerScaleUp" && strings.Contains(e.Message, "it wouldn't fit if a new node is added") {
						t.Log("NotTriggerScaleUp event found")
						eventFound = true
						break EventsLoop
					}
				}
			}
			require.Equal(t, eventFound, true)
			// Verify that cluster size is not changed
			require.NoError(t, WaitForClusterSizeFunc(f,
				func(size int) bool { return size <= nodeCount }, time.Second))
		})

		simpleScaleUpTest := func(unready int) {
			ReserveMemoryAsync(f, "memory-reservation", 100, int(1.1*float64(nodeCount*memAllocatableMb)), false, 1*time.Second)
			defer e2erc.DeleteRCAndWaitForGC(f.ClientSet, f.Namespace.Name, "memory-reservation")

			// Verify that cluster size is increased
			require.NoError(t, WaitForClusterSizeFuncWithUnready(f,
				func(size int) bool { return size >= nodeCount+1 }, scaleUpTimeout, unready))
			require.NoError(t, waitForAllCaPodsReadyInNamespace(f))
		}

		it("should increase cluster size if pending pods are small [Feature:ClusterSizeAutoscalingScaleUp]",
			func() { simpleScaleUpTest(0) })

		it("shouldn't trigger additional scale-ups during processing scale-up [Feature:ClusterSizeAutoscalingScaleUp]", func() {
			// Wait for the situation to stabilize - CA should be running and have up-to-date node readiness info.
			status, err := waitForScaleUpStatus(f, func(s *scaleUpStatus) bool {
				return s.ready == s.target && s.ready <= nodeCount
			}, scaleUpTriggerTimeout)
			require.NoError(t, err)

			unmanagedNodes := nodeCount - status.ready

			t.Log("Schedule more pods than can fit and wait for cluster to scale-up")
			ReserveMemoryAsync(f, "memory-reservation", 100, nodeCount*memAllocatableMb, false, 1*time.Second)
			defer e2erc.DeleteRCAndWaitForGC(f.ClientSet, f.Namespace.Name, "memory-reservation")

			status, err = waitForScaleUpStatus(f, func(s *scaleUpStatus) bool {
				return s.status == caOngoingScaleUpStatus
			}, scaleUpTriggerTimeout)
			require.NoError(t, err)
			target := status.target
			require.NoError(t, waitForAllCaPodsReadyInNamespace(f))

			t.Log("Expect no more scale-up to be happening after all pods are scheduled")

			// wait for a while until scale-up finishes; we cannot read CA status immediately
			// after pods are scheduled as status config map is updated by CA once every loop iteration
			status, err = waitForScaleUpStatus(f, func(s *scaleUpStatus) bool {
				return s.status == caNoScaleUpStatus
			}, 2*freshStatusLimit)
			require.NoError(t, err)

			if status.target != target {
				f.T.Logf("Final number of nodes (%v) does not match initial scale-up target (%v).", status.target, target)
			}
			require.Equal(t, status.timestamp.Add(freshStatusLimit).Before(time.Now()), false)
			require.Equal(t, status.status, caNoScaleUpStatus)
			require.Equal(t, status.ready, status.target)
			nodes, err := e2enode.GetReadySchedulableNodes(f.ClientSet)
			require.NoError(t, err)
			require.Equal(t, len(nodes.Items), status.target+unmanagedNodes)
		})

		it("should increase cluster size if pods are pending due to host port conflict [Feature:ClusterSizeAutoscalingScaleUp]", func() {
			CreateHostPortPods(f, "host-port", nodeCount+2, false)
			defer e2erc.DeleteRCAndWaitForGC(f.ClientSet, f.Namespace.Name, "host-port")

			require.NoError(t, WaitForClusterSizeFunc(f,
				func(size int) bool { return size >= nodeCount+2 }, scaleUpTimeout))
			require.NoError(t, waitForAllCaPodsReadyInNamespace(f))
		})

		it("should increase cluster size if pods are pending due to pod anti-affinity [Feature:ClusterSizeAutoscalingScaleUp]", func() {
			pods := nodeCount
			newPods := 2
			labels := map[string]string{
				"anti-affinity": "yes",
			}
			t.Log("starting a pod with anti-affinity on each node")
			require.NoError(t, runAntiAffinityPods(f, f.Namespace.Name, pods, "some-pod", labels, labels))
			defer e2erc.DeleteRCAndWaitForGC(f.ClientSet, f.Namespace.Name, "some-pod")
			require.NoError(t, waitForAllCaPodsReadyInNamespace(f))

			t.Log("scheduling extra pods with anti-affinity to existing ones")
			require.NoError(t, runAntiAffinityPods(f, f.Namespace.Name, newPods, "extra-pod", labels, labels))
			defer e2erc.DeleteRCAndWaitForGC(f.ClientSet, f.Namespace.Name, "extra-pod")

			require.NoError(t, waitForAllCaPodsReadyInNamespace(f))
			require.NoError(t, e2enode.WaitForReadyNodes(f.ClientSet, nodeCount+newPods, scaleUpTimeout))
		})

		it("should increase cluster size if pod requesting EmptyDir volume is pending [Feature:ClusterSizeAutoscalingScaleUp]", func() {
			t.Log("creating pods")
			pods := nodeCount
			newPods := 1
			labels := map[string]string{
				"anti-affinity": "yes",
			}
			require.NoError(t, runAntiAffinityPods(f, f.Namespace.Name, pods, "some-pod", labels, labels))
			defer e2erc.DeleteRCAndWaitForGC(f.ClientSet, f.Namespace.Name, "some-pod")

			t.Log("waiting for all pods before triggering scale up")
			require.NoError(t, waitForAllCaPodsReadyInNamespace(f))

			t.Log("creating a pod requesting EmptyDir")
			require.NoError(t, runVolumeAntiAffinityPods(f, f.Namespace.Name, newPods, "extra-pod", labels, labels, emptyDirVolumes))
			defer e2erc.DeleteRCAndWaitForGC(f.ClientSet, f.Namespace.Name, "extra-pod")

			require.NoError(t, waitForAllCaPodsReadyInNamespace(f))
			require.NoError(t, e2enode.WaitForReadyNodes(f.ClientSet, nodeCount+newPods, scaleUpTimeout))
		})

		simpleScaleDownTest := func(unready int) {
			cleanup, err := addKubeSystemPdbs(f)
			defer cleanup()
			require.NoError(t, err)

			t.Log("Manually increase cluster size")
			increasedSize := 0
			newSizes := make(map[string]int)
			for key, val := range originalSizes {
				newSizes[key] = val + 2 + unready
				increasedSize += val + 2 + unready
			}
			setMigSizes(f, newSizes)
			require.NoError(t, WaitForClusterSizeFuncWithUnready(f,
				func(size int) bool { return size >= increasedSize }, manualResizeTimeout, unready))

			t.Log("Some node should be removed")
			require.NoError(t, WaitForClusterSizeFuncWithUnready(f,
				func(size int) bool { return size < increasedSize }, scaleDownTimeout, unready))
		}

		it("should correctly scale down after a node is not needed [Feature:ClusterSizeAutoscalingScaleDown]",
			func() { simpleScaleDownTest(0) })

		it("should be able to scale down when rescheduling a pod is required and pdb allows for it[Feature:ClusterSizeAutoscalingScaleDown]", func() {
			runDrainTest(f, originalSizes, f.Namespace.Name, 1, 1, func(increasedSize int) {
				t.Log("Some node should be removed")
				require.NoError(t, WaitForClusterSizeFunc(f,
					func(size int) bool { return size < increasedSize }, scaleDownTimeout))
			})
		})

		it("shouldn't be able to scale down when rescheduling a pod is required, but pdb doesn't allow drain[Feature:ClusterSizeAutoscalingScaleDown]", func() {
			runDrainTest(f, originalSizes, f.Namespace.Name, 1, 0, func(increasedSize int) {
				t.Log("No nodes should be removed")
				time.Sleep(scaleDownTimeout)
				nodes, err := e2enode.GetReadySchedulableNodes(f.ClientSet)
				require.NoError(t, err)
				require.Equal(t, len(nodes.Items), increasedSize)
			})
		})

		it("should be able to scale down by draining multiple pods one by one as dictated by pdb[Feature:ClusterSizeAutoscalingScaleDown]", func() {
			runDrainTest(f, originalSizes, f.Namespace.Name, 2, 1, func(increasedSize int) {
				t.Log("Some node should be removed")
				require.NoError(t, WaitForClusterSizeFunc(f,
					func(size int) bool { return size < increasedSize }, scaleDownTimeout))
			})
		})

		it("should be able to scale down by draining system pods with pdb[Feature:ClusterSizeAutoscalingScaleDown]", func() {
			runDrainTest(f, originalSizes, "kube-system", 2, 1, func(increasedSize int) {
				t.Log("Some node should be removed")
				require.NoError(t, WaitForClusterSizeFunc(f,
					func(size int) bool { return size < increasedSize }, scaleDownTimeout))
			})
		})

		it("shouldn't scale up when expendable pod is created [Feature:ClusterSizeAutoscalingScaleUp]", func() {
			defer createPriorityClasses(f)()
			// Create nodesCountAfterResize+1 pods allocating 0.7 allocatable on present nodes. One more node will have to be created.
			cleanupFunc := ReserveMemoryWithPriority(f, "memory-reservation", nodeCount+1, int(float64(nodeCount+1)*float64(0.7)*float64(memAllocatableMb)), false, time.Second, expendablePriorityClassName)
			defer cleanupFunc()
			t.Logf("Waiting for scale up hoping it won't happen, sleep for %s", scaleUpTimeout.String())
			time.Sleep(scaleUpTimeout)
			// Verify that cluster size is not changed
			require.NoError(t, WaitForClusterSizeFunc(f,
				func(size int) bool { return size == nodeCount }, time.Second))
		})

		it("should scale up when non expendable pod is created [Feature:ClusterSizeAutoscalingScaleUp]", func() {
			defer createPriorityClasses(f)()
			// Create nodesCountAfterResize+1 pods allocating 0.7 allocatable on present nodes. One more node will have to be created.
			cleanupFunc := ReserveMemoryWithPriority(f, "memory-reservation", nodeCount+1, int(float64(nodeCount+1)*float64(0.7)*float64(memAllocatableMb)), true, scaleUpTimeout, highPriorityClassName)
			defer cleanupFunc()
			// Verify that cluster size is not changed
			require.NoError(t, WaitForClusterSizeFunc(f,
				func(size int) bool { return size > nodeCount }, time.Second))
		})

		it("shouldn't scale up when expendable pod is preempted [Feature:ClusterSizeAutoscalingScaleUp]", func() {
			defer createPriorityClasses(f)()
			// Create nodesCountAfterResize pods allocating 0.7 allocatable on present nodes - one pod per node.
			cleanupFunc1 := ReserveMemoryWithPriority(f, "memory-reservation1", nodeCount, int(float64(nodeCount)*float64(0.7)*float64(memAllocatableMb)), true, defaultTimeout, expendablePriorityClassName)
			defer cleanupFunc1()
			// Create nodesCountAfterResize pods allocating 0.7 allocatable on present nodes - one pod per node. Pods created here should preempt pods created above.
			cleanupFunc2 := ReserveMemoryWithPriority(f, "memory-reservation2", nodeCount, int(float64(nodeCount)*float64(0.7)*float64(memAllocatableMb)), true, defaultTimeout, highPriorityClassName)
			defer cleanupFunc2()
			require.NoError(t, WaitForClusterSizeFunc(f,
				func(size int) bool { return size == nodeCount }, time.Second))
		})

		it("should scale down when expendable pod is running [Feature:ClusterSizeAutoscalingScaleDown]", func() {
			defer createPriorityClasses(f)()
			increasedSize := manuallyIncreaseClusterSize(f, originalSizes)
			// Create increasedSize pods allocating 0.7 allocatable on present nodes - one pod per node.
			cleanupFunc := ReserveMemoryWithPriority(f, "memory-reservation", increasedSize, int(float64(increasedSize)*float64(0.7)*float64(memAllocatableMb)), true, scaleUpTimeout, expendablePriorityClassName)
			defer cleanupFunc()
			t.Log("Waiting for scale down")
			require.NoError(t, WaitForClusterSizeFunc(f,
				func(size int) bool { return size == nodeCount }, scaleDownTimeout))
		})

		it("shouldn't scale down when non expendable pod is running [Feature:ClusterSizeAutoscalingScaleDown]", func() {
			defer createPriorityClasses(f)()
			increasedSize := manuallyIncreaseClusterSize(f, originalSizes)
			// Create increasedSize pods allocating 0.7 allocatable on present nodes - one pod per node.
			cleanupFunc := ReserveMemoryWithPriority(f, "memory-reservation", increasedSize, int(float64(increasedSize)*float64(0.7)*float64(memAllocatableMb)), true, scaleUpTimeout, highPriorityClassName)
			defer cleanupFunc()
			t.Logf("Waiting for scale down hoping it won't happen, sleep for %s", scaleDownTimeout.String())
			time.Sleep(scaleDownTimeout)
			require.NoError(t, WaitForClusterSizeFunc(f,
				func(size int) bool { return size == increasedSize }, time.Second))
		})
	})
}
