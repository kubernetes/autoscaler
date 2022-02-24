package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 安全组信息
type SecurityGroups struct {
	// 安全组ID

	Id string `json:"id"`
}

func (o SecurityGroups) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "SecurityGroups struct{}"
	}

	return strings.Join([]string{"SecurityGroups", string(data)}, " ")
}
