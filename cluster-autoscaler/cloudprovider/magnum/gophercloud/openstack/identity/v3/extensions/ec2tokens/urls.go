package ec2tokens

import "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"

func ec2tokensURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("ec2tokens")
}

func s3tokensURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("s3tokens")
}
