package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 增强型负载均衡器信息
type LbaasListenersResult struct {
	// 监听器ID

	ListenerId *string `json:"listener_id,omitempty"`
	// 后端云服务器组ID

	PoolId *string `json:"pool_id,omitempty"`
	// 后端协议号，指后端云服务器监听的端口，取值范围[1,65535]

	ProtocolPort *int32 `json:"protocol_port,omitempty"`
	// 权重，指后端云服务器经分发得到的请求数量比例，取值范围[0,1]，默认为1。

	Weight *int32 `json:"weight,omitempty"`
}

func (o LbaasListenersResult) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "LbaasListenersResult struct{}"
	}

	return strings.Join([]string{"LbaasListenersResult", string(data)}, " ")
}
