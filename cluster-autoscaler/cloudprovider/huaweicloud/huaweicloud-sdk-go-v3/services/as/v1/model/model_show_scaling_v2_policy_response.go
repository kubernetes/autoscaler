package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ShowScalingV2PolicyResponse struct {
	ScalingPolicy  *ScalingV2PolicyDetail `json:"scaling_policy,omitempty"`
	HttpStatusCode int                    `json:"-"`
}

func (o ShowScalingV2PolicyResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowScalingV2PolicyResponse struct{}"
	}

	return strings.Join([]string{"ShowScalingV2PolicyResponse", string(data)}, " ")
}
