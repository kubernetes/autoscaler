/*
 * As
 *
 * 弹性伸缩API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 创建伸缩配置请求
type CreateScalingConfigRequestBody struct {
	// 伸缩配置名称(1-64个字符)，只能包含中文、字母、数字、下划线或中划线。
	ScalingConfigurationName string          `json:"scaling_configuration_name"`
	InstanceConfig           *InstanceConfig `json:"instance_config"`
}

func (o CreateScalingConfigRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingConfigRequestBody", string(data)}, " ")
}
