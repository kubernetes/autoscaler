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

// 带宽信息
type Bandwidth struct {
	// 带宽（Mbit/s），取值范围为[1,300]。
	Size int32 `json:"size"`
	// 带宽的共享类型。共享类型枚举：PER：独享型。WHOLE：共享型。
	ShareType *BandwidthShareType `json:"share_type,omitempty"`
	// 带宽的计费类型。字段值为“bandwidth”，表示按带宽计费。字段值为“traffic”，表示按流量计费。字段为其它值，会导致创建云服务器失败。如果share_type是PER，该参数为必选项。如果share_type是WHOLE，会忽略该参数。
	ChargingMode BandwidthChargingMode `json:"charging_mode"`
	// 带宽ID，使用共享型带宽时，可以选择之前创建的共享带宽来创建弹性IP。如果share_type是PER，会忽略该参数。如果share_type是WHOLE，该参数为必选项。
	Id *string `json:"id,omitempty"`
}

func (o Bandwidth) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"Bandwidth", string(data)}, " ")
}

type BandwidthShareType struct {
	value string
}

type BandwidthShareTypeEnum struct {
	PER   BandwidthShareType
	WHOLE BandwidthShareType
}

func GetBandwidthShareTypeEnum() BandwidthShareTypeEnum {
	return BandwidthShareTypeEnum{
		PER: BandwidthShareType{
			value: "PER",
		},
		WHOLE: BandwidthShareType{
			value: "WHOLE",
		},
	}
}

func (c BandwidthShareType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BandwidthShareType) UnmarshalJSON(b []byte) error {
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

type BandwidthChargingMode struct {
	value string
}

type BandwidthChargingModeEnum struct {
	BANDWIDTH BandwidthChargingMode
	TRAFFIC   BandwidthChargingMode
}

func GetBandwidthChargingModeEnum() BandwidthChargingModeEnum {
	return BandwidthChargingModeEnum{
		BANDWIDTH: BandwidthChargingMode{
			value: "bandwidth",
		},
		TRAFFIC: BandwidthChargingMode{
			value: "traffic",
		},
	}
}

func (c BandwidthChargingMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BandwidthChargingMode) UnmarshalJSON(b []byte) error {
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
