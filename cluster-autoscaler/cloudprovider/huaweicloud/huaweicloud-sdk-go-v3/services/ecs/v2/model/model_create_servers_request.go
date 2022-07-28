package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type CreateServersRequest struct {
	Body *CreateServersRequestBody `json:"body,omitempty"`
}

func (o CreateServersRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateServersRequest struct{}"
	}

	return strings.Join([]string{"CreateServersRequest", string(data)}, " ")
}
