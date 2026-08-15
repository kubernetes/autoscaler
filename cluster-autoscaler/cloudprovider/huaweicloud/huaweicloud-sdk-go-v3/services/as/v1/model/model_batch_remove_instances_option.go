package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 批量移除实例
type BatchRemoveInstancesOption struct {
	// 云服务器ID。

	InstancesId []string `json:"instances_id"`
	// 从伸缩组中移出实例时，是否删除云服务器。默认为no；可选值为yes或no。只有action为REMOVE时，这个字段才生效。

	InstanceDelete *BatchRemoveInstancesOptionInstanceDelete `json:"instance_delete,omitempty"`
	// 批量操作实例action标识：添加：ADD  移除： REMOVE  设置实例保护： PROTECT  取消实例保护： UNPROTECT；转入备用状态：ENTER_STANDBY 移出备用状态:EXIT_STANDBY

	Action BatchRemoveInstancesOptionAction `json:"action"`
	// 将实例移入备用状态时，是否补充新的云服务器。取值如下：no：不补充新的实例，默认情况为no。yes：补充新的实例。只有action为ENTER_STANDBY时，这个字段才生效。

	InstanceAppend *BatchRemoveInstancesOptionInstanceAppend `json:"instance_append,omitempty"`
}

func (o BatchRemoveInstancesOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchRemoveInstancesOption struct{}"
	}

	return strings.Join([]string{"BatchRemoveInstancesOption", string(data)}, " ")
}

type BatchRemoveInstancesOptionInstanceDelete struct {
	value string
}

type BatchRemoveInstancesOptionInstanceDeleteEnum struct {
	YES BatchRemoveInstancesOptionInstanceDelete
	NO  BatchRemoveInstancesOptionInstanceDelete
}

func GetBatchRemoveInstancesOptionInstanceDeleteEnum() BatchRemoveInstancesOptionInstanceDeleteEnum {
	return BatchRemoveInstancesOptionInstanceDeleteEnum{
		YES: BatchRemoveInstancesOptionInstanceDelete{
			value: "yes",
		},
		NO: BatchRemoveInstancesOptionInstanceDelete{
			value: "no",
		},
	}
}

func (c BatchRemoveInstancesOptionInstanceDelete) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchRemoveInstancesOptionInstanceDelete) UnmarshalJSON(b []byte) error {
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

type BatchRemoveInstancesOptionAction struct {
	value string
}

type BatchRemoveInstancesOptionActionEnum struct {
	REMOVE BatchRemoveInstancesOptionAction
}

func GetBatchRemoveInstancesOptionActionEnum() BatchRemoveInstancesOptionActionEnum {
	return BatchRemoveInstancesOptionActionEnum{
		REMOVE: BatchRemoveInstancesOptionAction{
			value: "REMOVE",
		},
	}
}

func (c BatchRemoveInstancesOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchRemoveInstancesOptionAction) UnmarshalJSON(b []byte) error {
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

type BatchRemoveInstancesOptionInstanceAppend struct {
	value string
}

type BatchRemoveInstancesOptionInstanceAppendEnum struct {
	NO  BatchRemoveInstancesOptionInstanceAppend
	YES BatchRemoveInstancesOptionInstanceAppend
}

func GetBatchRemoveInstancesOptionInstanceAppendEnum() BatchRemoveInstancesOptionInstanceAppendEnum {
	return BatchRemoveInstancesOptionInstanceAppendEnum{
		NO: BatchRemoveInstancesOptionInstanceAppend{
			value: "no",
		},
		YES: BatchRemoveInstancesOptionInstanceAppend{
			value: "yes",
		},
	}
}

func (c BatchRemoveInstancesOptionInstanceAppend) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchRemoveInstancesOptionInstanceAppend) UnmarshalJSON(b []byte) error {
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
