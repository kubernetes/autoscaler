<!--- For help refer to https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.20.md?plain=1 as example --->

- [v1.21.0](#v1210)
    - [Synced with which upstream CA](#synced-with-which-upstream-ca)
    - [Changes made](#changes-made)
        - [To FAQ](#to-faq)
        - [During merging](#during-merging)
        - [During vendoring k8s](#during-vendoring-k8s)
        - [Others](#others)
        

# v1.21.0


## Synced with which upstream CA

[v1.21.0](https://github.com/kubernetes/autoscaler/tree/cluster-autoscaler-1.21.0/cluster-autoscaler)

## Changes made

### To FAQ

- ignored integration test package files during running unit tests.
### During merging

- Removed `clusterapi.ProviderName` in builder_all.go
### During vendoring k8s
- Used the new `update_vendor.sh` to vendor.
- “k8s.io/kubernetes/pkg/kubelet/apis" => “k8s.io/kubelet/pkg/apis"
- mcm 0.42 -> mcm 0.44.1
- mcm-provider-aws 0.8.0 -> 0.9.0
- mcm-provider-azure 0.5.0 -> 0.6.0
- google.golang.org/grpc -> v1.29.0
- nodegroup interface GetOptions() trivial implementation done
### Others
- [Release matrix](../README.md#releases-gardenerautoscaler) of Gardener Autoscaler updated.