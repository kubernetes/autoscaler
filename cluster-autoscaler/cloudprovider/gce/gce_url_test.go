/*
Copyright 2016 The Kubernetes Authors.

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

package gce

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUrl(t *testing.T) {
	proj, zone, name, err := parseGceUrl("https://www.googleapis.com/compute/v1/projects/mwielgus-proj/zones/us-central1-b/instanceGroups/kubernetes-minion-group", "instanceGroups")
	assert.Nil(t, err)
	assert.Equal(t, "mwielgus-proj", proj)
	assert.Equal(t, "us-central1-b", zone)
	assert.Equal(t, "kubernetes-minion-group", name)

	// Cluster Autoscaler previously used this format for MIG id (with "content" instead of "www"). Make sure it's still accepted
	// just to be safe.
	proj, zone, name, err = parseGceUrl("https://content.googleapis.com/compute/v1/projects/mwielgus-proj/zones/us-central1-b/instanceGroups/kubernetes-minion-group", "instanceGroups")
	assert.Nil(t, err)
	assert.Equal(t, "mwielgus-proj", proj)
	assert.Equal(t, "us-central1-b", zone)
	assert.Equal(t, "kubernetes-minion-group", name)

	_, _, _, err = parseGceUrl("www.onet.pl", "instanceGroups")
	assert.NotNil(t, err)

	_, _, _, err = parseGceUrl("https://www.googleapis.com/compute/vabc/projects/mwielgus-proj/zones/us-central1-b/instanceGroups/kubernetes-minion-group", "instanceGroups")
	assert.NotNil(t, err)
}
