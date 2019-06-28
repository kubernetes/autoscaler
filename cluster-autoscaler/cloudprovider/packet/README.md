# Cluster Autoscaler for Packet

The cluster autoscaler for [Packet](https://packet.com) worker nodes performs
autoscaling within any specified nodepool. It will run as a `Deployment` in
your cluster. The nodepool is specified using tags on Packet.

This README will go over some of the necessary steps required to get
the cluster autoscaler up and running.

## Permissions and credentials

The autoscaler needs a `ServiceAccount` with permissions for Kubernetes and
requires credentials for interacting with Packet.

An example `ServiceAccount` is given in [examples/cluster-autoscaler-svcaccount.yaml](examples/cluster-autoscaler-svcaccount.yaml).

The credentials for authenticating with Packet are stored in a secret and
provided as an env var to the container. [examples/cluster-autoscaler-secret](examples/cluster-autoscaler-secret.yaml)
In the above file you can modify the following fields:

| Secret                          | Key                     | Value                        |
|---------------------------------|-------------------------|------------------------------------------------------------------------------------------------------------------------------------|
| cluster-autoscaler-packet       | authtoken               | Your Packet API token                                                                                                              |
| cluster-autoscaler-cloud-config | Global/project-id       | Your Packet project id                                                                                                             |
| cluster-autoscaler-cloud-config | Global/api-server       | The ip:port for you cluster's k8s api (e.g. K8S_MASTER_PUBLIC_IP:6443)                                                             |
| cluster-autoscaler-cloud-config | Global/facility         | The Packet facility for the devices in your nodepool (eg: ams1)                                                                    |
| cluster-autoscaler-cloud-config | Global/plan             | The Packet plan (aka size/flavor) for new nodes in the nodepool (eg: t1.small.x86)                                                 |
| cluster-autoscaler-cloud-config | Global/billing          | The billing interval for new nodes (default: hourly)                                                                               |
| cluster-autoscaler-cloud-config | Global/os               | The OS image to use for new nodes (default: ubuntu_18_04). If you change this also update cloudinit.                               |
| cluster-autoscaler-cloud-config | Global/cloudinit        | The base64 encoded [user data](https://support.packet.com/kb/articles/user-data) submitted when provisioning devices. In the example file, the default value has been tested with Ubuntu 18.04 to install Docker & kubelet and then to bootstrap the node into the cluster using kubeadm. For a different base OS or bootstrap method, this needs to be customized accordingly.  |
| cluster-autoscaler-cloud-config | Global/reservation      | The values "require" or "prefer" will request the next available hardware reservation for new devices in selected facility & plan. If no hardware reservations match, "require" will trigger a failure, while "prefer" will launch on-demand devices instead (default: none)  |
| cluster-autoscaler-cloud-config | Global/hostname-pattern | The pattern for the names of new Packet devices (default: "k8s-{{.ClusterName}}-{{.NodeGroup}}-{{.RandString8}}" )                  |

## Configure nodepool and cluster names using Packet tags

The Packet API does not yet have native support for groups or pools of devices. So we use tags to specify them. Each Packet device that's a member of the "cluster1" cluster should have the tag k8s-cluster-cluster1. The devices that are members of the "pool1" nodepool should also have the tag k8s-nodepool-pool1. Once you have a Kubernetes cluster running on Packet, use the Packet Portal or API to tag the nodes accordingly.

## Autoscaler deployment

The deployment in `examples/cluster-autoscaler-deployment.yaml` can be used,
but the arguments passed to the autoscaler will need to be changed
to match your cluster.

| Argument         | Usage                                                                                                      |
|------------------|------------------------------------------------------------------------------------------------------------|
| --cluster-name   | The name of your Kubernetes cluster. It should correspond to the tags that have been applied to the nodes. |
| --nodes          | Of the form `min:max:NodepoolName`. Only a single node pool is currently supported.                        |

## Notes

The autoscaler will not remove nodes which have non-default kube-system pods.
This prevents the node that the autoscaler is running on from being scaled down.
If you are deploying the autoscaler into a cluster which already has more than one node,
it is best to deploy it onto any node which already has non-default kube-system pods,
to minimise the number of nodes which cannot be removed when scaling.
