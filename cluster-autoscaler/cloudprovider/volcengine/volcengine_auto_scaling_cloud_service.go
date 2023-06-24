/*
Copyright 2023 The Kubernetes Authors.

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

package volcengine

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/service/autoscaling"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/credentials"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/volcengine/volcengine-go-sdk/volcengine/session"
	"k8s.io/klog/v2"
)

// AutoScalingService is the interface for volcengine auto-scaling service
type AutoScalingService interface {
	GetScalingGroupById(groupId string) (*autoscaling.ScalingGroupForDescribeScalingGroupsOutput, error)
	ListScalingInstancesByGroupId(groupId string) ([]*autoscaling.ScalingInstanceForDescribeScalingInstancesOutput, error)
	GetScalingConfigurationById(configurationId string) (*autoscaling.ScalingConfigurationForDescribeScalingConfigurationsOutput, error)
	RemoveInstances(groupId string, instanceIds []string) error
	SetAsgTargetSize(groupId string, targetSize int) error
	SetAsgDesireCapacity(groupId string, desireCapacity int) error
}

type autoScalingService struct {
	autoscalingClient *autoscaling.AUTOSCALING
}

func (a *autoScalingService) SetAsgDesireCapacity(groupId string, desireCapacity int) error {
	_, err := a.autoscalingClient.ModifyScalingGroupCommon(&map[string]interface{}{
		"ScalingGroupId":       groupId,
		"DesireInstanceNumber": desireCapacity,
	})
	return err
}

func (a *autoScalingService) SetAsgTargetSize(groupId string, targetSize int) error {
	uid, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	resp, err := a.autoscalingClient.CreateScalingPolicyCommon(&map[string]interface{}{
		"AdjustmentType":             "TotalCapacity",
		"AdjustmentValue":            targetSize,
		"Cooldown":                   0,
		"ScalingGroupId":             groupId,
		"ScalingPolicyName":          fmt.Sprintf("autoscaler-autogen-scaling-policy-%s", uid.String()),
		"ScalingPolicyType":          "Scheduled",
		"ScheduledPolicy.LaunchTime": time.Now().Add(2 * time.Minute).UTC().Format("2006-01-02T15:04Z"),
	})
	if err != nil {
		klog.Errorf("failed to create scaling policy, err: %v", err)
		return err
	}

	scalingPolicyId := (*resp)["Result"].(map[string]interface{})["ScalingPolicyId"].(string)
	klog.Infof("create scaling policy response: %v, scalingPolicyId: %s", resp, scalingPolicyId)

	defer func() {
		// delete scaling policy
		_, err = a.autoscalingClient.DeleteScalingPolicyCommon(&map[string]interface{}{
			"ScalingPolicyId": scalingPolicyId,
		})
		if err != nil {
			klog.Warningf("failed to delete scaling policy %s, err: %v", scalingPolicyId, err)
		}
	}()

	_, err = a.autoscalingClient.EnableScalingPolicyCommon(&map[string]interface{}{
		"ScalingPolicyId": scalingPolicyId,
	})

	if err != nil {
		klog.Errorf("failed to enable scaling policy %s, err: %v", scalingPolicyId, err)
		return err
	}

	return wait.Poll(5*time.Second, 30*time.Minute, func() (bool, error) {
		// check scaling group status
		group, err := a.GetScalingGroupById(groupId)
		if err != nil {
			return false, err
		}
		if *group.DesireInstanceNumber == int32(targetSize) {
			return true, nil
		}
		return false, nil
	})
}

func (a *autoScalingService) RemoveInstances(groupId string, instanceIds []string) error {
	if len(instanceIds) == 0 {
		return nil
	}

	instanceIdGroups := StringSliceInGroupsOf(instanceIds, 20)
	for _, instanceIdGroup := range instanceIdGroups {
		_, err := a.autoscalingClient.RemoveInstances(&autoscaling.RemoveInstancesInput{
			ScalingGroupId: volcengine.String(groupId),
			InstanceIds:    volcengine.StringSlice(instanceIdGroup),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *autoScalingService) GetScalingConfigurationById(configurationId string) (*autoscaling.ScalingConfigurationForDescribeScalingConfigurationsOutput, error) {
	configurations, err := a.autoscalingClient.DescribeScalingConfigurations(&autoscaling.DescribeScalingConfigurationsInput{
		ScalingConfigurationIds: volcengine.StringSlice([]string{configurationId}),
	})
	if err != nil {
		return nil, err
	}
	if len(configurations.ScalingConfigurations) == 0 ||
		volcengine.StringValue(configurations.ScalingConfigurations[0].ScalingConfigurationId) != configurationId {
		return nil, fmt.Errorf("scaling configuration %s not found", configurationId)
	}
	return configurations.ScalingConfigurations[0], nil
}

func (a *autoScalingService) ListScalingInstancesByGroupId(groupId string) ([]*autoscaling.ScalingInstanceForDescribeScalingInstancesOutput, error) {
	req := &autoscaling.DescribeScalingInstancesInput{
		ScalingGroupId: volcengine.String(groupId),
		PageSize:       volcengine.Int32(50),
		PageNumber:     volcengine.Int32(1),
	}
	resp, err := a.autoscalingClient.DescribeScalingInstances(req)
	if err != nil {
		return nil, err
	}

	total := volcengine.Int32Value(resp.TotalCount)
	if total <= 50 {
		return resp.ScalingInstances, nil
	}

	res := make([]*autoscaling.ScalingInstanceForDescribeScalingInstancesOutput, 0)
	res = append(res, resp.ScalingInstances...)
	totalNumber := math.Ceil(float64(total) / 50)
	for i := 2; i <= int(totalNumber); i++ {
		req.PageNumber = volcengine.Int32(int32(i))
		resp, err = a.autoscalingClient.DescribeScalingInstances(req)
		if err != nil {
			return nil, err
		}
		res = append(res, resp.ScalingInstances...)
	}

	return res, nil
}

func (a *autoScalingService) GetScalingGroupById(groupId string) (*autoscaling.ScalingGroupForDescribeScalingGroupsOutput, error) {
	groups, err := a.autoscalingClient.DescribeScalingGroups(&autoscaling.DescribeScalingGroupsInput{
		ScalingGroupIds: volcengine.StringSlice([]string{groupId}),
	})
	if err != nil {
		return nil, err
	}
	if len(groups.ScalingGroups) == 0 {
		return nil, fmt.Errorf("scaling group %s not found", groupId)
	}
	return groups.ScalingGroups[0], nil
}

func newAutoScalingService(cloudConfig *cloudConfig) AutoScalingService {
	config := volcengine.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(cloudConfig.getAccessKey(), cloudConfig.getSecretKey(), "")).
		WithRegion(cloudConfig.getRegion()).
		WithEndpoint(cloudConfig.getEndpoint())
	sess, _ := session.NewSession(config)
	client := autoscaling.New(sess)
	return &autoScalingService{
		autoscalingClient: client,
	}
}
