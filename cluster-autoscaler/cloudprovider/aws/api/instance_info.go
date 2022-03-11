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

package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

const (
	instanceInfoCacheMaxAge      = time.Hour * 6
	instanceOperatingSystemLinux = "Linux"
	instanceTenancyShared        = "Shared"
)

// TODO <ylallemant> find some API for this map - support case opened
var (
	regionNameMap = map[string]string{
		"us-east-2":      "USA East (Ohio)",
		"us-east-1":      "USA East (N. Virginia)",
		"us-west-1":      "USA West (N. California)",
		"us-west-2":      "USA West (Oregon)",
		"ap-south-1":     "Asia Pacific (Mumbai)",
		"ap-northeast-3": "Asia Pacific (Osaka-Local)",
		"ap-northeast-2": "Asia Pacific (Seoul)",
		"ap-southeast-1": "Asia Pacific (Singapore)",
		"ap-southeast-2": "Asia Pacific (Sydney)",
		"ap-northeast-1": "Asia Pacific (Tokyo)",
		"ca-central-1":   "Canada (Central)",
		"cn-north-1":     "China (Beijing)",
		"cn-northwest-1": "China (Ningxia)",
		"eu-central-1":   "EU (Frankfurt)",
		"eu-west-1":      "EU (Ireland)",
		"eu-west-2":      "EU (London)",
		"eu-west-3":      "EU (Paris)",
		"eu-north-1":     "EU (Stockholm)",
		"sa-east-1":      "South America (São Paulo)",
		"us-gov-east-1":  "AWS GovCloud (US-East)",
		"us-gov-west-1":  "AWS GovCloud (USA)",
	}
)

// InstanceInfo holds AWS EC2 instance information
type InstanceInfo struct {
	// InstanceType of the described instance
	InstanceType string
	// OnDemandPrice in USD of the ec2 instance
	OnDemandPrice float64
	// VCPU count of this instance
	VCPU int64
	// MemoryMb size in megabytes of this instance
	MemoryMb int64
	// GPU count of this instance
	GPU int64
}

type awsClient interface {
	GetProducts(input *pricing.GetProductsInput) (*pricing.GetProductsOutput, error)
}

// NewEC2InstanceInfoService is the constructor of instanceInfoService which is a wrapper for AWS Pricing API.
func NewEC2InstanceInfoService(client awsClient) *instanceInfoService {
	return &instanceInfoService{
		client: client,
		cache:  make(instanceInfoCache),
	}
}

type instanceInfoService struct {
	client awsClient
	cache  instanceInfoCache
	sync.RWMutex
}

// DescribeInstanceInfo returns the corresponding aws instance info by given instance type and availability zone.
func (s *instanceInfoService) DescribeInstanceInfo(instanceType string, region string) (*InstanceInfo, error) {
	if s.shouldSync(region) {
		if err := s.sync(region); err != nil {
			// TODO <mrcrgl> may this be tolerated for resilience
			return nil, fmt.Errorf("failed to sync aws product and price information: %v", err)
		}
	}

	if bucket, found := s.cache[region]; found {
		for _, info := range bucket.info {
			if info.InstanceType == instanceType {
				return &info, nil
			}
		}
	}
	return nil, fmt.Errorf("instance info not available for instance type %s region %s", instanceType, region)
}

func (s *instanceInfoService) shouldSync(region string) bool {
	bucket, found := s.cache[region]
	if !found {
		return true
	}

	return bucket.LastSync().Before(time.Now().Truncate(instanceInfoCacheMaxAge))
}

func (s *instanceInfoService) sync(region string) error {
	s.Lock()
	defer s.Unlock()

	start := time.Now()

	bucket, found := s.cache[region]
	if !found {
		bucket = new(regionalInstanceInfoBucket)
		s.cache[region] = bucket
	}

	response, err := s.fetch(region, bucket.ETag)
	if err != nil {
		return err
	}

	defer func() {
		klog.V(4).Infof("Synchronized aws ec2 instance information for region %s - took %s", region, time.Now().Sub(start).String())
	}()

	if response == nil {
		bucket.SetLastSync()
		return nil
	}

	instances := make([]InstanceInfo, 0)
	now := time.Now()

	for _, product := range response.Products {
		sku := product.SKU
		attr := product.Attributes

		// TODO <mrcrgl> find better solution for the case of windows installations for instance.
		if attr.OperatingSystem != instanceOperatingSystemLinux {
			continue
		}

		// We do actually only support Shared tenancy instances.
		// See for more information: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-purchasing-options.html
		if attr.Tenancy != instanceTenancyShared {
			continue
		}

		if len(attr.InstanceType) == 0 {
			continue
		}

		i := InstanceInfo{
			InstanceType: attr.InstanceType,
		}

		var err error
		if attr.Memory != "" && attr.Memory != "NA" {
			if i.MemoryMb, err = parseMemory(attr.Memory); err != nil {
				return errors.Wrapf(err, "error parsing memory for SKU %s [%s]", sku, attr.Memory)
			}
		}

		if attr.VCPU != "" {
			if i.VCPU, err = parseCPU(attr.VCPU); err != nil {
				return errors.Wrapf(err, "error parsing VCPU for SKU %s [%s]", sku, attr.VCPU)
			}
		}
		if attr.GPU != "" {
			if i.GPU, err = parseCPU(attr.GPU); err != nil {
				return errors.Wrapf(err, "error parsing GPU for SKU %s [%s]", sku, attr.GPU)
			}
		}

		for priceSKU, offer := range response.Terms.OnDemand {
			if priceSKU != sku {
				continue
			}

			var lastOfferTime time.Time
			var lastOfferPrice float64

			if offer.EffectiveDate.After(now) {
				continue
			}

			for _, price := range offer.PriceDimensions {
				if price.EndRange != "Inf" || price.Unit != "Hrs" {
					continue
				}
				p, err := strconv.ParseFloat(price.PricePerUnit.USD, 64)
				if err != nil {
					return errors.Wrapf(err, "error parsing price for SKU %s [%s]", sku, price.PricePerUnit.USD)
				}

				if p == 0.0 {
					continue
				}

				if lastOfferTime.IsZero() || lastOfferTime.After(offer.EffectiveDate) {
					lastOfferTime = offer.EffectiveDate
					lastOfferPrice = p
				}
			}

			i.OnDemandPrice = lastOfferPrice
		}

		instances = append(instances, i)
	}

	bucket.Clear()
	bucket.Add(instances...)
	bucket.SetLastSync()

	return nil
}

func (s *instanceInfoService) fetch(region string, etag string) (*response, error) {

	regionName, err := regionFullName(region)
	if err != nil {
		return nil, err
	}

	var data = new(response)
	data.Products = make(map[string]product, 0)
	data.Terms.OnDemand = make(map[string]productOffer, 0)

	// Perform first iteration
	r, err2, done, nextToken := fetchInstanceInfoIteration(region, regionName, err, s, data, "")
	if done {
		return r, err2
	}

	// Continue while 'next' token is present
	for nextToken != "" {

		r, err2, done, currentNextToken := fetchInstanceInfoIteration(region, regionName, err, s, data, nextToken)
		if done {
			return r, err2
		}

		nextToken = currentNextToken
	}

	if len(data.Products) == 0 {
		return nil, fmt.Errorf("no price information found for region %s", region)
	}

	return data, nil
}

func fetchInstanceInfoIteration(region string, regionName string, err error, s *instanceInfoService, data *response, nextToken string) (*response, error, bool, string) {
	input := &pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
		Filters: []*pricing.Filter{
			{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("location"),
				Value: aws.String(regionName),
			},
			{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("operatingSystem"),
				Value: aws.String("Linux"),
			},
			{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("capacitystatus"),
				Value: aws.String("Used"),
			},
			{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("tenancy"),
				Value: aws.String("Shared"),
			},
			{
				Type:  aws.String("TERM_MATCH"),
				Field: aws.String("preInstalledSw"),
				Value: aws.String("NA"),
			},
		},
	}
	if nextToken != "" {
		input.NextToken = aws.String(nextToken)
	}

	output, err := s.client.GetProducts(input)
	if err != nil {
		return nil, errors.Wrapf(err, "could not fetch products for region %s", region), true, ""
	}

	for _, entry := range output.PriceList {
		raw, err := protocol.EncodeJSONValue(entry, protocol.NoEscape)
		if err != nil {
			return nil, errors.Wrap(err, "could not encode back aws sdk pricing response"), true, ""
		}

		var entry = new(priceListEntry)
		if err := json.Unmarshal([]byte(raw), entry); err != nil {
			return nil, errors.Wrapf(err, "error unmarshaling pricing list entry: %s", raw), true, ""
		}

		var validTerm productOffer
		for _, term := range entry.Terms.OnDemand {
			for _, priceDimension := range term.PriceDimensions {
				if priceDimension.BeginRange == "0" && priceDimension.EndRange == "Inf" && !strings.HasPrefix(priceDimension.PricePerUnit.USD, "0.000000") {
					validTerm = term
				}
			}
		}

		if validTerm.SKU == "" {
			klog.Warningf("no on demand price was not found for instance type %s in region %s", entry.Product.Attributes.InstanceType, region)
			continue
		}

		data.Products[entry.Product.SKU] = entry.Product
		data.Terms.OnDemand[entry.Product.SKU] = validTerm

	}

	returnNextToken := ""
	if output.NextToken != nil {
		returnNextToken = *output.NextToken
	}

	return nil, nil, false, returnNextToken
}

type instanceInfoCache map[string]*regionalInstanceInfoBucket

type regionalInstanceInfoBucket struct {
	sync.RWMutex
	lastSync time.Time
	ETag     string
	info     []InstanceInfo
}

func (b *regionalInstanceInfoBucket) SetLastSync() {
	b.Lock()
	defer b.Unlock()

	b.lastSync = time.Now()
}

func (b *regionalInstanceInfoBucket) LastSync() time.Time {
	b.RLock()
	defer b.RUnlock()

	return b.lastSync
}

func (b *regionalInstanceInfoBucket) Clear() {
	b.Lock()
	defer b.Unlock()

	b.info = make([]InstanceInfo, 0)
}

func (b *regionalInstanceInfoBucket) Add(info ...InstanceInfo) {
	b.Lock()
	defer b.Unlock()

	b.info = append(b.info, info...)
}

type priceListEntry struct {
	Product product `json:"product"`
	Terms   terms   `json:"terms"`
}

type response struct {
	Products map[string]product `json:"products"`
	Terms    terms              `json:"terms"`
}

type terms struct {
	OnDemand map[string]productOffer `json:"OnDemand"`
}

type productOffer struct {
	OfferTermCode   string                           `json:"offerTermCode"`
	EffectiveDate   time.Time                        `json:"effectiveDate"`
	SKU             string                           `json:"sku"`
	PriceDimensions map[string]productPriceDimension `json:"priceDimensions"`
}

type productPriceDimension struct {
	RateCode     string       `json:"rateCode"`
	Description  string       `json:"description"`
	Unit         string       `json:"unit"`
	BeginRange   string       `json:"beginRange"`
	EndRange     string       `json:"endRange"`
	PricePerUnit pricePerUnit `json:"pricePerUnit"`
}

type pricePerUnit struct {
	USD string `json:"USD"`
}

type product struct {
	SKU        string            `json:"sku"`
	Attributes productAttributes `json:"attributes"`
}

type productAttributes struct {
	Tenancy         string `json:"tenancy"`
	InstanceType    string `json:"instanceType"`
	VCPU            string `json:"vcpu"`
	Memory          string `json:"memory"`
	GPU             string `json:"gpu"`
	OperatingSystem string `json:"operatingSystem"`
}

func parseMemory(memory string) (int64, error) {
	reg, err := regexp.Compile("[^0-9\\.]+")
	if err != nil {
		return 0, fmt.Errorf("error compiling regex %v", err)
	}

	parsed := strings.TrimSpace(reg.ReplaceAllString(memory, ""))
	mem, err := strconv.ParseFloat(parsed, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing memory [%s] %v", memory, err)
	}

	return int64(mem * float64(1024)), nil
}

func parseCPU(cpu string) (int64, error) {
	i, err := strconv.ParseInt(cpu, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing cpu [%s] %v", cpu, err)
	}
	return i, nil
}

func regionFullName(region string) (string, error) {
	if fullName, ok := regionNameMap[region]; ok {
		return fullName, nil
	}

	return "", errors.New(fmt.Sprintf("region full name not found for region: %s", region))
}
