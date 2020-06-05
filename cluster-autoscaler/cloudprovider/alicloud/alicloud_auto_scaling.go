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

package alicloud

import (
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/sdk/requests"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/alicloud/alibaba-cloud-sdk-go/services/ess"
	klog "k8s.io/klog/v2"
	"time"
)

const (
	refreshClientInterval   = 60 * time.Minute
	acsAutogenIncreaseRules = "acs-autogen-increase-rules"
	defaultAdjustmentType   = "TotalCapacity"
)

//autoScaling define the interface usage in alibaba-cloud-sdk-go.
type autoScaling interface {
	DescribeScalingGroups(req *ess.DescribeScalingGroupsRequest) (*ess.DescribeScalingGroupsResponse, error)
	DescribeScalingConfigurations(req *ess.DescribeScalingConfigurationsRequest) (*ess.DescribeScalingConfigurationsResponse, error)
	DescribeScalingRules(req *ess.DescribeScalingRulesRequest) (*ess.DescribeScalingRulesResponse, error)
	DescribeScalingInstances(req *ess.DescribeScalingInstancesRequest) (*ess.DescribeScalingInstancesResponse, error)
	CreateScalingRule(req *ess.CreateScalingRuleRequest) (*ess.CreateScalingRuleResponse, error)
	ModifyScalingGroup(req *ess.ModifyScalingGroupRequest) (*ess.ModifyScalingGroupResponse, error)
	RemoveInstances(req *ess.RemoveInstancesRequest) (*ess.RemoveInstancesResponse, error)
	ExecuteScalingRule(req *ess.ExecuteScalingRuleRequest) (*ess.ExecuteScalingRuleResponse, error)
	ModifyScalingRule(req *ess.ModifyScalingRuleRequest) (*ess.ModifyScalingRuleResponse, error)
	DeleteScalingRule(req *ess.DeleteScalingRuleRequest) (*ess.DeleteScalingRuleResponse, error)
}

func newAutoScalingWrapper(cfg *cloudConfig) (*autoScalingWrapper, error) {
	if cfg.isValid() == false {
		//Never reach here.
		return nil, fmt.Errorf("your cloud config is not valid")
	}
	asw := &autoScalingWrapper{
		cfg: cfg,
	}
	if cfg.STSEnabled == true {
		go func(asw *autoScalingWrapper, cfg *cloudConfig) {
			timer := time.NewTicker(refreshClientInterval)
			defer timer.Stop()
			for {
				select {
				case <-timer.C:
					client, err := getEssClient(cfg)
					if err == nil {
						asw.autoScaling = client
					}
				}
			}
		}(asw, cfg)
	}
	client, err := getEssClient(cfg)
	if err == nil {
		asw.autoScaling = client
	}
	return asw, err
}

func getEssClient(cfg *cloudConfig) (client *ess.Client, err error) {
	region := cfg.getRegion()
	if cfg.STSEnabled == true {
		auth, err := cfg.getSTSToken()
		if err != nil {
			klog.Errorf("Failed to get sts token from metadata,Because of %s", err.Error())
			return nil, err
		}
		client, err = ess.NewClientWithStsToken(region, auth.AccessKeyId, auth.AccessKeySecret, auth.SecurityToken)
		if err != nil {
			klog.Errorf("Failed to create client with sts in metadata because of %s", err.Error())
		}
	} else {
		client, err = ess.NewClientWithAccessKey(region, cfg.AccessKeyID, cfg.AccessKeySecret)
		if err != nil {
			klog.Errorf("Failed to create ess client with AccessKeyId and AccessKeySecret,Because of %s", err.Error())
		}
	}
	return
}

//autoScalingWrapper will serve as the
type autoScalingWrapper struct {
	autoScaling
	cfg *cloudConfig
}

func (m autoScalingWrapper) getScalingGroupConfigurationByID(configID string, asgId string) (*ess.ScalingConfiguration, error) {
	params := ess.CreateDescribeScalingConfigurationsRequest()
	params.ScalingConfigurationId1 = configID
	params.ScalingGroupId = asgId

	resp, err := m.DescribeScalingConfigurations(params)
	if err != nil {
		klog.Errorf("failed to get ScalingConfiguration info request for %s,because of %s", configID, err.Error())
		return nil, err
	}

	configurations := resp.ScalingConfigurations.ScalingConfiguration

	if len(configurations) < 1 {
		return nil, fmt.Errorf("unable to get first ScalingConfiguration for %s", configID)
	}
	if len(configurations) > 1 {
		klog.Warningf("more than one ScalingConfiguration found for config(%q) and ASG(%q)", configID, asgId)
	}

	return &configurations[0], nil
}

func (m autoScalingWrapper) getScalingGroupByID(groupID string) (*ess.ScalingGroup, error) {
	params := ess.CreateDescribeScalingGroupsRequest()
	params.ScalingGroupId1 = groupID

	resp, err := m.DescribeScalingGroups(params)
	if err != nil {
		return nil, err
	}
	groups := resp.ScalingGroups.ScalingGroup
	if len(groups) < 1 {
		return nil, fmt.Errorf("unable to get first ScalingGroup for %s", groupID)
	}
	if len(groups) > 1 {
		klog.Warningf("more than one ScalingGroup for %s, use first one", groupID)
	}
	return &groups[0], nil
}

func (m autoScalingWrapper) getScalingGroupByName(groupName string) (*ess.ScalingGroup, error) {
	params := ess.CreateDescribeScalingGroupsRequest()
	params.ScalingGroupName = groupName

	resp, err := m.DescribeScalingGroups(params)
	if err != nil {
		return nil, err
	}
	groups := resp.ScalingGroups.ScalingGroup
	if len(groups) < 1 {
		return nil, fmt.Errorf("unable to get first ScalingGroup for %q", groupName)
	}
	if len(groups) > 1 {
		klog.Warningf("more than one ScalingGroup for %q, use first one", groupName)
	}
	return &groups[0], nil
}

func (m autoScalingWrapper) getScalingInstancesByGroup(asgId string) ([]ess.ScalingInstance, error) {
	params := ess.CreateDescribeScalingInstancesRequest()
	params.ScalingGroupId = asgId
	resp, err := m.DescribeScalingInstances(params)
	if err != nil {
		klog.Errorf("failed to request scaling instances for %s,Because of %s", asgId, err.Error())
		return nil, err
	}
	return resp.ScalingInstances.ScalingInstance, nil
}

func (m autoScalingWrapper) setCapcityInstanceSize(groupId string, capcityInstanceSize int64) error {
	var (
		ruleId         string
		scalingRuleAri string
	)
	req := ess.CreateDescribeScalingRulesRequest()
	req.RegionId = m.cfg.getRegion()
	req.ScalingGroupId = groupId
	req.ScalingRuleName1 = acsAutogenIncreaseRules
	resp, err := m.DescribeScalingRules(req)
	if err != nil {
		//need to handle
		return err
	}

	defer func() {
		deleteReq := ess.CreateDeleteScalingRuleRequest()
		deleteReq.ScalingRuleId = ruleId
		deleteReq.RegionId = m.cfg.getRegion()
		_, err := m.DeleteScalingRule(deleteReq)
		if err != nil {
			klog.Warningf("failed to clean scaling group rules,Because of %s", err.Error())
		}
	}()

	if len(resp.ScalingRules.ScalingRule) == 0 {
		//found the specific rules
		createReq := ess.CreateCreateScalingRuleRequest()
		createReq.RegionId = m.cfg.getRegion()
		createReq.ScalingGroupId = groupId
		createReq.AdjustmentType = defaultAdjustmentType
		createReq.AdjustmentValue = requests.NewInteger64(capcityInstanceSize)
		resp, err := m.CreateScalingRule(createReq)
		if err != nil {
			return err
		}
		ruleId = resp.ScalingRuleId
		scalingRuleAri = resp.ScalingRuleAri
	} else {
		ruleId = resp.ScalingRules.ScalingRule[0].ScalingRuleId
		scalingRuleAri = resp.ScalingRules.ScalingRule[0].ScalingRuleAri
	}

	modifyReq := ess.CreateModifyScalingRuleRequest()
	modifyReq.RegionId = m.cfg.getRegion()
	modifyReq.ScalingRuleId = ruleId
	modifyReq.AdjustmentType = defaultAdjustmentType
	modifyReq.AdjustmentValue = requests.NewInteger64(capcityInstanceSize)
	_, err = m.ModifyScalingRule(modifyReq)
	if err != nil {
		return err
	}
	executeReq := ess.CreateExecuteScalingRuleRequest()
	executeReq.RegionId = m.cfg.getRegion()
	executeReq.ScalingRuleAri = scalingRuleAri

	_, err = m.ExecuteScalingRule(executeReq)
	if err != nil {
		return err
	}
	return nil
}
