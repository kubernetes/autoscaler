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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateInstanceUrl(t *testing.T) {
	tests := []struct {
		name      string
		domainUrl string
		ref       GceRef
		want      string
	}{
		{
			name: "empty domain url",
			ref: GceRef{
				Project: "proj1",
				Name:    "name1",
				Zone:    "us-central1-a",
			},
			want: "https://www.googleapis.com/compute/v1/projects/proj1/zones/us-central1-a/instances/name1",
		},
		{
			name: "custom url",
			ref: GceRef{
				Project: "proj1",
				Name:    "name1",
				Zone:    "us-central1-a",
			},
			domainUrl: "https://www.googleapis.com/compute-custom/v2",
			want:      "https://www.googleapis.com/compute-custom/v2/projects/proj1/zones/us-central1-a/instances/name1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GenerateInstanceUrl(tt.domainUrl, tt.ref), "GenerateInstanceUrl(%v, %v)", tt.domainUrl, tt.ref)
		})
	}
}

func TestGenerateMigUrl(t *testing.T) {
	tests := []struct {
		name      string
		domainUrl string
		ref       GceRef
		want      string
	}{
		{
			name: "no custom url",
			ref: GceRef{
				Project: "proj1",
				Name:    "name1",
				Zone:    "us-central1-a",
			},
			want: "https://www.googleapis.com/compute/v1/projects/proj1/zones/us-central1-a/instanceGroups/name1",
		},
		{
			name:      "custom url",
			domainUrl: "https://www.googleapis.com/compute-custom/v2",
			ref: GceRef{
				Project: "proj1",
				Name:    "name1",
				Zone:    "us-central1-a",
			},
			want: "https://www.googleapis.com/compute-custom/v2/projects/proj1/zones/us-central1-a/instanceGroups/name1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GenerateMigUrl(tt.domainUrl, tt.ref), "GenerateMigUrl(%v, %v)", tt.domainUrl, tt.ref)
		})
	}
}

func TestParseIgmUrl(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantProject string
		wantZone    string
		wantName    string
		wantErr     error
	}{
		{
			name:        "default domain",
			url:         "https://www.googleapis.com/compute/v1/projects/proj1/zones/us-central1-a/instanceGroupManagers/name1",
			wantProject: "proj1",
			wantName:    "name1",
			wantZone:    "us-central1-a",
		},
		{
			name:        "custom domain",
			url:         "https://www.googleapis.com/compute_test/v1/projects/proj1/zones/us-central1-a/instanceGroupManagers/name1",
			wantProject: "proj1",
			wantName:    "name1",
			wantZone:    "us-central1-a",
		},
		{
			name:    "incorrect domain",
			url:     "https://www.googleapis.com/compute_test/v1/projects2/proj1/zones/us-central1-a/instanceGroupManagers2/name1",
			wantErr: fmt.Errorf("wrong url: expected format <url>/projects/<project-id>/zones/<zone>/instanceGroupManagers/<name>, got https://www.googleapis.com/compute_test/v1/projects2/proj1/zones/us-central1-a/instanceGroupManagers2/name1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProject, gotZone, gotName, err := ParseIgmUrl(tt.url)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			assert.Equalf(t, tt.wantProject, gotProject, "ParseIgmUrl(%v)", tt.url)
			assert.Equalf(t, tt.wantZone, gotZone, "ParseIgmUrl(%v)", tt.url)
			assert.Equalf(t, tt.wantName, gotName, "ParseIgmUrl(%v)", tt.url)
		})
	}
}

func TestParseInstanceUrl(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantProject string
		wantZone    string
		wantName    string
		wantErr     error
	}{
		{
			name:        "default domain",
			url:         "https://www.googleapis.com/compute/v1/projects/proj1/zones/us-central1-a/instances/name1",
			wantProject: "proj1",
			wantName:    "name1",
			wantZone:    "us-central1-a",
		},
		{
			name:        "custom domain",
			url:         "https://www.googleapis.com/compute_test/v1/projects/proj1/zones/us-central1-a/instances/name1",
			wantProject: "proj1",
			wantName:    "name1",
			wantZone:    "us-central1-a",
		},
		{
			name:    "incorrect domain",
			url:     "https://www.googleapis.com/compute_test/v1/projects2/proj1/zones/us-central1-a/instances2/name1",
			wantErr: fmt.Errorf("wrong url: expected format <url>/projects/<project-id>/zones/<zone>/instances/<name>, got https://www.googleapis.com/compute_test/v1/projects2/proj1/zones/us-central1-a/instances2/name1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProject, gotZone, gotName, err := ParseInstanceUrl(tt.url)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			assert.Equalf(t, tt.wantProject, gotProject, "ParseInstanceUrl(%v)", tt.url)
			assert.Equalf(t, tt.wantZone, gotZone, "ParseInstanceUrl(%v)", tt.url)
			assert.Equalf(t, tt.wantName, gotName, "ParseInstanceUrl(%v)", tt.url)
		})
	}
}

func TestParseInstanceUrlRef(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    GceRef
		wantErr error
	}{
		{
			name: "default domain",
			url:  "https://www.googleapis.com/compute/v1/projects/proj1/zones/us-central1-a/instances/name1",
			want: GceRef{
				Project: "proj1",
				Name:    "name1",
				Zone:    "us-central1-a",
			},
		},
		{
			name: "custom domain",
			url:  "https://www.googleapis.com/compute_test/v1/projects/proj1/zones/us-central1-a/instances/name1",
			want: GceRef{
				Project: "proj1",
				Name:    "name1",
				Zone:    "us-central1-a",
			},
		},
		{
			name:    "incorrect domain",
			url:     "https://www.googleapis.com/compute_test/v1/projects2/proj1/zones/us-central1-a/instances2/name1",
			wantErr: fmt.Errorf("wrong url: expected format <url>/projects/<project-id>/zones/<zone>/instances/<name>, got https://www.googleapis.com/compute_test/v1/projects2/proj1/zones/us-central1-a/instances2/name1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInstanceUrlRef(tt.url)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			assert.Equalf(t, tt.want, got, "ParseInstanceUrlRef(%v)", tt.url)
		})
	}
}

func TestParseMigUrl(t *testing.T) {

	tests := []struct {
		name        string
		url         string
		wantProject string
		wantZone    string
		wantName    string
		wantErr     error
	}{
		{
			name:        "default domain",
			url:         "https://www.googleapis.com/compute/v1/projects/proj1/zones/us-central1-a/instanceGroups/name1",
			wantProject: "proj1",
			wantName:    "name1",
			wantZone:    "us-central1-a",
		},
		{
			name:        "custom domain",
			url:         "https://www.googleapis.com/compute_test/v1/projects/proj1/zones/us-central1-a/instanceGroups/name1",
			wantProject: "proj1",
			wantName:    "name1",
			wantZone:    "us-central1-a",
		},
		{
			name:    "incorrect domain",
			url:     "https://www.googleapis.com/compute_test/v1/projects2/proj1/zones/us-central1-a/instanceGroups/name1",
			wantErr: fmt.Errorf("wrong url: expected format <url>/projects/<project-id>/zones/<zone>/instanceGroups/<name>, got https://www.googleapis.com/compute_test/v1/projects2/proj1/zones/us-central1-a/instanceGroups/name1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProject, gotZone, gotName, err := ParseMigUrl(tt.url)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			assert.Equalf(t, tt.wantProject, gotProject, "ParseMigUrl(%v)", tt.url)
			assert.Equalf(t, tt.wantZone, gotZone, "ParseMigUrl(%v)", tt.url)
			assert.Equalf(t, tt.wantName, gotName, "ParseMigUrl(%v)", tt.url)
		})
	}
}

func TestIsInstanceTemplateRegional(t *testing.T) {
	tests := []struct {
		name           string
		templateUrl    string
		expectRegional bool
		wantErr        error
	}{
		{
			name:           "Has regional instance url",
			templateUrl:    "https://www.googleapis.com/compute/v1/projects/test-project/regions/us-central1/instanceTemplates/instance-template",
			expectRegional: true,
		},
		{
			name:           "Has global instance url",
			templateUrl:    "https://www.googleapis.com/compute/v1/projects/test-project/global/instanceTemplates/instance-template",
			expectRegional: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regional, err := IsInstanceTemplateRegional(tt.templateUrl)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			assert.Equal(t, tt.expectRegional, regional)
		})
	}
}
