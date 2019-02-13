package v1

import (
	"testing"

	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	"github.com/gophercloud/gophercloud/pagination"

	th "github.com/gophercloud/gophercloud/testhelper"
)

func TestNodesCreateDestroy(t *testing.T) {
	clients.RequireLong(t)

	client, err := clients.NewBareMetalV1Client()
	th.AssertNoErr(t, err)
	client.Microversion = "1.38"

	node, err := CreateNode(t, client)
	th.AssertNoErr(t, err)
	defer DeleteNode(t, client, node)

	found := false
	err = nodes.List(client, nodes.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		nodeList, err := nodes.ExtractNodes(page)
		if err != nil {
			return false, err
		}

		for _, n := range nodeList {
			if n.UUID == node.UUID {
				found = true
				return true, nil
			}
		}

		return false, nil
	})
	th.AssertNoErr(t, err)

	th.AssertEquals(t, found, true)
}

func TestNodesUpdate(t *testing.T) {
	clients.RequireLong(t)

	client, err := clients.NewBareMetalV1Client()
	th.AssertNoErr(t, err)
	client.Microversion = "1.38"

	node, err := CreateNode(t, client)
	th.AssertNoErr(t, err)
	defer DeleteNode(t, client, node)

	updated, err := nodes.Update(client, node.UUID, nodes.UpdateOpts{
		nodes.UpdateOperation{
			Op:    nodes.ReplaceOp,
			Path:  "/maintenance",
			Value: "true",
		},
	}).Extract()
	th.AssertNoErr(t, err)

	th.AssertEquals(t, updated.Maintenance, true)
}
