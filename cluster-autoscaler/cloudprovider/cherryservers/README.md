# Cluster Autoscaler for Cherry Servers

The cluster autoscaler for [Cherry Servers](https://cherryservers.com) worker nodes performs
autoscaling within any specified nodepools. It will run as a `Deployment` in
your cluster. The nodepools are specified using tags on Cherry Servers.

This README will go over some of the necessary steps required to get
the cluster autoscaler up and running.

## Permissions and credentials

The autoscaler needs a `ServiceAccount` with permissions for Kubernetes and
requires credentials, specifically API tokens, for interacting with Cherry Servers.

An example `ServiceAccount` is given in [examples/cluster-autoscaler-svcaccount.yaml](examples/cluster-autoscaler-svcaccount.yaml).

The credentials for authenticating with Cherry Servers are stored in a secret and
provided as an env var to the container. [examples/cluster-autoscaler-secret](examples/cluster-autoscaler-secret.yaml)
In the above file you can modify the following fields:

| Secret                          | Key                     | Value                        |
|---------------------------------|-------------------------|------------------------------------------------------------------------------------------------------------------------------------|
| cluster-autoscaler-cherry       | authtoken               | Your Cherry Servers API token. It must be base64 encoded.                                                                                 |
| cluster-autoscaler-cloud-config | Global/project-id       | Your Cherry Servers project id                                                                                                             |
| cluster-autoscaler-cloud-config | Global/api-server       | The ip:port for you cluster's k8s api (e.g. K8S_MASTER_PUBLIC_IP:6443)                                                             |
| cluster-autoscaler-cloud-config | Global/region         | The Cherry Servers region slug for the servers in your nodepool (eg: `eu_nord_1`)                                                                    |
| cluster-autoscaler-cloud-config | Global/plan             | The Cherry Servers plan slug for new nodes in the nodepool (eg: `e5_1620v4`)                                                 |
| cluster-autoscaler-cloud-config | Global/os               | The OS image slug to use for new nodes, e.g. `ubuntu_18_04`. If you change this also update cloudinit.                               |
| cluster-autoscaler-cloud-config | Global/cloudinit        | The base64 encoded user data submitted when provisioning servers. In the example file, the default value has been tested with Ubuntu 18.04 to install Docker & kubelet and then to bootstrap the node into the cluster using kubeadm. The kubeadm, kubelet, kubectl are pinned to version 1.17.4. For a different base OS or bootstrap method, this needs to be customized accordingly. It will use go templates to inject runtime information; see below.|
| cluster-autoscaler-cloud-config | Global/hostname-pattern | The pattern for the names of new Cherry Servers servers (default: "k8s-{{.ClusterName}}-{{.NodeGroup}}-{{.RandString8}}" )                  |
| cluster-autoscaler-cloud-config | Global/os-partition-size | The OS partition size in gigabytes for new nodes in the nodepool (eg: `60`, default: `none`)                 |

You can always update the secret with more nodepool definitions (with different plans etc.) as shown in the example, but you should always provide a default nodepool configuration.

The userdata use the following fields to inject runtime information into userdata:

* `BootstrapTokenID`: Kubernetes bootstrap token ID, 6 alphanumeric characters, e.g. `nf8atf`
* `BootstrapTokenSecret`: Kubernetes bootstrap token secret, 16 alphanumeric characters, e.g. `kwfs1t15pjmhk7n4`
* `APIServerEndpoint`: endpoint to connect to the Kubernetes server
* `NodeGroup`: name of the nodegroup to which the newly deployed node belongs.

For example:

```sh
kubeadm join --discovery-token-unsafe-skip-ca-verification --token {{.BootstrapTokenID}}.{{.BootstrapTokenSecret}} {{.APIServerEndpoint}}
```

## Configure nodepool and cluster names using Cherry Servers tags

The Cherry Servers API does not yet have native support for groups or pools of servers. So we use tags to specify them. Each Cherry Servers server that's a member of the "cluster1" cluster should have the tag k8s-cluster-cluster1. The servers that are members of the "pool1" nodepool should also have the tag k8s-nodepool-pool1. Once you have a Kubernetes cluster running on Cherry Servers, use the Cherry Servers Portal, API or CLI to tag the nodes accordingly.

## Autoscaler deployment

The yaml files in [examples](./examples) can be used. You will need to change several of the files
to match your cluster:

* [cluster-autoscaler-rbac.yaml](./examples/cluster-autoscaler-rbac.yaml) unchanged
* [cluster-autoscaler-svcaccount.yaml](./examples/cluster-autoscaler-svcaccount.yaml) unchanged
* [cluster-autoscaler-secret.yaml](./examples/cluster-autoscaler-secret.yaml) requires entering the correct tokens, project ID, plan type, etc. for your cluster; see the file comments
* [cluster-autoscaler-deployment.yaml](./examples/cluster-autoscaler-deployment.yaml) requires setting the arguments passed to the autoscaler to match your cluster.

| Argument              | Usage                                                                                                      |
|-----------------------|------------------------------------------------------------------------------------------------------------|
| --cluster-name        | The name of your Kubernetes cluster. It should correspond to the tags that have been applied to the nodes. |
| --nodes               | Of the form `min:max:NodepoolName`. For multiple nodepools you can add the same argument multiple times. E.g. for pool1, pool2 you would add `--nodes=0:10:pool1` and `--nodes=0:10:pool2`. In addition, each node provisioned by the autoscaler will have a label with key: `pool` and with value: `NodepoolName`. These labels can be useful when there is a need to target specific nodepools. |
| --expander=random      |  This is an optional argument which allows the cluster-autoscaler to take into account various algorithms when scaling with multiple nodepools, see [expanders](../../FAQ.md#what-are-expanders). |

## Target Specific Nodepools

In case you want to target a specific nodepool(s) for e.g. a deployment, you can add a `nodeAffinity` with the key `pool` and with value the nodepool name that you want to target. This functionality is not backwards compatible, which means that nodes provisioned with older cluster-autoscaler images won't have the key `pool`. But you can overcome this limitation by manually adding the correct labels. Here are some examples:

Target a nodepool with a specific name:
```
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: pool
          operator: In
          values:
          - pool3
```
Target a nodepool with a specific Cherry Servers instance:
```
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: beta.kubernetes.io/instance-type
          operator: In
          values:
          - e5_1620v4
```

## CCM and Controller node labels

### CCM

By default, autoscaler assumes that you have a recent version of
[Cherry Servers CCM](https://github.com/cherryservers/cloud-provider-cherry)
installed in your
cluster. 

## Notes

The autoscaler will not remove nodes which have non-default kube-system pods.
This prevents the node that the autoscaler is running on from being scaled down.
If you are deploying the autoscaler into a cluster which already has more than one node,
it is best to deploy it onto any node which already has non-default kube-system pods,
to minimise the number of nodes which cannot be removed when scaling. For this reason in
the provided example the autoscaler pod has a nodeaffinity which forces it to deploy on
the control plane (previously referred to as master) node.

## Development

### Testing

The Cherry Servers cluster-autoscaler includes a series of tests, which are executed
against a mock backend server included in the package. It will **not** execute them
against the real Cherry Servers API.

If you want to execute them against the real Cherry Servers API, set the
environment variable:

```sh
CHERRY_USE_PRODUCTION_API=true
```

### Running Locally

To run the CherryServers cluster-autoscaler locally:

1. Save the desired cloud-config to a local file, e.g. `/tmp/cloud-config`. The contents of the file can be extracted from the value in [examples/cluster-autoscaler-secret.yaml](./examples/cluster-autoscaler-secret.yaml), secret named `cluster-autoscaler-cloud-config`, key `cloud-config`.
1. Export the following environment variables:
   * `BOOTSTRAP_TOKEN_ID`: the bootstrap token ID, i.e. the leading 6 characters of the entire bootstrap token, before the `.`
   * `BOOTSTRAP_TOKEN_SECRET`: the bootstrap token secret, i.e. the trailing 16 characters of the entire bootstrap token, after the `.`
   * `CHERRY_AUTH_TOKEN`: your CherryServers authentication token
   * `KUBECONFIG`: a kubeconfig file with permissions to your cluster
   * `CLUSTER_NAME`: the name for your cluster, e.g. `cluster1`
   * `CLOUD_CONFIG`: the path to your cloud-config file, e.g. `/tmp/cloud-config`
1. Run the autoscaler per the command-line below.

The command-line format is:

```
cluster-autoscaler --alsologtostderr --cluster-name=$CLUSTER_NAME --cloud-config=$CLOUD_CONFIG \     
  --cloud-provider=cherryservers \
  --nodes=0:10:pool1 \
  --nodes=0:10:pool2 \
  --scale-down-unneeded-time=1m0s --scale-down-delay-after-add=1m0s --scale-down-unready-time=1m0s \
  --kubeconfig=$KUBECONFIG \
  --v=2
```

You can set `--nodes=` as many times as you like. The format for each `--nodes=` is:

```
--nodes=<min>:<max>:<poolname>
```

* `<min>` and `<max>` must be integers, and `<max>` must be greater than `<min>`
* `<poolname>` must be a pool that exists in the `cloud-config`

If the poolname is not found, it will use the `default` pool, e.g.:

You also can make changes and run it directly, replacing the command with `go run`,
but this must be run from the `cluster-autoscaler` directory, i.e. not within the specific
cloudprovider implementation:

```
go run . --alsologtostderr --cluster-name=$CLUSTER_NAME --cloud-config=$CLOUD_CONFIG \     
  --cloud-provider=cherryservers \
  --nodes=0:10:pool1 \
  --nodes=0:10:pool2 \
  --scale-down-unneeded-time=1m0s --scale-down-delay-after-add=1m0s --scale-down-unready-time=1m0s \
  --kubeconfig=$KUBECONFIG \
  --v=2
```
