package v1

import (
	http_client "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"
)

type AsClient struct {
	hcClient *http_client.HcHttpClient
}

func NewAsClient(hcClient *http_client.HcHttpClient) *AsClient {
	return &AsClient{hcClient: hcClient}
}

func AsClientBuilder() *http_client.HcHttpClientBuilder {
	builder := http_client.NewHcHttpClientBuilder()
	return builder
}

//批量删除指定弹性伸缩配置。被伸缩组使用的伸缩配置不能被删除。单次最多删除伸缩配置个数为50。
func (c *AsClient) BatchDeleteScalingConfigs(request *model.BatchDeleteScalingConfigsRequest) (*model.BatchDeleteScalingConfigsResponse, error) {
	requestDef := GenReqDefForBatchDeleteScalingConfigs()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.BatchDeleteScalingConfigsResponse), nil
	}
}

//通过生命周期操作令牌或者通过实例ID和生命周期挂钩名称对伸缩实例指定的挂钩进行回调操作。如果在超时时间结束前已完成自定义操作，选择终止或继续完成生命周期操作。如果需要更多时间完成自定义操作，选择延长超时时间，实例保持等待状态的时间将增加1小时。只有实例的生命周期挂钩状态为 HANGING 时才可以进行回调操作。
func (c *AsClient) CompleteLifecycleAction(request *model.CompleteLifecycleActionRequest) (*model.CompleteLifecycleActionResponse, error) {
	requestDef := GenReqDefForCompleteLifecycleAction()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CompleteLifecycleActionResponse), nil
	}
}

//创建生命周期挂钩，可为伸缩组添加一个或多个生命周期挂钩，最多添加5个。添加生命周期挂钩后，当伸缩组进行伸缩活动时，实例将被生命周期挂钩挂起并置于等待状态（正在加入伸缩组或正在移出伸缩组），实例将保持此状态直至超时时间结束或者用户手动回调。用户能够在实例保持等待状态的时间段内执行自定义操作，例如，用户可以在新启动的实例上安装或配置软件，也可以在实例终止前从实例中下载日志文件。
func (c *AsClient) CreateLifyCycleHook(request *model.CreateLifyCycleHookRequest) (*model.CreateLifyCycleHookResponse, error) {
	requestDef := GenReqDefForCreateLifyCycleHook()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateLifyCycleHookResponse), nil
	}
}

//创建弹性伸缩配置。伸缩配置是伸缩组内实例（弹性云服务器云主机）的模板，定义了伸缩组内待添加的实例的规格数据。伸缩配置与伸缩组是解耦的，同一伸缩配置可以被多个伸缩组使用。默认最多可以创建100个伸缩配置。
func (c *AsClient) CreateScalingConfig(request *model.CreateScalingConfigRequest) (*model.CreateScalingConfigResponse, error) {
	requestDef := GenReqDefForCreateScalingConfig()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateScalingConfigResponse), nil
	}
}

//伸缩组是具有相同应用场景的实例的集合，是启停伸缩策略和进行伸缩活动的基本单位。伸缩组内定义了最大实例数、期望实例数、最小实例数、虚拟私有云、子网、负载均衡等信息。默认最多可以创建10个伸缩组。如果伸缩组配置了负载均衡，在添加或移除实例时，会自动为实例绑定或解绑负载均衡监听器。如果伸缩组使用负载均衡健康检查方式，伸缩组中的实例需要启用负载均衡器的监听端口才能通过健康检查。端口启用可在安全组中进行配置，可参考添加安全组规则进行操作。
func (c *AsClient) CreateScalingGroup(request *model.CreateScalingGroupRequest) (*model.CreateScalingGroupResponse, error) {
	requestDef := GenReqDefForCreateScalingGroup()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateScalingGroupResponse), nil
	}
}

//给弹性伸缩组配置通知功能。每调用一次该接口，伸缩组即配置一个通知主题及其通知场景，每个伸缩组最多可以增加5个主题。通知主题由用户事先在SMN创建并进行订阅，当通知主题对应的通知场景出现时，伸缩组会向用户的订阅终端发送通知。
func (c *AsClient) CreateScalingNotification(request *model.CreateScalingNotificationRequest) (*model.CreateScalingNotificationResponse, error) {
	requestDef := GenReqDefForCreateScalingNotification()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateScalingNotificationResponse), nil
	}
}

//创建弹性伸缩策略。伸缩策略定义了伸缩组内实例的扩张和收缩操作。如果执行伸缩策略造成伸缩组期望实例数与伸缩组内实例数不符，弹性伸缩会自动调整实例资源，以匹配期望实例数。当前伸缩策略支持告警触发策略，周期触发策略，定时触发策略。在策略执行具体动作中，可设置实例变化的个数，或根据当前实例的百分比数进行伸缩。
func (c *AsClient) CreateScalingPolicy(request *model.CreateScalingPolicyRequest) (*model.CreateScalingPolicyResponse, error) {
	requestDef := GenReqDefForCreateScalingPolicy()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateScalingPolicyResponse), nil
	}
}

//创建或删除指定资源的标签。每个伸缩组最多添加10个标签。
func (c *AsClient) CreateScalingTags(request *model.CreateScalingTagsRequest) (*model.CreateScalingTagsResponse, error) {
	requestDef := GenReqDefForCreateScalingTags()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.CreateScalingTagsResponse), nil
	}
}

//删除一个指定生命周期挂钩。伸缩组进行伸缩活动时，不允许删除该伸缩组内的生命周期挂钩。
func (c *AsClient) DeleteLifecycleHook(request *model.DeleteLifecycleHookRequest) (*model.DeleteLifecycleHookResponse, error) {
	requestDef := GenReqDefForDeleteLifecycleHook()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteLifecycleHookResponse), nil
	}
}

//删除一个指定弹性伸缩配置。
func (c *AsClient) DeleteScalingConfig(request *model.DeleteScalingConfigRequest) (*model.DeleteScalingConfigResponse, error) {
	requestDef := GenReqDefForDeleteScalingConfig()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteScalingConfigResponse), nil
	}
}

//删除一个指定弹性伸缩组。force_delete属性表示如果伸缩组存在ECS实例或正在进行伸缩活动，是否强制删除伸缩组并移出和释放ECS实例。默认值为no，表示不强制删除伸缩组。如果force_delete的值为no，必须满足以下两个条件，才能删除伸缩组：条件一：伸缩组没有正在进行的伸缩活动。条件二：伸缩组当前的ECS实例数量（current_instance_number）为0。如果force_delete的值为yes，伸缩组会被置于DELETING状态，拒绝接收新的伸缩活动请求，然后等待已有的伸缩活动完成，最后将伸缩组内所有ECS实例移出伸缩组（用户手动添加的ECS实例会被移出伸缩组，弹性伸缩自动创建的ECS实例会被自动删除）并删除伸缩组。
func (c *AsClient) DeleteScalingGroup(request *model.DeleteScalingGroupRequest) (*model.DeleteScalingGroupResponse, error) {
	requestDef := GenReqDefForDeleteScalingGroup()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteScalingGroupResponse), nil
	}
}

//从弹性伸缩组中移出一个指定实例。实例处于INSERVICE且移出后实例数不能小于伸缩组的最小实例数时才可以移出。当伸缩组没有伸缩活动时，才能移出实例。
func (c *AsClient) DeleteScalingInstance(request *model.DeleteScalingInstanceRequest) (*model.DeleteScalingInstanceResponse, error) {
	requestDef := GenReqDefForDeleteScalingInstance()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteScalingInstanceResponse), nil
	}
}

//删除指定的弹性伸缩组中指定的通知。
func (c *AsClient) DeleteScalingNotification(request *model.DeleteScalingNotificationRequest) (*model.DeleteScalingNotificationResponse, error) {
	requestDef := GenReqDefForDeleteScalingNotification()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteScalingNotificationResponse), nil
	}
}

//删除一个指定弹性伸缩策略。
func (c *AsClient) DeleteScalingPolicy(request *model.DeleteScalingPolicyRequest) (*model.DeleteScalingPolicyResponse, error) {
	requestDef := GenReqDefForDeleteScalingPolicy()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteScalingPolicyResponse), nil
	}
}

//创建或删除指定资源的标签。每个伸缩组最多添加10个标签。
func (c *AsClient) DeleteScalingTags(request *model.DeleteScalingTagsRequest) (*model.DeleteScalingTagsResponse, error) {
	requestDef := GenReqDefForDeleteScalingTags()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.DeleteScalingTagsResponse), nil
	}
}

//启用或停止一个指定弹性伸缩组。已停用状态的伸缩组，不会自动触发任何伸缩活动。当伸缩组正在进行伸缩活动，即使停用，正在进行的伸缩活动也不会立即停止。
func (c *AsClient) EnableOrDisableScalingGroup(request *model.EnableOrDisableScalingGroupRequest) (*model.EnableOrDisableScalingGroupResponse, error) {
	requestDef := GenReqDefForEnableOrDisableScalingGroup()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.EnableOrDisableScalingGroupResponse), nil
	}
}

//立即执行或启用或停止一个指定弹性伸缩策略。当伸缩组、伸缩策略状态处于INSERVICE时，伸缩策略才能被正确执行，否则会执行失败。
func (c *AsClient) ExecuteScalingPolicy(request *model.ExecuteScalingPolicyRequest) (*model.ExecuteScalingPolicyResponse, error) {
	requestDef := GenReqDefForExecuteScalingPolicy()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ExecuteScalingPolicyResponse), nil
	}
}

//添加生命周期挂钩后，当伸缩组进行伸缩活动时，实例将被挂钩挂起并置于等待状态，根据输入条件过滤查询弹性伸缩组中伸缩实例的挂起信息。可根据实例ID进行条件过滤查询。若不加过滤条件默认查询指定伸缩组内所有实例挂起信息。
func (c *AsClient) ListHookInstances(request *model.ListHookInstancesRequest) (*model.ListHookInstancesResponse, error) {
	requestDef := GenReqDefForListHookInstances()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListHookInstancesResponse), nil
	}
}

//根据伸缩组ID查询生命周期挂钩列表。
func (c *AsClient) ListLifeCycleHooks(request *model.ListLifeCycleHooksRequest) (*model.ListLifeCycleHooksResponse, error) {
	requestDef := GenReqDefForListLifeCycleHooks()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListLifeCycleHooksResponse), nil
	}
}

//根据项目ID查询指定资源类型的资源实例。资源、资源tag默认按照创建时间倒序。
func (c *AsClient) ListResourceInstances(request *model.ListResourceInstancesRequest) (*model.ListResourceInstancesResponse, error) {
	requestDef := GenReqDefForListResourceInstances()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListResourceInstancesResponse), nil
	}
}

//根据输入条件过滤查询伸缩活动日志。查询结果分页显示。可根据起始时间，截止时间，起始行号，记录数进行条件过滤查询。若不加过滤条件默认查询最多20条伸缩活动日志信息。
func (c *AsClient) ListScalingActivityLogs(request *model.ListScalingActivityLogsRequest) (*model.ListScalingActivityLogsResponse, error) {
	requestDef := GenReqDefForListScalingActivityLogs()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScalingActivityLogsResponse), nil
	}
}

//根据输入条件过滤查询弹性伸缩配置。查询结果分页显示。可以根据伸缩配置名称，镜像ID，起始行号，记录条数进行条件过滤查询。若不加过滤条件默认最多查询租户下20条伸缩配置信息。
func (c *AsClient) ListScalingConfigs(request *model.ListScalingConfigsRequest) (*model.ListScalingConfigsResponse, error) {
	requestDef := GenReqDefForListScalingConfigs()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScalingConfigsResponse), nil
	}
}

//根据输入条件过滤查询弹性伸缩组列表。查询结果分页显示。可根据伸缩组名称，伸缩配置ID，伸缩组状态，企业项目ID，起始行号，记录条数进行条件过滤查询。若不加过滤条件默认最多查询租户下20条伸缩组信息。
func (c *AsClient) ListScalingGroups(request *model.ListScalingGroupsRequest) (*model.ListScalingGroupsResponse, error) {
	requestDef := GenReqDefForListScalingGroups()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScalingGroupsResponse), nil
	}
}

//根据输入条件过滤查询弹性伸缩组中实例信息。查询结果分页显示。可根据实例在伸缩组中的生命周期状态，实例健康状态，实例保护状态，起始行号，记录条数进行条件过滤查询。若不加过滤条件默认查询组内最多20条实例信息
func (c *AsClient) ListScalingInstances(request *model.ListScalingInstancesRequest) (*model.ListScalingInstancesResponse, error) {
	requestDef := GenReqDefForListScalingInstances()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScalingInstancesResponse), nil
	}
}

//根据伸缩组ID查询指定弹性伸缩组的通知列表。
func (c *AsClient) ListScalingNotifications(request *model.ListScalingNotificationsRequest) (*model.ListScalingNotificationsResponse, error) {
	requestDef := GenReqDefForListScalingNotifications()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScalingNotificationsResponse), nil
	}
}

//根据输入条件过滤查询弹性伸缩策略。查询结果分页显示。可根据伸缩策略名称，策略类型，伸缩策略ID，起始行号，记录数进行条件过滤查询。若不加过滤条件默认查询租户下指定伸缩组内最多20条伸缩策略信息。
func (c *AsClient) ListScalingPolicies(request *model.ListScalingPoliciesRequest) (*model.ListScalingPoliciesResponse, error) {
	requestDef := GenReqDefForListScalingPolicies()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScalingPoliciesResponse), nil
	}
}

//根据输入条件过滤查询策略执行的历史记录。查询结果分页显示。可根据日志ID，伸缩资源类型，伸缩资源ID，策略执行类型，查询额起始，查询截止时间，查询起始行号，查询记录数进行条件过滤查询。若不加过滤条件默认查询最多20条策略执行日志信息。
func (c *AsClient) ListScalingPolicyExecuteLogs(request *model.ListScalingPolicyExecuteLogsRequest) (*model.ListScalingPolicyExecuteLogsResponse, error) {
	requestDef := GenReqDefForListScalingPolicyExecuteLogs()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScalingPolicyExecuteLogsResponse), nil
	}
}

//根据项目ID和资源ID查询指定资源类型的资源标签列表。
func (c *AsClient) ListScalingTagInfosByResourceId(request *model.ListScalingTagInfosByResourceIdRequest) (*model.ListScalingTagInfosByResourceIdResponse, error) {
	requestDef := GenReqDefForListScalingTagInfosByResourceId()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScalingTagInfosByResourceIdResponse), nil
	}
}

//根据项目ID查询指定资源类型的标签列表。
func (c *AsClient) ListScalingTagInfosByTenantId(request *model.ListScalingTagInfosByTenantIdRequest) (*model.ListScalingTagInfosByTenantIdResponse, error) {
	requestDef := GenReqDefForListScalingTagInfosByTenantId()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ListScalingTagInfosByTenantIdResponse), nil
	}
}

//根据伸缩组ID及生命周期挂钩名称查询指定的生命周期挂钩详情。
func (c *AsClient) ShowLifeCycleHook(request *model.ShowLifeCycleHookRequest) (*model.ShowLifeCycleHookResponse, error) {
	requestDef := GenReqDefForShowLifeCycleHook()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowLifeCycleHookResponse), nil
	}
}

//根据伸缩组ID查询指定弹性伸缩组下的伸缩策略和伸缩实例的配额总数及已使用配额数。
func (c *AsClient) ShowPolicyAndInstanceQuota(request *model.ShowPolicyAndInstanceQuotaRequest) (*model.ShowPolicyAndInstanceQuotaResponse, error) {
	requestDef := GenReqDefForShowPolicyAndInstanceQuota()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowPolicyAndInstanceQuotaResponse), nil
	}
}

//查询指定租户下的弹性伸缩组、伸缩配置、伸缩带宽策略、伸缩策略和伸缩实例的配额总数及已使用配额数。
func (c *AsClient) ShowResourceQuota(request *model.ShowResourceQuotaRequest) (*model.ShowResourceQuotaResponse, error) {
	requestDef := GenReqDefForShowResourceQuota()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowResourceQuotaResponse), nil
	}
}

//根据伸缩配置ID查询一个弹性伸缩配置的详细信息。
func (c *AsClient) ShowScalingConfig(request *model.ShowScalingConfigRequest) (*model.ShowScalingConfigResponse, error) {
	requestDef := GenReqDefForShowScalingConfig()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowScalingConfigResponse), nil
	}
}

//查询一个指定弹性伸缩组详情。
func (c *AsClient) ShowScalingGroup(request *model.ShowScalingGroupRequest) (*model.ShowScalingGroupResponse, error) {
	requestDef := GenReqDefForShowScalingGroup()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowScalingGroupResponse), nil
	}
}

//查询指定弹性伸缩策略信息。
func (c *AsClient) ShowScalingPolicy(request *model.ShowScalingPolicyRequest) (*model.ShowScalingPolicyResponse, error) {
	requestDef := GenReqDefForShowScalingPolicy()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.ShowScalingPolicyResponse), nil
	}
}

//修改一个指定生命周期挂钩中的信息。
func (c *AsClient) UpdateLifeCycleHook(request *model.UpdateLifeCycleHookRequest) (*model.UpdateLifeCycleHookResponse, error) {
	requestDef := GenReqDefForUpdateLifeCycleHook()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateLifeCycleHookResponse), nil
	}
}

//修改一个指定弹性伸缩组中的信息。更换伸缩组的伸缩配置，伸缩组中已经存在的使用之前伸缩配置创建的云服务器云主机不受影响。伸缩组为没有正在进行的伸缩活动时，可以修改伸缩组的子网、可用区和负载均衡配置。当伸缩组的期望实例数改变时，会触发伸缩活动加入或移出实例。期望实例数必须大于或等于最小实例数，必须小于或等于最大实例数。
func (c *AsClient) UpdateScalingGroup(request *model.UpdateScalingGroupRequest) (*model.UpdateScalingGroupResponse, error) {
	requestDef := GenReqDefForUpdateScalingGroup()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateScalingGroupResponse), nil
	}
}

//批量移出伸缩组中的实例或批量添加伸缩组外的实例。批量对伸缩组中的实例设置或取消其实例保护属性。批量将伸缩组中的实例转入或移出备用状态。
func (c *AsClient) UpdateScalingGroupInstance(request *model.UpdateScalingGroupInstanceRequest) (*model.UpdateScalingGroupInstanceResponse, error) {
	requestDef := GenReqDefForUpdateScalingGroupInstance()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateScalingGroupInstanceResponse), nil
	}
}

//修改指定弹性伸缩策略。
func (c *AsClient) UpdateScalingPolicy(request *model.UpdateScalingPolicyRequest) (*model.UpdateScalingPolicyResponse, error) {
	requestDef := GenReqDefForUpdateScalingPolicy()

	if resp, err := c.hcClient.Sync(request, requestDef); err != nil {
		return nil, err
	} else {
		return resp.(*model.UpdateScalingPolicyResponse), nil
	}
}
