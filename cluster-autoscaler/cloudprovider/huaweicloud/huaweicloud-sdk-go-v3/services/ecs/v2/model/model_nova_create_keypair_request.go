package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type NovaCreateKeypairRequest struct {
	// 微版本头

	OpenStackAPIVersion *string `json:"OpenStack-API-Version,omitempty"`

	Body *NovaCreateKeypairRequestBody `json:"body,omitempty"`
}

func (o NovaCreateKeypairRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NovaCreateKeypairRequest struct{}"
	}

	return strings.Join([]string{"NovaCreateKeypairRequest", string(data)}, " ")
}
