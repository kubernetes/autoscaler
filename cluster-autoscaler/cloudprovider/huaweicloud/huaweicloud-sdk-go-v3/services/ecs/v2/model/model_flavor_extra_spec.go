/*
 * ecs
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 云服务器规格的扩展字段。
type FlavorExtraSpec struct {
	// 云服务器规格的分类：  - normal：通用型 - cpuv1：计算I型 - cpuv2：计算II型 - highmem：内存优化型 - gpu：GPU加速型 - entry：通用入门型 - saphana：大内存型 - ultracpu：超高性能计算型 - diskintensive：磁盘增强型 - highio：超高I/O型 - fpga：FPGA加速型  > 说明：  - 早期注册的规格该字段为hws:performancetype。
	Ecsperformancetype string `json:"ecs:performancetype"`
	// 资源类型。resource_type是为了区分云服务器的物理主机类型。
	ResourceType string `json:"resource_type"`
	// 值格式为{type}:{count}:{size}:{safeFormat}，其中：  - type指磁盘类型，当前只支持hdd。 - count指本地磁盘数量，目前支持d1类型：3/6/12/24，d2类型：2/4/8/12/16/24，d3类型：2/4/8/12/16/24/28。 - size指单个磁盘容量，单位GB，目前只支持1675（实际磁盘大小为1800，格式化后可用大小为1675）。 - safeFormat指云服务器本地磁盘是否安全格式化，目前仅支持d1类型：FALSE，d2/d3类型：True。  > 说明：  - 磁盘增强型特有字段。
	QuotalocalDisk *string `json:"quota:local_disk,omitempty"`
	// 值格式为{type}:{spec}:{size}:{safeFormat}，其中：  - type指主机上配备的nvme ssd的单卡容量大小，当前只支持1.6T/3.2T。 - spec指nvme ssd的规格，包括large/small。large表示大规格，small表示小规格。目前仅支持i3类型：large。 - size指guest使用的盘的容量大小，单位为GB。在spec值为large的情况下，此项即为host单卡大小。在spec值为small的情况下，此为1/4规格或者1/2规格。 - safeFormat指云服务器本地磁盘是否安全格式化，目前仅支持i3类型：True。  > 说明：  - 超高I/O型特有字段。
	QuotanvmeSsd *string `json:"quota:nvme_ssd,omitempty"`
	// 弹性云服务器类型的代数。  - s1：通用型I代 - s2：通用型II代 - s3：通用型 - m1：内存优化型I代 - m2：内存优化型II代 - m3：内存优化型 - h1：高性能计算型I代 - h2：高性能计算型II代 - h3：高性能计算型 - hi3：超高性能计算型 - d1：密集存储型I代 - d2：密集存储型II代 - d3：磁盘增强型 - g1：GPU加速型I代 - g2：GPU加速型II代 - f1：FPGA高性能型 - f2：FPGA通用型 - c3：通用计算增强型 - e3：大内存型 - i3：超高I/O型
	Ecsgeneration *string `json:"ecs:generation,omitempty"`
	// 虚拟化类型。  - 如果值为“FusionCompute”，表示弹性云服务器使用基于XEN的虚拟化技术。 - 如果值为“CloudCompute”，表示弹性云服务器使用基于KVM的虚拟化技术。
	EcsvirtualizationEnvTypes *string `json:"ecs:virtualization_env_types,omitempty"`
	// 显卡是否直通。  值为“true”，表示GPU直通。
	PciPassthroughenableGpu string `json:"pci_passthrough:enable_gpu"`
	// G1型和G2型云服务器应用的技术，包括GPU虚拟化和GPU直通。  - 如果该规格的云服务器使用GPU虚拟化技术，且GPU卡的型号为M60-1Q，参数值可设置为“m60_1q:virt:1”。 - 如果该规格的云服务器使用GPU直通技术，且GPU卡的型号为M60，参数值可设置为“m60:direct_graphics:1”。
	PciPassthroughgpuSpecs string `json:"pci_passthrough:gpu_specs"`
	// P1型v本地直通GPU的型号和数量，参数值可设置为“nvidia-p100:1”，表示使用该规格创建的弹性云服务器将占用1张NVIDIA P100显卡。
	PciPassthroughalias string `json:"pci_passthrough:alias"`
	// 此参数是Region级配置，某个AZ没有在cond:operation:az参数中配置时默认使用此参数的取值。不配置或无此参数时等同于“normal”。取值范围：  - normal：正常商用 - abandon：下线（即不显示） - sellout：售罄 - obt：公测 - promotion：推荐(等同normal，也是商用)
	Condoperationstatus *string `json:"cond:operation:status,omitempty"`
	// 此参数是AZ级配置，某个AZ没有在此参数中配置时默认使用cond:operation:status参数的取值。此参数的配置格式“az(xx)”。()内为某个AZ的flavor状态，()内必须要填有状态，不填为无效配置。状态的取值范围与cond:operation:status参数相同。  例如：flavor在某个region的az0正常商用，az1售罄，az2公测，az3正常商用，其他az显示下线，可配置为：  - “cond:operation:status”设置为“abandon” - “cond:operation:az”设置为“az0(normal), az1(sellout), az2(obt), az3(normal)”  > 说明：  - 如果flavor在某个AZ下的状态与cond:operation:status配置状态不同，必须配置该参数。
	Condoperationaz *string `json:"cond:operation:az,omitempty"`
	// 最大带宽  - 单位Mbps，显示为Gbps时除以1000
	QuotamaxRate *string `json:"quota:max_rate,omitempty"`
	// 基准带宽  - 单位Mbps，显示为Gbps时除以1000
	QuotaminRate *string `json:"quota:min_rate,omitempty"`
	// 内网最大收发包能力  - 单位个，显示为xx万时除以10000
	QuotamaxPps *string `json:"quota:max_pps,omitempty"`
	// 计费类型  - 计费场景，不配置时都支持 - period，包周期 - demand，按需
	Condoperationcharge *string `json:"cond:operation:charge,omitempty"`
}

func (o FlavorExtraSpec) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"FlavorExtraSpec", string(data)}, " ")
}
