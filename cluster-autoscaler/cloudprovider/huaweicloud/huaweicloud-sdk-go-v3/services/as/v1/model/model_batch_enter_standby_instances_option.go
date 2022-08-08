package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 批量将实例转入备用状态
type BatchEnterStandbyInstancesOption struct {
	// 云服务器ID。

	InstancesId []string `json:"instances_id"`
	// 从伸缩组中移出实例时，是否删除云服务器。默认为no；可选值为yes或no。只有action为REMOVE时，这个字段才生效。

	InstanceDelete *BatchEnterStandbyInstancesOptionInstanceDelete `json:"instance_delete,omitempty"`
	// 批量操作实例action标识：添加：ADD  移除： REMOVE  设置实例保护： PROTECT  取消实例保护： UNPROTECT；转入备用状态：ENTER_STANDBY 移出备用状态:EXIT_STANDBY

	Action BatchEnterStandbyInstancesOptionAction `json:"action"`
	// 将实例移入备用状态时，是否补充新的云服务器。取值如下：no：不补充新的实例，默认情况为no。yes：补充新的实例。只有action为ENTER_STANDBY时，这个字段才生效。

	InstanceAppend *BatchEnterStandbyInstancesOptionInstanceAppend `json:"instance_append,omitempty"`
}

func (o BatchEnterStandbyInstancesOption) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchEnterStandbyInstancesOption struct{}"
	}

	return strings.Join([]string{"BatchEnterStandbyInstancesOption", string(data)}, " ")
}

type BatchEnterStandbyInstancesOptionInstanceDelete struct {
	value string
}

type BatchEnterStandbyInstancesOptionInstanceDeleteEnum struct {
	YES BatchEnterStandbyInstancesOptionInstanceDelete
	NO  BatchEnterStandbyInstancesOptionInstanceDelete
}

func GetBatchEnterStandbyInstancesOptionInstanceDeleteEnum() BatchEnterStandbyInstancesOptionInstanceDeleteEnum {
	return BatchEnterStandbyInstancesOptionInstanceDeleteEnum{
		YES: BatchEnterStandbyInstancesOptionInstanceDelete{
			value: "yes",
		},
		NO: BatchEnterStandbyInstancesOptionInstanceDelete{
			value: "no",
		},
	}
}

func (c BatchEnterStandbyInstancesOptionInstanceDelete) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchEnterStandbyInstancesOptionInstanceDelete) UnmarshalJSON(b []byte) error {
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

type BatchEnterStandbyInstancesOptionAction struct {
	value string
}

type BatchEnterStandbyInstancesOptionActionEnum struct {
	ENTER_STANDBY BatchEnterStandbyInstancesOptionAction
}

func GetBatchEnterStandbyInstancesOptionActionEnum() BatchEnterStandbyInstancesOptionActionEnum {
	return BatchEnterStandbyInstancesOptionActionEnum{
		ENTER_STANDBY: BatchEnterStandbyInstancesOptionAction{
			value: "ENTER_STANDBY",
		},
	}
}

func (c BatchEnterStandbyInstancesOptionAction) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchEnterStandbyInstancesOptionAction) UnmarshalJSON(b []byte) error {
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

type BatchEnterStandbyInstancesOptionInstanceAppend struct {
	value string
}

type BatchEnterStandbyInstancesOptionInstanceAppendEnum struct {
	NO  BatchEnterStandbyInstancesOptionInstanceAppend
	YES BatchEnterStandbyInstancesOptionInstanceAppend
}

func GetBatchEnterStandbyInstancesOptionInstanceAppendEnum() BatchEnterStandbyInstancesOptionInstanceAppendEnum {
	return BatchEnterStandbyInstancesOptionInstanceAppendEnum{
		NO: BatchEnterStandbyInstancesOptionInstanceAppend{
			value: "no",
		},
		YES: BatchEnterStandbyInstancesOptionInstanceAppend{
			value: "yes",
		},
	}
}

func (c BatchEnterStandbyInstancesOptionInstanceAppend) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BatchEnterStandbyInstancesOptionInstanceAppend) UnmarshalJSON(b []byte) error {
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
