package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type PrePaidServerPublicip struct {
	// 为待创建云服务器分配已有弹性IP时，分配的弹性IP的ID，UUID格式。  约束：只能分配状态（status）为DOWN的弹性IP。

	Id *string `json:"id,omitempty"`

	Eip *PrePaidServerEip `json:"eip,omitempty"`
	// 弹性公网IP随实例释放策略。  true：弹性公网IP随实例释放。 false：弹性公网IP不随实例释放。 默认值：false。

	DeleteOnTermination *bool `json:"delete_on_termination,omitempty"`
}

func (o PrePaidServerPublicip) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "PrePaidServerPublicip struct{}"
	}

	return strings.Join([]string{"PrePaidServerPublicip", string(data)}, " ")
}
