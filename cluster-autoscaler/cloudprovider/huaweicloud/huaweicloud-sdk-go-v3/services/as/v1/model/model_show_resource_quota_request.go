package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ShowResourceQuotaRequest struct {
}

func (o ShowResourceQuotaRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowResourceQuotaRequest struct{}"
	}

	return strings.Join([]string{"ShowResourceQuotaRequest", string(data)}, " ")
}
