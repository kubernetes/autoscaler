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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"gopkg.in/gcfg.v1"
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
	cloudCfg, err := readAWSCloudConfig(configReader)
	if err != nil {
		klog.Errorf("Couldn't read config: %v", err)
		return nil, err
	}

	if err = validateOverrides(cloudCfg); err != nil {
		klog.Errorf("Unable to validate custom endpoint overrides: %v", err)
		return nil, err
	}

	ctx := context.Background()
	
	// Load the AWS SDK v2 config
	var opts []func(*config.LoadOptions) error
	
	// Get region
	region := getRegion(ctx)
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	
	// Set max retries from environment
	maxRetries := os.Getenv("AWS_MAX_ATTEMPTS")
	if maxRetries != "" {
		num, err := strconv.Atoi(maxRetries)
		if err != nil {
			return nil, err
		}
		opts = append(opts, config.WithRetryMaxAttempts(num))
	}
	
	// Add custom endpoint resolver if configured
	if len(cloudCfg.ServiceOverride) > 0 {
		opts = append(opts, config.WithEndpointResolverWithOptions(getResolverV2(cloudCfg)))
	}
	
	// TODO: Add user agent middleware for cluster-autoscaler version tracking
	// User agent handling in SDK v2 requires custom middleware implementation
	
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	provider := &awsSDKProvider{
		cfg: cfg,
	}

	return provider, nil
}

type awsSDKProvider struct {
	cfg aws.Config
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

func getResolverV2(cfg *provider_aws.CloudConfig) aws.EndpointResolverWithOptionsFunc {
	return func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		for _, override := range cfg.ServiceOverride {
			if override.Service == service && override.Region == region {
				return aws.Endpoint{
					URL:           override.URL,
					SigningRegion: override.SigningRegion,
					SigningMethod: override.SigningMethod,
					SigningName:   override.SigningName,
				}, nil
			}
		}
		// Return unresolved to use default resolver
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	}
}

// getRegion deduces the current AWS Region.
func getRegion(ctx context.Context) string {
	region, present := os.LookupEnv("AWS_REGION")
	if !present {
		// Try to get region from EC2 instance metadata
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			klog.Errorf("Error loading AWS config while retrieving region: %v", err)
			return ""
		}
		
		client := imds.NewFromConfig(cfg)
		result, err := client.GetRegion(ctx, &imds.GetRegionInput{})
		if err == nil && result.Region != "" {
			region = result.Region
		}
	}
	return region
}
