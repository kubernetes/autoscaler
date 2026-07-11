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
  - [Interaction with CPU Startup Boost](#interaction-with-cpu-startup-boost)
  - [Validation](#validation)
  - [Status Condition](#status-condition)
  - [Metric](#metric)
  - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
  - [Graduation Criteria](#graduation-criteria)
  - [Version Skew](#version-skew)
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

Add an optional `initialDelaySeconds` field to `PodUpdatePolicy` that delays actuation of recommendations for a configurable duration after a VPA's creation. During the window the Recommender continues to compute and publish recommendations to `status.recommendation` normally, but nothing actuates them: the Updater performs no evictions or in-place resizes, and the Admission Controller does not apply recommendations to pods created during the window. The VPA behaves exactly as if `updateMode` were `Off`. Once the window elapses, the configured `updateMode` takes effect automatically with no operator action.

The window is computed as a pure function of `vpa.CreationTimestamp` and the current spec value, evaluated on every reconcile. No status writes are required. Modifying an existing VPA's spec does not reset the window — the anchor (`CreationTimestamp`) is immutable, and changing `initialDelaySeconds` itself simply moves the expiry relative to that anchor (see [Gate Evaluation](#gate-evaluation) for the full PATCH semantics).

## Motivation

When new `Deployment` objects are provisioned automatically — by CI, per-tenant provisioning, on-demand application creation, etc. — a VPA created alongside them and configured with a non-`Off` `updateMode` will start actuating on recommendations that have not stabilised yet. Recommendations are especially unstable in the first few hours of a workload's life because the recommender has few usage samples, and applying those recommendations causes disruptive churn: pods are restarted with the wrong resource requests, then restarted again a few minutes later with different requests.

The safe pattern is to run the VPA in `Off` mode for the first N hours to let recommendations stabilise, then flip to `Recreate` / `InPlaceOrRecreate` / `InPlace` / `Initial`. Doing that manually for every new VPA does not scale: cluster operators end up building external controllers, CronJobs, or admission-side patch scripts to sequence the mode transition. Every environment reinvents the same workflow.

This feature moves the sequencing into the VPA API itself. Operators declare "wait N seconds after creation before touching pods" alongside the update mode, and the Updater honours it.

### Goals

- Provide a declarative per-VPA field that delays all actuation — evictions, in-place resizes, and admission-time injection into newly created pods — for a configurable duration after VPA creation.
- Keep the Recommender's behaviour unchanged: recommendations are still computed and published to `status.recommendation` during the window, so operators can inspect them.
- Work uniformly across `Recreate`, `InPlaceOrRecreate`, `InPlace`, and `Initial`.
- Require no status writes and no PATCH-blocking admission validation in v1: the gate is a pure function of `CreationTimestamp` and the current spec value.
- Surface the gate state through both a status condition (`InitialDelayActive`) and a Prometheus gauge (`vpa_initial_delay_active`), so operators and dashboards can see when a VPA is gated without inspecting `spec`.

### Non-Goals

- **Recommender-confidence-based gating.** This feature is a user-declared policy, not a recommender-computed signal. Interoperation with `LowConfidence` / `RecommendationProvided` is left for future work.
- **Reset semantics on PATCH of other fields.** Changing `resourcePolicy`, `targetRef`, or the target workload does not reset the window. The window is anchored on `vpa.CreationTimestamp`; users who need to re-observe a workload after a material change should delete and recreate the VPA.
- **Cluster-wide default windows.** Different applications and different VPAs on the same target want different windows; this is a per-VPA concern.

## Proposal

Add a single optional field to `PodUpdatePolicy` (autoscaling.k8s.io/v1):

```go
// InitialDelaySeconds is the duration in seconds after this VPA's
// creation during which recommendations are not actuated: the Updater
// performs no evictions or in-place resizes, and the Admission
// Controller does not apply recommendations to newly created pods,
// regardless of the configured UpdateMode.
// +optional
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=7776000
InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`
```

Behaviour, in one sentence: **the Updater and the Admission Controller treat the VPA as if `updateMode` were `Off` until `now >= vpa.CreationTimestamp + spec.updatePolicy.initialDelaySeconds`.** After that, the configured `updateMode` takes effect.

Modifying the VPA's spec does not reset the window. The gate is a pure function of the immutable `vpa.CreationTimestamp` and the current value of `initialDelaySeconds`, re-evaluated on every reconcile — there is no per-object state to reset. The full PATCH semantics are enumerated in [Gate Evaluation](#gate-evaluation). While the gate is active, the [`InitialDelayActive` status condition](#status-condition) surfaces it on the VPA object.

## Design Details

### Workflow

1. The user creates a VPA with a non-`Off` `updateMode` and `initialDelaySeconds: N` set on `spec.updatePolicy`.
2. The Recommender begins computing recommendations on the normal schedule and populates `status.recommendation`. It is unaware of and unaffected by the observation window.
3. On every Updater reconcile, before deciding whether the VPA is eligible for actuation, the Updater evaluates the gate. If the gate is active, the VPA is treated as if `updateMode` were `Off` for that reconcile: no pods are evicted (`Recreate` / `InPlaceOrRecreate`) and no in-place resize is attempted (`InPlace` / `InPlaceOrRecreate`).
4. The Admission Controller evaluates the same gate whenever a pod matching the VPA's target is created. While the gate is active it does not patch the pod's resources — the pod is admitted with its original spec, exactly as under `updateMode: Off`. This applies to every mode, not just `Initial`: without it, pods created during the window (scale-ups, node replacements, crash restarts) would receive un-stabilised recommendations at admission time.
5. The Updater sets the `InitialDelayActive` status condition to `True` while the gate is active and `False` once it has elapsed. It emits the `vpa_initial_delay_active` gauge accordingly.
6. On the first reconcile after `CreationTimestamp + initialDelaySeconds`, the gate opens and the configured `updateMode` takes effect on subsequent reconciles and pod creations.

### API Changes

Extended `PodUpdatePolicy`:

```go
type PodUpdatePolicy struct {
    // ... existing fields (UpdateMode, MinReplicas, EvictionRequirements,
    //                     EvictAfterOOMSeconds) ...

    // InitialDelaySeconds is the duration in seconds after this
    // VPA's creation during which recommendations are not actuated,
    // regardless of the configured UpdateMode.
    //
    // During the window, the Recommender continues to compute and
    // publish recommendations to status.recommendation as usual, but the
    // Updater and the Admission Controller treat the VPA as if
    // UpdateMode were Off: no evictions, no in-place resizes, and no
    // recommendation injection into newly created pods. After the window
    // elapses (now >= CreationTimestamp + InitialDelaySeconds), the
    // configured UpdateMode takes effect on subsequent reconciles and
    // pod creations.
    //
    // The window is computed as a pure function of vpa.CreationTimestamp
    // and this spec field. PATCH is allowed; the new value takes effect
    // on the next reconcile. Removing the field or setting it to zero
    // closes the gate immediately. Setting a larger value after the
    // window has expired re-arms the gate only if the new expiry
    // (CreationTimestamp + InitialDelaySeconds) still lies in the
    // future; otherwise the gate stays open.
    //
    // If UpdateMode is Off, this field has no effect: Off already
    // suppresses Updater actuation indefinitely.
    //
    // Must be between 1 and 7776000 (90 days) if set.
    // +optional
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=7776000
    InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`
}
```

Follows the shape of `EvictAfterOOMSeconds` (the only existing time-based field on the same struct) and `DurationSeconds` from AEP-7862; the name matches `initialDelaySeconds` on Pod probes.

Gate state is surfaced via the [`InitialDelayActive` condition](#status-condition).

### Gate Evaluation

Reference implementation, exposed from a shared package (`pkg/utils/vpa`) and consumed by both actuating components:

```go
// inInitialDelayWindow returns true if the VPA is currently within its
// declared observation window and the Updater and Admission Controller
// should refrain from actuating recommendations.
func inInitialDelayWindow(vpa *vpa_types.VerticalPodAutoscaler, now time.Time) bool {
    p := vpa.Spec.UpdatePolicy
    if p == nil || p.InitialDelaySeconds == nil || *p.InitialDelaySeconds < 1 {
        return false
    }
    expiry := vpa.CreationTimestamp.Add(
        time.Duration(*p.InitialDelaySeconds) * time.Second,
    )
    return now.Before(expiry)
}
```

Insertion points:

- **Updater** — at the top of the eligibility check in `pkg/updater/logic/updater.go`, at the same call site where `UpdateModeOff` and `UpdateModeInitial` currently short-circuit the eviction path.
- **Admission Controller** — at the existing `UpdateModeOff` short-circuits in `pkg/admission-controller/resource/vpa/matcher.go` and `pkg/admission-controller/resource/pod/recommendation/recommendation_provider.go`, so pods created during the window are admitted with their original spec resources.

When `inInitialDelayWindow` returns `true`, both components take the same code path they take for `UpdateModeOff` today.

The gate is stateless: it is a pure function of `CreationTimestamp` (immutable on the object) and the current spec value (mutable). No caching, no status writes.

Consequences that fall out of this design:

| PATCH operation | Effect |
| --- | --- |
| Shorten `initialDelaySeconds` during window | Gate releases earlier on next reconcile. |
| Extend `initialDelaySeconds` during window | Gate stays open longer. |
| Extend `initialDelaySeconds` **after** expiry | Gate re-arms on the next reconcile **only if** the new expiry (`CreationTimestamp + newValue`) still lies in the future. A value whose expiry is already in the past leaves the gate open. |
| Remove `initialDelaySeconds` (or set to zero) | Gate closes immediately on next reconcile. |
| Modify `updateMode` | New mode takes effect when the window elapses (or immediately, if already elapsed). |
| Modify `resourcePolicy` / `targetRef` | No effect on the gate. |

Re-arm after expiry is possible but not guaranteed: because the window is anchored to `CreationTimestamp`, a PATCH re-arms the gate only when the new expiry still lies in the future. This is intentional — it lets users extend an observation window in response to observed instability without deleting the VPA, while a larger-but-already-elapsed value on an old VPA simply has no effect. If reviewers consider this a footgun serious enough to justify admission logic, we can add a check rejecting PATCHes that would push `expiry` past `now` once the window has already elapsed — see [Alternatives Considered](#alternatives-considered).

### Interaction with `updateMode`

The observation window is orthogonal to `updateMode`. While the gate is active, the Updater and the Admission Controller behave as if `updateMode` were `Off` regardless of the configured value: no evictions, no in-place resizes, and no recommendation injection into pods created during the window. Once the gate opens, the configured mode takes effect normally:

- `Recreate` → the Updater is now eligible to evict pods to apply recommendations.
- `InPlaceOrRecreate` → the Updater attempts in-place resize first, falling back to eviction.
- `InPlace` → the Updater attempts in-place resize only; no evictions.
- `Initial` → the admission-controller-side injection is now eligible to run on subsequent new-pod creations. Pods created during the window (or pods that pre-dated the VPA) are unaffected, matching the standard `Initial` semantics.
- `Off` → `initialDelaySeconds` has no effect; `Off` already suppresses actuation indefinitely.

### Interaction with CPU Startup Boost

CPU Startup Boost (AEP-7862) still applies during the delay window: the boost is a startup-safety mechanism computed from the pod's own spec, not a recommendation, and the Admission Controller already applies it even under `updateMode: Off` (the `Off` short-circuit in `resource/vpa/matcher.go` is bypassed for VPAs with a boost). The gate reuses the same `Off` semantics, so boosted containers remain boosted.

The interaction appears at boost expiry, when the Updater performs the post-boost in-place downscale:

- Delay window still active → the pod is scaled down to its original spec resources; no recommendation is actuated.
- Delay window elapsed → the pod is scaled down to the VPA recommendation, per normal post-boost behaviour.

### Validation

Enforced via CRD schema:

- `initialDelaySeconds` must be an integer between `1` and `7776000` (90 days) if set. The upper bound follows the [API conventions for numeric fields](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#numeric-fields); observation windows are realistically hours to days, so 90 days is generous headroom while still rejecting absurd values.

No dynamic (admission-webhook) validation is required in v1. All lifecycle transitions are handled declaratively by the gate.

### Status Condition

Add a new value to the existing `VerticalPodAutoscalerStatus.Conditions` slice:

- **Type:** `InitialDelayActive`
- **Status:** `True` while the gate is active, `False` once it has elapsed.
- **Reason:** `WindowActive` (True) / `WindowExpired` (False).
- **Message:** human-readable summary including expiry timestamp.

This lets `kubectl describe vpa` surface the gate without an operator having to compute `CreationTimestamp + initialDelaySeconds` mentally.

### Metric

Emit a new gauge from the Updater:

```
vpa_initial_delay_active{namespace, name} = 0 | 1
```

Value is `1` while the gate is active for that VPA, `0` otherwise. Enables fleet-wide visibility into how many VPAs are currently gated, useful for dashboards and alerting on stuck windows.

### Feature Enablement and Rollback

Feature gate: **`VPAInitialDelay`**.

Components consuming the gate:

- `updater` — reads `initialDelaySeconds` and honours the gate.
- `admission-controller` — gates acceptance of the field on new/updated VPAs behind the feature gate, and honours the gate by skipping recommendation injection for pods created during an active window.

Enabling the feature gate causes:

- `admission-controller` to accept `initialDelaySeconds` on new/updated VPAs, and to skip recommendation injection for pods created while a VPA's window is active.
- `updater` to honour the gate.

Disabling the feature gate causes:

- `admission-controller` to reject new VPAs that set `initialDelaySeconds`, with a descriptive error message indicating the feature gate is disabled.
- `updater` and `admission-controller` to ignore the field on existing objects — the VPA behaves as if the field were not set (fail-open). This is deliberate: disabling the gate should never *increase* actuation restrictions on existing objects.

### Graduation Criteria

**Alpha (initial release):**

- Feature gate `VPAInitialDelay` disabled by default.
- Unit and e2e coverage as described in the [Test Plan](#test-plan).

**Alpha → Beta (gate enabled by default):**

- E2e tests stable for at least one release.
- No open bugs against the feature gate.
- Positive user feedback on the field semantics (in particular the `CreationTimestamp`-anchored PATCH behaviour).

**Beta → GA (gate locked on, then removed):**

- No bug reports attributable to the feature for two consecutive releases.

### Version Skew

The Recommender is unaffected by this feature. The gate is fully effective only when both the Updater and the Admission Controller run a version that supports it, with the feature gate enabled on both. During a rollout with mixed versions, an older component simply ignores the field and behaves as it does today (fail-open, matching the gate-disabled semantics in [Feature Enablement and Rollback](#feature-enablement-and-rollback)):

- **Old Updater, new Admission Controller** — pods created during the window are admitted without injection, but the Updater may still evict or resize during the window.
- **New Updater, old Admission Controller** — no evictions or resizes during the window, but pods created during the window still receive recommendation injection.

Neither skew causes errors or corrupted state; the failure mode is only that part of the gate is not honoured until the rollout completes. Operators who need strict gating should enable the feature gate only after all components are upgraded.

### Kubernetes Version Compatibility

This feature is entirely internal to the VPA controllers and depends on no new Kubernetes APIs. It is compatible with any Kubernetes version supported by the corresponding VPA release.

## Test Plan

**Unit tests** (`pkg/updater/logic/`):

- `inInitialDelayWindow` returns `true` when `now < CreationTimestamp + initialDelaySeconds`.
- `inInitialDelayWindow` returns `false` when `now >= CreationTimestamp + initialDelaySeconds`.
- `inInitialDelayWindow` returns `false` when `initialDelaySeconds` is `nil`, `0`, or absent.
- Gate short-circuits the eligibility path identically to `updateMode: Off` for `Recreate`, `InPlaceOrRecreate`, `InPlace`, `Initial`.
- Gate has no effect when `updateMode: Off` — behavior matches plain `updateMode: Off` regardless of the field's value.
- Metric and condition are set correctly on both sides of the transition.
- PATCH scenarios: shortening releases the gate, extending stays gated, removing closes the gate, extending after expiry re-arms only when the new expiry still lies in the future.

**Admission-controller tests:**

- Accept `initialDelaySeconds >= 1` on all update modes, including `Off` (no-op semantics).
- Reject `initialDelaySeconds < 1` and `initialDelaySeconds > 7776000` via CRD schema validation.
- Reject `initialDelaySeconds` on any VPA when the `VPAInitialDelay` feature gate is disabled.
- Pods created while the gate is active are admitted with their original spec resources, for every update mode.
- Pods created after the window elapses receive recommendation injection per the configured mode's normal semantics.

**E2E tests** (`e2e/`):

- Create a VPA with `updateMode: Recreate` and `initialDelaySeconds: 60`. Verify no eviction occurs during the first 60 s. Advance time past the window; verify the next reconcile evicts as expected.
- Repeat for `InPlaceOrRecreate`, `InPlace`, and `Initial` (with pod-creation-timing tests for the last).
- Create a VPA with `updateMode: Recreate` and `initialDelaySeconds: 3600`; scale the target Deployment up during the window and verify the new pod keeps its template resources (no admission-time injection). After expiry, verify newly created pods receive injected resources.
- Create a VPA with `initialDelaySeconds: 3600`, PATCH to `60` after 30 s, verify eviction becomes eligible around the 60 s mark.
- Create a VPA with `initialDelaySeconds: 60`, wait for expiry, PATCH to `3600`, verify the gate re-arms (the new expiry, `CreationTimestamp + 3600s`, still lies in the future) and eviction is again suppressed until the new expiry.
- Verify the `InitialDelayActive` condition and `vpa_initial_delay_active` gauge match the gate state throughout each scenario.

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
    initialDelaySeconds: 21600   # 6 hours
```

For the first six hours after creation, `web` pods are not evicted by the Updater regardless of recommendations, and pods created during the window (scale-ups, restarts) keep the resources from the `Deployment` spec — no admission-time injection occurs. After the window, `Recreate` behaviour takes effect.

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
    initialDelaySeconds: 300     # 5 minutes
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
    initialDelaySeconds: 3600    # 1 hour
```

Pods created during the first hour receive their `Deployment`-spec resources unchanged. Pods created after the window get VPA-injected resources.

## Alternatives Considered

**1. Extend `EvictionRequirements` with a `TimeSinceCreationExceeds` value.** Reuses an existing extensibility hook (AEP-4831), but forces a duration to be encoded inside an enum meant for direction-of-change comparisons (`TargetHigherThanRequests` / `TargetLowerThanRequests`). Awkward and probably a design regression on that API.

**2. Recommender-side confidence gate.** Block `RecommendationProvided` from becoming `True` for N hours or until a confidence threshold is met. Cleaner conceptually but hides visibility — operators would see no recommendation at all in `status.recommendation` during the window, defeating the "collect first, inspect, then apply" workflow this feature is meant to support.

**3. Global Updater flag** (e.g. `--default-observation-window-seconds`). Wrong axis: different applications in the same cluster want different windows, and different VPAs on the same target want different windows.

**4. Do nothing; rely on operator tooling.** The status quo. Ships the complexity downstream to every VPA operator; the CronJob-plus-patch-controller dance ends up reinvented per environment.

**5. Immutable field after creation.** Reject PATCH that modifies `initialDelaySeconds` via admission. Simpler in that it eliminates the re-arm and shortening cases, but forces users into delete-and-recreate to change the window — an expensive operation that also discards the VPA's recommendation history. The proposed stateless-gate design gets the same "one-line mental model" without that cost.

**6. Reset on "material" spec changes** (`resourcePolicy`, `targetRef`). Arguably more correct — a workload with a different target really does have different history semantics — but forces us to define "material" and requires a status field to track the current anchor. Deferred: this AEP does not preclude adding such semantics in a future revision, and the `CreationTimestamp` anchor is a strict subset of any future anchor policy.

**7. Reject PATCH that re-arms an expired gate.** Split the difference on (5) and the proposed design: allow PATCH freely during the window, but reject PATCHes that would push `expiry` past `now` once the window has already elapsed. Rejected because it makes the admission logic dependent on real time (surprising) and does not eliminate the underlying "PATCHes can change gate state" surprise — the `InitialDelayActive` condition and gauge already surface any re-arm immediately.

## Implementation History

- (issue filed) 2026-07-06 — Issue [kubernetes/autoscaler#9936](https://github.com/kubernetes/autoscaler/issues/9936) filed.
- (triage accepted) 2026-07-07 — `/triage accepted` from SIG member.
- (AEP PR opened) 2026-07-10 — PR [kubernetes/autoscaler#9962](https://github.com/kubernetes/autoscaler/pull/9962).
- (initial implementation) TBD.
- (alpha) TBD.
- (beta) TBD.
