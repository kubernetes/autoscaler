package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type CreateScalingConfigResponse struct {
	// 伸缩配置ID

	ScalingConfigurationId *string `json:"scaling_configuration_id,omitempty"`
	HttpStatusCode         int     `json:"-"`
}

func (o CreateScalingConfigResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateScalingConfigResponse struct{}"
	}

	return strings.Join([]string{"CreateScalingConfigResponse", string(data)}, " ")
}
