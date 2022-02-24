package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type ClusterExtendParam struct {
	// 集群控制节点可用区配置。CCE支持的可用区请参考[[地区和终端节点](https://developer.huaweicloud.com/endpoint?CCE)](tag:hws)[[地区和终端节点](https://developer.huaweicloud.com/intl/zh-cn/endpoint?CCE)](tag:hws_hk)获取。   - multi_az：多可用区，可选。仅使用高可用集群时才可以配置多可用区。   - 专属云计算池可用区：用于指定专属云可用区部署集群控制节点。   如果需配置专属CCE集群，该字段为必选。例如“华北四-可用区一”取值为：cn-north-4a。更多信息请参见[[什么是专属计算集群？](https://support.huaweicloud.com/productdesc-dcc/zh-cn_topic_0016310838.html)](tag:hws)[[什么是专属计算集群？](https://support.huaweicloud.com/intl/zh-cn/productdesc-dcc/zh-cn_topic_0016310838.html)](tag:hws_hk)

	ClusterAZ *string `json:"clusterAZ,omitempty"`
	// 用于指定控制节点的系统盘和数据盘使用专属分布式存储，未指定或者值为空时，默认使用EVS云硬盘。 如果配置专属CCE集群，该字段为必选，请按照如下格式设置： ``` <rootVol.dssPoolID>.<rootVol.volType>;<dataVol.dssPoolID>.<dataVol.volType> ``` 字段说明： - rootVol为系统盘；dataVol为数据盘； - dssPoolID为专属分布式存储池ID； - volType为专属分布式存储池的存储类型，如SAS、SSD。  样例：c950ee97-587c-4f24-8a74-3367e3da570f.sas;6edbc2f4-1507-44f8-ac0d-eed1d2608d38.ssd > 非专属CCE集群不支持配置该字段。

	DssMasterVolumes *string `json:"dssMasterVolumes,omitempty"`
	// 集群所属的企业项目ID。 >   - 需要开通企业项目功能后才可配置企业项目，详情请参见[[如何进入企业管理页面](https://support.huaweicloud.com/usermanual-em/zh-cn_topic_0108763975.html)](tag:hws)[[如何进入企业管理页面](https://support.huaweicloud.com/intl/zh-cn/usermanual-em/zh-cn_topic_0108763975.html)](tag:hws_hk)。 >   - 集群所属的企业项目与集群下所关联的其他云服务资源所属的企业项目必须保持一致。

	EnterpriseProjectId *string `json:"enterpriseProjectId,omitempty"`
	// 服务转发模式，支持以下两种实现： - iptables：社区传统的kube-proxy模式，完全以iptables规则的方式来实现service负载均衡。该方式最主要的问题是在服务多的时候产生太多的iptables规则，非增量式更新会引入一定的时延，大规模情况下有明显的性能问题 - ipvs：主导开发并在社区获得广泛支持的kube-proxy模式，采用增量式更新，吞吐更高，速度更快，并可以保证service更新期间连接保持不断开，适用于大规模场景。 > 此参数已废弃，若同时指定此参数和ClusterSpec下的kubeProxyMode，以ClusterSpec下的为准。

	KubeProxyMode *string `json:"kubeProxyMode,omitempty"`
	// master 弹性公网IP

	ClusterExternalIP *string `json:"clusterExternalIP,omitempty"`
	// 容器网络固定IP池掩码位数，仅vpc-router网络支持。  该参数决定节点可分配容器IP数量，与创建节点时设置的maxPods参数共同决定节点最多可以创建多少个Pod， [具体请参见[节点最多可以创建多少Pod](https://support.huaweicloud.com/usermanual-cce/cce_01_0348.html)](tag:hws) [具体请参见[节点最多可以创建多少Pod](https://support.huaweicloud.com/intl/zh-cn/usermanual-cce/cce_01_0348.html)](tag:hws_hk)。   整数字符传取值范围: 24 ~ 28

	AlphaCceFixPoolMask *string `json:"alpha.cce/fixPoolMask,omitempty"`
	// 专属CCE集群指定可控制节点的规格。

	DecMasterFlavor *string `json:"decMasterFlavor,omitempty"`
	// 集群默认Docker的UmaskMode配置，可取值为secure或normal，不指定时默认为normal。

	DockerUmaskMode *string `json:"dockerUmaskMode,omitempty"`
	// 集群CPU管理策略。取值为none或static，默认为none。 - none：关闭工作负载实例独占CPU核的功能，优点是CPU共享池的可分配核数较多 - static：支持给节点上的工作负载实例配置CPU独占，适用于对CPU缓存和调度延迟敏感的工作负载，Turbo集群下仅对普通容器节点有效，安全容器节点无效。

	KubernetesIoCpuManagerPolicy *string `json:"kubernetes.io/cpuManagerPolicy,omitempty"`
	// 订单ID，集群付费类型为自动付费包周期类型时，响应中会返回此字段。

	OrderID *string `json:"orderID,omitempty"`
	// - month：月 - year：年 > billingMode为1（包周期）时生效，且为必选。

	PeriodType *string `json:"periodType,omitempty"`
	// 订购周期数，取值范围： - periodType=month（周期类型为月）时，取值为[1-9]。 - periodType=year（周期类型为年）时，取值为1。 > billingMode为1时生效，且为必选。

	PeriodNum *int32 `json:"periodNum,omitempty"`
	// 是否自动续订 - “true”：自动续订 - “false”：不自动续订 > billingMode为1时生效，不填写此参数时默认不会自动续费。

	IsAutoRenew *string `json:"isAutoRenew,omitempty"`
	// 是否自动扣款 - “true”：自动扣款 - “false”：不自动扣款 > billingMode为1时生效，不填写此参数时默认不会自动扣款。

	IsAutoPay *string `json:"isAutoPay,omitempty"`
	// 记录集群通过何种升级方式升级到当前版本。

	Upgradefrom *string `json:"upgradefrom,omitempty"`
}

func (o ClusterExtendParam) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ClusterExtendParam struct{}"
	}

	return strings.Join([]string{"ClusterExtendParam", string(data)}, " ")
}
