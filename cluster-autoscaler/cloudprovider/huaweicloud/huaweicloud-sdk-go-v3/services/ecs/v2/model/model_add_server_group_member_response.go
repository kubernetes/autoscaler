package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type AddServerGroupMemberResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o AddServerGroupMemberResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AddServerGroupMemberResponse struct{}"
	}

	return strings.Join([]string{"AddServerGroupMemberResponse", string(data)}, " ")
}
