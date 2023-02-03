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

package clusterapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

const (
	cpuStatusKey       = "cpu"
	memoryStatusKey    = "memory"
	nvidiaGpuStatusKey = "nvidia.com/gpu"
)

func TestSetSize(t *testing.T) {
	initialReplicas := 1
	updatedReplicas := 5

	test := func(t *testing.T, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		testResource := testConfig.machineSet
		if testConfig.machineDeployment != nil {
			testResource = testConfig.machineDeployment
		}

		sr, err := newUnstructuredScalableResource(controller, testResource)
		if err != nil {
			t.Fatal(err)
		}

		gvr, err := sr.GroupVersionResource()
		if err != nil {
			t.Fatal(err)
		}

		err = sr.SetSize(updatedReplicas)
		if err != nil {
			t.Fatal(err)
		}

		s, err := sr.controller.managementScaleClient.Scales(testResource.GetNamespace()).
			Get(context.TODO(), gvr.GroupResource(), testResource.GetName(), metav1.GetOptions{})

		if s.Spec.Replicas != int32(updatedReplicas) {
			t.Errorf("expected %v, got: %v", updatedReplicas, s.Spec.Replicas)
		}

		replicas, found, err := unstructured.NestedInt64(sr.unstructured.Object, "spec", "replicas")
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatal("replicas = 0")
		}
		if replicas != int64(updatedReplicas) {
			t.Errorf("expected %v, got: %v", updatedReplicas, replicas)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		test(t, createMachineSetTestConfig(
			RandomString(6),
			RandomString(6),
			RandomString(6),
			initialReplicas, map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			},
			nil,
		))
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		test(t, createMachineDeploymentTestConfig(
			RandomString(6),
			RandomString(6),
			RandomString(6),
			initialReplicas, map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			},
			nil,
		))
	})
}

func TestReplicas(t *testing.T) {
	initialReplicas := 1
	updatedReplicas := 5

	test := func(t *testing.T, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		testResource := testConfig.machineSet
		if testConfig.machineDeployment != nil {
			testResource = testConfig.machineDeployment
		}

		sr, err := newUnstructuredScalableResource(controller, testResource)
		if err != nil {
			t.Fatal(err)
		}

		gvr, err := sr.GroupVersionResource()
		if err != nil {
			t.Fatal(err)
		}

		i, err := sr.Replicas()
		if err != nil {
			t.Fatal(err)
		}

		if i != initialReplicas {
			t.Errorf("expected %v, got: %v", initialReplicas, i)
		}

		// fetch and update machineSet
		s, err := sr.controller.managementScaleClient.Scales(testResource.GetNamespace()).
			Get(context.TODO(), gvr.GroupResource(), testResource.GetName(), metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}

		s.Spec.Replicas = int32(updatedReplicas)

		ch := make(chan error)
		checkDone := func(obj interface{}) (bool, error) {
			u, ok := obj.(*unstructured.Unstructured)
			if !ok {
				return false, nil
			}
			sr, err := newUnstructuredScalableResource(controller, u)
			if err != nil {
				return true, err
			}
			i, err := sr.Replicas()
			if err != nil {
				return true, err
			}
			if i != updatedReplicas {
				return true, fmt.Errorf("expected %v, got: %v", updatedReplicas, i)
			}
			return true, nil
		}
		handler := cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				match, err := checkDone(obj)
				if match {
					ch <- err
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				match, err := checkDone(newObj)
				if match {
					ch <- err
				}
			},
		}

		controller.machineSetInformer.Informer().AddEventHandler(handler)
		controller.machineDeploymentInformer.Informer().AddEventHandler(handler)

		_, err = sr.controller.managementScaleClient.Scales(testResource.GetNamespace()).
			Update(context.TODO(), gvr.GroupResource(), s, metav1.UpdateOptions{})
		if err != nil {
			t.Fatal(err)
		}

		lastErr := fmt.Errorf("no updates received yet")
		for lastErr != nil {
			select {
			case err = <-ch:
				lastErr = err
			case <-time.After(1 * time.Second):
				t.Fatal(fmt.Errorf("timeout while waiting for update. Last error was: %v", lastErr))
			}
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		test(t, createMachineSetTestConfig(RandomString(6), RandomString(6), RandomString(6), initialReplicas, nil, nil))
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		test(t, createMachineDeploymentTestConfig(RandomString(6), RandomString(6), RandomString(6), initialReplicas, nil, nil))
	})
}

func TestSetSizeAndReplicas(t *testing.T) {
	initialReplicas := 1
	updatedReplicas := 5

	test := func(t *testing.T, testConfig *testConfig) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		testResource := testConfig.machineSet
		if testConfig.machineDeployment != nil {
			testResource = testConfig.machineDeployment
		}

		sr, err := newUnstructuredScalableResource(controller, testResource)
		if err != nil {
			t.Fatal(err)
		}

		i, err := sr.Replicas()
		if err != nil {
			t.Fatal(err)
		}

		if i != initialReplicas {
			t.Errorf("expected %v, got: %v", initialReplicas, i)
		}

		err = sr.SetSize(updatedReplicas)
		if err != nil {
			t.Fatal(err)
		}

		i, err = sr.Replicas()
		if err != nil {
			t.Fatal(err)
		}

		if i != updatedReplicas {
			t.Errorf("expected %v, got: %v", updatedReplicas, i)
		}
	}

	t.Run("MachineSet", func(t *testing.T) {
		test(t, createMachineSetTestConfig(
			RandomString(6),
			RandomString(6),
			RandomString(6),
			initialReplicas, map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			},
			nil,
		))
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		test(t, createMachineDeploymentTestConfig(
			RandomString(6),
			RandomString(6),
			RandomString(6),
			initialReplicas, map[string]string{
				nodeGroupMinSizeAnnotationKey: "1",
				nodeGroupMaxSizeAnnotationKey: "10",
			},
			nil,
		))
	})
}

func TestAnnotations(t *testing.T) {
	cpuQuantity := resource.MustParse("2")
	memQuantity := resource.MustParse("1024Mi")
	diskQuantity := resource.MustParse("100Gi")
	gpuQuantity := resource.MustParse("1")
	maxPodsQuantity := resource.MustParse("42")
	expectedTaints := []v1.Taint{{Key: "key1", Effect: v1.TaintEffectNoSchedule, Value: "value1"}, {Key: "key2", Effect: v1.TaintEffectNoExecute, Value: "value2"}}
	annotations := map[string]string{
		cpuKey:          cpuQuantity.String(),
		memoryKey:       memQuantity.String(),
		diskCapacityKey: diskQuantity.String(),
		gpuCountKey:     gpuQuantity.String(),
		maxPodsKey:      maxPodsQuantity.String(),
		taintsKey:       "key1=value1:NoSchedule,key2=value2:NoExecute",
		labelsKey:       "key3=value3,key4=value4,key5=value5",
	}

	test := func(t *testing.T, testConfig *testConfig, testResource *unstructured.Unstructured) {
		controller, stop := mustCreateTestController(t, testConfig)
		defer stop()

		sr, err := newUnstructuredScalableResource(controller, testResource)
		if err != nil {
			t.Fatal(err)
		}

		if cpu, err := sr.InstanceCPUCapacityAnnotation(); err != nil {
			t.Fatal(err)
		} else if cpuQuantity.Cmp(cpu) != 0 {
			t.Errorf("expected %v, got %v", cpuQuantity, cpu)
		}

		if mem, err := sr.InstanceMemoryCapacityAnnotation(); err != nil {
			t.Fatal(err)
		} else if memQuantity.Cmp(mem) != 0 {
			t.Errorf("expected %v, got %v", memQuantity, mem)
		}

		if disk, err := sr.InstanceEphemeralDiskCapacityAnnotation(); err != nil {
			t.Fatal(err)
		} else if diskQuantity.Cmp(disk) != 0 {
			t.Errorf("expected %v, got %v", diskQuantity, disk)
		}

		if gpu, err := sr.InstanceGPUCapacityAnnotation(); err != nil {
			t.Fatal(err)
		} else if gpuQuantity.Cmp(gpu) != 0 {
			t.Errorf("expected %v, got %v", gpuQuantity, gpu)
		}

		if maxPods, err := sr.InstanceMaxPodsCapacityAnnotation(); err != nil {
			t.Fatal(err)
		} else if maxPodsQuantity.Cmp(maxPods) != 0 {
			t.Errorf("expected %v, got %v", maxPodsQuantity, maxPods)
		}

		taints := sr.Taints()
		assert.Equal(t, expectedTaints, taints)

		labels := sr.Labels()
		assert.Len(t, labels, 3)
		assert.Equal(t, "value3", labels["key3"])
		assert.Equal(t, "value4", labels["key4"])
		assert.Equal(t, "value5", labels["key5"])
	}

	t.Run("MachineSet", func(t *testing.T) {
		testConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), RandomString(6), 1, annotations, nil)
		test(t, testConfig, testConfig.machineSet)
	})

	t.Run("MachineDeployment", func(t *testing.T) {
		testConfig := createMachineDeploymentTestConfig(RandomString(6), RandomString(6), RandomString(6), 1, annotations, nil)
		test(t, testConfig, testConfig.machineDeployment)
	})
}

func TestCanScaleFromZero(t *testing.T) {
	testConfigs := []struct {
		name        string
		annotations map[string]string
		capacity    map[string]string
		canScale    bool
	}{
		{
			"can scale from zero",
			map[string]string{
				cpuKey:    "1",
				memoryKey: "1024Mi",
			},
			nil,
			true,
		},
		{
			"with missing CPU info cannot scale from zero",
			map[string]string{
				memoryKey: "1024Mi",
			},
			nil,
			false,
		},
		{
			"with missing Memory info cannot scale from zero",
			map[string]string{
				cpuKey: "1",
			},
			nil,
			false,
		},
		{
			"with no information cannot scale from zero",
			map[string]string{},
			nil,
			false,
		},
		{
			"with capacity in machine template can scale from zero",
			map[string]string{},
			map[string]string{
				cpuStatusKey:    "1",
				memoryStatusKey: "4G",
			},
			true,
		},
		{
			"with missing cpu capacity in machine template cannot scale from zero",
			map[string]string{},
			map[string]string{
				memoryStatusKey: "4G",
			},
			false,
		},
		{
			"with missing memory capacity in machine template cannot scale from zero",
			map[string]string{},
			map[string]string{
				cpuStatusKey: "1",
			},
			false,
		},
		{
			"with both annotations and capacity in machine template can scale from zero",
			map[string]string{
				cpuKey:    "1",
				memoryKey: "1024Mi",
			},
			map[string]string{
				cpuStatusKey:    "1",
				memoryStatusKey: "4G",
			},
			true,
		},
		{
			"with incomplete annotations and capacity in machine template cannot scale from zero",
			map[string]string{
				cpuKey: "1",
			},
			map[string]string{
				nvidiaGpuStatusKey: "1",
			},
			false,
		},
		{
			"with complete information split across annotations and capacity in machine template can scale from zero",
			map[string]string{
				cpuKey: "1",
			},
			map[string]string{
				memoryStatusKey: "4G",
			},
			true,
		},
	}

	for _, tc := range testConfigs {
		testname := fmt.Sprintf("MachineSet %s", tc.name)
		t.Run(testname, func(t *testing.T) {
			msTestConfig := createMachineSetTestConfig(RandomString(6), RandomString(6), RandomString(6), 1, tc.annotations, tc.capacity)
			controller, stop := mustCreateTestController(t, msTestConfig)
			defer stop()

			testResource := msTestConfig.machineSet

			sr, err := newUnstructuredScalableResource(controller, testResource)
			if err != nil {
				t.Fatal(err)
			}

			canScale := sr.CanScaleFromZero()
			if canScale != tc.canScale {
				t.Errorf("expected %v, got %v", tc.canScale, canScale)
			}
		})
	}

	for _, tc := range testConfigs {
		testname := fmt.Sprintf("MachineDeployment %s", tc.name)
		t.Run(testname, func(t *testing.T) {
			msTestConfig := createMachineDeploymentTestConfig(RandomString(6), RandomString(6), RandomString(6), 1, tc.annotations, tc.capacity)
			controller, stop := mustCreateTestController(t, msTestConfig)
			defer stop()

			testResource := msTestConfig.machineDeployment

			sr, err := newUnstructuredScalableResource(controller, testResource)
			if err != nil {
				t.Fatal(err)
			}

			canScale := sr.CanScaleFromZero()
			if canScale != tc.canScale {
				t.Errorf("expected %v, got %v", tc.canScale, canScale)
			}
		})
	}
}
