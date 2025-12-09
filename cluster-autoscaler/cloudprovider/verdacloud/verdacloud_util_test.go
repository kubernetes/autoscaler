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
	"strings"
	"testing"
)

// Test constants for util tests
const (
	testUtilInstanceType   = "CPU.4V.16G"
	testUtilGPUType        = "1H100.80S.22V"
	testUtilLocation       = "FIN-03"
	testUtilAsgName        = "asg-test"
	testUtilProviderPrefix = "verdacloud://"
)

func TestParseAsgSpec_Valid(t *testing.T) {
	spec := "1:10:" + testUtilInstanceType + ":" + testUtilAsgName
	got, err := parseAsgSpec(spec)
	if err != nil {
		t.Fatalf("parseAsgSpec returned error: %v", err)
	}
	if got.minSize != 1 || got.maxSize != 10 || got.instanceType != testUtilInstanceType || got.name != testUtilAsgName {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestParseAsgSpec_InvalidFormat(t *testing.T) {
	_, err := parseAsgSpec("1:10:" + testUtilInstanceType) // only 3 parts
	if err == nil {
		t.Fatalf("expected error for invalid format, got nil")
	}
}

func TestInstanceRefFromProviderId_Valid(t *testing.T) {
	pid := testUtilProviderPrefix + "FIN-03/asg-x-77-FIN-03-1700000000"
	ref, err := instanceRefFromProviderId(pid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ProviderID != pid {
		t.Fatalf("providerID mismatch: %s", ref.ProviderID)
	}
	if ref.Hostname != "asg-x-77-FIN-03-1700000000" {
		t.Fatalf("hostname mismatch: %s", ref.Hostname)
	}
}

func TestInstanceRefFromProviderId_Invalid(t *testing.T) {
	_, err := instanceRefFromProviderId("verdacloud://malformed")
	if err == nil {
		t.Fatalf("expected error for malformed provider id, got nil")
	}
}

func TestExtractAsgNameFromHostname_NewFormat(t *testing.T) {
	host := "asg-prod-vm-fin-03-77"
	name, err := extractAsgNameFromHostname(host)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "asg-prod" {
		t.Fatalf("expected asg-prod, got %s", name)
	}
}

func TestExtractAsgNameFromHostname_NoMagic(t *testing.T) {
	_, err := extractAsgNameFromHostname("asg-prod-FIN-03-1700000000")
	if err == nil {
		t.Fatalf("expected error when magic separator is missing")
	}
}

func TestIsGPUInstanceType(t *testing.T) {
	if !isGPUInstanceType(testUtilGPUType) {
		t.Fatalf("expected GPU type to be detected")
	}
	if isGPUInstanceType(testUtilInstanceType) {
		t.Fatalf("expected CPU type not to be detected as GPU")
	}
}

func TestConvertConfigLabelsToK8sLabels(t *testing.T) {
	// Test CPU instance - should NOT have accelerator label
	cpuAsg := &Asg{AsgRef: AsgRef{Name: "asg-cpu"}, instanceType: testUtilInstanceType}
	cpuLabels := convertConfigLabelsToK8sLabels([]string{"env=prod"}, cpuAsg)
	if !containsAllSubstrings(t, cpuLabels, []string{"env=prod", NodeGroupLabelKey + "=asg-cpu"}) {
		t.Fatalf("unexpected CPU labels: %s", cpuLabels)
	}
	// CPU nodes should NOT have accelerator label
	if strings.Contains(cpuLabels, AcceleratorLabel) {
		t.Fatalf("CPU node should not have accelerator label: %s", cpuLabels)
	}

	// Test GPU instance - should have accelerator label
	gpuAsg := &Asg{AsgRef: AsgRef{Name: "asg-gpu"}, instanceType: testUtilGPUType}
	gpuLabels := convertConfigLabelsToK8sLabels([]string{"env=prod"}, gpuAsg)
	if !containsAllSubstrings(t, gpuLabels, []string{"env=prod", NodeGroupLabelKey + "=asg-gpu", AcceleratorLabel + "=" + testUtilGPUType}) {
		t.Fatalf("unexpected GPU labels: %s", gpuLabels)
	}
}

// containsAllSubstrings checks if string s contains all substrings in wants
func containsAllSubstrings(t *testing.T, s string, wants []string) bool {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(s, w) {
			return false
		}
	}
	return true
}

func TestConvertConfigLabelsToK8sLabels_ComplexLabels(t *testing.T) {
	tests := []struct {
		name        string
		inputLabels []string
		asg         *Asg
		expectAll   []string
	}{
		{
			name: "labels with dots and slashes",
			inputLabels: []string{
				"datacrunch.io/gpu.installed=8",
				"datacrunch.io/size-1-gpu=true",
				"kubernetes.io/role=8.6000PRO",
			},
			asg: &Asg{AsgRef: AsgRef{Name: "asg-pro6000-8x"}, instanceType: "GPU.6000PRO.x8"},
			expectAll: []string{
				"datacrunch.io/gpu.installed=8",
				"datacrunch.io/size-1-gpu=true",
				"kubernetes.io/role=8.6000PRO",
				NodeGroupLabelKey + "=asg-pro6000-8x",
				AcceleratorLabel + "=GPU.6000PRO.x8",
			},
		},
		{
			name: "labels with hyphens and numbers",
			inputLabels: []string{
				"node-type=gpu-worker",
				"version=v1.2.3",
				"count=100",
			},
			asg: &Asg{AsgRef: AsgRef{Name: "gpu-pool"}, instanceType: testUtilGPUType},
			expectAll: []string{
				"node-type=gpu-worker",
				"version=v1.2.3",
				"count=100",
				NodeGroupLabelKey + "=gpu-pool",
				AcceleratorLabel + "=" + testUtilGPUType,
			},
		},
		{
			name: "labels with underscores and colons",
			inputLabels: []string{
				"my_label=value_1",
				"url=http://example.com",
			},
			asg: &Asg{AsgRef: AsgRef{Name: "test-asg"}, instanceType: "CPU.c4m8"},
			expectAll: []string{
				"my_label=value_1",
				"url=http://example.com",
				NodeGroupLabelKey + "=test-asg",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertConfigLabelsToK8sLabels(tc.inputLabels, tc.asg)

			for _, expected := range tc.expectAll {
				if !strings.Contains(result, expected) {
					t.Errorf("expected label %q not found in result: %s", expected, result)
				}
			}

			if !strings.Contains(result, ",") && len(tc.inputLabels) > 0 {
				t.Errorf("expected comma-separated labels, got: %s", result)
			}
		})
	}
}

func TestEscapeDoubleQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special chars",
			input:    "datacrunch.io/gpu=8",
			expected: "datacrunch.io/gpu=8",
		},
		{
			name:     "with dollar sign",
			input:    "value=$HOME",
			expected: `value=\$HOME`,
		},
		{
			name:     "with quotes",
			input:    `value="test"`,
			expected: `value=\"test\"`,
		},
		{
			name:     "with backslash",
			input:    `path=C:\Users`,
			expected: `path=C:\\Users`,
		},
		{
			name:     "complex label string",
			input:    "key1=val1,key2=val2,special=$VAR",
			expected: `key1=val1,key2=val2,special=\$VAR`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := escapeDoubleQuotes(tc.input)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestPatchScript_LabelsInjection(t *testing.T) {
	tests := []struct {
		name           string
		labels         []string
		providerID     string
		expectContains []string
	}{
		{
			name: "labels with dots and slashes",
			labels: []string{
				"datacrunch.io/gpu.installed=8",
				"datacrunch.io/size-1-gpu=true",
				"kubernetes.io/role=8.6000PRO",
			},
			providerID: testUtilProviderPrefix + "FIN-03/node-123",
			expectContains: []string{
				`PROVIDER_ID="` + testUtilProviderPrefix + `FIN-03/node-123"`,
				`datacrunch.io/gpu.installed=8`,
				`datacrunch.io/size-1-gpu=true`,
				`kubernetes.io/role=8.6000PRO`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			script := []byte(`#!/usr/bin/env bash
PROVIDER_ID=""
LABELS=""
`)

			asg := &Asg{
				AsgRef:       AsgRef{Name: "test-asg"},
				instanceType: testUtilGPUType,
			}

			labelsStr := convertConfigLabelsToK8sLabels(tc.labels, asg)

			envMap := map[string]string{
				"PROVIDER_ID": tc.providerID,
				"LABELS":      labelsStr,
			}

			patched := patchScript(script, envMap)
			patchedStr := string(patched)

			for _, expected := range tc.expectContains {
				if !strings.Contains(patchedStr, expected) {
					t.Errorf("expected patched script to contain %q\nGot:\n%s", expected, patchedStr)
				}
			}
		})
	}
}
