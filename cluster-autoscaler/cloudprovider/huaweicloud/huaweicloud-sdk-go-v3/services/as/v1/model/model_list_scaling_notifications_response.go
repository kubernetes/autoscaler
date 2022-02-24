package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ListScalingNotificationsResponse struct {
	// 伸缩组通知列表。

	Topics         *[]Topics `json:"topics,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o ListScalingNotificationsResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListScalingNotificationsResponse struct{}"
	}

	return strings.Join([]string{"ListScalingNotificationsResponse", string(data)}, " ")
}
