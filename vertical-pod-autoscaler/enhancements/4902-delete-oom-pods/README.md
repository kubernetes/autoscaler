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
pods as a backup if the eviction fails.

## Motivation

The motivation behind the change is to give VPA users a way to recover a
failing deployment, if updated/increased limits would solve the problem.

### Goals

- Main: allow cluster administrators to enable deletion of pods
- Secondary: configurable deletion threshold to tune the deletion behaviour

### Non-Goals

- Get rid of or work around the existing eviction behaviour

## Proposal

The proposal is to add `--delete-on-eviction-error` to the VPA to enable
deletion of pods.

To add a bit of configuration an additional flag,
`--delete-on-eviction-error-threshold`, should be addedd. This value compared
to the amount of restarts a pod has gone through. The deletion wil only be
allowed if the amount of restarts exceeds the threshold. This is to further
ensure that only pods get deleted that are consistently crashing.

## Design Details

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

Also according to some discussions the general stance seems to be:
If you don't like the drain/evict behaviour, just use delete.
(https://github.com/kubernetes/kubernetes/issues/72320,
https://github.com/kubernetes/kubernetes/pull/105296)
