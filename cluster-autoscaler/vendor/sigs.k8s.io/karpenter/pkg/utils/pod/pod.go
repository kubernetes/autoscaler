/*
Copyright The Kubernetes Authors.

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

package pod

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NodeForPod(ctx context.Context, c client.Client, p *corev1.Pod) (*corev1.Node, error) {
	node := &corev1.Node{}
	if err := c.Get(ctx, client.ObjectKey{Name: p.Spec.NodeName}, node); err != nil {
		return nil, fmt.Errorf("getting node, %w", err)
	}
	return node, nil
}
