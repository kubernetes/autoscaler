/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// 伸缩组实例详情
type ScalingGroupInstance struct {
	// 实例ID。
	InstanceId *string `json:"instance_id,omitempty"`
	// 实例名称。
	InstanceName *string `json:"instance_name,omitempty"`
	// 实例所在伸缩组ID。
	ScalingGroupId *string `json:"scaling_group_id,omitempty"`
	// 实例所在伸缩组名称。
	ScalingGroupName *string `json:"scaling_group_name,omitempty"`
	// 实例在伸缩组中的实力状态周期：INSERVICE： 正在使用。PENDING：正在加入伸缩组。REMOVING：正在移出伸缩组。PENDING_WAIT：正在加入伸缩组：等待。REMOVING_WAIT：正在移出伸缩组：等待。
	LifeCycleState *ScalingGroupInstanceLifeCycleState `json:"life_cycle_state,omitempty"`
	// 实例健康状态:INITAILIZING:初始化；NORMAL：正常；ERROR：错误。
	HealthStatus *ScalingGroupInstanceHealthStatus `json:"health_status,omitempty"`
	// 伸缩配置名称。如果返回为空，表示伸缩配置已经被删除。如果返回MANNUAL_ADD，表示实例为手动加入。
	ScalingConfigurationName *string `json:"scaling_configuration_name,omitempty"`
	// 伸缩配置ID。
	ScalingConfigurationId *string `json:"scaling_configuration_id,omitempty"`
	// 实例加入伸缩组的时间，遵循UTC时间。
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
	// 实例的实例保护属性。
	ProtectFromScalingDown *bool `json:"protect_from_scaling_down,omitempty"`
}

func (o ScalingGroupInstance) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ScalingGroupInstance", string(data)}, " ")
}

type ScalingGroupInstanceLifeCycleState struct {
	value string
}

type ScalingGroupInstanceLifeCycleStateEnum struct {
	INSERVICE     ScalingGroupInstanceLifeCycleState
	PENDING       ScalingGroupInstanceLifeCycleState
	REMOVING      ScalingGroupInstanceLifeCycleState
	PENDING_WAIT  ScalingGroupInstanceLifeCycleState
	REMOVING_WAIT ScalingGroupInstanceLifeCycleState
}

func GetScalingGroupInstanceLifeCycleStateEnum() ScalingGroupInstanceLifeCycleStateEnum {
	return ScalingGroupInstanceLifeCycleStateEnum{
		INSERVICE: ScalingGroupInstanceLifeCycleState{
			value: "INSERVICE",
		},
		PENDING: ScalingGroupInstanceLifeCycleState{
			value: "PENDING",
		},
		REMOVING: ScalingGroupInstanceLifeCycleState{
			value: "REMOVING",
		},
		PENDING_WAIT: ScalingGroupInstanceLifeCycleState{
			value: "PENDING_WAIT",
		},
		REMOVING_WAIT: ScalingGroupInstanceLifeCycleState{
			value: "REMOVING_WAIT",
		},
	}
}

func (c ScalingGroupInstanceLifeCycleState) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ScalingGroupInstanceLifeCycleState) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}

type ScalingGroupInstanceHealthStatus struct {
	value string
}

type ScalingGroupInstanceHealthStatusEnum struct {
	NORMAL       ScalingGroupInstanceHealthStatus
	ERROR        ScalingGroupInstanceHealthStatus
	INITAILIZING ScalingGroupInstanceHealthStatus
}

func GetScalingGroupInstanceHealthStatusEnum() ScalingGroupInstanceHealthStatusEnum {
	return ScalingGroupInstanceHealthStatusEnum{
		NORMAL: ScalingGroupInstanceHealthStatus{
			value: "NORMAL",
		},
		ERROR: ScalingGroupInstanceHealthStatus{
			value: "ERROR",
		},
		INITAILIZING: ScalingGroupInstanceHealthStatus{
			value: "INITAILIZING",
		},
	}
}

func (c ScalingGroupInstanceHealthStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ScalingGroupInstanceHealthStatus) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
