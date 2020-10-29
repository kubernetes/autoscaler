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

package framework

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Framework struct {
	ClientConfig clientcmd.ClientConfig
	ClientSet    clientset.Interface
	Namespace    *v1.Namespace
	Provider     Provider
	T            *testing.T
}

type Provider interface {
	FrameworkBeforeEach(f *Framework)
	FrameworkAfterEach(f *Framework)

	DisableAutoscaler(nodeGroup string) error
	EnableAutoscaler(nodeGroup string, minSize, maxSize int) error
	GroupSize(group string) (int, error)
	ResizeGroup(group string, size int32) error
}
