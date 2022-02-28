package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type RemoveNodeRequest struct {
	// 集群 ID，获取方式请参见[[如何获取接口URI中参数](https://support.huaweicloud.com/api-cce/cce_02_0271.html)](tag:hws)[[如何获取接口URI中参数](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0271.html)](tag:hws_hk)

	ClusterId string `json:"cluster_id"`

	Body *RemoveNodesTask `json:"body,omitempty"`
}

func (o RemoveNodeRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "RemoveNodeRequest struct{}"
	}

	return strings.Join([]string{"RemoveNodeRequest", string(data)}, " ")
}
