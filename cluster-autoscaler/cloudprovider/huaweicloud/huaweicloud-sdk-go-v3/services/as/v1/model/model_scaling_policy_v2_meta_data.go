package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 附件信息
type ScalingPolicyV2MetaData struct {
	// 伸缩带宽策略中带宽对应的共享类型

	MetadataBandwidthShareType *string `json:"metadata_bandwidth_share_type,omitempty"`
	// 伸缩带宽策略中带宽对应的EIP的ID

	MetadataEipId *string `json:"metadata_eip_id,omitempty"`
	// 伸缩带宽策略中带宽对应的EIP地址

	MetadataEipAddress *string `json:"metadata_eip_address,omitempty"`
}

func (o ScalingPolicyV2MetaData) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ScalingPolicyV2MetaData struct{}"
	}

	return strings.Join([]string{"ScalingPolicyV2MetaData", string(data)}, " ")
}
