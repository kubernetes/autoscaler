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

package cloudprovider

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMIGAutoDiscoverySpecs(t *testing.T) {
	cases := []struct {
		name    string
		specs   []string
		want    []MIGAutoDiscoveryConfig
		wantErr bool
	}{
		{
			name: "GoodSpecs",
			specs: []string{
				"mig:namePrefix=pfx,min=0,max=10",
				"mig:namePrefix=anotherpfx,min=1,max=2",
			},
			want: []MIGAutoDiscoveryConfig{
				{Re: regexp.MustCompile("^pfx.+"), MinSize: 0, MaxSize: 10},
				{Re: regexp.MustCompile("^anotherpfx.+"), MinSize: 1, MaxSize: 2},
			},
		},
		{
			name:    "MissingMIGType",
			specs:   []string{"namePrefix=pfx,min=0,max=10"},
			wantErr: true,
		},
		{
			name:    "WrongType",
			specs:   []string{"asg:namePrefix=pfx,min=0,max=10"},
			wantErr: true,
		},
		{
			name:    "UnknownKey",
			specs:   []string{"mig:namePrefix=pfx,min=0,max=10,unknown=hi"},
			wantErr: true,
		},
		{
			name:    "NonIntegerMin",
			specs:   []string{"mig:namePrefix=pfx,min=a,max=10"},
			wantErr: true,
		},
		{
			name:    "NonIntegerMax",
			specs:   []string{"mig:namePrefix=pfx,min=1,max=donkey"},
			wantErr: true,
		},
		{
			name:    "PrefixDoesNotCompileToRegexp",
			specs:   []string{"mig:namePrefix=a),min=1,max=10"},
			wantErr: true,
		},
		{
			name:    "KeyMissingValue",
			specs:   []string{"mig:namePrefix=prefix,min=,max=10"},
			wantErr: true,
		},
		{
			name:    "ValueMissingKey",
			specs:   []string{"mig:namePrefix=prefix,=0,max=10"},
			wantErr: true,
		},
		{
			name:    "KeyMissingSeparator",
			specs:   []string{"mig:namePrefix=prefix,min,max=10"},
			wantErr: true,
		},
		{
			name:    "TooManySeparators",
			specs:   []string{"mig:namePrefix=prefix,min=0,max=10=20"},
			wantErr: true,
		},
		{
			name:    "PrefixIsEmpty",
			specs:   []string{"mig:namePrefix=,min=0,max=10"},
			wantErr: true,
		},
		{
			name:    "PrefixIsMissing",
			specs:   []string{"mig:min=0,max=10"},
			wantErr: true,
		},
		{
			name:    "MaxBelowMin",
			specs:   []string{"mig:namePrefix=prefix,min=10,max=1"},
			wantErr: true,
		},
		{
			name:    "MaxIsZero",
			specs:   []string{"mig:namePrefix=prefix,min=0,max=0"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			do := NodeGroupDiscoveryOptions{NodeGroupAutoDiscoverySpecs: tc.specs}
			got, err := do.ParseMIGAutoDiscoverySpecs()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, assert.ObjectsAreEqualValues(tc.want, got), "\ngot: %#v\nwant: %#v", got, tc.want)
		})
	}
}

func TestParseASGAutoDiscoverySpecs(t *testing.T) {
	cases := []struct {
		name    string
		specs   []string
		want    []ASGAutoDiscoveryConfig
		wantErr bool
	}{
		{
			name: "GoodSpecs",
			specs: []string{
				"asg:tag=tag,anothertag",
				"asg:tag=cooltag,anothertag",
				"asg:tag=label=value,anothertag",
			},
			want: []ASGAutoDiscoveryConfig{
				{Tags: map[string]string{"tag": "", "anothertag": ""}},
				{Tags: map[string]string{"cooltag": "", "anothertag": ""}},
				{Tags: map[string]string{"label": "value", "anothertag": ""}},
			},
		},
		{
			name:    "MissingASGType",
			specs:   []string{"tag=tag,anothertag"},
			wantErr: true,
		},
		{
			name:    "WrongType",
			specs:   []string{"mig:tag=tag,anothertag"},
			wantErr: true,
		},
		{
			name:    "KeyMissingValue",
			specs:   []string{"asg:tag="},
			wantErr: true,
		},
		{
			name:    "ValueMissingKey",
			specs:   []string{"asg:=tag"},
			wantErr: true,
		},
		{
			name:    "KeyMissingSeparator",
			specs:   []string{"asg:tag"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			do := NodeGroupDiscoveryOptions{NodeGroupAutoDiscoverySpecs: tc.specs}
			got, err := do.ParseASGAutoDiscoverySpecs()
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(t, assert.ObjectsAreEqualValues(tc.want, got), "\ngot: %#v\nwant: %#v", got, tc.want)
		})
	}
}
