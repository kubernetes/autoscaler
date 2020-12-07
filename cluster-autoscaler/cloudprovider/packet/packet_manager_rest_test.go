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

package packet

import (
	"context"
	"os"
	"testing"

	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// API call responses contain only the minimum information required by the cluster-autoscaler
const listPacketDevicesResponse = `
{"devices":[{"id":"cace3b27-dff8-4930-943d-b2a63a775f03","short_id":"cace3b27","hostname":"k8s-cluster2-pool3-gndxdmmw","description":null,"state":"active","tags":["k8s-cluster-cluster2","k8s-nodepool-pool3"]},{"id":"efc985f6-ba6a-4bc3-8ef4-9643b0e950a9","short_id":"efc985f6","hostname":"k8s-cluster2-master","description":null,"state":"active","tags":["k8s-cluster-cluster2"]}]}
`

const listPacketDevicesResponseAfterIncreasePool3 = `
{"devices":[{"id":"8fa90049-e715-4794-ba31-81c1c78cee84","short_id":"8fa90049","hostname":"k8s-cluster2-pool3-xpnrwgdf","description":null,"state":"active","tags":["k8s-cluster-cluster2","k8s-nodepool-pool3"]},{"id":"cace3b27-dff8-4930-943d-b2a63a775f03","short_id":"cace3b27","hostname":"k8s-cluster2-pool3-gndxdmmw","description":null,"state":"active","tags":["k8s-cluster-cluster2","k8s-nodepool-pool3"]},{"id":"efc985f6-ba6a-4bc3-8ef4-9643b0e950a9","short_id":"efc985f6","hostname":"k8s-cluster2-master","description":null,"state":"active","tags":["k8s-cluster-cluster2"]}]}
`

const listPacketDevicesResponseAfterIncreasePool2 = `
{"devices":[{"id":"0f5609af-1c27-451b-8edd-a1283f2c9440","short_id":"0f5609af","hostname":"k8s-cluster2-pool2-jssxcyzz","description":null,"state":"active","tags":["k8s-cluster-cluster2","k8s-nodepool-pool2"]},{"id":"8fa90049-e715-4794-ba31-81c1c78cee84","short_id":"8fa90049","hostname":"k8s-cluster2-pool3-xpnrwgdf","description":null,"state":"active","tags":["k8s-cluster-cluster2","k8s-nodepool-pool3"]},{"id":"cace3b27-dff8-4930-943d-b2a63a775f03","short_id":"cace3b27","hostname":"k8s-cluster2-pool3-gndxdmmw","description":null,"state":"active","tags":["k8s-cluster-cluster2","k8s-nodepool-pool3"]},{"id":"efc985f6-ba6a-4bc3-8ef4-9643b0e950a9","short_id":"efc985f6","hostname":"k8s-cluster2-master","description":null,"state":"active","tags":["k8s-cluster-cluster2"]}]}
`

const cloudinitDefault = "IyEvYmluL2Jhc2gKZXhwb3J0IERFQklBTl9GUk9OVEVORD1ub25pbnRlcmFjdGl2ZQphcHQtZ2V0IHVwZGF0ZSAmJiBhcHQtZ2V0IGluc3RhbGwgLXkgYXB0LXRyYW5zcG9ydC1odHRwcyBjYS1jZXJ0aWZpY2F0ZXMgY3VybCBzb2Z0d2FyZS1wcm9wZXJ0aWVzLWNvbW1vbgpjdXJsIC1mc1NMIGh0dHBzOi8vZG93bmxvYWQuZG9ja2VyLmNvbS9saW51eC91YnVudHUvZ3BnIHwgYXB0LWtleSBhZGQgLQpjdXJsIC1zIGh0dHBzOi8vcGFja2FnZXMuY2xvdWQuZ29vZ2xlLmNvbS9hcHQvZG9jL2FwdC1rZXkuZ3BnIHwgYXB0LWtleSBhZGQgLQpjYXQgPDxFT0YgPi9ldGMvYXB0L3NvdXJjZXMubGlzdC5kL2t1YmVybmV0ZXMubGlzdApkZWIgaHR0cHM6Ly9hcHQua3ViZXJuZXRlcy5pby8ga3ViZXJuZXRlcy14ZW5pYWwgbWFpbgpFT0YKYWRkLWFwdC1yZXBvc2l0b3J5ICAgImRlYiBbYXJjaD1hbWQ2NF0gaHR0cHM6Ly9kb3dubG9hZC5kb2NrZXIuY29tL2xpbnV4L3VidW50dSAgICQobHNiX3JlbGVhc2UgLWNzKSAgIHN0YWJsZSIKYXB0LWdldCB1cGRhdGUKYXB0LWdldCB1cGdyYWRlIC15CmFwdC1nZXQgaW5zdGFsbCAteSBrdWJlbGV0PTEuMTcuNC0wMCBrdWJlYWRtPTEuMTcuNC0wMCBrdWJlY3RsPTEuMTcuNC0wMAphcHQtbWFyayBob2xkIGt1YmVsZXQga3ViZWFkbSBrdWJlY3RsCmN1cmwgLWZzU0wgaHR0cHM6Ly9kb3dubG9hZC5kb2NrZXIuY29tL2xpbnV4L3VidW50dS9ncGcgfCBhcHQta2V5IGFkZCAtCmFkZC1hcHQtcmVwb3NpdG9yeSAiZGViIFthcmNoPWFtZDY0XSBodHRwczovL2Rvd25sb2FkLmRvY2tlci5jb20vbGludXgvdWJ1bnR1IGJpb25pYyBzdGFibGUiCmFwdCB1cGRhdGUKYXB0IGluc3RhbGwgLXkgZG9ja2VyLWNlPTE4LjA2LjJ+Y2V+My0wfnVidW50dQpjYXQgPiAvZXRjL2RvY2tlci9kYWVtb24uanNvbiA8PEVPRgp7CiAgImV4ZWMtb3B0cyI6IFsibmF0aXZlLmNncm91cGRyaXZlcj1zeXN0ZW1kIl0sCiAgImxvZy1kcml2ZXIiOiAianNvbi1maWxlIiwKICAibG9nLW9wdHMiOiB7CiAgICAibWF4LXNpemUiOiAiMTAwbSIKICB9LAogICJzdG9yYWdlLWRyaXZlciI6ICJvdmVybGF5MiIKfQpFT0YKbWtkaXIgLXAgL2V0Yy9zeXN0ZW1kL3N5c3RlbS9kb2NrZXIuc2VydmljZS5kCnN5c3RlbWN0bCBkYWVtb24tcmVsb2FkCnN5c3RlbWN0bCByZXN0YXJ0IGRvY2tlcgpzd2Fwb2ZmIC1hCm12IC9ldGMvZnN0YWIgL2V0Yy9mc3RhYi5vbGQgJiYgZ3JlcCAtdiBzd2FwIC9ldGMvZnN0YWIub2xkID4gL2V0Yy9mc3RhYgpjYXQgPDxFT0YgfCB0ZWUgL2V0Yy9kZWZhdWx0L2t1YmVsZXQKS1VCRUxFVF9FWFRSQV9BUkdTPS0tY2xvdWQtcHJvdmlkZXI9ZXh0ZXJuYWwgLS1ub2RlLWxhYmVscz1wb29sPXt7Lk5vZGVHcm91cH19CkVPRgprdWJlYWRtIGpvaW4gLS1kaXNjb3ZlcnktdG9rZW4tdW5zYWZlLXNraXAtY2EtdmVyaWZpY2F0aW9uIC0tdG9rZW4ge3suQm9vdHN0cmFwVG9rZW5JRH19Lnt7LkJvb3RzdHJhcFRva2VuU2VjcmV0fX0ge3suQVBJU2VydmVyRW5kcG9pbnR9fQo="

func newTestPacketManagerRest(t *testing.T, url string) *packetManagerRest {
	manager := &packetManagerRest{
		packetManagerNodePools: map[string]*packetManagerNodePool{
			"default": {
				baseURL:           url,
				clusterName:       "cluster2",
				projectID:         "3d27fd13-0466-4878-be22-9a4b5595a3df",
				apiServerEndpoint: "147.75.102.15:6443",
				facility:          "ams1",
				plan:              "t1.small.x86",
				os:                "ubuntu_18_04",
				billing:           "hourly",
				cloudinit:         cloudinitDefault,
				reservation:       "prefer",
				hostnamePattern:   "k8s-{{.ClusterName}}-{{.NodeGroup}}-{{.RandString8}}",
			},
			"pool2": {
				baseURL:           url,
				clusterName:       "cluster2",
				projectID:         "3d27fd13-0466-4878-be22-9a4b5595a3df",
				apiServerEndpoint: "147.75.102.15:6443",
				facility:          "ams1",
				plan:              "c1.small.x86",
				os:                "ubuntu_18_04",
				billing:           "hourly",
				cloudinit:         cloudinitDefault,
				reservation:       "prefer",
				hostnamePattern:   "k8s-{{.ClusterName}}-{{.NodeGroup}}-{{.RandString8}}",
			},
		},
	}
	return manager
}
func TestListPacketDevices(t *testing.T) {
	var m *packetManagerRest
	server := NewHttpServerMockWithContentType()
	defer server.Close()
	if len(os.Getenv("PACKET_AUTH_TOKEN")) > 0 {
		// If auth token set in env, hit the actual Packet API
		m = newTestPacketManagerRest(t, "https://api.packet.net")
	} else {
		// Set up a mock Packet API
		m = newTestPacketManagerRest(t, server.URL)
		t.Logf("server URL: %v", server.URL)
		t.Logf("default packetManagerNodePool baseURL: %v", m.packetManagerNodePools["default"].baseURL)
		server.On("handleWithContentType", "/projects/"+m.packetManagerNodePools["default"].projectID+"/devices").Return("application/json", listPacketDevicesResponse).Times(2)
	}

	_, err := m.listPacketDevices(context.TODO())
	assert.NoError(t, err)

	c, err := m.nodeGroupSize("pool3")
	assert.NoError(t, err)
	assert.Equal(t, int(1), c) // One device in nodepool

	mock.AssertExpectationsForObjects(t, server)
}
