# AEP-9936: Per-VPA Observation Window

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [Workflow](#workflow)
  - [API Changes](#api-changes)
  - [Gate Evaluation](#gate-evaluation)
  - [Interaction with <code>updateMode</code>](#interaction-with-updatemode)
  - [Validation](#validation)
  - [Status Condition](#status-condition)
  - [Metric](#metric)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
- [Test Plan](#test-plan)
- [Examples](#examples)
  - [Basic: <code>Recreate</code> with a six-hour window](#basic-recreate-with-a-six-hour-window)
  - [Fast iteration: <code>InPlace</code> with a short window](#fast-iteration-inplace-with-a-short-window)
  - [Pod-creation gating: <code>Initial</code> with a one-hour window](#pod-creation-gating-initial-with-a-one-hour-window)
- [Alternatives Considered](#alternatives-considered)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Add an optional `observationPeriodSeconds` field to `PodUpdatePolicy` that tells the Updater to refrain from applying recommendations for a configurable duration after a VPA's creation. During the window the Recommender continues to compute and publish recommendations to `status.recommendation` normally — only the Updater is gated. Once the window elapses, the configured `updateMode` takes effect automatically with no operator action.

The window is computed as a pure function of `vpa.CreationTimestamp` and the current spec value, evaluated on every reconcile. No status writes are required.

## Motivation

When new `Deployment` objects are provisioned automatically — by CI, per-tenant provisioning, on-demand application creation, etc. — a VPA created alongside them and configured with a non-`Off` `updateMode` will start actuating on recommendations that have not stabilised yet. Recommendations are especially unstable in the first few hours of a workload's life because the recommender has few usage samples, and applying those recommendations causes disruptive churn: pods are restarted with the wrong resource requests, then restarted again a few minutes later with different requests.

The safe pattern is to run the VPA in `Off` mode for the first N hours to let recommendations stabilise, then flip to `Recreate` / `InPlaceOrRecreate` / `InPlace` / `Initial`. Doing that manually for every new VPA does not scale: cluster operators end up building external controllers, CronJobs, or admission-side patch scripts to sequence the mode transition. Every environment reinvents the same workflow.

This feature moves the sequencing into the VPA API itself. Operators declare "wait N seconds after creation before touching pods" alongside the update mode, and the Updater honours it.

### Goals

- Provide a declarative per-VPA field that delays Updater actuation for a configurable duration after VPA creation.
- Keep the Recommender's behaviour unchanged: recommendations are still computed and published to `status.recommendation` during the window, so operators can inspect them.
- Work uniformly across `Recreate`, `InPlaceOrRecreate`, `InPlace`, and `Initial`.
- Require no status writes and no admission-side PATCH restrictions in v1: the gate is a pure function of `CreationTimestamp` and the current spec value.
- Surface the gate state through both a status condition (`ObservationPeriodActive`) and a Prometheus gauge (`vpa_in_observation_window`), so operators and dashboards can see when a VPA is gated without inspecting `spec`.

### Non-Goals

- **Recommender-confidence-based gating.** This feature is a user-declared policy, not a recommender-computed signal. Interoperation with `LowConfidence` / `RecommendationProvided` is left for future work.
- **Reset semantics on PATCH of other fields.** Changing `resourcePolicy`, `targetRef`, or the target workload does not reset the window. The window is anchored on `vpa.CreationTimestamp`; users who need to re-observe a workload after a material change should delete and recreate the VPA.
- **Cluster-wide default windows.** Different applications and different VPAs on the same target want different windows; this is a per-VPA concern.

## Proposal

Add a single optional field to `PodUpdatePolicy` (autoscaling.k8s.io/v1):

```go
// ObservationPeriodSeconds is the duration in seconds after this VPA's
// creation during which the Updater will refrain from applying
// recommendations, regardless of the configured UpdateMode.
// +optional
// +kubebuilder:validation:Minimum=1
ObservationPeriodSeconds *int32 `json:"observationPeriodSeconds,omitempty"`
```

Behaviour, in one sentence: **the Updater treats the VPA as if `updateMode` were `Off` until `now >= vpa.CreationTimestamp + spec.updatePolicy.observationPeriodSeconds`.** After that, the configured `updateMode` takes effect.

## Design Details

### Workflow

1. The user creates a VPA with a non-`Off` `updateMode` and `observationPeriodSeconds: N` set on `spec.updatePolicy`.
2. The Recommender begins computing recommendations on the normal schedule and populates `status.recommendation`. It is unaware of and unaffected by the observation window.
3. On every Updater reconcile, before deciding whether the VPA is eligible for actuation, the Updater evaluates the gate. If the gate is active, the VPA is treated as if `updateMode` were `Off` for that reconcile: no pods are evicted (`Recreate` / `InPlaceOrRecreate`), no in-place resize is attempted (`InPlace` / `InPlaceOrRecreate`), and no admission-controller-side injection happens on new-pod creation (`Initial`).
4. The Updater sets the `ObservationPeriodActive` status condition to `True` while the gate is active and `False` once it has elapsed. It emits the `vpa_in_observation_window` gauge accordingly.
5. On the first reconcile after `CreationTimestamp + observationPeriodSeconds`, the gate opens and the configured `updateMode` takes effect on subsequent reconciles.

### API Changes

Extended `PodUpdatePolicy`:

```go
type PodUpdatePolicy struct {
    // ... existing fields (UpdateMode, MinReplicas, EvictionRequirements,
    //                     EvictAfterOOMSeconds) ...

    // ObservationPeriodSeconds is the duration in seconds after this
    // VPA's creation during which the Updater will refrain from applying
    // recommendations, regardless of the configured UpdateMode.
    //
    // During the window, the Recommender continues to compute and
    // publish recommendations to status.recommendation as usual, but the
    // Updater treats the VPA as if UpdateMode were Off. After the window
    // elapses (now >= CreationTimestamp + ObservationPeriodSeconds), the
    // configured UpdateMode takes effect on subsequent reconciles.
    //
    // The window is computed as a pure function of vpa.CreationTimestamp
    // and this spec field. PATCH is allowed; the new value takes effect
    // on the next reconcile. Removing the field or setting it to zero
    // closes the gate immediately. Extending the value after the window
    // has already expired will re-arm the gate.
    //
    // If UpdateMode is Off, this field has no effect: Off already
    // suppresses Updater actuation indefinitely.
    //
    // Must be >= 1 if set.
    // +optional
    // +kubebuilder:validation:Minimum=1
    ObservationPeriodSeconds *int32 `json:"observationPeriodSeconds,omitempty"`
}
```

Follows the shape of `EvictAfterOOMSeconds` (the only existing time-based field on the same struct) and `DurationSeconds` from AEP-7862.

### Gate Evaluation

Reference implementation, invoked at the beginning of the Updater's per-VPA eligibility path:

```go
// inObservationWindow returns true if the VPA is currently within its
// declared observation window and the Updater should refrain from
// applying recommendations.
func inObservationWindow(vpa *vpa_types.VerticalPodAutoscaler, now time.Time) bool {
    p := vpa.Spec.UpdatePolicy
    if p == nil || p.ObservationPeriodSeconds == nil || *p.ObservationPeriodSeconds < 1 {
        return false
    }
    expiry := vpa.CreationTimestamp.Add(
        time.Duration(*p.ObservationPeriodSeconds) * time.Second,
    )
    return now.Before(expiry)
}
```

Insertion point: at the top of the Updater's eligibility check in `pkg/updater/logic/updater.go`, at the same call site where `UpdateModeOff` and `UpdateModeInitial` currently short-circuit the eviction path. When `inObservationWindow` returns `true`, the Updater takes the same code path it takes for `UpdateModeOff` today.

The gate is stateless: it is a pure function of `CreationTimestamp` (immutable on the object) and the current spec value (mutable). No caching, no status writes.

Consequences that fall out of this design:

| PATCH operation | Effect |
| --- | --- |
| Shorten `observationPeriodSeconds` during window | Gate releases earlier on next reconcile. |
| Extend `observationPeriodSeconds` during window | Gate stays open longer. |
| Extend `observationPeriodSeconds` **after** expiry | Gate re-arms on next reconcile. |
| Remove `observationPeriodSeconds` (or set to zero) | Gate closes immediately on next reconcile. |
| PATCH `updateMode` | New mode takes effect when the window elapses (or immediately, if already elapsed). |
| PATCH `resourcePolicy` / `targetRef` | No effect on the gate. |

The re-arm-after-expiry behaviour is intentional: it lets users manually extend an observation window in response to observed instability without deleting the VPA. If reviewers consider this a footgun serious enough to justify admission logic, we can add a check rejecting PATCHes that would push `expiry` past `now` once the window has already elapsed — see [Alternatives Considered](#alternatives-considered).

### Interaction with `updateMode`

The observation window is orthogonal to `updateMode`. While the gate is active, the Updater behaves as if `updateMode` were `Off` regardless of the configured value. Once the gate opens, the configured mode takes effect normally:

- `Recreate` → the Updater is now eligible to evict pods to apply recommendations.
- `InPlaceOrRecreate` → the Updater attempts in-place resize first, falling back to eviction.
- `InPlace` → the Updater attempts in-place resize only; no evictions.
- `Initial` → the admission-controller-side injection is now eligible to run on subsequent new-pod creations. Pods created during the window (or pods that pre-dated the VPA) are unaffected, matching the standard `Initial` semantics.
- `Off` → `observationPeriodSeconds` has no effect. `Off` already suppresses actuation indefinitely, so combining the two is a valid (but redundant) configuration that silently no-ops. No error, no warning; the field is simply ignored.

### Validation

Enforced via CRD schema:

- `observationPeriodSeconds` must be an integer `>= 1` if set.

No dynamic (admission-webhook) validation is required in v1. All lifecycle transitions are handled declaratively by the gate.

### Status Condition

Add a new value to the existing `VerticalPodAutoscalerStatus.Conditions` slice:

- **Type:** `ObservationPeriodActive`
- **Status:** `True` while the gate is active, `False` once it has elapsed.
- **Reason:** `WindowActive` (True) / `WindowExpired` (False).
- **Message:** human-readable summary including expiry timestamp.

This lets `kubectl describe vpa` surface the gate without an operator having to compute `CreationTimestamp + observationPeriodSeconds` mentally.

### Metric

Emit a new gauge from the Updater:

```
vpa_in_observation_window{namespace, name} = 0 | 1
```

Value is `1` while the gate is active for that VPA, `0` otherwise. Enables fleet-wide visibility into how many VPAs are currently gated, useful for dashboards and alerting on stuck windows.

### Feature Enablement and Rollback

Feature gate: **`VPAObservationWindow`**.

Components consuming the gate:

- `updater` — reads `observationPeriodSeconds` and honours the gate.
- `admission-controller` — gates acceptance of the field on new/updated VPAs behind the feature gate.

Enabling the feature gate causes:

- `admission-controller` to accept `observationPeriodSeconds` on new/updated VPAs.
- `updater` to honour the gate.

Disabling the feature gate causes:

- `admission-controller` to reject new VPAs that set `observationPeriodSeconds`, with a descriptive error message indicating the feature gate is disabled.
- `updater` to ignore the field on existing objects — the VPA behaves as if the field were not set (fail-open). This is deliberate: disabling the gate should never *increase* actuation restrictions on existing objects.

### Kubernetes Version Compatibility

This feature is entirely internal to the VPA controllers and depends on no new Kubernetes APIs. It is compatible with any Kubernetes version supported by the corresponding VPA release.

## Test Plan

**Unit tests** (`pkg/updater/logic/`):

- `inObservationWindow` returns `true` when `now < CreationTimestamp + observationPeriodSeconds`.
- `inObservationWindow` returns `false` when `now >= CreationTimestamp + observationPeriodSeconds`.
- `inObservationWindow` returns `false` when `observationPeriodSeconds` is `nil`, `0`, or absent.
- Gate short-circuits the eligibility path identically to `updateMode: Off` for `Recreate`, `InPlaceOrRecreate`, `InPlace`, `Initial`.
- Gate has no effect when `updateMode: Off` — behavior matches plain `updateMode: Off` regardless of the field's value.
- Metric and condition are set correctly on both sides of the transition.
- PATCH scenarios: shortening releases the gate, extending stays gated, removing closes the gate, extending after expiry re-arms.

**Admission-controller / CRD validation tests:**

- Accept `observationPeriodSeconds >= 1` on all update modes, including `Off` (no-op semantics).
- Reject `observationPeriodSeconds < 1` via CRD schema validation.
- Reject `observationPeriodSeconds` on any VPA when the `VPAObservationWindow` feature gate is disabled.

**E2E tests** (`e2e/`):

- Create a VPA with `updateMode: Recreate` and `observationPeriodSeconds: 60`. Verify no eviction occurs during the first 60 s. Advance time past the window; verify the next reconcile evicts as expected.
- Repeat for `InPlaceOrRecreate`, `InPlace`, and `Initial` (with pod-creation-timing tests for the last).
- Create a VPA with `observationPeriodSeconds: 3600`, PATCH to `60` after 30 s, verify eviction becomes eligible around the 60 s mark.
- Create a VPA with `observationPeriodSeconds: 60`, wait for expiry, PATCH to `3600`, verify the gate re-arms and eviction is again suppressed until the new expiry.
- Verify the `ObservationPeriodActive` condition and `vpa_in_observation_window` gauge match the gate state throughout each scenario.

## Examples

### Basic: `Recreate` with a six-hour window

Typical "let this new service settle for a work-day" configuration.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: web-vpa
  namespace: production
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: web
  updatePolicy:
    updateMode: Recreate
    observationPeriodSeconds: 21600   # 6 hours
```

For the first six hours after creation, `web` pods are not evicted regardless of recommendations. After the window, `Recreate` behaviour takes effect.

### Fast iteration: `InPlace` with a short window

CI-provisioned staging environments where operators want quick feedback but still want to avoid churn on the very first samples.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: staging-svc-vpa
  namespace: ci-runs
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: staging-svc
  updatePolicy:
    updateMode: InPlace
    observationPeriodSeconds: 300     # 5 minutes
```

### Pod-creation gating: `Initial` with a one-hour window

Services created by a per-tenant provisioner: the operator wants VPA-injected resources at pod-creation time, but only after an hour of stable recommendations.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: tenant-worker-vpa
  namespace: tenant-42
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: worker
  updatePolicy:
    updateMode: Initial
    observationPeriodSeconds: 3600    # 1 hour
```

Pods created during the first hour receive their `Deployment`-spec resources unchanged. Pods created after the window get VPA-injected resources.

## Alternatives Considered

**1. Extend `EvictionRequirements` with a `TimeSinceCreationExceeds` value.** Reuses an existing extensibility hook (AEP-4831), but forces a duration to be encoded inside an enum meant for direction-of-change comparisons (`TargetHigherThanRequests` / `TargetLowerThanRequests`). Awkward and probably a design regression on that API.

**2. Recommender-side confidence gate.** Block `RecommendationProvided` from becoming `True` for N hours or until a confidence threshold is met. Cleaner conceptually but hides visibility — operators would see no recommendation at all in `status.recommendation` during the window, defeating the "collect first, inspect, then apply" workflow this feature is meant to support.

**3. Global Updater flag** (e.g. `--default-observation-window-seconds`). Wrong axis: different applications in the same cluster want different windows, and different VPAs on the same target want different windows.

**4. Do nothing; rely on operator tooling.** The status quo. Ships the complexity downstream to every VPA operator; the CronJob-plus-patch-controller dance ends up reinvented per environment.

**5. Immutable field after creation.** Reject PATCH that modifies `observationPeriodSeconds` via admission. Simpler in that it eliminates the re-arm and shortening cases, but forces users into delete-and-recreate to change the window — an expensive operation that also discards the VPA's recommendation history. The proposed stateless-gate design gets the same "one-line mental model" without that cost.

**6. Reset on "material" spec changes** (`resourcePolicy`, `targetRef`). Arguably more correct — a workload with a different target really does have different history semantics — but forces us to define "material" and requires a status field to track the current anchor. Deferred: this AEP does not preclude adding such semantics in a future revision, and the `CreationTimestamp` anchor is a strict subset of any future anchor policy.

**7. Reject PATCH that re-arms an expired gate.** Split the difference on (5) and the proposed design: allow PATCH freely during the window, but reject PATCHes that would push `expiry` past `now` once the window has already elapsed. Rejected because it makes the admission logic dependent on real time (surprising) and does not eliminate the underlying "PATCHes can change gate state" surprise — the `ObservationPeriodActive` condition and gauge already surface any re-arm immediately.

## Implementation History

- (issue filed) 2026-07-06 — Issue [kubernetes/autoscaler#9936](https://github.com/kubernetes/autoscaler/issues/9936) filed.
- (triage accepted) 2026-07-07 — `/triage accepted` from SIG member.
- (AEP PR opened) TBD.
- (initial implementation) TBD.
- (alpha) TBD.
- (beta) TBD.
