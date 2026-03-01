/*
Copyright 2019 The Kubernetes Authors.

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

package verdacloud

import (
	"sort"
	"testing"

	apiv1 "k8s.io/api/core/v1"
)

func TestCloudConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		cfg         cloudConfig
		expectError bool
	}{
		{
			name: "valid config",
			cfg: cloudConfig{
				Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
				SSHKeyIDs:          []string{"key-1"},
				StartupScript:      "base64script",
				AvailableLocations: []string{"FIN-01"},
			},
			expectError: false,
		},
		{
			name: "valid config without groups (groups is optional)",
			cfg: cloudConfig{
				Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
				SSHKeyIDs:          []string{"key-1"},
				StartupScript:      "base64script",
				AvailableLocations: []string{"FIN-01"},
				Labels:             []string{"env=prod"},
				Taints:             []apiv1.Taint{{Key: "dedicated", Value: "test", Effect: apiv1.TaintEffectNoSchedule}},
				// Groups is nil - should be valid
			},
			expectError: false,
		},
		{
			name: "missing image",
			cfg: cloudConfig{
				SSHKeyIDs:          []string{"key-1"},
				StartupScript:      "base64script",
				AvailableLocations: []string{"FIN-01"},
			},
			expectError: true,
		},
		{
			name: "missing ssh keys",
			cfg: cloudConfig{
				Image:              imageConfig{GPU: "gpu-image"},
				StartupScript:      "base64script",
				AvailableLocations: []string{"FIN-01"},
			},
			expectError: true,
		},
		{
			name: "missing startup script",
			cfg: cloudConfig{
				Image:              imageConfig{GPU: "gpu-image"},
				SSHKeyIDs:          []string{"key-1"},
				AvailableLocations: []string{"FIN-01"},
			},
			expectError: true,
		},
		{
			name: "missing locations",
			cfg: cloudConfig{
				Image:         imageConfig{GPU: "gpu-image"},
				SSHKeyIDs:     []string{"key-1"},
				StartupScript: "base64script",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCloudConfig_GetNodeConfig_WithoutGroups(t *testing.T) {
	globalTaint := apiv1.Taint{Key: "dedicated", Value: "backend", Effect: apiv1.TaintEffectNoSchedule}

	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1", "key-2"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=prod", "team=backend"},
		StartupScript:      "base64script",
		StartupScriptEnv:   map[string]string{"MASTER_IP": "10.0.0.1"},
		AvailableLocations: []string{"FIN-01"},
		Taints:             []apiv1.Taint{globalTaint},
	}

	nodeCfg := cfg.GetNodeConfig("any-asg-name")

	if nodeCfg.Contract != "PAY_AS_YOU_GO" {
		t.Errorf("Expected Contract=PAY_AS_YOU_GO, got %s", nodeCfg.Contract)
	}
	if nodeCfg.IsSpot {
		t.Error("Expected IsSpot=false for PAY_AS_YOU_GO")
	}
	if !containsLabel(nodeCfg.Labels, "env=prod") || !containsLabel(nodeCfg.Labels, "team=backend") {
		t.Errorf("Expected global labels, got %v", nodeCfg.Labels)
	}
	if len(nodeCfg.Taints) != 1 || nodeCfg.Taints[0].Key != "dedicated" {
		t.Errorf("Expected global taints, got %v", nodeCfg.Taints)
	}
	if nodeCfg.StartupScriptEnv["MASTER_IP"] != "10.0.0.1" {
		t.Errorf("Expected MASTER_IP=10.0.0.1, got %s", nodeCfg.StartupScriptEnv["MASTER_IP"])
	}
}

func TestCloudConfig_GetNodeConfig_UnknownAsgUsesGlobal(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=prod"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		Taints:             []apiv1.Taint{{Key: "global", Value: "taint", Effect: apiv1.TaintEffectNoSchedule}},
		Groups: map[string]GroupConfig{
			"known-group": {
				Labels: []string{"hardware=gpu"},
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("unknown-asg")

	if !containsLabel(nodeCfg.Labels, "env=prod") {
		t.Errorf("Expected global label env=prod, got %v", nodeCfg.Labels)
	}
	if len(nodeCfg.Taints) != 1 || nodeCfg.Taints[0].Key != "global" {
		t.Errorf("Expected global taints, got %v", nodeCfg.Taints)
	}
}

func TestCloudConfig_GetNodeConfig_LabelsMerge(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=prod", "cluster=main"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		Groups: map[string]GroupConfig{
			"gpu-workers": {
				Labels: []string{"hardware=gpu", "team=ai"},
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("gpu-workers")

	expectedLabels := []string{"env=prod", "cluster=main", "hardware=gpu", "team=ai"}
	for _, expected := range expectedLabels {
		if !containsLabel(nodeCfg.Labels, expected) {
			t.Errorf("Expected label %s in merged labels, got %v", expected, nodeCfg.Labels)
		}
	}
	if len(nodeCfg.Labels) != 4 {
		t.Errorf("Expected 4 merged labels, got %d: %v", len(nodeCfg.Labels), nodeCfg.Labels)
	}
}

func TestCloudConfig_GetNodeConfig_LabelsMergeWithConflict(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=prod", "team=backend", "cluster=main"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		Groups: map[string]GroupConfig{
			"gpu-workers": {
				// env conflicts with global, team conflicts with global
				Labels: []string{"env=staging", "team=ai", "hardware=gpu"},
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("gpu-workers")

	if !containsLabel(nodeCfg.Labels, "env=staging") {
		t.Errorf("Expected group label env=staging to override global, got %v", nodeCfg.Labels)
	}
	if !containsLabel(nodeCfg.Labels, "team=ai") {
		t.Errorf("Expected group label team=ai to override global, got %v", nodeCfg.Labels)
	}
	if !containsLabel(nodeCfg.Labels, "cluster=main") {
		t.Errorf("Expected global label cluster=main to be preserved, got %v", nodeCfg.Labels)
	}
	if !containsLabel(nodeCfg.Labels, "hardware=gpu") {
		t.Errorf("Expected group label hardware=gpu to be added, got %v", nodeCfg.Labels)
	}
	if len(nodeCfg.Labels) != 4 {
		t.Errorf("Expected 4 merged labels (with conflicts resolved), got %d: %v", len(nodeCfg.Labels), nodeCfg.Labels)
	}
}

func TestCloudConfig_GetNodeConfig_TaintsMerge(t *testing.T) {
	globalTaint := apiv1.Taint{Key: "dedicated", Value: "backend", Effect: apiv1.TaintEffectNoSchedule}
	groupTaint := apiv1.Taint{Key: "nvidia.com/gpu", Value: "present", Effect: apiv1.TaintEffectNoSchedule}

	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=prod"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		Taints:             []apiv1.Taint{globalTaint},
		Groups: map[string]GroupConfig{
			"gpu-workers": {
				Taints: []apiv1.Taint{groupTaint},
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("gpu-workers")

	if len(nodeCfg.Taints) != 2 {
		t.Fatalf("Expected 2 taints (global + group), got %d", len(nodeCfg.Taints))
	}
	if !containsTaint(nodeCfg.Taints, "dedicated") {
		t.Errorf("Expected global taint 'dedicated', got %v", nodeCfg.Taints)
	}
	if !containsTaint(nodeCfg.Taints, "nvidia.com/gpu") {
		t.Errorf("Expected group taint 'nvidia.com/gpu', got %v", nodeCfg.Taints)
	}
}

func TestCloudConfig_GetNodeConfig_TaintsMergeWithConflict(t *testing.T) {
	globalTaint1 := apiv1.Taint{Key: "dedicated", Value: "backend", Effect: apiv1.TaintEffectNoSchedule}
	globalTaint2 := apiv1.Taint{Key: "tier", Value: "standard", Effect: apiv1.TaintEffectNoSchedule}
	groupTaint1 := apiv1.Taint{Key: "dedicated", Value: "gpu-workload", Effect: apiv1.TaintEffectNoExecute}
	groupTaint2 := apiv1.Taint{Key: "nvidia.com/gpu", Value: "present", Effect: apiv1.TaintEffectNoSchedule}

	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=prod"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		Taints:             []apiv1.Taint{globalTaint1, globalTaint2},
		Groups: map[string]GroupConfig{
			"gpu-workers": {
				Taints: []apiv1.Taint{groupTaint1, groupTaint2},
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("gpu-workers")

	if len(nodeCfg.Taints) != 3 {
		t.Fatalf("Expected 3 taints (with conflict resolved), got %d: %v", len(nodeCfg.Taints), nodeCfg.Taints)
	}

	for _, taint := range nodeCfg.Taints {
		if taint.Key == "dedicated" {
			if taint.Value != "gpu-workload" {
				t.Errorf("Expected group taint value 'gpu-workload' for key 'dedicated', got %s", taint.Value)
			}
			if taint.Effect != apiv1.TaintEffectNoExecute {
				t.Errorf("Expected group taint effect NoExecute for key 'dedicated', got %s", taint.Effect)
			}
		}
	}

	if !containsTaint(nodeCfg.Taints, "tier") {
		t.Errorf("Expected global taint 'tier' to be preserved, got %v", nodeCfg.Taints)
	}
	if !containsTaint(nodeCfg.Taints, "nvidia.com/gpu") {
		t.Errorf("Expected group taint 'nvidia.com/gpu', got %v", nodeCfg.Taints)
	}
}

func TestCloudConfig_GetNodeConfig_BillingConfigOverride(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=prod"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		Groups: map[string]GroupConfig{
			"spot-workers": {
				BillingConfig: &billingConfig{Price: "DYNAMIC", Contract: "SPOT"},
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("spot-workers")

	if nodeCfg.Contract != "SPOT" {
		t.Errorf("Expected Contract=SPOT, got %s", nodeCfg.Contract)
	}
	if !nodeCfg.IsSpot {
		t.Error("Expected IsSpot=true for SPOT contract")
	}

	if !containsLabel(nodeCfg.Labels, "env=prod") {
		t.Errorf("Expected global label env=prod, got %v", nodeCfg.Labels)
	}
}

func TestCloudConfig_GetNodeConfig_AllFieldsMerge(t *testing.T) {
	globalTaint := apiv1.Taint{Key: "global", Value: "taint", Effect: apiv1.TaintEffectNoSchedule}
	groupTaint := apiv1.Taint{Key: "nvidia.com/gpu", Value: "present", Effect: apiv1.TaintEffectNoSchedule}

	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=prod"},
		StartupScript:      "base64script",
		StartupScriptEnv:   map[string]string{"MASTER_IP": "10.0.0.1"},
		AvailableLocations: []string{"FIN-01"},
		Taints:             []apiv1.Taint{globalTaint},
		Groups: map[string]GroupConfig{
			"gpu-spot-workers": {
				Labels:        []string{"hardware=gpu"},
				Taints:        []apiv1.Taint{groupTaint},
				BillingConfig: &billingConfig{Price: "DYNAMIC", Contract: "SPOT"},
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("gpu-spot-workers")

	if nodeCfg.Contract != "SPOT" {
		t.Errorf("Expected Contract=SPOT, got %s", nodeCfg.Contract)
	}
	if !nodeCfg.IsSpot {
		t.Error("Expected IsSpot=true")
	}

	if !containsLabel(nodeCfg.Labels, "env=prod") {
		t.Errorf("Expected global label env=prod, got %v", nodeCfg.Labels)
	}
	if !containsLabel(nodeCfg.Labels, "hardware=gpu") {
		t.Errorf("Expected group label hardware=gpu, got %v", nodeCfg.Labels)
	}

	if len(nodeCfg.Taints) != 2 {
		t.Fatalf("Expected 2 taints, got %d", len(nodeCfg.Taints))
	}
	if !containsTaint(nodeCfg.Taints, "global") {
		t.Errorf("Expected global taint, got %v", nodeCfg.Taints)
	}
	if !containsTaint(nodeCfg.Taints, "nvidia.com/gpu") {
		t.Errorf("Expected group taint, got %v", nodeCfg.Taints)
	}

	if nodeCfg.StartupScriptEnv["MASTER_IP"] != "10.0.0.1" {
		t.Errorf("Expected global MASTER_IP, got %s", nodeCfg.StartupScriptEnv["MASTER_IP"])
	}
}

func TestCloudConfig_GetNodeConfig_EmptyGroupDoesNotMerge(t *testing.T) {
	globalTaint := apiv1.Taint{Key: "global", Value: "taint", Effect: apiv1.TaintEffectNoSchedule}

	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=prod"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		Taints:             []apiv1.Taint{globalTaint},
		Groups: map[string]GroupConfig{
			"empty-group": {},
		},
	}

	nodeCfg := cfg.GetNodeConfig("empty-group")

	if nodeCfg.Contract != "PAY_AS_YOU_GO" {
		t.Errorf("Expected global Contract, got %s", nodeCfg.Contract)
	}
	if !containsLabel(nodeCfg.Labels, "env=prod") {
		t.Errorf("Expected global label env=prod, got %v", nodeCfg.Labels)
	}
	if len(nodeCfg.Taints) != 1 || nodeCfg.Taints[0].Key != "global" {
		t.Errorf("Expected global taint only, got %v", nodeCfg.Taints)
	}
}

func TestCloudConfig_GetNodeConfig_NoMutation(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=prod"},
		StartupScript:      "base64script",
		StartupScriptEnv:   map[string]string{"MASTER_IP": "10.0.0.1"},
		AvailableLocations: []string{"FIN-01"},
		Taints:             []apiv1.Taint{{Key: "global", Value: "taint", Effect: apiv1.TaintEffectNoSchedule}},
		Groups: map[string]GroupConfig{
			"test-group": {
				Labels: []string{"team=test"},
				Taints: []apiv1.Taint{{Key: "group", Value: "taint", Effect: apiv1.TaintEffectNoSchedule}},
			},
		},
	}

	// Get config multiple times
	_ = cfg.GetNodeConfig("test-group")
	_ = cfg.GetNodeConfig("test-group")
	nodeCfg := cfg.GetNodeConfig("test-group")

	if len(cfg.Labels) != 1 {
		t.Errorf("Original Labels was mutated: expected 1, got %d", len(cfg.Labels))
	}
	if len(cfg.Taints) != 1 {
		t.Errorf("Original Taints was mutated: expected 1, got %d", len(cfg.Taints))
	}

	if len(nodeCfg.Labels) != 2 {
		t.Errorf("Expected 2 labels in result, got %d", len(nodeCfg.Labels))
	}
	if len(nodeCfg.Taints) != 2 {
		t.Errorf("Expected 2 taints in result, got %d", len(nodeCfg.Taints))
	}
}

func TestMergeLabels(t *testing.T) {
	tests := []struct {
		name     string
		global   []string
		group    []string
		expected []string
	}{
		{
			name:     "global only",
			global:   []string{"env=prod", "team=backend"},
			group:    nil,
			expected: []string{"env=prod", "team=backend"},
		},
		{
			name:     "group only",
			global:   nil,
			group:    []string{"hardware=gpu"},
			expected: []string{"hardware=gpu"},
		},
		{
			name:     "merge without conflict",
			global:   []string{"env=prod"},
			group:    []string{"hardware=gpu"},
			expected: []string{"env=prod", "hardware=gpu"},
		},
		{
			name:     "merge with conflict - group wins",
			global:   []string{"env=prod", "team=backend"},
			group:    []string{"env=staging", "hardware=gpu"},
			expected: []string{"env=staging", "team=backend", "hardware=gpu"},
		},
		{
			name:     "both empty",
			global:   nil,
			group:    nil,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeLabels(tt.global, tt.group)

			sort.Strings(result)
			sort.Strings(tt.expected)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d labels, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, label := range tt.expected {
				if result[i] != label {
					t.Errorf("Expected label %s at index %d, got %s", label, i, result[i])
				}
			}
		})
	}
}

func TestMergeTaints(t *testing.T) {
	tests := []struct {
		name     string
		global   []apiv1.Taint
		group    []apiv1.Taint
		expected []apiv1.Taint
	}{
		{
			name:     "global only",
			global:   []apiv1.Taint{{Key: "dedicated", Value: "backend", Effect: apiv1.TaintEffectNoSchedule}},
			group:    nil,
			expected: []apiv1.Taint{{Key: "dedicated", Value: "backend", Effect: apiv1.TaintEffectNoSchedule}},
		},
		{
			name:     "group only",
			global:   nil,
			group:    []apiv1.Taint{{Key: "nvidia.com/gpu", Value: "present", Effect: apiv1.TaintEffectNoSchedule}},
			expected: []apiv1.Taint{{Key: "nvidia.com/gpu", Value: "present", Effect: apiv1.TaintEffectNoSchedule}},
		},
		{
			name:   "merge with conflict - group wins",
			global: []apiv1.Taint{{Key: "dedicated", Value: "backend", Effect: apiv1.TaintEffectNoSchedule}},
			group:  []apiv1.Taint{{Key: "dedicated", Value: "gpu-workload", Effect: apiv1.TaintEffectNoExecute}},
			expected: []apiv1.Taint{
				{Key: "dedicated", Value: "gpu-workload", Effect: apiv1.TaintEffectNoExecute},
			},
		},
		{
			name:     "both empty",
			global:   nil,
			group:    nil,
			expected: []apiv1.Taint{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeTaints(tt.global, tt.group)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d taints, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for _, expectedTaint := range tt.expected {
				found := false
				for _, resultTaint := range result {
					if resultTaint.Key == expectedTaint.Key &&
						resultTaint.Value == expectedTaint.Value &&
						resultTaint.Effect == expectedTaint.Effect {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected taint %v not found in result %v", expectedTaint, result)
				}
			}
		})
	}
}

func TestCloudConfig_GetNodeConfig_OSVolumeSize_Default(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
	}

	nodeCfg := cfg.GetNodeConfig("any-asg")

	if nodeCfg.OSVolumeSize != 50 {
		t.Errorf("Expected default OSVolumeSize=50, got %d", nodeCfg.OSVolumeSize)
	}
}

func TestCloudConfig_GetNodeConfig_OSVolumeSize_GlobalOverride(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		OSVolumeSize:       100,
	}

	nodeCfg := cfg.GetNodeConfig("any-asg")

	if nodeCfg.OSVolumeSize != 100 {
		t.Errorf("Expected global OSVolumeSize=100, got %d", nodeCfg.OSVolumeSize)
	}
}

func TestCloudConfig_GetNodeConfig_OSVolumeSize_BelowMinimum(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		OSVolumeSize:       30,
	}

	nodeCfg := cfg.GetNodeConfig("any-asg")

	if nodeCfg.OSVolumeSize != 50 {
		t.Errorf("Expected OSVolumeSize=50 (below minimum should use default), got %d", nodeCfg.OSVolumeSize)
	}
}

func TestCloudConfig_GetNodeConfig_OSVolumeSize_GroupOverride(t *testing.T) {
	osVolSize := 200
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		OSVolumeSize:       100,
		Groups: map[string]GroupConfig{
			"gpu-workers": {
				OSVolumeSize: &osVolSize,
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("gpu-workers")

	if nodeCfg.OSVolumeSize != 200 {
		t.Errorf("Expected group OSVolumeSize=200, got %d", nodeCfg.OSVolumeSize)
	}
}

func TestCloudConfig_GetNodeConfig_OSVolumeSize_GroupBelowMinimum(t *testing.T) {
	osVolSize := 30
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		OSVolumeSize:       100,
		Groups: map[string]GroupConfig{
			"test-workers": {
				OSVolumeSize: &osVolSize,
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("test-workers")

	if nodeCfg.OSVolumeSize != 100 {
		t.Errorf("Expected global OSVolumeSize=100 (group below minimum), got %d", nodeCfg.OSVolumeSize)
	}
}

func TestCloudConfig_GetNodeConfig_OSVolumeSize_Precedence(t *testing.T) {
	tests := []struct {
		name            string
		globalSize      int
		groupSize       *int
		expectedSize    int
		expectedAsgName string
	}{
		{
			name:            "default only (no global, no group)",
			globalSize:      0,
			groupSize:       nil,
			expectedSize:    50,
			expectedAsgName: "test-asg",
		},
		{
			name:            "global overrides default",
			globalSize:      80,
			groupSize:       nil,
			expectedSize:    80,
			expectedAsgName: "test-asg",
		},
		{
			name:            "group overrides global",
			globalSize:      80,
			groupSize:       intPtr(150),
			expectedSize:    150,
			expectedAsgName: "test-asg",
		},
		{
			name:            "group overrides default (no global)",
			globalSize:      0,
			groupSize:       intPtr(120),
			expectedSize:    120,
			expectedAsgName: "test-asg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &cloudConfig{
				Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
				SSHKeyIDs:          []string{"key-1"},
				BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
				StartupScript:      "base64script",
				AvailableLocations: []string{"FIN-01"},
				OSVolumeSize:       tt.globalSize,
			}

			if tt.groupSize != nil {
				cfg.Groups = map[string]GroupConfig{
					tt.expectedAsgName: {
						OSVolumeSize: tt.groupSize,
					},
				}
			}

			nodeCfg := cfg.GetNodeConfig(tt.expectedAsgName)

			if nodeCfg.OSVolumeSize != tt.expectedSize {
				t.Errorf("Expected OSVolumeSize=%d, got %d", tt.expectedSize, nodeCfg.OSVolumeSize)
			}
		})
	}
}

func TestCloudConfig_GetNodeConfig_AvailableLocations_Default(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-02", "FIN-03"},
	}

	nodeCfg := cfg.GetNodeConfig("any-asg")

	if len(nodeCfg.AvailableLocations) != 2 {
		t.Fatalf("Expected 2 locations, got %d", len(nodeCfg.AvailableLocations))
	}
	if !containsString(nodeCfg.AvailableLocations, "FIN-02") {
		t.Errorf("Expected FIN-02 in locations, got %v", nodeCfg.AvailableLocations)
	}
	if !containsString(nodeCfg.AvailableLocations, "FIN-03") {
		t.Errorf("Expected FIN-03 in locations, got %v", nodeCfg.AvailableLocations)
	}
}

func TestCloudConfig_GetNodeConfig_AvailableLocations_GroupReplaces(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-02", "FIN-03"},
		Groups: map[string]GroupConfig{
			"gpu-workers": {
				AvailableLocations: []string{"FIN-03"},
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("gpu-workers")

	if len(nodeCfg.AvailableLocations) != 1 {
		t.Fatalf("Expected 1 location (group replaces global), got %d: %v", len(nodeCfg.AvailableLocations), nodeCfg.AvailableLocations)
	}
	if nodeCfg.AvailableLocations[0] != "FIN-03" {
		t.Errorf("Expected FIN-03, got %s", nodeCfg.AvailableLocations[0])
	}
}

func TestCloudConfig_GetNodeConfig_AvailableLocations_UnknownAsgUsesGlobal(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-02", "FIN-03"},
		Groups: map[string]GroupConfig{
			"gpu-workers": {
				AvailableLocations: []string{"FIN-03"},
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("cpu-workers")

	if len(nodeCfg.AvailableLocations) != 2 {
		t.Fatalf("Expected 2 locations (global), got %d", len(nodeCfg.AvailableLocations))
	}
	if !containsString(nodeCfg.AvailableLocations, "FIN-02") {
		t.Errorf("Expected FIN-02 in locations, got %v", nodeCfg.AvailableLocations)
	}
	if !containsString(nodeCfg.AvailableLocations, "FIN-03") {
		t.Errorf("Expected FIN-03 in locations, got %v", nodeCfg.AvailableLocations)
	}
}

func TestCloudConfig_GetNodeConfig_AdditionalVolumes_GroupOnly(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
		Groups: map[string]GroupConfig{
			"gpu-workers": {
				AdditionalVolumes: []additionalVolume{
					{Name: "data", Size: 500, Type: "SSD"},
					{Name: "scratch", Size: 200, Type: "SSD"},
				},
			},
		},
	}

	nodeCfg := cfg.GetNodeConfig("gpu-workers")
	if len(nodeCfg.Volumes) != 2 {
		t.Fatalf("Expected 2 volumes for gpu-workers, got %d", len(nodeCfg.Volumes))
	}
	if nodeCfg.Volumes[0].Name != "data" || nodeCfg.Volumes[0].Size != 500 {
		t.Errorf("Expected data volume (500GB), got %+v", nodeCfg.Volumes[0])
	}
	if nodeCfg.Volumes[1].Name != "scratch" || nodeCfg.Volumes[1].Size != 200 {
		t.Errorf("Expected scratch volume (200GB), got %+v", nodeCfg.Volumes[1])
	}

	nodeCfg2 := cfg.GetNodeConfig("cpu-workers")
	if len(nodeCfg2.Volumes) != 0 {
		t.Errorf("Expected 0 volumes for cpu-workers, got %d", len(nodeCfg2.Volumes))
	}
}

func TestCloudConfig_GetNodeConfig_AdditionalVolumes_NoGlobal(t *testing.T) {
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-01"},
	}

	nodeCfg := cfg.GetNodeConfig("any-asg")

	if len(nodeCfg.Volumes) != 0 {
		t.Errorf("Expected 0 volumes (no global, no group), got %d", len(nodeCfg.Volumes))
	}
}

func TestCloudConfig_GetNodeConfig_FullIntegration(t *testing.T) {
	osVolSize := 200
	cfg := &cloudConfig{
		Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
		SSHKeyIDs:          []string{"key-1"},
		BillingConfig:      billingConfig{Price: "DYNAMIC", Contract: "PAY_AS_YOU_GO"},
		Labels:             []string{"env=production"},
		StartupScript:      "base64script",
		AvailableLocations: []string{"FIN-02", "FIN-03"},
		OSVolumeSize:       100,
		Groups: map[string]GroupConfig{
			"gpu-workers": {
				OSVolumeSize:       &osVolSize,
				AvailableLocations: []string{"FIN-03"},
				AdditionalVolumes: []additionalVolume{
					{Name: "data", Size: 500, Type: "SSD"},
				},
				Labels: []string{"hardware=gpu"},
			},
			"cpu-workers": {
				OSVolumeSize: intPtr(80),
				Labels:       []string{"hardware=cpu"},
			},
		},
	}

	// Test gpu-workers
	gpuCfg := cfg.GetNodeConfig("gpu-workers")
	if gpuCfg.OSVolumeSize != 200 {
		t.Errorf("gpu-workers: Expected OSVolumeSize=200, got %d", gpuCfg.OSVolumeSize)
	}
	if len(gpuCfg.AvailableLocations) != 1 || gpuCfg.AvailableLocations[0] != "FIN-03" {
		t.Errorf("gpu-workers: Expected locations=[FIN-03], got %v", gpuCfg.AvailableLocations)
	}
	if len(gpuCfg.Volumes) != 1 {
		t.Errorf("gpu-workers: Expected 1 volume, got %d", len(gpuCfg.Volumes))
	}
	if !containsLabel(gpuCfg.Labels, "hardware=gpu") {
		t.Errorf("gpu-workers: Expected hardware=gpu label, got %v", gpuCfg.Labels)
	}

	cpuCfg := cfg.GetNodeConfig("cpu-workers")
	if cpuCfg.OSVolumeSize != 80 {
		t.Errorf("cpu-workers: Expected OSVolumeSize=80, got %d", cpuCfg.OSVolumeSize)
	}
	if len(cpuCfg.AvailableLocations) != 2 {
		t.Errorf("cpu-workers: Expected 2 locations (global), got %d: %v", len(cpuCfg.AvailableLocations), cpuCfg.AvailableLocations)
	}
	if len(cpuCfg.Volumes) != 0 {
		t.Errorf("cpu-workers: Expected 0 volumes, got %d", len(cpuCfg.Volumes))
	}
	if !containsLabel(cpuCfg.Labels, "hardware=cpu") {
		t.Errorf("cpu-workers: Expected hardware=cpu label, got %v", cpuCfg.Labels)
	}

	unknownCfg := cfg.GetNodeConfig("unknown-asg")
	if unknownCfg.OSVolumeSize != 100 {
		t.Errorf("unknown-asg: Expected OSVolumeSize=100 (global), got %d", unknownCfg.OSVolumeSize)
	}
	if len(unknownCfg.AvailableLocations) != 2 {
		t.Errorf("unknown-asg: Expected 2 locations (global), got %d", len(unknownCfg.AvailableLocations))
	}
	if len(unknownCfg.Volumes) != 0 {
		t.Errorf("unknown-asg: Expected 0 volumes, got %d", len(unknownCfg.Volumes))
	}
}

func containsLabel(labels []string, target string) bool {
	for _, label := range labels {
		if label == target {
			return true
		}
	}
	return false
}

// assertContainsLabel fails the test if the target label is not found
func assertContainsLabel(t *testing.T, labels []string, target string) {
	t.Helper()
	if !containsLabel(labels, target) {
		t.Errorf("expected label %q in %v", target, labels)
	}
}

func containsTaint(taints []apiv1.Taint, key string) bool {
	for _, taint := range taints {
		if taint.Key == key {
			return true
		}
	}
	return false
}

func containsString(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

func intPtr(i int) *int {
	return &i
}

func TestCloudConfig_IsValid(t *testing.T) {
	tests := []struct {
		name          string
		cfg           cloudConfig
		expectedValid bool
	}{
		{
			name: "valid config returns true",
			cfg: cloudConfig{
				Image:              imageConfig{GPU: "gpu-image", CPU: "cpu-image"},
				SSHKeyIDs:          []string{"key-1"},
				StartupScript:      "base64script",
				AvailableLocations: []string{"FIN-01"},
			},
			expectedValid: true,
		},
		{
			name: "missing image returns false",
			cfg: cloudConfig{
				SSHKeyIDs:          []string{"key-1"},
				StartupScript:      "base64script",
				AvailableLocations: []string{"FIN-01"},
			},
			expectedValid: false,
		},
		{
			name: "missing ssh keys returns false",
			cfg: cloudConfig{
				Image:              imageConfig{GPU: "gpu-image"},
				StartupScript:      "base64script",
				AvailableLocations: []string{"FIN-01"},
			},
			expectedValid: false,
		},
		{
			name: "missing startup script returns false",
			cfg: cloudConfig{
				Image:              imageConfig{GPU: "gpu-image"},
				SSHKeyIDs:          []string{"key-1"},
				AvailableLocations: []string{"FIN-01"},
			},
			expectedValid: false,
		},
		{
			name: "missing locations returns false",
			cfg: cloudConfig{
				Image:         imageConfig{GPU: "gpu-image"},
				SSHKeyIDs:     []string{"key-1"},
				StartupScript: "base64script",
			},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.isValid()
			if result != tt.expectedValid {
				t.Errorf("Expected isValid()=%v, got %v", tt.expectedValid, result)
			}
		})
	}
}
