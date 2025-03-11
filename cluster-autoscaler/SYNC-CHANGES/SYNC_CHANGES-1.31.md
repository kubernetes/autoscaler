<!--- For help refer to https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.20.md?plain=1 as example --->

- [v1.31.0](#v1310)
  - [Synced with which upstream CA](#synced-with-which-upstream-ca)
  - [Changes made](#changes-made)
    - [During merging](#during-merging)
    - [During vendoring k8s](#during-vendoring-k8s)
    - [Others](#others)


# v1.31.0


## Synced with which upstream CA

[v1.31.1](https://github.com/kubernetes/autoscaler/tree/cluster-autoscaler-1.31.1/cluster-autoscaler)

## Changes made
 - See general release notes of 1.31.1: https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.31.1
 - New flags added that help cluster-autoscaler provision nodes before all pending pods are created and marked as unschedulable by scheduler
    - New flag added: flag.Bool("enable-proactive-scaleup", false, "Whether to enable/disable proactive scale-ups, defaults to false")
    - New flag added: flag.Int("pod-injection-limit", 5000, "Limits total number of pods while injecting fake pods. If unschedulable pods already exceeds the limit, pod injection is disabled but pods are not truncated.")
- New flag added: flag.Duration("max-binpacking-time", 5*time.Minute, "Maximum time spend on binpacking for a single scale-up. If binpacking is limited by this, scale-up will continue with the already calculated scale-up options.")
- Support added for ProvisioningRequest v1 API
- Adds support for using in-cluster Kubernetes configuration as fallback in case neither path to Kubernetes configuration nor location of master is provided
- New expander added: `least-nodes`
- The following options have been added to `AutoscalingOptions`
  ```
  // MaxBinpackingTime is the maximum time spend on binpacking for a single scale-up.
  // If binpacking is limited by this, scale-up will continue with the already calculated scale-up options.
  MaxBinpackingTime time.Duration
  // AsyncNodeGroupsEnabled tells if CA creates/deletes node groups asynchronously.
  AsyncNodeGroupsEnabled bool
  ```

### During merging
  - `mcm/fakeclient.FakeObjectTracker` implements new methods to be in sync with `k8stesting.ObjectTracker`

### During vendoring k8s
- mcm v0.53.0 -> 0.55.0
- k8s.io/api v0.30.1 -> v0.31.1
- github.com/onsi/ginkgo/v2 v2.16.0 -> v2.19.0

### Others
- [Release matrix](../README.md#releases-gardenerautoscaler) of Gardener Autoscaler updated.