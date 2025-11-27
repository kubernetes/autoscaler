## Slicer Cloud Provider for Cluster Autoscaler

The cluster autoscaler for [Slicer](https://slicervm.com/) scales nodes on lightweight slicer microVMs.

The architecture is as follows:

* Slicer runs a K3s control plane on one virtualisation host.
* Slicer runs all agents on one or more additional virtualisation hosts running the Slicer REST API. Starting off with zero microVMs and relying on cluster autoscaler to add new ones as Pods cannot be scheduled to the existing set of nodes.

Check the documentation on [SlicerVM.com](https://docs.slicervm.com/examples/autoscaling-k3s/) for instructions on how to setup the cluster-autoscaler with with Slicer.

## Configuration

The `cluster-autoscaler` with Slicer needs a configuration file to work by using the `--cloud-config` parameter, it is an INI file with the following fields:

| Key | Value | Mandatory | Default |
|-----|-------|-----------|---------|
| `global/k3s-url` | The URL of the K3s control plane API server | yes | none |
| `global/k3s-token` | The K3s join token for adding new agent nodes | yes | none |
| `global/default-min-size` | Default minimum size of a node group (must be > 0) | no | 1 |
| `global/default-max-size` | Default maximum size of a node group | no | 8 |
| `nodegroup \"slicer_host_group_name\"/slicer-url` | The URL of the Slicer API server for this node group | yes | none |
| `nodegroup \"slicer_host_group_name\"/slicer-token` | The authentication token for the Slicer API server | yes | none |
| `nodegroup \"slicer_host_group_name\"/min-size` | Minimum size for a specific node group | no | global/defaut-min-size |
| `nodegroup \"slicer_host_group_name\"/max-size` | Maximum size for a specific node group | no | global/defaut-max-size |

## Development

Follow the instructions in the [slicer docs](https://docs.slicervm.com/examples/autoscaling-k3s/) to setup a K3S cluster and host groups for nodes.

Make sure you are inside the `cluster-autoscaler` path of the [autoscaler repository](https://github.com/kubernetes/autoscaler).

### Run out of cluster

Start the cluster-autoscaler:

```bash
#!/bin/bash
go run . \
    --cloud-provider=slicer \
    --kubeconfig $HOME/k3s-cp-kubeconfig \
    --scale-down-enabled=true \
    --scale-down-delay-after-add=30s \
    --scale-down-unneeded-time=30s \
    --expendable-pods-priority-cutoff=-10 \
    --cloud-config="$HOME/cloud-config.ini" \
    --v=4
```

### Run in cluster.

Build and publish an image:

```sh
REGISTRY=ttl.sh/openfaasltd BUILD_TAGS=slicer TAG=dev make dev-release
```

Create a the cloud-config secret:

```sh
kubectl create secret generic cluster-autoscaler-cloud-config \
  --from-file=cloud-config=cloud-config.ini \
  -n kube-system
```

Create a `values.yaml` for the cluster-autoscaler chart:

```yaml
image:
  repository: ttl.sh/openfaasltd/cluster-autoscaler-slicer-amd64
  tag: dev

cloudProvider: slicer

fullnameOverride: cluster-autoscaler-slicer

autoDiscovery:
  clusterName: k3s-slicer

# Mount the cluster-autoscaler-cloud-config secret
extraVolumeSecrets:
  cluster-autoscaler-cloud-config:
    name: cluster-autoscaler-cloud-config
    mountPath: /etc/slicer/
    items:
      - key: cloud-config
        path: cloud-config

# All your required parameters
extraArgs:
  cloud-config: /etc/slicer/cloud-config
  # Standard logging
  logtostderr: true
  stderrthreshold: info
  v: 4

  scale-down-enabled: true
  scale-down-delay-after-add: "30s"
  scale-down-unneeded-time: "30s"
  expendable-pods-priority-cutoff: -10
```

Deploy with Helm:

```sh
helm install cluster-autoscaler-slicer charts/cluster-autoscaler \
  --namespace=kube-system \
  --values=values.yaml
```

To test the autoscaler do one of the following:

* Scale a deployment higher than can fit on the current set of control-plane nodes, then wait for the autoscaler to scale up the cluster.
* Or, create a taint / affinity / anti-affinity rule that will prevent a pod from being scheduled to the existing set of nodes, then wait for the autoscaler to scale up the cluster.
