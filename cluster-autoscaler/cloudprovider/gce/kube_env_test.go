/*
Copyright 2024 The Kubernetes Authors.

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
	gce "google.golang.org/api/compute/v1"
)

func TestExtractKubeEnv(t *testing.T) {
	templateName := "instance-template"
	correctKubeEnv := "VAR1: VALUE1\nVAR2: VALUE2"
	someValue := "Lorem ipsum dolor sit amet"

	testCases := []struct {
		name        string
		template    *gce.InstanceTemplate
		wantKubeEnv KubeEnv
		wantErr     bool
	}{
		{
			name:     "template is nil",
			template: nil,
			wantErr:  true,
		},
		{
			name:     "template without instance properties",
			template: &gce.InstanceTemplate{},
			wantErr:  true,
		},
		{
			name: "template without instance properties metadata",
			template: &gce.InstanceTemplate{
				Properties: &gce.InstanceProperties{},
			},
			wantErr: true,
		},
		{
			name: "template without kube-env",
			template: &gce.InstanceTemplate{
				Name: templateName,
				Properties: &gce.InstanceProperties{
					Metadata: &gce.Metadata{
						Items: []*gce.MetadataItems{
							{Key: "key-1", Value: &someValue},
							{Key: "key-2", Value: &someValue},
						},
					},
				},
			},
			wantKubeEnv: KubeEnv{templateName: templateName},
		},
		{
			name: "template with nil kube-env",
			template: &gce.InstanceTemplate{
				Name: templateName,
				Properties: &gce.InstanceProperties{
					Metadata: &gce.Metadata{
						Items: []*gce.MetadataItems{
							{Key: "key-1", Value: &someValue},
							{Key: "key-2", Value: &someValue},
							{Key: "kube-env", Value: nil},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "template with incorrect kube-env",
			template: &gce.InstanceTemplate{
				Properties: &gce.InstanceProperties{
					Metadata: &gce.Metadata{
						Items: []*gce.MetadataItems{
							{Key: "key-1", Value: &someValue},
							{Key: "key-2", Value: &someValue},
							{Key: "kube-env", Value: &someValue},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "template with correct kube-env",
			template: &gce.InstanceTemplate{
				Name: templateName,
				Properties: &gce.InstanceProperties{
					Metadata: &gce.Metadata{
						Items: []*gce.MetadataItems{
							{Key: "key-1", Value: &someValue},
							{Key: "key-2", Value: &someValue},
							{Key: "kube-env", Value: &correctKubeEnv},
						},
					},
				},
			},
			wantKubeEnv: KubeEnv{
				templateName: templateName,
				env: map[string]string{
					"VAR1": "VALUE1",
					"VAR2": "VALUE2",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kubeEnv, err := ExtractKubeEnv(tc.template)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantKubeEnv, kubeEnv)
			}
		})
	}
}

func TestParseKubeEnv(t *testing.T) {
	templateName := "instance-template"
	testCases := []struct {
		name         string
		kubeEnvValue string
		wantKubeEnv  KubeEnv
		wantErr      bool
	}{
		{
			name:         "kube-env value is empty",
			kubeEnvValue: "",
			wantKubeEnv: KubeEnv{
				templateName: templateName,
				env:          map[string]string{},
			},
		},
		{
			name:         "kube-env value is incorrect",
			kubeEnvValue: "Lorem ipsum dolor sit amet",
			wantErr:      true,
		},
		{
			name:         "kube-env value is correct",
			kubeEnvValue: "VAR1: VALUE1\nVAR2: VALUE2",
			wantKubeEnv: KubeEnv{
				templateName: templateName,
				env: map[string]string{
					"VAR1": "VALUE1",
					"VAR2": "VALUE2",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kubeEnv, err := ParseKubeEnv(templateName, tc.kubeEnvValue)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantKubeEnv, kubeEnv)
			}
		})
	}
}

func TestKubeEnvVar(t *testing.T) {
	testCases := []struct {
		name      string
		kubeEnv   KubeEnv
		variable  string
		wantValue string
		wantFound bool
	}{
		{
			name:      "kube-env is nil",
			variable:  "VAR1",
			wantFound: false,
		},
		{
			name: "kube-env does not have this variable",
			kubeEnv: KubeEnv{
				env: map[string]string{
					"VAR1": "VALUE1",
					"VAR2": "VALUE2",
				},
			},
			variable:  "VAR3",
			wantFound: false,
		},
		{
			name: "kube-env has this variable",
			kubeEnv: KubeEnv{
				env: map[string]string{
					"VAR1": "VALUE1",
					"VAR2": "VALUE2",
				},
			},
			variable:  "VAR2",
			wantValue: "VALUE2",
			wantFound: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, found := tc.kubeEnv.Var(tc.variable)
			assert.Equal(t, tc.wantValue, value)
			assert.Equal(t, tc.wantFound, found)
		})
	}
}
