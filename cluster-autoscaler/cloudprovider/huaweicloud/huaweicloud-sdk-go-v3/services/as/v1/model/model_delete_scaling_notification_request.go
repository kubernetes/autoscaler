package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type DeleteScalingNotificationRequest struct {
	// 伸缩组标识。

	ScalingGroupId string `json:"scaling_group_id"`
	// SMN服务中Topic的唯一的资源标识。

	TopicUrn string `json:"topic_urn"`
}

func (o DeleteScalingNotificationRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteScalingNotificationRequest struct{}"
	}

	return strings.Join([]string{"DeleteScalingNotificationRequest", string(data)}, " ")
}
