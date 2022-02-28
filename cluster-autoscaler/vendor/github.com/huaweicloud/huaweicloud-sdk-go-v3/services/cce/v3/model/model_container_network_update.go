package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type ContainerNetworkUpdate struct {
	// 容器网络网段列表。1.21及新版本集群，当集群网络类型为vpc-router时，支持增量添加容器网段。  此参数在集群更新后不可更改，请谨慎选择。

	Cidrs *[]ContainerCidr `json:"cidrs,omitempty"`
}

func (o ContainerNetworkUpdate) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ContainerNetworkUpdate struct{}"
	}

	return strings.Join([]string{"ContainerNetworkUpdate", string(data)}, " ")
}
