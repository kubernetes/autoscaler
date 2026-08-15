package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 批量添加实例
type BatchAddInstancesOption struct {
	// 云服务器ID。

	InstancesId []string `json:"instances_id"`
	// 从伸缩组中移出实例时，是否删除云服务器。默认为no；可选值为yes或no。只有action为REMOVE时，这个字段才生效。

	InstanceDelete *BatchAddInstancesOptionInstanceDelete `json:"instance_delete,omitempty"`
	// 批量操作实例action标识：添加：ADD  移除： REMOVE  设置实例保护： PROTECT  取消实例保护： UNPROTECT；转入备用状态：ENTER_STANDBY 移出备用状态:EXIT_STANDBY

	Action BatchAddInstancesOptionAction `json:"action"`
	// 将实例移入备用状态时，是否补充新的云服务器。取值如下：no：不补充新的实例，默认情况为no。yes：补充新的实例。只有action为ENTER_STANDBY时，这个字段才生效。

	InstanceAppend *BatchAddInstancesOptionInstanceAppend `json:"instance_append,omitempty"`
}

func (o BatchAddInstancesOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchAddInstancesOption struct{}"
	}

	return strings.Join([]string{"BatchAddInstancesOption", string(data)}, " ")
}

type BatchAddInstancesOptionInstanceDelete struct {
	value string
}

type BatchAddInstancesOptionInstanceDeleteEnum struct {
	YES BatchAddInstancesOptionInstanceDelete
	NO  BatchAddInstancesOptionInstanceDelete
}

func GetBatchAddInstancesOptionInstanceDeleteEnum() BatchAddInstancesOptionInstanceDeleteEnum {
	return BatchAddInstancesOptionInstanceDeleteEnum{
		YES: BatchAddInstancesOptionInstanceDelete{
			value: "yes",
		},
		NO: BatchAddInstancesOptionInstanceDelete{
			value: "no",
		},
	}
}

func (c BatchAddInstancesOptionInstanceDelete) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchAddInstancesOptionInstanceDelete) UnmarshalJSON(b []byte) error {
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

type BatchAddInstancesOptionAction struct {
	value string
}

type BatchAddInstancesOptionActionEnum struct {
	ADD BatchAddInstancesOptionAction
}

func GetBatchAddInstancesOptionActionEnum() BatchAddInstancesOptionActionEnum {
	return BatchAddInstancesOptionActionEnum{
		ADD: BatchAddInstancesOptionAction{
			value: "ADD",
		},
	}
}

func (c BatchAddInstancesOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchAddInstancesOptionAction) UnmarshalJSON(b []byte) error {
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

type BatchAddInstancesOptionInstanceAppend struct {
	value string
}

type BatchAddInstancesOptionInstanceAppendEnum struct {
	NO  BatchAddInstancesOptionInstanceAppend
	YES BatchAddInstancesOptionInstanceAppend
}

func GetBatchAddInstancesOptionInstanceAppendEnum() BatchAddInstancesOptionInstanceAppendEnum {
	return BatchAddInstancesOptionInstanceAppendEnum{
		NO: BatchAddInstancesOptionInstanceAppend{
			value: "no",
		},
		YES: BatchAddInstancesOptionInstanceAppend{
			value: "yes",
		},
	}
}

func (c BatchAddInstancesOptionInstanceAppend) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchAddInstancesOptionInstanceAppend) UnmarshalJSON(b []byte) error {
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
