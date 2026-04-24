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

# AEP-NNNN: Your short, descriptive title

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

## Motivation

<!--
Explain why this change is worth doing. Link to issues, user reports, or
previous discussions that show the problem is real. Reviewers will weigh the
cost of the change against the motivation described here, so be concrete.
-->

### Goals

<!--
Bullet list of what this AEP is trying to achieve. Keep each goal testable —
something a reviewer could point at later to decide whether the AEP succeeded.
-->

### Non-Goals

<!--
What is explicitly out of scope. Listing non-goals is often more valuable than
listing goals — it keeps the discussion focused and prevents scope creep during
review.
-->

## Proposal

<!--
Describe the proposed change at a high level. This is the "what", not the
"how" — implementation details belong in Design Details below. A reviewer
should be able to read this section and understand the user-visible behavior
without reading any code.
-->

## Design Details

<!--
The "how". Include enough detail that a reader can evaluate whether the
approach is sound. API types, flag names, component interactions, and any
non-obvious behavior belong here. Code snippets and YAML examples are welcome
when they clarify intent.
-->

### API Changes

<!--
If this AEP adds or modifies fields in the VPA API (`autoscaling.k8s.io/v1`),
describe the new types, their validation rules, and default behavior. Include
the Go struct definitions when possible. If there are no API changes, remove
this subsection.
-->

### Test Plan

<!--
Describe how this change will be tested. At minimum, reviewers expect:
- unit tests for new logic,
- e2e tests for user-visible behavior (what scenarios will be covered?).
Integration tests are not required for most AEPs, but mention them if they
apply.
-->

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

### Graduation Criteria

<!--
A few bullets describing what needs to be true to move the feature from alpha
to beta and from beta to GA. For most AEPs this is short — typical signals are
"tests are stable for N releases", "no open bugs against the feature gate",
and "positive user feedback". Remove this subsection if the change does not go
through a graduation lifecycle (e.g. a pure bug fix).
-->

### Version Skew

<!--
The VPA ships multiple components (recommender, updater, admission-controller).
Describe what happens when they are not all running the same version during a
rollout — for example, a new recommender writing a field that an older updater
does not understand. If the feature gate fully mitigates skew (the gate must
be enabled on all components before the behavior takes effect), state that.
Remove this subsection if only one component is affected.
-->

### Kubernetes Version Compatibility

<!--
Call out any minimum Kubernetes version this feature requires, and what the
VPA does when running on an older version. Fill this in when the AEP depends
on an upstream Kubernetes feature (e.g. KEP-1287 for in-place updates).
Otherwise, remove this subsection.
-->

## Implementation History

<!--
Track major milestones using absolute dates (YYYY-MM-DD):
- initial version
- significant design changes
- the first VPA release where the feature shipped
- graduation to beta / GA
-->

- YYYY-MM-DD: initial version

## Alternatives

<!--
What other approaches were considered, and why were they rejected? Even a
short note here helps future readers understand the design space and prevents
the same alternatives from being re-proposed.
-->
