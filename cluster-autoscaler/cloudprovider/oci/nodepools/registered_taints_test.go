/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package nodepools

import (
	"reflect"
	"testing"

	apiv1 "k8s.io/api/core/v1"
	oke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/containerengine"
)

func Test_ociInitialTaintsGetterImpl_Get(t *testing.T) {
	tests := []struct {
		name    string
		np      *oke.NodePool
		want    []apiv1.Taint
		wantErr bool
	}{
		{
			name: "no kubelet-extra-args in node metadata",
			np: &oke.NodePool{
				NodeMetadata: map[string]string{},
			},
			want:    []apiv1.Taint{},
			wantErr: false,
		},
		{
			name: "empty kubelet-extra-args in node metadata",
			np: &oke.NodePool{
				NodeMetadata: map[string]string{"kubelet-extra-args": ""},
			},
			want:    []apiv1.Taint{},
			wantErr: false,
		},
		{
			name: "kubelet-extra-args has ignorable flags",
			np: &oke.NodePool{
				NodeMetadata: map[string]string{"kubelet-extra-args": "--add-dir-header --address 0.0.0.0"},
			},
			want:    []apiv1.Taint{},
			wantErr: false,
		},
		{
			name: "kubelet-extra-args has register-with-taint flag",
			np: &oke.NodePool{
				NodeMetadata: map[string]string{"kubelet-extra-args": "--register-with-taints=testTaint1=hello:NoSchedule"},
			},
			want: []apiv1.Taint{
				{
					Key:    "testTaint1",
					Value:  "hello",
					Effect: apiv1.TaintEffectNoSchedule,
				},
			},
			wantErr: false,
		},
		{
			name: "kubelet-extra-args has register-with-taint flag using space instead of =",
			np: &oke.NodePool{
				NodeMetadata: map[string]string{"kubelet-extra-args": "--register-with-taints testTaint1=hello:NoSchedule"},
			},
			want: []apiv1.Taint{
				{
					Key:    "testTaint1",
					Value:  "hello",
					Effect: apiv1.TaintEffectNoSchedule,
				},
			},
			wantErr: false,
		},
		{
			name: "kubelet-extra-args has register-with-taint and extra flag",
			np: &oke.NodePool{
				NodeMetadata: map[string]string{"kubelet-extra-args": "--register-with-taints=testTaint1=hello:NoSchedule --address 0.0.0.0"},
			},
			want: []apiv1.Taint{
				{
					Key:    "testTaint1",
					Value:  "hello",
					Effect: apiv1.TaintEffectNoSchedule,
				},
			},
			wantErr: false,
		},
		{
			name: "register-with-taint has multiple taints flag",
			np: &oke.NodePool{
				NodeMetadata: map[string]string{"kubelet-extra-args": "--register-with-taints=testTaint1=hello:NoSchedule,testTaint2=world:NoSchedule"},
			},
			want: []apiv1.Taint{
				{
					Key:    "testTaint1",
					Value:  "hello",
					Effect: apiv1.TaintEffectNoSchedule,
				},
				{
					Key:    "testTaint2",
					Value:  "world",
					Effect: apiv1.TaintEffectNoSchedule,
				},
			},
			wantErr: false,
		},
		{
			name: "register-with-taint has bad taint flag",
			np: &oke.NodePool{
				NodeMetadata: map[string]string{"kubelet-extra-args": "--register-with-taints=testTaint1=hello,world:NoSchedule"},
			},
			want:    []apiv1.Taint{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otg := &registeredTaintsGetterImpl{}
			got, err := otg.Get(tt.np)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}
