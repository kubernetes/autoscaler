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

package aws

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"gopkg.in/gcfg.v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/aws/middleware"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/config"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/autoscaling"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/ec2"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go-v2/service/eks"
	smithymiddleware "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/smithy-go/middleware"
	"k8s.io/autoscaler/cluster-autoscaler/version"
	provider_aws "k8s.io/cloud-provider-aws/pkg/providers/v1"
	"k8s.io/klog/v2"
)

// createAWSSDKProvider
//
// #1449 If running tests outside of AWS without AWS_REGION among environment
// variables, avoid a 5+ second EC2 Metadata lookup timeout in getRegion by
// setting and resetting AWS_REGION before calling createAWSSDKProvider:
//
// t.Setenv("AWS_REGION", "fanghorn")
func createAWSSDKProvider(configReader io.Reader) (*awsSDKProvider, error) {
	cloudConfig, err := readAWSCloudConfig(configReader)
	if err != nil {
		klog.Errorf("Couldn't read config: %v", err)
		return nil, err
	}

	if err = validateOverrides(cloudConfig); err != nil {
		klog.Errorf("Unable to validate custom endpoint overrides: %v", err)
		return nil, err
	}

	// Configure all options before building the config
	loadOpts := []func(*config.LoadOptions) error{
		config.WithRegion(getRegion()),
		config.WithAPIOptions([]func(*smithymiddleware.Stack) error{
			// add cluster-autoscaler to the user-agent to make it easier to identify
			middleware.AddUserAgentKeyValue("cluster-autoscaler", version.ClusterAutoscalerVersion),
		}),
	}
	if maxRetries, isSet, err := getMaxRetriesFromEnv(); isSet {
		if err != nil {
			klog.Errorf("Error getting max retries from env: %v", err)
			return nil, err
		}
		loadOpts = append(loadOpts, config.WithRetryMaxAttempts(maxRetries))
	}

	awsConfig, err := config.LoadDefaultConfig(context.Background(), loadOpts...)
	if err != nil {
		klog.Errorf("Unable to load default aws config: %v", err)
		return nil, err
	}

	provider := &awsSDKProvider{
		cfg:         awsConfig,
		cloudConfig: cloudConfig,
	}
	return provider, nil
}

// getMaxRetriesFromEnv retrieves the MaxRetries configuration by reading AWS_MAX_ATTEMPTS
// aws sdk does not auto-set these so instead of having more config options we can reuse what the aws cli
// does and read AWS_MAX_ATTEMPTS from the env https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
func getMaxRetriesFromEnv() (int, bool, error) {
	maxRetries := os.Getenv("AWS_MAX_ATTEMPTS")
	if maxRetries != "" {
		num, err := strconv.Atoi(maxRetries)
		if err != nil {
			return 0, true, err
		}
		return num, true, nil
	}
	return 0, false, nil
}

type awsSDKProvider struct {
	cfg         aws.Config
	cloudConfig *provider_aws.CloudConfig
}

// readAWSCloudConfig reads an instance of AWSCloudConfig from config reader.
func readAWSCloudConfig(config io.Reader) (*provider_aws.CloudConfig, error) {
	var cfg provider_aws.CloudConfig
	var err error

	if config != nil {
		err = gcfg.ReadInto(&cfg, config)
		if err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

func validateOverrides(cfg *provider_aws.CloudConfig) error {
	if len(cfg.ServiceOverride) == 0 {
		return nil
	}
	set := make(map[string]bool)
	for onum, ovrd := range cfg.ServiceOverride {
		// Note: gcfg does not space trim, so we have to when comparing to empty string ""
		name := strings.TrimSpace(ovrd.Service)
		if name == "" {
			return fmt.Errorf("service name is missing [Service is \"\"] in override %s", onum)
		}
		// insure the map service name is space trimmed
		ovrd.Service = name

		region := strings.TrimSpace(ovrd.Region)
		if region == "" {
			return fmt.Errorf("service region is missing [Region is \"\"] in override %s", onum)
		}
		// insure the map region is space trimmed
		ovrd.Region = region

		url := strings.TrimSpace(ovrd.URL)
		if url == "" {
			return fmt.Errorf("url is missing [URL is \"\"] in override %s", onum)
		}
		signingRegion := strings.TrimSpace(ovrd.SigningRegion)
		if signingRegion == "" {
			return fmt.Errorf("signingRegion is missing [SigningRegion is \"\"] in override %s", onum)
		}
		signature := name + "_" + region
		if set[signature] {
			return fmt.Errorf("duplicate entry found for service override [%s] (%s in %s)", onum, name, region)
		}
		set[signature] = true
	}
	return nil
}

type ec2OverrideResolver struct {
	defaultResolver ec2.EndpointResolver
	cloudConfig     *provider_aws.CloudConfig
}

var _ ec2.EndpointResolver = &ec2OverrideResolver{}

func newEc2OverrideResolver(cloudConfig *provider_aws.CloudConfig) *ec2OverrideResolver {
	return &ec2OverrideResolver{
		defaultResolver: ec2.NewDefaultEndpointResolver(),
		cloudConfig:     cloudConfig,
	}
}

func (r *ec2OverrideResolver) ResolveEndpoint(region string, options ec2.EndpointResolverOptions) (aws.Endpoint, error) {
	for _, override := range r.cloudConfig.ServiceOverride {
		if override.Service == "ec2" && override.Region == region {
			return aws.Endpoint{
				URL:           override.URL,
				SigningRegion: override.SigningRegion,
				SigningMethod: override.SigningMethod,
				SigningName:   override.SigningName,
			}, nil
		}
	}
	return r.defaultResolver.ResolveEndpoint(region, options)
}

type autoscalingOverrideResolver struct {
	defaultResolver autoscaling.EndpointResolver
	cloudConfig     *provider_aws.CloudConfig
}

var _ autoscaling.EndpointResolver = &autoscalingOverrideResolver{}

func newAutoscalingOverrideResolver(cloudConfig *provider_aws.CloudConfig) *autoscalingOverrideResolver {
	return &autoscalingOverrideResolver{
		defaultResolver: autoscaling.NewDefaultEndpointResolver(),
		cloudConfig:     cloudConfig,
	}
}

func (r *autoscalingOverrideResolver) ResolveEndpoint(region string, options autoscaling.EndpointResolverOptions) (aws.Endpoint, error) {
	for _, override := range r.cloudConfig.ServiceOverride {
		if override.Service == "autoscaling" && override.Region == region {
			return aws.Endpoint{
				URL:           override.URL,
				SigningRegion: override.SigningRegion,
				SigningMethod: override.SigningMethod,
				SigningName:   override.SigningName,
			}, nil
		}
	}
	return r.defaultResolver.ResolveEndpoint(region, options)
}

type eksOverrideResolver struct {
	defaultResolver eks.EndpointResolver
	cloudConfig     *provider_aws.CloudConfig
}

var _ eks.EndpointResolver = &eksOverrideResolver{}

func newEksOverrideResolver(cloudConfig *provider_aws.CloudConfig) *eksOverrideResolver {
	return &eksOverrideResolver{
		defaultResolver: eks.NewDefaultEndpointResolver(),
		cloudConfig:     cloudConfig,
	}
}

func (r *eksOverrideResolver) ResolveEndpoint(region string, options eks.EndpointResolverOptions) (aws.Endpoint, error) {
	for _, override := range r.cloudConfig.ServiceOverride {
		if override.Service == "eks" && override.Region == region {
			return aws.Endpoint{
				URL:           override.URL,
				SigningRegion: override.SigningRegion,
				SigningMethod: override.SigningMethod,
				SigningName:   override.SigningName,
			}, nil
		}
	}
	return r.defaultResolver.ResolveEndpoint(region, options)
}

// getRegion deduces the current AWS Region.
func getRegion() string {
	region, present := os.LookupEnv("AWS_REGION")
	if !present {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			klog.Errorf("Unable to load default aws config: %v", err)
		}
		region = cfg.Region
	}
	return region
}
