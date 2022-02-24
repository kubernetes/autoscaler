package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type ShowResourceQuotaResponse struct {
	Quotas         *AllQuotas `json:"quotas,omitempty"`
	HttpStatusCode int        `json:"-"`
}

func (o ShowResourceQuotaResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowResourceQuotaResponse struct{}"
	}

	return strings.Join([]string{"ShowResourceQuotaResponse", string(data)}, " ")
}
