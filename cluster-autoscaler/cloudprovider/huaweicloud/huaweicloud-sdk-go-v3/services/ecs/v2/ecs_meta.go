package v2

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	"net/http"
)

func GenReqDefForAddServerGroupMember(request *model.AddServerGroupMemberRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/os-server-groups/{server_group_id}/action").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForAddServerGroupMember() (*model.AddServerGroupMemberResponse, *def.HttpResponseDef) {
	resp := new(model.AddServerGroupMemberResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForAttachServerVolume(request *model.AttachServerVolumeRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/attachvolume").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForAttachServerVolume() (*model.AttachServerVolumeResponse, *def.HttpResponseDef) {
	resp := new(model.AttachServerVolumeResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForBatchAddServerNics(request *model.BatchAddServerNicsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/nics").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForBatchAddServerNics() (*model.BatchAddServerNicsResponse, *def.HttpResponseDef) {
	resp := new(model.BatchAddServerNicsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForBatchCreateServerTags(request *model.BatchCreateServerTagsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/tags/action").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForBatchCreateServerTags() (*model.BatchCreateServerTagsResponse, *def.HttpResponseDef) {
	resp := new(model.BatchCreateServerTagsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForBatchDeleteServerNics(request *model.BatchDeleteServerNicsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/nics/delete").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForBatchDeleteServerNics() (*model.BatchDeleteServerNicsResponse, *def.HttpResponseDef) {
	resp := new(model.BatchDeleteServerNicsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForBatchDeleteServerTags(request *model.BatchDeleteServerTagsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/tags/action").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForBatchDeleteServerTags() (*model.BatchDeleteServerTagsResponse, *def.HttpResponseDef) {
	resp := new(model.BatchDeleteServerTagsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForBatchRebootServers(request *model.BatchRebootServersRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/action").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForBatchRebootServers() (*model.BatchRebootServersResponse, *def.HttpResponseDef) {
	resp := new(model.BatchRebootServersResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForBatchStartServers(request *model.BatchStartServersRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/action").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForBatchStartServers() (*model.BatchStartServersResponse, *def.HttpResponseDef) {
	resp := new(model.BatchStartServersResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForBatchStopServers(request *model.BatchStopServersRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/action").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForBatchStopServers() (*model.BatchStopServersResponse, *def.HttpResponseDef) {
	resp := new(model.BatchStopServersResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForChangeServerOsWithCloudInit(request *model.ChangeServerOsWithCloudInitRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/cloudservers/{server_id}/changeos").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForChangeServerOsWithCloudInit() (*model.ChangeServerOsWithCloudInitResponse, *def.HttpResponseDef) {
	resp := new(model.ChangeServerOsWithCloudInitResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForCreatePostPaidServers(request *model.CreatePostPaidServersRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForCreatePostPaidServers() (*model.CreatePostPaidServersResponse, *def.HttpResponseDef) {
	resp := new(model.CreatePostPaidServersResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForCreateServerGroup(request *model.CreateServerGroupRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/os-server-groups").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForCreateServerGroup() (*model.CreateServerGroupResponse, *def.HttpResponseDef) {
	resp := new(model.CreateServerGroupResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForCreateServers(request *model.CreateServersRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1.1/{project_id}/cloudservers").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForCreateServers() (*model.CreateServersResponse, *def.HttpResponseDef) {
	resp := new(model.CreateServersResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForDeleteServerGroup(request *model.DeleteServerGroupRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v1/{project_id}/cloudservers/os-server-groups/{server_group_id}")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForDeleteServerGroup() (*model.DeleteServerGroupResponse, *def.HttpResponseDef) {
	resp := new(model.DeleteServerGroupResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForDeleteServerGroupMember(request *model.DeleteServerGroupMemberRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/os-server-groups/{server_group_id}/action").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_group_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForDeleteServerGroupMember() (*model.DeleteServerGroupMemberResponse, *def.HttpResponseDef) {
	resp := new(model.DeleteServerGroupMemberResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForDeleteServerMetadata(request *model.DeleteServerMetadataRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/metadata/{key}")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("key").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForDeleteServerMetadata() (*model.DeleteServerMetadataResponse, *def.HttpResponseDef) {
	resp := new(model.DeleteServerMetadataResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForDeleteServers(request *model.DeleteServersRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/delete").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForDeleteServers() (*model.DeleteServersResponse, *def.HttpResponseDef) {
	resp := new(model.DeleteServersResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForDetachServerVolume(request *model.DetachServerVolumeRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/detachvolume/{volume_id}")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("volume_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("delete_flag").
		WithLocationType(def.Query))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForDetachServerVolume() (*model.DetachServerVolumeResponse, *def.HttpResponseDef) {
	resp := new(model.DetachServerVolumeResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForListFlavors(request *model.ListFlavorsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v1/{project_id}/cloudservers/flavors")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("availability_zone").
		WithLocationType(def.Query))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForListFlavors() (*model.ListFlavorsResponse, *def.HttpResponseDef) {
	resp := new(model.ListFlavorsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForListResizeFlavors(request *model.ListResizeFlavorsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v1/{project_id}/cloudservers/resize_flavors")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("instance_uuid").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("sort_dir").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("sort_key").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("source_flavor_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("source_flavor_name").
		WithLocationType(def.Query))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForListResizeFlavors() (*model.ListResizeFlavorsResponse, *def.HttpResponseDef) {
	resp := new(model.ListResizeFlavorsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForListServerBlockDevices(request *model.ListServerBlockDevicesRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/block_device")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForListServerBlockDevices() (*model.ListServerBlockDevicesResponse, *def.HttpResponseDef) {
	resp := new(model.ListServerBlockDevicesResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForListServerInterfaces(request *model.ListServerInterfacesRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/os-interface")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForListServerInterfaces() (*model.ListServerInterfacesResponse, *def.HttpResponseDef) {
	resp := new(model.ListServerInterfacesResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForListServersDetails(request *model.ListServersDetailsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v1/{project_id}/cloudservers/detail")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("enterprise_project_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("flavor").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ip").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("not-tags").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("offset").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("reservation_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("status").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("tags").
		WithLocationType(def.Query))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForListServersDetails() (*model.ListServersDetailsResponse, *def.HttpResponseDef) {
	resp := new(model.ListServersDetailsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaAssociateSecurityGroup(request *model.NovaAssociateSecurityGroupRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2.1/{project_id}/servers/{server_id}/action").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaAssociateSecurityGroup() (*model.NovaAssociateSecurityGroupResponse, *def.HttpResponseDef) {
	resp := new(model.NovaAssociateSecurityGroupResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaCreateKeypair(request *model.NovaCreateKeypairRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2.1/{project_id}/os-keypairs").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("OpenStack-API-Version").
		WithLocationType(def.Header))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaCreateKeypair() (*model.NovaCreateKeypairResponse, *def.HttpResponseDef) {
	resp := new(model.NovaCreateKeypairResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaCreateServers(request *model.NovaCreateServersRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2.1/{project_id}/servers").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("OpenStack-API-Version").
		WithLocationType(def.Header))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaCreateServers() (*model.NovaCreateServersResponse, *def.HttpResponseDef) {
	resp := new(model.NovaCreateServersResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaDeleteKeypair(request *model.NovaDeleteKeypairRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v2.1/{project_id}/os-keypairs/{keypair_name}")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("keypair_name").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaDeleteKeypair() (*model.NovaDeleteKeypairResponse, *def.HttpResponseDef) {
	resp := new(model.NovaDeleteKeypairResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaDeleteServer(request *model.NovaDeleteServerRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodDelete).
		WithPath("/v2.1/{project_id}/servers/{server_id}")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaDeleteServer() (*model.NovaDeleteServerResponse, *def.HttpResponseDef) {
	resp := new(model.NovaDeleteServerResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaDisassociateSecurityGroup(request *model.NovaDisassociateSecurityGroupRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2.1/{project_id}/servers/{server_id}/action").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaDisassociateSecurityGroup() (*model.NovaDisassociateSecurityGroupResponse, *def.HttpResponseDef) {
	resp := new(model.NovaDisassociateSecurityGroupResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaListAvailabilityZones(request *model.NovaListAvailabilityZonesRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2.1/{project_id}/os-availability-zone")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaListAvailabilityZones() (*model.NovaListAvailabilityZonesResponse, *def.HttpResponseDef) {
	resp := new(model.NovaListAvailabilityZonesResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaListKeypairs(request *model.NovaListKeypairsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2.1/{project_id}/os-keypairs")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("marker").
		WithLocationType(def.Query))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("OpenStack-API-Version").
		WithLocationType(def.Header))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaListKeypairs() (*model.NovaListKeypairsResponse, *def.HttpResponseDef) {
	resp := new(model.NovaListKeypairsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaListServerSecurityGroups(request *model.NovaListServerSecurityGroupsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2.1/{project_id}/servers/{server_id}/os-security-groups")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaListServerSecurityGroups() (*model.NovaListServerSecurityGroupsResponse, *def.HttpResponseDef) {
	resp := new(model.NovaListServerSecurityGroupsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaListServersDetails(request *model.NovaListServersDetailsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2.1/{project_id}/servers/detail")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("changes-since").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("flavor").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("image").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ip").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("limit").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("marker").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("name").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("not-tags").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("reservation_id").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("sort_key").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("status").
		WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("tags").
		WithLocationType(def.Query))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("OpenStack-API-Version").
		WithLocationType(def.Header))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaListServersDetails() (*model.NovaListServersDetailsResponse, *def.HttpResponseDef) {
	resp := new(model.NovaListServersDetailsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForNovaShowServer(request *model.NovaShowServerRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2.1/{project_id}/servers/{server_id}")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("OpenStack-API-Version").
		WithLocationType(def.Header))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForNovaShowServer() (*model.NovaShowServerResponse, *def.HttpResponseDef) {
	resp := new(model.NovaShowServerResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForReinstallServerWithCloudInit(request *model.ReinstallServerWithCloudInitRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v2/{project_id}/cloudservers/{server_id}/reinstallos").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForReinstallServerWithCloudInit() (*model.ReinstallServerWithCloudInitResponse, *def.HttpResponseDef) {
	resp := new(model.ReinstallServerWithCloudInitResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForResetServerPassword(request *model.ResetServerPasswordRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/os-reset-password").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForResetServerPassword() (*model.ResetServerPasswordResponse, *def.HttpResponseDef) {
	resp := new(model.ResetServerPasswordResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForResizePostPaidServer(request *model.ResizePostPaidServerRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/resize").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForResizePostPaidServer() (*model.ResizePostPaidServerResponse, *def.HttpResponseDef) {
	resp := new(model.ResizePostPaidServerResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForResizeServer(request *model.ResizeServerRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1.1/{project_id}/cloudservers/{server_id}/resize").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForResizeServer() (*model.ResizeServerResponse, *def.HttpResponseDef) {
	resp := new(model.ResizeServerResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForShowResetPasswordFlag(request *model.ShowResetPasswordFlagRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/os-resetpwd-flag")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForShowResetPasswordFlag() (*model.ShowResetPasswordFlagResponse, *def.HttpResponseDef) {
	resp := new(model.ShowResetPasswordFlagResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForShowServer(request *model.ShowServerRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v1/{project_id}/cloudservers/{server_id}")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForShowServer() (*model.ShowServerResponse, *def.HttpResponseDef) {
	resp := new(model.ShowServerResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForShowServerLimits(request *model.ShowServerLimitsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v1/{project_id}/cloudservers/limits")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForShowServerLimits() (*model.ShowServerLimitsResponse, *def.HttpResponseDef) {
	resp := new(model.ShowServerLimitsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForShowServerRemoteConsole(request *model.ShowServerRemoteConsoleRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/remote_console").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForShowServerRemoteConsole() (*model.ShowServerRemoteConsoleResponse, *def.HttpResponseDef) {
	resp := new(model.ShowServerRemoteConsoleResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForShowServerTags(request *model.ShowServerTagsRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/tags")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForShowServerTags() (*model.ShowServerTagsResponse, *def.HttpResponseDef) {
	resp := new(model.ShowServerTagsResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForUpdateServer(request *model.UpdateServerRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPut).
		WithPath("/v1/{project_id}/cloudservers/{server_id}").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForUpdateServer() (*model.UpdateServerResponse, *def.HttpResponseDef) {
	resp := new(model.UpdateServerResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForUpdateServerMetadata(request *model.UpdateServerMetadataRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodPost).
		WithPath("/v1/{project_id}/cloudservers/{server_id}/metadata").
		WithContentType("application/json;charset=UTF-8")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("server_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithBodyJson(request.Body)

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForUpdateServerMetadata() (*model.UpdateServerMetadataResponse, *def.HttpResponseDef) {
	resp := new(model.UpdateServerMetadataResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}

func GenReqDefForShowJob(request *model.ShowJobRequest) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v1/{project_id}/jobs/{job_id}")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("job_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("project_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("domain_id").
		WithLocationType(def.Path))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GenRespForShowJob() (*model.ShowJobResponse, *def.HttpResponseDef) {
	resp := new(model.ShowJobResponse)
	respDefBuilder := def.NewHttpResponseDefBuilder().WithBodyJson(resp)
	responseDef := respDefBuilder.Build()
	return resp, responseDef
}
