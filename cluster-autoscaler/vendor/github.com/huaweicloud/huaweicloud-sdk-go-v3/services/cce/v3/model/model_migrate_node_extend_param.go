package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

type MigrateNodeExtendParam struct {
	// 节点最大允许创建的实例数(Pod)，该数量包含系统默认实例，取值范围为16~256。 该设置的目的为防止节点因管理过多实例而负载过重，请根据您的业务需要进行设置。

	MaxPods *int32 `json:"maxPods,omitempty"`
	// Docker数据盘配置项。 待迁移节点的磁盘类型须和创建时一致（即“DockerLVMConfigOverride”参数中“diskType”字段的值须和创建时一致），请确保单次接口调用时批量选择的节点磁盘类型一致。 默认配置示例如下： ``` \"DockerLVMConfigOverride\":\"dockerThinpool=vgpaas/90%VG;kubernetesLV=vgpaas/10%VG;diskType=evs;lvType=linear\" ``` 包含如下字段：   - userLV（可选）：用户空间的大小，示例格式：vgpaas/20%VG   - userPath（可选）：用户空间挂载路径，示例格式：/home/wqt-test   - diskType：磁盘类型，目前只有evs、hdd和ssd三种格式   - lvType：逻辑卷的类型，目前支持linear和striped两种，示例格式：striped   - dockerThinpool：Docker盘的空间大小，示例格式：vgpaas/60%VG   - kubernetesLV：Kubelet空间大小，示例格式：vgpaas/20%VG

	DockerLVMConfigOverride *string `json:"DockerLVMConfigOverride,omitempty"`
	// 安装前执行脚本 > 输入的值需要经过Base64编码，方法为echo -n \"待编码内容\" | base64。

	AlphaCcePreInstall *string `json:"alpha.cce/preInstall,omitempty"`
	// 安装后执行脚本 > 输入的值需要经过Base64编码，方法为echo -n \"待编码内容\" | base64。

	AlphaCcePostInstall *string `json:"alpha.cce/postInstall,omitempty"`
	// 指定待切换目标操作系统所使用的用户镜像ID。 当指定“alpha.cce/NodeImageID”参数时，“os”参数必须和用户自定义镜像的操作系统一致。 镜像需满足条件：[使用私有镜像制作工作节点镜像（公测）](https://support.huaweicloud.com/bestpractice-cce/cce_bestpractice_00026.html)或[使用私有镜像制作Turbo集群共池裸机工作节点镜像](https://support.huaweicloud.com/bestpractice-cce/cce_bestpractice_0134.html)

	AlphaCceNodeImageID *string `json:"alpha.cce/NodeImageID,omitempty"`
}

func (o MigrateNodeExtendParam) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "MigrateNodeExtendParam struct{}"
	}

	return strings.Join([]string{"MigrateNodeExtendParam", string(data)}, " ")
}
