package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListHookInstancesResponse struct {
	// 伸缩实例生命周期挂钩列表。

	InstanceHangingInfo *[]InstanceHangingInfos `json:"instance_hanging_info,omitempty"`
	HttpStatusCode      int                     `json:"-"`
}

func (o ListHookInstancesResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListHookInstancesResponse struct{}"
	}

	return strings.Join([]string{"ListHookInstancesResponse", string(data)}, " ")
}
