/*
Copyright 2022 The Kubernetes Authors.

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

package cherryservers

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// API call responses contain only the minimum information required by the cluster-autoscaler
const listCherryServersResponse = `
[{"id":1000,"name":"server-1000","hostname":"k8s-cluster2-pool3-gndxdmmw","state":"active","tags":{"k8s-cluster":"cluster2","k8s-nodepool":"pool3"}},{"id":1001,"name":"server-1001","hostname":"k8s-cluster2-master","state":"active","tags":{"k8s-cluster":"cluster2"}}]
`

const listCherryServersResponseAfterIncreasePool3 = `
[{"id":2000,"name":"server-2000","hostname":"k8s-cluster2-pool3-xpnrwgdf","state":"active","tags":{"k8s-cluster":"cluster2","k8s-nodepool":"pool3"}},{"id":1000,"name":"server-1000","hostname":"k8s-cluster2-pool3-gndxdmmw","state":"active","tags":{"k8s-cluster":"cluster2","k8s-nodepool":"pool3"}},{"id":1001,"name":"server-1001","hostname":"k8s-cluster2-master","state":"active","tags":{"k8s-cluster":"cluster2"}}]
`

const listCherryServersResponseAfterIncreasePool2 = `
[{"id":3000,"name":"server-3001","hostname":"k8s-cluster2-pool2-jssxcyzz","state":"active","tags":{"k8s-cluster":"cluster2","k8s-nodepool":"pool2"}},{"id":2000,"name":"server-2000","hostname":"k8s-cluster2-pool3-xpnrwgdf","state":"active","tags":{"k8s-cluster":"cluster2","k8s-nodepool":"pool3"}},{"id":1000,"name":"server-1000","hostname":"k8s-cluster2-pool3-gndxdmmw","state":"active","tags":{"k8s-cluster":"cluster2","k8s-nodepool":"pool3"}},{"id":1001,"name":"server-1001","hostname":"k8s-cluster2-master","state":"active","tags":{"k8s-cluster":"cluster2"}}]
`

const cloudinitDefault = "IyEvYmluL2Jhc2gKZXhwb3J0IERFQklBTl9GUk9OVEVORD1ub25pbnRlcmFjdGl2ZQphcHQtZ2V0IHVwZGF0ZSAmJiBhcHQtZ2V0IGluc3RhbGwgLXkgYXB0LXRyYW5zcG9ydC1odHRwcyBjYS1jZXJ0aWZpY2F0ZXMgY3VybCBzb2Z0d2FyZS1wcm9wZXJ0aWVzLWNvbW1vbgpjdXJsIC1mc1NMIGh0dHBzOi8vZG93bmxvYWQuZG9ja2VyLmNvbS9saW51eC91YnVudHUvZ3BnIHwgYXB0LWtleSBhZGQgLQpjdXJsIC1zIGh0dHBzOi8vcGFja2FnZXMuY2xvdWQuZ29vZ2xlLmNvbS9hcHQvZG9jL2FwdC1rZXkuZ3BnIHwgYXB0LWtleSBhZGQgLQpjYXQgPDxFT0YgPi9ldGMvYXB0L3NvdXJjZXMubGlzdC5kL2t1YmVybmV0ZXMubGlzdApkZWIgaHR0cHM6Ly9hcHQua3ViZXJuZXRlcy5pby8ga3ViZXJuZXRlcy14ZW5pYWwgbWFpbgpFT0YKYWRkLWFwdC1yZXBvc2l0b3J5ICAgImRlYiBbYXJjaD1hbWQ2NF0gaHR0cHM6Ly9kb3dubG9hZC5kb2NrZXIuY29tL2xpbnV4L3VidW50dSAgICQobHNiX3JlbGVhc2UgLWNzKSAgIHN0YWJsZSIKYXB0LWdldCB1cGRhdGUKYXB0LWdldCB1cGdyYWRlIC15CmFwdC1nZXQgaW5zdGFsbCAteSBrdWJlbGV0PTEuMTcuNC0wMCBrdWJlYWRtPTEuMTcuNC0wMCBrdWJlY3RsPTEuMTcuNC0wMAphcHQtbWFyayBob2xkIGt1YmVsZXQga3ViZWFkbSBrdWJlY3RsCmN1cmwgLWZzU0wgaHR0cHM6Ly9kb3dubG9hZC5kb2NrZXIuY29tL2xpbnV4L3VidW50dS9ncGcgfCBhcHQta2V5IGFkZCAtCmFkZC1hcHQtcmVwb3NpdG9yeSAiZGViIFthcmNoPWFtZDY0XSBodHRwczovL2Rvd25sb2FkLmRvY2tlci5jb20vbGludXgvdWJ1bnR1IGJpb25pYyBzdGFibGUiCmFwdCB1cGRhdGUKYXB0IGluc3RhbGwgLXkgZG9ja2VyLWNlPTE4LjA2LjJ+Y2V+My0wfnVidW50dQpjYXQgPiAvZXRjL2RvY2tlci9kYWVtb24uanNvbiA8PEVPRgp7CiAgImV4ZWMtb3B0cyI6IFsibmF0aXZlLmNncm91cGRyaXZlcj1zeXN0ZW1kIl0sCiAgImxvZy1kcml2ZXIiOiAianNvbi1maWxlIiwKICAibG9nLW9wdHMiOiB7CiAgICAibWF4LXNpemUiOiAiMTAwbSIKICB9LAogICJzdG9yYWdlLWRyaXZlciI6ICJvdmVybGF5MiIKfQpFT0YKbWtkaXIgLXAgL2V0Yy9zeXN0ZW1kL3N5c3RlbS9kb2NrZXIuc2VydmljZS5kCnN5c3RlbWN0bCBkYWVtb24tcmVsb2FkCnN5c3RlbWN0bCByZXN0YXJ0IGRvY2tlcgpzd2Fwb2ZmIC1hCm12IC9ldGMvZnN0YWIgL2V0Yy9mc3RhYi5vbGQgJiYgZ3JlcCAtdiBzd2FwIC9ldGMvZnN0YWIub2xkID4gL2V0Yy9mc3RhYgpjYXQgPDxFT0YgfCB0ZWUgL2V0Yy9kZWZhdWx0L2t1YmVsZXQKS1VCRUxFVF9FWFRSQV9BUkdTPS0tY2xvdWQtcHJvdmlkZXI9ZXh0ZXJuYWwgLS1ub2RlLWxhYmVscz1wb29sPXt7Lk5vZGVHcm91cH19CkVPRgprdWJlYWRtIGpvaW4gLS1kaXNjb3ZlcnktdG9rZW4tdW5zYWZlLXNraXAtY2EtdmVyaWZpY2F0aW9uIC0tdG9rZW4ge3suQm9vdHN0cmFwVG9rZW5JRH19Lnt7LkJvb3RzdHJhcFRva2VuU2VjcmV0fX0ge3suQVBJU2VydmVyRW5kcG9pbnR9fQo="

var useRealEndpoint bool

func init() {
	useRealEndpoint = strings.TrimSpace(os.Getenv("CHERRY_USE_PRODUCTION_API")) == "true"
}

// newTestCherryManagerRest creates a cherryManagerRest with two nodepools.
// If the url is provided, uses that as the Cherry Servers API endpoint, otherwise
// uses the system default.
func newTestCherryManagerRest(t *testing.T, serverUrl string) *cherryManagerRest {
	poolUrl := baseURL
	if serverUrl != "" {
		poolUrl = serverUrl
	}
	u, err := url.Parse(poolUrl)
	if err != nil {
		t.Fatalf("invalid request path %s: %v", poolUrl, err)
	}
	manager := &cherryManagerRest{
		baseURL: u,
		nodePools: map[string]*cherryManagerNodePool{
			"default": {
				clusterName:       "cluster2",
				projectID:         10001,
				apiServerEndpoint: "147.75.102.15:6443",
				region:            "eu_nord_1",
				plan:              "e5_1620v4",
				os:                "ubuntu_18_04",
				cloudinit:         cloudinitDefault,
				hostnamePattern:   "k8s-{{.ClusterName}}-{{.NodeGroup}}-{{.RandString8}}",
			},
			"pool2": {
				clusterName:       "cluster2",
				projectID:         10001,
				apiServerEndpoint: "147.75.102.15:6443",
				region:            "eu_nord_1",
				plan:              "e5_1620v4",
				os:                "ubuntu_18_04",
				cloudinit:         cloudinitDefault,
				hostnamePattern:   "k8s-{{.ClusterName}}-{{.NodeGroup}}-{{.RandString8}}",
			},
		},
	}
	return manager
}
func TestListCherryServers(t *testing.T) {
	server := NewHttpServerMock(MockFieldContentType, MockFieldResponse)
	defer server.Close()

	var m *cherryManagerRest
	// Set up a mock Cherry Servers API
	if useRealEndpoint {
		// If auth token set in env, hit the actual Cherry Servers API
		m = newTestCherryManagerRest(t, "")
	} else {
		m = newTestCherryManagerRest(t, server.URL)
		t.Logf("server URL: %v", server.URL)
		t.Logf("default cherryManager baseURL: %v", m.baseURL)
		// should get called 2 times: once for listCherryServers() below, and once
		// as part of nodeGroupSize()
		server.On("handle", fmt.Sprintf("/projects/%d/servers", m.nodePools["default"].projectID)).Return("application/json", listCherryServersResponse).Times(2)
	}

	_, err := m.listCherryServers(context.TODO())
	assert.NoError(t, err)

	c, err := m.nodeGroupSize("pool3")
	assert.NoError(t, err)
	assert.Equal(t, int(1), c) // One server in nodepool

	mock.AssertExpectationsForObjects(t, server)
}
