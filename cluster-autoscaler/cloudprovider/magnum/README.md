# Cluster Autoscaler for OpenStack Magnum
The cluster autoscaler for Magnum scales worker nodes within any
specified nodegroup. It will run as a `Deployment` in your cluster.
This README will go over some of the necessary steps required to get
the cluster autoscaler up and running.

## Permissions and credentials

The autoscaler needs a `ServiceAccount` with permissions for Kubernetes and
requires credentials for interacting with OpenStack.

An example `ServiceAccount` is given in [examples/cluster-autoscaler-svcaccount.yaml](examples/cluster-autoscaler-svcaccount.yaml).

The credentials for authenticating with OpenStack are stored in a secret and
mounted as a file inside the container. [examples/cluster-autoscaler-secret](examples/cluster-autoscaler-secret.yaml)
can be modified with the contents of your cloud-config. This file can be obtained from your master node,
in `/etc/kubernetes` (may be named `kube_openstack_config` instead of `cloud-config`).

## Autoscaler deployment

The deployment in `examples/cluster-autoscaler-deployment.yaml` can be used,
but the arguments passed to the autoscaler will need to be changed
to match your cluster.

| Argument         | Usage                                                                                                                                      |
|------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| --cluster-name   | The name of your Kubernetes cluster. If there are multiple clusters sharing the same name then the cluster IDs should be used instead.     |
| --cloud-provider | Can be omitted if the autoscaler is built with `BUILD_TAGS=magnum`.                                                                        |
| --nodes          | Of the form `min:max:NodeGroupName`. Node groups are not yet implemented in Magnum so only a single node group is currently supported.     |

## Notes

Magnum does not yet support multiple node groups within a single cluster, but this
is currently in development. Once node groups are available for Magnum, support
for autoscaling clusters using nodegroups will be made available by adding another
implementation of the [Magnum manager interface](./magnum_manager.go). 

The autoscaler will not remove nodes which have non-default kube-system pods.
This prevents the node that the autoscaler is running on from being scaled down.
If you are deploying the autoscaler into a cluster which already has more than one node,
it is best to deploy it onto any node which already has non-default kube-system pods,
to minimise the number of nodes which cannot be removed when scaling.

Or, if you are using a Magnum version which supports scheduling on the master node, then
the example deployment file
[examples/cluster-autoscaler-deployment-master.yaml](examples/cluster-autoscaler-deployment-master.yaml)
can be used.