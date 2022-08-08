package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type NovaDisassociateSecurityGroupResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o NovaDisassociateSecurityGroupResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NovaDisassociateSecurityGroupResponse struct{}"
	}

	return strings.Join([]string{"NovaDisassociateSecurityGroupResponse", string(data)}, " ")
}
