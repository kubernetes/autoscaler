# Cluster Autoscaler on Ionos Cloud Managed Kubernetes

The cluster autoscaler for the Ionos Cloud scales worker nodes within Managed Kubernetes cluster
node pools. It can be deployed as `Deployment` in your cluster.

## Deployment

### Managed

Managed autoscaling can be enabled or disabled via Kubernetes Manager in the [DCD](https://dcd.ionos.com/latest/)
or [API](https://devops.ionos.com/api/cloud/v5/#update-a-nodepool).
In this case a `Deployment` is not needed, since it will be deployed in the managed Kubernetes controlplane.

### In-cluster

Build and push a docker image in the `cluster-autoscaler` directory:

```sh
make build-in-docker BUILD_TAGS=ionoscloud
make make-image BUILD_TAGS=ionoscloud TAG='<tag>' REGISTRY='<registry>'
make push-image BUILD_TAGS=ionoscloud TAG='<tag>' REGISTRY='<registry>'
```

If you don't have a token, generate one:

```sh
curl -u '<username>' -p '<password>' https://api.ionos.com/auth/v1/tokens/generate
```

Edit [`cluster-autoscaler-standard.yaml`](./examples/cluster-autoscaler-standard.yaml) and deploy it:

```console
kubectl apply -f examples/cluster-autoscaler-standard.yaml
```
