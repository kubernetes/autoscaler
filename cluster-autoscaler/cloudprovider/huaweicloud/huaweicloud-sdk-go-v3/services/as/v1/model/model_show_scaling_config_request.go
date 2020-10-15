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

// Request Object
type ShowScalingConfigRequest struct {
	ScalingConfigurationId string `json:"scaling_configuration_id"`
}

func (o ShowScalingConfigRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"ShowScalingConfigRequest", string(data)}, " ")
}
