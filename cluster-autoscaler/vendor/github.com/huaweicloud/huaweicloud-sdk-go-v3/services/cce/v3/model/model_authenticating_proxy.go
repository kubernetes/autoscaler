package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// authenticatingProxy模式相关配置。认证模式为authenticating_proxy时必选
type AuthenticatingProxy struct {
	// authenticating_proxy模式配置的x509格式CA证书(base64编码)。当集群认证模式为authenticating_proxy时，此项必须填写。  最大长度：1M

	Ca *string `json:"ca,omitempty"`
	// authenticating_proxy模式配置的x509格式CA证书签发的客户端证书，用于kube-apiserver到扩展apiserver的认证。(base64编码)。当集群认证模式为authenticating_proxy时，此项必须填写。

	Cert *string `json:"cert,omitempty"`
	// authenticating_proxy模式配置的x509格式CA证书签发的客户端证书时对应的私钥，用于kube-apiserver到扩展apiserver的认证。Kubernetes集群使用的私钥尚不支持密码加密，请使用未加密的私钥。(base64编码)。当集群认证模式为authenticating_proxy时，此项必须填写。

	PrivateKey *string `json:"privateKey,omitempty"`
}

func (o AuthenticatingProxy) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AuthenticatingProxy struct{}"
	}

	return strings.Join([]string{"AuthenticatingProxy", string(data)}, " ")
}
