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
  - [Interaction with Checkpoints](#interaction-with-checkpoints)
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
- [Alternatives](#alternatives)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Add an optional `initialDelaySeconds` field to `PodUpdatePolicy` that delays actuation of recommendations for a configurable duration after a VPA's creation. During the window the Recommender continues to compute and publish recommendations to `status.recommendation` normally, but nothing actuates them: the Updater performs no evictions or in-place resizes, and the Admission Controller does not apply recommendations to pods created during the window. The VPA behaves exactly as if `updateMode` were `Off`. Once the window elapses, the configured `updateMode` takes effect automatically with no operator action.

The window is computed as a pure function of `vpa.CreationTimestamp` and the current spec value, evaluated on every reconcile. No status writes are required. Modifying an existing VPA's spec does not reset the window — the anchor (`CreationTimestamp`) is immutable, and changing `initialDelaySeconds` itself simply moves the expiry relative to that anchor (see [Gate Evaluation](#gate-evaluation) for the full modification semantics).

## Motivation

When new `Deployment` objects are provisioned automatically — by CI, per-tenant provisioning, on-demand application creation, etc. — a VPA created alongside them and configured with a non-`Off` `updateMode` will start actuating on recommendations that have not stabilised yet. Recommendations are especially unstable in the first few hours of a workload's life because the recommender has few usage samples, and applying those recommendations causes disruptive churn: pods are restarted with the wrong resource requests, then restarted again a few minutes later with different requests.

The safe pattern is to run the VPA in `Off` mode for the first N hours to let recommendations stabilise, then flip to `Recreate` / `InPlaceOrRecreate` / `InPlace` / `Initial`. Doing that manually for every new VPA does not scale: cluster operators end up building external controllers, CronJobs, or admission-side patch scripts to sequence the mode transition. Every environment reinvents the same workflow.

This feature moves the sequencing into the VPA API itself. Operators declare "wait N seconds after creation before touching pods" alongside the update mode, and the Updater honours it.

### Goals

- Provide a declarative per-VPA field that delays actuation of recommendations — evictions, in-place resizes, and admission-time injection into newly created pods — for a configurable duration after VPA creation. CPU Startup Boost is not delayed (see [Interaction with CPU Startup Boost](#interaction-with-cpu-startup-boost)).
- Keep the Recommender's behaviour unchanged: recommendations are still computed and published to `status.recommendation` during the window, so operators can inspect them.
- Work uniformly across `Recreate`, `InPlaceOrRecreate`, `InPlace`, and `Initial`.
- Surface the gate state through both a status condition (`InitialDelayActive`) and a Prometheus gauge (`vpa_initial_delay_active`), so operators and dashboards can see when a VPA is gated without inspecting `spec`.

### Non-Goals

- **Recommender-confidence-based gating.** This feature is a user-declared policy, not a recommender-computed signal.
- **Reset semantics on modification of other fields.** Changing `resourcePolicy`, `targetRef`, or the target workload does not reset the window. The window is anchored on `vpa.CreationTimestamp`; users who need to re-observe a workload after a material change should delete and recreate the VPA.
- **Cluster-wide default windows.** Out of scope for this AEP; operators can inject a default with a Mutating Admission Policy (see [Alternatives](#alternatives)).

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

With the new `initialDelaySeconds`, the Updater and the Admission Controller treat the VPA as if `updateMode` were `Off` until `now >= vpa.CreationTimestamp + spec.updatePolicy.initialDelaySeconds`. After that, the configured `updateMode` takes effect.

Modifying the VPA's spec does not reset the window. The gate is a pure function of the immutable `vpa.CreationTimestamp` and the current value of `initialDelaySeconds`, re-evaluated on every reconcile — there is no per-object state to reset. The full modification semantics are described in [Gate Evaluation](#gate-evaluation). While the gate is active, the [`InitialDelayActive` status condition](#status-condition) surfaces it on the VPA object.

## Design Details

### Workflow

1. The user creates a VPA with a non-`Off` `updateMode` and `initialDelaySeconds: N` set on `spec.updatePolicy`.
2. The Recommender begins computing recommendations on the normal schedule and populates `status.recommendation`. It is unaware of and unaffected by the observation window.
3. On every Updater reconcile, before deciding whether the VPA is eligible for actuation, the Updater evaluates the gate. If the gate is active, the VPA is treated as if `updateMode` were `Off` for that reconcile: no pods are evicted (`Recreate` / `InPlaceOrRecreate`) and no in-place resize is attempted (`InPlace` / `InPlaceOrRecreate`).
4. The Admission Controller evaluates the same gate whenever a pod matching the VPA's target is created. While the gate is active it does not patch the pod's resources — the pod is admitted with its original spec, exactly as under `updateMode: Off`. This applies to every mode, not just `Initial`: without it, pods created during the window (scale-ups, node replacements, crash restarts) would receive un-stabilised recommendations at admission time.
5. The Updater sets the `InitialDelayActive` status condition to `True` while the gate is active and `False` once it has elapsed. It emits the `vpa_initial_delay_active` gauge accordingly.
6. After the window elapses (i.e. `now >= CreationTimestamp + InitialDelaySeconds`), the gate opens and the configured `updateMode` takes effect on subsequent reconciles and pod creations.

### API Changes

Extended `PodUpdatePolicy`:

```go
type PodUpdatePolicy struct {
    // ... existing fields (UpdateMode, MinReplicas, EvictionRequirements,
    //                     EvictAfterOOMSeconds) ...

    // InitialDelaySeconds specifies the number of seconds to wait after
    // the VPA object is created before recommendations are actuated,
    // regardless of the configured UpdateMode.
    //
    // During the window, the Recommender still computes and publishes
    // recommendations to status.recommendation, but the Updater and
    // Admission Controller treat the VPA as if UpdateMode were Off: no
    // evictions, in-place resizes, or injection into new pods. The
    // configured UpdateMode takes effect once the window elapses.
    //
    // If UpdateMode is Off, this field has no effect. Must be between
    // 1 and 7776000 (90 days) if set.
    // +optional
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=7776000
    InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`
}
```

Follows the shape of `EvictAfterOOMSeconds` (the only existing time-based field on the same struct) and `DurationSeconds` from AEP-7862; the name matches `initialDelaySeconds` on Pod probes.

In addition to the spec field, a new `VerticalPodAutoscalerConditionType` value is added:

```go
// InitialDelayActive indicates whether the VPA is currently within its
// configured initial delay window. It is True while the window is
// active and False once it has elapsed.
InitialDelayActive VerticalPodAutoscalerConditionType = "InitialDelayActive"
```

Reasons and message format are described in [Status Condition](#status-condition).

### Gate Evaluation

Reference implementation, exposed from a shared package (`pkg/utils/vpa`) and consumed by both actuating components:

```go
// InInitialDelayWindow returns true if the VPA is currently within its
// declared observation window and the Updater and Admission Controller
// should refrain from actuating recommendations.
func InInitialDelayWindow(vpa *vpa_types.VerticalPodAutoscaler, now time.Time) bool {
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

When `InInitialDelayWindow` returns `true`, both components take the same code path they take for `UpdateModeOff` today.

`InInitialDelayWindow` is exported from the shared `pkg/utils/vpa` package so both the Updater and the Admission Controller consume the identical implementation. Each call site first checks the `VPAInitialDelay` feature gate (`features.Enabled(features.VPAInitialDelay)`); when the gate is disabled the helper is not consulted and the component behaves exactly as it does today (fail-open).

The gate is stateless: it is a pure function of `CreationTimestamp` (immutable on the object) and the current spec value (mutable). No caching, no status writes.

No VPA spec change affects the gate **except** modifying `initialDelaySeconds` itself: the new value simply moves the expiry (`CreationTimestamp + initialDelaySeconds`), so on the next reconcile the gate may open earlier (value shortened or removed) or stay closed longer (value extended). No other field — existing or added to the CRD in the future — participates in gate evaluation; `updateMode` changes only what happens once the gate opens.

Re-arm after expiry is possible but not guaranteed: because the window is anchored to `CreationTimestamp`, a modification re-arms the gate only when the new expiry still lies in the future. This is intentional — it lets users extend an observation window in response to observed instability without deleting the VPA, while a larger-but-already-elapsed value on an old VPA simply has no effect. If reviewers consider this a footgun serious enough to justify admission logic, we can add a check rejecting PATCHes that would push `expiry` past `now` once the window has already elapsed.

### Interaction with `updateMode`

The observation window is orthogonal to `updateMode`. While the gate is active, the Updater and the Admission Controller behave as if `updateMode` were `Off` regardless of the configured value: no evictions, no in-place resizes, and no recommendation injection into pods created during the window.

### Interaction with CPU Startup Boost

CPU Startup Boost (AEP-7862) still applies during the delay window: the boost is a startup-safety mechanism computed from the pod's own spec, not a recommendation, and the Admission Controller already applies it even under `updateMode: Off` (the `Off` short-circuit in `resource/vpa/matcher.go` is bypassed for VPAs with a boost). The gate reuses the same `Off` semantics, so boosted containers remain boosted.

The interaction appears at boost expiry, when the Updater performs the post-boost in-place resize:

- Delay window still active → the pod is scaled down to its original spec resources; no recommendation is actuated.
- Delay window elapsed → the pod is resized to the VPA recommendation, per normal post-boost behaviour.

### Interaction with Checkpoints

A VPA's checkpoint stores the recommender's history for that VPA. Checkpoints are keyed by VPA name, not by a unique object ID, and one is only deleted once no VPA with that name exists (the recommender's GC pass, every 10 minutes by default via `--checkpoints-gc-interval`).

So deleting a VPA and recreating it with the same name hands the new object the old checkpoint, as long as GC hasn't removed it yet. The recommender loads that history and fills `status.recommendation` with mature values almost immediately, even though the new VPA has a fresh UID and creation time.

This matters for the delay window, because a freshly recreated VPA can already hold mature recommendations while its window is still active. The window still starts from the new object's `CreationTimestamp`, whether the recommendations were restored from a checkpoint or built from scratch. Making the gate look at checkpoints would add state and turn `initialDelaySeconds` into a confidence signal, which is a [non-goal](#non-goals). Anyone who recreates a VPA and is happy with the restored recommendations can just set a smaller `initialDelaySeconds`, or leave it off, on the new object.

### Validation

`initialDelaySeconds` must be an integer between `1` and `7776000` (90 days) if set. The upper bound follows the [API conventions for numeric fields](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#numeric-fields); observation windows are realistically hours to days, so 90 days is generous headroom while still rejecting absurd values.

The range is enforced in two places:

- **CRD schema** — `Minimum=1`, `Maximum=7776000`, as the backstop.
- **Admission webhook** — the same check added to the existing VPA validation in `pkg/admission-controller/resource/vpa/validation.go`, which returns descriptive error messages, and which is also where the field is rejected when the `VPAInitialDelay` feature gate is disabled.

No cross-field validation or restrictions on modifying the field are required: all lifecycle transitions are handled declaratively by the gate.

### Status Condition

Add a new value to the existing `VerticalPodAutoscalerStatus.Conditions` slice:

- **Type:** `InitialDelayActive`
- **Status:** `True` while the gate is active, `False` once it has elapsed.
- **Reason:** `WindowActive` (True) / `WindowExpired` (False).
- **Message:** human-readable summary including expiry timestamp.

This lets `kubectl describe vpa` surface the gate without an operator having to compute `CreationTimestamp + initialDelaySeconds` mentally.

### Metric

Emit a new gauge from the Updater:

```text
vpa_initial_delay_active{namespace, name} = 0 | 1
```

Value is `1` while the gate is active for that VPA, `0` otherwise. Enables fleet-wide visibility into how many VPAs are currently gated, useful for dashboards and alerting on stuck windows.

The condition and gauge report the window state, not effective suppression — so under `updateMode: Off` they still read active until the window elapses.

### Feature Enablement and Rollback

Feature gate: **`VPAInitialDelay`**.

Both the `updater` and the `admission-controller` read `initialDelaySeconds` and honour the `VPAInitialDelay` feature gate. While the window is active:

- `updater` — makes no evictions or in-place resizes for the VPA.
- `admission-controller` — skips recommendation injection for newly created pods. It also validates the field on write and rejects it when the gate is disabled.

Disabling the gate causes:

- the `admission-controller` to reject new VPAs that set `initialDelaySeconds`, with a descriptive error.
- both components to ignore the field on existing objects, so the VPA behaves as if it were unset (fail-open).

### Graduation Criteria

**Alpha (initial release):**

- Feature gate `VPAInitialDelay` disabled by default.
- Unit and e2e coverage as described in the [Test Plan](#test-plan).

**Alpha → Beta (gate enabled by default):**

- E2e tests stable for at least one release.
- No open bugs against the feature gate.
- Positive user feedback on the field semantics.

**Beta → GA (gate locked to enabled):**

- No bug reports attributable to the feature during the beta release.

### Version Skew

The Recommender is unaffected by this feature. The gate is fully effective only when both the Updater and the Admission Controller run a version that supports it, with the feature gate enabled on both. During a rollout with mixed versions, an older component simply ignores the field and behaves as it does today (fail-open, matching the gate-disabled semantics in [Feature Enablement and Rollback](#feature-enablement-and-rollback)):

- **Old Updater, new Admission Controller** — pods created during the window are admitted without injection, but the Updater may still evict or resize during the window.
- **New Updater, old Admission Controller** — no evictions or resizes during the window, but pods created during the window still receive recommendation injection.

Neither skew causes errors or corrupted state; the failure mode is only that part of the gate is not honoured until the rollout completes. Operators who need strict gating should enable the feature gate only after all components are upgraded.

**CRD schema skew.** Apply the updated CRD with (or before) the controller upgrade; a newer controller against an older CRD simply has the field pruned by the apiserver and treats it as unset (fail-open).

### Kubernetes Version Compatibility

This feature is entirely internal to the VPA controllers and depends on no new Kubernetes APIs. It is compatible with any Kubernetes version supported by the corresponding VPA release.

## Test Plan

**Updater tests:**

- `InInitialDelayWindow` returns `true` when `now < CreationTimestamp + initialDelaySeconds`.
- `InInitialDelayWindow` returns `false` when `now >= CreationTimestamp + initialDelaySeconds`.
- `InInitialDelayWindow` returns `false` when `initialDelaySeconds` is `nil`, `0`, or absent.
- Gate short-circuits the eligibility path identically to `updateMode: Off` for `Recreate`, `InPlaceOrRecreate`, `InPlace`, `Initial`.
- Gate has no effect when `updateMode: Off` — behavior matches plain `updateMode: Off` regardless of the field's value.
- Metric and condition are set correctly on both sides of the transition.
- Modification scenarios: shortening releases the gate, extending stays gated, removing closes the gate, extending after expiry re-arms only when the new expiry still lies in the future.

**Admission-controller tests:**

- Accept `initialDelaySeconds >= 1` on all update modes, including `Off` (no-op semantics).
- Reject `initialDelaySeconds < 1` and `initialDelaySeconds > 7776000` via CRD schema validation.
- Reject `initialDelaySeconds` on any VPA when the `VPAInitialDelay` feature gate is disabled.
- Pods created while the gate is active are admitted with their original spec resources, for every update mode.
- Pods created after the window elapses receive recommendation injection per the configured mode's normal semantics.

**Integration tests** (updater and admission-controller):

- Create a VPA with `updateMode: Recreate`, `initialDelaySeconds: 60`, and a recommendation far enough from the pod's requests that eviction is expected. Watch for `InitialDelayActive=True` and verify no pods are evicted. Patch `initialDelaySeconds` to `1`; watch for `InitialDelayActive=False` and verify pods are evicted.
- Repeat for `InPlaceOrRecreate`, `InPlace`, and `Initial` (for `Initial`, assert on injection at new-pod creation instead of eviction).
- Admission gating: while `InitialDelayActive=True`, scale the target Deployment up and verify the new pod keeps its template resources; after patching `initialDelaySeconds` to `1` and observing `InitialDelayActive=False`, verify newly created pods receive injected resources.
- Re-arm: create a VPA with `initialDelaySeconds: 1`; watch for `InitialDelayActive=False`. Patch to `3600` (new expiry still lies in the future); watch for `InitialDelayActive=True` and verify actuation is suppressed again.
- Throughout each scenario, verify the `vpa_initial_delay_active` gauge matches the condition.

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

## Alternatives

**1. External tooling (today's status quo).** Without this feature, you get the same behavior by creating the VPA in `Off` mode and flipping it to an active mode after N hours — a CronJob or a small controller that patches `updateMode`. It works, but every team rebuilds the same automation, and there's no per-VPA declarative way to say "wait N hours before acting."

**2. Extend `EvictionRequirements` with a time-based condition.** `EvictionRequirements` (AEP-4831) already gates eviction on conditions. A duration doesn't fit its enum, which compares the recommendation against requests (`TargetHigherThanRequests` / `TargetLowerThanRequests`), and it only covers eviction — not admission-time injection or in-place resize.

**3. Default the field with a Mutating Admission Policy.** For a cluster-wide default, a [Mutating Admission Policy](https://kubernetes.io/docs/reference/access-authn-authz/mutating-admission-policy/) can set `initialDelaySeconds` on VPAs that don't specify it, using native Kubernetes with no VPA-side flag. This complements the field rather than replacing it.

## Implementation History

* 2026-07-10: Initial version.
