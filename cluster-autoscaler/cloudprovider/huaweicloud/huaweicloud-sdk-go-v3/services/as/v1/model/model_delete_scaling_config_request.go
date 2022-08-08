package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type DeleteScalingConfigRequest struct {
	// 伸缩配置ID。

	ScalingConfigurationId string `json:"scaling_configuration_id"`
}

func (o DeleteScalingConfigRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteScalingConfigRequest struct{}"
	}

	return strings.Join([]string{"DeleteScalingConfigRequest", string(data)}, " ")
}
