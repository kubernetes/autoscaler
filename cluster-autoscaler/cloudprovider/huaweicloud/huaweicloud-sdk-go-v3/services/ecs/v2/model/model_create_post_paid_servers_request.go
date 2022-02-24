package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type CreatePostPaidServersRequest struct {
	Body *CreatePostPaidServersRequestBody `json:"body,omitempty"`
}

func (o CreatePostPaidServersRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreatePostPaidServersRequest struct{}"
	}

	return strings.Join([]string{"CreatePostPaidServersRequest", string(data)}, " ")
}
