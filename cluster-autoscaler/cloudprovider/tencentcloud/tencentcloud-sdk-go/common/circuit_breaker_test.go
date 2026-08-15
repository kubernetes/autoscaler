/*
Copyright 2021 The Kubernetes Authors.

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

package common

import "testing"

func Test_checkDomain(t *testing.T) {
	type args struct {
		endpoint string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"valid endpoint", args{endpoint: "cvm.tencentcloudapi.com"}, true},
		{"valid endpoint", args{endpoint: "cvm.ap-beijing.tencentcloudapi.com"}, true},
		{"invalid endpoint", args{endpoint: "cvm.tencentcloud.com"}, false},
		{"invalid endpoint", args{endpoint: "cvm.com"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkEndpoint(tt.args.endpoint); got != tt.want {
				t.Errorf("checkEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_renewUrl(t *testing.T) {
	type args struct {
		oldDomain string
		region    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"success3", args{
			oldDomain: "cvm.tencentcloudapi.com",
			region:    "ap-beijing",
		}, "cvm.ap-beijing.tencentcloudapi.com"},
		{"success4", args{
			oldDomain: "cvm.ap-beijing.tencentcloudapi.com",
			region:    "ap-shanghai",
		}, "cvm.ap-shanghai.tencentcloudapi.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renewUrl(tt.args.oldDomain, tt.args.region); got != tt.want {
				t.Errorf("renewUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
