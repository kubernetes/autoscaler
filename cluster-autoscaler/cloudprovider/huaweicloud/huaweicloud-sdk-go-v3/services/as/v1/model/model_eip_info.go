package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// 配置云服务器的弹性IP信息
type EipInfo struct {
	// 弹性IP地址类型。类型枚举值：5_bgp：全动态BGP;5_sbgp：静态BGP;5_telcom：中国电信;5_union：中国联通;详情请参见《虚拟私有云接口参考》“申请弹性公网IP”章节的“publicip”字段说明。

	IpType EipInfoIpType `json:"ip_type"`

	Bandwidth *BandwidthInfo `json:"bandwidth"`
}

func (o EipInfo) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "EipInfo struct{}"
	}

	return strings.Join([]string{"EipInfo", string(data)}, " ")
}

type EipInfoIpType struct {
	value string
}

type EipInfoIpTypeEnum struct {
	E_5_BGP    EipInfoIpType
	E_5_SBGP   EipInfoIpType
	E_5_TELCOM EipInfoIpType
	E_5_UNION  EipInfoIpType
}

func GetEipInfoIpTypeEnum() EipInfoIpTypeEnum {
	return EipInfoIpTypeEnum{
		E_5_BGP: EipInfoIpType{
			value: "5_bgp",
		},
		E_5_SBGP: EipInfoIpType{
			value: "5_sbgp",
		},
		E_5_TELCOM: EipInfoIpType{
			value: "5_telcom",
		},
		E_5_UNION: EipInfoIpType{
			value: "5_union",
		},
	}
}

func (c EipInfoIpType) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *EipInfoIpType) UnmarshalJSON(b []byte) error {
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
