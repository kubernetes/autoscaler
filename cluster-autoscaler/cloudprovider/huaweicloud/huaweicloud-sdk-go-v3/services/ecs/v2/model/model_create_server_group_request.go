package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type CreateServerGroupRequest struct {
	Body *CreateServerGroupRequestBody `json:"body,omitempty"`
}

func (o CreateServerGroupRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateServerGroupRequest struct{}"
	}

	return strings.Join([]string{"CreateServerGroupRequest", string(data)}, " ")
}
