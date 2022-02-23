package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type ListScalingInstancesRequest struct {
	// 伸缩组ID。

	ScalingGroupId string `json:"scaling_group_id"`
	// 实例在伸缩组中的生命周期状态：INSERVICE： 正在使用。PENDING：正在加入伸缩组。REMOVING：正在移出伸缩组。PENDING_WAIT：正在加入伸缩组：等待。REMOVING_WAIT：正在移出伸缩组：等待。

	LifeCycleState *ListScalingInstancesRequestLifeCycleState `json:"life_cycle_state,omitempty"`
	// 实例健康状态：INITIALIZING：初始化。NORMAL：正常。ERROR：异常

	HealthStatus *ListScalingInstancesRequestHealthStatus `json:"health_status,omitempty"`
	// 实例保护状态：true：已设置实例保护。false：未设置实例保护。

	ProtectFromScalingDown *ListScalingInstancesRequestProtectFromScalingDown `json:"protect_from_scaling_down,omitempty"`
	// 查询的起始行号，默认为0。

	StartNumber *int32 `json:"start_number,omitempty"`
	// 查询的记录条数，默认为20。

	Limit *int32 `json:"limit,omitempty"`
}

func (o ListScalingInstancesRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingInstancesRequest struct{}"
	}

	return strings.Join([]string{"ListScalingInstancesRequest", string(data)}, " ")
}

type ListScalingInstancesRequestLifeCycleState struct {
	value string
}

type ListScalingInstancesRequestLifeCycleStateEnum struct {
	INSERVICE        ListScalingInstancesRequestLifeCycleState
	PENDING          ListScalingInstancesRequestLifeCycleState
	REMOVING         ListScalingInstancesRequestLifeCycleState
	PENDING_WAIT     ListScalingInstancesRequestLifeCycleState
	REMOVING_WAIT    ListScalingInstancesRequestLifeCycleState
	STANDBY          ListScalingInstancesRequestLifeCycleState
	ENTERING_STANDBY ListScalingInstancesRequestLifeCycleState
}

func GetListScalingInstancesRequestLifeCycleStateEnum() ListScalingInstancesRequestLifeCycleStateEnum {
	return ListScalingInstancesRequestLifeCycleStateEnum{
		INSERVICE: ListScalingInstancesRequestLifeCycleState{
			value: "INSERVICE",
		},
		PENDING: ListScalingInstancesRequestLifeCycleState{
			value: "PENDING",
		},
		REMOVING: ListScalingInstancesRequestLifeCycleState{
			value: "REMOVING",
		},
		PENDING_WAIT: ListScalingInstancesRequestLifeCycleState{
			value: "PENDING_WAIT",
		},
		REMOVING_WAIT: ListScalingInstancesRequestLifeCycleState{
			value: "REMOVING_WAIT",
		},
		STANDBY: ListScalingInstancesRequestLifeCycleState{
			value: "STANDBY",
		},
		ENTERING_STANDBY: ListScalingInstancesRequestLifeCycleState{
			value: "ENTERING_STANDBY",
		},
	}
}

func (c ListScalingInstancesRequestLifeCycleState) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListScalingInstancesRequestLifeCycleState) UnmarshalJSON(b []byte) error {
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

type ListScalingInstancesRequestHealthStatus struct {
	value string
}

type ListScalingInstancesRequestHealthStatusEnum struct {
	INITIALIZING ListScalingInstancesRequestHealthStatus
	NORMAL       ListScalingInstancesRequestHealthStatus
	ERROR        ListScalingInstancesRequestHealthStatus
}

func GetListScalingInstancesRequestHealthStatusEnum() ListScalingInstancesRequestHealthStatusEnum {
	return ListScalingInstancesRequestHealthStatusEnum{
		INITIALIZING: ListScalingInstancesRequestHealthStatus{
			value: "INITIALIZING",
		},
		NORMAL: ListScalingInstancesRequestHealthStatus{
			value: "NORMAL",
		},
		ERROR: ListScalingInstancesRequestHealthStatus{
			value: "ERROR",
		},
	}
}

func (c ListScalingInstancesRequestHealthStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListScalingInstancesRequestHealthStatus) UnmarshalJSON(b []byte) error {
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

type ListScalingInstancesRequestProtectFromScalingDown struct {
	value string
}

type ListScalingInstancesRequestProtectFromScalingDownEnum struct {
	TRUE  ListScalingInstancesRequestProtectFromScalingDown
	FALSE ListScalingInstancesRequestProtectFromScalingDown
}

func GetListScalingInstancesRequestProtectFromScalingDownEnum() ListScalingInstancesRequestProtectFromScalingDownEnum {
	return ListScalingInstancesRequestProtectFromScalingDownEnum{
		TRUE: ListScalingInstancesRequestProtectFromScalingDown{
			value: "true",
		},
		FALSE: ListScalingInstancesRequestProtectFromScalingDown{
			value: "false",
		},
	}
}

func (c ListScalingInstancesRequestProtectFromScalingDown) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ListScalingInstancesRequestProtectFromScalingDown) UnmarshalJSON(b []byte) error {
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
