package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 节点重装场景服务器相关配置
type ReinstallVolumeConfig struct {
	// Docker数据盘配置项。 默认配置示例如下： ``` \"lvmConfig\":\"dockerThinpool=vgpaas/90%VG;kubernetesLV=vgpaas/10%VG;diskType=evs;lvType=linear\" ``` 包含如下字段：   - userLV：用户空间的大小，示例格式：vgpaas/20%VG   - userPath：用户空间挂载路径，示例格式：/home/wqt-test   - diskType：磁盘类型，目前只有evs、hdd和ssd三种格式   - lvType：逻辑卷的类型，目前支持linear和striped两种，示例格式：striped   - dockerThinpool：Docker盘的空间大小，示例格式：vgpaas/60%VG   - kubernetesLV：Kubelet空间大小，示例格式：vgpaas/20%VG

	LvmConfig *string `json:"lvmConfig,omitempty"`

	Storage *Storage `json:"storage,omitempty"`
}

func (o ReinstallVolumeConfig) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ReinstallVolumeConfig struct{}"
	}

	return strings.Join([]string{"ReinstallVolumeConfig", string(data)}, " ")
}
