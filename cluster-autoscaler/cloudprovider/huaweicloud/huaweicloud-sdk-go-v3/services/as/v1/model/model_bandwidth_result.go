package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 带宽信息
type BandwidthResult struct {
	// 带宽（Mbit/s）。

	Size *int32 `json:"size,omitempty"`
	// 带宽的共享类型。共享类型枚举：PER，表示独享。目前只支持独享。

	ShareType *BandwidthResultShareType `json:"share_type,omitempty"`
	// 带宽的计费类型。字段值为“bandwidth”，表示按带宽计费。字段值为“traffic”，表示按流量计费。

	ChargingMode *BandwidthResultChargingMode `json:"charging_mode,omitempty"`
	// 带宽ID，创建WHOLE类型带宽的弹性IP时指定的共享带宽。

	Id *string `json:"id,omitempty"`
}

func (o BandwidthResult) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BandwidthResult struct{}"
	}

	return strings.Join([]string{"BandwidthResult", string(data)}, " ")
}

type BandwidthResultShareType struct {
	value string
}

type BandwidthResultShareTypeEnum struct {
	PER   BandwidthResultShareType
	WHOLE BandwidthResultShareType
}

func GetBandwidthResultShareTypeEnum() BandwidthResultShareTypeEnum {
	return BandwidthResultShareTypeEnum{
		PER: BandwidthResultShareType{
			value: "PER",
		},
		WHOLE: BandwidthResultShareType{
			value: "WHOLE",
		},
	}
}

func (c BandwidthResultShareType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BandwidthResultShareType) UnmarshalJSON(b []byte) error {
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

type BandwidthResultChargingMode struct {
	value string
}

type BandwidthResultChargingModeEnum struct {
	BANDWIDTH BandwidthResultChargingMode
	TRAFFIC   BandwidthResultChargingMode
}

func GetBandwidthResultChargingModeEnum() BandwidthResultChargingModeEnum {
	return BandwidthResultChargingModeEnum{
		BANDWIDTH: BandwidthResultChargingMode{
			value: "bandwidth",
		},
		TRAFFIC: BandwidthResultChargingMode{
			value: "traffic",
		},
	}
}

func (c BandwidthResultChargingMode) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *BandwidthResultChargingMode) UnmarshalJSON(b []byte) error {
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
