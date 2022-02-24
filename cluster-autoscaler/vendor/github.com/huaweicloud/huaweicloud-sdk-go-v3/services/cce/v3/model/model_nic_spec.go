package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 主网卡的描述信息。
type NicSpec struct {
	// 网卡所在子网的ID。

	SubnetId *string `json:"subnetId,omitempty"`
	// 主网卡的IP将通过fixedIps指定，数量不得大于创建的节点数。fixedIps或ipBlock同时只能指定一个。

	FixedIps *[]string `json:"fixedIps,omitempty"`
	// IP段的CIDR格式，创建的节点IP将属于该IP段内。fixedIps或ipBlock同时只能指定一个。

	IpBlock *string `json:"ipBlock,omitempty"`
}

func (o NicSpec) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NicSpec struct{}"
	}

	return strings.Join([]string{"NicSpec", string(data)}, " ")
}
