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

// 批量操作实例
type UpdateScalingGroupInstanceRequestBody struct {
	// 云服务器ID。
	InstancesId []string `json:"instances_id"`
	// 从伸缩组中移出实例时，是否删除云服务器。默认为no；可选值为yes或no。只有action为REMOVE时，这个字段才生效。
	InstanceDelete *string `json:"instance_delete,omitempty"`
	// 批量操作实例action标识：添加：ADD  移除： REMOVE  设置实例保护： PROTECT  取消实例保护： UNPROTECT；转入备用状态：ENTER_STANDBY 移出备用状态:EXIT_STANDBY
	Action UpdateScalingGroupInstanceRequestBodyAction `json:"action"`
	// 将实例移入备用状态时，是否补充新的云服务器。取值如下：no：不补充新的实例，默认情况为no。yes：补充新的实例。只有action为ENTER_STANDBY时，这个字段才生效。
	InstanceAppend *string `json:"instance_append,omitempty"`
}

func (o UpdateScalingGroupInstanceRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"UpdateScalingGroupInstanceRequestBody", string(data)}, " ")
}

type UpdateScalingGroupInstanceRequestBodyAction struct {
	value string
}

type UpdateScalingGroupInstanceRequestBodyActionEnum struct {
	ADD           UpdateScalingGroupInstanceRequestBodyAction
	REMOVE        UpdateScalingGroupInstanceRequestBodyAction
	PROTECT       UpdateScalingGroupInstanceRequestBodyAction
	UNPROTECT     UpdateScalingGroupInstanceRequestBodyAction
	ENTER_STANDBY UpdateScalingGroupInstanceRequestBodyAction
	EXIT_STANDBY  UpdateScalingGroupInstanceRequestBodyAction
}

func GetUpdateScalingGroupInstanceRequestBodyActionEnum() UpdateScalingGroupInstanceRequestBodyActionEnum {
	return UpdateScalingGroupInstanceRequestBodyActionEnum{
		ADD: UpdateScalingGroupInstanceRequestBodyAction{
			value: "ADD",
		},
		REMOVE: UpdateScalingGroupInstanceRequestBodyAction{
			value: "REMOVE",
		},
		PROTECT: UpdateScalingGroupInstanceRequestBodyAction{
			value: "PROTECT",
		},
		UNPROTECT: UpdateScalingGroupInstanceRequestBodyAction{
			value: "UNPROTECT",
		},
		ENTER_STANDBY: UpdateScalingGroupInstanceRequestBodyAction{
			value: "ENTER_STANDBY",
		},
		EXIT_STANDBY: UpdateScalingGroupInstanceRequestBodyAction{
			value: "EXIT_STANDBY",
		},
	}
}

func (c UpdateScalingGroupInstanceRequestBodyAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateScalingGroupInstanceRequestBodyAction) UnmarshalJSON(b []byte) error {
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
