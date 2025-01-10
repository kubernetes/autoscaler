// Copyright (c) 2016, 2018, 2024, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Kubernetes Engine API
//
// API for the Kubernetes Engine service (also known as the Container Engine for Kubernetes service). Use this API to build, deploy,
// and manage cloud-native applications. For more information, see
// Overview of Kubernetes Engine (https://docs.cloud.oracle.com/iaas/Content/ContEng/Concepts/contengoverview.htm).
//

package containerengine

import (
	"context"
	"fmt"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common/auth"
	"net/http"
)

// ContainerEngineClient a client for ContainerEngine
type ContainerEngineClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewContainerEngineClientWithConfigurationProvider Creates a new default ContainerEngine client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewContainerEngineClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client ContainerEngineClient, err error) {
	if enabled := common.CheckForEnabledServices("containerengine"); !enabled {
		return client, fmt.Errorf("the Developer Tool configuration disabled this service, this behavior is controlled by OciSdkEnabledServicesMap variables. Please check if your local developer-tool-configuration.json file configured the service you're targeting or contact the cloud provider on the availability of this service")
	}
	provider, err := auth.GetGenericConfigurationProvider(configProvider)
	if err != nil {
		return client, err
	}
	baseClient, e := common.NewClientWithConfig(provider)
	if e != nil {
		return client, e
	}
	return newContainerEngineClientFromBaseClient(baseClient, provider)
}

// NewContainerEngineClientWithOboToken Creates a new default ContainerEngine client with the given configuration provider.
// The obotoken will be added to default headers and signed; the configuration provider will be used for the signer
//
//	as well as reading the region
func NewContainerEngineClientWithOboToken(configProvider common.ConfigurationProvider, oboToken string) (client ContainerEngineClient, err error) {
	baseClient, err := common.NewClientWithOboToken(configProvider, oboToken)
	if err != nil {
		return client, err
	}

	return newContainerEngineClientFromBaseClient(baseClient, configProvider)
}

func newContainerEngineClientFromBaseClient(baseClient common.BaseClient, configProvider common.ConfigurationProvider) (client ContainerEngineClient, err error) {
	// ContainerEngine service default circuit breaker is enabled
	baseClient.Configuration.CircuitBreaker = common.NewCircuitBreaker(common.DefaultCircuitBreakerSettingWithServiceName("ContainerEngine"))
	common.ConfigCircuitBreakerFromEnvVar(&baseClient)
	common.ConfigCircuitBreakerFromGlobalVar(&baseClient)

	client = ContainerEngineClient{BaseClient: baseClient}
	client.BasePath = "20180222"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *ContainerEngineClient) SetRegion(region string) {
	client.Host = common.StringToRegion(region).EndpointForTemplate("containerengine", "https://containerengine.{region}.oci.{secondLevelDomain}")
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *ContainerEngineClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
	if ok, err := common.IsConfigurationProviderValid(configProvider); !ok {
		return err
	}

	// Error has been checked already
	region, _ := configProvider.Region()
	client.SetRegion(region)
	if client.Host == "" {
		return fmt.Errorf("invalid region or Host. Endpoint cannot be constructed without endpointServiceName or serviceEndpointTemplate for a dotted region")
	}
	client.config = &configProvider
	return nil
}

// ConfigurationProvider the ConfigurationProvider used in this client, or null if none set
func (client *ContainerEngineClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// ClusterMigrateToNativeVcn Initiates cluster migration to use native VCN.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ClusterMigrateToNativeVcn.go.html to see an example of how to use ClusterMigrateToNativeVcn API.
// A default retry strategy applies to this operation ClusterMigrateToNativeVcn()
func (client ContainerEngineClient) ClusterMigrateToNativeVcn(ctx context.Context, request ClusterMigrateToNativeVcnRequest) (response ClusterMigrateToNativeVcnResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.clusterMigrateToNativeVcn, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ClusterMigrateToNativeVcnResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ClusterMigrateToNativeVcnResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ClusterMigrateToNativeVcnResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ClusterMigrateToNativeVcnResponse")
	}
	return
}

// clusterMigrateToNativeVcn implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) clusterMigrateToNativeVcn(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPost, "/clusters/{clusterId}/actions/migrateToNativeVcn", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ClusterMigrateToNativeVcnResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ClusterMigrateToNativeVcn", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CompleteCredentialRotation Complete cluster credential rotation. Retire old credentials from kubernetes components.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/CompleteCredentialRotation.go.html to see an example of how to use CompleteCredentialRotation API.
// A default retry strategy applies to this operation CompleteCredentialRotation()
func (client ContainerEngineClient) CompleteCredentialRotation(ctx context.Context, request CompleteCredentialRotationRequest) (response CompleteCredentialRotationResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.completeCredentialRotation, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CompleteCredentialRotationResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CompleteCredentialRotationResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CompleteCredentialRotationResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CompleteCredentialRotationResponse")
	}
	return
}

// completeCredentialRotation implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) completeCredentialRotation(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPost, "/clusters/{clusterId}/actions/completeCredentialRotation", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response CompleteCredentialRotationResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "CompleteCredentialRotation", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateCluster Create a new cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/CreateCluster.go.html to see an example of how to use CreateCluster API.
// A default retry strategy applies to this operation CreateCluster()
func (client ContainerEngineClient) CreateCluster(ctx context.Context, request CreateClusterRequest) (response CreateClusterResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createCluster, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateClusterResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateClusterResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateClusterResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateClusterResponse")
	}
	return
}

// createCluster implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) createCluster(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPost, "/clusters", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response CreateClusterResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "CreateCluster", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateKubeconfig Create the Kubeconfig YAML for a cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/CreateKubeconfig.go.html to see an example of how to use CreateKubeconfig API.
// A default retry strategy applies to this operation CreateKubeconfig()
func (client ContainerEngineClient) CreateKubeconfig(ctx context.Context, request CreateKubeconfigRequest) (response CreateKubeconfigResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.createKubeconfig, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateKubeconfigResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateKubeconfigResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateKubeconfigResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateKubeconfigResponse")
	}
	return
}

// createKubeconfig implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) createKubeconfig(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPost, "/clusters/{clusterId}/kubeconfig/content", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response CreateKubeconfigResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "CreateKubeconfig", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateNodePool Create a new node pool.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/CreateNodePool.go.html to see an example of how to use CreateNodePool API.
// A default retry strategy applies to this operation CreateNodePool()
func (client ContainerEngineClient) CreateNodePool(ctx context.Context, request CreateNodePoolRequest) (response CreateNodePoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createNodePool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateNodePoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateNodePoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateNodePoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateNodePoolResponse")
	}
	return
}

// createNodePool implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) createNodePool(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPost, "/nodePools", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response CreateNodePoolResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "CreateNodePool", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateVirtualNodePool Create a new virtual node pool.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/CreateVirtualNodePool.go.html to see an example of how to use CreateVirtualNodePool API.
// A default retry strategy applies to this operation CreateVirtualNodePool()
func (client ContainerEngineClient) CreateVirtualNodePool(ctx context.Context, request CreateVirtualNodePoolRequest) (response CreateVirtualNodePoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createVirtualNodePool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateVirtualNodePoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateVirtualNodePoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateVirtualNodePoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateVirtualNodePoolResponse")
	}
	return
}

// createVirtualNodePool implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) createVirtualNodePool(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPost, "/virtualNodePools", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response CreateVirtualNodePoolResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "CreateVirtualNodePool", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateWorkloadMapping Create the specified workloadMapping for a cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/CreateWorkloadMapping.go.html to see an example of how to use CreateWorkloadMapping API.
// A default retry strategy applies to this operation CreateWorkloadMapping()
func (client ContainerEngineClient) CreateWorkloadMapping(ctx context.Context, request CreateWorkloadMappingRequest) (response CreateWorkloadMappingResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createWorkloadMapping, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateWorkloadMappingResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateWorkloadMappingResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateWorkloadMappingResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateWorkloadMappingResponse")
	}
	return
}

// createWorkloadMapping implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) createWorkloadMapping(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPost, "/clusters/{clusterId}/workloadMappings", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response CreateWorkloadMappingResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "CreateWorkloadMapping", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteCluster Delete a cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/DeleteCluster.go.html to see an example of how to use DeleteCluster API.
// A default retry strategy applies to this operation DeleteCluster()
func (client ContainerEngineClient) DeleteCluster(ctx context.Context, request DeleteClusterRequest) (response DeleteClusterResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteCluster, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteClusterResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteClusterResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteClusterResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteClusterResponse")
	}
	return
}

// deleteCluster implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) deleteCluster(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/clusters/{clusterId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response DeleteClusterResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "DeleteCluster", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteNode Delete node.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/DeleteNode.go.html to see an example of how to use DeleteNode API.
// A default retry strategy applies to this operation DeleteNode()
func (client ContainerEngineClient) DeleteNode(ctx context.Context, request DeleteNodeRequest) (response DeleteNodeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteNode, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteNodeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteNodeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteNodeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteNodeResponse")
	}
	return
}

// deleteNode implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) deleteNode(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/nodePools/{nodePoolId}/node/{nodeId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response DeleteNodeResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "DeleteNode", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteNodePool Delete a node pool.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/DeleteNodePool.go.html to see an example of how to use DeleteNodePool API.
// A default retry strategy applies to this operation DeleteNodePool()
func (client ContainerEngineClient) DeleteNodePool(ctx context.Context, request DeleteNodePoolRequest) (response DeleteNodePoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteNodePool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteNodePoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteNodePoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteNodePoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteNodePoolResponse")
	}
	return
}

// deleteNodePool implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) deleteNodePool(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/nodePools/{nodePoolId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response DeleteNodePoolResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "DeleteNodePool", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteVirtualNodePool Delete a virtual node pool.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/DeleteVirtualNodePool.go.html to see an example of how to use DeleteVirtualNodePool API.
// A default retry strategy applies to this operation DeleteVirtualNodePool()
func (client ContainerEngineClient) DeleteVirtualNodePool(ctx context.Context, request DeleteVirtualNodePoolRequest) (response DeleteVirtualNodePoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteVirtualNodePool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteVirtualNodePoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteVirtualNodePoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteVirtualNodePoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteVirtualNodePoolResponse")
	}
	return
}

// deleteVirtualNodePool implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) deleteVirtualNodePool(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/virtualNodePools/{virtualNodePoolId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response DeleteVirtualNodePoolResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "DeleteVirtualNodePool", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteWorkRequest Cancel a work request that has not started.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/DeleteWorkRequest.go.html to see an example of how to use DeleteWorkRequest API.
// A default retry strategy applies to this operation DeleteWorkRequest()
func (client ContainerEngineClient) DeleteWorkRequest(ctx context.Context, request DeleteWorkRequestRequest) (response DeleteWorkRequestResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteWorkRequest, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteWorkRequestResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteWorkRequestResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteWorkRequestResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteWorkRequestResponse")
	}
	return
}

// deleteWorkRequest implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) deleteWorkRequest(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/workRequests/{workRequestId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response DeleteWorkRequestResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "DeleteWorkRequest", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteWorkloadMapping Delete workloadMapping for a provisioned cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/DeleteWorkloadMapping.go.html to see an example of how to use DeleteWorkloadMapping API.
// A default retry strategy applies to this operation DeleteWorkloadMapping()
func (client ContainerEngineClient) DeleteWorkloadMapping(ctx context.Context, request DeleteWorkloadMappingRequest) (response DeleteWorkloadMappingResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteWorkloadMapping, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteWorkloadMappingResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteWorkloadMappingResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteWorkloadMappingResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteWorkloadMappingResponse")
	}
	return
}

// deleteWorkloadMapping implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) deleteWorkloadMapping(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/clusters/{clusterId}/workloadMappings/{workloadMappingId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response DeleteWorkloadMappingResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "DeleteWorkloadMapping", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DisableAddon Disable addon for a provisioned cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/DisableAddon.go.html to see an example of how to use DisableAddon API.
// A default retry strategy applies to this operation DisableAddon()
func (client ContainerEngineClient) DisableAddon(ctx context.Context, request DisableAddonRequest) (response DisableAddonResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.disableAddon, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DisableAddonResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DisableAddonResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DisableAddonResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DisableAddonResponse")
	}
	return
}

// disableAddon implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) disableAddon(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/clusters/{clusterId}/addons/{addonName}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response DisableAddonResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "DisableAddon", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetAddon Get the specified addon for a cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetAddon.go.html to see an example of how to use GetAddon API.
// A default retry strategy applies to this operation GetAddon()
func (client ContainerEngineClient) GetAddon(ctx context.Context, request GetAddonRequest) (response GetAddonResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getAddon, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetAddonResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetAddonResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetAddonResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetAddonResponse")
	}
	return
}

// getAddon implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getAddon(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/clusters/{clusterId}/addons/{addonName}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetAddonResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetAddon", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetCluster Get the details of a cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetCluster.go.html to see an example of how to use GetCluster API.
// A default retry strategy applies to this operation GetCluster()
func (client ContainerEngineClient) GetCluster(ctx context.Context, request GetClusterRequest) (response GetClusterResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCluster, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetClusterResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetClusterResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetClusterResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetClusterResponse")
	}
	return
}

// getCluster implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getCluster(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/clusters/{clusterId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetClusterResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetCluster", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetClusterMigrateToNativeVcnStatus Get details on a cluster's migration to native VCN.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetClusterMigrateToNativeVcnStatus.go.html to see an example of how to use GetClusterMigrateToNativeVcnStatus API.
// A default retry strategy applies to this operation GetClusterMigrateToNativeVcnStatus()
func (client ContainerEngineClient) GetClusterMigrateToNativeVcnStatus(ctx context.Context, request GetClusterMigrateToNativeVcnStatusRequest) (response GetClusterMigrateToNativeVcnStatusResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getClusterMigrateToNativeVcnStatus, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetClusterMigrateToNativeVcnStatusResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetClusterMigrateToNativeVcnStatusResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetClusterMigrateToNativeVcnStatusResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetClusterMigrateToNativeVcnStatusResponse")
	}
	return
}

// getClusterMigrateToNativeVcnStatus implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getClusterMigrateToNativeVcnStatus(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/clusters/{clusterId}/migrateToNativeVcnStatus", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetClusterMigrateToNativeVcnStatusResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetClusterMigrateToNativeVcnStatus", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetClusterOptions Get options available for clusters.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetClusterOptions.go.html to see an example of how to use GetClusterOptions API.
// A default retry strategy applies to this operation GetClusterOptions()
func (client ContainerEngineClient) GetClusterOptions(ctx context.Context, request GetClusterOptionsRequest) (response GetClusterOptionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getClusterOptions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetClusterOptionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetClusterOptionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetClusterOptionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetClusterOptionsResponse")
	}
	return
}

// getClusterOptions implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getClusterOptions(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/clusterOptions/{clusterOptionId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetClusterOptionsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetClusterOptions", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetCredentialRotationStatus Get cluster credential rotation status.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetCredentialRotationStatus.go.html to see an example of how to use GetCredentialRotationStatus API.
// A default retry strategy applies to this operation GetCredentialRotationStatus()
func (client ContainerEngineClient) GetCredentialRotationStatus(ctx context.Context, request GetCredentialRotationStatusRequest) (response GetCredentialRotationStatusResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCredentialRotationStatus, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetCredentialRotationStatusResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetCredentialRotationStatusResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetCredentialRotationStatusResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetCredentialRotationStatusResponse")
	}
	return
}

// getCredentialRotationStatus implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getCredentialRotationStatus(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/clusters/{clusterId}/credentialRotationStatus", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetCredentialRotationStatusResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetCredentialRotationStatus", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetNodePool Get the details of a node pool.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetNodePool.go.html to see an example of how to use GetNodePool API.
// A default retry strategy applies to this operation GetNodePool()
func (client ContainerEngineClient) GetNodePool(ctx context.Context, request GetNodePoolRequest) (response GetNodePoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getNodePool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetNodePoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetNodePoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetNodePoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetNodePoolResponse")
	}
	return
}

// getNodePool implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getNodePool(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/nodePools/{nodePoolId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetNodePoolResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetNodePool", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetNodePoolOptions Get options available for node pools.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetNodePoolOptions.go.html to see an example of how to use GetNodePoolOptions API.
// A default retry strategy applies to this operation GetNodePoolOptions()
func (client ContainerEngineClient) GetNodePoolOptions(ctx context.Context, request GetNodePoolOptionsRequest) (response GetNodePoolOptionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getNodePoolOptions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetNodePoolOptionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetNodePoolOptionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetNodePoolOptionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetNodePoolOptionsResponse")
	}
	return
}

// getNodePoolOptions implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getNodePoolOptions(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/nodePoolOptions/{nodePoolOptionId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetNodePoolOptionsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetNodePoolOptions", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetVirtualNode Get the details of a virtual node.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetVirtualNode.go.html to see an example of how to use GetVirtualNode API.
// A default retry strategy applies to this operation GetVirtualNode()
func (client ContainerEngineClient) GetVirtualNode(ctx context.Context, request GetVirtualNodeRequest) (response GetVirtualNodeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getVirtualNode, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetVirtualNodeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetVirtualNodeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetVirtualNodeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetVirtualNodeResponse")
	}
	return
}

// getVirtualNode implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getVirtualNode(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/virtualNodePools/{virtualNodePoolId}/virtualNodes/{virtualNodeId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetVirtualNodeResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetVirtualNode", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetVirtualNodePool Get the details of a virtual node pool.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetVirtualNodePool.go.html to see an example of how to use GetVirtualNodePool API.
// A default retry strategy applies to this operation GetVirtualNodePool()
func (client ContainerEngineClient) GetVirtualNodePool(ctx context.Context, request GetVirtualNodePoolRequest) (response GetVirtualNodePoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getVirtualNodePool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetVirtualNodePoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetVirtualNodePoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetVirtualNodePoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetVirtualNodePoolResponse")
	}
	return
}

// getVirtualNodePool implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getVirtualNodePool(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/virtualNodePools/{virtualNodePoolId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetVirtualNodePoolResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetVirtualNodePool", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetWorkRequest Get the details of a work request.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetWorkRequest.go.html to see an example of how to use GetWorkRequest API.
// A default retry strategy applies to this operation GetWorkRequest()
func (client ContainerEngineClient) GetWorkRequest(ctx context.Context, request GetWorkRequestRequest) (response GetWorkRequestResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getWorkRequest, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetWorkRequestResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetWorkRequestResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetWorkRequestResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetWorkRequestResponse")
	}
	return
}

// getWorkRequest implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getWorkRequest(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/workRequests/{workRequestId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetWorkRequestResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetWorkRequest", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetWorkloadMapping Get the specified workloadMapping for a cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/GetWorkloadMapping.go.html to see an example of how to use GetWorkloadMapping API.
// A default retry strategy applies to this operation GetWorkloadMapping()
func (client ContainerEngineClient) GetWorkloadMapping(ctx context.Context, request GetWorkloadMappingRequest) (response GetWorkloadMappingResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getWorkloadMapping, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetWorkloadMappingResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetWorkloadMappingResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetWorkloadMappingResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetWorkloadMappingResponse")
	}
	return
}

// getWorkloadMapping implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) getWorkloadMapping(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/clusters/{clusterId}/workloadMappings/{workloadMappingId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response GetWorkloadMappingResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "GetWorkloadMapping", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// InstallAddon Install the specified addon for a cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/InstallAddon.go.html to see an example of how to use InstallAddon API.
// A default retry strategy applies to this operation InstallAddon()
func (client ContainerEngineClient) InstallAddon(ctx context.Context, request InstallAddonRequest) (response InstallAddonResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.installAddon, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = InstallAddonResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = InstallAddonResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(InstallAddonResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into InstallAddonResponse")
	}
	return
}

// installAddon implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) installAddon(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPost, "/clusters/{clusterId}/addons", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response InstallAddonResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "InstallAddon", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListAddonOptions Get list of supported addons for a specific kubernetes version.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListAddonOptions.go.html to see an example of how to use ListAddonOptions API.
// A default retry strategy applies to this operation ListAddonOptions()
func (client ContainerEngineClient) ListAddonOptions(ctx context.Context, request ListAddonOptionsRequest) (response ListAddonOptionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listAddonOptions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListAddonOptionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListAddonOptionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListAddonOptionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListAddonOptionsResponse")
	}
	return
}

// listAddonOptions implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listAddonOptions(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/addonOptions", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListAddonOptionsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListAddonOptions", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListAddons List addon for a provisioned cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListAddons.go.html to see an example of how to use ListAddons API.
// A default retry strategy applies to this operation ListAddons()
func (client ContainerEngineClient) ListAddons(ctx context.Context, request ListAddonsRequest) (response ListAddonsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listAddons, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListAddonsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListAddonsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListAddonsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListAddonsResponse")
	}
	return
}

// listAddons implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listAddons(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/clusters/{clusterId}/addons", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListAddonsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListAddons", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListClusters List all the cluster objects in a compartment.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListClusters.go.html to see an example of how to use ListClusters API.
// A default retry strategy applies to this operation ListClusters()
func (client ContainerEngineClient) ListClusters(ctx context.Context, request ListClustersRequest) (response ListClustersResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listClusters, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListClustersResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListClustersResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListClustersResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListClustersResponse")
	}
	return
}

// listClusters implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listClusters(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/clusters", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListClustersResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListClusters", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListNodePools List all the node pools in a compartment, and optionally filter by cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListNodePools.go.html to see an example of how to use ListNodePools API.
// A default retry strategy applies to this operation ListNodePools()
func (client ContainerEngineClient) ListNodePools(ctx context.Context, request ListNodePoolsRequest) (response ListNodePoolsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listNodePools, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListNodePoolsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListNodePoolsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListNodePoolsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListNodePoolsResponse")
	}
	return
}

// listNodePools implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listNodePools(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/nodePools", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListNodePoolsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListNodePools", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListPodShapes List all the Pod Shapes in a compartment.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListPodShapes.go.html to see an example of how to use ListPodShapes API.
// A default retry strategy applies to this operation ListPodShapes()
func (client ContainerEngineClient) ListPodShapes(ctx context.Context, request ListPodShapesRequest) (response ListPodShapesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listPodShapes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListPodShapesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListPodShapesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListPodShapesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListPodShapesResponse")
	}
	return
}

// listPodShapes implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listPodShapes(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/podShapes", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListPodShapesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListPodShapes", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListVirtualNodePools List all the virtual node pools in a compartment, and optionally filter by cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListVirtualNodePools.go.html to see an example of how to use ListVirtualNodePools API.
// A default retry strategy applies to this operation ListVirtualNodePools()
func (client ContainerEngineClient) ListVirtualNodePools(ctx context.Context, request ListVirtualNodePoolsRequest) (response ListVirtualNodePoolsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listVirtualNodePools, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListVirtualNodePoolsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListVirtualNodePoolsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListVirtualNodePoolsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListVirtualNodePoolsResponse")
	}
	return
}

// listVirtualNodePools implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listVirtualNodePools(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/virtualNodePools", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListVirtualNodePoolsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListVirtualNodePools", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListVirtualNodes List virtual nodes in a virtual node pool.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListVirtualNodes.go.html to see an example of how to use ListVirtualNodes API.
// A default retry strategy applies to this operation ListVirtualNodes()
func (client ContainerEngineClient) ListVirtualNodes(ctx context.Context, request ListVirtualNodesRequest) (response ListVirtualNodesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listVirtualNodes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListVirtualNodesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListVirtualNodesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListVirtualNodesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListVirtualNodesResponse")
	}
	return
}

// listVirtualNodes implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listVirtualNodes(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/virtualNodePools/{virtualNodePoolId}/virtualNodes", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListVirtualNodesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListVirtualNodes", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListWorkRequestErrors Get the errors of a work request.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListWorkRequestErrors.go.html to see an example of how to use ListWorkRequestErrors API.
// A default retry strategy applies to this operation ListWorkRequestErrors()
func (client ContainerEngineClient) ListWorkRequestErrors(ctx context.Context, request ListWorkRequestErrorsRequest) (response ListWorkRequestErrorsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listWorkRequestErrors, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListWorkRequestErrorsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListWorkRequestErrorsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListWorkRequestErrorsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListWorkRequestErrorsResponse")
	}
	return
}

// listWorkRequestErrors implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listWorkRequestErrors(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/workRequests/{workRequestId}/errors", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListWorkRequestErrorsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListWorkRequestErrors", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListWorkRequestLogs Get the logs of a work request.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListWorkRequestLogs.go.html to see an example of how to use ListWorkRequestLogs API.
// A default retry strategy applies to this operation ListWorkRequestLogs()
func (client ContainerEngineClient) ListWorkRequestLogs(ctx context.Context, request ListWorkRequestLogsRequest) (response ListWorkRequestLogsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listWorkRequestLogs, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListWorkRequestLogsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListWorkRequestLogsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListWorkRequestLogsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListWorkRequestLogsResponse")
	}
	return
}

// listWorkRequestLogs implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listWorkRequestLogs(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/workRequests/{workRequestId}/logs", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListWorkRequestLogsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListWorkRequestLogs", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListWorkRequests List all work requests in a compartment.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListWorkRequests.go.html to see an example of how to use ListWorkRequests API.
// A default retry strategy applies to this operation ListWorkRequests()
func (client ContainerEngineClient) ListWorkRequests(ctx context.Context, request ListWorkRequestsRequest) (response ListWorkRequestsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listWorkRequests, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListWorkRequestsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListWorkRequestsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListWorkRequestsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListWorkRequestsResponse")
	}
	return
}

// listWorkRequests implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listWorkRequests(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/workRequests", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListWorkRequestsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListWorkRequests", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListWorkloadMappings List workloadMappings for a provisioned cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/ListWorkloadMappings.go.html to see an example of how to use ListWorkloadMappings API.
// A default retry strategy applies to this operation ListWorkloadMappings()
func (client ContainerEngineClient) ListWorkloadMappings(ctx context.Context, request ListWorkloadMappingsRequest) (response ListWorkloadMappingsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listWorkloadMappings, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListWorkloadMappingsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListWorkloadMappingsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListWorkloadMappingsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListWorkloadMappingsResponse")
	}
	return
}

// listWorkloadMappings implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) listWorkloadMappings(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodGet, "/clusters/{clusterId}/workloadMappings", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response ListWorkloadMappingsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "ListWorkloadMappings", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// StartCredentialRotation Start cluster credential rotation by adding new credentials, old credentials will still work after this operation.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/StartCredentialRotation.go.html to see an example of how to use StartCredentialRotation API.
// A default retry strategy applies to this operation StartCredentialRotation()
func (client ContainerEngineClient) StartCredentialRotation(ctx context.Context, request StartCredentialRotationRequest) (response StartCredentialRotationResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.startCredentialRotation, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = StartCredentialRotationResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = StartCredentialRotationResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(StartCredentialRotationResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into StartCredentialRotationResponse")
	}
	return
}

// startCredentialRotation implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) startCredentialRotation(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPost, "/clusters/{clusterId}/actions/startCredentialRotation", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response StartCredentialRotationResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "StartCredentialRotation", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateAddon Update addon details for a cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/UpdateAddon.go.html to see an example of how to use UpdateAddon API.
// A default retry strategy applies to this operation UpdateAddon()
func (client ContainerEngineClient) UpdateAddon(ctx context.Context, request UpdateAddonRequest) (response UpdateAddonResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateAddon, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateAddonResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateAddonResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateAddonResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateAddonResponse")
	}
	return
}

// updateAddon implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) updateAddon(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPut, "/clusters/{clusterId}/addons/{addonName}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response UpdateAddonResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "UpdateAddon", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateCluster Update the details of a cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/UpdateCluster.go.html to see an example of how to use UpdateCluster API.
// A default retry strategy applies to this operation UpdateCluster()
func (client ContainerEngineClient) UpdateCluster(ctx context.Context, request UpdateClusterRequest) (response UpdateClusterResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateCluster, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateClusterResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateClusterResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateClusterResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateClusterResponse")
	}
	return
}

// updateCluster implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) updateCluster(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPut, "/clusters/{clusterId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response UpdateClusterResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "UpdateCluster", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateClusterEndpointConfig Update the details of the cluster endpoint configuration.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/UpdateClusterEndpointConfig.go.html to see an example of how to use UpdateClusterEndpointConfig API.
// A default retry strategy applies to this operation UpdateClusterEndpointConfig()
func (client ContainerEngineClient) UpdateClusterEndpointConfig(ctx context.Context, request UpdateClusterEndpointConfigRequest) (response UpdateClusterEndpointConfigResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateClusterEndpointConfig, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateClusterEndpointConfigResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateClusterEndpointConfigResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateClusterEndpointConfigResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateClusterEndpointConfigResponse")
	}
	return
}

// updateClusterEndpointConfig implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) updateClusterEndpointConfig(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPost, "/clusters/{clusterId}/actions/updateEndpointConfig", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response UpdateClusterEndpointConfigResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "UpdateClusterEndpointConfig", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateNodePool Update the details of a node pool.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/UpdateNodePool.go.html to see an example of how to use UpdateNodePool API.
// A default retry strategy applies to this operation UpdateNodePool()
func (client ContainerEngineClient) UpdateNodePool(ctx context.Context, request UpdateNodePoolRequest) (response UpdateNodePoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateNodePool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateNodePoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateNodePoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateNodePoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateNodePoolResponse")
	}
	return
}

// updateNodePool implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) updateNodePool(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPut, "/nodePools/{nodePoolId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response UpdateNodePoolResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "UpdateNodePool", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateVirtualNodePool Update the details of a virtual node pool.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/UpdateVirtualNodePool.go.html to see an example of how to use UpdateVirtualNodePool API.
// A default retry strategy applies to this operation UpdateVirtualNodePool()
func (client ContainerEngineClient) UpdateVirtualNodePool(ctx context.Context, request UpdateVirtualNodePoolRequest) (response UpdateVirtualNodePoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateVirtualNodePool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateVirtualNodePoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateVirtualNodePoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateVirtualNodePoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateVirtualNodePoolResponse")
	}
	return
}

// updateVirtualNodePool implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) updateVirtualNodePool(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPut, "/virtualNodePools/{virtualNodePoolId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response UpdateVirtualNodePoolResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "UpdateVirtualNodePool", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateWorkloadMapping Update workloadMapping details for a cluster.
//
// # See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/containerengine/UpdateWorkloadMapping.go.html to see an example of how to use UpdateWorkloadMapping API.
// A default retry strategy applies to this operation UpdateWorkloadMapping()
func (client ContainerEngineClient) UpdateWorkloadMapping(ctx context.Context, request UpdateWorkloadMappingRequest) (response UpdateWorkloadMappingResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.DefaultRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateWorkloadMapping, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateWorkloadMappingResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateWorkloadMappingResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateWorkloadMappingResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateWorkloadMappingResponse")
	}
	return
}

// updateWorkloadMapping implements the OCIOperation interface (enables retrying operations)
func (client ContainerEngineClient) updateWorkloadMapping(ctx context.Context, request common.OCIRequest, binaryReqBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (common.OCIResponse, error) {

	httpRequest, err := request.HTTPRequest(http.MethodPut, "/clusters/{clusterId}/workloadMappings/{workloadMappingId}", binaryReqBody, extraHeaders)
	if err != nil {
		return nil, err
	}

	var response UpdateWorkloadMappingResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		apiReferenceLink := ""
		err = common.PostProcessServiceError(err, "ContainerEngine", "UpdateWorkloadMapping", apiReferenceLink)
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}
