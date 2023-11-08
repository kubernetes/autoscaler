/*
Copyright 2021 The Kubernetes Authors.

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

package tencentcloud

import (
	"context"
	"fmt"

	gerrors "github.com/pkg/errors"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/client"
	as "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/as/v20180419"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/common"
	cvm "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/cvm/v20170312"
	vpc "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/vpc/v20170312"
)

// CloudService is used for communicating with Tencentcloud API.
type CloudService interface {
	// FetchAsgInstances returns instances of the specified ASG.
	FetchAsgInstances(TcRef) ([]cloudprovider.Instance, error)
	// DeleteInstances remove instances of specified ASG.
	DeleteInstances(Asg, []string) error
	// GetAsgRefByInstanceRef returns asgRef according to instanceRef
	GetAsgRefByInstanceRef(TcRef) (*TcRef, error)
	// GetAutoScalingGroups queries and returns a set of ASG.
	GetAutoScalingGroups([]string) ([]as.AutoScalingGroup, error)
	// GetAutoscalingConfigs queries and returns a set of ASG launchconfiguration.
	GetAutoscalingConfigs([]string) ([]as.LaunchConfiguration, error)
	// GetAutoScalingGroups queries and returns a set of ASG.
	GetAutoScalingGroup(TcRef) (*as.AutoScalingGroup, error)
	// ResizeAsg set the target size of ASG.
	ResizeAsg(TcRef, uint64) error
	// GetAutoScalingInstances returns instances of specific ASG.
	GetAutoScalingInstances(TcRef) ([]*as.Instance, error)
	// GetTencentcloudInstanceRef returns a Tencentcloud ref.
	GetTencentcloudInstanceRef(*as.Instance) (*TcRef, error)
	// GetInstanceInfoByType queries the number of CPU, memory, and GPU resources of the model configured for generating template
	GetInstanceInfoByType(string) (*InstanceInfo, error)
	// GetZoneBySubnetID return zone by subnetID.
	GetZoneBySubnetID(string) (string, error)
	// GetZoneInfo invokes cvm.DescribeZones to query zone information.
	GetZoneInfo(string) (*cvm.ZoneInfo, error)
}

// CloudServiceImpl provides several utility methods over the auto-scaling cloudService provided by Tencentcloud SDK
type CloudServiceImpl struct {
	asClient, cvmClient, vpcClient client.Client
}

const (
	maxRecordsReturnedByAPI = 100
)

var zoneInfos = make(map[string]*cvm.ZoneInfo)

// SubnetInfo represents subnet's detail
type SubnetInfo struct {
	SubnetID string
	Zone     string
	ZoneID   int
}

// NewCloudService creates an instance of caching CloudServiceImpl
func NewCloudService(cvmClient, vpcClient, asClient client.Client) CloudService {
	return &CloudServiceImpl{
		cvmClient: cvmClient,
		vpcClient: vpcClient,
		asClient:  asClient,
	}
}

// FetchAsgInstances returns instances of the specified ASG.
func (ts *CloudServiceImpl) FetchAsgInstances(asgRef TcRef) ([]cloudprovider.Instance, error) {
	tencentcloudInstances, err := ts.GetAutoScalingInstances(asgRef)
	if err != nil {
		klog.V(4).Infof("Failed ASG info request for %s %s: %v", asgRef.Zone, asgRef.ID, err)
		return nil, err
	}
	infos := []cloudprovider.Instance{}
	for _, instance := range tencentcloudInstances {
		ref, err := ts.GetTencentcloudInstanceRef(instance)
		if err != nil {
			return nil, err
		}
		infos = append(infos, cloudprovider.Instance{
			Id: ref.ToProviderID(),
		})
	}
	return infos, nil
}

// DeleteInstances remove instances of specified ASG.
func (ts *CloudServiceImpl) DeleteInstances(asg Asg, instances []string) error {
	if ts.asClient == nil {
		return fmt.Errorf("asClient is not initialized")
	}

	req := as.NewRemoveInstancesRequest()
	req.AutoScalingGroupId = common.StringPtr(asg.Id())
	req.InstanceIds = common.StringPtrs(instances)
	res := as.NewRemoveInstancesResponse()
	err := ts.asClient.Send(context.TODO(), req, res)
	if err != nil {
		return gerrors.Wrap(err, "[CloudAPIError]")
	}
	if res == nil || res.Response == nil ||
		res.Response.ActivityId == nil ||
		res.Response.RequestId == nil {
		return fmt.Errorf("[InvalidResponse] %s:%s", req.GetService(), req.GetAction())
	}
	klog.V(4).Infof("Remove instances %v, asaID: %s, requestID: %s", instances, *res.Response.ActivityId, *res.Response.RequestId)

	return nil
}

// GetAutoscalingConfigs queries and returns a set of ASG launchconfiguration.
func (ts *CloudServiceImpl) GetAutoscalingConfigs(ascs []string) ([]as.LaunchConfiguration, error) {
	if ts.asClient == nil {
		return nil, fmt.Errorf("asClient is not initialized")
	}

	if len(ascs) > maxRecordsReturnedByAPI {
		klog.Warning("The number of Launch Configuration IDs exceeds 100: ", len(ascs))
	}

	// 查询AS，启动配置对应机型
	req := as.NewDescribeLaunchConfigurationsRequest()
	req.LaunchConfigurationIds = common.StringPtrs(ascs)
	req.Limit = common.Uint64Ptr(maxRecordsReturnedByAPI)
	resp := as.NewDescribeLaunchConfigurationsResponse()
	err := ts.asClient.Send(context.TODO(), req, resp)
	if err != nil {
		return nil, gerrors.Wrap(err, "[CloudAPIError]")
	}

	if resp == nil || resp.Response == nil || resp.Response.LaunchConfigurationSet == nil {
		return nil, fmt.Errorf("[InvalidResponse] %s:%s", req.GetService(), req.GetAction())
	}

	res := make([]as.LaunchConfiguration, 0)
	for _, lc := range resp.Response.LaunchConfigurationSet {
		if lc != nil {
			res = append(res, *lc)
		}
	}

	return res, nil
}

// InstanceInfo represents CVM's detail
type InstanceInfo struct {
	CPU            int64
	Memory         int64
	GPU            int64
	InstanceFamily string
	InstanceType   string
}

// GetInstanceInfoByType queries the number of CPU, memory, and GPU resources of the model configured for generating template
func (ts *CloudServiceImpl) GetInstanceInfoByType(instanceType string) (*InstanceInfo, error) {
	if ts.cvmClient == nil {
		return nil, fmt.Errorf("cvmClient is not initialized")
	}

	// DescribeZoneInstanceConfigInfos
	req := cvm.NewDescribeInstanceTypeConfigsRequest()
	req.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("instance-type"),
			Values: []*string{&instanceType},
		},
	}
	res := cvm.NewDescribeInstanceTypeConfigsResponse()
	err := ts.cvmClient.Send(context.TODO(), req, res)
	if err != nil {
		return nil, gerrors.Wrap(err, "[CloudAPIError]")
	}

	if res == nil ||
		res.Response == nil ||
		res.Response.InstanceTypeConfigSet == nil ||
		len(res.Response.InstanceTypeConfigSet) < 1 ||
		res.Response.InstanceTypeConfigSet[0].CPU == nil ||
		res.Response.InstanceTypeConfigSet[0].Memory == nil ||
		res.Response.InstanceTypeConfigSet[0].GPU == nil ||
		res.Response.InstanceTypeConfigSet[0].InstanceFamily == nil ||
		res.Response.InstanceTypeConfigSet[0].InstanceType == nil ||
		res.Response.InstanceTypeConfigSet[0].Zone == nil {
		return nil, fmt.Errorf("[InvalidResponse] %s:%s", req.GetService(), req.GetAction())
	}

	return &InstanceInfo{
		*res.Response.InstanceTypeConfigSet[0].CPU,
		*res.Response.InstanceTypeConfigSet[0].Memory,
		*res.Response.InstanceTypeConfigSet[0].GPU,
		*res.Response.InstanceTypeConfigSet[0].InstanceFamily,
		*res.Response.InstanceTypeConfigSet[0].InstanceType,
	}, nil
}

// GetAutoScalingGroups queries and returns a set of ASG.
func (ts *CloudServiceImpl) GetAutoScalingGroups(asgIds []string) ([]as.AutoScalingGroup, error) {
	if ts.asClient == nil {
		return nil, fmt.Errorf("asClient is not initialized")
	}

	if len(asgIds) > maxRecordsReturnedByAPI {
		klog.Warning("The number of ASG IDs exceeds 100: ", len(asgIds))
	}

	req := as.NewDescribeAutoScalingGroupsRequest()
	req.AutoScalingGroupIds = common.StringPtrs(asgIds)
	req.Limit = common.Uint64Ptr(maxRecordsReturnedByAPI)

	res := as.NewDescribeAutoScalingGroupsResponse()
	err := ts.asClient.Send(context.TODO(), req, res)
	if err != nil {
		return nil, gerrors.Wrap(err, "[CloudAPIError]")
	}

	if res == nil || res.Response == nil || res.Response.AutoScalingGroupSet == nil {
		return nil, fmt.Errorf("[InvalidResponse] %s:%s", req.GetService(), req.GetAction())
	}

	asgs := make([]as.AutoScalingGroup, 0)
	for _, asg := range res.Response.AutoScalingGroupSet {
		if asg != nil {
			asgs = append(asgs, *asg)
		}
	}

	return asgs, nil
}

// GetAutoScalingGroup returns the specific ASG.
func (ts *CloudServiceImpl) GetAutoScalingGroup(asgRef TcRef) (*as.AutoScalingGroup, error) {
	if ts.asClient == nil {
		return nil, fmt.Errorf("asClient is not initialized")
	}

	req := as.NewDescribeAutoScalingGroupsRequest()
	req.AutoScalingGroupIds = []*string{&asgRef.ID}
	res := as.NewDescribeAutoScalingGroupsResponse()
	err := ts.asClient.Send(context.TODO(), req, res)
	if err != nil {
		return nil, gerrors.Wrap(err, "[CloudAPIError]")
	}

	if res == nil || res.Response == nil || res.Response.AutoScalingGroupSet == nil ||
		len(res.Response.AutoScalingGroupSet) != 1 || res.Response.AutoScalingGroupSet[0] == nil {
		return nil, fmt.Errorf("as:DescribeAutoScalingGroups returned a invalid response: %s", res.ToJsonString())
	}

	return res.Response.AutoScalingGroupSet[0], nil
}

// GetAutoScalingInstances returns instances of specific ASG.
func (ts *CloudServiceImpl) GetAutoScalingInstances(asgRef TcRef) ([]*as.Instance, error) {
	if ts.asClient == nil {
		return nil, fmt.Errorf("asClient is not initialized")
	}

	req := as.NewDescribeAutoScalingInstancesRequest()
	filter := as.Filter{
		Name:   common.StringPtr("auto-scaling-group-id"),
		Values: common.StringPtrs([]string{asgRef.ID}),
	}
	req.Filters = []*as.Filter{&filter}
	req.Limit = common.Int64Ptr(maxRecordsReturnedByAPI)
	req.Offset = common.Int64Ptr(0)
	resp := as.NewDescribeAutoScalingInstancesResponse()
	err := ts.asClient.Send(context.TODO(), req, resp)
	if err != nil {
		return nil, gerrors.Wrap(err, "[CloudAPIError]")
	}
	res := resp.Response.AutoScalingInstanceSet
	totalCount := uint64(0)
	if resp.Response.TotalCount != nil {
		totalCount = *resp.Response.TotalCount
	}
	for uint64(len(res)) < totalCount {
		req.Offset = common.Int64Ptr(int64(len(res)))
		resp := as.NewDescribeAutoScalingInstancesResponse()
		err := ts.asClient.Send(context.TODO(), req, resp)
		if err != nil {
			return nil, gerrors.Wrap(err, "[CloudAPIError]")
		}
		if resp.Response == nil || resp.Response.AutoScalingInstanceSet == nil {
			return nil, fmt.Errorf("[InvalidResponse] %s:%s", req.GetService(), req.GetAction())
		}
		if resp.Response.TotalCount != nil {
			if totalCount != *resp.Response.TotalCount {
				klog.Warningf("%s instance totalCount changed: %d->%d, reset request", asgRef.ID, totalCount, *resp.Response.TotalCount)
				totalCount = *resp.Response.TotalCount
				res = []*as.Instance{}
			}
		}
		res = append(res, resp.Response.AutoScalingInstanceSet...)
	}
	return res, nil
}

// GetAsgRefByInstanceRef returns asgRef according to instanceRef
func (ts *CloudServiceImpl) GetAsgRefByInstanceRef(instanceRef TcRef) (*TcRef, error) {
	if ts.asClient == nil {
		return nil, fmt.Errorf("asClient is not initialized")
	}

	req := as.NewDescribeAutoScalingInstancesRequest()
	req.InstanceIds = common.StringPtrs([]string{instanceRef.ID})
	res := as.NewDescribeAutoScalingInstancesResponse()
	err := ts.asClient.Send(context.TODO(), req, res)
	if err != nil {
		return nil, gerrors.Wrap(err, "[CloudAPIError]")
	}

	if res == nil || res.Response == nil ||
		res.Response.AutoScalingInstanceSet == nil ||
		res.Response.TotalCount == nil ||
		*res.Response.TotalCount != 1 ||
		len(res.Response.AutoScalingInstanceSet) != 1 {
		if *res.Response.TotalCount == 0 || len(res.Response.AutoScalingInstanceSet) == 0 {
			return nil, nil
		} else if res.Response.AutoScalingInstanceSet[0] == nil ||
			res.Response.AutoScalingInstanceSet[0].AutoScalingGroupId == nil {
			return nil, fmt.Errorf("[InvalidResponse] %s:%s", req.GetService(), req.GetAction())
		}
	}

	return &TcRef{
		ID: *res.Response.AutoScalingInstanceSet[0].AutoScalingGroupId,
	}, nil
}

// ResizeAsg set the target size of ASG.
func (ts *CloudServiceImpl) ResizeAsg(ref TcRef, size uint64) error {
	if ts.asClient == nil {
		return fmt.Errorf("asClient is not initialized")
	}
	req := as.NewModifyAutoScalingGroupRequest()
	req.AutoScalingGroupId = common.StringPtr(ref.ID)
	req.DesiredCapacity = common.Uint64Ptr(size)
	res := as.NewModifyAutoScalingGroupResponse()
	err := ts.asClient.Send(context.TODO(), req, res)
	if err != nil {
		return gerrors.Wrap(err, "[CloudAPIError]")
	}
	if res == nil || res.Response == nil ||
		res.Response.RequestId == nil {
		return fmt.Errorf("[InvalidResponse] %s:%s", req.GetService(), req.GetAction())
	}

	klog.V(4).Infof("ResizeAsg size %d, requestID: %s", size, *res.Response.RequestId)

	return nil
}

// GetZoneBySubnetID 查询子网的所属可用区
func (ts *CloudServiceImpl) GetZoneBySubnetID(subnetID string) (string, error) {
	if ts.vpcClient == nil {
		return "", fmt.Errorf("vpcClient is not initialized")
	}

	if ts.cvmClient == nil {
		return "", fmt.Errorf("cvmClient is not initialized")
	}
	req := vpc.NewDescribeSubnetsRequest()
	req.SubnetIds = []*string{common.StringPtr(subnetID)}
	res := vpc.NewDescribeSubnetsResponse()
	err := ts.vpcClient.Send(context.TODO(), req, res)
	if err != nil {
		return "", gerrors.Wrap(err, "[CloudAPIError]")
	} else if res.Response == nil || len(res.Response.SubnetSet) < 1 ||
		res.Response.SubnetSet[0] == nil || res.Response.SubnetSet[0].Zone == nil {
		return "", fmt.Errorf("[InvalidResponse] %s:%s", req.GetService(), req.GetAction())
	}
	zone := *res.Response.SubnetSet[0].Zone
	return zone, nil
}

// GetZoneInfo invokes cvm.DescribeZones to query zone information.
// zoneInfo will be cache.
func (ts *CloudServiceImpl) GetZoneInfo(zone string) (*cvm.ZoneInfo, error) {
	if zone == "" {
		return nil, fmt.Errorf("param is invalid: zone is empty")
	}
	if zoneInfo, exist := zoneInfos[zone]; exist {
		return zoneInfo, nil
	}
	req := cvm.NewDescribeZonesRequest()
	res := cvm.NewDescribeZonesResponse()
	err := ts.cvmClient.Send(context.TODO(), req, res)
	if err != nil {
		return nil, gerrors.Wrap(err, "[CloudAPIError]")
	} else if res.Response == nil || res.Response.TotalCount == nil || *res.Response.TotalCount < 1 {
		return nil, fmt.Errorf("[InvalidResponse] %s:%s", req.GetService(), req.GetAction())
	}
	for _, it := range res.Response.ZoneSet {
		if it != nil && it.Zone != nil && *it.Zone == zone {
			zoneInfos[zone] = it
			return it, nil
		}
	}
	return nil, fmt.Errorf("[InvalidResponse] %s:%s", req.GetService(), req.GetAction())
}

// GetTencentcloudInstanceRef returns a Tencentcloud ref.
func (ts *CloudServiceImpl) GetTencentcloudInstanceRef(instance *as.Instance) (*TcRef, error) {
	if instance == nil {
		return nil, fmt.Errorf("instance is nil")
	}

	zoneID := ""
	if instance.Zone != nil && *instance.Zone != "" {
		zoneInfo, err := ts.GetZoneInfo(*instance.Zone)
		if err != nil {
			return nil, err
		}
		zoneID = *zoneInfo.ZoneId
	}

	return &TcRef{
		ID:   *instance.InstanceId,
		Zone: zoneID,
	}, nil
}
