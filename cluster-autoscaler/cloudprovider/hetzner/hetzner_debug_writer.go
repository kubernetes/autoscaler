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

package hetzner

import (
	"k8s.io/klog/v2"
)

// debugWriter is a writer that logs to klog at level 5.
type debugWriter struct{}

func (d debugWriter) Write(p []byte) (n int, err error) {
	klog.V(5).Info(string(p))
	return len(p), nil
}
