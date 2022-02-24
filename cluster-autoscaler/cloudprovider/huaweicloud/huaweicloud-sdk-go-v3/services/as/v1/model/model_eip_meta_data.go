package model

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 用户自定义键值对
type EipMetaData struct {
	// 伸缩带宽策略中带宽对应的共享类型。

	MetadataBandwidthShareType *string `json:"metadata_bandwidth_share_type,omitempty"`
	// 伸缩带宽策略中带宽对应的EIP的ID。

	MetadataEipId *string `json:"metadata_eip_id,omitempty"`
	// 伸缩带宽策略中带宽对应的EIP地址。

	MetadataeipAddress *string `json:"metadataeip_address,omitempty"`
}

func (o EipMetaData) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "EipMetaData struct{}"
	}

	return strings.Join([]string{"EipMetaData", string(data)}, " ")
}
