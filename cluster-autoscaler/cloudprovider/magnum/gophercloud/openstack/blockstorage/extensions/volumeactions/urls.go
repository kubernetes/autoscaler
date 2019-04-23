package volumeactions

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"

func actionURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("volumes", id, "action")
}
