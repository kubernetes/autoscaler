package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type NovaListKeypairsResult struct {
	Keypair *NovaSimpleKeypair `json:"keypair"`
}

func (o NovaListKeypairsResult) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NovaListKeypairsResult struct{}"
	}

	return strings.Join([]string{"NovaListKeypairsResult", string(data)}, " ")
}
