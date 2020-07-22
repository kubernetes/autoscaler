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

package magnum

import (
	"fmt"
	"strings"

	"k8s.io/api/core/v1"
)

const (
	scaleToZeroSupported = false
)

// NodeRef stores the name, systemUUID and providerID of a node.
// For refs which are created from fake nodes, IsFake should be true.
type NodeRef struct {
	Name       string
	SystemUUID string
	ProviderID string
	IsFake     bool
}

// isFakeNode returns true if a node object was created from a CA cloudprovider.Instance,
// or false if it is from an actual node in the cluster.
func isFakeNode(node *v1.Node) bool {
	// An actual node will have an object UID.
	// If there is no UID then the node is fake.
	return len(node.ObjectMeta.UID) == 0
}

// parseFakeProviderID takes a fake provider ID in the format
// fake:///<nodeGroupID>/<minionIndex> and returns the
// node group ID  and minion index.
func parseFakeProviderID(id string) (string, string, error) {
	id2 := strings.TrimPrefix(id, "fake:///")
	parts := strings.Split(id2, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("could not parse fake node provider ID %q", id)
	}
	return parts[0], parts[1], nil
}
