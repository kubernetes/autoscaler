package model

import (
	"errors"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"
)

type VersionInfo struct {
	// API版本ID。

	Id *VersionInfoId `json:"id,omitempty"`
	// API的URL相关信息。

	Links *[]Links `json:"links,omitempty"`
	// 该版本API支持的最小微版本号。

	MinVersion *string `json:"min_version,omitempty"`
	// 版本状态，为如下3种：CURRENT：表示该版本为主推版本；SUPPORT：表示为老版本，但是现在还继续支持；DEPRECATED：表示为废弃版本，存在后续删除的可能。

	Status *VersionInfoStatus `json:"status,omitempty"`
	// 版本发布时间，使用UTC时间。

	Update *sdktime.SdkTime `json:"update,omitempty"`
	// 该版本API支持的最大微版本号。

	Version *string `json:"version,omitempty"`
}

func (o VersionInfo) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "VersionInfo struct{}"
	}

	return strings.Join([]string{"VersionInfo", string(data)}, " ")
}

type VersionInfoId struct {
	value string
}

type VersionInfoIdEnum struct {
	V1 VersionInfoId
	V2 VersionInfoId
}

func GetVersionInfoIdEnum() VersionInfoIdEnum {
	return VersionInfoIdEnum{
		V1: VersionInfoId{
			value: "v1",
		},
		V2: VersionInfoId{
			value: "v2",
		},
	}
}

func (c VersionInfoId) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *VersionInfoId) UnmarshalJSON(b []byte) error {
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

type VersionInfoStatus struct {
	value string
}

type VersionInfoStatusEnum struct {
	CURRENT    VersionInfoStatus
	SUPPORT    VersionInfoStatus
	DEPRECATED VersionInfoStatus
}

func GetVersionInfoStatusEnum() VersionInfoStatusEnum {
	return VersionInfoStatusEnum{
		CURRENT: VersionInfoStatus{
			value: "CURRENT",
		},
		SUPPORT: VersionInfoStatus{
			value: "SUPPORT",
		},
		DEPRECATED: VersionInfoStatus{
			value: "DEPRECATED",
		},
	}
}

func (c VersionInfoStatus) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *VersionInfoStatus) UnmarshalJSON(b []byte) error {
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
