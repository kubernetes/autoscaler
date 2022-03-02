package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ShowScalingConfigRequest struct {
	// 伸缩配置ID，查询唯一配置。

	ScalingConfigurationId string `json:"scaling_configuration_id"`
}

func (o ShowScalingConfigRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowScalingConfigRequest struct{}"
	}

	return strings.Join([]string{"ShowScalingConfigRequest", string(data)}, " ")
}
