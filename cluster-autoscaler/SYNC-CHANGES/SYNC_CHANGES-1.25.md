<!--- For help refer to https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.20.md?plain=1 as example --->

- [v1.25.0](#v1250)
    - [Synced with which upstream CA](#synced-with-which-upstream-ca)
    - [Changes made](#changes-made)
        - [To FAQ](#to-faq)
        - [During merging](#during-merging)
        - [During vendoring k8s](#during-vendoring-k8s)
        - [Others](#others)


# v1.25.0


## Synced with which upstream CA

[v1.25.0](https://github.com/kubernetes/autoscaler/tree/cluster-autoscaler-1.25.0/cluster-autoscaler)

## Changes made

### To FAQ
- new FAQ added —> How can I see all the events from Cluster Autoscaler?
- changes in some answers
    - Dynamic provisioning
    - how does scale down work
- Updated list of autoscaler startup parameters.
    - ignore-daemonsets-utilization
    - ignore-mirror-pods-utilization
    - balancing-label
    - cordon-node-before-terminating
    - record-duplicated-events

### During merging

- In `builder/Dockerfile`
  - Dockerfile go version is updated to 1.18.1 from 1.19

- In `.ci`
  - Updated go version to 1.18.1 from 1.16, added permissions as well.

- In `cluster-autoscaler/core`
  - removed all the scale down logic in a new directory called scaledown. The entire code for scale down is refactored. See upstream release notes for more info.
	
- In `cluster-autoscaler/metrics`
  - Added new metric called `skippedScaleEventsCount` which keeps count of scaling events that the CA has chosen to skip because of CPU/Memory limits.

- In `charts/cluster-autoscaler`
    - `README.md` —> PSP keys updated.
    - `values.yaml` —> autoDiscovery enabled for azure, cloudConfigPath changed to “” for aws

- In `cluster-autoscaler/processors`
  - nodegroupset —> updated ignoredLabels for ClusterAPI

- Following new flags are introduced :-
    - `max-pod-eviction-time` —> Maximum time CA tries to evict a pod before giving up (Default 2min)
    - `balancing-label` —> Specifies a label to use for comparing if two node groups are similar, rather than the built in heuristics. Setting this flag disables all other comparison logic, and cannot be combined with --balancing-ignore-label
    - `initial-node-group-backoff-duration` —>  initialNodeGroupBackoffDuration is the duration of first backoff after a new node failed to start (Default 5*time.Minute)
    - `max-node-group-backoff-duration` —>  the maximum backoff duration for a NodeGroup after new nodes failed to start (Default 30*time.Minute)
    - `node-group-backoff-reset-timeout` —>  the time after last failed scale-up when the backoff duration is reset.  (Default 3*time.Hour)
    - `max-scale-down-parallelism` —>  Maximum number of nodes (both empty and needing drain) that can be deleted in parallel (Default 10)
    - `max-drain-parallelism` —> Maximum number of nodes needing drain, that can be drained and deleted in parallel. (Default 1)
    - `gce-expander-ephemeral-storage-support` —>  Whether scale-up takes ephemeral storage resources into account for GCE cloud provider (Default false)
    - `record-duplicated-events` —> enable duplication of similar events within a 5 minute window (Default false)
    - `max-nodes-per-scaleup` —> Max nodes added in a single scale-up. This is intended strictly for optimizing CA algorithm latency and not a tool to rate-limit scale-up throughput (Default 1000)
    - `max-nodegroup-binpacking-duration` —> Maximum time that will be spent in binpacking simulation for each NodeGroup (Default 10*time.Second)

- There are other code changes as well, refer to the release notes to find more about them.

### During vendoring k8s
- mcm v0.46.0 -> 0.47.0
- mcm-provider-aws v0.12.0 -> v0.15.0
- mcm-provider-azure v0.8.0 -> v0.9.0

### Others
- [Release matrix](../README.md#releases-gardenerautoscaler) of Gardener Autoscaler updated.
