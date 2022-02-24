package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type SecurityId struct {
	// 安全组ID。

	Id *string `json:"id,omitempty"`
}

func (o SecurityId) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "SecurityId struct{}"
	}

	return strings.Join([]string{"SecurityId", string(data)}, " ")
}
