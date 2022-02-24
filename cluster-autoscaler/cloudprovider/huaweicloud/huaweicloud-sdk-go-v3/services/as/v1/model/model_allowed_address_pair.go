package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 是否开启源/目的检查开关。
type AllowedAddressPair struct {
	// 是否开启源/目的检查开关。  默认是开启，不允许置空。  关闭：1.1.1.1/0 开启：除“1.1.1.1/0”以外的其余值均按开启处理

	IpAddress *string `json:"ip_address,omitempty"`
}

func (o AllowedAddressPair) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AllowedAddressPair struct{}"
	}

	return strings.Join([]string{"AllowedAddressPair", string(data)}, " ")
}
