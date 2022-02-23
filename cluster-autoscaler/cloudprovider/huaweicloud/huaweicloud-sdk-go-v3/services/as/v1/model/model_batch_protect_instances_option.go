package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 批量设置实例保护
type BatchProtectInstancesOption struct {
	// 云服务器ID。

	InstancesId []string `json:"instances_id"`
	// 从伸缩组中移出实例时，是否删除云服务器。默认为no；可选值为yes或no。只有action为REMOVE时，这个字段才生效。

	InstanceDelete *BatchProtectInstancesOptionInstanceDelete `json:"instance_delete,omitempty"`
	// 批量操作实例action标识：添加：ADD  移除： REMOVE  设置实例保护： PROTECT  取消实例保护： UNPROTECT；转入备用状态：ENTER_STANDBY 移出备用状态:EXIT_STANDBY

	Action BatchProtectInstancesOptionAction `json:"action"`
	// 将实例移入备用状态时，是否补充新的云服务器。取值如下：no：不补充新的实例，默认情况为no。yes：补充新的实例。只有action为ENTER_STANDBY时，这个字段才生效。

	InstanceAppend *BatchProtectInstancesOptionInstanceAppend `json:"instance_append,omitempty"`
}

func (o BatchProtectInstancesOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchProtectInstancesOption struct{}"
	}

	return strings.Join([]string{"BatchProtectInstancesOption", string(data)}, " ")
}

type BatchProtectInstancesOptionInstanceDelete struct {
	value string
}

type BatchProtectInstancesOptionInstanceDeleteEnum struct {
	YES BatchProtectInstancesOptionInstanceDelete
	NO  BatchProtectInstancesOptionInstanceDelete
}

func GetBatchProtectInstancesOptionInstanceDeleteEnum() BatchProtectInstancesOptionInstanceDeleteEnum {
	return BatchProtectInstancesOptionInstanceDeleteEnum{
		YES: BatchProtectInstancesOptionInstanceDelete{
			value: "yes",
		},
		NO: BatchProtectInstancesOptionInstanceDelete{
			value: "no",
		},
	}
}

func (c BatchProtectInstancesOptionInstanceDelete) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchProtectInstancesOptionInstanceDelete) UnmarshalJSON(b []byte) error {
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

type BatchProtectInstancesOptionAction struct {
	value string
}

type BatchProtectInstancesOptionActionEnum struct {
	PROTECT BatchProtectInstancesOptionAction
}

func GetBatchProtectInstancesOptionActionEnum() BatchProtectInstancesOptionActionEnum {
	return BatchProtectInstancesOptionActionEnum{
		PROTECT: BatchProtectInstancesOptionAction{
			value: "PROTECT",
		},
	}
}

func (c BatchProtectInstancesOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchProtectInstancesOptionAction) UnmarshalJSON(b []byte) error {
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

type BatchProtectInstancesOptionInstanceAppend struct {
	value string
}

type BatchProtectInstancesOptionInstanceAppendEnum struct {
	NO  BatchProtectInstancesOptionInstanceAppend
	YES BatchProtectInstancesOptionInstanceAppend
}

func GetBatchProtectInstancesOptionInstanceAppendEnum() BatchProtectInstancesOptionInstanceAppendEnum {
	return BatchProtectInstancesOptionInstanceAppendEnum{
		NO: BatchProtectInstancesOptionInstanceAppend{
			value: "no",
		},
		YES: BatchProtectInstancesOptionInstanceAppend{
			value: "yes",
		},
	}
}

func (c BatchProtectInstancesOptionInstanceAppend) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchProtectInstancesOptionInstanceAppend) UnmarshalJSON(b []byte) error {
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
