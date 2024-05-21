<!--- For help refer to https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.20.md?plain=1 as example --->

- [v1.29.2](#v1290)
    - [Synced with which upstream CA](#synced-with-which-upstream-ca)
    - [Changes made](#changes-made)
        - [During merging](#during-merging)
        - [During vendoring k8s](#during-vendoring-k8s)
        - [Others](#others)


# v1.29.0


## Synced with which upstream CA

[v1.29.2](https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.29.2)

## Changes made
- See general release notes of 1.29.2: https://github.com/kubernetes/autoscaler/releases/tag/cluster-autoscaler-1.29.2
- New flag added in autoscaling options: flag.Bool("scale-down-delay-type-local", false, "Should --scale-down-delay-after-* flags be applied locally per nodegroup or globally across all nodegroups")
- New flag added in autoscaling options: pflag.StringSlice("bypassed-scheduler-names", []string{}, fmt.Sprintf("Names of schedulers to bypass. If set to non-empty value, CA will not wait for pods to reach a certain age before triggering a scale-up."))
- New flag added in autoscaling options: flag.String("drain-priority-config", "",
  "List of ',' separated pairs (priority:terminationGracePeriodSeconds) of integers separated by ':' enables priority evictor. Priority evictor groups pods into priority groups based on pod priority and evict pods in the ascending order of group priorities"+
  "--max-graceful-termination-sec flag should not be set when this flag is set. Not setting this flag will use unordered evictor by default."+
  "Priority evictor reuses the concepts of drain logic in kubelet(https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/2712-pod-priority-based-graceful-node-shutdown#migration-from-the-node-graceful-shutdown-feature)."+
  "Eg. flag usage:  '10000:20,1000:100,0:60'")
- New flag added in autoscaling options: flag.Bool("dynamic-node-delete-delay-after-taint-enabled", false, "Enables dynamic adjustment of NodeDeleteDelayAfterTaint based of the latency between CA and api-server")
- New flag added in autoscaling options: flag.String("kube-api-content-type", "application/vnd.kubernetes.protobuf", "Content type of requests sent to apiserver.")

### During merging
- import package for `machine-controller-manager-provider-azure` was updated after vendoring latest version.

### During vendoring k8s
- mcm v0.50.1 -> 0.53.0
- mcm-provider-azure v0.11.1 -> v0.12.1

### Others
- [Release matrix](../README.md#releases-gardenerautoscaler) of Gardener Autoscaler updated.