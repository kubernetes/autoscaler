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

// 批量删除伸缩配置请求
type BatchDeleteScalingConfigsRequestBody struct {
	// 伸缩配置ID。
	ScalingConfigurationId *[]string `json:"scaling_configuration_id,omitempty"`
}

func (o BatchDeleteScalingConfigsRequestBody) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"BatchDeleteScalingConfigsRequestBody", string(data)}, " ")
}
