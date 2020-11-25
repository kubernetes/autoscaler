/*
Copyright 2020 The Kubernetes Authors.

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

package ionoscloud

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ionos "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ionoscloud/ionos-cloud-sdk-go"
	"k8s.io/utils/pointer"
)

var (
	kubernetesNodes = []ionos.KubernetesNode{
		{
			Id: pointer.StringPtr("1"),
			Metadata: &ionos.KubernetesNodeMetadata{
				State: pointer.StringPtr(K8sNodeStateProvisioning),
			},
			Properties: &ionos.KubernetesNodeProperties{
				Name: pointer.StringPtr("node1"),
			},
		},
		{
			Id: pointer.StringPtr("2"),
			Metadata: &ionos.KubernetesNodeMetadata{
				State: pointer.StringPtr(K8sNodeStateProvisioned),
			},
			Properties: &ionos.KubernetesNodeProperties{
				Name: pointer.StringPtr("node2"),
			},
		},
		{
			Id: pointer.StringPtr("3"),
			Metadata: &ionos.KubernetesNodeMetadata{
				State: pointer.StringPtr(K8sNodeStateRebuilding),
			},
			Properties: &ionos.KubernetesNodeProperties{
				Name: pointer.StringPtr("node3"),
			},
		},
		{
			Id: pointer.StringPtr("4"),
			Metadata: &ionos.KubernetesNodeMetadata{
				State: pointer.StringPtr(K8sNodeStateTerminating),
			},
			Properties: &ionos.KubernetesNodeProperties{
				Name: pointer.StringPtr("node4"),
			},
		},
		{
			Id: pointer.StringPtr("5"),
			Metadata: &ionos.KubernetesNodeMetadata{
				State: pointer.StringPtr(K8sNodeStateReady),
			},
			Properties: &ionos.KubernetesNodeProperties{
				Name: pointer.StringPtr("node5"),
			},
		},
	}
	cloudproviderInstances = []cloudprovider.Instance{
		{
			Id: "ionos://1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		}, {
			Id: "ionos://2",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		}, {
			Id: "ionos://3",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		}, {
			Id: "ionos://4",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		}, {
			Id: "ionos://5",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
	}
)

func TestUtils_ConvertToInstanceId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		in := "1-2-3-4"
		want := "ionos://1-2-3-4"
		got := convertToInstanceId(in)
		require.Equal(t, want, got)
	})
}

func TestUtils_ConvertToNodeId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		in := "ionos://1-2-3-4"
		want := "1-2-3-4"
		got := convertToNodeId(in)
		require.Equal(t, want, got)
	})
}

func TestUtils_ConvertToInstances(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		in := ionos.KubernetesNodes{
			Items: &kubernetesNodes,
		}
		want := cloudproviderInstances
		got := convertToInstances(&in)
		require.Equal(t, want, got)
	})
}

func TestUtils_ConvertToInstance(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		in := ionos.KubernetesNode{
			Id: pointer.StringPtr("1"),
			Metadata: &ionos.KubernetesNodeMetadata{
				State: pointer.StringPtr(K8sNodeStateReady),
			},
		}
		want := cloudprovider.Instance{
			Id: "ionos://1",
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		}
		got := convertToInstance(in)
		require.Equal(t, want, got)
	})
}

func TestUtils_ConvertToInstanceStatus(t *testing.T) {
	tests := []struct {
		in, name string
		want     *cloudprovider.InstanceStatus
	}{
		{
			name: "success, ionos server provisioning",
			in:   K8sNodeStateProvisioning,
			want: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		}, {
			name: "success, ionos server provisioned",
			in:   K8sNodeStateProvisioned,
			want: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		}, {
			name: "success, ionos server rebuiling",
			in:   K8sNodeStateRebuilding,
			want: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		}, {
			name: "success, ionos server terminating",
			in:   K8sNodeStateTerminating,
			want: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeleting,
			},
		}, {
			name: "success, ionos server ready",
			in:   K8sNodeStateReady,
			want: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := convertToInstanceStatus(tc.in)
			require.Equal(t, tc.want, got)
		})
	}

	t.Run("Fail, unknown node state", func(t *testing.T) {
		want := &cloudprovider.InstanceStatus{
			State: 0,
			ErrorInfo: &cloudprovider.InstanceErrorInfo{
				ErrorClass:   cloudprovider.OtherErrorClass,
				ErrorCode:    ErrorCodeUnknownState,
				ErrorMessage: "Unknown node state: wrong_state",
			},
		}
		got := convertToInstanceStatus("wrong_state")
		require.Equal(t, want, got)
	})
}
