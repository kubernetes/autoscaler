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

package spotinst

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/log"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/stringutil"
	"gopkg.in/gcfg.v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
)

// CloudManager holds the config and client.
type CloudManager struct {
	groupService    elastigroup.Service
	groups          []*Group
	refreshedAt     time.Time
	refreshInterval time.Duration
	interruptCh     chan struct{}
	cacheMu         sync.Mutex
	cache           map[string]*Group // k: InstanceID, v: Group
}

// CloudConfig holds the configuration parsed from the --cloud-config flag.
// All fields are required unless otherwise specified.
type CloudConfig struct {
	Global struct{}
}

// NewCloudManager constructs manager object.
func NewCloudManager(config io.Reader, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions) (*CloudManager, error) {
	glog.Info("Building Spotinst cloud manager")

	cfg, err := readCloudConfig(config)
	if err != nil {
		return nil, err
	}

	svc, err := newService(cfg)
	if err != nil {
		return nil, err
	}

	manager := &CloudManager{
		groupService:    svc,
		refreshInterval: time.Minute,
		interruptCh:     make(chan struct{}),
		groups:          make([]*Group, 0),
		cache:           make(map[string]*Group),
	}

	if err := manager.addNodeGroups(discoveryOpts.NodeGroupSpecs); err != nil {
		return nil, err
	}

	go wait.Until(func() {
		if err := manager.forceRefresh(); err != nil {
			glog.Errorf("Error while refreshing cache: %v", err)
		}
	}, time.Hour, manager.interruptCh)

	return manager, nil
}

// newService returns a new instance of Spotinst Service.
func newService(cloudConfig *CloudConfig) (elastigroup.Service, error) {
	// Create a new config.
	config := spotinst.DefaultConfig()
	config.WithLogger(newStdLogger())
	config.WithUserAgent("Kubernetes-ClusterAutoscaler")

	// Create a new session.
	sess := session.New(config)

	// Create a new service.
	svc := elastigroup.New(sess)

	return svc, nil
}

func newStdLogger() log.Logger {
	return log.LoggerFunc(func(format string, args ...interface{}) {
		glog.V(4).Infof(format, args...)
	})
}

// readCloudConfig reads an instance of Config from config reader.
func readCloudConfig(config io.Reader) (*CloudConfig, error) {
	var cfg CloudConfig

	if config != nil {
		if err := gcfg.ReadInto(&cfg, config); err != nil {
			return nil, fmt.Errorf("couldn't read Spotinst config: %v", err)
		}
	}

	return &cfg, nil
}

func (mgr *CloudManager) addNodeGroups(specs []string) error {
	glog.Info("Attempting to add node groups")

	for _, spec := range specs {
		if err := mgr.addNodeGroup(spec); err != nil {
			return fmt.Errorf("could not register group with spec %s: %v", spec, err)
		}
	}

	return nil
}

func (mgr *CloudManager) addNodeGroup(spec string) error {
	glog.Infof("Attempting to add node group: %s", spec)

	group, err := mgr.buildGroupFromSpec(spec)
	if err != nil {
		return fmt.Errorf("could not parse spec for node group: %v", err)
	}
	mgr.RegisterGroup(group)

	glog.Infof("Node group added: %s", group.groupID)
	return nil
}

func (mgr *CloudManager) buildGroupFromSpec(value string) (*Group, error) {
	spec, err := dynamic.SpecFromString(value, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}
	group := &Group{
		manager: mgr,
		groupID: spec.Name,
		minSize: spec.MinSize,
		maxSize: spec.MaxSize,
	}
	return group, nil
}

// RegisterGroup registers a resource group in Spotinst Manager.
func (mgr *CloudManager) RegisterGroup(grp *Group) error {
	mgr.cacheMu.Lock()
	defer mgr.cacheMu.Unlock()

	group, err := mgr.getResourceForGroup(grp.Id())
	if err != nil {
		return err
	}
	grp.group = group

	mgr.groups = append(mgr.groups, grp)
	return nil
}

// GetGroupSize gets the current size of the group.
func (mgr *CloudManager) GetGroupSize(grp *Group) (int64, error) {
	group, err := mgr.getResourceForGroup(grp.Id())
	if err != nil {
		return -1, err
	}
	size := spotinst.IntValue(group.Capacity.Target)
	return int64(size), nil
}

// SetGroupSize sets the instances count in a Group by updating a
// predefined Spotinst stack parameter (specified by the user).
func (mgr *CloudManager) SetGroupSize(grp *Group, size int64) error {
	in := &aws.UpdateGroupInput{
		Group: &aws.Group{
			ID: spotinst.String(grp.Id()),
			Capacity: &aws.Capacity{
				Target:  spotinst.Int(int(size)),
				Minimum: spotinst.Int(grp.minSize),
				Maximum: spotinst.Int(grp.maxSize),
			},
		},
	}
	_, err := mgr.groupService.CloudProviderAWS().Update(context.Background(), in)
	if err != nil {
		return err
	}
	return nil
}

// GetGroupForInstance retrieves the resource group that contains
// a given instance.
func (mgr *CloudManager) GetGroupForInstance(instanceID string) (*Group, error) {
	mgr.cacheMu.Lock()
	defer mgr.cacheMu.Unlock()

	if group, ok := mgr.cache[instanceID]; ok {
		return group, nil
	}

	if err := mgr.forceRefresh(); err != nil {
		return nil, err
	}

	if group, ok := mgr.cache[instanceID]; ok {
		return group, nil
	}

	glog.Warningf("Instance `%s` does not belong to any managed group", instanceID)
	return nil, nil
}

// DeleteInstances deletes the specified instances from the
// OpenStack resource group
func (mgr *CloudManager) DeleteInstances(instanceIDs []string) error {
	if len(instanceIDs) == 0 {
		return nil
	}
	commonGroup, err := mgr.GetGroupForInstance(instanceIDs[0])
	if err != nil {
		return err
	}
	for _, instanceID := range instanceIDs {
		instanceGroup, err := mgr.GetGroupForInstance(instanceID)
		if err != nil {
			return err
		}
		if instanceGroup.groupID != commonGroup.groupID {
			return errors.New("connot delete instances which don't belong to the same group")
		}
	}
	in := &aws.DetachGroupInput{
		GroupID:                       spotinst.String(commonGroup.groupID),
		InstanceIDs:                   instanceIDs,
		ShouldDecrementTargetCapacity: spotinst.Bool(true),
		ShouldTerminateInstances:      spotinst.Bool(true),
	}
	if _, err := mgr.groupService.CloudProviderAWS().Detach(context.Background(), in); err != nil {
		return fmt.Errorf("failed to detach instances from group %s: %v", commonGroup.groupID, err)
	}
	return nil
}

func (mgr *CloudManager) getResourceForGroup(groupID string) (*aws.Group, error) {
	in := &aws.ReadGroupInput{
		GroupID: spotinst.String(groupID),
	}
	out, err := mgr.groupService.CloudProviderAWS().Read(context.Background(), in)
	if err != nil {
		return nil, err
	}
	if out.Group == nil {
		return nil, fmt.Errorf("failed to get group %s", groupID)
	}
	return out.Group, nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (mgr *CloudManager) Cleanup() error {
	close(mgr.interruptCh)
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (mgr *CloudManager) Refresh() error {
	if mgr.refreshedAt.Add(mgr.refreshInterval).After(time.Now()) {
		return nil
	}
	return mgr.forceRefresh()
}

func (mgr *CloudManager) forceRefresh() error {
	mgr.regenerateCache()
	mgr.refreshedAt = time.Now()
	glog.V(2).Infof("Refreshed, next refresh after %v", mgr.refreshedAt.Add(mgr.refreshInterval))
	return nil
}

func (mgr *CloudManager) regenerateCache() {
	mgr.cacheMu.Lock()
	defer mgr.cacheMu.Unlock()

	mgr.cache = make(map[string]*Group)
	for _, group := range mgr.groups {
		glog.V(4).Infof("Regenerating resource group information for %s", group.groupID)
		if err := mgr.refreshGroupNodes(group); err != nil {
			glog.Warningf("Could not retrieve nodes for group %s: %v", group.groupID, err)
		}
	}
}

func (mgr *CloudManager) refreshGroupNodes(grp *Group) error {
	in := &aws.StatusGroupInput{
		GroupID: spotinst.String(grp.Id()),
	}
	status, err := mgr.groupService.CloudProviderAWS().Status(context.Background(), in)
	if err != nil {
		return err
	}
	for _, instance := range status.Instances {
		if instance.ID != nil {
			instanceID := spotinst.StringValue(instance.ID)
			glog.Infof("Managing AWS instance with ID %s in group %s", instanceID, grp.Id())
			mgr.cache[instanceID] = grp
		}
	}
	return nil
}

type groupTemplate struct {
	InstanceType *instanceType
	Region       string
	Zone         string
	Tags         []*aws.Tag
}

func (mgr *CloudManager) buildGroupTemplate(groupID string) (*groupTemplate, error) {
	glog.Infof("Building template for group %s", groupID)

	group, err := mgr.getResourceForGroup(groupID)
	if err != nil {
		return nil, err
	}

	if len(group.Compute.AvailabilityZones) < 1 {
		return nil, fmt.Errorf("unable to get first AvailabilityZone for %s", groupID)
	}

	zone := spotinst.StringValue(group.Compute.AvailabilityZones[0].Name)
	region := zone[0 : len(zone)-1]

	if len(group.Compute.AvailabilityZones) > 1 {
		glog.Warningf("Found multiple availability zones, using %s", zone)
	}

	tmpl := &groupTemplate{
		InstanceType: InstanceTypes[spotinst.StringValue(group.Compute.InstanceTypes.OnDemand)],
		Region:       region,
		Zone:         zone,
		Tags:         group.Compute.LaunchSpecification.Tags,
	}

	return tmpl, nil
}

func (mgr *CloudManager) buildNodeFromTemplate(group *Group, template *groupTemplate) (*apiv1.Node, error) {
	glog.Infof("Building node from template of group %s", group.Id())

	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-group-%d", group.groupID, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}

	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(template.InstanceType.VCPU, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(template.InstanceType.MemoryMb*1024*1024, resource.DecimalSI)
	node.Status.Allocatable = node.Status.Capacity

	// NodeLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, extractLabelsFromGroup(template.Tags))

	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Spec.Taints = extractTaintsFromGroup(template.Tags)
	node.Status.Conditions = cloudprovider.BuildReadyConditions()

	glog.V(4).Infof("Node `%s` labels: %s", nodeName, stringutil.Stringify(node.Labels))
	glog.V(4).Infof("Node `%s` taints: %s", nodeName, stringutil.Stringify(node.Spec.Taints))

	return &node, nil
}

func buildGenericLabels(template *groupTemplate, nodeName string) map[string]string {
	result := make(map[string]string)

	result[kubeletapis.LabelArch] = cloudprovider.DefaultArch
	result[kubeletapis.LabelOS] = cloudprovider.DefaultOS
	result[kubeletapis.LabelInstanceType] = template.InstanceType.InstanceType
	result[kubeletapis.LabelZoneRegion] = template.Region
	result[kubeletapis.LabelZoneFailureDomain] = template.Zone
	result[kubeletapis.LabelHostname] = nodeName

	return result
}

func extractLabelsFromGroup(tags []*aws.Tag) map[string]string {
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

func extractTaintsFromGroup(tags []*aws.Tag) []apiv1.Taint {
	taints := make([]apiv1.Taint, 0)

	for _, tag := range tags {
		k := *tag.Key
		v := *tag.Value
		splits := strings.Split(k, "k8s.io/cluster-autoscaler/node-template/taint/")
		if len(splits) > 1 {
			values := strings.SplitN(v, ":", 2)
			taints = append(taints, apiv1.Taint{
				Key:    splits[1],
				Value:  values[0],
				Effect: apiv1.TaintEffect(values[1]),
			})
		}
	}

	return taints
}
