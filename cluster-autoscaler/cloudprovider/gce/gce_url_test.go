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
	// Valid 'www' URL for instanceGroups
	proj, zone, name, err := parseGceUrl("https://www.googleapis.com/compute/v1/projects/mwielgus-proj/zones/us-central1-b/instanceGroups/kubernetes-minion-group", "instanceGroups")
	assert.Nil(t, err)
	assert.Equal(t, "mwielgus-proj", proj)
	assert.Equal(t, "us-central1-b", zone)
	assert.Equal(t, "kubernetes-minion-group", name)

	// Valid 'www' URL for instances
	proj, zone, name, err = parseGceUrl("https://www.googleapis.com/compute/v1/projects/mwielgus-proj/zones/us-central1-b/instances/my-instance-1", "instances")
	assert.Nil(t, err)
	assert.Equal(t, "mwielgus-proj", proj)
	assert.Equal(t, "us-central1-b", zone)
	assert.Equal(t, "my-instance-1", name)

	// Valid 'content' URL for instanceGroups
	proj, zone, name, err = parseGceUrl("https://content.googleapis.com/compute/v1/projects/mwielgus-proj/zones/us-central1-b/instanceGroups/kubernetes-minion-group", "instanceGroups")
	assert.Nil(t, err)
	assert.Equal(t, "mwielgus-proj", proj)
	assert.Equal(t, "us-central1-b", zone)
	assert.Equal(t, "kubernetes-minion-group", name)

	// Valid 'content' URL for instances
	proj, zone, name, err = parseGceUrl("https://content.googleapis.com/compute/v1/projects/mwielgus-proj/zones/us-central1-f/instances/my-instance-2", "instanceGroups")
	assert.Nil(t, err)
	assert.Equal(t, "mwielgus-proj", proj)
	assert.Equal(t, "us-central1-f", zone)
	assert.Equal(t, "my-instance-2", name)

	// Don't validate subdomains; test that we accept anything ending in .googleapis.com
	proj, zone, name, err = parseGceUrl("https://18adjadj391.googleapis.com/compute/v1/projects/mwielgus-proj/zones/us-central1-b/instanceGroups/kubernetes-minion-group", "instanceGroups")
	assert.Nil(t, err)
	assert.Equal(t, "mwielgus-proj", proj)
	assert.Equal(t, "us-central1-b", zone)
	assert.Equal(t, "kubernetes-minion-group", name)

	// Invalid TLD
	_, _, _, err = parseGceUrl("https://www.googleapis.net/compute/v1/projects/mwielgus-proj/zones/us-central1-b/instanceGroups/kubernetes-minion-group", "instanceGroups")
	assert.NotNil(t, err)

	// Empty 'expectedResource'
	_, _, _, err = parseGceUrl("googleapis.net/compute/v1/projects/mwielgus-proj/zones/us-central1-b/instanceGroups/kubernetes-minion-group", "")
	assert.NotNil(t, err)

	// Invalid domain & URL
	_, _, _, err = parseGceUrl("www.onet.pl", "instanceGroups")
	assert.NotNil(t, err)

	// Validate the URL suffix
	_, _, _, err = parseGceUrl("https://content.googleapis.com/compute/vabc/projects/mwielgus-proj/zones/us-central1-b/instanceGroups/kubernetes-minion-group", "instanceGroups")
	assert.NotNil(t, err)
}
