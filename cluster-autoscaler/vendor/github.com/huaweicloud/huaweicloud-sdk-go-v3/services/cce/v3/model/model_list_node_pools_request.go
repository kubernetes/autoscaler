package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ListNodePoolsRequest struct {
	// 集群 ID，获取方式请参见[[如何获取接口URI中参数](https://support.huaweicloud.com/api-cce/cce_02_0271.html)](tag:hws)[[如何获取接口URI中参数](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0271.html)](tag:hws_hk)

	ClusterId string `json:"cluster_id"`
	// 集群状态兼容Error参数，用于API平滑切换。 兼容场景下，errorStatus为空则屏蔽Error状态为Deleting状态。

	ErrorStatus *string `json:"errorStatus,omitempty"`
	// 是否展示默认节点池。默认不展示，指定为“true”时展示默认节点池。

	ShowDefaultNodePool *string `json:"showDefaultNodePool,omitempty"`
}

func (o ListNodePoolsRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListNodePoolsRequest struct{}"
	}

	return strings.Join([]string{"ListNodePoolsRequest", string(data)}, " ")
}
