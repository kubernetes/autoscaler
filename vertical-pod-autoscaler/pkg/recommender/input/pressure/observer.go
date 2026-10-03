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

// Package pressure polls the kubelet Summary API for per-container memory PSI and reports sustained
// memory stall on containers whose working set is near their memory limit.
package pressure

import (
	"context"
	"net/http"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/utils/clock"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// Transport selects how the observer reaches the kubelet Summary API.
type Transport string

const (
	// TransportDirect fetches /stats/summary from each kubelet endpoint.
	TransportDirect Transport = "direct"
	// TransportAPIServerProxy fetches through the API server node proxy.
	TransportAPIServerProxy Transport = "apiserver-proxy"
)

// Config holds the detector, scheduling and transport settings.
type Config struct {
	// FetchInterval is how often each eligible node is polled.
	FetchInterval time.Duration
	// StallRatio is the memory PSI "some" stall rate a window must exceed.
	StallRatio float64
	// WorkingSetFraction is the working-set/limit ratio a window must reach.
	WorkingSetFraction float64
	// Cooldown suppresses further events for a container after one is emitted.
	Cooldown time.Duration
	// StartupGrace suppresses events this long after a container starts.
	StartupGrace time.Duration
	// PSITotalUnit is the duration of one unit of the Summary PSI total counter.
	PSITotalUnit time.Duration
	// MaxConcurrent caps in-flight fetches.
	MaxConcurrent int
	// RequestTimeout bounds one fetch.
	RequestTimeout time.Duration
	// SweepTimeout bounds one sweep over all eligible nodes.
	SweepTimeout time.Duration
	// MaxBackoff caps the delay before retrying a failing node.
	MaxBackoff time.Duration
	// MaxResponseBytes caps the size of one Summary response.
	MaxResponseBytes int64
	// MaxSampleAge discards samples older than this.
	MaxSampleAge time.Duration
	// Transport selects how the kubelet Summary API is reached.
	Transport Transport
}

// DefaultConfig returns the default settings.
func DefaultConfig() Config {
	return Config{
		FetchInterval:      60 * time.Second,
		StallRatio:         0.10,
		WorkingSetFraction: 0.90,
		Cooldown:           10 * time.Minute,
		StartupGrace:       60 * time.Second,
		PSITotalUnit:       time.Microsecond,
		MaxConcurrent:      50,
		RequestTimeout:     5 * time.Second,
		SweepTimeout:       45 * time.Second,
		MaxBackoff:         5 * time.Minute,
		MaxResponseBytes:   16 << 20,
		MaxSampleAge:       2 * time.Minute,
		Transport:          TransportDirect,
	}
}

// queueSize bounds pending observations; a full queue drops and counts the observation.
const queueSize = 5000

// Target is one eligible container.
type Target struct {
	ContainerID model.ContainerID
	PodUID      string
	Node        string
	LimitBytes  int64
	// ResizeState is resize_in_progress, resize_deferred or resize_infeasible while a resize is
	// outstanding, and empty otherwise.
	ResizeState string
}

// Info is a pressure observation for one container.
type Info struct {
	ContainerID model.ContainerID
	PodUID      string
	SampleTime  time.Time
	WorkingSet  model.ResourceAmount
	StallRate   float64
}

// KubeletObserver polls each eligible node's kubelet Summary API and emits an Info when sustained
// memory stall coincides with a working set near the limit.
type KubeletObserver struct {
	client rest.Interface
	cfg    Config
	ch     chan Info
	clock  clock.PassiveClock

	// Direct transport.
	httpClient *http.Client
	nodeAddr   func(node string) (string, error)

	mu      sync.Mutex // guards targets and state
	targets map[targetKey]Target
	state   map[targetKey]*containerState

	schedMu   sync.Mutex // guards the scheduling fields below
	cursor    int
	carry     []string        // cancelled by the last sweep's deadline; fetched first next sweep
	done      map[string]bool // completed in the current sweep
	latencies []time.Duration
	lastOK    map[string]time.Time
	nodeFlag  map[string]string
	backoff   map[string]backoffState
}

type backoffState struct {
	until time.Time
	delay time.Duration
}

// NewObserver creates an observer. client is a CoreV1 REST client, used by the apiserver-proxy transport.
func NewObserver(client rest.Interface, cfg Config) *KubeletObserver {
	return &KubeletObserver{
		client:   client,
		cfg:      cfg,
		ch:       make(chan Info, queueSize),
		clock:    clock.RealClock{},
		targets:  map[targetKey]Target{},
		state:    map[targetKey]*containerState{},
		lastOK:   map[string]time.Time{},
		nodeFlag: map[string]string{},
		backoff:  map[string]backoffState{},
	}
}

// Channel returns emitted observations.
func (o *KubeletObserver) Channel() <-chan Info { return o.ch }

// SetTargets replaces the set of containers to observe and drops state for containers that left it.
func (o *KubeletObserver) SetTargets(targets []Target) {
	m := make(map[targetKey]Target, len(targets))
	for _, t := range targets {
		m[targetKey{podUID: t.PodUID, container: t.ContainerID.ContainerName}] = t
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.targets = m
	for k := range o.state {
		if _, ok := m[k]; !ok {
			delete(o.state, k)
		}
	}
}

// Eligible reports whether the observed container is still in the current target set.
func (o *KubeletObserver) Eligible(info Info) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	_, ok := o.targets[targetKey{podUID: info.PodUID, container: info.ContainerID.ContainerName}]
	return ok
}

// Run polls until ctx is done. Sweeps start every FetchInterval regardless of how long the previous one
// took, because a node revisited less often than every two intervals can never qualify.
func (o *KubeletObserver) Run(ctx context.Context) {
	wait.NonSlidingUntilWithContext(ctx, o.sweep, o.cfg.FetchInterval)
}

// targetKey identifies a container across Summary responses.
type targetKey struct {
	podUID    string
	container string
}
