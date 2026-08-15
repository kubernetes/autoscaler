package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type NovaListAvailabilityZonesRequest struct {
}

func (o NovaListAvailabilityZonesRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NovaListAvailabilityZonesRequest struct{}"
	}

	return strings.Join([]string{"NovaListAvailabilityZonesRequest", string(data)}, " ")
}
