package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DeleteScalingNotificationResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteScalingNotificationResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteScalingNotificationResponse struct{}"
	}

	return strings.Join([]string{"DeleteScalingNotificationResponse", string(data)}, " ")
}
