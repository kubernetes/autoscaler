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
	"strings"
)

// Request Object
type ListScalingInstancesRequest struct {
	ScalingGroupId         string                                             `json:"scaling_group_id"`
	LifeCycleState         *ListScalingInstancesRequestLifeCycleState         `json:"life_cycle_state,omitempty"`
	HealthStatus           *ListScalingInstancesRequestHealthStatus           `json:"health_status,omitempty"`
	ProtectFromScalingDown *ListScalingInstancesRequestProtectFromScalingDown `json:"protect_from_scaling_down,omitempty"`
	StartNumber            *int32                                             `json:"start_number,omitempty"`
	Limit                  *int32                                             `json:"limit,omitempty"`
}

func (o ListScalingInstancesRequest) String() string {
	data, _ := json.Marshal(o)
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
	return json.Marshal(c.value)
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
	return json.Marshal(c.value)
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
	return json.Marshal(c.value)
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
