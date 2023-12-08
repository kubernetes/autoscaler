# Cluster Autoscaler for Civo Cloud

The cluster autoscaler for Civo Cloud scales worker nodes within any specified Civo Cloud Kubernetes cluster.

## Configuration

As there is no concept of a node group within Civo Cloud's Kubernetes offering, the configuration required is quite
simple. You need to set:

- Your Civo Cloud API Key
- The Kubernetes Cluster's ID (not the name)
- The region of the cluster
- The minimum and maximum number of **worker** nodes you want (the master is excluded)

See the [cluster-autoscaler-standard.yaml](examples/cluster-autoscaler-standard.yaml) example configuration, but to
summarise you should set a `nodes` startup parameter for cluster autoscaler to specify a node group called `workers`
e.g. `--nodes=1:10:workers`.

The remaining parameters can be set via environment variables (`CIVO_API_KEY`, `CIVO_CLUSTER_ID` and `CIVO_REGION`) as in the
example YAML.

It is also possible to get these parameters through a YAML file mounted into the container
(for example via a Kubernetes Secret). The path configured with a startup parameter e.g.
`--cloud-config=/etc/kubernetes/cloud.config`. In this case the YAML keys are `api_url`, `api_key`, `cluster_id` and `region`.
