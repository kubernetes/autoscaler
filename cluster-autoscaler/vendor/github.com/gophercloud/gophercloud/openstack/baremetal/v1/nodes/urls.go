package nodes

import "github.com/gophercloud/gophercloud"

func createURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL("nodes")
}

func listURL(client *gophercloud.ServiceClient) string {
	return createURL(client)
}

func listDetailURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL("nodes", "detail")
}

func deleteURL(client *gophercloud.ServiceClient, id string) string {
	return client.ServiceURL("nodes", id)
}

func getURL(client *gophercloud.ServiceClient, id string) string {
	return deleteURL(client, id)
}

func updateURL(client *gophercloud.ServiceClient, id string) string {
	return deleteURL(client, id)
}
