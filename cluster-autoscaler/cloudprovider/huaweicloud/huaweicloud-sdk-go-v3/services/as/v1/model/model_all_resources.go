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

// 配额资源
type AllResources struct {
	// 查询配额的类型。scaling_Group：伸缩组配额。scaling_Config：伸缩配置配额。scaling_Policy：伸缩策略配额。scaling_Instance：伸缩实例配额。bandwidth_scaling_policy：伸缩带宽策略配额。
	Type *AllResourcesType `json:"type,omitempty"`
	// 已使用的配额数量。当type为scaling_Policy和scaling_Instance时，该字段为保留字段，返回-1。可通过查询弹性伸缩策略和伸缩实例配额查询指定弹性伸缩组下的弹性伸缩策略和伸缩实例已使用的配额数量。
	Used *int32 `json:"used,omitempty"`
	// 配额总数量。
	Quota *int32 `json:"quota,omitempty"`
	// 配额上限。
	Max *int32 `json:"max,omitempty"`
}

func (o AllResources) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"AllResources", string(data)}, " ")
}

type AllResourcesType struct {
	value string
}

type AllResourcesTypeEnum struct {
	SCALING_GROUP            AllResourcesType
	SCALING_CONFIG           AllResourcesType
	SCALING_POLICY           AllResourcesType
	SCALING_INSTANCE         AllResourcesType
	BANDWIDTH_SCALING_POLICY AllResourcesType
}

func GetAllResourcesTypeEnum() AllResourcesTypeEnum {
	return AllResourcesTypeEnum{
		SCALING_GROUP: AllResourcesType{
			value: "scaling_group",
		},
		SCALING_CONFIG: AllResourcesType{
			value: "scaling_config",
		},
		SCALING_POLICY: AllResourcesType{
			value: "scaling_Policy",
		},
		SCALING_INSTANCE: AllResourcesType{
			value: "scaling_Instance",
		},
		BANDWIDTH_SCALING_POLICY: AllResourcesType{
			value: "bandwidth_scaling_policy",
		},
	}
}

func (c AllResourcesType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AllResourcesType) UnmarshalJSON(b []byte) error {
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
