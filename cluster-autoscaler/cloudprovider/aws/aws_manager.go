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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	provider_aws "k8s.io/kubernetes/pkg/cloudprovider/providers/aws"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
)

const (
	operationWaitTimeout    = 5 * time.Second
	operationPollInterval   = 100 * time.Millisecond
	maxRecordsReturnedByAPI = 100
	maxAsgNamesPerDescribe  = 50
	refreshInterval         = 10 * time.Second
)

// AwsManager is handles aws communication and data caching.
type AwsManager struct {
	autoScalingService autoScalingWrapper
	ec2Service         ec2Wrapper
	asgCache           *asgCache
	lastRefresh        time.Time
}

type asgTemplate struct {
	InstanceType *instanceType
	Region       string
	Zone         string
	Tags         []*autoscaling.TagDescription
}

// getRegion deduces the current AWS Region.
func getRegion(cfg ...*aws.Config) string {
	region, present := os.LookupEnv("AWS_REGION")
	if !present {
		svc := ec2metadata.New(session.New(), cfg...)
		if r, err := svc.Region(); err == nil {
			region = r
		}
	}
	return region
}

// createAwsManagerInternal allows for a customer autoScalingWrapper to be passed in by tests
func createAWSManagerInternal(
	configReader io.Reader,
	discoveryOpts cloudprovider.NodeGroupDiscoveryOptions,
	autoScalingService *autoScalingWrapper,
	ec2Service *ec2Wrapper,
) (*AwsManager, error) {
	if configReader != nil {
		var cfg provider_aws.CloudConfig
		if err := gcfg.ReadInto(&cfg, configReader); err != nil {
			glog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
	}

	if autoScalingService == nil || ec2Service == nil {
		sess := session.New(aws.NewConfig().WithRegion(getRegion()))

		if autoScalingService == nil {
			autoScalingService = &autoScalingWrapper{autoscaling.New(sess)}
		}

		if ec2Service == nil {
			ec2Service = &ec2Wrapper{ec2.New(sess)}
		}
	}

	specs, err := discoveryOpts.ParseASGAutoDiscoverySpecs()
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
	}

	if err := manager.forceRefresh(); err != nil {
		return nil, err
	}

	return manager, nil
}

// CreateAwsManager constructs awsManager object.
func CreateAwsManager(configReader io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*AwsManager, error) {
	return createAWSManagerInternal(configReader, discoveryOpts, nil, nil)
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
		glog.Errorf("Failed to regenerate ASG cache: %v", err)
		return err
	}
	m.lastRefresh = time.Now()
	glog.V(2).Infof("Refreshed ASG list, next refresh after %v", m.lastRefresh.Add(refreshInterval))
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
	return m.asgCache.DeleteInstances(instances)
}

// GetAsgNodes returns Asg nodes.
func (m *AwsManager) GetAsgNodes(ref AwsRef) ([]AwsInstanceRef, error) {
	return m.asgCache.InstancesByAsg(ref)
}

func (m *AwsManager) getAsgTemplate(asg *asg) (*asgTemplate, error) {
	if len(asg.AvailabilityZones) < 1 {
		return nil, fmt.Errorf("Unable to get first AvailabilityZone for ASG %q", asg.Name)
	}

	az := asg.AvailabilityZones[0]
	region := az[0 : len(az)-1]

	if len(asg.AvailabilityZones) > 1 {
		glog.Warningf("Found multiple availability zones for ASG %q; using %s\n", asg.Name, az)
	}

	instanceTypeName, err := m.buildInstanceType(asg)
	if err != nil {
		return nil, err
	}

	if t, ok := InstanceTypes[instanceTypeName]; ok {
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
	} else if asg.LaunchTemplateName != "" && asg.LaunchTemplateVersion != "" {
		return m.ec2Service.getInstanceTypeByLT(asg.LaunchTemplateName, asg.LaunchTemplateVersion)
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

	result[kubeletapis.LabelInstanceType] = template.InstanceType.InstanceType

	result[kubeletapis.LabelZoneRegion] = template.Region
	result[kubeletapis.LabelZoneFailureDomain] = template.Zone
	result[kubeletapis.LabelHostname] = nodeName
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
