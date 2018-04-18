# Cluster Autoscaler on Azure

The cluster autoscaler on Azure scales worker nodes within any specified autoscaling group. It will run as a Kubernetes deployment in your cluster. This README will go over some of the necessary steps required to get the cluster autoscaler up and running.

## Kubernetes Version

Kubernetes v1.10.X and Cluster autoscaler v1.2+  are required to run on Azure.

Cluster autoscaler support two VM types with Azure cloud provider:

- **vmss**: For kubernetes cluster running on VMSS instances. Azure cloud provider's `vmType` parameter must be configured as 'vmss'. It requires Kubernetes with Azure VMSS support ([kubernetes#43287](https://github.com/kubernetes/kubernetes/issues/43287)).
- **standard**: For kubernetes cluster running on VMAS instances. Azure cloud provider's `vmType` parameter must be configured as 'standard' or left as empty string. It only supports Kubernetes cluster deployed via [acs-engine](https://github.com/Azure/acs-engine).

## Permissions

Get azure credentials by running the following command

```sh
# replace <subscription-id> with yours.
az ad sp create-for-rbac --role="Contributor" --scopes="/subscriptions/<subscription-id>" --output json
```

## Deployment manifests

### VMSS deployment

Pre-requirements:

- Get credentials from above `permissions` step.
- Get the scale set name which is used for nodes scaling.
- Encode each data with base64.

Fill the values of cluster-autoscaler-azure secret in [cluster-autoscaler-vmss.yaml](cluster-autoscaler-vmss.yaml), including

- ClientID: `<base64-encoded-client-id>`
- ClientSecret: `<base64-encoded-client-secret>`
- ResourceGroup: `<base64-encoded-resource-group>`
- SubscriptionID: `<base64-encode-subscription-id>`
- TenantID: `<base64-encoded-tenant-id>`

> Note that all data above should be encoded with base64.

And fill the node groups in container command by `--nodes`, e.g.

```yaml
        - --nodes=1:10:vmss1
```

or multiple node groups:

```yaml
        - --nodes=1:10:vmss1
        - --nodes=1:10:vmss2
```

Then deploy cluster-autoscaler by running

```sh
kubectl create -f cluster-autoscaler-vmss.yaml
```

To run a CA pod in master node - CA deployment should tolerate the master `taint` and `nodeSelector` should be used to schedule the pods in master node.

```sh
kubectl create -f cluster-autoscaler-vmss-master.yaml
```

To run a CA pod with Azure managed service identity (MSI), use [cluster-autoscaler-vmss-msi.yaml](cluster-autoscaler-vmss-msi.yaml) instead:

```sh
kubectl create -f cluster-autoscaler-vmss-msi.yaml
```

### Standard deployment

Pre-requirements:

- Get credentials from above `permissions` step.
- Get the initial Azure deployment name from azure portal. If you have multiple deployments (e.g. have run `acs-engine scale` command), make sure to get the first one.
- Get a node pool name for nodes scaling from acs-engine deployment manifests
- Encode each data with base64.

Fill the values of cluster-autoscaler-azure secret in [cluster-autoscaler-standard-master.yaml](cluster-autoscaler-standard-master.yaml), including

- ClientID: `<base64-encoded-client-id>`
- ClientSecret: `<base64-encoded-client-secret>`
- ResourceGroup: `<base64-encoded-resource-group>`
- SubscriptionID: `<base64-encode-subscription-id>`
- TenantID: `<base64-encoded-tenant-id>`
- Deployment: `<base64-encoded-azure-initial-deploy-name>`

> Note that all data above should be encoded with base64.

And fill the node groups in container command by `--nodes`, e.g.

```yaml
        - --nodes=1:10:agentpool1
```

or multiple node groups:

```yaml
        - --nodes=1:10:agentpool1
        - --nodes=1:10:agentpool2
```

Create Azure deploy parameters secret `cluster-autoscaler-azure-deploy-parameters` by running

```sh
kubectl -n kube-system create secret generic cluster-autoscaler-azure-deploy-parameters --from-file=deploy-parameters=./_output/<your-output-path>/azuredeploy.parameters.json
```

Then deploy cluster-autoscaler by running

```sh
kubectl create -f cluster-autoscaler-standard-master.yaml
```

To run a CA pod with Azure managed service identity (MSI), use [cluster-autoscaler-standard-msi.yaml](cluster-autoscaler-standard-msi.yaml) instead:

```sh
kubectl create -f cluster-autoscaler-standard-msi.yaml
```