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

// Response Object
type CreateScalingConfigResponse struct {
	// 伸缩配置ID
	ScalingConfigurationId *string `json:"scaling_configuration_id,omitempty"`
}

func (o CreateScalingConfigResponse) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"CreateScalingConfigResponse", string(data)}, " ")
}
