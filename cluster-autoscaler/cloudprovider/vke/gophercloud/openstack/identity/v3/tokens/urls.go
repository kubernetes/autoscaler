package tokens

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vke/gophercloud"

func tokenURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("auth", "tokens")
}
