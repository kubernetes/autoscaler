package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ShowScalingPolicyResponse struct {
	ScalingPolicy  *ScalingV1PolicyDetail `json:"scaling_policy,omitempty"`
	HttpStatusCode int                    `json:"-"`
}

func (o ShowScalingPolicyResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowScalingPolicyResponse struct{}"
	}

	return strings.Join([]string{"ShowScalingPolicyResponse", string(data)}, " ")
}
