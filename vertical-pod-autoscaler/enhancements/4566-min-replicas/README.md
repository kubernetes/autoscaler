# KEP-4566: MinReplicas per VPA object

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
   - [Existing Behaviour](#existing-behaviour)
   - [Reuse Cluster Autoscaler Annotations](#reuse-cluster-autoscaler-annotations)
<!-- /toc -->

## Summary

The default behaviour of VPA Updater is to allow Pod eviction only if there are
at least 2 live replicas, in order to avoid temporary total unavailability of a
workload under VPA in Auto mode. This can be changed globally with
`--min-replicas` flag. However, such a change might be deemed dangerous in a
centrally-managed multi-tenant cluster.

This proposal addresses the problem by allowing to specify per VPA object a
minimum number of replicas required to start Pod eviction (thus active VPA
actuation). This allows to keep the current relatively safe default and override
it explicitly only where needed.

## Motivation

The motivation behind the change is due to VPA users needing the configuration
option in managed environments to allow for eviction of a single replica, see
[issue #3986](https://github.com/kubernetes/autoscaler/issues/3986) or
[issue #1828](https://github.com/kubernetes/autoscaler/issues/1828).

### Goals

- Main: allow workload owner to use VPA in Auto/Recreate mode on a single-replica
  workloads,
- Secondary: allow workload owner to specify a higher number of replicas required
  to be kept alive. Note that this can also be achieved with
  [PDB](https://kubernetes.io/docs/tasks/run-application/configure-pdb/).

### Non-Goals

- Create any advanced form of policy around VPA-introduced disruption.

## Proposal

The proposal is to add `minReplicas` field under VPA `spec.updatePolicy` and
alter VPA Updater behaviour so it respects the new field:
- if the field is set, use its value instead of the global `--min-replicas`
  flag, only for this particular VPA object,
- if the field is not set, use the global `--min-replicas` flag, keeping
  backward compatibility.

The field is ignored by other VPA components (Recommender, Admission
Controller).

Since the change is backward-compatible the suggestion is to simply extend `v1`
version of VPA API, avoiding the hassle of introducing a new API version.

## Design Details

Suggested implementation is present in [PR
4560](https://github.com/kubernetes/autoscaler/pull/4560).

### Test Plan

Add automated E2E tests that update VPA objects, altering the value of
`spec.updatePolicy.minReplicas` flag and verifying that the behaviour of VPA
Updater changes accordingly.

## Implementation History

- 2021-12-28: initial version

## Alternatives

### Existing Behaviour

The only existing alternative to achieve the behaviour allowed by the proposed
change is to set global `--min-replicas` flag to `1` and use Pod Disruption
Budget to protect single-replica workloads from being disrupted, if needed.

While this allows for the cluster to end up with a very similar (if not identical)
behaviour, it changes the default behaviour to an unsafe one.
As a result this alternative requires a cluster-wide coordination - PDBs might
not be configured for all workloads needing protection.
This might be hard to achieve in multitenant environments.

Also the flag might not be accessible in managed environments (for example GKE).

### Reuse Cluster Autoscaler Annotations

An alternative suggested in
[issue #3986](https://github.com/kubernetes/autoscaler/issues/3986) was to reuse
Cluster Autoscaler annotation `cluster-autoscaler.kubernetes.io/safe-to-evict`.

While it's not clear if there are use cases when one would like to allow VPA or
CA to evict Pods but disallow it for the other controller, coupling them this
way seems dangerous.

Of course VPA could introduce its own annotation to express the same but it
seems much more elegant to keep the information in the API object (a solution
not possible for CA due to lack of such an object) instead of effectively
spreading the API between VPA objects and Pod annotations.

