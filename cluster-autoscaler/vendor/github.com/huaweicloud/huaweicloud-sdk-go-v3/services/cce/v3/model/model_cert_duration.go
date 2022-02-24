package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 集群证书有效期
type CertDuration struct {
	// 集群证书有效时间，单位为天，用户可申请1-10950天，若填写-1则为最大值10950天，10950天约为30年。

	Duration int32 `json:"duration"`
}

func (o CertDuration) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CertDuration struct{}"
	}

	return strings.Join([]string{"CertDuration", string(data)}, " ")
}
