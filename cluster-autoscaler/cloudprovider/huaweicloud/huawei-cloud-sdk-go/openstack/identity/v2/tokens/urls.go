package tokens

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"

// CreateURL generates the URL used to create new Tokens.
func CreateURL(client *huawei_cloud_sdk_go.ServiceClient) string {
	return client.ServiceURL("tokens")
}

// GetURL generates the URL used to Validate Tokens.
func GetURL(client *huawei_cloud_sdk_go.ServiceClient, token string) string {
	return client.ServiceURL("tokens", token)
}
