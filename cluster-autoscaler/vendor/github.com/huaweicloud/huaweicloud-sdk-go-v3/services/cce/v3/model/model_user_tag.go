package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type UserTag struct {
	// 云服务器标签的键。不得以\"CCE-\"或\"__type_baremetal\"开头

	Key *string `json:"key,omitempty"`
	// 云服务器标签的值

	Value *string `json:"value,omitempty"`
}

func (o UserTag) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "UserTag struct{}"
	}

	return strings.Join([]string{"UserTag", string(data)}, " ")
}
