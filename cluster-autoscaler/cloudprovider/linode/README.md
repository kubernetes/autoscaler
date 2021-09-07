# Cluster Autoscaler for Linode

The cluster autoscaler for Linode scales nodes in a LKE cluster.

## Linode Kubernetes Engine

Linode Kubernetes Engine ([LKE](https://www.linode.com/docs/guides/deploy-and-manage-a-cluster-with-linode-kubernetes-engine-a-tutorial/)) is the managed K8s solution provided by Linode.

## Configuration

The cluster autoscaler automatically selects every every Node Pool in the LKE Cluster with autoscaler enabled (via the API), so there is no need define the `node-group-auto-discovery` or `nodes` flags, see [examples/cluster-autoscaler-autodiscover.yaml](examples/cluster-autoscaler-autodiscover.yaml) for an example of a kubernetes deployment.

You can optionally use a JSON `cloud-config` file to configure the autoscaler. All required fields have environment variable bindings, so the cloud config is only necessary for additional configuration. See [examples/cluster-autoscaler-secret.yaml](examples/cluster-autoscaler-secret.yaml) for an example config.

| Key | Value | Required | Default | Environment Variable |
|-----|-------|-----------|---------|----|
| token | Linode API Token with Read/Write permission for LKE Clusters and NodePools | yes | none | LINODE_API_TOKEN |
| clusterID | The numeric ID of the LKE cluster | yes | none | LKE_CLUSTER_ID
| baseURL | The base URL to set on the Linodego client | no | api.linode.com |
| apiVersion | The Linode API version to set on the Linodego client | no | v4 |

## Development

Make sure you are inside the `cluster-autoscaler` path of the [autoscaler repository](https://github.com/kubernetes/autoscaler).

1. Build the `cluster-autoscaler` binary
    ```
    make build-in-docker
    ```

2. Build the docker image
    ```
    REGISTRY=linode TAG=canary make make-image
    ```

3. Push the docker image to DockerHub:
    ```
    REGISTRY=linode TAG=canary make push-image
    ```
