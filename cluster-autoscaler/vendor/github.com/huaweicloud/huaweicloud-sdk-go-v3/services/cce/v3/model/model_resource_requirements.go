package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type ResourceRequirements struct {
	// 资源限制，创建时指定无效

	Limits map[string]string `json:"limits,omitempty"`
	// 资源需求，创建时指定无效

	Requests map[string]string `json:"requests,omitempty"`
}

func (o ResourceRequirements) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResourceRequirements struct{}"
	}

	return strings.Join([]string{"ResourceRequirements", string(data)}, " ")
}
