package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DeleteLifecycleHookResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteLifecycleHookResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteLifecycleHookResponse struct{}"
	}

	return strings.Join([]string{"DeleteLifecycleHookResponse", string(data)}, " ")
}
