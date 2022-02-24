package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ListAddonInstancesRequest struct {
	// 含义：想要筛选的插件名称  属性：隐藏参数

	AddonTemplateName *string `json:"addon_template_name,omitempty"`
	// 集群 ID，获取方式请参见[[如何获取接口URI中参数](https://support.huaweicloud.com/api-cce/cce_02_0271.html)](tag:hws)[[如何获取接口URI中参数](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0271.html)](tag:hws_hk)

	ClusterId string `json:"cluster_id"`
}

func (o ListAddonInstancesRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListAddonInstancesRequest struct{}"
	}

	return strings.Join([]string{"ListAddonInstancesRequest", string(data)}, " ")
}
