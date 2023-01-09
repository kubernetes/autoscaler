# Cluster Autoscaler for Brightbox Cloud

This cloud provider implements the autoscaling function for
[Brightbox Cloud](https://www.brightbox.com). The autoscaler should
work on any Kubernetes clusters running on Brightbox Cloud, however
the approach is tailored to clusters built with the [Kubernetes Cluster
Builder](https://github.com/brightbox/kubernetes-cluster)

# How Autoscaler works on Brightbox Cloud

The autoscaler looks for the first [Config
Map](https://api.gb1.brightbox.com/1.0/index.html) with a name that has
a suffix the same as the `cluster-name` option passed to the autoscaler
(`--cluster-name`).

The config map data consist of a colon separated key-value pairs. The
key and the value are treated as strings.

```
server_group: grp-sda44
min: 1
max: 4
default_group: grp-y6cai
additional_groups: grp-abcde,grp-testy,grp-winga
image: img-testy
zone: zon-testy
user_data: <base64 encoded userdata>
```

The `server_group`, `min` and `max` items are required. All the rest
are optional. Additional Groups should be comma separated without spaces.

The names of the autocreated servers are derived from the name of the config map.

The Brightbox Cloud provider only supports auto-discovery mode using
this pattern. `node-group-auto-discovery` and `nodes` options are
effectively ignored.

## Cluster configuration

If you are using the [Kubernetes Cluster
Builder](https://github.com/brightbox/kubernetes-cluster) set the
`worker_min` and `worker_max` values to scale the worker group, and the
`storage_min` and `storage_max` values to scale the storage group.

The Cluster Builder will ensure the group name and description are
updated with the correct values in the format that autoscaler can recognise.

Generally it is best to keep the `min` and the `count` values to be the
same within the Cluster Buider and let autoscaler create and destroy
servers dynamically up the the `max` value.

While using autoscaler you may find that the Cluster Builder recreates
servers that have been scaled down, if you use the manifests to maintain
the cluster for other reasons (changing the management address for
example). This is a limitation of the Terraform state database, and
autoscaler will scale the cluster back down during the next few minutes.

# Autoscaler Brightbox cloudprovider configuration

The Brightbox Cloud cloudprovider is configured via Environment Variables
suppied to the autoscaler pod. The easiest way to do this is to [create
a secret](https://kubernetes.io/docs/concepts/configuration/secret/#creating-a-secret-manually) containing the variables within the `kube-system` namespace.

```
apiVersion: v1
kind: Secret
metadata:
  name: brightbox-credentials
  namespace: kube-system
type: Opaque
data:
  BRIGHTBOX_API_URL: <base 64 of api URL>
  BRIGHTBOX_CLIENT: <bas64 of Brighbox Cloud client id>
  BRIGHTBOX_CLIENT_SECRET: <base64 of Brightbox Cloud client id secret>
  BRIGHTBOX_KUBE_JOIN_COMMAND: <base64 of cluster join command>
  BRIGHTBOX_KUBE_VERSION: <base 64 of installed k8s version>
```

The join command can be obtained from the kubeadm token command

```
$ kubeadm token create --ttl 0 --description 'Cluster autoscaling token' --print-join-command
```

[Brightbox API
Clients](https://www.brightbox.com/docs/guides/manager/api-clients/)
can be created in the [Brightbox
Manager](https://www.brightbox.com/docs/guides/manager/)

## Cluster Configuration

The [Kubernetes Cluster
Builder](https://github.com/brightbox/kubernetes-cluster) creates a
`brightbox-credentials` secret in the `kube-system` namespace ready
to use.

## Checking the environment

You can check the brightbox-credentials secret by running the `check-env`
job from the examples directory.

```
$ kubectl apply -f examples/check-env.yaml
job.batch/check-env created
$ kubectl -n kube-system logs job/check-env
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=check-env-hbh6m
_BASH_GPG_KEY=7C0135FB088AAF6C66C650B9BB5869F064EA74AB
_BASH_VERSION=5.0
_BASH_PATCH_LEVEL=0
_BASH_LATEST_PATCH=11
BRIGHTBOX_KUBE_VERSION=1.17.0
...
$ kubectl delete -f examples/check-env.yaml
job.batch "check-env" deleted
```

# Running the Autoscaler

1. Clone this repository and change into this directory.
1. Edit the `examples/config.rb` file and adjust the config hash.
2. Alter the cluster name if
required. (If you are using the [Kubernetes Cluster
Builder](https://github.com/brightbox/kubernetes-cluster), this will be
`cluster_name` and `cluster_domainname` joined with a '.')

Then generate and apply the manifests
```
$ make deploy TAG=<version>
```

where TAG is the version you wish to use (1.17, 1.18, etc.)

As the Brightbox cloud-provider auto-detects and potentially scales all
the worker groups, the example deployment file runs the autoscaler on
the master nodes. This avoids it accidentally killing itself.

## Viewing the cluster-autoscaler options

Cluster autoscaler has many options that can be adjusted to better fit
the needs of your application. To view them run

```
$ kubectl create job ca-options --image=brightbox/cluster-autoscaler-brightbox:dev -- ./cluster-autoscaler -h
$ kubectl log job/ca-options
```

Remove the job in the normal way with `kubectl delete job/ca-options`

You can read more details about some of the options in the [main FAQ](../../FAQ.md)


# Building the Brightbox Cloud autoscaler

Extract the repository to a machine running docker and then run the make command

```
$ make build
```

This builds an autoscaler containing only the Brightbox Cloud provider,
tagged as `brightbox/cluster-autoscaler-brightbox:dev`. To build any
other version add a TAG variable

```
make build TAG=1.1x
```
