# Cluster Autoscaler for OpenStack Magnum
The cluster autoscaler for Magnum scales worker nodes within any
specified nodegroup. It will run as a `Deployment` in your cluster.
This README will go over some of the necessary steps required to get
the cluster autoscaler up and running.

## Compatibility

* For Magnum Rocky or earlier: cluster autoscaler v1.18 or lower.
* For Magnum Train or later: cluster autoscaler v1.19 or higher.

Cluster autoscaler versions v1.18 and lower will continue to work on Magnum Train and later versions,
but will only support the single default node group. No extra node groups should be added to clusters
using the cluster autoscaler v1.18 or lower.

## Updates

* CA 1.19
  * Update to support Magnum node groups (introduced in Magnum Train).
    * Add node group autodiscovery based on the group's role property.
  * Report upcoming/failed nodes so that CA can back off if the OpenStack project quota is being exceeded.
* CA 1.15
  * Initial release.

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

| Argument                    | Usage                                                                                                                                            |
|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| --cluster-name              | The name of your Kubernetes cluster. If there are multiple clusters sharing the same name then the cluster IDs should be used instead.           |
| --cloud-provider            | Can be omitted if the autoscaler is built with `BUILD_TAGS=magnum`, otherwise use `--cloud-provider=magnum`.                                     |
| --nodes                     | Used to select a specific node group to autoscale and constrain its node count. Of the form `min:max:NodeGroupName`. Can be used multiple times. |
| --node-group-auto-discovery | See below.                                                                                                                                       |

#### Deployment with helm

Alternatively, the autoscaler can be deployed with the cluster autoscaler helm chart.
A minimal values.yaml file looks like:

```yaml
cloudProvider: "magnum"

magnumClusterName: "cluster name or ID"

autoscalingGroups:
- name: default-worker
  maxSize: 5
  minSize: 1

cloudConfigPath: "/etc/kubernetes/cloud-config"
```

For running on the master node and other suggested settings, see
[examples/values-example.yaml](examples/values-example.yaml).
To deploy with node group autodiscovery (for cluster autoscaler v1.19+), see
[examples/values-autodiscovery.yaml](examples/values-autodiscovery.yaml).


## Node group auto discovery

Instead of using `--nodes` to select specific node groups by name,
node group auto discovery can be used to to let the autoscaler find which node groups
to autoscale by itself, by checking every node group in the cluster against a set of conditions.

The first condition is given in the auto discovery parameter,
to select one or more node group roles which should be autoscalable.

```
--node-group-auto-discovery=magnum:role=worker,autoscaling
```

The above configuration means that for the Magnum provider, any node group which
has a role of "worker" or "autoscaling" should be managed by the cluster autoscaler.
The auto discovery parameter can be used multiple times, so the same configuration could be written as:

```
--node-group-auto-discovery=magnum:role=worker
--node-group-auto-discovery=magnum:role=autoscaling
```

The second condition is that the node group must have a maximum node count set in Magnum.
This can be done using the following command:

```
$ openstack coe nodegroup update <cluster> <nodegroup> replace /max_node_count=5
```

which would set the maximum node count to 5 for whichever node group is updated.

By default the `min_node_count` for a node group is 1, but this can also be changed.

The role of a node group can not be changed after is had been created, but to disable autoscaling
for a node group it is enough to unset the maximum node count.

```
$ openstack coe nodegroup update <cluster> <nodegroup> remove /max_node_count
```

## Notes

The autoscaler will not remove nodes which have non-default kube-system pods.
This prevents the node that the autoscaler is running on from being scaled down.
If you are deploying the autoscaler into a cluster which already has more than one node,
it is best to deploy it onto any node which already has non-default kube-system pods,
to minimise the number of nodes which cannot be removed when scaling.

Or, if you are using a Magnum version which supports scheduling on the master node, then
the example deployment file
[examples/cluster-autoscaler-deployment-master.yaml](examples/cluster-autoscaler-deployment-master.yaml)
can be used.
