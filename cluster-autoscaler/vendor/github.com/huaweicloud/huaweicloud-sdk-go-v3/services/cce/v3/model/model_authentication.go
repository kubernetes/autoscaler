package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type Authentication struct {
	// 集群认证模式。   - kubernetes 1.11及之前版本的集群支持“x509”、“rbac”和“authenticating_proxy”，默认取值为“x509”。  - kubernetes 1.13及以上版本的集群支持“rbac”和“authenticating_proxy”，默认取值为“rbac”。

	Mode *string `json:"mode,omitempty"`

	AuthenticatingProxy *AuthenticatingProxy `json:"authenticatingProxy,omitempty"`
}

func (o Authentication) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "Authentication struct{}"
	}

	return strings.Join([]string{"Authentication", string(data)}, " ")
}
