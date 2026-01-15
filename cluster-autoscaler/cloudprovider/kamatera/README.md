# Cluster Autoscaler for Kamatera

The cluster autoscaler for Kamatera scales nodes in a Kamatera cluster.

## Kamatera Kubernetes

[Kamatera](https://www.kamatera.com/express/compute/) supports Kubernetes clusters using any Kubernetes distribution / provisioning tool.

See the following example Terraform setup using RKE2 for a recommended production ready Kubernetes cluster:

[Kamatera RKE2 Kubernetes Terraform](https://github.com/Kamatera/kamatera-rke2-kubernetes-terraform-example/blob/main/README.md)

## Cluster Autoscaler Node Groups

An autoscaler node group is composed of multiple Kamatera servers with the same server configuration.
All servers belonging to a node group are identified by Kamatera server tags `k8sca-CLUSTER_NAME`, `k8scang-NODEGROUP_NAME`.
The cluster and node groups must be specified in the autoscaler cloud configuration file.

## Deployment

See [examples/](examples/) and modify the configurations as needed.

## Configuration

The cluster autoscaler only considers the cluster and node groups defined in the configuration file.

**Important Note:** The cluster and node group names must be 15 characters or less.

it is an INI file with the following fields:

| Key                                    | Value                                                                                                                                                   | Mandatory | Default                            |
|----------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|-----------|------------------------------------|
| global/kamatera-api-client-id          | Kamatera API Client ID                                                                                                                                  | yes       | none                               |
| global/kamatera-api-secret             | Kamatera API Secret                                                                                                                                     | yes       | none                               |
| global/cluster-name                    | **max 15 characters: english letters, numbers, dash, underscore, space, dot**: distinct string used to set the cluster server tag                       | yes       | none                               |
| global/filter-name-prefix              | autoscaler will only handle server names that start with this prefix                                                                                    | no        | none                               |
| global/provider-id-prefix              | prefix used for Kubernetes node `.spec.providerID` (and for matching nodes to Kamatera instances)                                                       | no        | kamatera://                        |
| global/poweroff-on-scale-down          | boolean - set to true to power-off servers instead of terminating them                                                                                  | no        | false                              |
| global/poweron-on-scale-up             | boolean - set to true to look for powered off servers to use for scale up before creating additional servers                                            | no        | false                              |
| global/default-min-size                | default minimum size of a node group (must be > 0)                                                                                                      | no        | 1                                  |
| global/default-max-size                | default maximum size of a node group                                                                                                                    | no        | 254                                |
| global/default-<SERVER_CONFIG_KEY>     | replace <SERVER_CONFIG_KEY> with the relevant configuration key                                                                                         | see below | see below                          |
| nodegroup \"name\"                     | **max 15 characters: english letters, numbers, dash, underscore, space, dot**: distinct string within the cluster used to set the node group server tag | yes       | none                               |
| nodegroup \"name\"/min-size            | minimum size for a specific node group                                                                                                                  | no        | global/defaut-min-size             |
| nodegroup \"name\"/max-size            | maximum size for a specific node group                                                                                                                  | no        | global/defaut-min-size             |
| nodegroup \"name\"/template-label      | Set labels on the node template used for scale up checks (See below for details)                                                                        | no        | none                               |
| nodegroup \"name\"/<SERVER_CONFIG_KEY> | replace <SERVER_CONFIG_KEY> with the relevant configuration key                                                                                         | no        | global/default-<SERVER_CONFIG_KEY> |

### Server configuration keys

Following are the supported server configuration keys:

| Key | Value | Mandatory | Default |
|-----|-------|-----------|---------|
| name-prefix | Prefix for all created server names | no | none |
| password | Server root password | no | none |
| ssh-key | Public SSH key to add to the server authorized keys | no | none |
| datacenter | Datacenter ID | yes | none |
| image | Image ID or name | yes | none |
| cpu | CPU type and size identifier | yes | none |
| ram | RAM size in MB | yes | none |
| disk | Disk specifications - see below for details | yes | none |
| dailybackup | boolean - set to true to enable daily backups | no | false |
| managed | boolean - set to true to enable managed services | no | false |
| network | Network specifications - see below for details | yes | none |
| billingcycle | \"hourly\" or \"monthly\" | no | \"hourly\" |
| monthlypackage | For monthly billing only - the monthly network package to use | no | none |
| script-base64 | base64 encoded server initialization script, must be provided to connect the server to the cluster, see below for details | no | none |

### Disk specifications

Server disks are specified using an array of strings which are the same as the cloudcli `--disk` argument
as specified in [cloudcli server create](https://github.com/cloudwm/cloudcli/blob/master/docs/cloudcli_server_create.md).
For multiple disks, include the configuration multiple times, example:

```
[global]
; default for all node groups: single 100gb disk
default-disk = "size=100"

[nodegroup "ng1"]
; this node group will use the default

[nodegroup "ng2"]
; override the default and use 2 disks
disk = "size=100"
disk = "size=200"
```

### Network specifications

Networks are specified using an array of strings which are the same as the cloudcli `--network` argument
as specified in [cloudcli server create](https://github.com/cloudwm/cloudcli/blob/master/docs/cloudcli_server_create.md).
For multiple networks, include the configuration multiple times, example:

```
[global]
; default for all node groups: single public network with auto-assigned ip
default-network = "name=wan,ip=auto"

[nodegroup "ng1"]
; this node group will use the default

[nodegroup "ng2"]
; override the default and attach 2 networks - 1 public and 1 private
network = "name=wan,ip=auto"
network = "name=lan-12345-abcde,ip=auto"
```

### Server Initialization Script

This script is required so that the server will connect to the relevant cluster. The specific script depends on
how you create and manage the cluster.

The script needs to be provided as a base64 encoded string. You can encode your script using the following command: 
`cat script.sh | base64 -w0`.

### Node Templates

When autoscaler makes scaling decisions it checks if added nodes will be able to run pending pods.

If pods have node selectors or affinity restrictions, the autoscaler needs to know if the new nodes will match these requirements.

Following example shows how to set labels on the node template used for scale up checks:

```
[nodegroup "ng1"]
template-label = "disktype=ssd"
template-label = "kubernetes.io/os=linux"
```

This will cause the relevant node group to be considered for pending pods that require those labels.

It's still your responsibility to make sure the actual nodes created by the autoscaler will have these labels.

## Development

Make sure you are inside the `cluster-autoscaler` path of the [autoscaler repository](https://github.com/kubernetes/autoscaler).

Run unit tests:

```
go test -v k8s.io/autoscaler/cluster-autoscaler/cloudprovider/kamatera
```

Setup a Kamatera cluster, you can use [Kamatera RKE2 Kubernetes Terraform](https://github.com/Kamatera/kamatera-rke2-kubernetes-terraform-example/blob/main/README.md)

Get the cluster kubeconfig and set in local file and set in the `KUBECONFIG` environment variable.
Make sure you are connected to the cluster using `kubectl get nodes`.
Create a cloud config file according to the above documentation and set it's path in `CLOUD_CONFIG_FILE` env var.

Build the binary and run it:

```
make build &&\
./cluster-autoscaler-amd64 --cloud-config $CLOUD_CONFIG_FILE --cloud-provider kamatera --kubeconfig $KUBECONFIG -v2
```

Build the docker image:

```
make container
```

Tag and push it to a Docker registry

```
docker tag staging-k8s.gcr.io/cluster-autoscaler-amd64:dev ghcr.io/github_username_lowercase/cluster-autoscaler-amd64
docker push ghcr.io/github_username_lowercase/cluster-autoscaler-amd64
```

Make sure relevant cluster has access to this registry/image.

Follow the documentation for deploying the image and using the autoscaler.
