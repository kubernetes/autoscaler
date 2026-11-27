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

package clusterapi

import (
	"context"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/rest"
	clientgotesting "k8s.io/client-go/testing"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	testprovider "sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider/test"
	ca_processors "sigs.k8s.io/cluster-autoscaler/pkg/processors"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/nodes"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/scaledowncandidates"
)

const unreachableHost = "http://127.0.0.1:1"

func assertStopped(t *testing.T, p *provider) {
	t.Helper()
	select {
	case <-p.controller.stopChannel:
	case <-time.After(wait.ForeverTestTimeout):
		t.Fatal("expected the controller stop channel to be closed")
	}
}

func assertRunning(t *testing.T, p *provider) {
	t.Helper()
	select {
	case <-p.controller.stopChannel:
		t.Fatal("expected the controller stop channel to be open")
	default:
	}
}

func TestNewProviderValidatesOptions(t *testing.T) {
	config := &rest.Config{Host: unreachableHost}

	for _, tc := range []struct {
		name string
		opts ProviderOptions
	}{
		{
			name: "nil management config",
			opts: ProviderOptions{WorkloadConfig: config},
		},
		{
			name: "nil workload config",
			opts: ProviderOptions{ManagementConfig: config},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewProvider(context.Background(), tc.opts)
			if err == nil {
				t.Fatal("expected an error")
			}
			if p != nil {
				t.Fatalf("expected a nil provider, got %T", p)
			}
		})
	}
}

func TestNewProviderDoesNotMutateConfig(t *testing.T) {
	management := &rest.Config{Host: unreachableHost, QPS: 7, Burst: 11}
	workload := &rest.Config{Host: unreachableHost, QPS: 13, Burst: 17}
	wantManagement := rest.CopyConfig(management)
	wantWorkload := rest.CopyConfig(workload)

	ctx, cancel := context.WithTimeout(context.Background(), wait.ForeverTestTimeout)
	defer cancel()

	if _, err := NewProvider(ctx, ProviderOptions{
		ManagementConfig: management,
		WorkloadConfig:   workload,
	}); err == nil {
		t.Fatal("expected an error for an unreachable management cluster")
	}

	if !reflect.DeepEqual(management, wantManagement) {
		t.Errorf("management config was mutated: got %+v, want %+v", management, wantManagement)
	}
	if !reflect.DeepEqual(workload, wantWorkload) {
		t.Errorf("workload config was mutated: got %+v, want %+v", workload, wantWorkload)
	}
}

func TestNewProviderFromClientsStopsOnContextCancel(t *testing.T) {
	clients, _ := newTestProviderClients(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p, err := newProviderFromClients(ctx, clients, ProviderOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRunning(t, p)
	cancel()
	assertStopped(t, p)
}

func TestCleanupStopsProvider(t *testing.T) {
	clients, _ := newTestProviderClients(t)

	p, err := newProviderFromClients(context.Background(), clients, ProviderOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRunning(t, p)
	if err := p.Cleanup(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertStopped(t, p)

	// A second Cleanup must be a no-op.
	if err := p.Cleanup(context.Background()); err != nil {
		t.Fatalf("unexpected error on second cleanup: %v", err)
	}
}

func TestCleanupAfterContextCancel(t *testing.T) {
	clients, _ := newTestProviderClients(t)
	ctx, cancel := context.WithCancel(context.Background())

	p, err := newProviderFromClients(ctx, clients, ProviderOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cancel()
	assertStopped(t, p)
	if err := p.Cleanup(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCleanupWithoutStop(t *testing.T) {
	p := newProvider(ProviderName, nil, nil)
	if err := p.Cleanup(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewProviderFromClientsReturnsError(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*providerClients)
		options ProviderOptions
	}{
		{
			name: "missing Cluster API group",
			mutate: func(c *providerClients) {
				c.discovery = &fakediscovery.FakeDiscovery{Fake: &clientgotesting.Fake{}}
			},
		},
		{
			name: "invalid auto discovery spec",
			options: ProviderOptions{
				DiscoveryOptions: cloudprovider.NodeGroupDiscoveryOptions{
					NodeGroupAutoDiscoverySpecs: []string{"invalid"},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clients, _ := newTestProviderClients(t)
			if tc.mutate != nil {
				tc.mutate(&clients)
			}

			p, err := newProviderFromClients(context.Background(), clients, tc.options)
			if err == nil {
				t.Fatal("expected an error")
			}
			if p != nil {
				t.Fatal("expected a nil provider")
			}
		})
	}
}

func TestRegisterProcessors(t *testing.T) {
	clients, _ := newTestProviderClients(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p, err := newProviderFromClients(ctx, clients, ProviderOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, tc := range []struct {
		name       string
		provider   cloudprovider.CloudProvider
		processors *ca_processors.AutoscalingProcessors
		wantErr    bool
	}{
		{
			name:     "combined scale down processor",
			provider: p,
			processors: &ca_processors.AutoscalingProcessors{
				ScaleDownNodeProcessor: scaledowncandidates.NewCombinedScaleDownCandidatesProcessor(),
			},
		},
		{
			name:     "wrong scale down processor type",
			provider: p,
			processors: &ca_processors.AutoscalingProcessors{
				ScaleDownNodeProcessor: nodes.NewPreFilteringScaleDownNodeProcessor(),
			},
			wantErr: true,
		},
		{
			name:     "non Cluster API cloud provider",
			provider: testprovider.NewTestCloudProviderBuilder().Build(),
			processors: &ca_processors.AutoscalingProcessors{
				ScaleDownNodeProcessor: scaledowncandidates.NewCombinedScaleDownCandidatesProcessor(),
			},
			wantErr: true,
		},
		{
			name:     "nil processors",
			provider: p,
			wantErr:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := RegisterProcessors(tc.provider, tc.processors)
			if tc.wantErr && err == nil {
				t.Fatal("expected an error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
