// +build acceptance

package v1

import (
	"testing"

	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/acceptance/tools"
	"github.com/gophercloud/gophercloud/openstack/orchestration/v1/stackresources"
	th "github.com/gophercloud/gophercloud/testhelper"
)

func TestStackResources(t *testing.T) {
	client, err := clients.NewOrchestrationV1Client()
	th.AssertNoErr(t, err)

	stack, err := CreateStack(t, client)
	th.AssertNoErr(t, err)
	defer DeleteStack(t, client, stack.Name, stack.ID)

	resource, err := stackresources.Get(client, stack.Name, stack.ID, basicTemplateResourceName).Extract()
	th.AssertNoErr(t, err)
	tools.PrintResource(t, resource)

	metadata, err := stackresources.Metadata(client, stack.Name, stack.ID, basicTemplateResourceName).Extract()
	th.AssertNoErr(t, err)
	tools.PrintResource(t, metadata)

	allPages, err := stackresources.List(client, stack.Name, stack.ID, nil).AllPages()
	th.AssertNoErr(t, err)
	allResources, err := stackresources.ExtractResources(allPages)
	th.AssertNoErr(t, err)

	var found bool
	for _, v := range allResources {
		if v.Name == basicTemplateResourceName {
			found = true
		}
	}

	th.AssertEquals(t, found, true)
}
