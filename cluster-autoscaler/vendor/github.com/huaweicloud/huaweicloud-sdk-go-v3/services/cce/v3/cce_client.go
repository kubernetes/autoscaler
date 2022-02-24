package v3

import (
	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
)

type CceClient struct {
	HcClient *http_client.HcHttpClient
}

func NewCceClient(hcClient *http_client.HcHttpClient) *CceClient {
	return &CceClient{HcClient: hcClient}
}

func CceClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//该API用于在指定集群下纳管节点。 >集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) AddNode(request *model.AddNodeRequest) (*model.AddNodeResponse, error) {
	requestDef := GenReqDefForAddNode()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AddNodeResponse), nil
	}
}

//集群唤醒用于唤醒已休眠的集群，唤醒后，将继续收取控制节点资源费用。
func (c *CceClient) AwakeCluster(request *model.AwakeClusterRequest) (*model.AwakeClusterResponse, error) {
	requestDef := GenReqDefForAwakeCluster()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.AwakeClusterResponse), nil
	}
}

//根据提供的插件模板，安装插件实例。
func (c *CceClient) CreateAddonInstance(request *model.CreateAddonInstanceRequest) (*model.CreateAddonInstanceResponse, error) {
	requestDef := GenReqDefForCreateAddonInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateAddonInstanceResponse), nil
	}
}

//该API用于在指定的Namespace下通过云存储服务中的云存储（EVS、SFS、OBS）去创建PVC（PersistentVolumeClaim）。  >存储管理的URL格式为：https://{clusterid}.Endpoint/uri。其中{clusterid}为集群ID，uri为资源路径，也即API访问的路径。如果使用https://Endpoint/uri，则必须指定请求header中的X-Cluster-ID参数。
func (c *CceClient) CreateCloudPersistentVolumeClaims(request *model.CreateCloudPersistentVolumeClaimsRequest) (*model.CreateCloudPersistentVolumeClaimsResponse, error) {
	requestDef := GenReqDefForCreateCloudPersistentVolumeClaims()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateCloudPersistentVolumeClaimsResponse), nil
	}
}

//该API用于创建一个空集群（即只有控制节点Master，没有工作节点Node）。请在调用本接口完成集群创建之后，通过[[创建节点](https://support.huaweicloud.com/api-cce/cce_02_0242.html)](tag:hws)[[创建节点](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0242.html)](tag:hws_hk)添加节点。  >   - 集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。 >   - 调用该接口创建集群时，默认不安装ICAgent。ICAgent是应用性能管理APM的采集代理，运行在应用所在的服务器上，用于实时采集探针所获取的数据，安装ICAgent是使用应用性能管理APM的前提。若需安装ICAgent，请参照[[安装ICAgent](https://support.huaweicloud.com/usermanual-apm/apm_02_0013.html)](tag:hws)[[安装ICAgent](https://support.huaweicloud.com/intl/zh-cn/usermanual-apm/apm_02_0013.html)](tag:hws_hk)。 >   - 默认情况下，一个帐户只能创建5个集群（每个Region下），如果您需要创建更多的集群，请[[提交工单](https://console.huaweicloud.com/console/#/quota)](tag:hws)[[提交工单](https://console-intl.huaweicloud.com/console/?locale=zh-cn#/quota)](tag:hws_hk)申请增加配额。
func (c *CceClient) CreateCluster(request *model.CreateClusterRequest) (*model.CreateClusterResponse, error) {
	requestDef := GenReqDefForCreateCluster()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateClusterResponse), nil
	}
}

//该API用于获取指定集群的证书信息。
func (c *CceClient) CreateKubernetesClusterCert(request *model.CreateKubernetesClusterCertRequest) (*model.CreateKubernetesClusterCertResponse, error) {
	requestDef := GenReqDefForCreateKubernetesClusterCert()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateKubernetesClusterCertResponse), nil
	}
}

//该API用于在指定集群下创建节点。 > - 若无集群，请先[[创建集群](https://support.huaweicloud.com/api-cce/cce_02_0236.html)](tag:hws)[[创建集群](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0236.html)](tag:hws_hk)。 > - 集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) CreateNode(request *model.CreateNodeRequest) (*model.CreateNodeResponse, error) {
	requestDef := GenReqDefForCreateNode()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateNodeResponse), nil
	}
}

//该API用于在指定集群下创建节点池。仅支持集群在处于可用、扩容、缩容状态时调用。1.21版本的集群创建节点池时支持绑定安全组，每个节点池最多绑定五个安全组。更新节点池的安全组后，只针对新创的pod生效，建议驱逐节点上原有的pod。  > 若无集群，请先[[创建集群](https://support.huaweicloud.com/api-cce/cce_02_0236.html)](tag:hws)[[创建集群](https://support.huaweicloud.com/intl/zh-cn/api-cce/cce_02_0236.html)](tag:hws_hk)。  > 集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径
func (c *CceClient) CreateNodePool(request *model.CreateNodePoolRequest) (*model.CreateNodePoolResponse, error) {
	requestDef := GenReqDefForCreateNodePool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateNodePoolResponse), nil
	}
}

//删除插件实例的功能。
func (c *CceClient) DeleteAddonInstance(request *model.DeleteAddonInstanceRequest) (*model.DeleteAddonInstanceResponse, error) {
	requestDef := GenReqDefForDeleteAddonInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteAddonInstanceResponse), nil
	}
}

//该API用于删除指定Namespace下的PVC（PersistentVolumeClaim）对象，并可以选择保留后端的云存储。  >存储管理的URL格式为：https://{clusterid}.Endpoint/uri。其中{clusterid}为集群ID，uri为资源路径，也即API访问的路径。如果使用https://Endpoint/uri，则必须指定请求header中的X-Cluster-ID参数。
func (c *CceClient) DeleteCloudPersistentVolumeClaims(request *model.DeleteCloudPersistentVolumeClaimsRequest) (*model.DeleteCloudPersistentVolumeClaimsResponse, error) {
	requestDef := GenReqDefForDeleteCloudPersistentVolumeClaims()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteCloudPersistentVolumeClaimsResponse), nil
	}
}

//该API用于删除一个指定的集群。 >集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) DeleteCluster(request *model.DeleteClusterRequest) (*model.DeleteClusterResponse, error) {
	requestDef := GenReqDefForDeleteCluster()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteClusterResponse), nil
	}
}

//该API用于删除指定的节点。 >集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径
func (c *CceClient) DeleteNode(request *model.DeleteNodeRequest) (*model.DeleteNodeResponse, error) {
	requestDef := GenReqDefForDeleteNode()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteNodeResponse), nil
	}
}

//该API用于删除指定的节点池。 > 集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径
func (c *CceClient) DeleteNodePool(request *model.DeleteNodePoolRequest) (*model.DeleteNodePoolResponse, error) {
	requestDef := GenReqDefForDeleteNodePool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteNodePoolResponse), nil
	}
}

//集群休眠用于将运行中的集群置于休眠状态，休眠后，将不再收取控制节点资源费用。
func (c *CceClient) HibernateCluster(request *model.HibernateClusterRequest) (*model.HibernateClusterResponse, error) {
	requestDef := GenReqDefForHibernateCluster()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.HibernateClusterResponse), nil
	}
}

//获取集群所有已安装插件实例
func (c *CceClient) ListAddonInstances(request *model.ListAddonInstancesRequest) (*model.ListAddonInstancesResponse, error) {
	requestDef := GenReqDefForListAddonInstances()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAddonInstancesResponse), nil
	}
}

//插件模板查询接口，查询插件信息。
func (c *CceClient) ListAddonTemplates(request *model.ListAddonTemplatesRequest) (*model.ListAddonTemplatesResponse, error) {
	requestDef := GenReqDefForListAddonTemplates()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListAddonTemplatesResponse), nil
	}
}

//该API用于获取指定项目下所有集群的详细信息。
func (c *CceClient) ListClusters(request *model.ListClustersRequest) (*model.ListClustersResponse, error) {
	requestDef := GenReqDefForListClusters()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListClustersResponse), nil
	}
}

//该API用于获取集群下所有节点池。 > - 集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径 > - nodepool是集群中具有相同配置的节点实例的子集。
func (c *CceClient) ListNodePools(request *model.ListNodePoolsRequest) (*model.ListNodePoolsResponse, error) {
	requestDef := GenReqDefForListNodePools()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListNodePoolsResponse), nil
	}
}

//该API用于通过集群ID获取指定集群下所有节点的详细信息。 >集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) ListNodes(request *model.ListNodesRequest) (*model.ListNodesResponse, error) {
	requestDef := GenReqDefForListNodes()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListNodesResponse), nil
	}
}

//该API用于在指定集群下迁移节点到另一集群。 >集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) MigrateNode(request *model.MigrateNodeRequest) (*model.MigrateNodeResponse, error) {
	requestDef := GenReqDefForMigrateNode()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.MigrateNodeResponse), nil
	}
}

//该API用于在指定集群下移除节点。 >集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) RemoveNode(request *model.RemoveNodeRequest) (*model.RemoveNodeResponse, error) {
	requestDef := GenReqDefForRemoveNode()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.RemoveNodeResponse), nil
	}
}

//该API用于在指定集群下重置节点。 >集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) ResetNode(request *model.ResetNodeRequest) (*model.ResetNodeResponse, error) {
	requestDef := GenReqDefForResetNode()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ResetNodeResponse), nil
	}
}

//获取插件实例详情。
func (c *CceClient) ShowAddonInstance(request *model.ShowAddonInstanceRequest) (*model.ShowAddonInstanceResponse, error) {
	requestDef := GenReqDefForShowAddonInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowAddonInstanceResponse), nil
	}
}

//该API用于获取指定集群的详细信息。 >集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) ShowCluster(request *model.ShowClusterRequest) (*model.ShowClusterResponse, error) {
	requestDef := GenReqDefForShowCluster()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowClusterResponse), nil
	}
}

//该API用于获取任务信息。通过某一任务请求下发后返回的jobID来查询指定任务的进度。 > - 集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径 > - 该接口通常使用场景为： >   - 创建、删除集群时，查询相应任务的进度。 >   - 创建、删除节点时，查询相应任务的进度。
func (c *CceClient) ShowJob(request *model.ShowJobRequest) (*model.ShowJobResponse, error) {
	requestDef := GenReqDefForShowJob()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowJobResponse), nil
	}
}

//该API用于通过节点ID获取指定节点的详细信息。 >集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) ShowNode(request *model.ShowNodeRequest) (*model.ShowNodeResponse, error) {
	requestDef := GenReqDefForShowNode()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowNodeResponse), nil
	}
}

//该API用于获取指定节点池的详细信息。 > 集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径
func (c *CceClient) ShowNodePool(request *model.ShowNodePoolRequest) (*model.ShowNodePoolResponse, error) {
	requestDef := GenReqDefForShowNodePool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowNodePoolResponse), nil
	}
}

//该API用于查询CCE服务下的资源配额。
func (c *CceClient) ShowQuotas(request *model.ShowQuotasRequest) (*model.ShowQuotasResponse, error) {
	requestDef := GenReqDefForShowQuotas()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowQuotasResponse), nil
	}
}

//更新插件实例的功能。
func (c *CceClient) UpdateAddonInstance(request *model.UpdateAddonInstanceRequest) (*model.UpdateAddonInstanceResponse, error) {
	requestDef := GenReqDefForUpdateAddonInstance()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateAddonInstanceResponse), nil
	}
}

//该API用于更新指定的集群。 >集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) UpdateCluster(request *model.UpdateClusterRequest) (*model.UpdateClusterResponse, error) {
	requestDef := GenReqDefForUpdateCluster()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateClusterResponse), nil
	}
}

//该API用于更新指定的节点。 > - 当前仅支持更新metadata下的name字段，即节点的名字。 > - 集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径。
func (c *CceClient) UpdateNode(request *model.UpdateNodeRequest) (*model.UpdateNodeResponse, error) {
	requestDef := GenReqDefForUpdateNode()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateNodeResponse), nil
	}
}

//该API用于更新指定的节点池。仅支持集群在处于可用、扩容、缩容状态时调用。  > - 集群管理的URL格式为：https://Endpoint/uri。其中uri为资源路径，也即API访问的路径  > - 当前仅支持更新节点池名称，spec下的initialNodeCount，k8sTags， taints，login，userTags与节点池的扩缩容配置相关字段。
func (c *CceClient) UpdateNodePool(request *model.UpdateNodePoolRequest) (*model.UpdateNodePoolResponse, error) {
	requestDef := GenReqDefForUpdateNodePool()

	if resp, err := c.HcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateNodePoolResponse), nil
	}
}
