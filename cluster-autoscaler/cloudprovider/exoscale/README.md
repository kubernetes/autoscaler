# Cluster Autoscaler for Exoscale

The Cluster Autoscaler (CA) for Exoscale scales worker nodes running in
Exoscale SKS Nodepools or Instance Pools.

- [Cluster Autoscaler for Exoscale](#cluster-autoscaler-for-exoscale)
  - [Configuration](#configuration)
    - [Authenticating to the Exoscale API](#authenticating-to-the-exoscale-api)
    - [Optional configuration](#optional-configuration)
  - [Deployment](#deployment)
    - [Helm](#helm)
    - [Manifest](#manifest)
  - [⚠️  Important Notes](#️--important-notes)

## Configuration

### Authenticating to the Exoscale API

> Note: the following guide assumes you have the permissions to create
> resources in the `kube-system` namespace of the target Kubernetes cluster.

In order to interact with the Exoscale API, the Exoscale CA must be configured
with API credentials. This can be achieved using Kubernetes
[*Secrets*][k8s-secrets], by exposing those as container environment variables.

We provide a convenience script that generates and applies a k8s manifest
declaring Exoscale API credentials as a k8s *Secret* in your cluster from your
local shell environment variables: once created, this *Secret* can be used in
the CA *Deployment*.

First, start by exporting the Exoscale API credentials (we recommend that you
create dedicated API credentials using the [Exoscale IAM][exo-iam] service) to
provide to the CA in your shell, as well as the zone the target Kubernetes
cluster is located in:

```sh
export EXOSCALE_API_KEY="EXOxxxxxxxxxxxxxxxxxxxxxxxx"
export EXOSCALE_API_SECRET="xxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export EXOSCALE_ZONE="ch-gva-2"
```

Next, run the following command from the same shell:

```
./examples/generate-secret.sh
```

Next, ensure that the `exoscale-api-credentials` *Secret* has been created
successfully by running the following command:

```
kubectl get secret --namespace kube-system exoscale-api-credentials
```

Finally, a `EXOSCALE_ZONE` variable must be set to the target Kubernetes
cluster zone along with the API credentials in the CA *Deployment* shell
environment.

You can restrict API operation your IAM key can perform:

* When deploying the Cluster Autoscaler in SKS, your can restrict your IAM access key
to these API operations :

```
evict-sks-nodepool-members
get-instance
get-instance-pool
get-operation
get-quota
list-sks-clusters
scale-sks-nodepool
```

* When deploying the Cluster Autoscaler in an unmanaged cluster, the cluster needs to have
nodes belonging to at least an instance-pool. In this case, you can rather restrict your
IAM key to these API operations:

```
evict-instance-pool-members
get-instance
get-instance-pool
get-operation
get-quota
scale-instance-pool
```

### Optional configuration

By default, all nodepools in the k8s cluster are considered for scaling.
The flag `--nodes=<min>:<max>:<nodepool-name>` may be specified to limit the minimum and
maximum size of a particular nodepool.

## Deployment

### Helm

See the [Helm Chart README](https://github.com/kubernetes/autoscaler/tree/master/charts/cluster-autoscaler).

### Manifest

To deploy the CA on your Kubernetes cluster, you can use the manifest provided as example:

```bash
kubectl apply -f ./examples/cluster-autoscaler-run-on-control-plane.yaml
```

This manifest contains a deployment which is designed to schedule the CA Pod on control-plane nodes.
If you want to deploy the CA Pod on regular Nodes (not on the control-plane) or in SKS, you can
use this manifest instead:

```bash
kubectl apply -f ./examples/cluster-autoscaler.yaml
```

## ⚠️  Important Notes

* The minimum and maximum node group size of particular nodepools
  may be specified via the `--nodes` flag, if omitted (default),
  the minimum is 1 and maximum is computed based on the current [Compute instances limit][exo-limits]
  of the Exoscale account the Cluster Autoscaler is running in.
* The Instance Pool candidate for scaling is determined based on the Compute
  instance the Kubernetes node is running on, depending on cluster resource
  constraining events emitted by the Kubernetes scheduler.


[exo-iam]: https://community.exoscale.com/documentation/iam/quick-start/
[exo-limits]: https://portal.exoscale.com/organization/quotas
[k8s-secrets]: https://kubernetes.io/docs/concepts/configuration/secret/
