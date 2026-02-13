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
			result := escapeForDoubleQuotes(tc.input)
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

			patched := injectEnvVarsIntoScript(script, envMap)
			patchedStr := string(patched)

			for _, expected := range tc.expectContains {
				if !strings.Contains(patchedStr, expected) {
					t.Errorf("expected patched script to contain %q\nGot:\n%s", expected, patchedStr)
				}
			}
		})
	}
}

// =============================================================================
// Shell Utility Function Tests
// =============================================================================

func TestIsShellSafeUnquoted(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Safe values (alphanumeric + allowed special chars)
		{"empty string", "", true},
		{"alphanumeric", "abc123", true},
		{"with underscore", "my_var", true},
		{"with dot", "file.txt", true},
		{"with slash", "/usr/bin/bash", true},
		{"with colon", "http://example.com", true},
		{"with hyphen", "my-value", true},
		{"with at sign", "user@host", true},
		{"mixed safe chars", "node.kubernetes.io/role", true},
		{"path with multiple dots", "/path/to/file.tar.gz", true},

		// Unsafe values (require quoting)
		{"with space", "hello world", false},
		{"with dollar", "$HOME", false},
		{"with double quote", `say "hello"`, false},
		{"with single quote", "it's here", false},
		{"with backtick", "`cmd`", false},
		{"with pipe", "cmd|other", false},
		{"with semicolon", "cmd;other", false},
		{"with ampersand", "cmd&", false},
		{"with parentheses", "(subshell)", false},
		{"with brackets", "[array]", false},
		{"with asterisk", "*.txt", false},
		{"with question mark", "file?.txt", false},
		{"with newline", "line1\nline2", false},
		{"with tab", "col1\tcol2", false},
		{"with backslash", `path\to\file`, false},
		{"with equals in value", "key=value", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isShellSafeUnquoted(tc.input)
			if result != tc.expected {
				t.Errorf("isShellSafeUnquoted(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestDetectShellQuoteStyle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected shellQuoteStyle
	}{
		// No quotes
		{"empty string", "", shellQuoteNone},
		{"bare word", "value", shellQuoteNone},
		{"bare path", "/usr/bin/bash", shellQuoteNone},

		// Double quotes
		{"double quoted", `"hello"`, shellQuoteDouble},
		{"double quoted with space", `"hello world"`, shellQuoteDouble},
		{"double quoted empty", `""`, shellQuoteDouble},
		{"double quoted with var", `"$HOME"`, shellQuoteDouble},

		// Single quotes
		{"single quoted", `'hello'`, shellQuoteSingle},
		{"single quoted with space", `'hello world'`, shellQuoteSingle},
		{"single quoted empty", `''`, shellQuoteSingle},
		{"single quoted literal dollar", `'$HOME'`, shellQuoteSingle},

		// Edge cases - not properly quoted
		{"only opening double", `"hello`, shellQuoteNone},
		{"only closing double", `hello"`, shellQuoteNone},
		{"only opening single", `'hello`, shellQuoteNone},
		{"only closing single", `hello'`, shellQuoteNone},
		{"mixed quotes start double", `"hello'`, shellQuoteNone},
		{"mixed quotes start single", `'hello"`, shellQuoteNone},

		// With whitespace
		{"double quoted with leading space", `  "value"`, shellQuoteDouble},
		{"single quoted with trailing space", `'value'  `, shellQuoteSingle},

		// Windows line endings (carriage return)
		{"double quoted with CR", "\"value\"\r", shellQuoteDouble},
		{"single quoted with CR", "'value'\r", shellQuoteSingle},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectShellQuoteStyle(tc.input)
			if result != tc.expected {
				t.Errorf("detectShellQuoteStyle(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatWithShellQuotes(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		style    shellQuoteStyle
		expected string
	}{
		// shellQuoteNone - defaults to double quotes
		{"none style simple", "hello", shellQuoteNone, `"hello"`},
		{"none style with space", "hello world", shellQuoteNone, `"hello world"`},
		{"none style with dollar", "$HOME", shellQuoteNone, `"\$HOME"`},

		// shellQuoteDouble
		{"double style simple", "hello", shellQuoteDouble, `"hello"`},
		{"double style with dollar", "$VAR", shellQuoteDouble, `"\$VAR"`},
		{"double style with quote", `say "hi"`, shellQuoteDouble, `"say \"hi\""`},
		{"double style with backslash", `path\to`, shellQuoteDouble, `"path\\to"`},
		{"double style complex", `$HOME/"dir"`, shellQuoteDouble, `"\$HOME/\"dir\""`},

		// shellQuoteSingle
		{"single style simple", "hello", shellQuoteSingle, `'hello'`},
		{"single style with dollar", "$VAR", shellQuoteSingle, `'$VAR'`},
		{"single style with space", "hello world", shellQuoteSingle, `'hello world'`},
		// Single quote with embedded single quote falls back to double quotes
		{"single style with single quote", "it's", shellQuoteSingle, `"it's"`},
		{"single style with single quote and dollar", "it's $HOME", shellQuoteSingle, `"it's \$HOME"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatWithShellQuotes(tc.value, tc.style)
			if result != tc.expected {
				t.Errorf("formatWithShellQuotes(%q, %v) = %q, expected %q",
					tc.value, tc.style, result, tc.expected)
			}
		})
	}
}

func TestTrimCarriageReturn(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no CR", "hello", "hello"},
		{"single CR", "hello\r", "hello"},
		{"multiple CR", "hello\r\r\r", "hello"},
		{"CR in middle", "hel\rlo", "hel\rlo"},
		{"only CR", "\r\r", ""},
		{"empty", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := trimCarriageReturn(tc.input)
			if result != tc.expected {
				t.Errorf("trimCarriageReturn(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestRewriteShellAssignment(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		envMap   map[string]string
		expected string
	}{
		// Basic replacement
		{
			name:     "simple replacement",
			line:     `MY_VAR=""`,
			envMap:   map[string]string{"MY_VAR": "new-value"},
			expected: `MY_VAR="new-value"`,
		},
		{
			name:     "with export",
			line:     `export MY_VAR=""`,
			envMap:   map[string]string{"MY_VAR": "new-value"},
			expected: `export MY_VAR="new-value"`,
		},
		// No replacement (var not in map)
		{
			name:     "var not in map",
			line:     `OTHER_VAR="value"`,
			envMap:   map[string]string{"MY_VAR": "new-value"},
			expected: `OTHER_VAR="value"`,
		},
		// Non-assignment lines
		{
			name:     "not an assignment",
			line:     `echo "hello"`,
			envMap:   map[string]string{"echo": "test"},
			expected: `echo "hello"`,
		},
		{
			name:     "comment line",
			line:     `# MY_VAR=value`,
			envMap:   map[string]string{"MY_VAR": "new-value"},
			expected: `# MY_VAR=value`,
		},
		// Preserve comments
		{
			name:     "preserve inline comment",
			line:     `MY_VAR="" # important comment`,
			envMap:   map[string]string{"MY_VAR": "new-value"},
			expected: `MY_VAR="new-value" # important comment`,
		},
		// Safe unquoted values
		{
			name:     "safe unquoted replacement",
			line:     `MY_VAR=oldvalue`,
			envMap:   map[string]string{"MY_VAR": "simple"},
			expected: `MY_VAR=simple`,
		},
		// Unsafe values get quoted
		{
			name:     "unsafe value gets quoted",
			line:     `MY_VAR=old`,
			envMap:   map[string]string{"MY_VAR": "has space"},
			expected: `MY_VAR="has space"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := rewriteShellAssignment(tc.line, tc.envMap)
			if result != tc.expected {
				t.Errorf("rewriteShellAssignment(%q, %v) = %q, expected %q",
					tc.line, tc.envMap, result, tc.expected)
			}
		})
	}
}

func TestParseHeredocDelimiter(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		expectMatch bool
		expectDelim string
	}{
		// Valid heredocs
		{"basic heredoc", "cat <<EOF", true, "EOF"},
		{"heredoc with dash", "cat <<-EOF", true, "EOF"},
		{"quoted heredoc", "cat <<'EOF'", true, "EOF"},
		{"different delimiter", "cat <<END", true, "END"},
		{"heredoc with underscore", "cat <<MY_DOC", true, "MY_DOC"},
		{"heredoc with numbers", "cat <<DOC123", true, "DOC123"},

		// Not heredocs
		{"not heredoc", "echo hello", false, ""},
		{"less than only", "if [ $a < $b ]", false, ""},
		{"assignment with lt", "VAR=<<", false, ""},
		{"empty", "", false, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			delim, ok := parseHeredocDelimiter(tc.line)
			if ok != tc.expectMatch {
				t.Errorf("parseHeredocDelimiter(%q) match = %v, expected %v", tc.line, ok, tc.expectMatch)
			}
			if ok && delim != tc.expectDelim {
				t.Errorf("parseHeredocDelimiter(%q) delim = %q, expected %q", tc.line, delim, tc.expectDelim)
			}
		})
	}
}

func TestInjectEnvVarsIntoScript_HeredocHandling(t *testing.T) {
	// Test that variables inside heredocs are NOT replaced
	script := []byte(`#!/bin/bash
MY_VAR=""
cat <<EOF
MY_VAR="should not be replaced"
EOF
ANOTHER_VAR=""
`)

	envMap := map[string]string{
		"MY_VAR":      "replaced-value",
		"ANOTHER_VAR": "also-replaced",
	}

	result := string(injectEnvVarsIntoScript(script, envMap))

	// Variables outside heredoc should be replaced
	if !strings.Contains(result, `MY_VAR="replaced-value"`) {
		t.Error("MY_VAR outside heredoc should be replaced")
	}
	if !strings.Contains(result, `ANOTHER_VAR="also-replaced"`) {
		t.Error("ANOTHER_VAR outside heredoc should be replaced")
	}

	// Variable inside heredoc should NOT be replaced
	if !strings.Contains(result, `MY_VAR="should not be replaced"`) {
		t.Error("MY_VAR inside heredoc should NOT be replaced")
	}
}
