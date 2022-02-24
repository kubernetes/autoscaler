package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type NovaDeleteKeypairRequest struct {
	// 密钥名称。

	KeypairName string `json:"keypair_name"`
}

func (o NovaDeleteKeypairRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NovaDeleteKeypairRequest struct{}"
	}

	return strings.Join([]string{"NovaDeleteKeypairRequest", string(data)}, " ")
}
