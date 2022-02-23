package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 增强型负载均衡器信息
type LbaasListener struct {
	// 监听器ID。

	ListenerId *string `json:"listener_id,omitempty"`
	// 后端云服务器组ID。

	PoolId *string `json:"pool_id,omitempty"`
	// 后端协议端口，指后端云服务器监听的端口。

	ProtocolPort *int32 `json:"protocol_port,omitempty"`
	// 权重，指后端云服务器分发得到请求的数量比例。

	Weight *int32 `json:"weight,omitempty"`
}

func (o LbaasListener) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "LbaasListener struct{}"
	}

	return strings.Join([]string{"LbaasListener", string(data)}, " ")
}
