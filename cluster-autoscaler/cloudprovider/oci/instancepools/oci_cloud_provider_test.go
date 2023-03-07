/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package instancepools

import (
	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	"testing"
)

func Test_getPoolType(t *testing.T) {
	tests := []struct {
		name    string
		groups  []string
		want    string
		wantErr bool
	}{
		{
			name:    "base case single type nodepool",
			groups:  []string{"ocid1.nodepool.oc1.ap-melbourne-1.xxx", "ocid1.nodepool.oc1.ap-melbourne-1.yyy", "ocid1.nodepool.oc1.ap-melbourne-1.zzz"},
			want:    "nodepool",
			wantErr: false,
		},
		{
			name:    "base case single type instancepool",
			groups:  []string{"ocid1.instancepool.oc1.ap-melbourne-1.xxx", "ocid1.instancepool.oc1.ap-melbourne-1.yyy", "ocid1.instancepool.oc1.ap-melbourne-1.zzz"},
			want:    "instancepool",
			wantErr: false,
		},
		{
			name:    "empty should pass through",
			groups:  []string{},
			want:    "",
			wantErr: false,
		},
		{
			name:    "mixed type",
			groups:  []string{"ocid1.nodepool.oc1.ap-melbourne-1.xxx", "ocid1.instancepool.oc1.ap-melbourne-1.yyy"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ocicommon.GetAllPoolTypes(tt.groups)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPoolType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getPoolType() got = %v, want %v", got, tt.want)
			}
		})
	}
}
