# Cluster Autoscaler for Oracle Cloud Infrastructure (OCI)

**Note**: this implementation of Cluster Autoscaler is intended for use with self-managed Kubernetes running on Oracle Cloud Infrastructure and not [Oracle Container Engine for Kubernetes](https://www.oracle.com/cloud-native/container-engine-kubernetes/). Refer to [Using the Kubernetes Cluster Autoscaler](https://docs.oracle.com/en-us/iaas/Content/ContEng/Tasks/contengusingclusterautoscaler.htm#Using_Kubernetes_Horizontal_Pod_Autoscaler), for information about using Cluster Autoscaler with Oracle Container Engine for Kubernetes.


When operating a self-managed Kubernetes cluster in OCI, the Cluster Autoscaler utilizes [Instance Pools](https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/creatinginstancepool.htm)
combined with [Instance Configurations](https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/creatinginstanceconfig.htm) to
automatically resize a cluster's nodes based on application workload demands by:

- adding nodes to static instance-pool(s) when a pod cannot be scheduled in the cluster because of insufficient resource constraints.
- removing nodes from an instance-pool(s) when the nodes have been underutilized for an extended time, and when pods can be placed on other existing nodes.

The Cluster Autoscaler works on a per-instance pool basis. You configure the Cluster Autoscaler to tell it which instance pools to target
for expansion and contraction, the minimum and maximum sizes for each pool, and how you want the autoscaling to take place.
Instance pools not referenced in the configuration file are not managed by the Cluster Autoscaler.

## Create Required OCI Resources

### IAM Policy (if using Instance Principals)

We recommend setting up and configuring the Cluster Autoscaler to use
[Instance Principals](https://docs.oracle.com/en-us/iaas/Content/Identity/Tasks/callingservicesfrominstances.htm)
to authenticate to the OCI APIs.

The following policy provides the minimum privileges necessary for Cluster Autoscaler to run:

1: Create a compartment-level dynamic group containing the nodes (compute instances) in the cluster:

```
All {instance.compartment.id = 'ocid1.compartment.oc1..aaaaaaaa7ey4sg3a6b5wnv5hlkjlkjadslkfjalskfjalsadfadsf'}
```

2: Create a *tenancy-level* policy to allow nodes to manage instance-pools:

```
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to manage instance-pools in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to manage instance-configurations in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to manage instance-family in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to use subnets in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to read virtual-network-family in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to use vnics in compartment <compartment-name>
Allow dynamic-group acme-oci-cluster-autoscaler-dyn-grp to inspect compartments in compartment <compartment-name>
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

## Configure Cluster Autoscaler

Use the `--nodes=<min-nodes>:<max-nodes>:<instancepool-ocid>` parameter to specify which pre-existing instance
pools to target for automatic expansion and contraction, the minimum and maximum sizes for each node pool, and how you
want the autoscaling to take place. Instance pools not referenced in the configuration file are not managed by the
autoscaler where:

- `<min-nodes>` is the minimum number of nodes allowed in the instance-pool.
- `<max-nodes>` is the maximum number of nodes allowed in the instance-pool. Make sure the maximum number of nodes you specify does not exceed the tenancy limits for the node shape defined for the node pool.
- `<instancepool-ocid>` is the OCIDs of a pre-existing instance-pool.

If you are authenticating via instance principals, be sure the `OCI_REGION` environment variable is set to the correct
value in the deployment e.g.:

```yaml
env:
  - name: OCI_REGION
    value: "us-phoenix-1"
```

### Optional cloud-config file

_Optional_ cloud-config file mounted in the path specified by `--cloud-config`.

An example, of passing optional configuration via `cloud-config` file that uses configures the cluster-autoscaler to use
instance-principals authenticating via instance principalsand only see configured instance-pools in a single compartment:

```ini
[Global]
compartment-id = ocid1.compartment.oc1..aaaaaaaa7ey4sg3a6b5wnv5hlkjlkjadslkfjalskfjalsadfadsf
region = uk-london-1
use-instance-principals = true
```

### Environment variables

Configuration via environment-variables:

- `OCI_USE_INSTANCE_PRINCIPAL` - Whether to use Instance Principals for authentication rather than expecting an OCI config file to be mounted in the container. Defaults to false.
- `OCI_REGION` - **Required** when using Instance Principals. e.g. `OCI_REGION=us-phoenix-1`. See [region list](https://docs.oracle.com/en-us/iaas/Content/General/Concepts/regions.htm) for identifiers.
- `OCI_COMPARTMENT_ID` - Restrict the cluster-autoscaler to instance-pools in a single compartment. When unset, the cluster-autoscaler will manage each specified instance-pool no matter which compartment they are in.
- `OCI_REFRESH_INTERVAL` - Optional refresh interval to sync internal cache with OCI API defaults to `2m`.
- `OCI_USE_NON_POOL_MEMBER_ANNOTATION` - Optional. If true, the node will be annotated as non-pool-member if it doesn't belong to any instance pool and the time-consuming instance pool lookup will be skipped.

## Deployment

### Create OCI config secret (only if _not_ using Instance Principals)

If you are opting for a file based OCI configuration (as opposed to instance principals), the OCI config file and private key need to be mounted into the container filesystem using a secret volume.

The following policy is required when the specified is not an administrator to run the cluster-autoscaler:

```
Allow group acme-oci-cluster-autoscaler-user-grp to manage instance-pools in compartment <compartment-name>
Allow group acme-oci-cluster-autoscaler-user-grp to manage instance-configurations in compartment <compartment-name>
Allow group acme-oci-cluster-autoscaler-user-grp to manage instance-family in compartment <compartment-name>
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
```

Instance principal based authentication deployment:

Substitute the OCIDs of _your_ instance pool(s) and set the `OCI_REGION` environment variable to the region where your
instance pool(s) reside before applying the deployment:

```
kubectl apply -f ./cloudprovider/oci/examples/oci-ip-cluster-autoscaler-w-principals.yaml
```

OCI config file based authentication deployment:

```
kubectl apply -f ./cloudprovider/oci/examples/oci-ip-cluster-autoscaler-w-config.yaml
```

## Common Notes and Gotchas:
- You must configure the instance configuration of new compute instances to join the existing cluster when they start. This can
  be accomplished with `cloud-init` / `user-data` in the instance launch configuration [example](./examples/instance-details.json).
- If opting for a file based OCI configuration (as opposed to instance principals), ensure the OCI config and private-key
  PEM files are mounted into the container filesystem at the [expected path](https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm). Note the `key_file` option in the example `~/.oci/config` above references a private-key file mounted into container by the example [volumeMount](./examples/oci-ip-cluster-autoscaler-w-config.yaml#L165)
- Make sure the maximum number of nodes you specify does not exceed the limit for the instance-pool or the tenancy.
- We recommend creating multiple instance-pools with one availability domain specified so new nodes can be created to meet
  affinity requirements across availability domains.
- If you are authenticating via instance principals, be sure the `OCI_REGION` environment variable is set to the correct
  value in the deployment.
- The Cluster Autoscaler will not automatically remove scaled down (terminated) `Node` objects from the Kubernetes API
  without assistance from the [OCI Cloud Controller Manager](https://github.com/oracle/oci-cloud-controller-manager) (CCM).
  If scaled down nodes are lingering in your cluster in the `NotReady` status, ensure the OCI CCM is installed and running
  correctly (`oci-cloud-controller-manager`).
- Avoid manually changing node pools that are managed by the Cluster Autoscaler. For example, do not add or remove nodes
  using kubectl, or using the Console (or the Oracle Cloud Infrastructure CLI or API).
- `--node-group-auto-discovery` and `--node-autoprovisioning-enabled=true` are not supported.
- We set a `nvidia.com/gpu:NoSchedule` taint on nodes in a GPU enabled instance-pool.

## Helpful links
- [Oracle Cloud Infrastructure home](https://cloud.oracle.com)
- [OCI instance configuration documentation](https://docs.oracle.com/en-us/iaas/Content/Compute/Tasks/creatinginstanceconfig.htm)
- [instance principals](https://docs.oracle.com/en-us/iaas/Content/Identity/Tasks/callingservicesfrominstances.htm)
- [OCI Cloud Controller Manager](https://github.com/oracle/oci-cloud-controller-manager)
- [OCI Container Storage Interface driver](https://github.com/oracle/oci-cloud-controller-manager/blob/master/container-storage-interface.md)
