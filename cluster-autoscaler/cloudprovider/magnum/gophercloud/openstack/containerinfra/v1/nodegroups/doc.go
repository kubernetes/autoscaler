/*
Package nodegroups provides methods for interacting with the Magnum node group API.

All node group actions must be performed on a specific cluster,
so the cluster UUID/name is required as a parameter in each method.


Create a client to use:

    opts, err := openstack.AuthOptionsFromEnv()
    if err != nil {
        panic(err)
    }

    provider, err := openstack.AuthenticatedClient(opts)
    if err != nil {
        panic(err)
    }

    client, err := openstack.NewContainerInfraV1(provider, gophercloud.EndpointOpts{Region: os.Getenv("OS_REGION_NAME")})
    if err != nil {
        panic(err)
    }

    client.Microversion = "1.9"


Example of Getting a node group:

    ng, err := nodegroups.Get(client, clusterUUID, nodeGroupUUID).Extract()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%#v\n", ng)


Example of Listing node groups:

    listOpts := nodegroup.ListOpts{
        Role: "worker",
    }

    allPages, err := nodegroups.List(client, clusterUUID, listOpts).AllPages()
    if err != nil {
        panic(err)
    }

    ngs, err := nodegroups.ExtractNodeGroups(allPages)
    if err != nil {
        panic(err)
    }

    for _, ng := range ngs {
        fmt.Printf("%#v\n", ng)
    }


Example of Creating a node group:

    // Labels, node image and node flavor will be inherited from the cluster value if not set.
    // Role will default to "worker" if not set.

    // To add a label to the new node group, need to know the cluster labels
    cluster, err := clusters.Get(client, clusterUUID).Extract()
    if err != nil {
        panic(err)
    }

    // Add the new label
    labels := cluster.Labels
    labels["availability_zone"] = "A"

    maxNodes := 5
    createOpts := nodegroups.CreateOpts{
        Name:         "new-nodegroup",
        MinNodeCount: 2,
        MaxNodeCount: &maxNodes,
        Labels: labels,
    }

    ng, err := nodegroups.Create(client, clusterUUID, createOpts).Extract()
    if err != nil {
        panic(err)
    }

    fmt.Printf("%#v\n", ng)


Example of Updating a node group:

    // Valid paths are "/min_node_count" and "/max_node_count".
    // Max node count can be unset with the "remove" op to have
    // no enforced maximum node count.

    updateOpts := []nodegroups.UpdateOptsBuilder{
        nodegroups.UpdateOpts{
            Op:    nodegroups.ReplaceOp,
            Path:  "/max_node_count",
            Value: 10,
        },
    }

    ng, err = nodegroups.Update(client, clusterUUID, nodeGroupUUID, updateOpts).Extract()
    if err != nil {
        panic(err)
    }

    fmt.Printf("%#v\n", ng)


Example of Deleting a node group:

     err = nodegroups.Delete(client, clusterUUID, nodeGroupUUID).ExtractErr()
     if err != nil {
         panic(err)
     }
*/
package nodegroups
