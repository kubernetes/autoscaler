# KEP-5546: Scaling based on container count

<!-- toc -->
- [Summary](#summary)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
    - [Notes](#notes)
    - [Risks and Mitigations](#risks-and-mitigations)
- [Design Details](#design-details)
    - [Test Plan](#test-plan)
<!-- /toc -->

## Summary

Currently Addon Resizer supports scaling based on the number of nodes. Some workloads use resources proportionally to
the number of containers in the cluster. Since number of containers per node is very different in different clusters
it's more resource-efficient to scale such workloads based directly on the container count. 

### Goals

- Allow scaling workloads based on count of containers in a cluster.
- Allow this for Addon Resizer 1.8 ([used by metrics  server]).

### Non-Goals

- Using both node and container count to scale workloads.
- Bringing this change to the `master` branch of Addon Resizer.

## Proposal

Add flag `--scaling-mode` to Addon Resizer on the [`addon-resizer-release-1.8`] branch. Flag will
have two valid values:

- `node-proportional` - default, current behavior.
- `container-proportional` - addon resizer will set resources, using the same algorithm it's using now but using number
  of containers where it's currently using number of nodes.

### Notes

Addon Resizer 1.8 assumes in multiple places that it's scaling based on the number of nodes:

- [Flag descriptions] that directly reference node counts (`--extra-cpu`, `--extra-memory`, `--extra-storage`, and
  `--minClusterSize`) will need to be updated to instead refer to cluster size.
- [README] will need to be updated to reference cluster size instead of node count and explain that cluster size refers
  to either node count or container count, depending on the value of the `--scaling-mode` flag.
- Many variable names in code which now refer to node count will refer to cluster size and should be renamed accordingly.

In addition to implementing the feature we should also clean up the code and documentation.  

### Risks and Mitigations

One potential risk is that Addon resizer can obtain cluster size (node count or container count):
- from metrics or
- by querying Cluster Api Server to list all objects of the appropriate type

depending on the configuration. There can be many times more containers in a cluster that there are nodes. So listing
all containers could result in higher load on the Cluster API server. Since Addon Resizer is requesting very few fields
I don't expect this effect to be noticeable.

Also I expect metrics-server to test for this before using the feature and any other users of Addon Resizer are likely
better off using metrics (which don't have this problem). 

## Design Details

- Implement function `kubernetesClient.CountContainers()`. It will be analogous to the existing
  [`kubernetesClient.CountNodes()`] function.
  - If using metrics to determine number of containers in the cluster:
    - Fetch pod metrics (similar to [fetching node metrics] but use `/pods` URI instead of `/nodes`).
    - For each pod obtain number of containers (length of the `containers` field).
    - Sum container counts for all pods.
  - If using API server:
    - Fetch list pods (similar to [listing nodes])
      - Fetch only [`Spec.InitContainers`], [`Spec.Containers`], and [`Spec.EphemeralContainers`] fields.
      - Exclude pods in terminal states ([selector excluding pods in terminal states in VPA])
    - Sum container count over pods.
- Add the `--scaling-mode` flag, with two valid values:
  - `node-proportional` - default, current behavior, scaling based on clusters node count and
  - `container-proportional` - new behavior, scaling based on clusters container count
- Pass value indicating if we should use node count or container count to the [`updateResources()`] function.
- In `updateResources()` use node count or container count, depending on the value.

Check that listing containers directly works

Coinsider listing pods, getting containers only for working pods

### Test Plan

In addition to unit tests we will run manual e2e test:

- Create config based on [`example.yaml`] but scaling the deployment based on the number of containers in the cluster.
- Create config starting deployment with 100 `pause` containers.

Test the feature by:

- Starting the deployment scaled by Addon Resizer, based on node count.
- Observe size of the deployment and that it's stable.
- Start deployment with 100 `pause` containers.
- Observe the scaled deployment change resources appropriately.

Test the node-based scaling:

- Apply [`example.yaml`].
- Observe amount and stability assigned resources.
- Resize cluster.
- Observe change in assigned resources.

Both tests should be performed with metrics- and API- based scaling.

[used by metrics  server]: https://github.com/kubernetes-sigs/metrics-server/blob/0c47555e9b49cfe0719db1a0b7fb6c8dcdff3d38/charts/metrics-server/values.yaml#L121
[`addon-resizer-release-1.8`]: https://github.com/kubernetes/autoscaler/tree/addon-resizer-release-1.8
[Flag descriptions]: https://github.com/kubernetes/autoscaler/blob/da500188188d275a382be578ad3d0a758c3a170f/addon-resizer/nanny/main/pod_nanny.go#L47
[README]: https://github.com/kubernetes/autoscaler/blob/da500188188d275a382be578ad3d0a758c3a170f/addon-resizer/README.md?plain=1#L1
[`kubernetesClient.CountNodes()`]: https://github.com/kubernetes/autoscaler/blob/da500188188d275a382be578ad3d0a758c3a170f/addon-resizer/nanny/kubernetes_client.go#L58
[fetching node metrics]: https://github.com/kubernetes/autoscaler/blob/da500188188d275a382be578ad3d0a758c3a170f/addon-resizer/nanny/kubernetes_client.go#L150
[listing nodes]: https://github.com/kubernetes/autoscaler/blob/da500188188d275a382be578ad3d0a758c3a170f/addon-resizer/nanny/kubernetes_client.go#L71
[`Spec.InitContainers`]: https://github.com/kubernetes/api/blob/1528256abbdf8ff2510112b28a6aacd239789a36/core/v1/types.go#L3143
[`Spec.Containers`]: https://github.com/kubernetes/api/blob/1528256abbdf8ff2510112b28a6aacd239789a36/core/v1/types.go#L3150
[`Spec.EphemeralContainers`]: https://github.com/kubernetes/api/blob/1528256abbdf8ff2510112b28a6aacd239789a36/core/v1/types.go#L3158
[`Status.Phase`]: https://github.com/kubernetes/api/blob/1528256abbdf8ff2510112b28a6aacd239789a36/core/v1/types.go#L4011
[selector excluding pods in terminal states in VPA]: https://github.com/kubernetes/autoscaler/blob/04e5bfc88363b4af9fdeb9dfd06c362ec5831f51/vertical-pod-autoscaler/e2e/v1beta2/common.go#L195
[`updateResources()`]: https://github.com/kubernetes/autoscaler/blob/da500188188d275a382be578ad3d0a758c3a170f/addon-resizer/nanny/nanny_lib.go#L126
[`example.yaml`]: https://github.com/kubernetes/autoscaler/blob/c8d612725c4f186d5de205ed0114f21540a8ed39/addon-resizer/deploy/example.yaml