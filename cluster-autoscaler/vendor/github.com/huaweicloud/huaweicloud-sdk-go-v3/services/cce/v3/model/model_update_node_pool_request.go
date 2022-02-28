package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type UpdateNodePoolRequest struct {
	// 集群 ID，获取方式请参见[[如何获取接口URI中参数](https://support.huaweicloud.com/api-cce/cce_02_0271.html)](tag:hws)[[如何获取接口URI中参数](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0271.html)](tag:hws_hk)

	ClusterId string `json:"cluster_id"`
	// 节点池ID

	NodepoolId string `json:"nodepool_id"`
	// 集群状态兼容Error参数，用于API平滑切换。 兼容场景下，errorStatus为空则屏蔽Error状态为Deleting状态。

	ErrorStatus *string `json:"errorStatus,omitempty"`

	Body *NodePool `json:"body,omitempty"`
}

func (o UpdateNodePoolRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "UpdateNodePoolRequest struct{}"
	}

	return strings.Join([]string{"UpdateNodePoolRequest", string(data)}, " ")
}
