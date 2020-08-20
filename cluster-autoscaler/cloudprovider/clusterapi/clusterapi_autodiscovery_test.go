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
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/labels"
)

func Test_parseAutoDiscoverySpec(t *testing.T) {
	for _, tc := range []struct {
		name    string
		spec    string
		want    *clusterAPIAutoDiscoveryConfig
		wantErr bool
	}{{
		name:    "missing ':'",
		spec:    "foo",
		wantErr: true,
	}, {
		name:    "wrong provider given",
		spec:    "asg:tag=k8s.io/cluster-autoscaler/enabled,k8s.io/cluster-autoscaler/clustername",
		wantErr: true,
	}, {
		name:    "invalid key/value pair given",
		spec:    "clusterapi:invalid",
		wantErr: true,
	}, {
		name: "no attributes specified",
		spec: "clusterapi:",
		want: &clusterAPIAutoDiscoveryConfig{
			labelSelector: labels.NewSelector(),
		},
		wantErr: false,
	}, {
		name: "only clusterName given",
		spec: "clusterapi:clusterName=foo",
		want: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			labelSelector: labels.NewSelector(),
		},
		wantErr: false,
	}, {
		name: "only namespace given",
		spec: "clusterapi:namespace=default",
		want: &clusterAPIAutoDiscoveryConfig{
			namespace:     "default",
			labelSelector: labels.NewSelector(),
		},
		wantErr: false,
	}, {
		name: "no clustername or namespace given, key provided without value",
		spec: "clusterapi:mylabel=",
		want: &clusterAPIAutoDiscoveryConfig{
			labelSelector: labels.SelectorFromSet(labels.Set{"mylabel": ""}),
		},
		wantErr: false,
	}, {
		name: "no clustername or namespace given, single key/value pair for labels",
		spec: "clusterapi:mylabel=myval",
		want: &clusterAPIAutoDiscoveryConfig{
			labelSelector: labels.SelectorFromSet(labels.Set{"mylabel": "myval"}),
		},
		wantErr: false,
	}, {
		name: "no clustername or namespace given, multiple key/value pair for labels",
		spec: "clusterapi:color=blue,shape=square",
		want: &clusterAPIAutoDiscoveryConfig{
			labelSelector: labels.SelectorFromSet(labels.Set{"color": "blue", "shape": "square"}),
		},
		wantErr: false,
	}, {
		name: "no clustername given, multiple key/value pair for labels",
		spec: "clusterapi:namespace=test,color=blue,shape=square",
		want: &clusterAPIAutoDiscoveryConfig{
			namespace:     "test",
			labelSelector: labels.SelectorFromSet(labels.Set{"color": "blue", "shape": "square"}),
		},
		wantErr: false,
	}, {
		name: "no clustername given, single key/value pair for labels",
		spec: "clusterapi:namespace=test,color=blue",
		want: &clusterAPIAutoDiscoveryConfig{
			namespace:     "test",
			labelSelector: labels.SelectorFromSet(labels.Set{"color": "blue"}),
		},
		wantErr: false,
	}, {
		name: "no namespace given, multiple key/value pair for labels",
		spec: "clusterapi:clusterName=foo,color=blue,shape=square",
		want: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			labelSelector: labels.SelectorFromSet(labels.Set{"color": "blue", "shape": "square"}),
		},
		wantErr: false,
	}, {
		name: "no namespace given, single key/value pair for labels",
		spec: "clusterapi:clusterName=foo,shape=square",
		want: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			labelSelector: labels.SelectorFromSet(labels.Set{"shape": "square"}),
		},
		wantErr: false,
	}, {
		name: "clustername, namespace, multiple key/value pair for labels provided",
		spec: "clusterapi:namespace=test,color=blue,shape=square,clusterName=foo",
		want: &clusterAPIAutoDiscoveryConfig{
			namespace:     "test",
			clusterName:   "foo",
			labelSelector: labels.SelectorFromSet(labels.Set{"color": "blue", "shape": "square"}),
		},
		wantErr: false,
	}, {
		name: "clustername, namespace, single key/value pair for labels provided",
		spec: "clusterapi:namespace=test,color=blue,clusterName=foo",
		want: &clusterAPIAutoDiscoveryConfig{
			namespace:     "test",
			clusterName:   "foo",
			labelSelector: labels.SelectorFromSet(labels.Set{"color": "blue"}),
		},
		wantErr: false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAutoDiscoverySpec(tc.spec)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseAutoDiscoverySpec() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseAutoDiscoverySpec() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_parseAutoDiscovery(t *testing.T) {
	for _, tc := range []struct {
		name    string
		spec    []string
		want    []*clusterAPIAutoDiscoveryConfig
		wantErr bool
	}{{
		name:    "contains invalid spec",
		spec:    []string{"foo", "clusterapi:color=green"},
		wantErr: true,
	}, {
		name: "clustername, namespace, single key/value pair for labels provided",
		spec: []string{
			"clusterapi:namespace=test,color=blue,clusterName=foo",
			"clusterapi:namespace=default,color=green,clusterName=bar",
		},
		want: []*clusterAPIAutoDiscoveryConfig{
			{
				namespace:     "test",
				clusterName:   "foo",
				labelSelector: labels.SelectorFromSet(labels.Set{"color": "blue"}),
			},
			{
				namespace:     "default",
				clusterName:   "bar",
				labelSelector: labels.SelectorFromSet(labels.Set{"color": "green"}),
			},
		},
		wantErr: false,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAutoDiscovery(tc.spec)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseAutoDiscoverySpec() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if len(got) != len(tc.want) {
				t.Errorf("parseAutoDiscoverySpec() expected length of got to be = %v, got %v", len(tc.want), len(got))
			}
			if err == nil && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseAutoDiscoverySpec() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_allowedByAutoDiscoverySpec(t *testing.T) {
	for _, tc := range []struct {
		name                string
		testSpec            testSpec
		autoDiscoveryConfig *clusterAPIAutoDiscoveryConfig
		additionalLabels    map[string]string
		shouldMatch         bool
	}{{
		name:                "no clustername, namespace, or label selector specified should match any MachineSet",
		testSpec:            createTestSpec(RandomString(6), RandomString(6), RandomString(6), 1, false, nil),
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{labelSelector: labels.NewSelector()},
		shouldMatch:         true,
	}, {
		name:                "no clustername, namespace, or label selector specified should match any MachineDeployment",
		testSpec:            createTestSpec(RandomString(6), RandomString(6), RandomString(6), 1, true, nil),
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{labelSelector: labels.NewSelector()},
		shouldMatch:         true,
	}, {
		name:     "clustername specified does not match MachineSet, namespace matches, no labels specified",
		testSpec: createTestSpec("default", RandomString(6), RandomString(6), 1, false, nil),
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			namespace:     "default",
			labelSelector: labels.NewSelector(),
		},
		shouldMatch: false,
	}, {
		name:     "clustername specified does not match MachineDeployment, namespace matches, no labels specified",
		testSpec: createTestSpec("default", RandomString(6), RandomString(6), 1, true, nil),
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			namespace:     "default",
			labelSelector: labels.NewSelector(),
		},
		shouldMatch: false,
	}, {
		name:     "namespace specified does not match MachineSet, clusterName matches, no labels specified",
		testSpec: createTestSpec(RandomString(6), "foo", RandomString(6), 1, false, nil),
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			namespace:     "default",
			labelSelector: labels.NewSelector(),
		},
		shouldMatch: false,
	}, {
		name:     "clustername specified does not match MachineDeployment, namespace matches, no labels specified",
		testSpec: createTestSpec(RandomString(6), "foo", RandomString(6), 1, true, nil),
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			namespace:     "default",
			labelSelector: labels.NewSelector(),
		},
		shouldMatch: false,
	}, {
		name:     "namespace and clusterName matches MachineSet, no labels specified",
		testSpec: createTestSpec("default", "foo", RandomString(6), 1, false, nil),
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			namespace:     "default",
			labelSelector: labels.NewSelector(),
		},
		shouldMatch: true,
	}, {
		name:     "namespace and clusterName matches MachineDeployment, no labels specified",
		testSpec: createTestSpec("default", "foo", RandomString(6), 1, true, nil),
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			namespace:     "default",
			labelSelector: labels.NewSelector(),
		},
		shouldMatch: true,
	}, {
		name:     "namespace and clusterName matches MachineSet, does not match label selector",
		testSpec: createTestSpec("default", "foo", RandomString(6), 1, false, nil),
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			namespace:     "default",
			labelSelector: labels.SelectorFromSet(labels.Set{"color": "green"}),
		},
		shouldMatch: false,
	}, {
		name:     "namespace and clusterName matches MachineDeployment, does not match label selector",
		testSpec: createTestSpec("default", "foo", RandomString(6), 1, true, nil),
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			namespace:     "default",
			labelSelector: labels.SelectorFromSet(labels.Set{"color": "green"}),
		},
		shouldMatch: false,
	}, {
		name:             "namespace, clusterName, and label selector matches MachineSet",
		testSpec:         createTestSpec("default", "foo", RandomString(6), 1, false, nil),
		additionalLabels: map[string]string{"color": "green"},
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			namespace:     "default",
			labelSelector: labels.SelectorFromSet(labels.Set{"color": "green"}),
		},
		shouldMatch: true,
	}, {
		name:             "namespace, clusterName, and label selector matches MachineDeployment",
		testSpec:         createTestSpec("default", "foo", RandomString(6), 1, true, nil),
		additionalLabels: map[string]string{"color": "green"},
		autoDiscoveryConfig: &clusterAPIAutoDiscoveryConfig{
			clusterName:   "foo",
			namespace:     "default",
			labelSelector: labels.SelectorFromSet(labels.Set{"color": "green"}),
		},
		shouldMatch: true,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			testConfigs := createTestConfigs(tc.testSpec)
			resource := testConfigs[0].machineSet
			if tc.testSpec.rootIsMachineDeployment {
				resource = testConfigs[0].machineDeployment
			}
			if tc.additionalLabels != nil {
				resource.SetLabels(labels.Merge(resource.GetLabels(), tc.additionalLabels))
			}
			got := allowedByAutoDiscoverySpec(tc.autoDiscoveryConfig, resource)

			if got != tc.shouldMatch {
				t.Errorf("allowedByAutoDiscoverySpec got = %v, want %v", got, tc.shouldMatch)
			}
		})
	}
}
