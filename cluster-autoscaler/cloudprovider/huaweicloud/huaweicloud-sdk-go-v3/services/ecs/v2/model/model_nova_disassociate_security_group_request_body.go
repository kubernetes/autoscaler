package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// This is a auto create Body Object
type NovaDisassociateSecurityGroupRequestBody struct {
	RemoveSecurityGroup *NovaRemoveSecurityGroupOption `json:"removeSecurityGroup"`
}

func (o NovaDisassociateSecurityGroupRequestBody) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NovaDisassociateSecurityGroupRequestBody struct{}"
	}

	return strings.Join([]string{"NovaDisassociateSecurityGroupRequestBody", string(data)}, " ")
}
