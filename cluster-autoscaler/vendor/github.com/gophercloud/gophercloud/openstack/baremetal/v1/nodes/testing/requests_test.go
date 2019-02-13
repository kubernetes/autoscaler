package testing

import (
	"testing"

	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"
	"github.com/gophercloud/gophercloud/pagination"
	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
)

func TestListDetailNodes(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleNodeListDetailSuccessfully(t)

	pages := 0
	err := nodes.ListDetail(client.ServiceClient(), nodes.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		pages++

		actual, err := nodes.ExtractNodes(page)
		if err != nil {
			return false, err
		}

		if len(actual) != 3 {
			t.Fatalf("Expected 3 nodes, got %d", len(actual))
		}
		th.CheckDeepEquals(t, NodeFoo, actual[0])
		th.CheckDeepEquals(t, NodeBar, actual[1])
		th.CheckDeepEquals(t, NodeBaz, actual[2])

		return true, nil
	})

	th.AssertNoErr(t, err)

	if pages != 1 {
		t.Errorf("Expected 1 page, saw %d", pages)
	}
}

func TestListNodes(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleNodeListSuccessfully(t)

	pages := 0
	err := nodes.List(client.ServiceClient(), nodes.ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		pages++

		actual, err := nodes.ExtractNodes(page)
		if err != nil {
			return false, err
		}

		if len(actual) != 3 {
			t.Fatalf("Expected 3 nodes, got %d", len(actual))
		}
		th.AssertEquals(t, "foo", actual[0].Name)
		th.AssertEquals(t, "bar", actual[1].Name)
		th.AssertEquals(t, "baz", actual[2].Name)

		return true, nil
	})

	th.AssertNoErr(t, err)

	if pages != 1 {
		t.Errorf("Expected 1 page, saw %d", pages)
	}
}

func TestListOpts(t *testing.T) {
	// Detail cannot take Fields
	opts := nodes.ListOpts{
		Fields: []string{"name", "uuid"},
	}

	_, err := opts.ToNodeListDetailQuery()
	th.AssertEquals(t, err.Error(), "fields is not a valid option when getting a detailed listing of nodes")

	// Regular ListOpts can
	query, err := opts.ToNodeListQuery()
	th.AssertEquals(t, query, "?fields=name&fields=uuid")
	th.AssertNoErr(t, err)
}

func TestCreateNode(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleNodeCreationSuccessfully(t, SingleNodeBody)

	actual, err := nodes.Create(client.ServiceClient(), nodes.CreateOpts{
		Name:          "foo",
		Driver:        "ipmi",
		BootInterface: "pxe",
		DriverInfo: map[string]interface{}{
			"ipmi_port":      "6230",
			"ipmi_username":  "admin",
			"deploy_kernel":  "http://172.22.0.1/images/tinyipa-stable-rocky.vmlinuz",
			"ipmi_address":   "192.168.122.1",
			"deploy_ramdisk": "http://172.22.0.1/images/tinyipa-stable-rocky.gz",
			"ipmi_password":  "admin",
		},
	}).Extract()
	th.AssertNoErr(t, err)

	th.CheckDeepEquals(t, NodeFoo, *actual)
}

func TestDeleteNode(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleNodeDeletionSuccessfully(t)

	res := nodes.Delete(client.ServiceClient(), "asdfasdfasdf")
	th.AssertNoErr(t, res.Err)
}

func TestGetNode(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleNodeGetSuccessfully(t)

	c := client.ServiceClient()
	actual, err := nodes.Get(c, "1234asdf").Extract()
	if err != nil {
		t.Fatalf("Unexpected Get error: %v", err)
	}

	th.CheckDeepEquals(t, NodeFoo, *actual)
}

func TestUpdateNode(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleNodeUpdateSuccessfully(t, SingleNodeBody)

	c := client.ServiceClient()
	actual, err := nodes.Update(c, "1234asdf", nodes.UpdateOpts{
		nodes.UpdateOperation{
			Op:    nodes.ReplaceOp,
			Path:  "/driver",
			Value: "new-driver",
		},
	}).Extract()
	if err != nil {
		t.Fatalf("Unexpected Update error: %v", err)
	}

	th.CheckDeepEquals(t, NodeFoo, *actual)

}
