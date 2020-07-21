package endpoints

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"

func listURL(client *huawei_cloud_sdk_go.ServiceClient) string {
	return client.ServiceURL("endpoints")
}

func endpointURL(client *huawei_cloud_sdk_go.ServiceClient, endpointID string) string {
	return client.ServiceURL("endpoints", endpointID)
}
