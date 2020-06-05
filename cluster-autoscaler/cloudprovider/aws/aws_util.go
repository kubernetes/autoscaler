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
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"io/ioutil"
	klog "k8s.io/klog/v2"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	ec2MetaDataServiceUrl        = "http://169.254.169.254/latest/dynamic/instance-identity/document"
	ec2PricingServiceUrlTemplate = "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/%s/index.json"
	staticListLastUpdateTime     = "2019-10-14"
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
}

// GenerateEC2InstanceTypes returns a map of ec2 resources
func GenerateEC2InstanceTypes(region string) (map[string]*InstanceType, error) {
	instanceTypes := make(map[string]*InstanceType)

	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()

	for _, p := range partitions {
		for _, r := range p.Regions() {
			if region != "" && region != r.ID() {
				continue
			}

			url := fmt.Sprintf(ec2PricingServiceUrlTemplate, r.ID())
			klog.V(1).Infof("fetching %s\n", url)
			res, err := http.Get(url)
			if err != nil {
				klog.Warningf("Error fetching %s skipping...\n", url)
				continue
			}

			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				klog.Warningf("Error parsing %s skipping...\n", url)
				continue
			}

			var unmarshalled = response{}
			err = json.Unmarshal(body, &unmarshalled)
			if err != nil {
				klog.Warningf("Error unmarshalling %s, skip...\n", url)
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

// GetCurrentAwsRegion return region of current cluster without building awsManager
func GetCurrentAwsRegion() (string, error) {
	region, present := os.LookupEnv("AWS_REGION")

	if !present {
		klog.V(1).Infof("fetching %s\n", ec2MetaDataServiceUrl)
		res, err := http.Get(ec2MetaDataServiceUrl)
		if err != nil {
			return "", fmt.Errorf("Error fetching %s", ec2MetaDataServiceUrl)
		}

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", fmt.Errorf("Error parsing %s", ec2MetaDataServiceUrl)
		}

		var unmarshalled = map[string]string{}
		err = json.Unmarshal(body, &unmarshalled)
		if err != nil {
			klog.Warningf("Error unmarshalling %s, skip...\n", ec2MetaDataServiceUrl)
		}

		region = unmarshalled["region"]
	}

	return region, nil
}
