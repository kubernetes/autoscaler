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

//go:generate go run ec2_instance_types/gen.go

package aws

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gopkg.in/gcfg.v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	provider_aws "k8s.io/legacy-cloud-providers/aws"
)

const (
	operationWaitTimeout    = 5 * time.Second
	operationPollInterval   = 100 * time.Millisecond
	maxRecordsReturnedByAPI = 100
	maxAsgNamesPerDescribe  = 50
	refreshInterval         = 1 * time.Minute
	autoDiscovererTypeASG   = "asg"
	asgAutoDiscovererKeyTag = "tag"
)

// AwsManager is handles aws communication and data caching.
type AwsManager struct {
	autoScalingService autoScalingWrapper
	ec2Service         ec2Wrapper
	asgCache           *asgCache
	lastRefresh        time.Time
	instanceTypes      map[string]*InstanceType
}

type asgTemplate struct {
	InstanceType *InstanceType
	Region       string
	Zone         string
	Tags         []*autoscaling.TagDescription
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

type awsSDKProvider struct {
	cfg *provider_aws.CloudConfig
}

func newAWSSDKProvider(cfg *provider_aws.CloudConfig) *awsSDKProvider {
	return &awsSDKProvider{
		cfg: cfg,
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

// createAwsManagerInternal allows for a customer autoScalingWrapper to be passed in by tests
//
// #1449 If running tests outside of AWS without AWS_REGION among environment
// variables, avoid a 5+ second EC2 Metadata lookup timeout in getRegion by
// setting and resetting AWS_REGION before calling createAWSManagerInternal:
//
//	defer resetAWSRegion(os.LookupEnv("AWS_REGION"))
//	os.Setenv("AWS_REGION", "fanghorn")
func createAWSManagerInternal(
	configReader io.Reader,
	discoveryOpts cloudprovider.NodeGroupDiscoveryOptions,
	autoScalingService *autoScalingWrapper,
	ec2Service *ec2Wrapper,
	instanceTypes map[string]*InstanceType,
) (*AwsManager, error) {

	cfg, err := readAWSCloudConfig(configReader)
	if err != nil {
		klog.Errorf("Couldn't read config: %v", err)
		return nil, err
	}

	if err = validateOverrides(cfg); err != nil {
		klog.Errorf("Unable to validate custom endpoint overrides: %v", err)
		return nil, err
	}

	if autoScalingService == nil || ec2Service == nil {
		awsSdkProvider := newAWSSDKProvider(cfg)
		sess, err := session.NewSession(aws.NewConfig().WithRegion(getRegion()).
			WithEndpointResolver(getResolver(awsSdkProvider.cfg)))
		if err != nil {
			return nil, err
		}

		if autoScalingService == nil {
			c := newLaunchConfigurationInstanceTypeCache()
			autoScalingService = &autoScalingWrapper{autoscaling.New(sess), c}
		}

		if ec2Service == nil {
			ec2Service = &ec2Wrapper{ec2.New(sess)}
		}
	}

	specs, err := parseASGAutoDiscoverySpecs(discoveryOpts)
	if err != nil {
		return nil, err
	}

	cache, err := newASGCache(*autoScalingService, discoveryOpts.NodeGroupSpecs, specs)
	if err != nil {
		return nil, err
	}

	manager := &AwsManager{
		autoScalingService: *autoScalingService,
		ec2Service:         *ec2Service,
		asgCache:           cache,
		instanceTypes:      instanceTypes,
	}

	if err := manager.forceRefresh(); err != nil {
		return nil, err
	}

	return manager, nil
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

// CreateAwsManager constructs awsManager object.
func CreateAwsManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, instanceTypes map[string]*InstanceType) (*AwsManager, error) {
	return createAWSManagerInternal(configReader, discoveryOpts, nil, nil, instanceTypes)
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (m *AwsManager) Refresh() error {
	if m.lastRefresh.Add(refreshInterval).After(time.Now()) {
		return nil
	}
	return m.forceRefresh()
}

func (m *AwsManager) forceRefresh() error {
	if err := m.asgCache.regenerate(); err != nil {
		klog.Errorf("Failed to regenerate ASG cache: %v", err)
		return err
	}
	m.lastRefresh = time.Now()
	klog.V(2).Infof("Refreshed ASG list, next refresh after %v", m.lastRefresh.Add(refreshInterval))
	return nil
}

// GetAsgForInstance returns AsgConfig of the given Instance
func (m *AwsManager) GetAsgForInstance(instance AwsInstanceRef) *asg {
	return m.asgCache.FindForInstance(instance)
}

// Cleanup the ASG cache.
func (m *AwsManager) Cleanup() {
	m.asgCache.Cleanup()
}

func (m *AwsManager) getAsgs() []*asg {
	return m.asgCache.Get()
}

// SetAsgSize sets ASG size.
func (m *AwsManager) SetAsgSize(asg *asg, size int) error {
	return m.asgCache.SetAsgSize(asg, size)
}

// DeleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (m *AwsManager) DeleteInstances(instances []*AwsInstanceRef) error {
	if err := m.asgCache.DeleteInstances(instances); err != nil {
		return err
	}
	klog.V(2).Infof("Some ASG instances might have been deleted, forcing ASG list refresh")
	return m.forceRefresh()
}

// GetAsgNodes returns Asg nodes.
func (m *AwsManager) GetAsgNodes(ref AwsRef) ([]AwsInstanceRef, error) {
	return m.asgCache.InstancesByAsg(ref)
}

func (m *AwsManager) getAsgTemplate(asg *asg) (*asgTemplate, error) {
	if len(asg.AvailabilityZones) < 1 {
		return nil, fmt.Errorf("unable to get first AvailabilityZone for ASG %q", asg.Name)
	}

	az := asg.AvailabilityZones[0]
	region := az[0 : len(az)-1]

	if len(asg.AvailabilityZones) > 1 {
		klog.Warningf("Found multiple availability zones for ASG %q; using %s\n", asg.Name, az)
	}

	instanceTypeName, err := m.buildInstanceType(asg)
	if err != nil {
		return nil, err
	}

	if t, ok := m.instanceTypes[instanceTypeName]; ok {
		return &asgTemplate{
			InstanceType: t,
			Region:       region,
			Zone:         az,
			Tags:         asg.Tags,
		}, nil
	}
	return nil, fmt.Errorf("ASG %q uses the unknown EC2 instance type %q", asg.Name, instanceTypeName)
}

func (m *AwsManager) buildInstanceType(asg *asg) (string, error) {
	if asg.LaunchConfigurationName != "" {
		return m.autoScalingService.getInstanceTypeByLCName(asg.LaunchConfigurationName)
	} else if asg.LaunchTemplate != nil {
		return m.ec2Service.getInstanceTypeByLT(asg.LaunchTemplate)
	} else if asg.MixedInstancesPolicy != nil {
		// always use first instance
		if len(asg.MixedInstancesPolicy.instanceTypesOverrides) != 0 {
			return asg.MixedInstancesPolicy.instanceTypesOverrides[0], nil
		}

		return m.ec2Service.getInstanceTypeByLT(asg.MixedInstancesPolicy.launchTemplate)
	}

	return "", errors.New("Unable to get instance type from launch config or launch template")
}

func (m *AwsManager) buildNodeFromTemplate(asg *asg, template *asgTemplate) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-asg-%d", asg.Name, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	// TODO: get a real value.
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(template.InstanceType.VCPU, resource.DecimalSI)
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(template.InstanceType.GPU, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.InstanceType.MemoryMb*1024*1024, resource.DecimalSI)

	resourcesFromTags := extractAllocatableResourcesFromAsg(template.Tags)
	for resourceName, val := range resourcesFromTags {
		node.Status.Capacity[apiv1.ResourceName(resourceName)] = *val
	}

	// TODO: use proper allocatable!!
	node.Status.Allocatable = node.Status.Capacity

	// NodeLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, extractLabelsFromAsg(template.Tags))
	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Spec.Taints = extractTaintsFromAsg(template.Tags)

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(template *asgTemplate, nodeName string) map[string]string {
	result := make(map[string]string)
	// TODO: extract it somehow
	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS

	result[apiv1.LabelInstanceType] = template.InstanceType.InstanceType

	result[apiv1.LabelZoneRegion] = template.Region
	result[apiv1.LabelZoneFailureDomain] = template.Zone
	result[apiv1.LabelHostname] = nodeName
	return result
}

func extractLabelsFromAsg(tags []*autoscaling.TagDescription) map[string]string {
	result := make(map[string]string)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/label/")
		if len(splits) > 1 {
			label := splits[1]
			if label != "" {
				result[label] = v
			}
		}
	}

	return result
}

func extractAllocatableResourcesFromAsg(tags []*autoscaling.TagDescription) map[string]*resource.Quantity {
	result := make(map[string]*resource.Quantity)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/resources/")
		if len(splits) > 1 {
			label := splits[1]
			if label != "" {
				quantity, err := resource.ParseQuantity(v)
				if err != nil {
					continue
				}
				result[label] = &quantity
			}
		}
	}

	return result
}

func extractTaintsFromAsg(tags []*autoscaling.TagDescription) []apiv1.Taint {
	taints := make([]apiv1.Taint, 0)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		// The tag value must be in the format <tag>:NoSchedule
		r, _ := regexp.Compile("(.*):(?:NoSchedule|NoExecute|PreferNoSchedule)")
		if r.MatchString(v) {
			splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/taint/")
			if len(splits) > 1 {
				values := strings.SplitN(v, ":", 2)
				if len(values) > 1 {
					taints = append(taints, apiv1.Taint{
						Key:    splits[1],
						Value:  values[0],
						Effect: apiv1.TaintEffect(values[1]),
					})
				}
			}
		}
	}
	return taints
}

// An asgAutoDiscoveryConfig specifies how to autodiscover AWS ASGs.
type asgAutoDiscoveryConfig struct {
	// Tags to match on.
	// Any ASG with all of the provided tag keys will be autoscaled.
	Tags map[string]string
}

// ParseASGAutoDiscoverySpecs returns any provided NodeGroupAutoDiscoverySpecs
// parsed into configuration appropriate for ASG autodiscovery.
func parseASGAutoDiscoverySpecs(o cloudprovider.NodeGroupDiscoveryOptions) ([]asgAutoDiscoveryConfig, error) {
	cfgs := make([]asgAutoDiscoveryConfig, len(o.NodeGroupAutoDiscoverySpecs))
	var err error
	for i, spec := range o.NodeGroupAutoDiscoverySpecs {
		cfgs[i], err = parseASGAutoDiscoverySpec(spec)
		if err != nil {
			return nil, err
		}
	}
	return cfgs, nil
}

func parseASGAutoDiscoverySpec(spec string) (asgAutoDiscoveryConfig, error) {
	cfg := asgAutoDiscoveryConfig{}

	tokens := strings.Split(spec, ":")
	if len(tokens) != 2 {
		return cfg, fmt.Errorf("invalid node group auto discovery spec specified via --node-group-auto-discovery: %s", spec)
	}
	discoverer := tokens[0]
	if discoverer != autoDiscovererTypeASG {
		return cfg, fmt.Errorf("unsupported discoverer specified: %s", discoverer)
	}
	param := tokens[1]
	kv := strings.SplitN(param, "=", 2)
	if len(kv) != 2 {
		return cfg, fmt.Errorf("invalid key=value pair %s", kv)
	}
	k, v := kv[0], kv[1]
	if k != asgAutoDiscovererKeyTag {
		return cfg, fmt.Errorf("unsupported parameter key \"%s\" is specified for discoverer \"%s\". The only supported key is \"%s\"", k, discoverer, asgAutoDiscovererKeyTag)
	}
	if v == "" {
		return cfg, errors.New("tag value not supplied")
	}
	p := strings.Split(v, ",")
	if len(p) == 0 {
		return cfg, fmt.Errorf("invalid ASG tag for auto discovery specified: ASG tag must not be empty")
	}
	cfg.Tags = make(map[string]string, len(p))
	for _, label := range p {
		lp := strings.SplitN(label, "=", 2)
		if len(lp) > 1 {
			cfg.Tags[lp[0]] = lp[1]
			continue
		}
		cfg.Tags[lp[0]] = ""
	}
	return cfg, nil
}
