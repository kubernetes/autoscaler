package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 批量取消实例保护
type BatchUnprotectInstancesOption struct {
	// 云服务器ID。

	InstancesId []string `json:"instances_id"`
	// 从伸缩组中移出实例时，是否删除云服务器。默认为no；可选值为yes或no。只有action为REMOVE时，这个字段才生效。

	InstanceDelete *BatchUnprotectInstancesOptionInstanceDelete `json:"instance_delete,omitempty"`
	// 批量操作实例action标识：添加：ADD  移除： REMOVE  设置实例保护： PROTECT  取消实例保护： UNPROTECT；转入备用状态：ENTER_STANDBY 移出备用状态:EXIT_STANDBY

	Action BatchUnprotectInstancesOptionAction `json:"action"`
	// 将实例移入备用状态时，是否补充新的云服务器。取值如下：no：不补充新的实例，默认情况为no。yes：补充新的实例。只有action为ENTER_STANDBY时，这个字段才生效。

	InstanceAppend *BatchUnprotectInstancesOptionInstanceAppend `json:"instance_append,omitempty"`
}

func (o BatchUnprotectInstancesOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchUnprotectInstancesOption struct{}"
	}

	return strings.Join([]string{"BatchUnprotectInstancesOption", string(data)}, " ")
}

type BatchUnprotectInstancesOptionInstanceDelete struct {
	value string
}

type BatchUnprotectInstancesOptionInstanceDeleteEnum struct {
	YES BatchUnprotectInstancesOptionInstanceDelete
	NO  BatchUnprotectInstancesOptionInstanceDelete
}

func GetBatchUnprotectInstancesOptionInstanceDeleteEnum() BatchUnprotectInstancesOptionInstanceDeleteEnum {
	return BatchUnprotectInstancesOptionInstanceDeleteEnum{
		YES: BatchUnprotectInstancesOptionInstanceDelete{
			value: "yes",
		},
		NO: BatchUnprotectInstancesOptionInstanceDelete{
			value: "no",
		},
	}
}

func (c BatchUnprotectInstancesOptionInstanceDelete) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchUnprotectInstancesOptionInstanceDelete) UnmarshalJSON(b []byte) error {
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

type BatchUnprotectInstancesOptionAction struct {
	value string
}

type BatchUnprotectInstancesOptionActionEnum struct {
	UNPROTECT BatchUnprotectInstancesOptionAction
}

func GetBatchUnprotectInstancesOptionActionEnum() BatchUnprotectInstancesOptionActionEnum {
	return BatchUnprotectInstancesOptionActionEnum{
		UNPROTECT: BatchUnprotectInstancesOptionAction{
			value: "UNPROTECT",
		},
	}
}

func (c BatchUnprotectInstancesOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchUnprotectInstancesOptionAction) UnmarshalJSON(b []byte) error {
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

type BatchUnprotectInstancesOptionInstanceAppend struct {
	value string
}

type BatchUnprotectInstancesOptionInstanceAppendEnum struct {
	NO  BatchUnprotectInstancesOptionInstanceAppend
	YES BatchUnprotectInstancesOptionInstanceAppend
}

func GetBatchUnprotectInstancesOptionInstanceAppendEnum() BatchUnprotectInstancesOptionInstanceAppendEnum {
	return BatchUnprotectInstancesOptionInstanceAppendEnum{
		NO: BatchUnprotectInstancesOptionInstanceAppend{
			value: "no",
		},
		YES: BatchUnprotectInstancesOptionInstanceAppend{
			value: "yes",
		},
	}
}

func (c BatchUnprotectInstancesOptionInstanceAppend) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchUnprotectInstancesOptionInstanceAppend) UnmarshalJSON(b []byte) error {
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
