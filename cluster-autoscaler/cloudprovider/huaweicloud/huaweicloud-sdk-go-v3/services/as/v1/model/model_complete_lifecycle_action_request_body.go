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

// 伸缩实例生命周期回调
type CompleteLifecycleActionRequestBody struct {
	// 生命周期操作令牌，通过查询伸缩实例挂起信息接口获取。指定生命周期回调对象，当不传入instance_id字段时，该字段为必选。当该字段与instance_id字段都传入，优先使用该字段进行回调。
	LifecycleActionKey *string `json:"lifecycle_action_key,omitempty"`
	// 实例ID。指定生命周期回调对象，当不传入lifecycle_action_key字段时，该字段为必选。
	InstanceId *string `json:"instance_id,omitempty"`
	// 生命周期挂钩名称。指定生命周期回调对象，当不传入lifecycle_action_key字段时，该字段为必选。
	LifecycleHookName *string `json:"lifecycle_hook_name,omitempty"`
	// 生命周期回调操作。ABANDON：终止。CONTINUE：继续。EXTEND：延长超时时间，每次延长1小时。
	LifecycleActionResult CompleteLifecycleActionRequestBodyLifecycleActionResult `json:"lifecycle_action_result"`
}

func (o CompleteLifecycleActionRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CompleteLifecycleActionRequestBody", string(data)}, " ")
}

type CompleteLifecycleActionRequestBodyLifecycleActionResult struct {
	value string
}

type CompleteLifecycleActionRequestBodyLifecycleActionResultEnum struct {
	ABANDON  CompleteLifecycleActionRequestBodyLifecycleActionResult
	CONTINUE CompleteLifecycleActionRequestBodyLifecycleActionResult
	EXTEND   CompleteLifecycleActionRequestBodyLifecycleActionResult
}

func GetCompleteLifecycleActionRequestBodyLifecycleActionResultEnum() CompleteLifecycleActionRequestBodyLifecycleActionResultEnum {
	return CompleteLifecycleActionRequestBodyLifecycleActionResultEnum{
		ABANDON: CompleteLifecycleActionRequestBodyLifecycleActionResult{
			value: "ABANDON",
		},
		CONTINUE: CompleteLifecycleActionRequestBodyLifecycleActionResult{
			value: "CONTINUE",
		},
		EXTEND: CompleteLifecycleActionRequestBodyLifecycleActionResult{
			value: "EXTEND",
		},
	}
}

func (c CompleteLifecycleActionRequestBodyLifecycleActionResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CompleteLifecycleActionRequestBodyLifecycleActionResult) UnmarshalJSON(b []byte) error {
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
