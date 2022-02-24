package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type ClusterCert struct {
	// 服务器地址。

	Server *string `json:"server,omitempty"`
	// 证书授权数据。

	CertificateAuthorityData *string `json:"certificate-authority-data,omitempty"`
	// 不校验服务端证书，在 cluster 类型为 externalCluster 时，该值为 true。

	InsecureSkipTlsVerify *bool `json:"insecure-skip-tls-verify,omitempty"`
}

func (o ClusterCert) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ClusterCert struct{}"
	}

	return strings.Join([]string{"ClusterCert", string(data)}, " ")
}
