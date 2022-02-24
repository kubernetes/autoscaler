package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 批量将实例移出备用状态
type BatchExitStandByInstancesOption struct {
	// 云服务器ID。

	InstancesId []string `json:"instances_id"`
	// 从伸缩组中移出实例时，是否删除云服务器。默认为no；可选值为yes或no。只有action为REMOVE时，这个字段才生效。

	InstanceDelete *BatchExitStandByInstancesOptionInstanceDelete `json:"instance_delete,omitempty"`
	// 批量操作实例action标识：添加：ADD  移除： REMOVE  设置实例保护： PROTECT  取消实例保护： UNPROTECT；转入备用状态：ENTER_STANDBY 移出备用状态:EXIT_STANDBY

	Action BatchExitStandByInstancesOptionAction `json:"action"`
	// 将实例移入备用状态时，是否补充新的云服务器。取值如下：no：不补充新的实例，默认情况为no。yes：补充新的实例。只有action为ENTER_STANDBY时，这个字段才生效。

	InstanceAppend *BatchExitStandByInstancesOptionInstanceAppend `json:"instance_append,omitempty"`
}

func (o BatchExitStandByInstancesOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchExitStandByInstancesOption struct{}"
	}

	return strings.Join([]string{"BatchExitStandByInstancesOption", string(data)}, " ")
}

type BatchExitStandByInstancesOptionInstanceDelete struct {
	value string
}

type BatchExitStandByInstancesOptionInstanceDeleteEnum struct {
	YES BatchExitStandByInstancesOptionInstanceDelete
	NO  BatchExitStandByInstancesOptionInstanceDelete
}

func GetBatchExitStandByInstancesOptionInstanceDeleteEnum() BatchExitStandByInstancesOptionInstanceDeleteEnum {
	return BatchExitStandByInstancesOptionInstanceDeleteEnum{
		YES: BatchExitStandByInstancesOptionInstanceDelete{
			value: "yes",
		},
		NO: BatchExitStandByInstancesOptionInstanceDelete{
			value: "no",
		},
	}
}

func (c BatchExitStandByInstancesOptionInstanceDelete) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchExitStandByInstancesOptionInstanceDelete) UnmarshalJSON(b []byte) error {
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

type BatchExitStandByInstancesOptionAction struct {
	value string
}

type BatchExitStandByInstancesOptionActionEnum struct {
	EXIT_STANDBY BatchExitStandByInstancesOptionAction
}

func GetBatchExitStandByInstancesOptionActionEnum() BatchExitStandByInstancesOptionActionEnum {
	return BatchExitStandByInstancesOptionActionEnum{
		EXIT_STANDBY: BatchExitStandByInstancesOptionAction{
			value: "EXIT_STANDBY",
		},
	}
}

func (c BatchExitStandByInstancesOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchExitStandByInstancesOptionAction) UnmarshalJSON(b []byte) error {
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

type BatchExitStandByInstancesOptionInstanceAppend struct {
	value string
}

type BatchExitStandByInstancesOptionInstanceAppendEnum struct {
	NO  BatchExitStandByInstancesOptionInstanceAppend
	YES BatchExitStandByInstancesOptionInstanceAppend
}

func GetBatchExitStandByInstancesOptionInstanceAppendEnum() BatchExitStandByInstancesOptionInstanceAppendEnum {
	return BatchExitStandByInstancesOptionInstanceAppendEnum{
		NO: BatchExitStandByInstancesOptionInstanceAppend{
			value: "no",
		},
		YES: BatchExitStandByInstancesOptionInstanceAppend{
			value: "yes",
		},
	}
}

func (c BatchExitStandByInstancesOptionInstanceAppend) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchExitStandByInstancesOptionInstanceAppend) UnmarshalJSON(b []byte) error {
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
