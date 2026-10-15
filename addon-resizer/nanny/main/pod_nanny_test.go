/*
Copyright 2016 The Kubernetes Authors.

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

package main

import (
    "testing"

    "k8s.io/autoscaler/addon-resizer/nanny"

    "github.com/google/go-cmp/cmp"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig"
    nannyconfigalpha "k8s.io/autoscaler/addon-resizer/nanny/apis/nannyconfig/v1alpha1"
    "k8s.io/client-go/kubernetes/fake"
)

func TestCurrentResources(t *testing.T) {
    defaultConfig := &nannyconfigalpha.NannyConfiguration{
        BaseCPU:       "30",
        CPUPerNode:    "1",
        BaseMemory:    "30Mi",
        MemoryPerNode: "1Mi",
    }
    defaultResources := []nanny.Resource{
        {
            Base:             resource.MustParse("30"),
            ExtraPerResource: resource.MustParse("1"),
            Name:             "cpu",
        },
        {
            Base:             resource.MustParse("30Mi"),
            ExtraPerResource: resource.MustParse("1Mi"),
            Name:             "memory",
        },
    }
    namespace := "fake-namespace"
    nannyConfigMapName := "fake-config"
    baseStorage := nannyconfig.NoValue

    testCases := []struct {
        name          string
        configMap     *corev1.ConfigMap
        wantResources []nanny.Resource
        wantErr       bool
    }{
        {
            name:          "no matching configmap, using default config",
            configMap:     &corev1.ConfigMap{},
            wantResources: defaultResources,
            wantErr:       false,
        },
        {
            name: "valid configmap, overriding default config",
            configMap: &corev1.ConfigMap{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      nannyConfigMapName,
                    Namespace: namespace,
                },
                Data: map[string]string{
                    "NannyConfiguration": "{\"apiVersion\":\"nannyconfig/v1alpha1\",\"kind\":\"NannyConfiguration\",\"baseCPU\":\"20\",\"baseMemory\":\"20Mi\",\"cpuPerNode\":\"2\",\"memoryPerNode\":\"2Mi\"}",
                },
            },
            wantResources: []nanny.Resource{
                {
                    Base:             resource.MustParse("20"),
                    ExtraPerResource: resource.MustParse("2"),
                    Name:             "cpu",
                },
                {
                    Base:             resource.MustParse("20Mi"),
                    ExtraPerResource: resource.MustParse("2Mi"),
                    Name:             "memory",
                },
            },
            wantErr: false,
        },
        {
            name: "invalid configmap data, using default config",
            configMap: &corev1.ConfigMap{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      nannyConfigMapName,
                    Namespace: namespace,
                },
                Data: map[string]string{
                    "NannyConfiguration": "fake-data",
                },
            },
            wantResources: defaultResources,
            wantErr:       false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            fakeClientset := fake.NewSimpleClientset(tc.configMap)
            nanyConfigUpdater := newNannyConfigUpdater(fakeClientset, defaultConfig, namespace, nannyConfigMapName, baseStorage)
            gotResources, gotErr := nanyConfigUpdater.CurrentResources()

            if gotErr != nil && !tc.wantErr {
                t.Errorf("unexpected error when getting CurrentResources(): %v", gotErr)
            }
            if gotErr == nil {
                if tc.wantErr {
                    t.Errorf("expected error but received none")
                } else {
                    if diff := cmp.Diff(tc.wantResources, gotResources); diff != "" {
                        t.Errorf("unexpected resources found, diff (-want +got):\n%s", diff)
                    }
                }
            }
        })
    }
}