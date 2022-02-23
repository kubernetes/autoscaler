package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ShowScalingConfigResponse struct {
	ScalingConfiguration *ScalingConfiguration `json:"scaling_configuration,omitempty"`
	HttpStatusCode       int                   `json:"-"`
}

func (o ShowScalingConfigResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowScalingConfigResponse struct{}"
	}

	return strings.Join([]string{"ShowScalingConfigResponse", string(data)}, " ")
}
