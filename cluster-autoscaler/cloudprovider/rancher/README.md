# Cluster Autoscaler for Rancher with RKE2

This cluster autoscaler for Rancher scales nodes in clusters which use RKE2
provisioning (Rancher v2.6+). It uses a combination of the Rancher API and the
underlying cluster-api types of RKE2.

## Configuration

The `cluster-autoscaler` for Rancher needs a configuration file to work by using
`--cloud-config` parameter. A `cluster-autoscaler` instance can target a single
downstream RKE2 cluster specified in the config. An up-to-date example can be
found in [examples/config.yaml](./examples/config.yaml).

### Configuration via environment variables

In order to override URL, token or clustername use following environment variables:
 - RANCHER_URL
 - RANCHER_TOKEN
 - RANCHER_CLUSTER_NAME

### Permissions

The Rancher server account provided in the `cloud-config` requires the
following permissions on the Rancher server:

* Get/Update of the `clusters.provisioning.cattle.io` resource to autoscale
* List of `machines.cluster.x-k8s.io` in the namespace of the cluster resource

## Running the Autoscaler

The `cluster-autoscaler` can be run inside the RKE2 cluster, on the Rancher
server cluster or on a completely separate machine. To run it outside the RKE2
cluster, make sure to provide a kubeconfig with `--kubeconfig`.

To start the autoscaler with the Rancher provider, the cloud provider needs to
be specified:

```bash
cluster-autoscaler --cloud-provider=rancher --cloud-config=config.yaml
```

## Enabling Autoscaling

In order for the autoscaler to function, the RKE2 cluster needs to be
configured accordingly. The autoscaler works by adjusting the `quantity` of a
`machinePool` dynamically. For the autoscaler to know the min/max size of a
`machinePool` we need to set a few annotations using the
`machineDeploymentAnnotations` field. That field has been chosen because
updating it does not trigger a full rollout of a `machinePool`.

```yaml
apiVersion: provisioning.cattle.io/v1
kind: Cluster
spec:
  rkeConfig:
    machinePools:
    - name: pool-1
      quantity: 1
      workerRole: true
      machineDeploymentAnnotations:
        cluster.provisioning.cattle.io/autoscaler-min-size: "1"
        cluster.provisioning.cattle.io/autoscaler-max-size: "3"
```

Optionally in order to enable scaling a `machinePool` from and to 0 nodes, we
need to add a few more annotations to let the autoscaler know, which resources
a single node in a pool provides:

```yaml
apiVersion: provisioning.cattle.io/v1
kind: Cluster
spec:
  rkeConfig:
    machinePools:
    - name: pool-1
      machineDeploymentAnnotations:
        cluster.provisioning.cattle.io/autoscaler-min-size: "0"
        cluster.provisioning.cattle.io/autoscaler-max-size: "3"
        cluster.provisioning.cattle.io/autoscaler-resource-cpu: "1"
        cluster.provisioning.cattle.io/autoscaler-resource-ephemeral-storage: 50Gi
        cluster.provisioning.cattle.io/autoscaler-resource-memory: 4Gi
```
