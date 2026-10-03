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

package pressure

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

type fakeKubelets struct {
	mu      sync.Mutex
	delay   map[string]time.Duration
	mode    map[string]string // "", "429", "oversized"
	fetches map[string]int
}

// newFakeKubelets serves /stats/summary for n nodes named node-<i>, identified by the request Host.
func newFakeKubelets(t *testing.T, n int, cfg Config) (*KubeletObserver, *fakeKubelets) {
	f := &fakeKubelets{delay: map[string]time.Duration{}, mode: map[string]string{}, fetches: map[string]int{}}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		node, _, _ := strings.Cut(r.Host, ":")
		f.mu.Lock()
		f.fetches[node]++
		d, mode := f.delay[node], f.mode[node]
		f.mu.Unlock()
		select {
		case <-time.After(d):
		case <-r.Context().Done():
			return
		}
		switch mode {
		case "429":
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
		case "oversized":
			_, _ = w.Write([]byte(`{"pods":[],"pad":"` + strings.Repeat("x", 2048) + `"}`))
		default:
			_, _ = w.Write([]byte(`{"pods":[]}`))
		}
	}))
	t.Cleanup(srv.Close)
	tr := srv.Client().Transport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // test server cert; node identity is the Host header
	tr.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, srv.Listener.Addr().String())
	}
	o := NewObserver(nil, cfg)
	o.httpClient = &http.Client{Transport: tr}
	o.nodeAddr = func(node string) (string, error) { return node + ":443", nil }
	var targets []Target
	for i := range n {
		node := fmt.Sprintf("node-%03d", i)
		targets = append(targets, Target{ContainerID: model.ContainerID{PodID: model.PodID{Namespace: "ns", PodName: node}, ContainerName: "c"},
			PodUID: node, Node: node, LimitBytes: 512 << 20})
	}
	o.SetTargets(targets)
	return o, f
}

func TestSweepFetchesEveryNode(t *testing.T) {
	cfg := DefaultConfig()
	o, f := newFakeKubelets(t, 200, cfg)
	o.sweep(context.Background())
	assert.Len(t, f.fetches, 200)
	assert.Len(t, o.lastOK, 200)
	assert.Empty(t, o.carry)
	assert.Empty(t, o.backoff)
}

func TestSweepDeadlineCarriesNodesAndCoversAll(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SweepTimeout = 300 * time.Millisecond
	cfg.MaxConcurrent = 2
	o, f := newFakeKubelets(t, 60, cfg)
	for i := range 60 {
		f.delay[fmt.Sprintf("node-%03d", i)] = 40 * time.Millisecond
	}
	o.sweep(context.Background())
	first := len(o.lastOK)
	assert.Less(t, first, 60, "the deadline must cut the first sweep short")
	assert.NotEmpty(t, o.carry, "requests cut by the deadline are carried")
	assert.Empty(t, o.backoff, "deadline cancellations must not back off")
	for range 20 {
		o.sweep(context.Background())
		if len(o.lastOK) == 60 {
			break
		}
	}
	assert.Len(t, o.lastOK, 60, "rotation plus carry reaches every node")
	assert.Empty(t, o.backoff)
}

func TestSweepBacksOffThrottledAndOversizedNodes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxResponseBytes = 1024
	o, f := newFakeKubelets(t, 10, cfg)
	f.mode["node-003"] = "429"
	f.mode["node-007"] = "oversized"
	o.sweep(context.Background())
	assert.Len(t, o.lastOK, 8)
	b := o.backoff["node-003"]
	assert.GreaterOrEqual(t, b.delay, 4*time.Second, "min(5s) backoff with jitter; Retry-After 2s does not lower it")
	assert.Contains(t, o.backoff, "node-007")
	o.sweep(context.Background())
	assert.Equal(t, 1, f.fetches["node-003"], "a node in backoff is skipped")
	assert.Equal(t, 2, f.fetches["node-000"])
}

func TestSweepForgetsIneligibleNodes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxResponseBytes = 1024
	o, f := newFakeKubelets(t, 3, cfg)
	f.mode["node-002"] = "429"
	o.sweep(context.Background())
	assert.Contains(t, o.lastOK, "node-000")
	assert.Contains(t, o.backoff, "node-002")

	o.SetTargets(nil)
	o.sweep(context.Background())
	assert.Empty(t, o.lastOK, "nodes that left the target set are forgotten")
	assert.Empty(t, o.backoff)
	assert.Empty(t, o.nodeFlag)
}

func TestFetchSummarySizeLimit(t *testing.T) {
	testCases := []struct {
		name    string
		mode    string
		wantErr string
	}{
		{name: "response within the limit", mode: ""},
		{name: "response over the limit", mode: "oversized", wantErr: "exceeds 1024 bytes"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.MaxResponseBytes = 1024
			o, f := newFakeKubelets(t, 1, cfg)
			f.mode["node-000"] = tc.mode
			_, err := o.fetchSummary(context.Background(), "node-000")
			if tc.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func TestKubeletRESTConfig(t *testing.T) {
	testCases := []struct {
		name       string
		caFile     string
		wantCAFile string
		wantCAData []byte
	}{
		{name: "keeps the API server CA by default", caFile: "", wantCAFile: "/var/run/apiserver-ca.crt", wantCAData: []byte("apiserver-ca")},
		{name: "replaces the CA when a file is given", caFile: "/etc/kubelet-ca.crt", wantCAFile: "/etc/kubelet-ca.crt"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			restCfg := &rest.Config{TLSClientConfig: rest.TLSClientConfig{CAFile: "/var/run/apiserver-ca.crt", CAData: []byte("apiserver-ca"), Insecure: true}}
			got := kubeletRESTConfig(restCfg, tc.caFile)
			assert.Equal(t, tc.wantCAFile, got.CAFile)
			assert.Equal(t, tc.wantCAData, got.CAData)
			assert.False(t, got.Insecure, "insecure mode is never allowed")
			assert.True(t, restCfg.Insecure, "the input config is not modified")
		})
	}
}
