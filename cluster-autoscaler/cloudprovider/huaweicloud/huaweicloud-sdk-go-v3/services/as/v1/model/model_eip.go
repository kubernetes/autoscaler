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

// 配置云服务器的弹性IP信息
type Eip struct {
	// 弹性IP地址类型。类型枚举值：5_bgp：全动态BGP;5_sbgp：静态BGP;5_telcom：中国电信;5_union：中国联通;详情请参见《虚拟私有云接口参考》“申请弹性公网IP”章节的“publicip”字段说明。
	IpType    EipIpType  `json:"ip_type"`
	Bandwidth *Bandwidth `json:"bandwidth"`
}

func (o Eip) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"Eip", string(data)}, " ")
}

type EipIpType struct {
	value string
}

type EipIpTypeEnum struct {
	E_5_BGP    EipIpType
	E_5_SBGP   EipIpType
	E_5_TELCOM EipIpType
	E_5_UNION  EipIpType
}

func GetEipIpTypeEnum() EipIpTypeEnum {
	return EipIpTypeEnum{
		E_5_BGP: EipIpType{
			value: "5_bgp",
		},
		E_5_SBGP: EipIpType{
			value: "5_sbgp",
		},
		E_5_TELCOM: EipIpType{
			value: "5_telcom",
		},
		E_5_UNION: EipIpType{
			value: "5_union",
		},
	}
}

func (c EipIpType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *EipIpType) UnmarshalJSON(b []byte) error {
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
