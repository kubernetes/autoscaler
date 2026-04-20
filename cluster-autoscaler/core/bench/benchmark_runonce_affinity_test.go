package bench

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

const (
	nodesCount  = 5000
	podsPerNode = 10 // 50,000 pods total
	surgeCount  = 100

	expectedTotal = nodesCount + surgeCount
)

func BenchmarkRunOnceAffinitySurge_Default(b *testing.B) {
	s := scenario{
		setup:  setupAffinitySurge(nodesCount, podsPerNode, surgeCount, 50),
		verify: verifyTotalTargetSizeMassive(expectedTotal),
		config: func(opts *config.AutoscalingOptions) {
			configureAffinitySurge(opts)
			opts.FastPredicatesEnabled = false
		},
	}
	s.run(b)
}

func BenchmarkRunOnceAffinitySurge_FastPredicates(b *testing.B) {
	s := scenario{
		setup:  setupAffinitySurge(nodesCount, podsPerNode, surgeCount, 50),
		verify: verifyTotalTargetSizeMassive(expectedTotal),
		config: func(opts *config.AutoscalingOptions) {
			configureAffinitySurge(opts)
			opts.FastPredicatesEnabled = true
		},
	}
	s.run(b)
}

func configureAffinitySurge(opts *config.AutoscalingOptions) {
	opts.MaxNodesPerScaleUp = 1000000
	opts.ScaleUpFromZero = true
	opts.MaxCoresTotal = 1000000000
	opts.MaxMemoryTotal = 1000000000
	opts.MaxNodesTotal = 1000000
	opts.PredicateParallelism = 16
	// Use high binpacking limits to avoid benchmarks being cut by internal timeouts.
	opts.MaxBinpackingTime = 10 * time.Minute
	opts.MaxNodeGroupBinpackingDuration = 10 * time.Minute
}

func setupAffinitySurge(nodesCount, podsPerNode, surgePodsCount, nodesPerGroup int) func(*integration.FakeSet) error {
	go func() {
		importHttp := true
		if importHttp {
			_ = http.ListenAndServe("localhost:6060", nil)
		}
	}()
	return func(clusterFakes *integration.FakeSet) error {
		clusterFakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 1000000000)
		clusterFakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameMemory, 0, 1000000000)

		podGlobalIdx := 0
		numGroups := nodesCount / nodesPerGroup

		for i := 0; i < numGroups; i++ {
			ngName := fmt.Sprintf("ng-affinity-%d", i)
			nTemplate := BuildTestNode(fmt.Sprintf("%s-template", ngName), nodeCPU, nodeMem)
			if nTemplate.Labels == nil {
				nTemplate.Labels = make(map[string]string)
			}
			nTemplate.Labels["kubernetes.io/hostname"] = nTemplate.Name
			SetNodeReadyState(nTemplate, true, time.Now())

			clusterFakes.CloudProvider.AddNodeGroup(ngName,
				testprovider.WithTemplate(framework.NewNodeInfo(nTemplate, nil)),
				testprovider.WithNGSize(nodesPerGroup, 1000000),
			)

			ng := clusterFakes.CloudProvider.GetNodeGroup(ngName)
			if err := ng.IncreaseSize(nodesPerGroup); err != nil {
				return err
			}
			ngID := ng.Id()

			for j := 0; j < nodesPerGroup; j++ {
				nodeName := fmt.Sprintf("%s-node-%d", ngID, j)

				node := BuildTestNode(nodeName, nodeCPU, nodeMem)
				if node.Labels == nil {
					node.Labels = make(map[string]string)
				}
				node.Labels["kubernetes.io/hostname"] = nodeName
				SetNodeReadyState(node, true, time.Now())
				clusterFakes.K8s.AddNode(node)

				for p := 0; p < podsPerNode; p++ {
					podName := fmt.Sprintf("filler-%s-pod-%d", ngID, podGlobalIdx)
					pod := BuildTestPod(podName, 10, 10)
					pod.Spec.NodeName = nodeName

					// Create many small deployments using anti-affinity.
					// Every 2 pods belong to the same "deployment".
					deploymentID := podGlobalIdx / 2
					appLabel := fmt.Sprintf("app-%d", deploymentID)

					pod.Labels = map[string]string{"app": appLabel}
					pod.Spec.Affinity = &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"app": appLabel},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					}

					clusterFakes.K8s.AddPod(pod)
					podGlobalIdx++
				}
			}
		}

		for i := 0; i < surgePodsCount; i++ {
			podName := fmt.Sprintf("surge-pod-%d", i)
			pod := BuildTestPod(podName, nodeCPU, nodeMem, MarkUnschedulable())

			// Surge pods also have anti-affinity
			deploymentID := (podGlobalIdx + i) / 2
			appLabel := fmt.Sprintf("app-%d", deploymentID)

			pod.Labels = map[string]string{"app": appLabel}
			pod.Spec.Affinity = &corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": appLabel},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			}

			clusterFakes.K8s.AddPod(pod)
		}

		return nil
	}
}

func verifyTotalTargetSizeMassive(expectedTotalTargetSize int) func(*integration.FakeSet) error {
	return func(clusterFakes *integration.FakeSet) error {
		total := 0
		for _, ng := range clusterFakes.CloudProvider.NodeGroups() {
			targetSize, err := ng.TargetSize()
			if err != nil {
				return err
			}
			total += targetSize
		}
		if total != expectedTotalTargetSize {
			return fmt.Errorf("verify failed: expected total target size %d, got %d", expectedTotalTargetSize, total)
		}
		return nil
	}
}
