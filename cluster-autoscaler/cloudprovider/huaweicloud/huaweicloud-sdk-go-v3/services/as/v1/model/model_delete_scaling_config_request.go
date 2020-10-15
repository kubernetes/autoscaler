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
type DeleteScalingConfigRequest struct {
	ScalingConfigurationId string `json:"scaling_configuration_id"`
}

func (o DeleteScalingConfigRequest) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"DeleteScalingConfigRequest", string(data)}, " ")
}
