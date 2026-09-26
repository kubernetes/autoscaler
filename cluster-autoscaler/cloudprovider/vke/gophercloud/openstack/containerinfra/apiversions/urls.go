package apiversions

import (
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vke/gophercloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vke/gophercloud/openstack/utils"
)

func getURL(c *gophercloud.ServiceClient, version string) string {
	baseEndpoint, _ := utils.BaseEndpoint(c.Endpoint)
	endpoint := strings.TrimRight(baseEndpoint, "/") + "/" + strings.TrimRight(version, "/") + "/"
	return endpoint
}

func listURL(c *gophercloud.ServiceClient) string {
	baseEndpoint, _ := utils.BaseEndpoint(c.Endpoint)
	endpoint := strings.TrimRight(baseEndpoint, "/") + "/"
	return endpoint
}
