/*
Copyright 2025 The Kubernetes Authors.

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

package sidecar

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestIsNativeSidecar(t *testing.T) {
	always := corev1.ContainerRestartPolicyAlways

	tests := []struct {
		name      string
		container *corev1.Container
		want      bool
	}{
		{
			name:      "nil RestartPolicy",
			container: &corev1.Container{Name: "regular-init"},
			want:      false,
		},
		{
			name: "RestartPolicy Always",
			container: &corev1.Container{
				Name:          "sidecar",
				RestartPolicy: &always,
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsNativeSidecar(tc.container))
		})
	}
}
