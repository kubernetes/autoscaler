/*
Copyright 2017 The Kubernetes Authors.

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

package aws

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/ec2metadata"
	provider_aws "k8s.io/legacy-cloud-providers/aws"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestGetRegion ensures correct source supplies AWS Region.
func TestGetRegion(t *testing.T) {
	key := "AWS_REGION"
	// Ensure environment variable retains precedence.
	expected1 := "the-shire-1"
	t.Setenv(key, expected1)
	assert.Equal(t, expected1, getRegion())
	// Ensure without environment variable, EC2 Metadata is used.
	expected2 := "mordor-2"
	expectedjson := ec2metadata.EC2InstanceIdentityDocument{Region: expected2}
	js, _ := json.Marshal(expectedjson)
	os.Unsetenv(key)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(js)
	}))
	cfg := aws.NewConfig().WithEndpoint(server.URL)
	assert.Equal(t, expected2, getRegion(cfg))
}

func TestOverridesActiveConfig(t *testing.T) {
	tests := []struct {
		name string

		reader io.Reader
		aws    provider_aws.Services

		expectError        bool
		active             bool
		servicesOverridden []ServiceDescriptor
	}{
		{
			"No overrides",
			strings.NewReader(`
				[global]
				`),
			nil,
			false, false,
			[]ServiceDescriptor{},
		},
		{
			"Missing Service Name",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Region=sregion
				URL=https://s3.foo.bar
				SigningRegion=sregion
				SigningMethod = sign
				`),
			nil,
			true, false,
			[]ServiceDescriptor{},
		},
		{
			"Missing Service Region",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Service=s3
				URL=https://s3.foo.bar
				SigningRegion=sregion
				SigningMethod = sign
				`),
			nil,
			true, false,
			[]ServiceDescriptor{},
		},
		{
			"Missing URL",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Service="s3"
				Region=sregion
				SigningRegion=sregion
				SigningMethod = sign
				`),
			nil,
			true, false,
			[]ServiceDescriptor{},
		},
		{
			"Missing Signing Region",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Service=s3
				Region=sregion
				URL=https://s3.foo.bar
				SigningMethod = sign
				`),
			nil,
			true, false,
			[]ServiceDescriptor{},
		},
		{
			"Active Overrides",
			strings.NewReader(`
				[Global]
				[ServiceOverride "1"]
				Service = "s3      "
				Region = sregion
				URL = https://s3.foo.bar
				SigningRegion = sregion
				SigningMethod = v4
				`),
			nil,
			false, true,
			[]ServiceDescriptor{{name: "s3", region: "sregion", signingRegion: "sregion", signingMethod: "v4"}},
		},
		{
			"Multiple Overridden Services",
			strings.NewReader(`
				[Global]
				vpc = vpc-abc1234567
				[ServiceOverride "1"]
				Service=s3
				Region=sregion1
				URL=https://s3.foo.bar
				SigningRegion=sregion1
				SigningMethod = v4
				[ServiceOverride "2"]
				Service=ec2
				Region=sregion2
				URL=https://ec2.foo.bar
				SigningRegion=sregion2
				SigningMethod = v4
				`),
			nil,
			false, true,
			[]ServiceDescriptor{{name: "s3", region: "sregion1", signingRegion: "sregion1", signingMethod: "v4"},
				{name: "ec2", region: "sregion2", signingRegion: "sregion2", signingMethod: "v4"}},
		},
		{
			"Duplicate Services",
			strings.NewReader(`
				[Global]
				vpc = vpc-abc1234567
				[ServiceOverride "1"]
				Service=s3
				Region=sregion1
				URL=https://s3.foo.bar
				SigningRegion=sregion
				SigningMethod = sign
				[ServiceOverride "2"]
				Service=s3
				Region=sregion1
				URL=https://s3.foo.bar
				SigningRegion=sregion
				SigningMethod = sign
				`),
			nil,
			true, false,
			[]ServiceDescriptor{},
		},
		{
			"Multiple Overridden Services in Multiple regions",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
			 	Service=s3
				Region=region1
				URL=https://s3.foo.bar
				SigningRegion=sregion1
				[ServiceOverride "2"]
				Service=ec2
				Region=region2
				URL=https://ec2.foo.bar
				SigningRegion=sregion
				SigningMethod = v4
				`),
			nil,
			false, true,
			[]ServiceDescriptor{{name: "s3", region: "region1", signingRegion: "sregion1", signingMethod: ""},
				{name: "ec2", region: "region2", signingRegion: "sregion", signingMethod: "v4"}},
		},
		{
			"Multiple regions, Same Service",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Service=s3
				Region=region1
				URL=https://s3.foo.bar
				SigningRegion=sregion1
				SigningMethod = v3
				[ServiceOverride "2"]
				Service=s3
				Region=region2
				URL=https://s3.foo.bar
				SigningRegion=sregion1
				SigningMethod = v4
				SigningName = "name"
				`),
			nil,
			false, true,
			[]ServiceDescriptor{{name: "s3", region: "region1", signingRegion: "sregion1", signingMethod: "v3"},
				{name: "s3", region: "region2", signingRegion: "sregion1", signingMethod: "v4", signingName: "name"}},
		},
	}

	for _, test := range tests {
		t.Logf("Running test case %s", test.name)
		cfg, err := readAWSCloudConfig(test.reader)
		if err == nil {
			err = validateOverrides(cfg)
		}
		if test.expectError {
			if err == nil {
				t.Errorf("Should error for case %s (cfg=%v)", test.name, cfg)
			}
		} else {
			if err != nil {
				t.Errorf("Should succeed for case: %s, got %v", test.name, err)
			}

			if len(cfg.ServiceOverride) != len(test.servicesOverridden) {
				t.Errorf("Expected %d overridden services, received %d for case %s",
					len(test.servicesOverridden), len(cfg.ServiceOverride), test.name)
			} else {
				for _, sd := range test.servicesOverridden {
					var found *struct {
						Service       string
						Region        string
						URL           string
						SigningRegion string
						SigningMethod string
						SigningName   string
					}
					for _, v := range cfg.ServiceOverride {
						if v.Service == sd.name && v.Region == sd.region {
							found = v
							break
						}
					}
					if found == nil {
						t.Errorf("Missing override for service %s in case %s",
							sd.name, test.name)
					} else {
						if found.SigningRegion != sd.signingRegion {
							t.Errorf("Expected signing region '%s', received '%s' for case %s",
								sd.signingRegion, found.SigningRegion, test.name)
						}
						if found.SigningMethod != sd.signingMethod {
							t.Errorf("Expected signing method '%s', received '%s' for case %s",
								sd.signingMethod, found.SigningRegion, test.name)
						}
						targetName := fmt.Sprintf("https://%s.foo.bar", sd.name)
						if found.URL != targetName {
							t.Errorf("Expected Endpoint '%s', received '%s' for case %s",
								targetName, found.URL, test.name)
						}
						if found.SigningName != sd.signingName {
							t.Errorf("Expected signing name '%s', received '%s' for case %s",
								sd.signingName, found.SigningName, test.name)
						}

						fn := getResolver(cfg)
						ep1, e := fn(sd.name, sd.region, nil)
						if e != nil {
							t.Errorf("Expected a valid endpoint for %s in case %s",
								sd.name, test.name)
						} else {
							targetName := fmt.Sprintf("https://%s.foo.bar", sd.name)
							if ep1.URL != targetName {
								t.Errorf("Expected endpoint url: %s, received %s in case %s",
									targetName, ep1.URL, test.name)
							}
							if ep1.SigningRegion != sd.signingRegion {
								t.Errorf("Expected signing region '%s', received '%s' in case %s",
									sd.signingRegion, ep1.SigningRegion, test.name)
							}
							if ep1.SigningMethod != sd.signingMethod {
								t.Errorf("Expected signing method '%s', received '%s' in case %s",
									sd.signingMethod, ep1.SigningRegion, test.name)
							}
						}
					}
				}
			}
		}
	}
}
