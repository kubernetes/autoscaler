/*
Copyright 2018 The Kubernetes Authors.

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

package baiducloud

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/bce"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/baiducloud/baiducloud-sdk-go/cce"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
)

const (

	// CceUserAgent is a prefix of the http header UserAgent.
	CceUserAgent = "cce-k8s:"

	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// BaiducloudManager handles baiducloud communication and data caching.
type BaiducloudManager struct {
	cloudConfig *CloudConfig
	cceClient   *cce.Client

	asgs      *autoScalingGroups
	interrupt chan struct{}
}

type asgInformation struct {
	config   *Asg
	basename string
}

type asgTemplate struct {
	InstanceType     int
	Region           string
	Zone             string
	CPU              int
	Memory           int
	GpuCount         int
	EphemeralStorage int
	Tags             map[string]string
}

// CreateBaiducloudManager constructs baiducloudManager object.
func CreateBaiducloudManager(configReader io.Reader) (*BaiducloudManager, error) {
	cfg := &CloudConfig{}
	if configReader != nil {
		configContents, err := ioutil.ReadAll(configReader)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(configContents, cfg)
		if err != nil {
			return nil, err
		}
	}
	err := cfg.validate()
	if err != nil {
		return nil, err
	}

	bceConfig := bce.NewConfig(bce.NewCredentials(cfg.AccessKeyID, cfg.SecretAccessKey))
	bceConfig.Region = cfg.Region
	bceConfig.Timeout = 20 * time.Second
	bceConfig.Endpoint = cfg.Endpoint + "/internal-api"
	bceConfig.UserAgent = CceUserAgent + cfg.ClusterID
	cceClient := cce.NewClient(cce.NewConfig(bceConfig))
	cceClient.SetDebug(true)
	manager := &BaiducloudManager{
		cloudConfig: cfg,
		cceClient:   cceClient,
		asgs:        newAutoScalingGroups(cfg, cceClient),
		interrupt:   make(chan struct{}),
	}

	go wait.Until(func() {
		manager.asgs.cacheMutex.Lock()
		defer manager.asgs.cacheMutex.Unlock()
		if err := manager.asgs.regenerateCache(); err != nil {
			klog.Errorf("Error while regenerating cache: %v\n", err)
		}
	}, time.Hour, manager.interrupt)

	return manager, nil
}

// RegisterAsg registers asg in the BCE Manager.
func (m *BaiducloudManager) RegisterAsg(asg *Asg) {
	m.asgs.Register(asg)
}

// GetAsgForInstance returns AsgConfig.
func (m *BaiducloudManager) GetAsgForInstance(instanceID string) (*Asg, error) {
	return m.asgs.FindForInstance(instanceID)
}

// GetAsgSize gets asg's size.
func (m *BaiducloudManager) GetAsgSize(asg *Asg) (int64, error) {
	instanceList, err := m.cceClient.GetAsgNodes(asg.Name, m.cloudConfig.ClusterID)
	if err != nil {
		return -1, err
	}
	size := int64(0)
	for _, instance := range instanceList {
		klog.V(4).Infof("Group: %s instances status: %v \n", asg.Name, instance.Status)
		if instance.Status == cce.InstanceStatusRunning || instance.Status == cce.InstanceStatusCreating || instance.Status == "" {
			size++
		}
	}
	printNodeStatusCount(instanceList, asg.Name, size)
	return size, nil
}

func printNodeStatusCount(instanceList []cce.CceInstance, asgName string, size int64) {
	runningCount := 0
	creatingCount := 0
	deletingCount := 0
	exception := 0
	for _, instance := range instanceList {
		switch instance.Status {
		case "RUNNING":
			runningCount++
		case "CREATING":
			creatingCount++
		case "DELETING":
			deletingCount++
		default:
			exception++
			klog.V(4).Infof("EXCEPTION instance: %v, status: %v", instance.InstanceId, instance.Status)
		}
	}
	klog.V(4).Infof("Group: %s TotalSize: %d, CA Size: %d ,instances RUNNING: %v, CREATING: %v, DELETING: %v, EXCEPTION: %v",
		asgName, len(instanceList), size, runningCount, creatingCount, deletingCount, exception)
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// ScaleUpCluster scales up cluster.
func (m *BaiducloudManager) ScaleUpCluster(delta int, groupID string) error {
	var args *cce.ScaleUpClusterWithGroupIDArgs
	args = &cce.ScaleUpClusterWithGroupIDArgs{
		ClusterID: m.cloudConfig.ClusterID,
		Num:       delta,
		GroupID:   groupID,
	}
	err := m.cceClient.ScaleUpClusterWithGroupID(args)
	if err != nil {
		return fmt.Errorf("[bce] ScaleUpCluster error: %v", err)
	}

	return nil
}

// ScaleDownCluster decreases nodes belonging to cluster.
func (m *BaiducloudManager) ScaleDownCluster(instances []string) error {
	klog.V(4).Infof("scaleDownCluster: %v\n", instances)
	if len(instances) == 0 {
		return nil
	}
	nodeinfos := make([]cce.NodeInfo, len(instances))
	for _, id := range instances {
		info := cce.NodeInfo{
			InstanceID: id,
		}
		nodeinfos = append(nodeinfos, info)
	}
	scaledownArg := &cce.ScaleDownClusterArgs{
		ClusterID: m.cloudConfig.ClusterID,
		NodeInfos: nodeinfos,
	}
	return m.cceClient.ScaleDownCluster(scaledownArg)
}

// GetAsgNodes returns Asg nodes.
func (m *BaiducloudManager) GetAsgNodes(asg *Asg) ([]string, error) {
	result := make([]string, 0)
	instanceList, err := m.cceClient.GetAsgNodes(asg.Name, m.cloudConfig.ClusterID)
	if err != nil {
		return []string{}, err
	}
	for _, instance := range instanceList {
		result = append(result, fmt.Sprintf("cce://%s", instance.InstanceId))
	}
	klog.V(5).Infof("GetAsgNodes: %s\n", result)
	return result, nil
}

func (m *BaiducloudManager) getAsgTemplate(name string) (*asgTemplate, error) {
	cceGroup, err := m.cceClient.DescribeGroup(name, m.cloudConfig.ClusterID)
	if err != nil {
		klog.V(4).Infof("describeCluster err: %s\n", err)
		return nil, err
	}

	tags := make(map[string]string)
	for _, tag := range cceGroup.Tags {
		tags[tag.Key] = tag.Value
	}

	return &asgTemplate{
		InstanceType:     cceGroup.InstanceType,
		Region:           m.cloudConfig.Region,
		CPU:              cceGroup.CPU,
		Memory:           cceGroup.Memory,
		GpuCount:         cceGroup.GpuCount,
		EphemeralStorage: cceGroup.EphemeralStorage,
		Tags:             tags,
	}, nil
}

func (m *BaiducloudManager) buildNodeFromTemplate(asg *Asg, template *asgTemplate) (*apiv1.Node, error) {
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
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(int64(template.CPU), resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(int64(template.Memory*1024*1024*1024), resource.DecimalSI)

	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template))

	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(int64(template.GpuCount), resource.DecimalSI)

	// add ephemeral-storage
	node.Status.Capacity[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(int64(template.EphemeralStorage*1024*1024*1024), resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func buildGenericLabels(template *asgTemplate) map[string]string {
	result := make(map[string]string)
	// append custom node labels
	for key, value := range template.Tags {
		result[key] = value
	}
	return result
}
