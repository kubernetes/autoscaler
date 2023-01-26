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
	"fmt"
	"gopkg.in/gcfg.v1"
	"io"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/ec2metadata"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/endpoints"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aws/aws-sdk-go/aws/session"
	"k8s.io/klog/v2"
	provider_aws "k8s.io/legacy-cloud-providers/aws"
	"os"
	"strconv"
	"strings"
)

// createAWSSDKProvider
//
// #1449 If running tests outside of AWS without AWS_REGION among environment
// variables, avoid a 5+ second EC2 Metadata lookup timeout in getRegion by
// setting and resetting AWS_REGION before calling createAWSSDKProvider:
//
// t.Setenv("AWS_REGION", "fanghorn")
func createAWSSDKProvider(configReader io.Reader) (*awsSDKProvider, error) {
	cfg, err := readAWSCloudConfig(configReader)
	if err != nil {
		klog.Errorf("Couldn't read config: %v", err)
		return nil, err
	}

	if err = validateOverrides(cfg); err != nil {
		klog.Errorf("Unable to validate custom endpoint overrides: %v", err)
		return nil, err
	}

	config := aws.NewConfig().
		WithRegion(getRegion()).
		WithEndpointResolver(getResolver(cfg))

	config, err = setMaxRetriesFromEnv(config)
	if err != nil {
		return nil, err
	}

	sess, err := session.NewSession(config)

	if err != nil {
		return nil, err
	}

	provider := &awsSDKProvider{
		session: sess,
	}

	return provider, nil
}

// setMaxRetriesFromEnv sets aws config MaxRetries by reading AWS_MAX_ATTEMPTS
// aws sdk does not auto-set these so instead of having more config options we can reuse what the aws cli
// does and read AWS_MAX_ATTEMPTS from the env https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
func setMaxRetriesFromEnv(config *aws.Config) (*aws.Config, error) {
	maxRetries := os.Getenv("AWS_MAX_ATTEMPTS")
	if maxRetries != "" {
		num, err := strconv.Atoi(maxRetries)
		if err != nil {
			return nil, err
		}
		config = config.WithMaxRetries(num)
	}
	return config, nil
}

type awsSDKProvider struct {
	session *session.Session
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

func getResolver(cfg *provider_aws.CloudConfig) endpoints.ResolverFunc {
	defaultResolver := endpoints.DefaultResolver()
	defaultResolverFn := func(service, region string,
		optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		return defaultResolver.EndpointFor(service, region, optFns...)
	}
	if len(cfg.ServiceOverride) == 0 {
		return defaultResolverFn
	}

	return func(service, region string,
		optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		for _, override := range cfg.ServiceOverride {
			if override.Service == service && override.Region == region {
				return endpoints.ResolvedEndpoint{
					URL:           override.URL,
					SigningRegion: override.SigningRegion,
					SigningMethod: override.SigningMethod,
					SigningName:   override.SigningName,
				}, nil
			}
		}
		return defaultResolver.EndpointFor(service, region, optFns...)
	}
}

// getRegion deduces the current AWS Region.
func getRegion(cfg ...*aws.Config) string {
	region, present := os.LookupEnv("AWS_REGION")
	if !present {
		sess, err := session.NewSession()
		if err != nil {
			klog.Errorf("Error getting AWS session while retrieving region: %v", err)
		} else {
			svc := ec2metadata.New(sess, cfg...)
			if r, err := svc.Region(); err == nil {
				region = r
			}
		}
	}
	return region
}
