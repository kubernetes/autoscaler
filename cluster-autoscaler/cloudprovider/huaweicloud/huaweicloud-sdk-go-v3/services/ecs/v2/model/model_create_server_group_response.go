package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type CreateServerGroupResponse struct {
	ServerGroup    *CreateServerGroupResult `json:"server_group,omitempty"`
	HttpStatusCode int                      `json:"-"`
}

func (o CreateServerGroupResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateServerGroupResponse struct{}"
	}

	return strings.Join([]string{"CreateServerGroupResponse", string(data)}, " ")
}
