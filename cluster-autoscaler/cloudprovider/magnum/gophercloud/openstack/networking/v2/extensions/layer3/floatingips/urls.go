package floatingips

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"

const resourcePath = "floatingips"

func rootURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL(resourcePath)
}

func resourceURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL(resourcePath, id)
}
