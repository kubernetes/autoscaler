// +build acceptance networking

package portsbinding

import (
	"testing"

	"github.com/gophercloud/gophercloud/acceptance/clients"
	networking "github.com/gophercloud/gophercloud/acceptance/openstack/networking/v2"
	"github.com/gophercloud/gophercloud/acceptance/tools"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	th "github.com/gophercloud/gophercloud/testhelper"
)

func TestPortsbindingCRUD(t *testing.T) {
	clients.RequireAdmin(t)

	client, err := clients.NewNetworkV2Client()
	th.AssertNoErr(t, err)

	// Create Network
	network, err := networking.CreateNetwork(t, client)
	th.AssertNoErr(t, err)
	defer networking.DeleteNetwork(t, client, network.ID)

	// Create Subnet
	subnet, err := networking.CreateSubnet(t, client, network.ID)
	th.AssertNoErr(t, err)
	defer networking.DeleteSubnet(t, client, subnet.ID)

	// Define a host
	hostID := "localhost"

	// Create port
	port, err := CreatePortsbinding(t, client, network.ID, subnet.ID, hostID)
	th.AssertNoErr(t, err)
	defer networking.DeletePort(t, client, port.ID)

	tools.PrintResource(t, port)

	// Update port
	newPortName := ""
	newPortDescription := ""
	updateOpts := ports.UpdateOpts{
		Name:        &newPortName,
		Description: &newPortDescription,
	}
	newPort, err := ports.Update(client, port.ID, updateOpts).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, newPort)
	th.AssertEquals(t, newPort.Description, newPortName)
	th.AssertEquals(t, newPort.Description, newPortDescription)
}
