package services

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"

func listURL(client *huawei_cloud_sdk_go.ServiceClient) string {
	return client.ServiceURL("services")
}

func createURL(client *huawei_cloud_sdk_go.ServiceClient) string {
	return client.ServiceURL("services")
}

func serviceURL(client *huawei_cloud_sdk_go.ServiceClient, serviceID string) string {
	return client.ServiceURL("services", serviceID)
}

func updateURL(client *huawei_cloud_sdk_go.ServiceClient, serviceID string) string {
	return client.ServiceURL("services", serviceID)
}
