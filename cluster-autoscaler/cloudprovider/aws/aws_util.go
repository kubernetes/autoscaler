/*
Copyright 2019 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"

	klog "k8s.io/klog/v2"
)

var (
	ec2MetaDataServiceUrl          = "http://169.254.169.254"
	ec2PricingServiceUrlTemplate   = "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/%s/index.json"
	ec2PricingServiceUrlTemplateCN = "https://pricing.cn-north-1.amazonaws.com.cn/offers/v1.0/cn/AmazonEC2/current/%s/index.json"
	staticListLastUpdateTime       = "2020-12-07"
	ec2Arm64Processors             = []string{"AWS Graviton Processor", "AWS Graviton2 Processor"}
)

type response struct {
	Products map[string]product `json:"products"`
}

type product struct {
	Attributes productAttributes `json:"attributes"`
}

type productAttributes struct {
	InstanceType string `json:"instanceType"`
	VCPU         string `json:"vcpu"`
	Memory       string `json:"memory"`
	GPU          string `json:"gpu"`
	Architecture string `json:"physicalProcessor"`
}

// GenerateEC2InstanceTypes returns a map of ec2 resources
func GenerateEC2InstanceTypes(region string) (map[string]*InstanceType, error) {
	var pricingUrlTemplate string
	if strings.HasPrefix(region, "cn-") {
		pricingUrlTemplate = ec2PricingServiceUrlTemplateCN
	} else {
		pricingUrlTemplate = ec2PricingServiceUrlTemplate
	}

	instanceTypes := make(map[string]*InstanceType)

	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()

	for _, p := range partitions {
		for _, r := range p.Regions() {
			if region != "" && region != r.ID() {
				continue
			}

			url := fmt.Sprintf(pricingUrlTemplate, r.ID())
			klog.V(1).Infof("fetching %s\n", url)
			res, err := http.Get(url)
			if err != nil {
				klog.Warningf("Error fetching %s skipping...\n%s\n", url, err)
				continue
			}

			defer res.Body.Close()

			unmarshalled, err := unmarshalProductsResponse(res.Body)
			if err != nil {
				klog.Warningf("Error parsing %s skipping...\n%s\n", url, err)
				continue
			}

			for _, product := range unmarshalled.Products {
				attr := product.Attributes
				if attr.InstanceType != "" {
					instanceTypes[attr.InstanceType] = &InstanceType{
						InstanceType: attr.InstanceType,
					}
					if attr.Memory != "" && attr.Memory != "NA" {
						instanceTypes[attr.InstanceType].MemoryMb = parseMemory(attr.Memory)
					}
					if attr.VCPU != "" {
						instanceTypes[attr.InstanceType].VCPU = parseCPU(attr.VCPU)
					}
					if attr.GPU != "" {
						instanceTypes[attr.InstanceType].GPU = parseCPU(attr.GPU)
					}
					if attr.Architecture != "" {
						instanceTypes[attr.InstanceType].Architecture = parseArchitecture(attr.Architecture)
					}
				}
			}
		}
	}

	if len(instanceTypes) == 0 {
		return nil, errors.New("unable to load EC2 Instance Type list")
	}

	return instanceTypes, nil
}

// GetStaticEC2InstanceTypes return pregenerated ec2 instance type list
func GetStaticEC2InstanceTypes() (map[string]*InstanceType, string) {
	return InstanceTypes, staticListLastUpdateTime
}

func unmarshalProductsResponse(r io.Reader) (*response, error) {
	dec := json.NewDecoder(r)
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := t.(json.Delim); !ok || delim.String() != "{" {
		return nil, errors.New("Invalid products json")
	}

	unmarshalled := response{map[string]product{}}

	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return nil, err
		}

		if t == "products" {
			tt, err := dec.Token()
			if err != nil {
				return nil, err
			}
			if delim, ok := tt.(json.Delim); !ok || delim.String() != "{" {
				return nil, errors.New("Invalid products json")
			}
			for dec.More() {
				productCode, err := dec.Token()
				if err != nil {
					return nil, err
				}

				prod := product{}
				if err = dec.Decode(&prod); err != nil {
					return nil, err
				}
				unmarshalled.Products[productCode.(string)] = prod
			}
		}
	}

	t, err = dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := t.(json.Delim); !ok || delim.String() != "}" {
		return nil, errors.New("Invalid products json")
	}

	return &unmarshalled, nil
}

func parseMemory(memory string) int64 {
	reg, err := regexp.Compile("[^0-9\\.]+")
	if err != nil {
		klog.Fatal(err)
	}

	parsed := strings.TrimSpace(reg.ReplaceAllString(memory, ""))
	mem, err := strconv.ParseFloat(parsed, 64)
	if err != nil {
		klog.Fatal(err)
	}

	return int64(mem * float64(1024))
}

func parseCPU(cpu string) int64 {
	i, err := strconv.ParseInt(cpu, 10, 64)
	if err != nil {
		klog.Fatal(err)
	}
	return i
}

func parseArchitecture(archName string) string {
	for _, processor := range ec2Arm64Processors {
		if archName == processor {
			return "arm64"
		}
	}
	return "amd64"
}

// GetCurrentAwsRegion return region of current cluster without building awsManager
func GetCurrentAwsRegion() (string, error) {
	region, present := os.LookupEnv("AWS_REGION")

	if !present {
		c := aws.NewConfig().
			WithEndpoint(ec2MetaDataServiceUrl)
		sess, err := session.NewSession()
		if err != nil {
			return "", fmt.Errorf("failed to create session")
		}
		return ec2metadata.New(sess, c).Region()
	}

	return region, nil
}
