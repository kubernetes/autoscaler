/*
Copyright 2018 The Kubernetes Authors.

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

package instrumentation

import "testing"

func Test_preparePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			"simple test",
			"/v1/volumes/123456",
			"/volumes/",
		},
		{
			"simple test",
			"/v1/volumes/123456/actions/attach",
			"/volumes/actions/attach",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := preparePathForLabel(tt.path); got != tt.want {
				t.Errorf("preparePathForLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}
