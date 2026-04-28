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
	"bytes"
	"context"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"

	"github.com/google/uuid"
	"k8s.io/klog/v2"
)

// Config holds configuration for the Manager.
type Config struct {
	// Directory is the path to store diagnostic reports. Empty string disables the diagnostic system.
	Directory string
	// MaxLoopTime is the threshold for a single loop duration to trigger diagnostics.
	MaxLoopTime time.Duration
	// MaxCollectionTime is the duration to collect CPU and Trace profiles.
	MaxCollectionTime time.Duration
	// CooldownPeriod is the minimum time between two diagnostic collections.
	CooldownPeriod time.Duration
	// EnabledProfiles is a list of profiles to collect (e.g., "cpu", "trace", "heap", "allocs").
	EnabledProfiles []string

	// TraceMaxMB is the maximum size of the flight recorder window in megabytes.
	TraceMaxMB uint64
}

// Manager monitors loop durations and triggers diagnostic reports.
type Manager struct {
	sink       DiagnosticSink
	config     Config
	collectors map[string]ProfileCollector

	mu           sync.Mutex
	lastTrigger  time.Time
	isCollecting bool
}

// NewManager creates a new Manager.
func NewManager(sink DiagnosticSink, config Config) *Manager {
	collectors := make(map[string]ProfileCollector)

	for _, p := range config.EnabledProfiles {
		switch p {
		case "cpu":
			collectors["cpu"] = &cpuCollector{}
		case "trace":
			if tc, err := newTraceCollector(config); err == nil {
				collectors["trace"] = tc
			}
		case "heap":
			collectors["heap"] = &heapCollector{}
		case "allocs":
			collectors["allocs"] = &allocsCollector{}
		}
	}

	return NewManagerWithCollectors(sink, config, collectors)
}

// NewManagerWithCollectors creates a new Manager with custom collectors.
func NewManagerWithCollectors(sink DiagnosticSink, config Config, collectors map[string]ProfileCollector) *Manager {
	dm := &Manager{
		sink:       sink,
		config:     config,
		collectors: collectors,
	}

	dm.notifyCooldownEnd()

	return dm
}

func (dm *Manager) notifyCooldownStart() {
	for _, c := range dm.collectors {
		if obs, ok := c.(CooldownObserver); ok {
			obs.OnCooldownStart()
		}
	}
}

func (dm *Manager) notifyCooldownEnd() {
	for _, c := range dm.collectors {
		if obs, ok := c.(CooldownObserver); ok {
			obs.OnCooldownEnd()
		}
	}
}

type cpuCollector struct{}

func (c *cpuCollector) Collect(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := pprof.StartCPUProfile(&buf); err != nil {
		return nil, err
	}
	<-ctx.Done()
	pprof.StopCPUProfile()
	return buf.Bytes(), nil
}

type traceCollector struct {
	flightRecorder *trace.FlightRecorder
	mu             sync.Mutex
	running        bool
}

func newTraceCollector(config Config) (*traceCollector, error) {
	// Calculate minimum age to ensure we hold at least (MaxLoopTime + MaxCollectionTime).
	// This guarantees that when WriteTo is called at the end of collection,
	// the resulting trace contains the full history of the event.
	minAge := config.MaxLoopTime + config.MaxCollectionTime
	// Add a 10% safety buffer
	minAge += minAge / 10

	frConfig := trace.FlightRecorderConfig{
		MaxBytes: config.TraceMaxMB * 1024 * 1024,
		MinAge:   minAge,
	}
	fr := trace.NewFlightRecorder(frConfig)
	return &traceCollector{flightRecorder: fr}, nil
}

func (c *traceCollector) OnCooldownEnd() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.running {
		return
	}
	if err := c.flightRecorder.Start(); err != nil {
		klog.ErrorS(err, "Failed to start flight recorder")
		return
	}
	c.running = true
}

func (c *traceCollector) OnCooldownStart() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.running {
		return
	}
	c.flightRecorder.Stop()
	c.running = false
}

func (c *traceCollector) Collect(ctx context.Context) ([]byte, error) {
	<-ctx.Done()
	var buf bytes.Buffer
	if _, err := c.flightRecorder.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type heapCollector struct{}

func (c *heapCollector) Collect(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := pprof.Lookup("heap").WriteTo(&buf, 0); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type allocsCollector struct{}

func (c *allocsCollector) Collect(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer
	if err := pprof.Lookup("allocs").WriteTo(&buf, 0); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RunWithDiagnostic executes the provided function. If it takes longer than MaxLoopTime,
// diagnostic collection is triggered automatically in the background.
func (dm *Manager) RunWithDiagnostic(ctx context.Context, runFn func()) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		select {
		case <-time.After(dm.config.MaxLoopTime):
			dm.mu.Lock()
			now := time.Now()
			if dm.isCollecting || now.Sub(dm.lastTrigger) < dm.config.CooldownPeriod {
				dm.mu.Unlock()
				return
			}
			dm.isCollecting = true
			dm.lastTrigger = now
			dm.mu.Unlock()

			klog.InfoS("Diagnostic triggered diagnostic collection", "threshold", dm.config.MaxLoopTime)
			go dm.collectAndStore(ctx)
		case <-ctx.Done():
		}
	}()

	runFn()
}

func (dm *Manager) collectAndStore(ctx context.Context) {
	defer func() {
		dm.mu.Lock()
		dm.isCollecting = false
		dm.notifyCooldownStart()

		// Schedule the end of the cooldown period.
		// Using the lastTrigger to ensure we respect the full period from the start of the event.
		nextEnable := time.Until(dm.lastTrigger.Add(dm.config.CooldownPeriod))
		time.AfterFunc(nextEnable, func() {
			dm.mu.Lock()
			defer dm.mu.Unlock()
			dm.notifyCooldownEnd()
		})
		dm.mu.Unlock()
	}()

	report := &DiagnosticReport{
		ID:       uuid.New().String(),
		Profiles: make(map[string][]byte),
		Metadata: make(map[string]string),
	}

	report.Metadata["threshold_exceeded"] = dm.config.MaxLoopTime.String()
	report.Metadata["trigger_time"] = time.Now().Format(time.RFC3339)

	// collectionCtx will be cancelled when either MaxCollectionTime expires
	// or the monitored function completes (via ctx).
	collectionCtx, collectionCancel := context.WithTimeout(ctx, dm.config.MaxCollectionTime)
	defer collectionCancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	for name, collector := range dm.collectors {
		wg.Add(1)
		go func(n string, c ProfileCollector) {
			defer wg.Done()
			data, err := c.Collect(collectionCtx)
			if err != nil {
				klog.ErrorS(err, "Failed to collect profile", "profile", n)
				return
			}
			mu.Lock()
			report.Profiles[n] = data
			mu.Unlock()
		}(name, collector)
	}
	wg.Wait()

	// Use context.Background() for storage to ensure it persists even if the monitored loop has finished.
	if err := dm.sink.Store(context.Background(), report); err != nil {
		klog.ErrorS(err, "Failed to store diagnostic report", "reportID", report.ID)
	} else {
		klog.InfoS("Diagnostic report stored successfully", "reportID", report.ID)
	}
}
