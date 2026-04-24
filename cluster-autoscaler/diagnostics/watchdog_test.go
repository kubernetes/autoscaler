/*
Copyright The Kubernetes Authors.

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

package diagnostics

import (
	"context"
	"os"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockSink struct {
	mu      sync.Mutex
	reports []*DiagnosticReport
}

func (m *mockSink) Store(ctx context.Context, report *DiagnosticReport) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reports = append(m.reports, report)
	return nil
}

type mockCollector struct {
	profileType string
	mu          sync.Mutex
	startCalls  int
	stopCalls   int
}

func (m *mockCollector) OnCooldownEnd() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startCalls++
}

func (m *mockCollector) OnCooldownStart() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopCalls++
}

func (m *mockCollector) Collect(ctx context.Context) ([]byte, error) {
	return []byte("test-" + m.profileType + "-profile"), nil
}

func TestManagerLifecycle(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &mockSink{}
		config := Config{
			MaxLoopTime:       10 * time.Second,
			MaxCollectionTime: 5 * time.Second,
			CooldownPeriod:    60 * time.Second,
		}

		collector := &mockCollector{profileType: "test"}
		collectors := map[string]ProfileCollector{"test": collector}
		dm := NewManagerWithCollectors(sink, config, collectors)

		// 1. Verify initial start
		assert.Equal(t, 1, collector.startCalls, "Should call OnCooldownEnd on creation")

		// 2. Trigger collection
		ctx := context.Background()
		dm.RunWithDiagnostic(ctx, func() {
			time.Sleep(20 * time.Second)
		})
		synctest.Wait()

		// Verify stop was called after collection
		assert.Equal(t, 1, collector.stopCalls, "Should call OnCooldownStart after collection")

		// 3. Verify automatic resume after cooldown
		// Advance time past the 60s cooldown
		time.Sleep(65 * time.Second)
		synctest.Wait()

		assert.Equal(t, 2, collector.startCalls, "Should call OnCooldownEnd after cooldown expires via timer")
	})
}

func TestDiagnosticManager(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sink := &mockSink{}
		config := Config{
			MaxLoopTime:       10 * time.Second,
			MaxCollectionTime: 5 * time.Second,
			CooldownPeriod:    60 * time.Second,
			TraceMaxMB:        64,
		}

		collectors := map[string]ProfileCollector{
			"cpu":  &mockCollector{profileType: "cpu"},
			"heap": &mockCollector{profileType: "heap"},
		}
		dm := NewManagerWithCollectors(sink, config, collectors)
		ctx := context.Background()

		// Test 1: Loop duration below threshold
		t.Log("Testing below threshold...")
		dm.RunWithDiagnostic(ctx, func() {
			time.Sleep(5 * time.Second)
		})
		synctest.Wait()

		sink.mu.Lock()
		assert.Equal(t, 0, len(sink.reports), "Should not collect if below threshold")
		sink.mu.Unlock()

		// Test 2: Loop duration above threshold
		t.Log("Testing above threshold...")
		dm.RunWithDiagnostic(ctx, func() {
			time.Sleep(20 * time.Second)
		})
		// Wait for the background goroutines to finish collection
		synctest.Wait()

		sink.mu.Lock()
		assert.Equal(t, 1, len(sink.reports), "Should collect if above threshold")
		if len(sink.reports) == 1 {
			report := sink.reports[0]
			assert.Contains(t, report.Profiles, "cpu")
			assert.Contains(t, report.Profiles, "heap")
		}
		sink.mu.Unlock()

		// Test 3: Cooldown period
		t.Log("Testing cooldown...")
		// Trigger another one quickly - should be blocked by cooldown
		dm.RunWithDiagnostic(ctx, func() {
			time.Sleep(20 * time.Second)
		})
		synctest.Wait()

		sink.mu.Lock()
		assert.Equal(t, 1, len(sink.reports), "Should not collect during cooldown")
		sink.mu.Unlock()

		// Advance time past cooldown
		time.Sleep(65 * time.Second)
		dm.RunWithDiagnostic(ctx, func() {
			time.Sleep(20 * time.Second)
		})
		synctest.Wait()

		sink.mu.Lock()
		assert.Equal(t, 2, len(sink.reports), "Should collect after cooldown")
		sink.mu.Unlock()

	})
}

func TestFileSink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "diagnostics-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	sink, err := NewFileSink(tmpDir)
	assert.NoError(t, err)

	report := &DiagnosticReport{
		ID: "test-report",
		Profiles: map[string][]byte{
			"cpu":  []byte("cpu-data"),
			"heap": []byte("heap-data"),
		},
	}

	err = sink.Store(context.Background(), report)
	assert.NoError(t, err)

	_, err = os.Stat(tmpDir + "/test-report.cpu")
	assert.NoError(t, err)
	_, err = os.Stat(tmpDir + "/test-report.heap")
	assert.NoError(t, err)

	data, err := os.ReadFile(tmpDir + "/test-report.cpu")
	assert.NoError(t, err)
	assert.Equal(t, []byte("cpu-data"), data)
}
