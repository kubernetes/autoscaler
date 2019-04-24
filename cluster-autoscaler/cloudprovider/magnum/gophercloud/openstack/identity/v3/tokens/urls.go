package tokens

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"

func tokenURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("auth", "tokens")
}
