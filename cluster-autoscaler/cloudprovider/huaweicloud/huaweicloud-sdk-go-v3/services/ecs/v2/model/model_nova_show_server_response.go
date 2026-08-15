package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type NovaShowServerResponse struct {
	Server         *NovaServer `json:"server,omitempty"`
	HttpStatusCode int         `json:"-"`
}

func (o NovaShowServerResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NovaShowServerResponse struct{}"
	}

	return strings.Join([]string{"NovaShowServerResponse", string(data)}, " ")
}
