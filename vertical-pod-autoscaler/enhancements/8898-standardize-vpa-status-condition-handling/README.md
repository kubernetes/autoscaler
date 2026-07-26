# AEP-8898: Standardize VPA status condition handling

<!-- toc -->
- [Summary](#summary)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Motivation](#motivation)
- [Proposal](#proposal)
  - [Improve existing conditions](#improve-existing-conditions)
  - [Add new conditions](#add-new-conditions)
    - [ScalingBlocked](#scalingblocked)
    - [ScalingRequired](#scalingrequired)
    - [ScalingActionSucceeded](#scalingactionsucceeded)
- [Design Details](#design-details)
  - [Test Plan](#test-plan)
  - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

The Condition field on resources in Kubernetes is a standard mechanism to provide reporting from a controller. The purpose of this AEP is to change the behaviour of the existing conditions, bringing them in line with modern practices, and in addition to that, add some additional conditions that will provide users with additional information that can be useful to monitor the behaviour of their VPAs.

### Goals

- Update existing conditions to conform to current [guidance](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties) from sig-architecture, by modifying their status, rather than deleting or adding conditions.
- Add new conditions that indicate useful status to users

### Non-Goals

- Removing unused conditions from the API

## Motivation

The current VPA implementation handles conditions inconsistently - some conditions are deleted when they become "false" rather than having their status updated. This behavior deviates from [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties), which recommend that conditions persist and toggle between `True` and `False` status values.

This inconsistency causes several problems:

1. **Monitoring and alerting**: Tools like Prometheus that watch for condition changes cannot reliably alert on VPA state because conditions appear and disappear rather than transitioning. Users cannot easily set up alerts like "alert if ConfigUnsupported has been True for more than 5 minutes" when the condition might not exist at all.

2. **E2E testing**: Tests that verify VPA behavior must resort to waiting for arbitrary time periods rather than checking for specific condition states. With proper condition semantics, tests can wait for a condition like `ScalingRequired=False` to confirm that scaling has completed, providing faster and more reliable test results.

3. **Observability**: Users and operators lack visibility into why a VPA is or isn't taking action. New conditions like `ScalingBlocked`, `ScalingRequired`, and `ScalingActionSucceeded` provide richer information about the VPA's decision-making process.

## Proposal

This proposal is essentially two parts: Improving existing conditions and adding new conditions.

### Improve existing conditions

Existing VPA conditions will be updated to persist with `status: False` instead of being deleted when their state becomes false. The following conditions are affected:

- `ConfigDeprecated`
- `ConfigUnsupported`
- `NoPodsMatched`

### Add new conditions

In addition to changing existing conditions, this AEP also proposes to add new conditions. The purpose of this AEP is to retrofit conditions that should have always been there. The list is a work in progress and will be amended as new retrofitted conditions are required.

Any new feature requiring new conditions will list those conditions in that feature's AEP.

#### ScalingBlocked

Type: ScalingBlocked

True status reasons:

- InsufficientReplicas
- ScalingDisabled

False status reasons:

- SufficientReplicas

#### ScalingRequired

Type: ScalingRequired

True status reasons:

- PodResourcesDiverged (pod resources differ from recommendation)

False status reasons:

- NoEligiblePodsForScaling (no pods require scaling - resources match recommendation or no pods can be safely updated)

#### ScalingActionSucceeded

Type: ScalingActionSucceeded

This condition is only set when a scaling action is actually attempted.

True status reasons:

- InPlaceResizeSuccessful
- EvictionSuccess

False status reasons:

- InPlaceResizeFailure
- EvictionFailed

## Design Details

### Component Responsibilities

The following table shows which VPA component is responsible for setting each condition:

| Condition | Component | Notes |
|-----------|-----------|-------|
| `ConfigDeprecated` | Recommender | Set when VPA uses deprecated configuration |
| `ConfigUnsupported` | Recommender | Set when VPA configuration is invalid |
| `NoPodsMatched` | Recommender | Set based on whether pods match the VPA selector |
| `RecommendationProvided` | Recommender | Set when a recommendation has been calculated |
| `ScalingBlocked` | Updater | Set when scaling cannot proceed (insufficient replicas, disabled mode) |
| `ScalingRequired` | Updater | Set based on whether pods need resource adjustments |
| `ScalingActionSucceeded` | Updater | Set after an actual scaling action (eviction or in-place resize) is attempted |

### Test Plan

**Unit Tests:**

- Test condition state transitions in the updater
- Test that conditions persist with `status: False` rather than being deleted
- Test each condition/reason combination is set correctly based on VPA state

**E2E Tests:**

- Verify `ScalingRequired=False` is set when no pods need scaling
- Verify `ScalingActionSucceeded=True` is set after successful eviction or in-place resize
- Verify `ScalingActionSucceeded=False` is set after failed scaling attempts
- Verify `ScalingBlocked=True` with appropriate reasons when scaling cannot proceed

**Note:** Several E2E tests already exist and can be updated to use the new `ScalingRequired` condition instead of waiting for arbitrary timeouts.

### Upgrade / Downgrade Strategy

#### Upgrade

- Existing VPAs will gain new conditions (`ScalingBlocked`, `ScalingRequired`, `ScalingActionSucceeded`) on first reconciliation after upgrade
- Existing conditions that were previously deleted when their state became "false" will now persist with `status: False`
- This is non-breaking - clients that don't understand new conditions will ignore them per standard Kubernetes behavior

#### Downgrade

- New conditions will remain on VPA objects but won't be updated by older controllers
- Older clients ignore unknown conditions (standard Kubernetes behavior)
- No manual cleanup required

### Feature Enablement and Rollback

#### How can this feature be enabled / disabled in a live cluster?

This feature is always enabled and does not require a feature gate. The changes consist of:

1. **Bug fixes** to existing condition handling (persisting conditions with `status: False` instead of deleting them) - this aligns VPA with Kubernetes API conventions and is always active.

2. **New conditions** (`ScalingBlocked`, `ScalingRequired`, `ScalingActionSucceeded`) - these are additive and do not affect existing functionality.

#### Rollback

To rollback, downgrade the VPA components (recommender, updater) to a previous version. After rollback:

- New conditions will remain on VPA objects but will no longer be updated
- Existing conditions will revert to the old behavior (being deleted instead of set to `False`)
- No manual cleanup is required

## Implementation History

- 2025-12-07: initial version
