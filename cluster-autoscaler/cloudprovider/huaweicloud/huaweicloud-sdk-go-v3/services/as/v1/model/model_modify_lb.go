package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 负载均衡器
type ModifyLb struct {
	LbaasListener *LbaasListener `json:"lbaas_listener,omitempty"`
	// 经典型负载均衡器信息

	Listener *string `json:"listener,omitempty"`
	// 负载均衡器迁移失败原因。

	FailedReason *string `json:"failed_reason,omitempty"`
	// 负载均衡器迁移失败详情。

	FailedDetails *string `json:"failed_details,omitempty"`
}

func (o ModifyLb) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ModifyLb struct{}"
	}

	return strings.Join([]string{"ModifyLb", string(data)}, " ")
}
