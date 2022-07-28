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

package tencentcloud

import (
	"fmt"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/metrics"
	as "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	cvm "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	tke "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	vpc "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

const deleteInstanceModeTerminate = "terminate"

// CloudService is used for communicating with Tencentcloud API.
type CloudService interface {
	// FetchAsgInstances returns instances of the specified ASG.
	FetchAsgInstances(TcRef) ([]cloudprovider.Instance, error)
	// DeleteInstances remove instances of specified ASG.
	DeleteInstances(Asg, []string) error
	// GetAsgRefByInstanceRef returns asgRef according to instanceRef
	GetAsgRefByInstanceRef(TcRef) (TcRef, error)
	// GetAutoScalingGroups queries and returns a set of ASG.
	GetAutoScalingGroups([]string) ([]as.AutoScalingGroup, error)
	// GetAutoscalingConfigs queries and returns a set of ASG launchconfiguration.
	GetAutoscalingConfigs([]string) ([]as.LaunchConfiguration, error)
	// GetAutoScalingGroups queries and returns a set of ASG.
	GetAutoScalingGroup(TcRef) (as.AutoScalingGroup, error)
	// ResizeAsg set the target size of ASG.
	ResizeAsg(TcRef, uint64) error
	// GetAutoScalingInstances returns instances of specific ASG.
	GetAutoScalingInstances(TcRef) ([]*as.Instance, error)
	// GetTencentcloudInstanceRef returns a Tencentcloud ref.
	GetTencentcloudInstanceRef(*as.Instance) (TcRef, error)
	// DescribeVpcCniPodLimits list network limits
	DescribeVpcCniPodLimits(string) (*tke.PodLimitsInstance, error)
	// GetInstanceInfoByType queries the number of CPU, memory, and GPU resources of the model configured for generating template
	GetInstanceInfoByType(string) (*InstanceInfo, error)
	// GetNodePoolInfo returns nodepool information from TKE.
	GetNodePoolInfo(string, string) (*NodePoolInfo, error)
	// GetZoneBySubnetID return zone by subnetID.
	GetZoneBySubnetID(string) (string, error)
	// GetZoneInfo invokes cvm.DescribeZones to query zone information.
	GetZoneInfo(string) (*cvm.ZoneInfo, error)
}

// CloudServiceImpl provides several utility methods over the auto-scaling cloudService provided by Tencentcloud SDK
type CloudServiceImpl struct {
	asClient  *as.Client
	cvmClient *cvm.Client
	tkeClient *tke.Client
	vpcClient *vpc.Client
}

const (
	maxRecordsReturnedByAPI = 100
)

// CVM stop mode
const (
	StoppedModeStopCharging = "STOP_CHARGING"
	StoppedModeKeepCharging = "KEEP_CHARGING"
)

// AS scaling mode
const (
	ScalingModeClassic       = "CLASSIC_SCALING"
	ScalingModeWakeUpStopped = "WAKE_UP_STOPPED_SCALING"
)

var zoneInfos = make(map[string]*cvm.ZoneInfo)

// SubnetInfo represents subnet's detail
type SubnetInfo struct {
	SubnetID string
	Zone     string
	ZoneID   int
}

// NewCloudService creates an instance of caching CloudServiceImpl
func NewCloudService(cvmClient *cvm.Client, vpcClient *vpc.Client, asClient *as.Client, tkeClient *tke.Client) CloudService {
	return &CloudServiceImpl{
		cvmClient: cvmClient,
		vpcClient: vpcClient,
		asClient:  asClient,
		tkeClient: tkeClient,
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
// NOTICE: 一般情况下都是移除一个节点，只有在创建失败时才有批量移除，暂未分页处理
// 如果是关机模式的伸缩组进行关机操作，其它进行移除操作
// 目前执行移除操作走 TKE 接口，需要 cluster 属性，开源时切换为 as 接口
func (ts *CloudServiceImpl) DeleteInstances(asg Asg, instances []string) error {
	// TODO 处理缩容保护

	if asg.GetScalingType() == ScalingModeWakeUpStopped {
		return ts.stopInstances(asg, instances)
	}

	if ts.tkeClient == nil {
		return fmt.Errorf("tkeClient is not initialized")
	}
	req := tke.NewDeleteClusterInstancesRequest()
	req.InstanceIds = common.StringPtrs(instances)
	req.ClusterId = &cloudConfig.ClusterID
	req.InstanceDeleteMode = common.StringPtr(deleteInstanceModeTerminate)

	_, err := ts.tkeClient.DeleteClusterInstances(req)
	if err != nil {
		if e, ok := err.(*errors.TencentCloudSDKError); ok {
			metrics.RegisterCloudAPIInvokedError("tke", "DeleteClusterInstances", e.Code)
		}
	}

	return err
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
	resp, err := ts.asClient.DescribeLaunchConfigurations(req)
	metrics.RegisterCloudAPIInvoked("as", "DescribeLaunchConfigurations", err)
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Response == nil ||
		resp.Response.LaunchConfigurationSet == nil {
		return nil, fmt.Errorf("DescribeLaunchConfigurations returned a invalid response")
	}

	res := make([]as.LaunchConfiguration, 0)
	for _, lc := range resp.Response.LaunchConfigurationSet {
		if lc != nil {
			res = append(res, *lc)
		}
	}

	if len(res) != len(ascs) {
		return nil, fmt.Errorf("DescribeLaunchConfigurations need: %d, real: %d", len(ascs), len(res))
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
	instanceTypeRequest := cvm.NewDescribeInstanceTypeConfigsRequest()
	instanceTypeRequest.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("instance-type"),
			Values: []*string{&instanceType},
		},
	}

	resp, err := ts.cvmClient.DescribeInstanceTypeConfigs(instanceTypeRequest)
	metrics.RegisterCloudAPIInvoked("cvm", "DescribeInstanceTypeConfigs", err)
	if err != nil {
		return nil, err
	}

	if resp == nil ||
		resp.Response == nil ||
		resp.Response.InstanceTypeConfigSet == nil ||
		len(resp.Response.InstanceTypeConfigSet) < 1 ||
		resp.Response.InstanceTypeConfigSet[0].CPU == nil ||
		resp.Response.InstanceTypeConfigSet[0].Memory == nil ||
		resp.Response.InstanceTypeConfigSet[0].GPU == nil ||
		resp.Response.InstanceTypeConfigSet[0].InstanceFamily == nil ||
		resp.Response.InstanceTypeConfigSet[0].InstanceType == nil ||
		resp.Response.InstanceTypeConfigSet[0].Zone == nil {
		return nil, fmt.Errorf("DescribeInstanceTypeConfigs returned a invalid response")
	}

	return &InstanceInfo{
		*resp.Response.InstanceTypeConfigSet[0].CPU,
		*resp.Response.InstanceTypeConfigSet[0].Memory,
		*resp.Response.InstanceTypeConfigSet[0].GPU,
		*resp.Response.InstanceTypeConfigSet[0].InstanceFamily,
		*resp.Response.InstanceTypeConfigSet[0].InstanceType,
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

	response, err := ts.asClient.DescribeAutoScalingGroups(req)
	metrics.RegisterCloudAPIInvoked("as", "DescribeAutoScalingGroups", err)
	if err != nil {
		return nil, err
	}

	if response == nil || response.Response == nil || response.Response.AutoScalingGroupSet == nil {
		return nil, fmt.Errorf("DescribeAutoScalingGroups returned a invalid response")
	}

	asgs := make([]as.AutoScalingGroup, 0)
	for _, asg := range response.Response.AutoScalingGroupSet {
		if asg != nil {
			asgs = append(asgs, *asg)
		}
	}

	return asgs, nil
}

// GetAutoScalingGroup returns the specific ASG.
func (ts *CloudServiceImpl) GetAutoScalingGroup(asgRef TcRef) (as.AutoScalingGroup, error) {
	if ts.asClient == nil {
		return as.AutoScalingGroup{}, fmt.Errorf("asClient is not initialized")
	}

	req := as.NewDescribeAutoScalingGroupsRequest()
	req.AutoScalingGroupIds = []*string{&asgRef.ID}
	resp, err := ts.asClient.DescribeAutoScalingGroups(req)
	metrics.RegisterCloudAPIInvoked("as", "DescribeAutoScalingGroups", err)
	if err != nil {
		return as.AutoScalingGroup{}, err
	}

	if resp == nil || resp.Response == nil || resp.Response.AutoScalingGroupSet == nil ||
		len(resp.Response.AutoScalingGroupSet) != 1 || resp.Response.AutoScalingGroupSet[0] == nil {
		return as.AutoScalingGroup{}, fmt.Errorf("DescribeAutoScalingGroups returned a invalid response")
	}

	return *resp.Response.AutoScalingGroupSet[0], nil
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
	resp, err := ts.asClient.DescribeAutoScalingInstances(req)
	metrics.RegisterCloudAPIInvoked("as", "DescribeAutoScalingInstances", err)
	if err != nil {
		return nil, err
	}
	res := resp.Response.AutoScalingInstanceSet
	totalCount := uint64(0)
	if resp.Response.TotalCount != nil {
		totalCount = *resp.Response.TotalCount
	}
	for uint64(len(res)) < totalCount {
		req.Offset = common.Int64Ptr(int64(len(res)))
		resp, err = ts.asClient.DescribeAutoScalingInstances(req)
		metrics.RegisterCloudAPIInvoked("as", "DescribeAutoScalingInstances", err)
		if err != nil {
			return nil, err
		}
		if resp.Response == nil || resp.Response.AutoScalingInstanceSet == nil {
			return res, fmt.Errorf("query auto-scaling instances returned incorrect results")
		}
		if resp.Response.TotalCount != nil {
			if totalCount != *resp.Response.TotalCount {
				klog.Warningf("%s instance totalCount changed: %d->%d, reset request", totalCount, *resp.Response.TotalCount)
				totalCount = *resp.Response.TotalCount
				res = []*as.Instance{}
			}
		}
		res = append(res, resp.Response.AutoScalingInstanceSet...)
	}
	return res, nil
}

// GetAsgRefByInstanceRef returns asgRef according to instanceRef
func (ts *CloudServiceImpl) GetAsgRefByInstanceRef(instanceRef TcRef) (TcRef, error) {
	if ts.asClient == nil {
		return TcRef{}, fmt.Errorf("asClient is not initialized")
	}

	req := as.NewDescribeAutoScalingInstancesRequest()
	req.InstanceIds = common.StringPtrs([]string{instanceRef.ID})
	resp, err := ts.asClient.DescribeAutoScalingInstances(req)
	metrics.RegisterCloudAPIInvoked("as", "DescribeAutoScalingInstances", err)
	if err != nil {
		return TcRef{}, err
	}

	if resp == nil || resp.Response == nil ||
		resp.Response.AutoScalingInstanceSet == nil ||
		resp.Response.TotalCount == nil ||
		*resp.Response.TotalCount != 1 ||
		len(resp.Response.AutoScalingInstanceSet) != 1 {
		if *resp.Response.TotalCount == 0 || len(resp.Response.AutoScalingInstanceSet) == 0 {
			return TcRef{}, nil
		} else if resp.Response.AutoScalingInstanceSet[0] == nil ||
			resp.Response.AutoScalingInstanceSet[0].AutoScalingGroupId == nil {
			return TcRef{}, fmt.Errorf("DescribeAutoScalingInstances response is invalid by instance %v: %#v", instanceRef, resp)
		}
	}

	return TcRef{
		ID: *resp.Response.AutoScalingInstanceSet[0].AutoScalingGroupId,
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
	resp, err := ts.asClient.ModifyAutoScalingGroup(req)
	metrics.RegisterCloudAPIInvoked("as", "ModifyAutoScalingGroup", err)
	if err != nil {
		return err
	}
	if resp == nil || resp.Response == nil ||
		resp.Response.RequestId == nil {
		return fmt.Errorf("ModifyAutoScalingGroup returned a invalid response")
	}

	klog.V(4).Infof("ResizeAsg size %d, requestID: %s", size, *resp.Response.RequestId)

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
	resp, err := ts.vpcClient.DescribeSubnets(req)
	if err != nil {
		return "", err
	} else if resp.Response == nil || len(resp.Response.SubnetSet) < 1 ||
		resp.Response.SubnetSet[0] == nil || resp.Response.SubnetSet[0].Zone == nil {
		return "", fmt.Errorf("Failed to invoke DescribeSubnets from vpc, subnetId %v, err: response subenet is nil", subnetID)
	}
	zone := *resp.Response.SubnetSet[0].Zone
	return zone, nil
}

// GetZoneInfo invokes cvm.DescribeZones to query zone information.
// zoneInfo will be cache.
func (ts *CloudServiceImpl) GetZoneInfo(zone string) (*cvm.ZoneInfo, error) {
	if zone == "" {
		return nil, fmt.Errorf("Param is invalid: zone is empty")
	}
	if zoneInfo, exist := zoneInfos[zone]; exist {
		return zoneInfo, nil
	}
	req := cvm.NewDescribeZonesRequest()
	resp, err := ts.cvmClient.DescribeZones(req)
	metrics.RegisterCloudAPIInvoked("cvm", "DescribeZones", err)
	if err != nil {
		return nil, err
	} else if resp.Response == nil || resp.Response.TotalCount == nil || *resp.Response.TotalCount < 1 {
		return nil, fmt.Errorf("DescribeZones returns a invalid response")
	}
	for _, it := range resp.Response.ZoneSet {
		if it != nil && it.Zone != nil && *it.Zone == zone {
			zoneInfos[zone] = it
			return it, nil
		}
	}
	return nil, fmt.Errorf("Failed to get zoneInfo: %s is not exist", zone)
}

// GetTencentcloudInstanceRef returns a Tencentcloud ref.
func (ts *CloudServiceImpl) GetTencentcloudInstanceRef(instance *as.Instance) (TcRef, error) {
	if instance == nil {
		return TcRef{}, fmt.Errorf("instance is nil")
	}

	zoneID := ""
	if instance.Zone != nil && *instance.Zone != "" {
		zoneInfo, err := ts.GetZoneInfo(*instance.Zone)
		if err != nil {
			return TcRef{}, err
		}
		zoneID = *zoneInfo.ZoneId
	}

	return TcRef{
		ID:   *instance.InstanceId,
		Zone: zoneID,
	}, nil
}

// NodePoolInfo represents the information nodePool or clusterAsg from dashboard
type NodePoolInfo struct {
	Labels []*tke.Label
	Taints []*tke.Taint
}

// GetNodePoolInfo returns nodepool information from TKE.
func (ts *CloudServiceImpl) GetNodePoolInfo(clusterID string, asgID string) (*NodePoolInfo, error) {
	if ts.tkeClient == nil {
		return nil, fmt.Errorf("tkeClient is not initialized")
	}
	// From NodePool
	npReq := tke.NewDescribeClusterNodePoolsRequest()
	npReq.ClusterId = common.StringPtr(clusterID)
	npResp, err := ts.tkeClient.DescribeClusterNodePools(npReq)
	if e, ok := err.(*errors.TencentCloudSDKError); ok {
		metrics.RegisterCloudAPIInvokedError("tke", "DescribeClusterNodePools", e.Code)
		return nil, e
	}
	if npResp == nil || npResp.Response == nil || npResp.Response.RequestId == nil {
		return nil, errors.NewTencentCloudSDKError("DASHBOARD_ERROR", "empty response", "-")
	}
	var targetNodePool *tke.NodePool
	for _, np := range npResp.Response.NodePoolSet {
		if np.AutoscalingGroupId != nil && *np.AutoscalingGroupId == asgID {
			targetNodePool = np
			break
		}
	}
	if targetNodePool != nil {
		return &NodePoolInfo{
			Labels: targetNodePool.Labels,
			Taints: targetNodePool.Taints,
		}, nil
	}

	// Compatible with DEPRECATED autoScalingGroups
	asgReq := tke.NewDescribeClusterAsGroupsRequest()
	asgReq.AutoScalingGroupIds = []*string{common.StringPtr(asgID)}
	asgReq.ClusterId = common.StringPtr(clusterID)
	asgResp, err := ts.tkeClient.DescribeClusterAsGroups(asgReq)
	if err != nil {
		if e, ok := err.(*errors.TencentCloudSDKError); ok {
			metrics.RegisterCloudAPIInvokedError("tke", "DescribeClusterAsGroups", e.Code)
		}
		return nil, err
	}
	if asgResp == nil || asgResp.Response == nil {
		return nil, errors.NewTencentCloudSDKError("DASHBOARD_ERROR", "empty response", "-")
	}
	asgCount := len(asgResp.Response.ClusterAsGroupSet)
	if asgCount != 1 {
		return nil, errors.NewTencentCloudSDKError("UNEXPECTED_ERROR",
			fmt.Sprintf("%s get %d autoScalingGroup", asgID, asgCount), *asgResp.Response.RequestId)
	}
	asg := asgResp.Response.ClusterAsGroupSet[0]
	if asg == nil {
		return nil, errors.NewTencentCloudSDKError("UNEXPECTED_ERROR", "asg is nil", *asgResp.Response.RequestId)
	}

	return &NodePoolInfo{
		Labels: asg.Labels,
	}, nil
}

// DescribeVpcCniPodLimits list network limits
func (ts *CloudServiceImpl) DescribeVpcCniPodLimits(instanceType string) (*tke.PodLimitsInstance, error) {
	if ts.tkeClient == nil {
		return nil, fmt.Errorf("tkeClient is not initialized")
	}
	req := tke.NewDescribeVpcCniPodLimitsRequest()
	req.InstanceType = common.StringPtr(instanceType)
	resp, err := ts.tkeClient.DescribeVpcCniPodLimits(req)
	if err != nil {
		if e, ok := err.(*errors.TencentCloudSDKError); ok {
			metrics.RegisterCloudAPIInvokedError("tke", "DescribeVpcCniPodLimits", e.Code)
		}
		return nil, err
	}
	if resp == nil || resp.Response == nil || resp.Response.RequestId == nil {
		return nil, errors.NewTencentCloudSDKError("DASHBOARD_ERROR", "empty response", "-")
	}
	if len(resp.Response.PodLimitsInstanceSet) == 0 {
		return nil, nil
	}

	// PodLimitsInstanceSet 分可用区返回，会存在多组值，不过内容都一样，取第一个
	return resp.Response.PodLimitsInstanceSet[0], nil
}

func (ts *CloudServiceImpl) stopAutoScalingInstancesWithRetry(req *as.StopAutoScalingInstancesRequest) error {
	var err error
	scalingActivityID := ""
	for i := 0; i < retryCountStop; i++ {
		if i > 0 {
			time.Sleep(intervalTimeStop)
		}
		var resp = &as.StopAutoScalingInstancesResponse{}
		resp, err = ts.asClient.StopAutoScalingInstances(req)
		metrics.RegisterCloudAPIInvoked("as", "StopAutoScalingInstances", err)
		if err != nil {
			if asErr, ok := err.(*errors.TencentCloudSDKError); ok {
				// 仍然有不支持的
				if asErr.Code == "ResourceUnavailable.StoppedInstanceWithInconsistentChargingMode" {
					continue
				}
			}
			// 如果错误不是因为机型的原因，就重试
			klog.Warningf("StopAutoScalingInstances failed %v, %d retry", err.Error(), i)
		} else {
			if resp.Response.ActivityId != nil {
				scalingActivityID = *resp.Response.ActivityId
			}
			break
		}
	}

	if err != nil {
		return err
	}

	// check activity
	err = ts.ensureAutoScalingActivityDone(scalingActivityID)
	if err != nil {
		return err
	}

	return nil
}

// removeInstances invoke as.RemoveInstances
// api document: https://cloud.tencent.com/document/api/377/20431
func (ts *CloudServiceImpl) removeInstances(asg Asg, instances []string) error {
	if ts.asClient == nil {
		return fmt.Errorf("asClient is not initialized")
	}

	req := as.NewRemoveInstancesRequest()
	req.AutoScalingGroupId = common.StringPtr(asg.Id())
	req.InstanceIds = common.StringPtrs(instances)

	resp, err := ts.asClient.RemoveInstances(req)
	metrics.RegisterCloudAPIInvoked("as", "RemoveInstances", err)
	if err != nil {
		return err
	}
	if resp == nil || resp.Response == nil ||
		resp.Response.ActivityId == nil ||
		resp.Response.RequestId == nil {
		return fmt.Errorf("RemoveInstances returned a invalid response")
	}
	klog.V(4).Infof("Remove instances %v, asaID: %s, requestID: %s", instances, *resp.Response.ActivityId, *resp.Response.RequestId)

	return nil
}

// stopInstances 关闭instanceList中符合关机不收费的机型的机器，如果不支持的机器，就进行关机收费操作
// TODO 注意：该方法仅供上层调用，instanceList的长度需要上层控制，长度最大限制：100
func (ts *CloudServiceImpl) stopInstances(asg Asg, instances []string) error {
	if ts.asClient == nil {
		return fmt.Errorf("asClient is not initialized")
	}
	req := as.NewStopAutoScalingInstancesRequest()
	req.AutoScalingGroupId = common.StringPtr(asg.Id())
	req.InstanceIds = common.StringPtrs(instances)
	req.StoppedMode = common.StringPtr(StoppedModeStopCharging)

	// 没有dry run，所以先试跑一次，如果超过了，就直接超过了，
	keepChargingIns := make([]string, 0)
	stopChargingIns := make([]string, 0)
	var errOut error
	scalingActivityID := ""
	for i := 0; i < retryCountStop; i++ {
		// 从第二次开始，等待5s钟（一般autoscaling移出节点的时间为3s）
		if i > 0 {
			time.Sleep(intervalTimeStop)
		}
		resp, err := ts.asClient.StopAutoScalingInstances(req)
		metrics.RegisterCloudAPIInvoked("as", "StopAutoScalingInstances", err)
		if e, ok := err.(*errors.TencentCloudSDKError); ok {
			metrics.RegisterCloudAPIInvokedError("as", "StopAutoScalingInstances", e.Code)
		}
		if err == nil {
			// 一次成功
			klog.Info("StopAutoScalingInstances succeed")
			klog.V(4).Infof("res:%#v", resp.Response)
			if resp.Response.ActivityId != nil {
				scalingActivityID = *resp.Response.ActivityId
			}
			break
		} else if asErr, ok := err.(*errors.TencentCloudSDKError); ok &&
			(asErr.Code == "ResourceUnavailable.StoppedInstanceWithInconsistentChargingMode" ||
				asErr.Code == "ResourceUnavailable.InstanceNotSupportStopCharging") { // TODO 这里拿code和msg返回做判断，有点危险
			stopChargingIns, keepChargingIns = getInstanceIdsFromMessage(instances, asErr.Message)
			break
		} else {
			errOut = err
			klog.Warningf("Failed to StopAutoScalingInstances res:%#v", err)
		}
	}

	// 如果是一次就过了，说明instance是全部可以关机不收费的，直接结束
	if errOut == nil && scalingActivityID != "" {
		// check activity
		err := ts.ensureAutoScalingActivityDone(scalingActivityID)
		if err != nil {
			return err
		}
		return nil
	} else if errOut != nil {
		// 如果一直在报其他的错误，就返回错误
		return errOut
	}
	// 如果有不支持关机不收费的话，就分别执行两次。

	// 支持关机不收费的实例，进行`StopCharging`操作
	if len(stopChargingIns) != 0 {
		req.InstanceIds = common.StringPtrs(stopChargingIns)
		req.StoppedMode = common.StringPtr(StoppedModeStopCharging)
		err := ts.stopAutoScalingInstancesWithRetry(req)
		if err != nil {
			errOut = err
		}
	}

	if len(keepChargingIns) != 0 {
		// 不支持关机不收费的实例，进行`KeepCharging`操作
		req.InstanceIds = common.StringPtrs(keepChargingIns)
		req.StoppedMode = common.StringPtr(StoppedModeKeepCharging)
		err := ts.stopAutoScalingInstancesWithRetry(req)
		if err != nil {
			errOut = err
		}
	}
	if errOut != nil {
		return errOut
	}
	return nil
}

func (ts *CloudServiceImpl) ensureAutoScalingActivityDone(scalingActivityID string) error {
	if scalingActivityID == "" {
		return fmt.Errorf("ActivityId is nil")
	}

	checker := func(r interface{}, e error) bool {
		if e != nil {
			return false
		}
		resp, ok := r.(*as.DescribeAutoScalingActivitiesResponse)
		if !ok || resp.Response == nil || len(resp.Response.ActivitySet) != 1 {
			return false
		}
		if resp.Response.ActivitySet[0].StatusCode != nil {
			if *resp.Response.ActivitySet[0].StatusCode == "INIT" || *resp.Response.ActivitySet[0].StatusCode == "RUNNING" {
				return false
			}
			return true
		}
		return true
	}
	do := func() (interface{}, error) {
		if ts.asClient == nil {
			return nil, fmt.Errorf("asClient is not initialized")
		}
		req := as.NewDescribeAutoScalingActivitiesRequest()
		req.ActivityIds = common.StringPtrs([]string{scalingActivityID})

		resp, err := ts.asClient.DescribeAutoScalingActivities(req)
		metrics.RegisterCloudAPIInvoked("as", "DescribeAutoScalingActivities", err)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	ret, isTimeout, err := retryDo(do, checker, 1200, 2)
	if err != nil {
		return fmt.Errorf("EnsureAutoScalingActivityDone scalingActivityId:%s failed:%v", scalingActivityID, err)
	}

	if isTimeout {
		return fmt.Errorf("EnsureAutoScalingActivityDone scalingActivityId:%s timeout", scalingActivityID)
	}
	resp, ok := ret.(*as.DescribeAutoScalingActivitiesResponse)
	if !ok || resp.Response == nil || len(resp.Response.ActivitySet) != 1 {
		return fmt.Errorf("EnsureAutoScalingActivityDone scalingActivityId:%s failed", scalingActivityID)
	}
	if resp.Response.ActivitySet[0].StatusCode != nil && *resp.Response.ActivitySet[0].StatusCode != "SUCCESSFUL" {
		if resp.Response.ActivitySet[0].StatusMessageSimplified == nil {
			resp.Response.ActivitySet[0].StatusMessageSimplified = common.StringPtr("no message")
		}
		return fmt.Errorf("AutoScalingActivity scalingActivityId:%s %s %s", scalingActivityID, *resp.Response.ActivitySet[0].StatusCode, *resp.Response.ActivitySet[0].StatusMessageSimplified)
	}
	return nil
}
