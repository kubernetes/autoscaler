package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type Volume struct {
	// 磁盘大小，单位为GB  - 系统盘取值范围：40~1024 - 数据盘取值范围：100~32768

	Size int32 `json:"size"`
	// 磁盘类型，取值请参见创建云服务器 中“root_volume字段数据结构说明”。  - SAS：高IO，是指由SAS存储提供资源的磁盘类型。 - SSD：超高IO，是指由SSD存储提供资源的磁盘类型。 - SATA：普通IO，是指由SATA存储提供资源的磁盘类型。EVS已下线SATA磁盘，仅存量节点有此类型的磁盘。

	Volumetype string `json:"volumetype"`
	// 磁盘扩展参数，取值请参见[[创建云服务器](https://support.huaweicloud.com/api-ecs/zh-cn_topic_0020212668.html)](tag:hws)[[创建云服务器](https://support.huaweicloud.com/intl/zh-cn/api-ecs/zh-cn_topic_0020212668.html)](tag:hws_hk)中“extendparam”参数的描述。

	ExtendParam map[string]interface{} `json:"extendParam,omitempty"`
	// 云服务器系统盘对应的存储池的ID。仅用作专属云集群，专属分布式存储DSS的存储池ID，即dssPoolID。  获取方法请参见获取单个专属分布式存储池详情中“表3 响应参数”的ID字段。

	ClusterId *string `json:"cluster_id,omitempty"`
	// 云服务器系统盘对应的磁盘存储类型。仅用作专属云集群，固定取值为dss。

	ClusterType *string `json:"cluster_type,omitempty"`
	// - 使用SDI规格创建虚拟机时请关注该参数，如果该参数值为true，说明创建的为SCSI类型的卷 - 节点池类型为ElasticBMS时，此参数必须填写为true

	Hwpassthrough *bool `json:"hw:passthrough,omitempty"`

	Metadata *VolumeMetadata `json:"metadata,omitempty"`
}

func (o Volume) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "Volume struct{}"
	}

	return strings.Join([]string{"Volume", string(data)}, " ")
}
