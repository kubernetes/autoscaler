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
type ListScalingGroupsRequest struct {
	ScalingGroupName       *string                                     `json:"scaling_group_name,omitempty"`
	ScalingConfigurationId *string                                     `json:"scaling_configuration_id,omitempty"`
	ScalingGroupStatus     *ListScalingGroupsRequestScalingGroupStatus `json:"scaling_group_status,omitempty"`
	StartNumber            *int32                                      `json:"start_number,omitempty"`
	Limit                  *int32                                      `json:"limit,omitempty"`
}

func (o ListScalingGroupsRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListScalingGroupsRequest", string(data)}, " ")
}

type ListScalingGroupsRequestScalingGroupStatus struct {
	value string
}

type ListScalingGroupsRequestScalingGroupStatusEnum struct {
	INSERVICE ListScalingGroupsRequestScalingGroupStatus
	PAUSED    ListScalingGroupsRequestScalingGroupStatus
	ERROR     ListScalingGroupsRequestScalingGroupStatus
	DELETING  ListScalingGroupsRequestScalingGroupStatus
}

func GetListScalingGroupsRequestScalingGroupStatusEnum() ListScalingGroupsRequestScalingGroupStatusEnum {
	return ListScalingGroupsRequestScalingGroupStatusEnum{
		INSERVICE: ListScalingGroupsRequestScalingGroupStatus{
			value: "INSERVICE",
		},
		PAUSED: ListScalingGroupsRequestScalingGroupStatus{
			value: "PAUSED",
		},
		ERROR: ListScalingGroupsRequestScalingGroupStatus{
			value: "ERROR",
		},
		DELETING: ListScalingGroupsRequestScalingGroupStatus{
			value: "DELETING",
		},
	}
}

func (c ListScalingGroupsRequestScalingGroupStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListScalingGroupsRequestScalingGroupStatus) UnmarshalJSON(b []byte) error {
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
