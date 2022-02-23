package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DeleteScalingTagInfoResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteScalingTagInfoResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteScalingTagInfoResponse struct{}"
	}

	return strings.Join([]string{"DeleteScalingTagInfoResponse", string(data)}, " ")
}
