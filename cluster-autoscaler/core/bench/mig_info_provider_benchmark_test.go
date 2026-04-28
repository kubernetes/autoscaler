/*
Copyright 2026 The Kubernetes Authors.

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
	"regexp"
	"testing"
	"time"

	compute "google.golang.org/api/compute/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/test/integration"
)

// mockMig implements gce.Mig interface for benchmarking.
type mockMig struct {
	ref gce.GceRef
}

func (m *mockMig) MaxSize() int                                   { return 1000 }
func (m *mockMig) MinSize() int                                   { return 0 }
func (m *mockMig) TargetSize() (int, error)                       { return 10, nil }
func (m *mockMig) IncreaseSize(delta int) error                   { return nil }
func (m *mockMig) DeleteNodes([]*corev1.Node) error               { return nil }
func (m *mockMig) DecreaseTargetSize(delta int) error             { return nil }
func (m *mockMig) Id() string                                     { return m.ref.Name }
func (m *mockMig) Debug() string                                  { return m.ref.Name }
func (m *mockMig) Nodes() ([]cloudprovider.Instance, error)       { return nil, nil }
func (m *mockMig) TemplateNodeInfo() (*framework.NodeInfo, error) { return nil, nil }
func (m *mockMig) Exist() bool                                    { return true }
func (m *mockMig) Create() (cloudprovider.NodeGroup, error)       { return nil, nil }
func (m *mockMig) Delete() error                                  { return nil }
func (m *mockMig) Autoprovisioned() bool                          { return false }
func (m *mockMig) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return &defaults, nil
}
func (m *mockMig) GceRef() gce.GceRef                          { return m.ref }
func (m *mockMig) IsStable() (bool, error)                     { return true, nil }
func (m *mockMig) AtomicIncreaseSize(delta int) error          { return nil }
func (m *mockMig) ForceDeleteNodes(nodes []*corev1.Node) error { return nil }

// benchmarkGceManager implements gce.GceManager interface for benchmarking.
type benchmarkGceManager struct {
	provider gce.MigInfoProvider
	migs     []gce.Mig
}

func (m *benchmarkGceManager) Refresh() error     { return nil }
func (m *benchmarkGceManager) Cleanup() error     { return nil }
func (m *benchmarkGceManager) GetMigs() []gce.Mig { return m.migs }
func (m *benchmarkGceManager) GetMigNodes(mig gce.Mig) ([]gce.GceInstance, error) {
	return m.provider.GetMigInstances(mig.GceRef())
}
func (m *benchmarkGceManager) GetMigForInstance(instance gce.GceRef) (gce.Mig, error) {
	return m.provider.GetMigForInstance(instance)
}
func (m *benchmarkGceManager) GetMigTemplateNode(mig gce.Mig) (*corev1.Node, error) { return nil, nil }
func (m *benchmarkGceManager) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return nil, nil
}
func (m *benchmarkGceManager) GetMigSize(mig gce.Mig) (int64, error) {
	return m.provider.GetMigTargetSize(mig.GceRef())
}
func (m *benchmarkGceManager) GetMigOptions(mig gce.Mig, defaults config.NodeGroupAutoscalingOptions) *config.NodeGroupAutoscalingOptions {
	return &defaults
}
func (m *benchmarkGceManager) IsMigStable(mig gce.Mig) (bool, error)          { return true, nil }
func (m *benchmarkGceManager) SetMigSize(mig gce.Mig, size int64) error       { return nil }
func (m *benchmarkGceManager) DeleteInstances(instances []gce.GceRef) error   { return nil }
func (m *benchmarkGceManager) CreateInstances(mig gce.Mig, delta int64) error { return nil }

// mockAutoscalingGceClient implements gce.AutoscalingGceClient for benchmarking.
type mockAutoscalingGceClient struct {
	fetchMigInstances  func(gce.GceRef) ([]gce.GceInstance, error)
	fetchMigTargetSize func(gce.GceRef) (int64, error)
	fetchMigBasename   func(gce.GceRef) (string, error)
}

func (client *mockAutoscalingGceClient) FetchMachineType(zone, machineName string) (*compute.MachineType, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) FetchMachineTypes(_ string) ([]*compute.MachineType, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) FetchAllMigs(zone string) ([]*compute.InstanceGroupManager, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) FetchAllInstances(project, zone string, filter string) ([]gce.GceInstance, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) FetchMig(migRef gce.GceRef) (*compute.InstanceGroupManager, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) FetchMigTargetSize(migRef gce.GceRef) (int64, error) {
	return client.fetchMigTargetSize(migRef)
}
func (client *mockAutoscalingGceClient) FetchMigBasename(migRef gce.GceRef) (string, error) {
	return client.fetchMigBasename(migRef)
}
func (client *mockAutoscalingGceClient) FetchListManagedInstancesResults(migRef gce.GceRef) (string, error) {
	return "", nil
}
func (client *mockAutoscalingGceClient) FetchMigInstances(migRef gce.GceRef) ([]gce.GceInstance, error) {
	return client.fetchMigInstances(migRef)
}
func (client *mockAutoscalingGceClient) FetchMigTemplateName(migRef gce.GceRef) (gce.InstanceTemplateName, error) {
	return gce.InstanceTemplateName{}, nil
}
func (client *mockAutoscalingGceClient) FetchMigTemplate(migRef gce.GceRef, templateName string, regional bool) (*compute.InstanceTemplate, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) FetchMigsWithName(_ string, _ *regexp.Regexp) ([]string, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) FetchZones(_ string) ([]string, error) { return nil, nil }
func (client *mockAutoscalingGceClient) FetchAvailableCpuPlatforms() (map[string][]string, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) FetchAvailableDiskTypes(_ string) ([]string, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) CreateInstances(gce.GceRef, string, int64, []string) error {
	return nil
}
func (client *mockAutoscalingGceClient) DeleteInstances(gce.GceRef, []gce.GceRef) error { return nil }
func (client *mockAutoscalingGceClient) FetchReservations() ([]*compute.Reservation, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) FetchReservationsInProject(projectId string) ([]*compute.Reservation, error) {
	return nil, nil
}
func (client *mockAutoscalingGceClient) ResizeMig(gce.GceRef, int64) error { return nil }
func (client *mockAutoscalingGceClient) WaitForOperation(operationName, operationType, project, zone string) error {
	return nil
}

func BenchmarkRunOnceWithGce(b *testing.B) {
	b.StopTimer()
	// klog.LogToStderr(false)
	// klog.SetOutput(io.Discard)

	const (
		numMigs            = 200
		totalReadyNodes    = 200
		totalDeletedNodes  = 100
		totalUpcomingNodes = 100
	)

	getCounts := func(migIdx int) (active, deleted, upcoming int) {
		active = totalReadyNodes / numMigs
		if migIdx < totalReadyNodes%numMigs {
			active++
		}
		deleted = totalDeletedNodes / numMigs
		if migIdx < totalDeletedNodes%numMigs {
			deleted++
		}
		upcoming = totalUpcomingNodes / numMigs
		if migIdx < totalUpcomingNodes%numMigs {
			upcoming++
		}
		return
	}

	s := scenario{
		createCloudProvider: func(fakes *integration.FakeSet) cloudprovider.CloudProvider {
			client := &mockAutoscalingGceClient{}
			cache := gce.NewGceCache()
			migLister := gce.NewMigLister(cache)

			migs := make([]gce.Mig, numMigs)
			for j := 0; j < numMigs; j++ {
				migRef := gce.GceRef{
					Project: "project",
					Zone:    "us-central1-a",
					Name:    fmt.Sprintf("mig-%d", j),
				}
				migs[j] = &mockMig{ref: migRef}
				cache.RegisterMig(migs[j])
				cache.SetMigBasename(migRef, fmt.Sprintf("mig-%d", j))
			}

			provider := gce.NewCachingMigInfoProvider(cache, migLister, client, "project", 1, time.Hour, false, false)
			manager := &benchmarkGceManager{
				provider: provider,
				migs:     migs,
			}
			gceProvider, _ := gce.BuildGceCloudProvider(manager, nil, nil)

			client.fetchMigTargetSize = func(ref gce.GceRef) (int64, error) {
				var idx int
				fmt.Sscanf(ref.Name, "mig-%d", &idx)
				active, _, upcoming := getCounts(idx)
				return int64(active + upcoming), nil
			}

			client.fetchMigBasename = func(ref gce.GceRef) (string, error) {
				return ref.Name, nil
			}

			client.fetchMigInstances = func(migRef gce.GceRef) ([]gce.GceInstance, error) {
				var idx int
				fmt.Sscanf(migRef.Name, "mig-%d", &idx)
				active, _, _ := getCounts(idx)
				var instances []gce.GceInstance
				for k := 0; k < active; k++ {
					nodeName := fmt.Sprintf("%s-instance-%d", migRef.Name, k)
					instances = append(instances, gce.GceInstance{
						Instance: cloudprovider.Instance{
							Id:     fmt.Sprintf("gce://project/us-central1-a/%s", nodeName),
							Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
						},
						Igm: migRef,
					})
				}
				return instances, nil
			}

			return gceProvider
		},
		setup: func(fakes *integration.FakeSet) error {
			kubeClient := fakes.KubeClient
			// Add nodes to K8s
			for j := 0; j < numMigs; j++ {
				migName := fmt.Sprintf("mig-%d", j)
				active, deleted, _ := getCounts(j)
				// active nodes
				for k := 0; k < active; k++ {
					nodeName := fmt.Sprintf("%s-instance-%d", migName, k)
					node := &corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: nodeName,
						},
						Spec: corev1.NodeSpec{
							ProviderID: fmt.Sprintf("gce://project/us-central1-a/%s", nodeName),
						},
					}
					kubeClient.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
				}

				// deleted nodes
				for k := 0; k < deleted; k++ {
					deletedNodeName := fmt.Sprintf("%s-deleted-%d-%d", migName, j, k)
					deletedNode := &corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Name: deletedNodeName,
						},
						Spec: corev1.NodeSpec{
							ProviderID: fmt.Sprintf("gce://project/us-central1-a/%s", deletedNodeName),
						},
					}
					kubeClient.CoreV1().Nodes().Create(context.Background(), deletedNode, metav1.CreateOptions{})
				}
			}

			// Add 1 unschedulable pod to trigger scale-up logic
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "unschedulable-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "pause",
							Image: "registry.k8s.io/pause:3.1",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodScheduled,
							Status: corev1.ConditionFalse,
							Reason: "Unschedulable",
						},
					},
				},
			}
			kubeClient.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
			return nil
		},
		config: func(opts *config.AutoscalingOptions) {
			opts.EstimatorName = "binpacking"
			opts.ExpanderNames = "least-waste"
			opts.NodeGroupDefaults.ScaleDownUnneededTime = 1 * time.Minute
			opts.NodeGroupDefaults.MaxNodeProvisionTime = 10 * time.Minute
		},
	}

	s.run(b)
}
