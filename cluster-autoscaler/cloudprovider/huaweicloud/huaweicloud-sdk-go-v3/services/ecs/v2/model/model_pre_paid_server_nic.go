/*
 * ecs
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 待创建云服务器的网卡信息。
type PrePaidServerNic struct {
	// 待创建云服务器的网卡信息。   需要指定vpcid对应VPC下已创建的网络（network）的ID，UUID格式。
	SubnetId string `json:"subnet_id"`
	// 待创建云服务器网卡的IP地址，IPv4格式。  约束：  - 不填或空字符串，默认在子网（subnet）中自动分配一个未使用的IP作网卡的IP地址。 - 若指定IP地址，该IP地址必须在子网（subnet）对应的网段内，且未被使用。
	IpAddress *string `json:"ip_address,omitempty"`
	// 是否支持ipv6。  取值为true时，标识此网卡支持ipv6。
	Ipv6Enable    *bool                       `json:"ipv6_enable,omitempty"`
	Ipv6Bandwidth *PrePaidServerIpv6Bandwidth `json:"ipv6_bandwidth,omitempty"`
}

func (o PrePaidServerNic) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"PrePaidServerNic", string(data)}, " ")
}
