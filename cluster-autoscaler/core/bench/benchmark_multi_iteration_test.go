/*
Copyright 2024 The Kubernetes Authors.

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
	"sync"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	targetNodesCount = flag.Int("target-nodes-count", 20000, "Target number of nodes for multi-iteration benchmark")
	stepsCount       = flag.Int("steps-count", 10, "Number of steps in multi-iteration benchmark")
)

var (
	initLogsOnce sync.Once
)

func initLogs() {
	if !flag.Parsed() {
		flag.Parse()
	}
	klog.LogToStderr(false)
	klog.SetOutput(io.Discard)
	ctrl.SetLogger(klog.Background())
}

var e2TypesMap = map[string]struct {
	cpu int64
	mem int64
}{
	"e2-standard-2":  {2000, 8 * units.GiB},
	"e2-standard-4":  {4000, 16 * units.GiB},
	"e2-standard-8":  {8000, 32 * units.GiB},
	"e2-standard-16": {16000, 64 * units.GiB},
	"e2-standard-32": {32000, 128 * units.GiB},
	"e2-highcpu-2":   {2000, 2 * units.GiB},
	"e2-highcpu-4":   {4000, 4 * units.GiB},
	"e2-highcpu-8":   {8000, 8 * units.GiB},
	"e2-highcpu-16":  {16000, 16 * units.GiB},
	"e2-highcpu-32":  {32000, 32 * units.GiB},
	"e2-highmem-2":   {2000, 16 * units.GiB},
	"e2-highmem-4":   {4000, 32 * units.GiB},
	"e2-highmem-8":   {8000, 64 * units.GiB},
	"e2-highmem-16":  {16000, 128 * units.GiB},
}

var e2Zones = []string{"us-central1-a", "us-central1-b", "us-central1-c"}

func setupMultiIteration(clusterFakes *integration.FakeSet) error {
	clusterFakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameCores, 0, 10000000)
	clusterFakes.CloudProvider.SetResourceLimit(cloudprovider.ResourceNameMemory, 0, 10000000*units.GiB)

	// Create 42 NodeGroups (3 zones * 14 instance types)
	for _, zone := range e2Zones {
		for itName, typeDetails := range e2TypesMap {
			ngName := fmt.Sprintf("ng-%s-%s", zone, itName)
			nTemplate := BuildTestNode(fmt.Sprintf("%s-template", ngName), typeDetails.cpu, typeDetails.mem)
			nTemplate.Labels[apiv1.LabelArchStable] = "amd64"
			nTemplate.Labels[apiv1.LabelOSStable] = "linux"
			nTemplate.Labels["topology.kubernetes.io/zone"] = zone
			nTemplate.Labels["node.kubernetes.io/instance-type"] = itName
			nTemplate.Labels["kubernetes.io/hostname"] = nTemplate.Name
			nTemplate.Labels["cluster-autoscaler.kubernetes.io/nodegroup-id"] = ngName
			SetNodeReadyState(nTemplate, true, time.Now())

			clusterFakes.CloudProvider.AddNodeGroup(ngName,
				testprovider.WithTemplate(framework.NewNodeInfo(nTemplate, nil)),
				testprovider.WithNGSize(0, 1000000),
			)
		}
	}
	return nil
}

func (s *scenario) runMultiIteration(b *testing.B) {
	initLogsOnce.Do(initLogs)
	n := *targetNodesCount
	k := *stepsCount

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		clusterFakes := integration.NewFakeSet()
		if s.setup != nil {
			if err := s.setup(clusterFakes); err != nil {
				b.Fatalf("setup failed: %v", err)
			}
		}

		// Add 10 DaemonSets
		for dsIdx := 0; dsIdx < 10; dsIdx++ {
			dsName := fmt.Sprintf("ds-%d", dsIdx)
			ds := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dsName,
					Namespace: "default",
				},
				Spec: appsv1.DaemonSetSpec{
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": dsName}},
					Template: apiv1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": dsName}},
						Spec: apiv1.PodSpec{
							Containers: []apiv1.Container{
								{
									Name:  "c",
									Image: "i",
									Resources: apiv1.ResourceRequirements{
										Requests: apiv1.ResourceList{
											apiv1.ResourceCPU:    *resource.NewMilliQuantity(50, resource.DecimalSI),
											apiv1.ResourceMemory: *resource.NewQuantity(50*1024*1024, resource.BinarySI),
										},
									},
								},
							},
						},
					},
				},
			}
			_, _ = clusterFakes.KubeClient.AppsV1().DaemonSets("default").Create(context.TODO(), ds, metav1.CreateOptions{})
		}

		autoscaler := newAutoscaler(context.Background(), b, *s, clusterFakes)

		for step := 1; step <= k; step++ {
			b.StopTimer()
			targetNumPods := step * n / k
			currentPods, _ := clusterFakes.KubeClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
			for p := len(currentPods.Items); p < targetNumPods; p++ {
				pod := BuildTestPod(fmt.Sprintf("pod-%d", p), 1500, 512*1024*1024,
					WithLabels(map[string]string{
						"app":                 "workload",
						apiv1.LabelArchStable: "amd64",
						apiv1.LabelOSStable:   "linux",
					}),
					WithPodAntiAffinity(map[string]string{"app": "workload"}, apiv1.LabelHostname),
					MarkUnschedulable(),
				)
				pod.Status.Phase = apiv1.PodPending
				clusterFakes.K8s.AddPod(pod)
			}

			// Wait for informer to catch up.
			for i := 0; i < 100; i++ {
				pods, _ := clusterFakes.InformerFactory.Core().V1().Pods().Lister().List(labels.Everything())
				if len(pods) >= targetNumPods {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}

			b.StartTimer()
			err := autoscaler.RunOnce(time.Now().Add(10 * time.Second))
			if err != nil {
				b.Fatalf("run failed at step %d: %v", step, err)
			}
			b.StopTimer()
		}

		if s.verify != nil {
			err := s.verify(clusterFakes)
			if err != nil {
				b.Fatalf("verify failed: %v", err)
			}
		}

		// Print node group distribution
		grandTotal := 0
		fmt.Printf("\nNode Group Distribution:\n")
		for _, ng := range clusterFakes.CloudProvider.NodeGroups() {
			size, _ := ng.TargetSize()
			if size > 0 {
				fmt.Printf("- %s: %d nodes\n", ng.Id(), size)
				grandTotal += size
			}
		}
		fmt.Printf("GRAND TOTAL: %d nodes\n\n", grandTotal)
	}
}

func BenchmarkMultiIteration_Default(b *testing.B) {
	s := scenario{
		setup:  setupMultiIteration,
		verify: verifyTotalTargetSizeAtLeast(*targetNodesCount),
		config: func(opts *config.AutoscalingOptions) {
			configureAffinitySurge(opts)
			opts.KarpenterSimulatorEnabled = false
			opts.BalancingExtraIgnoredLabels = []string{"ca.prototype/nodegroup-id"}
		},
	}
	s.runMultiIteration(b)
}

func BenchmarkMultiIteration_Karpenter(b *testing.B) {
	s := scenario{
		setup:  setupMultiIteration,
		verify: verifyTotalTargetSizeAtLeast(*targetNodesCount),
		config: func(opts *config.AutoscalingOptions) {
			configureAffinitySurge(opts)
			opts.KarpenterSimulatorEnabled = true
			opts.BalancingExtraIgnoredLabels = []string{"ca.prototype/nodegroup-id"}
		},
	}
	s.runMultiIteration(b)
}
