package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type NovaDisassociateSecurityGroupRequest struct {
	// 云服务器ID。

	ServerId string `json:"server_id"`

	Body *NovaDisassociateSecurityGroupRequestBody `json:"body,omitempty"`
}

func (o NovaDisassociateSecurityGroupRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NovaDisassociateSecurityGroupRequest struct{}"
	}

	return strings.Join([]string{"NovaDisassociateSecurityGroupRequest", string(data)}, " ")
}
