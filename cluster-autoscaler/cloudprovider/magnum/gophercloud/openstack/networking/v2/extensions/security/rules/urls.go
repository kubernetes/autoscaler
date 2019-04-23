package rules

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"

const rootPath = "security-group-rules"

func rootURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL(rootPath)
}

func resourceURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL(rootPath, id)
}
