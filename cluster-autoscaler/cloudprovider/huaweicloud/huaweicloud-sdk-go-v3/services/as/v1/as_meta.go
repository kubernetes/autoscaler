package v1

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"
	"net/http"
)

func GenReqDefForBatchDeleteScalingConfigs() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_configurations").
		WithResponse(new(model.BatchDeleteScalingConfigsResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCompleteLifecycleAction() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_instance_hook/{scaling_group_id}/callback").
		WithResponse(new(model.CompleteLifecycleActionResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateLifyCycleHook() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_lifecycle_hook/{scaling_group_id}").
		WithResponse(new(model.CreateLifyCycleHookResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateScalingConfig() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_configuration").
		WithResponse(new(model.CreateScalingConfigResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateScalingGroup() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_group").
		WithResponse(new(model.CreateScalingGroupResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateScalingNotification() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_notification/{scaling_group_id}").
		WithResponse(new(model.CreateScalingNotificationResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateScalingPolicy() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_policy").
		WithResponse(new(model.CreateScalingPolicyResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForCreateScalingTags() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/{resource_type}/{resource_id}/tags/action").
		WithResponse(new(model.CreateScalingTagsResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceId").
		WithJsonTag("resource_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteLifecycleHook() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_lifecycle_hook/{scaling_group_id}/{lifecycle_hook_name}").
		WithResponse(new(model.DeleteLifecycleHookResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LifecycleHookName").
		WithJsonTag("lifecycle_hook_name").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteScalingConfig() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_configuration/{scaling_configuration_id}").
		WithResponse(new(model.DeleteScalingConfigResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingConfigurationId").
		WithJsonTag("scaling_configuration_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteScalingGroup() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_group/{scaling_group_id}").
		WithResponse(new(model.DeleteScalingGroupResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ForceDelete").
		WithJsonTag("force_delete").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteScalingInstance() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_group_instance/{instance_id}").
		WithResponse(new(model.DeleteScalingInstanceResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("InstanceId").
		WithJsonTag("instance_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("InstanceDelete").
		WithJsonTag("instance_delete").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteScalingNotification() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_notification/{scaling_group_id}/{topic_urn}").
		WithResponse(new(model.DeleteScalingNotificationResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("TopicUrn").
		WithJsonTag("topic_urn").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteScalingPolicy() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_policy/{scaling_policy_id}").
		WithResponse(new(model.DeleteScalingPolicyResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingPolicyId").
		WithJsonTag("scaling_policy_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForDeleteScalingTags() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/{resource_type}/{resource_id}/tags/action").
		WithResponse(new(model.DeleteScalingTagsResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceId").
		WithJsonTag("resource_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForEnableOrDisableScalingGroup() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_group/{scaling_group_id}/action").
		WithResponse(new(model.EnableOrDisableScalingGroupResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForExecuteScalingPolicy() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_policy/{scaling_policy_id}/action").
		WithResponse(new(model.ExecuteScalingPolicyResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingPolicyId").
		WithJsonTag("scaling_policy_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListHookInstances() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_instance_hook/{scaling_group_id}/list").
		WithResponse(new(model.ListHookInstancesResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("InstanceId").
		WithJsonTag("instance_id").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListLifeCycleHooks() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_lifecycle_hook/{scaling_group_id}/list").
		WithResponse(new(model.ListLifeCycleHooksResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListResourceInstances() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/{resource_type}/resource_instances/action").
		WithResponse(new(model.ListResourceInstancesResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListScalingActivityLogs() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_activity_log/{scaling_group_id}").
		WithResponse(new(model.ListScalingActivityLogsResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("StartTime").
		WithJsonTag("start_time").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EndTime").
		WithJsonTag("end_time").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("StartNumber").
		WithJsonTag("start_number").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListScalingConfigs() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_configuration").
		WithResponse(new(model.ListScalingConfigsResponse))

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingConfigurationName").
		WithJsonTag("scaling_configuration_name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ImageId").
		WithJsonTag("image_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("StartNumber").
		WithJsonTag("start_number").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListScalingGroups() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_group").
		WithResponse(new(model.ListScalingGroupsResponse))

	// request

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupName").
		WithJsonTag("scaling_group_name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingConfigurationId").
		WithJsonTag("scaling_configuration_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupStatus").
		WithJsonTag("scaling_group_status").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("StartNumber").
		WithJsonTag("start_number").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListScalingInstances() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_group_instance/{scaling_group_id}/list").
		WithResponse(new(model.ListScalingInstancesResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LifeCycleState").
		WithJsonTag("life_cycle_state").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("HealthStatus").
		WithJsonTag("health_status").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ProtectFromScalingDown").
		WithJsonTag("protect_from_scaling_down").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("StartNumber").
		WithJsonTag("start_number").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListScalingNotifications() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_notification/{scaling_group_id}").
		WithResponse(new(model.ListScalingNotificationsResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListScalingPolicies() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_policy/{scaling_group_id}/list").
		WithResponse(new(model.ListScalingPoliciesResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingPolicyName").
		WithJsonTag("scaling_policy_name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingPolicyType").
		WithJsonTag("scaling_policy_type").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingPolicyId").
		WithJsonTag("scaling_policy_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("StartNumber").
		WithJsonTag("start_number").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListScalingPolicyExecuteLogs() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_policy_execute_log/{scaling_policy_id}").
		WithResponse(new(model.ListScalingPolicyExecuteLogsResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingPolicyId").
		WithJsonTag("scaling_policy_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LogId").
		WithJsonTag("log_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingResourceType").
		WithJsonTag("scaling_resource_type").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingResourceId").
		WithJsonTag("scaling_resource_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ExecuteType").
		WithJsonTag("execute_type").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("StartTime").
		WithJsonTag("start_time").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("EndTime").
		WithJsonTag("end_time").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("StartNumber").
		WithJsonTag("start_number").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Limit").
		WithJsonTag("limit").
		WithLocationType(def.Query))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListScalingTagInfosByResourceId() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/{resource_type}/{resource_id}/tags").
		WithResponse(new(model.ListScalingTagInfosByResourceIdResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceId").
		WithJsonTag("resource_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForListScalingTagInfosByTenantId() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/{resource_type}/tags").
		WithResponse(new(model.ListScalingTagInfosByTenantIdResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ResourceType").
		WithJsonTag("resource_type").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowLifeCycleHook() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_lifecycle_hook/{scaling_group_id}/{lifecycle_hook_name}").
		WithResponse(new(model.ShowLifeCycleHookResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LifecycleHookName").
		WithJsonTag("lifecycle_hook_name").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowPolicyAndInstanceQuota() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/quotas/{scaling_group_id}").
		WithResponse(new(model.ShowPolicyAndInstanceQuotaResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowResourceQuota() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/quotas").
		WithResponse(new(model.ShowResourceQuotaResponse))

	// request

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowScalingConfig() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_configuration/{scaling_configuration_id}").
		WithResponse(new(model.ShowScalingConfigResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingConfigurationId").
		WithJsonTag("scaling_configuration_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowScalingGroup() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_group/{scaling_group_id}").
		WithResponse(new(model.ShowScalingGroupResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForShowScalingPolicy() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_policy/{scaling_policy_id}").
		WithResponse(new(model.ShowScalingPolicyResponse))

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingPolicyId").
		WithJsonTag("scaling_policy_id").
		WithLocationType(def.Path))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateLifeCycleHook() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_lifecycle_hook/{scaling_group_id}/{lifecycle_hook_name}").
		WithResponse(new(model.UpdateLifeCycleHookResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("LifecycleHookName").
		WithJsonTag("lifecycle_hook_name").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateScalingGroup() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_group/{scaling_group_id}").
		WithResponse(new(model.UpdateScalingGroupResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateScalingGroupInstance() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_group_instance/{scaling_group_id}/action").
		WithResponse(new(model.UpdateScalingGroupInstanceResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingGroupId").
		WithJsonTag("scaling_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenReqDefForUpdateScalingPolicy() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/autoscaling-api/v1/{project_id}/scaling_policy/{scaling_policy_id}").
		WithResponse(new(model.UpdateScalingPolicyResponse)).
		WithContentType("application/json;charset=UTF-8")

	// request
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ScalingPolicyId").
		WithJsonTag("scaling_policy_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("Body").
		WithLocationType(def.Body))

	// response

	requestDef := reqDefBuilder.Build()
	return requestDef
}
