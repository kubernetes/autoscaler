# Cluster Autoscaler for Oracle Cloud Infrastructure (OCI)

**Note**: this implementation of Cluster Autoscaler is intended for use for **both** self-managed Kubernetes running on Oracle Cloud Infrastructure and [Oracle Container Engine for Kubernetes](https://www.oracle.com/cloud-native/container-engine-kubernetes/).

The Cluster Autoscaler automatically resizes a cluster's nodes based on application workload demands by:

- adding nodes to static pool(s) when a pod cannot be scheduled in the cluster because of insufficient resource constraints.
- removing nodes from pool(s) when the nodes have been underutilized for an extended time, and when pods can be placed on other existing nodes.

The Cluster Autoscaler works on a per-pool basis. You configure the Cluster Autoscaler to tell it which pools to target
for expansion and contraction, the minimum and maximum sizes for each pool, and how you want the autoscaling to take place.
Pools not referenced in the configuration file are not managed by the Cluster Autoscaler.

When operating a self-managed Kubernetes cluster in OCI, the Cluster Autoscaler utilizes [Instance Pools](https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/creatinginstancepool.htm)
combined with [Instance Configurations](https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/creatinginstanceconfig.htm).

When running with Oracle Container Engine for Kubernetes, the Cluster Autoscaler utilizes [Node Pools](https://docs.oracle.com/en-us/iaas/Content/ContEng/Tasks/contengscalingclusters.htm). More details on using the Cluster Autoscaler with Node Pools here: [Using the Kubernetes Cluster Autoscaler](https://docs.oracle.com/en-us/iaas/Content/ContEng/Tasks/contengusingclusterautoscaler.htm#Using_Kubernetes_Horizontal_Pod_Autoscaler).

## Create Required OCI Resources

### IAM Policy (if using Instance Principals)

We recommend setting up and configuring the Cluster Autoscaler to use
[Instance Principals](https://docs.oracle.com/en-us/iaas/Content/Identity/Tasks/callingservicesfrominstances.htm)
to authenticate to the OCI APIs.

The following policy provides the privileges necessary for Cluster Autoscaler to run:

1: Create a compartment-level dynamic group containing the nodes (compute instances) in the cluster:

```
All {instance.compartment.id = 'ocid1.compartment.oc1..aaaaaaaa7ey4sg3a6b5wnv5hlkjlkjadslkfjalskfjalsadfadsf'}
```

Note: the matching rule in the dynamic group above includes all instances
in the specified compartment. If this is too broad for your requirements,
you can add more conditions for example

```
All {instance.compartment.id = '...', tag.MyTagNamespace.MyNodeRole = 'MyTagValue'}
```

here `MyTagValue` is the defined-tag assigned to all nodes where `cluster-autoscaler` pods will be scheduled
(for example, with `nodeSeletor`).
See [node-pool](https://docs.oracle.com/en-us/iaas/Content/ContEng/Tasks/contengtaggingclusterresources_tagging-oke-resources_node-tags.htm)
or [instance-pool](https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/creatinginstanceconfig.htm)
and also [managing dynamic groups](https://docs.oracle.com/en-us/iaas/Content/Identity/Tasks/managingdynamicgroups.htm).


2: Create a *tenancy-level* policy to allow nodes to manage node-pools and/or instance-pools:

```
# if using node pools
Allow dynamic-group acme-oke-cluster-autoscaler-dyn-grp to manage cluster-node-pools in compartment <compartment-name>
# if using instance pools
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to manage instance-pools in compartment <compartment-name>

Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to manage instance-configurations in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to manage instance-family in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to manage volume-family in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to use subnets in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to read virtual-network-family in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to use vnics in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to inspect compartments in compartment <compartment-name>
```

### If using Workload Identity

Note: This is available to use with OKE Node Pools or OCI Managed Instance Pools with OKE Enhanced Clusters only.

See the [documentation](https://docs.oracle.com/en-us/iaas/Content/ContEng/Tasks/contenggrantingworkloadaccesstoresources.htm) for more details

When using a mix of nodes, make sure to add proper lables and affinities on the cluster-autoscaler deployment to prevent it from being deployed on non-OCI managed nodes.

```
Allow any-user to manage cluster-node-pools in compartment <compartment name> where ALL {request.principal.type='workload', request.principal.namespace ='<namespace>', request.principal.service_account = 'cluster-autoscaler', request.principal.cluster_id = 'ocid1.cluster.oc1....'}
Allow any-user to manage instance-family in compartment <compartment name> where ALL {request.principal.type='workload', request.principal.namespace ='<namespace>', request.principal.service_account = 'cluster-autoscaler', request.principal.cluster_id = 'ocid1.cluster.oc1....'}
Allow any-user to use subnets in compartment <compartment name> where ALL {request.principal.type='workload', request.principal.namespace ='<namespace>', request.principal.service_account = 'cluster-autoscaler', request.principal.cluster_id = 'ocid1.cluster.oc1....'}
Allow any-user to read virtual-network-family in compartment <compartment name> where ALL {request.principal.type='workload', request.principal.namespace ='<namespace>', request.principal.service_account = 'cluster-autoscaler', request.principal.cluster_id = 'ocid1.cluster.oc1....'}
Allow any-user to use vnics in compartment <compartment name> where ALL {request.principal.type='workload', request.principal.namespace ='<namespace>', request.principal.service_account = 'cluster-autoscaler', request.principal.cluster_id = 'ocid1.cluster.oc1....'}
Allow any-user to inspect compartments in compartment <compartment name> where ALL {request.principal.type='workload', request.principal.namespace ='<namespace>', request.principal.service_account = 'cluster-autoscaler', request.principal.cluster_id = 'ocid1.cluster.oc1....'}
```

### Instance Pool and Instance Configurations

Before you deploy the Cluster Autoscaler on OCI, your need to create one or more static Instance Pools and Instance
Configuration with `cloud-init` specified in the launch details so new nodes automatically joins the existing cluster on
start up.

Advanced Instance Pool and Instance Configuration configuration is out of scope for this document. However, a
working [instance-details.json](./examples/instance-details.json) and [placement-config.json](./examples/placement-config.json)
([example](./examples/instance-details.json) based on Rancher [RKE](https://rancher.com/products/rke/)) using [cloud-init](https://cloudinit.readthedocs.io/en/latest/) are
included in the examples, which can be applied using the [OCI CLI](https://docs.oracle.com/en-us/iaas/Content/API/SDKDocs/cliinstall.htm).

Modify the `user_data` in the example [instance-details.json](./examples/instance-details.json) to suit your needs, re-base64 encode, apply:

```bash
# e.g. cloud-init. Modify, re-encode, and update user_data in instance-details.json to suit your needs:

$ echo IyEvYmluL2Jhc2gKdG91hci9saWIvYXB0L....1yZXRyeSAzIGhG91Y2ggL3RtcC9jbG91ZC1pbml0LWZpbmlzaGVkCg==  | base64 -D

#!/bin/bash
groupadd docker
usermod -aG docker ubuntu
curl --retry 3 https://releases.rancher.com/install-docker/20.10.sh | sh
docker run -d --privileged --restart=unless-stopped --net=host -v /etc/kubernetes:/etc/kubernetes -v /var/run:/var/run rancher/rancher-agent:v2.5.5 --server https://my-rancher.com --token xxxxxx  --worker
```

```bash
$ oci compute-management instance-configuration create --instance-details file://./cluster-autoscaler/cloudprovider/oci/examples/instance-details.json --compartment-id ocid1.compartment.oc1..aaaaaaaa7ey4sg3a6b5wnv5hlkjlkjadslkfjalskfjalsadfadsf --query 'data.id' --raw-output

ocid1.instanceconfiguration.oc1.phx.aaaaaaaa3neul67zb3goz43lybosc2o3fv67gj3zazexbb3vfcbypmpznhtq

$ oci compute-management instance-pool create --compartment-id ocid1.compartment.oc1..aaaaaaaa7ey4sg3a6b5wnv5hlkjlkjadslkfjalskfjalsadfadsf --instance-configuration-id ocid1.instanceconfiguration.oc1.phx.aaaaaaaa3neul67zb3goz43lybosc2o3fv67gj3zazexbb3vfcbypmpznhtq --placement-configurations file://./cluster-autoscaler/cloudprovider/oci/examples/placement-config.json --size 0 --wait-for-state RUNNING --query 'data.id' --raw-output

Action completed. Waiting until the resource has entered state: ('RUNNING',)
ocid1.instancepool.oc1.phx.aaaaaaaayd5bxwrzomzr2b2enchm4mof7uhw7do5hc2afkhks576syikk2ca
```

### Node Pool

There are no required actions to setup node pools besides creating them:

```
oci ce node-pool create --cluster-id ... --compartment-id, ... --name ... --node-shape ... --size 0 ...
```

## Configure Cluster Autoscaler

Use the `--nodes=<min-nodes>:<max-nodes>:<instancepool-ocid>` parameter to specify which pre-existing instance
pools to target for automatic expansion and contraction, the minimum and maximum sizes for each node pool, and how you
want the autoscaling to take place. The current iteration of the Cluster Autoscaler accepts **either** Instance Pools **or**
Node Pools, but does not accept a mixed set of pool types (yet).

Pools not referenced in the configuration file are not managed by the autoscaler where:

- `<min-nodes>` is the minimum number of nodes allowed in the pool.
- `<max-nodes>` is the maximum number of nodes allowed in the pool. Make sure the maximum number of nodes you specify does not exceed the tenancy limits for the instance shape defined for the pool.
- `<instancepool-ocid>` is the OCIDs of a pre-existing pool.

### Optional cloud-config file

_Optional_ cloud-config file mounted in the path specified by `--cloud-config`.

An example, of passing optional configuration via `cloud-config` file that uses configures the cluster-autoscaler to use
instance-principals authenticating via instance principals and only see configured instance-pools in a single compartment:

```ini
[Global]
compartment-id = ocid1.compartment.oc1..aaaaaaaa7ey4sg3a6b5wnv5hlkjlkjadslkfjalskfjalsadfadsf
region = uk-london-1
use-instance-principals = true
```

### Configuration via environment-variables:

- `OCI_USE_INSTANCE_PRINCIPAL` - Whether to use Instance Principals for authentication rather than expecting an OCI config file to be mounted in the container. Defaults to false.
- `OCI_USE_WORKLOAD_IDENTITY` - Whether to use Workload Identity for authentication (Available with node pools and OCI managed nodepools in OKE Enhanced Clusters only). Setting to `true` takes precedence over `OCI_USE_INSTANCE_PRINCIPAL`. When using this flag, the `OCI_RESOURCE_PRINCIPAL_VERSION` (1.1 or 2.2) and `OCI_RESOURCE_PRINCIPAL_REGION` also need to be set. See this [blog post](https://blogs.oracle.com/cloud-infrastructure/post/oke-workload-identity-greater-control-access#:~:text=The%20OKE%20Workload%20Identity%20feature,having%20to%20run%20fewer%20nodes.) for more details on setting the policies for this auth mode.
- `OCI_REFRESH_INTERVAL` - Optional. Refresh interval to sync internal cache with OCI API. Defaults to `2m`.

#### Instance Pool specific environment-variables

- `OCI_USE_NON_POOL_MEMBER_ANNOTATION` - Optional. If true, the node will be annotated as non-pool-member if it doesn't belong to any instance pool and the time-consuming instance pool lookup will be skipped.

#### Environment-variables applicable _only_ for Node Pools

n/a

### Node Group Auto Discovery
`--node-group-auto-discovery` could be given in below pattern. It would discover the nodepools under given compartment by matching the nodepool tags (either they are Freeform or Defined tags). All of the parameters are mandatory.
```
clusterId:<clusterId>,compartmentId:<compartmentId>,nodepoolTags:<tagKey1>=<tagValue1>&<tagKey2>=<tagValue2>,min:<min>,max:<max>
```
Auto discovery can not be used along with static discovery (`node` parameter) to prevent conflicts.

## Deployment

### Create OCI config secret (only if _not_ using Instance Principals)

If you are opting for a file based OCI configuration (as opposed to instance principals), the OCI config file and private key need to be mounted into the container filesystem using a secret volume.

The following policy is required when the specified is not an administrator to run the cluster-autoscaler:

```
# if using node pools
Allow group acme-oke-cluster-autoscaler-dyn-grp to manage cluster-node-pools in compartment <compartment-name>
# if using instance pools
Allow group acme-oci-cluster-autoscaler-dyn-grp to manage instance-pools in compartment <compartment-name>

Allow group acme-oci-cluster-autoscaler-user-grp to manage instance-configurations in compartment <compartment-name>
Allow group acme-oci-cluster-autoscaler-user-grp to manage instance-family in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to manage volume-family in compartment <compartment-name>
Allow group acme-oci-cluster-autoscaler-user-grp to use subnets in compartment <compartment-name>
Allow group acme-oci-cluster-autoscaler-user-grp to read virtual-network-family in compartment <compartment-name>
Allow group acme-oci-cluster-autoscaler-user-grp to use vnics in compartment <compartment-name>
Allow group acme-oci-cluster-autoscaler-user-grp to inspect compartments in compartment <compartment-name>
```

Example OCI config file (note `key_file` is the expected path and filename of the OCI API private-key from the perspective of the container):

```bash
$ cat ~/.oci/config

[DEFAULT]
user=ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
fingerprint=xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx
key_file=/root/.oci/api_key.pem
tenancy=ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
pass_phrase=
region=us-phoenix-1
```

Create the secret (`api_key.pem` key name is required):

```bash
kubectl create secret generic oci-config -n kube-system --from-file=/Users/me/.oci/config --from-file=api_key.pem=/Users/me/.oci/my_api_key.pem
```

### Example Deployment

Two example deployments of the Cluster Autoscaler that manage instancepools are located in the [examples](./examples/) directory.
[oci-ip-cluster-autoscaler-w-principals.yaml](./examples/oci-ip-cluster-autoscaler-w-principals.yaml) uses
instance principals, and [oci-ip-cluster-autoscaler-w-config.yaml](./examples/oci-ip-cluster-autoscaler-w-config.yaml) uses file
based authentication.

Note the 3 specified instance-pools are intended to correspond to different availability domains in the Phoenix, AZ region:

```yaml
...
      containers:
        - image: registry.k8s.io/autoscaling/cluster-autoscaler:{{ ca_version }}
          name: cluster-autoscaler
          command:
            - ./cluster-autoscaler
            - --cloud-provider=oci
            - --nodes=1:10:ocid1.instancepool.oc1.phx.aaaaaaaaqdxy35acq32zjfvkybjmvlbdgj6q3m55qkwwctxhsprmz633k62q
            - --nodes=0:10:ocid1.instancepool.oc1.phx.aaaaaaaazldzcu4mi5spz56upbtwnsynz2nk6jvmx7zi4hsta4uggxbulbua
            - --nodes=0:20:ocid1.instancepool.oc1.phx.aaaaaaaal3jhoc32ljsfaeif4x2ssfa2a63oehjgqryiueivieee6yaqbkia

            # if using node pools
            - --nodes=1:10:ocid1.nodepool.oc1.phx.aaaaaaaaqdxy35acq32zjfvkybjmvlbdgj6q3m55qkwwctxhsprmz633k62q
            - --nodes=0:10:ocid1.nodepool.oc1.phx.aaaaaaaazldzcu4mi5spz56upbtwnsynz2nk6jvmx7zi4hsta4uggxbulbua
            - --nodes=0:20:ocid1.nodepool.oc1.phx.aaaaaaaal3jhoc32ljsfaeif4x2ssfa2a63oehjgqryiueivieee6yaqbkia
```

Instance principal based authentication deployment:

Substitute the OCIDs of _your_ instance pool(s) before applying the deployment:

```
kubectl apply -f ./cloudprovider/oci/examples/oci-ip-cluster-autoscaler-w-principals.yaml
```

OCI config file based authentication deployment:

```
kubectl apply -f ./cloudprovider/oci/examples/oci-ip-cluster-autoscaler-w-config.yaml
```

OCI with node pool yamls:

```
# First substitute any values mentioned in the file and then apply
kubectl apply -f ./cloudprovider/oci/examples/oci-nodepool-cluster-autoscaler-w-principals.yaml
```

## Common Notes and Gotchas:
- For instance pools, you must configure the instance configuration of new compute instances to join the existing cluster when they start. This can
  be accomplished with `cloud-init` / `user-data` in the instance launch configuration [example](./examples/instance-details.json).
- If opting for a file based OCI configuration (as opposed to instance principals), ensure the OCI config and private-key
  PEM files are mounted into the container filesystem at the [expected path](https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm). Note the `key_file` option in the example `~/.oci/config` above references a private-key file mounted into container by the example [volumeMount](./examples/oci-ip-cluster-autoscaler-w-config.yaml#L165)
- Make sure the maximum number of nodes you specify does not exceed the limit for the pool or the tenancy.
- We recommend creating multiple pools with one availability domain specified so new nodes can be created to meet
  affinity requirements across availability domains.
- The Cluster Autoscaler will not automatically remove scaled down (terminated) `Node` objects from the Kubernetes API
  without assistance from the [OCI Cloud Controller Manager](https://github.com/oracle/oci-cloud-controller-manager) (CCM).
  If scaled down nodes are lingering in your cluster in the `NotReady` status, ensure the OCI CCM is installed and running
  correctly (`oci-cloud-controller-manager`).
- Avoid manually changing pools that are managed by the Cluster Autoscaler. For example, do not add or remove nodes
  using kubectl, or using the Console (or the Oracle Cloud Infrastructure CLI or API).
- `--node-autoprovisioning-enabled=true` are not supported.
- `--node-group-auto-discovery` and `node` parameters can not be used together as it can cause conflicts.
- We set a `nvidia.com/gpu:NoSchedule` taint on nodes in a GPU enabled pools.

## Helpful links
- [Oracle Cloud Infrastructure home](https://cloud.oracle.com)
- [OCI instance configuration documentation](https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/creatinginstanceconfig.htm)
- [instance principals](https://docs.oracle.com/en-us/iaas/Content/Identity/Tasks/callingservicesfrominstances.htm)
- [OCI Cloud Controller Manager](https://github.com/oracle/oci-cloud-controller-manager)
- [OCI Container Storage Interface driver](https://github.com/oracle/oci-cloud-controller-manager/blob/master/container-storage-interface.md)
- [OCI Cluster Autoscaler](https://docs.oracle.com/en-us/iaas/Content/ContEng/Tasks/contengusingclusterautoscaler.htm)
