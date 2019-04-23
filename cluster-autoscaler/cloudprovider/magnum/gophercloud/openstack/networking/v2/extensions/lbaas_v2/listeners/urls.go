package listeners

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"

const (
	rootPath     = "lbaas"
	resourcePath = "listeners"
)

func rootURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL(rootPath, resourcePath)
}

func resourceURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL(rootPath, resourcePath, id)
}
