package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type NodeBandwidth struct {
	//    带宽的计费类型： - 未传该字段，表示按带宽计费。 - 字段值为空，表示按带宽计费。 - 字段值为“traffic”，表示按流量计费。 - 字段为其它值，会导致创建云服务器失败。  > -  按带宽计费：按公网传输速率（单位为Mbps）计费。当您的带宽利用率高于10%时，建议优先选择按带宽计费。  > - 按流量计费：只允许在创建按需节点时指定，按公网传输的数据总量（单位为GB）计费。当您的带宽利用率低于10%时，建议优先选择按流量计费。

	Chargemode *string `json:"chargemode,omitempty"`
	// 带宽大小，取值请参见“[[创建云服务器](https://support.huaweicloud.com/api-ecs/zh-cn_topic_0167957246.html)](tag:hws)[[创建云服务器](https://support.huaweicloud.com/intl/zh-cn/api-ecs/zh-cn_topic_0167957246.html)](tag:hws_hk) > bandwidth字段数据结构说明”表中“size”参数的描述。

	Size *int32 `json:"size,omitempty"`
	// 带宽的共享类型，取值请参见“[[创建云服务器](https://support.huaweicloud.com/api-ecs/zh-cn_topic_0167957246.html)](tag:hws)[[创建云服务器](https://support.huaweicloud.com/intl/zh-cn/api-ecs/zh-cn_topic_0167957246.html)](tag:hws_hk) > bandwidth字段数据结构说明”表中“sharetype”参数的描述。

	Sharetype *string `json:"sharetype,omitempty"`
}

func (o NodeBandwidth) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NodeBandwidth struct{}"
	}

	return strings.Join([]string{"NodeBandwidth", string(data)}, " ")
}
