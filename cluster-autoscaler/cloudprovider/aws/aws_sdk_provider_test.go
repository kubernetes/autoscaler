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
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/stretchr/testify/assert"
	provider_aws "k8s.io/cloud-provider-aws/pkg/providers/v1"
)

// TestGetRegion ensures correct source supplies AWS Region.
func TestGetRegionFromEnvironmentVariable(t *testing.T) {
	key := "AWS_REGION"
	// Ensure environment variable retains precedence.
	expected1 := "the-shire-1"
	t.Setenv(key, expected1)
	assert.Equal(t, expected1, getRegion())
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
				URL=https://eks.foo.bar
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
				Service=eks
				URL=https://eks.foo.bar
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
				Service="eks"
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
				Service=eks
				Region=sregion
				URL=https://eks.foo.bar
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
				Service = "eks      "
				Region = sregion
				URL = https://eks.foo.bar
				SigningRegion = sregion
				SigningMethod = v4
				`),
			nil,
			false, true,
			[]ServiceDescriptor{{name: "eks", region: "sregion", signingRegion: "sregion", signingMethod: "v4"}},
		},
		{
			"Multiple Overridden Services",
			strings.NewReader(`
				[Global]
				vpc = vpc-abc1234567
				[ServiceOverride "1"]
				Service=eks
				Region=sregion1
				URL=https://eks.foo.bar
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
			[]ServiceDescriptor{{name: "eks", region: "sregion1", signingRegion: "sregion1", signingMethod: "v4"},
				{name: "ec2", region: "sregion2", signingRegion: "sregion2", signingMethod: "v4"}},
		},
		{
			"Duplicate Services",
			strings.NewReader(`
				[Global]
				vpc = vpc-abc1234567
				[ServiceOverride "1"]
				Service=eks
				Region=sregion1
				URL=https://eks.foo.bar
				SigningRegion=sregion
				SigningMethod = sign
				[ServiceOverride "2"]
				Service=eks
				Region=sregion1
				URL=https://eks.foo.bar
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
			 	Service=eks
				Region=region1
				URL=https://eks.foo.bar
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
			[]ServiceDescriptor{{name: "eks", region: "region1", signingRegion: "sregion1", signingMethod: ""},
				{name: "ec2", region: "region2", signingRegion: "sregion", signingMethod: "v4"}},
		},
		{
			"Multiple regions, Same Service",
			strings.NewReader(`
				[global]
				[ServiceOverride "1"]
				Service=eks
				Region=region1
				URL=https://eks.foo.bar
				SigningRegion=sregion1
				SigningMethod = v3
				[ServiceOverride "2"]
				Service=eks
				Region=region2
				URL=https://eks.foo.bar
				SigningRegion=sregion1
				SigningMethod = v4
				SigningName = "name"
				`),
			nil,
			false, true,
			[]ServiceDescriptor{{name: "eks", region: "region1", signingRegion: "sregion1", signingMethod: "v3"},
				{name: "eks", region: "region2", signingRegion: "sregion1", signingMethod: "v4", signingName: "name"}},
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

						var endpoint aws.Endpoint
						switch sd.name {
						case "autoscaling":
							resolver := newAutoscalingOverrideResolver(cfg)
							endpoint, err = resolver.ResolveEndpoint(sd.region, autoscaling.EndpointResolverOptions{})
						case "ec2":
							resolver := newEc2OverrideResolver(cfg)
							endpoint, err = resolver.ResolveEndpoint(sd.region, ec2.EndpointResolverOptions{})
						case "eks":
							resolver := newEksOverrideResolver(cfg)
							endpoint, err = resolver.ResolveEndpoint(sd.region, eks.EndpointResolverOptions{})
						default:
							t.Errorf("Unexpected service %s not supported", sd.name)
						}
						if err != nil {
							t.Errorf("Failed to resolve endpoint for service %s in case %s", sd.name, test.name)
						}

						if endpoint.URL != targetName {
							t.Errorf("Expected endpoint url: %s, received %s in case %s",
								targetName, endpoint.URL, test.name)
						}
						if endpoint.SigningRegion != sd.signingRegion {
							t.Errorf("Expected signing region '%s', received '%s' in case %s",
								sd.signingRegion, endpoint.SigningRegion, test.name)
						}
						if endpoint.SigningMethod != sd.signingMethod {
							t.Errorf("Expected signing method '%s', received '%s' in case %s",
								sd.signingMethod, endpoint.SigningRegion, test.name)
						}
					}
				}
			}
		}
	}
}
