# Cluster Autoscaler on Azure

The cluster autoscaler on Azure dynamically scales Kubernetes worker nodes. It runs as a deployment in your cluster.

This README will help you get cluster autoscaler running on your Azure Kubernetes cluster.

## Kubernetes Version

Kubernetes v1.10.x or later is required to use cluster autoscaler on Azure. See the "[Releases][]" section in the README for more information.

## CA Version

Cluster autoscaler v1.2.x or later is required for Azure. See the "[Releases][]" section in the README for more information.

> **_NOTE_**: In the deployment manifests referenced below, be sure to replace the `{{ ca_version }}` placeholder with an actual release, such as `v1.14.2`.

## Permissions

Get Azure credentials by running the following [Azure CLI][] command:

```sh
# replace <subscription-id> with yours.
az ad sp create-for-rbac --role="Contributor" --scopes="/subscriptions/<subscription-id>" --output json
```

This will create a new [service principal][] with "Contributor" role scoped to your subscription. Save the JSON output, because it will be needed to configure the cluster autoscaler deployment in the next step.

## Deployment manifests

Cluster autoscaler supports four Kubernetes cluster options on Azure:

- [**vmss**](#vmss-deployment): Autoscale VMSS instances by setting the Azure cloud provider's `vmType` parameter to `vmss` or to an empty string. This supports clusters deployed with [aks-engine][].
- [**standard**](#standard-deployment): Autoscale VMAS instances by setting the Azure cloud provider's `vmType` parameter to `standard`. This supports clusters deployed with [aks-engine][].
- [**aks**](#aks-deployment): Supports an Azure Kubernetes Service ([AKS][]) cluster.
- *DEPRECATED* [**acs**](#acs-deployment): Supports an Azure Container Service ([ACS][]) cluster.

> **_NOTE_**: only the `vmss` option supports scaling down to zero nodes.

### VMSS deployment

Prerequisites:

- Get Azure credentials from the [**Permissions**](#permissions) step above.
- Get the name of the VM scale set associated with the cluster's node pool. You can find this in the [Azure Portal][] or with the `az vmss list` command.

Make a copy of [cluster-autoscaler-vmss.yaml](examples/cluster-autoscaler-vmss.yaml). Fill in the placeholder values for the `cluster-autoscaler-azure` secret data by base64-encoding each of your Azure credential fields.

- ClientID: `<base64-encoded-client-id>`
- ClientSecret: `<base64-encoded-client-secret>`
- ResourceGroup: `<base64-encoded-resource-group>`
- SubscriptionID: `<base64-encoded-subscription-id>`
- TenantID: `<base64-encoded-tenant-id>`

> **_NOTE_**: Use a command such as `echo $CLIENT_ID | base64` to encode each of the fields above.

> **_NOTE_** (optional) to specify the TTL of VMSS ASG cache to prevent throttling issue, please provide the env `AZURE_ASG_CACHE_TTL` in seconds which is set to one hour by default.

In the `cluster-autoscaler` spec, find the `image:` field and replace `{{ ca_version }}` with a specific cluster autoscaler release.

Below that, in the `command:` section, update the `--nodes=` arguments to reference your node limits and VMSS name. For example, if node pool "k8s-nodepool-1-vmss" should scale from 1 to 10 nodes:

```yaml
        - --nodes=1:10:k8s-nodepool-1-vmss
```

or to autoscale multiple VM scale sets:

```yaml
        - --nodes=1:10:k8s-nodepool-1-vmss
        - --nodes=1:10:k8s-nodepool-2-vmss
```

Note that it doesn't mean the number of nodes in nodepool is restricted in the 
range from 1 to 10. It means when ca is downscaling (upscaling) the nodepool, 
it will never break the limit of 1 (10). If the current node pool size is lower than the specified minimum or greater than the specified maximum when you enable autoscaling, the autoscaler waits to take effect until a new node is needed in the node pool or until a node can be safely deleted from the node pool.

To allow scaling similar node pools simultaneously, or when using separate node groups per zone and to keep nodes balanced across zones, use the `--balance-similar-node-groups` flag (default false). Add it to the `command` section to enable it:

```yaml
        - --balance-similar-node-groups=true
```

See the [FAQ](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#im-running-cluster-with-nodes-in-multiple-zones-for-ha-purposes-is-that-supported-by-cluster-autoscaler) for more details.

Save the updated deployment manifest, then deploy cluster-autoscaler by running:

```sh
kubectl create -f cluster-autoscaler-vmss.yaml
```

To run a cluster autoscaler pod on a master node, the deployment should tolerate the `master` taint, and `nodeSelector` should be used to schedule pods. Use [cluster-autoscaler-vmss-master.yaml](examples/cluster-autoscaler-vmss-master.yaml) in this case.

To run a cluster autoscaler pod with Azure managed service identity (MSI), use [cluster-autoscaler-vmss-msi.yaml](examples/cluster-autoscaler-vmss-msi.yaml) instead.

### Standard deployment

Prerequisites:

- Get Azure credentials from the [**Permissions**](#permissions) step above.
- Get the name of the initial Azure deployment resource for the cluster. You can find this in the [Azure Portal](https://portal.azure.com) or with the `az deployment list` command. If there are multiple deployments, get the name of the first one.

Make a copy of [cluster-autoscaler-standard-master.yaml](examples/cluster-autoscaler-standard-master.yaml). Fill in the placeholder values for the `cluster-autoscaler-azure` secret data by base64-encoding each of your Azure credential fields.

- ClientID: `<base64-encoded-client-id>`
- ClientSecret: `<base64-encoded-client-secret>`
- ResourceGroup: `<base64-encoded-resource-group>`
- SubscriptionID: `<base64-encoded-subscription-id>`
- TenantID: `<base64-encoded-tenant-id>`
- Deployment: `<base64-encoded-azure-initial-deployment-name>`

> **_NOTE_**: Use a command such as `echo $CLIENT_ID | base64` to encode each of the fields above.

In the `cluster-autoscaler` spec, find the `image:` field and replace `{{ ca_version }}` with a specific cluster autoscaler release.

Below that, in the `command:` section, update the `--nodes=` arguments to reference your node limits and node pool name (tips: node pool name is NOT availability set name, e.g., the corresponding node pool name of the availability set 
`agentpool1-availabilitySet-xxxxxxxx` would be `agentpool1`). For example, if node pool "k8s-nodepool-1" should scale from 1 to 10 nodes:

```yaml
        - --nodes=1:10:k8s-nodepool-1
```

or to autoscale multiple VM scale sets:

```yaml
        - --nodes=1:10:k8s-nodepool-1
        - --nodes=1:10:k8s-nodepool-2
```

Create the Azure deploy parameters secret `cluster-autoscaler-azure-deploy-parameters` by running:

```sh
kubectl -n kube-system create secret generic cluster-autoscaler-azure-deploy-parameters --from-file=deploy-parameters=./_output/<your-output-path>/azuredeploy.parameters.json
```

Then deploy cluster-autoscaler by running:

```sh
kubectl create -f cluster-autoscaler-standard-master.yaml
```

To run a cluster autoscaler pod with Azure managed service identity (MSI), use [cluster-autoscaler-standard-msi.yaml](examples/cluster-autoscaler-standard-msi.yaml) instead.

> **_WARNING_**: Cluster autoscaler depends on user-provided deployment parameters to provision new nodes. After upgrading your Kubernetes cluster, cluster autoscaler must also be redeployed with new parameters to prevent provisioning nodes with an old version.

### AKS deployment

Autoscaling VM scale sets with AKS is supported for Kubernetes v1.12.4 and later. The option to enable cluster autoscaler is available in the [Azure Portal][] or with the [Azure CLI][]:

```sh
az aks create \
  --resource-group myResourceGroup \
  --name myAKSCluster \
  --kubernetes-version 1.13.5 \
  --node-count 1 \
  --enable-vmss \
  --enable-cluster-autoscaler \
  --min-count 1 \
  --max-count 3
```

Please see the [AKS autoscaler documentation][] for details.

### ACS deployment

> **_NOTE_**: [ACS will retire on January 31, 2020][ACS].

Prerequisites:

- Get Azure credentials from the [**Permissions**](#permissions) step above.
- Get the cluster name with the `az acs list` command.
- Get the name of a node pool from the value of the label **agentpool**

```sh
kubectl get nodes --show-labels
```

Make a copy of [cluster-autoscaler-containerservice](examples/cluster-autoscaler-containerservice.yaml). Fill in the placeholder values for the `cluster-autoscaler-azure` secret data by base64-encoding each of your Azure credential fields.

- ClientID: `<base64-encoded-client-id>`
- ClientSecret: `<base64-encoded-client-secret>`
- ResourceGroup: `<base64-encoded-resource-group>` (Note: ResourceGroup is case-sensitive)
- SubscriptionID: `<base64-encoded-subscription-id>`
- TenantID: `<base64-encoded-tenant-id>`
- ClusterName: `<base64-encoded-clustername>`
- NodeResourceGroup: `<base64-encoded-node-resource-group>` (Note: node resource group is not resource group and can be obtained in the corresponding label of the nodepool)

> **_NOTE_**: Use a command such as `echo $CLIENT_ID | base64` to encode each of the fields above.

In the `cluster-autoscaler` spec, find the `image:` field and replace `{{ ca_version }}` with a specific cluster autoscaler release.

Below that, in the `command:` section, update the `--nodes=` arguments to reference your node limits and node pool name. For example, if node pool "k8s-nodepool-1" should scale from 1 to 10 nodes:

```yaml
        - --nodes=1:10:k8s-nodepool-1
```

or to autoscale multiple VM scale sets:

```yaml
        - --nodes=1:10:k8s-nodepool-1
        - --nodes=1:10:k8s-nodepool-2
```

Then deploy cluster-autoscaler by running

```sh
kubectl create -f cluster-autoscaler-containerservice.yaml
```


[ACS]: https://azure.microsoft.com/updates/azure-container-service-will-retire-on-january-31-2020/
[AKS]: https://docs.microsoft.com/azure/aks/
[AKS autoscaler documentation]: https://docs.microsoft.com/azure/aks/autoscaler
[aks-engine]: https://github.com/Azure/aks-engine
[Azure CLI]: https://docs.microsoft.com/cli/azure/install-azure-cli
[Azure Portal]: https://portal.azure.com
[Releases]: ../../README.md#releases
[service principal]: https://docs.microsoft.com/azure/active-directory/develop/app-objects-and-service-principals
