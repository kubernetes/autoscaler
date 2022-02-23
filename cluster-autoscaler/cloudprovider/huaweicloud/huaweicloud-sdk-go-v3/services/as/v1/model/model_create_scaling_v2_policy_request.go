package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type CreateScalingV2PolicyRequest struct {
	Body *CreateScalingPolicyV2Option `json:"body,omitempty"`
}

func (o CreateScalingV2PolicyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateScalingV2PolicyRequest struct{}"
	}

	return strings.Join([]string{"CreateScalingV2PolicyRequest", string(data)}, " ")
}
