**WARNING**
:warning: We no longer intend to implemnt this KEP. Instead we recommend using
[Unhealthy Pod Eviction Policy for PDBs](https://github.com/kubernetes/enhancements/blob/master/keps/sig-apps/3017-pod-healthy-policy-for-pdb/README.md).

# KEP-4902: Delete OOM Pods

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
   - [Goals](#goals)
   - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
   - [Test Plan](#test-plan)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
   - [Update the eviction API](#update-the-eviction-api)
<!-- /toc -->

## Summary

The default behaviour of VPA Updater is to evict Pods when new resource
recommendations are available. This works fine for most cases.
A problem that can arise is when there are multiple replicas and a
PodDisruptionBudget (PDB) which allows one disruption. Now the eviction is at
risk, because as soon as multiple replicas run into for example the memory
limit and get killed they will never recover as the eviciton API will not allow
any further disruptions.

This proposal addresses the problem by allowing users to enable the deletion of
OOMing pods as a backup if the eviction fails.

## Motivation

The motivation behind the change is to give VPA users a way to recover a
failing deployment, if updated/increased limits would solve the problem.

### Goals

- Main: allow cluster administrators and other users to enable deletion of pods

### Non-Goals

- Get rid of or work around the existing eviction behaviour

## Proposal

The proposal is to add a new field to the VPA resource and a global flag.

A new global flag (`--delete-ooming-on-eviction-error`) shall be added to the
VPA updater to enable the new feature globally.

Additionally a new field in the VPA resource
(`Spec.UpdatePolicy.DeleteOomingOnEvictionError `) which takes precedence to
the global flag. When unset the value of the flag is taken. This allows cluster
administrators to enable the flag for all VPA resources and at the same time
disable it again for specific deployments, or only enable it for specific
deployments.

This should give users the most flexible way of configuring this feature to
fit their needs.

## Design Details

When the eviction fails the pod will not just get blindy deleted, but further
checks will occur. Which gives us the following checklist:
- [ ] Was at least one container in the Pod terminated due to being OOM
      (`OOMKilled`)?
- [ ] Is at least one container in the Pod currently waiting due to being in
      `CrashLoopBackOff`?

Additionally deletion should only occur when an OOM event was recorded and if
it is planned to actually increase the memory limit. Those information are
already present in the updater, they just need to be made available.

This should make sure to not accidentally disrupt deployments as they might
still heal to a point where eviction then might be possible.

Suggested implementation is present in [PR
4898](https://github.com/kubernetes/autoscaler/pull/4898).

### Test Plan

Add unit tests that cover the new code paths.


## Implementation History

- 2022-05-19: initial version

## Alternatives

### Update the eviction API

Instead of implementing this change on the client side, the VPA in this case,
it could be implemented on the API side. This would have the advantage that it
would work for all clients. On the other hand this would introduce breaking
behaviour and most likely would result in a new api version.

Another possibility would be to allow the eviction of disrupted workloads, but
this might be hard to decide for the API.
It would need to allow the eviction if the container is currently not running.
This could be introduced as an additional flag in the PDB, to allow this
behaviour.

Also according to some discussions the general stance seems to be:
If you don't like the drain/evict behaviour, just use delete.
(https://github.com/kubernetes/kubernetes/issues/72320,
https://github.com/kubernetes/kubernetes/pull/105296)
