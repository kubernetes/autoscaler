package tokens

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go"

func tokenURL(c *huawei_cloud_sdk_go.ServiceClient) string {
	return c.ServiceURL("auth", "tokens")
}
