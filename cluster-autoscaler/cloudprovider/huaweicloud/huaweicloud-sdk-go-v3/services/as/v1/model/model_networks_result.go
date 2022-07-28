package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 网络信息
type NetworksResult struct {
	// 网络ID

	Id *string `json:"id,omitempty"`
	// 是否启用IPv6。取值为true时，标识此网卡已启用IPv6。

	Ipv6Enable *bool `json:"ipv6_enable,omitempty"`

	Ipv6Bandwidth *Ipv6Bandwidth `json:"ipv6_bandwidth,omitempty"`
	// 是否开启源/目的检查开关。

	AllowedAddressPairs *[]AllowedAddressPair `json:"allowed_address_pairs,omitempty"`
}

func (o NetworksResult) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NetworksResult struct{}"
	}

	return strings.Join([]string{"NetworksResult", string(data)}, " ")
}
