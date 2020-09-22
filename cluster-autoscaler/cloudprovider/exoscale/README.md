# Cluster Autoscaler for Exoscale

The Cluster Autoscaler (CA) for Exoscale scales worker nodes running in
Exoscale Instance Pools.


## Configuration

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
provide to the CA in your shell:

```sh
export EXOSCALE_API_KEY="EXOxxxxxxxxxxxxxxxxxxxxxxxx"
export EXOSCALE_API_SECRET="xxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```

Next, run the following command from the same shell:

```
./examples/generate-secret.sh
```

Finally, ensure that the `exoscale-secret` *Secret* has been created
successfully by running the following command:

```
kubectl get secret --namespace kube-system exoscale-credentials
```


### Deploying the Cluster Autoscaler

To deploy the CA on your Kubernetes cluster, you can use the manifest provided
as example:

```
kubectl apply -f ./examples/cluster-autoscaler-run-on-master.yaml
```


## ⚠️  Important Notes

* The minimum node group size is 1
* The maximum node group size is computed based on the current [Compute
  instances limit][exo-limits] of the Exoscale account the Cluster Autoscaler
  is running in.
* It is not possible to target which Exoscale Instance Pool will be scaled. The
  Instance Pool candidate for scaling is determined based on the Compute
  instance the Kubernetes node is running on, depending on cluster resource
  constraining events emitted by the Kubernetes scheduler.


[exo-iam]: https://community.exoscale.com/documentation/iam/quick-start/
[exo-limits]: https://portal.exoscale.com/account/limits
[k8s-secrets]: https://kubernetes.io/docs/concepts/configuration/secret/
