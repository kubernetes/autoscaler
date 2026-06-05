<!--
**Note:** When your AEP is complete, all of these comment blocks should be
removed.

AEPs (Autoscaler Enhancement Proposals) are a lightweight version of Kubernetes
KEPs, scoped to the Vertical Pod Autoscaler subproject. Use this template as
the skeleton for new AEPs so reviewers see a consistent structure.

To get started:

- [ ] Open a tracking issue in kubernetes/autoscaler describing the problem.
- [ ] Copy this directory to `NNNN-short-descriptive-title`, where `NNNN` is
      the issue number.
- [ ] Fill in `Summary` and `Motivation` first — these are enough to start a
      design discussion.
- [ ] Open a PR for the new AEP and iterate. Merging an AEP does not mean it
      is approved or complete; aim for tightly-scoped PRs per topic.
- [ ] Fill in the remaining sections as the design firms up.

One AEP corresponds to one "feature" or "enhancement" for its whole lifecycle.
If major changes emerge after implementation, edit the AEP rather than opening
a new one.
-->

# AEP-9726: Capacity-Aware In-Place Updates in VPA Updater

<!--
Keep the title short and descriptive. It is used in the TOC, commit messages,
and PR titles.
-->

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [API Changes](#api-changes)
  - [Test Plan](#test-plan)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [Graduation Criteria](#graduation-criteria)
  - [Version Skew](#version-skew)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
- [Implementation History](#implementation-history)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary

<!--
A paragraph or two that captures what this AEP is about and why it matters.
Write this section so that someone unfamiliar with the VPA internals can read
it and understand the shape of the proposal.
-->
The InPlace update mode is intended for workloads that cannot tolerate the disruption caused by pod recreation.
Currently, the updater is unaware of the capacity constraints of the node on which a pod is running. As a result, it may attempt an in-place resize without first verifying whether the node has sufficient available resources.

This can lead to infeasible resize attempts that could otherwise be avoided, wasting CPU cycles in the updater and, more importantly, generating unnecessary load on the API server.

This AEP proposes that the updater skip in-place resize attempts for pods using the VPA InPlace policy when the node does not have sufficient allocatable capacity to accommodate the recommendation.

## Motivation

<!--
Explain why this change is worth doing. Link to issues, user reports, or
previous discussions that show the problem is real. Reviewers will weigh the
cost of the change against the motivation described here, so be concrete.
-->
The motivation for this feature is to reduce API server load caused by infeasible in-place resize attempts.
Given the large number of pod-node combinations, the existing cache is a useful optimization; however, further improvement is achievable by inspecting node allocatable resources and determining whether an in-place resize attempt should be initiated at all.

### Goals

<!--
Bullet list of what this AEP is trying to achieve. Keep each goal testable —
something a reviewer could point at later to decide whether the AEP succeeded.
-->
This proposal is a pure optimization. Its goals are:
1. Reduce CPU cycles spent by the updater on infeasible resize attempts.
2. Reduce API server load incurred by admission checks for infeasible in-place resize requests.

### Non-Goals

<!--
What is explicitly out of scope. Listing non-goals is often more valuable than
listing goals — it keeps the discussion focused and prevents scope creep during
review.
-->
This AEP does not introduce a new recommendation system.

## Proposal

<!--
Describe the proposed change at a high level. This is the "what", not the
"how" — implementation details belong in Design Details below. A reviewer
should be able to read this section and understand the user-visible behavior
without reading any code.
-->
The updater will be enhanced with a Node lister, enabling it to list and retrieve node objects.
When processing a recommendation for a pod that uses the InPlace update mode, the updater will verify whether the node has sufficient allocatable resources to accommodate the requested resource changes.
If the node cannot accommodate the recommendation, the resize attempt will not be initiated and the `CanInPlace` function will return an `InPlaceInfeasible` decision.

## Design Details

<!--
The "how". Include enough detail that a reader can evaluate whether the
approach is sound. API types, flag names, component interactions, and any
non-obvious behavior belong here. Code snippets and YAML examples are welcome
when they clarify intent.
-->
The `CanInPlace` function signature is extended to accept a node parameter:
```go
CanInPlaceUpdate(pod *corev1.Pod, node *corev1.Node, vpa *vpa_types.VerticalPodAutoscaler, infeasibleAttempts map[k8stypes.UID]*vpa_types.RecommendedPodResources) utils.InPlaceDecision
```
The function then evaluates whether the update mode is InPlace, the feature gate is enabled, and the node has sufficient allocatable capacity for the recommendation:
```go
if updateMode == vpa_types.UpdateModeInPlace && node != nil && features.Enabled(features.InPlaceCapacityAware) && !checkAllocatableNodeForInPlace(pod, recommendation, node.Status.Allocatable) {
	return utils.InPlaceInfeasible
}
```

### API Changes

<!--
If this AEP adds or modifies fields in the VPA API (`autoscaling.k8s.io/v1`),
describe the new types, their validation rules, and default behavior. Include
the Go struct definitions when possible. If there are no API changes, remove
this subsection.
-->
No API changes are introduced. This is a backend-only optimization.

### Test Plan

<!--
Describe how this change will be tested. At minimum, reviewers expect:
- unit tests for new logic,
- e2e tests for user-visible behavior (what scenarios will be covered?).
Integration tests are not required for most AEPs, but mention them if they
apply.
-->
New test cases will be added to the `TestCanInPlaceUpdate` function (https://github.com/kubernetes/autoscaler/blob/b7645aed576a7a1f5a30d91a70c0ba85717c2f2a/vertical-pod-autoscaler/pkg/updater/restriction/pods_inplace_restriction_test.go#L56):
1. `CanInPlaceUpdate` with InPlace mode — returns `InPlaceInfeasible` when node allocatable capacity is insufficient for the recommendation.
2. `CanInPlaceUpdate` with InPlace mode — returns `InPlaceApproved` when the pod already consumes most of the node's resources and the incremental delta fits within remaining capacity.

### Feature Enablement and Rollback

<!--
Answer the following if this AEP is gated by a feature flag:

- Feature gate name:
- Components depending on the feature gate (e.g. updater, admission-controller,
  recommender):
- What happens when the gate is enabled?
- What happens when the gate is disabled after being enabled? In particular,
  what happens to VPA objects already configured with the new field?

If the change is not gated (for example, a backward-compatible API extension
with safe defaults), state that explicitly and explain why no gate is needed.
-->
- Feature gate name: `InPlaceCapacityAware`
- Components depending on the feature gate: updater, admission-controller
- What happens when the gate is enabled: The updater verifies that the node has sufficient allocatable capacity for the recommendation before initiating an in-place resize attempt for pods using the InPlace update mode.
- What happens when the gate is disabled after being enabled: The updater does not verify node capacity before resize attempts, which may result in a higher rate of infeasible resize attempts.


### Graduation Criteria

<!--
A few bullets describing what needs to be true to move the feature from alpha
to beta and from beta to GA. For most AEPs this is short — typical signals are
"tests are stable for N releases", "no open bugs against the feature gate",
and "positive user feedback". Remove this subsection if the change does not go
through a graduation lifecycle (e.g. a pure bug fix).
-->
- Tests are stable for 3 releases.
- No open bugs against the feature gate.
- Positive user feedback.

### Version Skew

<!--
The VPA ships multiple components (recommender, updater, admission-controller).
Describe what happens when they are not all running the same version during a
rollout — for example, a new recommender writing a field that an older updater
does not understand. If the feature gate fully mitigates skew (the gate must
be enabled on all components before the behavior takes effect), state that.
Remove this subsection if only one component is affected.
-->
This change introduces no API modifications; the feature gate therefore fully mitigates any version skew.

### Kubernetes Version Compatibility

<!--
Call out any minimum Kubernetes version this feature requires, and what the
VPA does when running on an older version. Fill this in when the AEP depends
on an upstream Kubernetes feature (e.g. KEP-1287 for in-place updates).
Otherwise, remove this subsection.
-->
Kubernetes 1.27 or later is required, with the `InPlacePodVerticalScaling` feature gate enabled.



## Implementation History

<!--
Track major milestones using absolute dates (YYYY-MM-DD):
- initial version
- significant design changes
- the first VPA release where the feature shipped
- graduation to beta / GA
-->

- 2026-06-20: initial version

## Alternatives

<!--
What other approaches were considered, and why were they rejected? Even a
short note here helps future readers understand the design space and prevents
the same alternatives from being re-proposed.
-->
