/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 增强型负载均衡器信息
type LbaasListenersResult struct {
	// 监听器ID
	ListenersId *string `json:"listeners_id,omitempty"`
	// 后端云服务器组ID
	PoolId *string `json:"pool_id,omitempty"`
	// 后端协议号，指后端云服务器监听的端口，取值范围[1,65535]
	ProtocolPort *int32 `json:"protocol_port,omitempty"`
	// 权重，指后端云服务器经分发得到的请求数量比例，取值范围[0,1]，默认为1。
	Weight *int32 `json:"weight,omitempty"`
}

func (o LbaasListenersResult) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"LbaasListenersResult", string(data)}, " ")
}
