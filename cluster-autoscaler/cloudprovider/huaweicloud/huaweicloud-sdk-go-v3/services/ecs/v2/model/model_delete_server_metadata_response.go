package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DeleteServerMetadataResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteServerMetadataResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteServerMetadataResponse struct{}"
	}

	return strings.Join([]string{"DeleteServerMetadataResponse", string(data)}, " ")
}
