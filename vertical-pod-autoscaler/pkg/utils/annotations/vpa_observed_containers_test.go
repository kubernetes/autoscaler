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

package annotations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestGetVpaObservedContainersValue(t *testing.T) {
	tests := []struct {
		name string
		pod  *v1.Pod
		want string
	}{
		{
			name: "creating vpa observed containers annotation",
			pod: test.Pod().
				AddContainer(test.Container().WithName("test1").Get()).
				AddContainer(test.Container().WithName("test2").Get()).
				AddContainer(test.Container().WithName("test3").Get()).
				Get(),
			want: "test1, test2, test3",
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			got := GetVpaObservedContainersValue(tc.pod)
			assert.Equal(t, got, tc.want)
		})
	}
}

func TestParseVpaObservedContainersValue(t *testing.T) {
	tests := []struct {
		name       string
		annotation string
		want       []string
		wantErr    bool
	}{
		{
			name:       "parsing correct vpa observed containers annotation",
			annotation: "test1, test2, test3",
			want:       []string{"test1", "test2", "test3"},
			wantErr:    false,
		},
		{
			name:       "parsing vpa observed containers annotation with incorrect container name",
			annotation: "test1, test2, test3_;';s",
			want:       []string(nil),
			wantErr:    true,
		},
		{
			name:       "parsing empty vpa observed containers annotation",
			annotation: "",
			want:       []string{},
			wantErr:    false,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("test case: %s", tc.name), func(t *testing.T) {
			got, gotErr := ParseVpaObservedContainersValue(tc.annotation)
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("gotErr %v, wantErr %v", (gotErr != nil), tc.wantErr)
			}
			assert.Equal(t, got, tc.want)
		})
	}
}
