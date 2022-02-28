package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// CCE资源标签
type ResourceTag struct {
	// Key值。 - 支持最大长度为36个UTF-8字符。 - 不支持特殊字符[\\=\\*\\<\\>\\\\\\,\\|/]+ - 不支持ASCII控制字符(0-31)

	Key *string `json:"key,omitempty"`
	// Value值。 - 支持最大长度为43个UTF-8字符。 - 不支持特殊字符[\\=\\*\\<\\>\\\\\\,\\|/]+ - 不支持ASCII控制字符(0-31)

	Value *string `json:"value,omitempty"`
}

func (o ResourceTag) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ResourceTag struct{}"
	}

	return strings.Join([]string{"ResourceTag", string(data)}, " ")
}
