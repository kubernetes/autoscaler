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
type ListScalingPoliciesRequest struct {
	ScalingGroupId    string                                       `json:"scaling_group_id"`
	ScalingPolicyName *string                                      `json:"scaling_policy_name,omitempty"`
	ScalingPolicyType *ListScalingPoliciesRequestScalingPolicyType `json:"scaling_policy_type,omitempty"`
	ScalingPolicyId   *string                                      `json:"scaling_policy_id,omitempty"`
	StartNumber       *int32                                       `json:"start_number,omitempty"`
	Limit             *int32                                       `json:"limit,omitempty"`
}

func (o ListScalingPoliciesRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ListScalingPoliciesRequest", string(data)}, " ")
}

type ListScalingPoliciesRequestScalingPolicyType struct {
	value string
}

type ListScalingPoliciesRequestScalingPolicyTypeEnum struct {
	ALARM      ListScalingPoliciesRequestScalingPolicyType
	SCHEDULED  ListScalingPoliciesRequestScalingPolicyType
	RECURRENCE ListScalingPoliciesRequestScalingPolicyType
}

func GetListScalingPoliciesRequestScalingPolicyTypeEnum() ListScalingPoliciesRequestScalingPolicyTypeEnum {
	return ListScalingPoliciesRequestScalingPolicyTypeEnum{
		ALARM: ListScalingPoliciesRequestScalingPolicyType{
			value: "ALARM",
		},
		SCHEDULED: ListScalingPoliciesRequestScalingPolicyType{
			value: "SCHEDULED",
		},
		RECURRENCE: ListScalingPoliciesRequestScalingPolicyType{
			value: "RECURRENCE",
		},
	}
}

func (c ListScalingPoliciesRequestScalingPolicyType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListScalingPoliciesRequestScalingPolicyType) UnmarshalJSON(b []byte) error {
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
