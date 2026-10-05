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

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"pkg/utils/metrics/metrics.go": `package metrics
const TopMetricsNamespace = "vpa_"
func CreateExecutionTimeMetric(namespace string, help string) *prometheus.HistogramVec {
 return prometheus.NewHistogramVec(prometheus.HistogramOpts{Namespace: namespace, Name: "execution_latency_seconds", Help: help}, []string{"step"})
}`,
		"pkg/utils/metrics/recommender/recommender.go": `package recommender
const metricsNamespace = metrics.TopMetricsNamespace + "recommender"
var count = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: metricsNamespace, Name: "objects", Help: "Object | count"}, []string{"mode"})
var latency = metrics.CreateExecutionTimeMetric(metricsNamespace, "Loop duration")
func Register() { prometheus.MustRegister(count, latency) }`,
		"docs/recommender-metrics.md": "Introduction\n" + startMarker + "\nold table\n" + endMarker + "\nNotes\n",
	}
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func change(t *testing.T, root, name, old, replacement string) {
	t.Helper()
	path := filepath.Join(root, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), old) {
		t.Fatalf("fixture does not contain %q", old)
	}
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(string(data), old, replacement)), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateAndCheck(t *testing.T) {
	root := fixture(t)
	path := filepath.Join(root, "docs/recommender-metrics.md")
	before, _ := os.ReadFile(path)
	if err := run(root, true); err == nil {
		t.Fatal("stale documentation passed verification")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("verification modified the document")
	}
	if err := run(root, false); err != nil {
		t.Fatal(err)
	}
	if err := run(root, true); err != nil {
		t.Fatal(err)
	}
	doc, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Introduction\n", "\nNotes\n",
		"| `vpa_recommender_objects` | Gauge | `mode` | Object \\| count |",
		"| `vpa_recommender_execution_latency_seconds` | Histogram | `step` | Loop duration |",
	} {
		if !strings.Contains(string(doc), want) {
			t.Errorf("missing %q in generated document", want)
		}
	}
	if err := run(root, false); err != nil {
		t.Fatal(err)
	}
	again, _ := os.ReadFile(path)
	if string(doc) != string(again) {
		t.Fatal("generation is not deterministic")
	}
}

func TestDefinitionChanges(t *testing.T) {
	cases := []struct{ name, file, old, replacement, want string }{
		{"namespace", "pkg/utils/metrics/metrics.go", `"vpa_"`, `"custom_"`, "`custom_recommender_objects`"},
		{"helper labels", "pkg/utils/metrics/metrics.go", `[]string{"step"}`, `[]string{"phase", "result"}`, "`phase`, `result`"},
		{"helper metric name", "pkg/utils/metrics/metrics.go", `"execution_latency_seconds"`, `"loop_seconds"`, "`vpa_recommender_loop_seconds`"},
		{"description", "pkg/utils/metrics/recommender/recommender.go", `"Object | count"`, `"Current objects"`, "Current objects"},
		{"added collector", "pkg/utils/metrics/recommender/recommender.go", "func Register() { prometheus.MustRegister(count, latency) }", `var requests = prometheus.NewCounter(prometheus.CounterOpts{Namespace: metricsNamespace, Name: "requests", Help: "Requests"})
func Register() { prometheus.MustRegister(count, latency, requests) }`, "| `vpa_recommender_requests` | Counter | None | Requests |"},
		{"removed collector", "pkg/utils/metrics/recommender/recommender.go", "MustRegister(count, latency)", "MustRegister(count)", "Object"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := fixture(t)
			if err := run(root, false); err != nil {
				t.Fatal(err)
			}
			change(t, root, tc.file, tc.old, tc.replacement)
			if err := run(root, true); err == nil {
				t.Fatal("changed definition passed verification")
			}
			if err := run(root, false); err != nil {
				t.Fatal(err)
			}
			if err := run(root, true); err != nil {
				t.Fatal(err)
			}
			doc, _ := os.ReadFile(filepath.Join(root, "docs/recommender-metrics.md"))
			if !strings.Contains(string(doc), tc.want) {
				t.Fatalf("missing %q", tc.want)
			}
			if tc.name == "removed collector" && strings.Contains(string(doc), "execution_latency_seconds") {
				t.Fatal("removed collector is still documented")
			}
		})
	}
}

func TestUnsupportedDefinitionFailsWithoutWriting(t *testing.T) {
	root := fixture(t)
	change(t, root, "pkg/utils/metrics/recommender/recommender.go", `[]string{"mode"}`, `labelNames()`)
	path := filepath.Join(root, "docs/recommender-metrics.md")
	before, _ := os.ReadFile(path)
	if err := run(root, false); err == nil || !strings.Contains(err.Error(), "literal label names") {
		t.Fatalf("expected unsupported labels error, got %v", err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("failed generation modified the document")
	}
}

func TestInvalidMarkers(t *testing.T) {
	for _, doc := range []string{"no markers", endMarker + startMarker, startMarker + startMarker + endMarker} {
		if _, err := replaceTable(doc, "table"); err == nil {
			t.Errorf("accepted invalid markers in %q", doc)
		}
	}
}
