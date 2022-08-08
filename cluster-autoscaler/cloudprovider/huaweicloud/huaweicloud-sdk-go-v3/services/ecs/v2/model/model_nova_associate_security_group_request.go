package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type NovaAssociateSecurityGroupRequest struct {
	// 弹性云服务器ID。

	ServerId string `json:"server_id"`

	Body *NovaAssociateSecurityGroupRequestBody `json:"body,omitempty"`
}

func (o NovaAssociateSecurityGroupRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NovaAssociateSecurityGroupRequest struct{}"
	}

	return strings.Join([]string{"NovaAssociateSecurityGroupRequest", string(data)}, " ")
}
