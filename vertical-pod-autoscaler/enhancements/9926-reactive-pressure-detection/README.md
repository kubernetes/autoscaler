# AEP-9926: Reactive Memory Pressure Detection for VPA

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [Terminology](#terminology)
  - [User Stories](#user-stories)
    - [Story 1: a rollout leaves new Pods stalled for up to 12 hours](#story-1-a-rollout-leaves-new-pods-stalled-for-up-to-12-hours)
    - [Story 2: a startup spike that never stalls is left alone](#story-2-a-startup-spike-that-never-stalls-is-left-alone)
  - [How It Works](#how-it-works)
- [Design Details](#design-details)
  - [Detection Rule](#detection-rule)
  - [Pressure Observer](#pressure-observer)
  - [Sample Injection](#sample-injection)
  - [Quick Pressure Update](#quick-pressure-update)
  - [Scalability](#scalability)
  - [Feature Interactions](#feature-interactions)
  - [Observability](#observability)
  - [Notes, Constraints, and Caveats](#notes-constraints-and-caveats)
    - [One node can raise a cohort](#one-node-can-raise-a-cohort)
    - [Known false negatives](#known-false-negatives)
    - [No ceiling unless maxAllowed is set](#no-ceiling-unless-maxallowed-is-set)
    - [An infeasible upsize stays wedged](#an-infeasible-upsize-stays-wedged)
    - [Disabling does not undo a raise](#disabling-does-not-undo-a-raise)
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

VPA reacts to hard memory failure in two steps. `RecordOOM` adds an inflated peak to the histogram, and
the updater's quick-OOM path lets the raised recommendation bypass the pod-lifetime and minimum-change
gates. Linux PSI exposes a softer failure: a container that spends much of its time stalled on memory
reclaim, stays alive, and may never be killed. Neither OOM step fires for that container.

This AEP adds both steps for pressure. An opt-in observer in the recommender reads per-container memory
PSI from the kubelet Summary API. When sustained stalls coincide with a working set near the limit, the
recommender adds an inflated memory peak to the histogram and records the time in VPA status. For a
bounded window after that, the updater applies the current recommendation in place to any Pod whose
memory request is below the target, without waiting for the usual gates. There is no new update mode,
and VPA remains the only writer of Pod resources.

## Motivation

[Issue #9926](https://github.com/kubernetes/autoscaler/issues/9926) asks VPA to respond to memory
pressure before it becomes a kill. The gap that matters most is actuation.

VPA applies a new target only when the request falls outside `[lowerBound, upperBound]` or the Pod is
older than `PodLifetimeUpdateThreshold` (12 hours). Quick OOM bypasses those gates, but a container
stalled on memory reclaim is often never killed, so it waits. Users have reported Pods held this way
([#6243](https://github.com/kubernetes/autoscaler/issues/6243)) and asked for urgent recovery to be kept
apart from ordinary upsizing ([#10174](https://github.com/kubernetes/autoscaler/issues/10174)).

MemoryQoS ([KEP-2570](https://github.com/kubernetes/enhancements/issues/2570)) sets `memory.high`, which
the kernel documents as a boundary for a management agent to watch and act on. SIG Node reviewed an
in-kubelet response ([KEP-5986](https://github.com/kubernetes/enhancements/pull/6141)) and preferred one
that goes through the API, so a resize persists in the Pod spec and a restart does not undo it
([discussion](https://github.com/kubernetes/enhancements/pull/6141#issuecomment-4652914248)). This AEP is
that response for VPA-managed workloads.

### Goals

- For opted-in containers under sustained memory reclaim stall near their limit, contribute
  pressure-derived memory peaks to VPA's existing recommendation model
- When pressure is observed on a container whose request is below its target, apply the target in place
  within a bounded window instead of after the pod-lifetime gate
- Preserve existing usage and OOM handling, and every updater safety constraint other than the two gates
  quick OOM already bypasses
- If PSI is unavailable or observations are invalid, retain usage-and-OOM recommendations unchanged

This does not promise that every pressure event raises the target or prevents an OOM. Recommendations
pool aggregate states by container name, so a pressure peak is evidence added to a cohort.

### Non-Goals

- A new `UpdateMode`. Quick pressure updates act inside `InPlace` and `InPlaceOrRecreate`.
- Evicting a Pod because of pressure.
- CPU pressure. PSI cannot separate contention from throttling against a user-requested limit
  ([enhancements#5062](https://github.com/kubernetes/enhancements/issues/5062)).
- `memory.events` as a severity signal. It is absent from the Summary API, and `events:high` falls as a
  stall worsens, because a throttled task allocates less often.
- Containers without a memory limit, pod-level VPAs
  ([AEP-7571](https://github.com/kubernetes/autoscaler/pull/9988)), which aggregate separately from
  container histograms, and device memory such as GPU VRAM.
- Writing any cgroup value or acting at node level.

## Proposal

### Terminology

- **Stall rate**: `delta(Memory.PSI.Some.Total) / delta(Memory.Time)` from the Summary API, the fraction
  of time at least one task in the container waited on memory.
- **Working-set fraction**: `Memory.WorkingSetBytes` divided by the container's memory limit.
- **Pressure peak**: an inflated memory sample the recommender adds after sustained stall, like the OOM
  peak.
- **Cohort**: the aggregate states VPA merges, by container name, into one recommendation.
- **Quick pressure**: the updater path that applies the target in place within a window after a pressure
  peak, like quick OOM.

### User Stories

#### Story 1: a rollout leaves new Pods stalled for up to 12 hours

A Deployment under a VPA with days of history rolls to a version that needs more memory. The admission
controller sizes the new Pods from the old recommendation, so they start stalled. The recommender raises
the target within minutes, but the history keeps the lower bound under the request, so the updater waits
for the 12-hour gate. A proof of concept reproduced this on single-node kind with MemoryQoS enabled, a
synthetic workload holding ~503Mi (two replicas), and 8 days of history injected through a checkpoint:

| | stock VPA | this feature |
|---|---|---|
| Memory target after rollout | 599Mi | 683Mi |
| In-place resize | none in 20 minutes | ~230s and ~290s after rollout (one Pod per loop) |
| Memory PSI `some` stall rate | ~0.6 throughout | ~0.6 → 0.00 |
| Workload operations/s | ~1 | ~650–850 across runs |

For a VPA without history the confidence term lifts the lower bound past the request within minutes and
stock VPA resizes on its own, so the benefit is specific to VPAs with history.

#### Story 2: a startup spike that never stalls is left alone

New Pods spike during startup and settle without any reclaim stall. Relaxing the 12-hour gate globally
would resize them and leave them over-provisioned. This feature does not, because it acts only after a
stall is measured in the cohort. The bypass is cohort-wide, not per Pod: once one replica stalls, other
replicas below the target can be updated too (see [Quick Pressure Update](#quick-pressure-update)).

### How It Works

1. An observer in the recommender polls the kubelet Summary API on nodes running opted-in containers.
2. When the stall rate and the working-set fraction both stay high over two windows, it emits an event.
3. The recommender adds a pressure peak to the container's histogram and sets `lastPressureTime` in VPA
   status.
4. For `--pressure-quick-update-window`, the updater applies the target in place to any Pod whose memory
   request is below it, raising resources only and never evicting.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: my-app-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  updatePolicy:
    updateMode: InPlaceOrRecreate
  resourcePolicy:
    containerPolicies:
      - containerName: '*'
        pressureDetection: Enabled
        maxAllowed:
          memory: 1Gi
```

`GetContainerResourcePolicy` returns the named policy or the wildcard, whole, with no field merge, so a
container with its own named policy is not opted in by `pressureDetection` on `'*'`.

Worked example, request 256Mi, limit 512Mi, two polls 60s apart: `Memory.PSI.Some.Total` moves from
41.9s to 55.7s, a stall rate of `13.8 / 60 = 0.23` (above the `0.10` default), and the working set of
486Mi is `0.95` of the limit (above `0.90`). If both also held on the previous pair, the candidate peak
is `max(486Mi + 64Mi, 486Mi * 1.15) ≈ 559Mi`, which replaces the interval peak only if higher.

## Design Details

### Detection Rule

For each eligible container, both conditions must hold over two consecutive windows (three fresh
samples):

1. **Stall rate** above `--pressure-stall-ratio` (default `0.10`).
2. **Working-set fraction** at the end of each window at least `--pressure-working-set-fraction`
   (default `0.90`).

A failed, missing or invalid poll resets the count. A baseline begins `--pressure-startup-grace`
(default 60s) after the container starts. After an accepted sample the container is quiet for
`--pressure-cooldown` (default 10 minutes), then needs two fresh windows. While a resize is pending or in
progress, including one started outside VPA, nothing is injected and the container is reported as
`resize_in_progress`, `resize_deferred` or `resize_infeasible`.

With the defaults, the earliest detection is about three minutes after a container starts: the grace
period, a baseline sample and two 60s windows. Recommender and updater loops each add up to a minute.

The rule uses `some`, not `full`: memory PSI is accounted per CPU, so `full` stays low whenever another
task can run on the stalled task's CPU. It uses cumulative `total`, not `avg10`, so the observer defines
its own windows, rejects stale samples and requires new evidence after acting; the averages keep
reporting a stall for tens of seconds after it ends. Thresholds are not borrowed from systemd-oomd or
Senpai, which use PSI to kill and to reclaim.

`memory.high` is not required: `memoryThrottlingFactor` defaults to `nil` in 1.37
([k/k#140007](https://github.com/kubernetes/kubernetes/pull/140007)), so the rule keys on reclaim stall,
not on the throttle boundary.

**Availability.** PSI may need `psi=1` at boot, and the cadvisor-backed Summary path returns zeros
rather than nothing when accounting is off. The observer never claims availability from zeros: an absent
PSI field or a transport error clears the window count, and zeros simply never qualify.

### Pressure Observer

The observer has its own input path, like `pkg/recommender/input/oom/observer.go`, because the usage
pipeline carries CPU and memory quantities only. Each recommender loop rebuilds its target set from the
Pod lister, not `BasicPodSpec`, which lacks UID, node name and resize conditions. A container is eligible
when it is running, belongs to a VPA handled by this recommender, resolves to `pressureDetection:
Enabled`, permits memory recommendations and has a positive memory limit. Init containers are skipped,
as in the usage path.

**Identity and freshness.** Summary entries are matched by Pod UID and container name. A new container
start, a changed limit or a decreasing counter resets the baseline. Samples older than
`--pressure-max-sample-age` (default 2 minutes) or in the future are discarded, and a gap longer than two
fetch intervals starts a new baseline.

**Units.** The Summary PSI total is in microseconds on the cadvisor stats path and nanoseconds on the
CRI stats path, on the same kubelet version. `--pressure-psi-total-unit` (default `microseconds`) is
therefore explicit, and a computed rate above `1 + 1s/Δt` is discarded, since a stall rate cannot exceed
1 and Summary timestamps have one-second resolution. Without that guard a unit mismatch makes every busy
container qualify.

**Delivery.** The channel holds 5000 events with non-blocking sends; a full channel drops and counts. A
dropped event starts no cooldown and resets the container's window count, so a container still under
pressure qualifies again two windows later; the sweep's rotating start spreads which containers reach a
nearly full queue first. The detector drops stale samples, replaced containers and containers with a
resize outstanding before emitting. `LoadRealTimeMetrics` drains the channel after usage samples, drops
an event whose container has left the target set since emission (an opt-out or a deleted Pod), and calls
`RecordPressure`. Gate-off constructs no observer.

### Sample Injection

`addMemorySample` takes an `isOOM bool` that selects the peak field and bypasses the out-of-order guard.
A third source needs both independently, so it becomes an enumeration:

```golang
type memorySampleSource int

const (
	sampleSourceUsage memorySampleSource = iota
	sampleSourceOOM
	sampleSourcePressure
)
```

`ContainerState` gains `pressurePeak`, zeroed on window roll and included in `GetMaxMemoryPeak` and in
the checkpoint writer's `subtractCurrentContainerMemoryPeak`. `RecordPressure` accepts only a timestamp
within the current interval, never rolls the interval and never moves the usage watermark, so it cannot
reject or displace a later usage sample. It is accepted even when older than the latest usage sample,
because the observer's sample can trail the metrics sample:

```golang
func (container *ContainerState) RecordPressure(timestamp time.Time, workingSet ResourceAmount) error {
	intervalStart := container.WindowEnd.Add(-container.GetMemoryAggregationIntervalDuration())
	if container.WindowEnd.IsZero() || timestamp.Before(intervalStart) || !timestamp.Before(container.WindowEnd) {
		return errors.New("pressure sample outside active memory aggregation interval")
	}
	// Reading memoryPeak, never pressurePeak, prevents reinflating a previous synthetic peak.
	memoryUsed := ResourceAmountMax(workingSet, container.memoryPeak)
	memoryNeeded := ResourceAmountMax(memoryUsed+container.GetPressureMinBumpUp(),
		ScaleResource(memoryUsed, container.GetPressureBumpUpRatio()))
	sample := ContainerUsageSample{MeasureStart: timestamp, Usage: memoryNeeded, Resource: ResourceMemory}
	if !container.addMemorySample(&sample, sampleSourcePressure) {
		return errors.New("adding pressure sample failed")
	}
	return nil
}
```

`RecordOOM` also reads only `memoryPeak`, so a pressure bump followed by an OOM in the same interval does
not stack. Across containers the estimator uses the pooled, decayed histogram, so one pressure peak among
many equal peaks may not move P90: an observation can leave the target unchanged, and a changed target
applies to the cohort.

### Quick Pressure Update

**Recommender.** An accepted sample records its time on the container's aggregate state, merged across
Pods by taking the latest. When the recommender writes `status.recommendation`, it sets
`lastPressureTime` from that time, carries the existing value forward otherwise, and clears it for any
container whose policy is no longer `pressureDetection: Enabled`. The field is bounded by containers per
VPA, not by Pods.

The stamp does not require the pressure peak to raise the target. Whether a stalled Pod needs a resize
depends on its request against the current target, and in the rollout case the usage path had already
raised it: with the bump set to zero, quick pressure still resized both Pods.

**Updater.** In `UpdatePriorityCalculator.AddPod`, a Pod that the lifetime or minimum-change gate would
hold is admitted when all of these hold:

1. The gate is on and the update mode is `InPlace` or `InPlaceOrRecreate`.
2. No resize is pending or in progress for the Pod.
3. Some container has `lastPressureTime` within `--pressure-quick-update-window`. The updater reads it
   from the raw VPA status, because post-processors may rebuild processed entries without the field.
4. That container still resolves to `pressureDetection: Enabled`.
5. Its memory request is below its processed target.

Like quick OOM, it bypasses only pod lifetime and minimum change, and skips a Pod whose resources would
not change. `maxAllowed`, the in-place disruption tolerance and rate limiter, CPUStartupBoost exclusion
and the infeasible-retry rule still apply. Where it differs from quick OOM, it does so on purpose:

| | quick OOM | quick pressure |
|---|---|---|
| Trigger lasts | while `LastTerminationState` shows the OOM, which under `InPlace` can re-trigger every loop ([#10173](https://github.com/kubernetes/autoscaler/pull/10173)) | a window after the last accepted sample |
| Resources | whole recommendation, including decreases | `max(current request, target)` per resource |
| Eviction | may evict | never; the Pod is skipped for that loop |
| Pods per loop | all qualifying | at most `--pressure-quick-update-fraction` (default `0.1`, at least one) per VPA |
| Update modes | all | `InPlace`, `InPlaceOrRecreate` |

- **Why clamp.** The in-place patch applies the whole recommendation, so a memory stall could apply a CPU
  decrease; on stock VPA a memory-driven update cut a young Pod's CPU from 1 to 25m. Skipping such Pods
  instead left them stalled. The in-place update still applies `maxAllowed` and LimitRange caps, so if a
  cap below the current request would lower a resource anyway, the Pod is skipped for that loop.
- **Why never evict.** With a ResourceQuota that rejected the resize, eviction left the Deployment a
  replica short while the quota blocked the replacement.
- **Why a cap.** The stamp is per container name, so every replica below the target qualifies, and with
  the chart's defaults (`--in-place-skip-disruption-budget=true`, and an in-place rate limiter that
  shares `--eviction-rate-limit`, default `-1`) nothing else throttles in-place updates. The cap is a
  fraction of all the VPA's Pods, not of this loop's candidates. A cohort larger than the loops left in
  the window is not covered in one window: the remaining Pods wait for the next accepted sample, which
  restamps while the stall lasts, or for normal gating. Only Pods that quick pressure alone admits, and
  whose resources would change, count; a Pod the stock gates admit keeps the stock path, including its
  eviction fallback.

The existing `InPlaceResizedByVPA` event says when a Pod was admitted by quick pressure; otherwise the
reason appears only in an updater log line at verbosity 4.

`--pressure-quick-update-window` defaults to 10 minutes, matching `--evict-after-oom-threshold`.

### Scalability

**Node discovery.** Nodes come from `spec.nodeName` in the existing Pod informer, plus a shared Node
informer for addresses. Only nodes with eligible containers are queried; with no opt-ins there are no
fetches. This does not depend on VPASlice ([#10229](https://github.com/kubernetes/autoscaler/pull/10229)).

**Transport.** `--pressure-transport` selects `direct` (default) or `apiserver-proxy`. Direct fetches
`/stats/summary?only_cpu_and_memory=true` from the kubelet with the service-account token, a serving
certificate verified against `--pressure-kubelet-ca-file` (else the API client CA) and no redirects, and
needs `get nodes/stats` and read access to Nodes. It dials the node's InternalIP and verifies the
certificate against the node name, because kubelet serving certificates often carry only a DNS SAN. It
therefore requires serving certificates from one CA the recommender trusts (`serverTLSBootstrap: true`);
with kubeadm's per-node self-signed certificates, use the proxy. Proxy needs `get nodes/proxy`, which
also permits command execution in any container, and is never a fallback. Each permission ships as a
separate opt-in RBAC manifest, not in the default install, and the chart grants neither by default
(`recommender.memoryPressure.nodeStatsAccess`, `recommender.memoryPressure.nodeProxyAccess`).

**Scheduling.** One fetch per node per `--pressure-fetch-interval` (default 60s), at most
`--pressure-max-concurrent-fetches` (default 50) in flight, 5s per request, 45s per sweep, 16MiB per
response. Failing nodes back off exponentially with jitter up to 5 minutes, honouring `Retry-After`.
The observer runs outside the recommender loop.

**Coverage.** Detection needs two windows, so a node revisited less often than every two fetch intervals
can never qualify. Sweeps rotate their starting point, and nodes cut by the sweep deadline are fetched
first in the next sweep, so the same tail is not starved. Concurrency is derived each sweep as
`ceil(eligibleNodes * p95FetchLatency / sweepTimeout)`, clamped to the maximum. When even the maximum
cannot cover every node, uncovered nodes are reported as `undersampled`.

Measured with KWOK, 5000 nodes with one opted-in Pod each and 50 stalled, over the direct transport:

| fetch latency | revisit p50 / max | stalled nodes detected |
|---|---|---|
| 20ms | 63s / 76s | 50 / 50 |
| 450ms | 61s / 95s | 50 / 50 |
| 1000ms | 133s / 182s | 0 / 50 |

With the defaults, full coverage holds up to about 450ms mean fetch latency; beyond that, alert on the
revisit histogram. The load is about 83 requests/s at 5000 nodes, each response carrying every Pod on
the node. Observer state measured 581 bytes per eligible container, about 305MiB at 550,000 containers.

### Feature Interactions

| Feature | Interaction |
|---|---|
| Eviction requirements ([#10181](https://github.com/kubernetes/autoscaler/issues/10181)) | Apply unchanged. Quick pressure never authorizes an eviction. |
| CPUStartupBoost | Boosting Pods are excluded from updates; a raised target waits for the boost to end. |
| Observation windows ([AEP-9936](https://github.com/kubernetes/autoscaler/pull/9962)) | Collection continues, so a window can accumulate pressure peaks. A `lastPressureTime` older than the quick window when actuation reopens gives no bypass. |
| Per-VPA percentiles ([AEP-9970](https://github.com/kubernetes/autoscaler/pull/9994)) | A low memory percentile can absorb a single pressure peak, as it can an OOM peak. |
| Request/limit ratio ([AEP-8515](https://github.com/kubernetes/autoscaler/pull/8516)) | Only request recommendations change; limit derivation stays with existing policy. `RequestsOnly` can cap the usable request at an unchanged limit. |
| Resize preemption ([KEP-5836](https://github.com/kubernetes/enhancements/issues/5836)) | Honoured as configured; never enabled by this feature. |
| Swap | `disallowResizeForSwappableContainers` still rejects restart-free memory resizes of swappable containers. |
| Guaranteed QoS | MemoryQoS sets no container `memory.high` when request equals limit, so a Guaranteed container stalls only in limit reclaim shortly before an OOM kill. The detector does not need `memory.high`, but has not been measured on Guaranteed Pods. An [opt-in `memory.high` for Guaranteed](https://github.com/kubernetes/enhancements/pull/6141#issuecomment-4908522670) would give it a throttle window to act in. |
| Proactive reclaim | Out of scope ([enhancements#6398](https://github.com/kubernetes/enhancements/issues/6398) was closed leaving it to node agents). Without swap, `memory.reclaim` cannot lower the working set. |
| MemoryQoS | A VPA resize re-runs the kubelet's `memory.high` formula, so VPA should be the only actor on a container's boundary. Three kubelet gaps limit the result, measured on v1.37.0: the Summary API does not expose effective `memory.high`; after a resize that made request equal limit the container kept its old `memory.high` ([k/k#142569](https://github.com/kubernetes/kubernetes/issues/142569)); and under `memoryReservationPolicy: TieredReservation` a request-only resize leaves pod-level `memory.low` at the old request ([k/k#142572](https://github.com/kubernetes/kubernetes/issues/142572)). |

### Observability

| Metric | Type | Labels | Description |
|---|---|---|---|
| `vpa_recommender_pressure_events_total` | Counter | `result` | Detector: `stale`, `identity_changed`, `rate_out_of_range`, `resize_pending`, `queue_full`. Feeder: `accepted`, `policy_changed`, `window_mismatch`, `unknown_container` |
| `vpa_recommender_pressure_containers` | Gauge | `state` | Containers meeting both conditions in the latest window: `injectable`, `cooldown`, `resize_in_progress`, `resize_deferred`, `resize_infeasible` |
| `vpa_recommender_pressure_estimate_to_recommendation_ratio` | Histogram | none | Accepted pressure estimate divided by the last memory recommendation |
| `vpa_recommender_pressure_fetch_duration_seconds` | Histogram | `result` | Per-node fetch latency: `success`, `timeout`, `throttled`, `error` |
| `vpa_recommender_pressure_nodes` | Gauge | `state` | Nodes by last fetch: `observed`, `missing_psi`, `failed`, `stale`, `undersampled` |
| `vpa_recommender_pressure_node_revisit_seconds` | Histogram | none | Time between consecutive successful fetches of a node |
| `vpa_updater_pressure_quick_updates_total` | Counter | `result` | `admitted`, `capped`, `no_change`, `eviction_skipped`, `decrease_blocked` when caps would still lower a resource, and `decrease_clamped` per resource held at its request |

A pressure estimate is not a usage measurement, so it is not fed into the `vpa_quality_*` series.
Evaluation compares opted-in and stock VPA on the same consuming workloads and reports OOM kill rate and
memory slack together, with throughput at a fixed tail-latency target and time stalled: slack alone
rewards under-provisioning and OOM rate alone rewards excess. The in-tree KWOK benchmark runs no
consuming applications, so this needs a real-workload harness.

### Notes, Constraints, and Caveats

#### One node can raise a cohort

PSI does not say whether a stall comes from the container's limit, node-level reclaim or page-cache
refaults, so a container near its limit on a pressured node can qualify. With few replicas one bad node
can raise the whole cohort's recommendation; with many, one peak may not move the percentile. The rule
is a selection heuristic, not attribution.

#### Known false negatives

Limit reclaim of page cache can happen at a low working-set fraction, but clean cache is cheap to
reclaim and adds little stall. A low `memory.high` (for example `memoryThrottlingFactor=0.5`) throttles
below the working-set threshold, because the Summary API does not expose the effective boundary.

#### No ceiling unless maxAllowed is set

Existing policy caps apply, but without `maxAllowed` or a global maximum the feature adds no ceiling. The
bump applies to a sample, not the request: a 559Mi peak becomes a ~643Mi target after the 15% safety
margin, and a 2:1 limit ratio gives ~1.26Gi. The documentation recommends a workload-appropriate
`maxAllowed.memory`. A fixed input cannot reinflate its own peak, but rising usage over later intervals
can produce larger estimates.

#### An infeasible upsize stays wedged

Under `InPlace`, a cached infeasible attempt is retried only for a lower recommendation, so an infeasible
upsize is not retried until the Pod is replaced. The container stays reported as `resize_infeasible`:
this is a wedged Pod the feature should surface, not hide.

#### Disabling does not undo a raise

Injected peaks stay in the histogram and decay like OOM peaks, and completed-interval peaks survive a
restart through checkpoints.

### API Changes

```golang
type ContainerResourcePolicy struct {
	// ... existing fields ...

	// pressureDetection controls whether the recommender observes this container for sustained
	// memory reclaim stall. Defaults to Disabled.
	// +kubebuilder:validation:Enum=Enabled;Disabled
	// +optional
	PressureDetection *PressureDetectionMode `json:"pressureDetection,omitempty"`
}

type PressureDetectionMode string

const (
	PressureDetectionEnabled  PressureDetectionMode = "Enabled"
	PressureDetectionDisabled PressureDetectionMode = "Disabled"
)

type RecommendedContainerResources struct {
	// ... existing fields ...

	// lastPressureTime is when the recommender last accepted a memory pressure sample for this
	// container.
	// +optional
	LastPressureTime *metav1.Time `json:"lastPressureTime,omitempty"`
}
```

The bump is set by recommender flags, `--pressure-bump-up-ratio` (default `1.15`, in `[1, 2]`) and
`--pressure-min-bump-up-bytes` (default 64Mi), as `--oom-bump-up-ratio` and `--oom-min-bump-up-bytes` were
before `PerVPAConfig` added per-VPA overrides; per-VPA tuning can follow the same route. The webhook and
CRD schema validate the mode enum. Recommender and updater flags are validated at startup: bump ranges,
request timeout ≤ sweep timeout < fetch interval ≤ maximum sample age, a positive window and a fraction
in (0, 1].

### Test Plan

Time-dependent tests use an injected clock.

1. `RecordPressure` uses `max(workingSet, memoryPeak)`, writes only `pressurePeak`, replaces the interval
   peak once and never reinflates itself; removing the source enum or the `GetMaxMemoryPeak` inclusion
   fails the test.
2. Pressure never rolls an interval or moves the usage watermark, and is recorded when older than the
   latest usage sample. Gate-off usage and OOM results match
   the pre-feature code, including checkpoint exclusion of the current peak.
3. Detector: both conditions, three samples, grace, cooldown, post-resize reset, stale, duplicate and
   future timestamps, identity changes, zero PSI followed by real pressure, and both counter units.
4. A saturated channel never blocks the recommender; a queued event is dropped once its container leaves
   the target set. Run under the race detector.
5. An accepted sample sets `lastPressureTime` whether or not it moves the target; the stamp is carried
   forward and cleared on opt-out.
6. Updater: a young Pod below its target after a recent sample is updated in place within the window and
   not after it, including a replica that did not produce the sample; no update when resources would not
   change; no update with a resize pending; decreases are clamped; the per-loop cap holds; a failed or
   evicting in-place decision does not evict; `Recreate` mode and gate-off never apply quick pressure; a
   Pod the stock gates admit keeps its eviction fallback.
7. Webhook and CRD parity for the mode enum and the gate-off rules below.
8. Transport: direct TLS verified by node name, explicit proxy and RBAC, 429 and timeout handling,
   oversized responses, and coverage across 5000 fake nodes with slow responders.
9. e2e: a workload stalled near its limit under a VPA with an aged checkpoint is resized in place within
   the window with the gate on, and not with the gate off. The VPA e2e jobs run on GCE clusters built by
   kube-up, and none sets a MemoryQoS throttling factor, so no node writes `memory.high`. Page-cache
   churn alone stalls too briefly on recent kernels to qualify, so the test needs `memory.high`. Alpha
   adds a variant of an existing alpha-beta job that sets `memoryThrottlingFactor`; elsewhere the test
   skips with a clear message when Summary reports no PSI or the node sets no `memory.high`.
   `hack/run-e2e-locally.sh` runs it on kind.

### Feature Enablement and Rollback

- Feature gate: `ReactiveMemoryPressureDetection`, alpha, off by default
- Components: admission-controller, recommender, updater

Enabling starts the observer, allows pressure injection and `lastPressureTime`, lets the updater apply
quick pressure, and lets the admission controller accept `pressureDetection`.

Disabling stops polling and injection, and the updater ignores `lastPressureTime`. The admission
controller rejects new opt-ins but allows a container policy that already opts in to stay, be removed,
or change to Disabled.
Existing peaks are not removed (see [Disabling does not undo a raise](#disabling-does-not-undo-a-raise)).

### Graduation Criteria

Alpha:

- Gate registered, off by default
- Observer, `RecordPressure`, `lastPressureTime` and quick pressure update, with the unit, integration
  and e2e tests above
- Bounded queue, concurrency and responses, fair retries, verified direct TLS and opt-in proxy RBAC
- `pressureDetection` with webhook and schema validation; the metrics above; user documentation

Beta:

- Transport measured on a real cluster of 1000 nodes or more with sparse and dense populations
- Defaults justified by at least two consuming workloads with positive and negative controls, reporting
  known false negatives
- Detection and actuation delays measured separately, including cohort-wide quick updates
- Compatibility matrix for Kubernetes and runtime versions, counter units, resize and swap policies
- No open correctness bugs against the gate for one release

### Version Skew

Upgrade the CRD first; the API server prunes unknown fields. Then upgrade components with the gate off
and enable it once all run. An old updater ignores `lastPressureTime` and applies pressure-influenced
recommendations under the normal gates; an old recommender never sets it. To downgrade, disable the gate
first; keep the newer CRD if stored opt-ins must survive.

### Kubernetes Version Compatibility

Kubernetes 1.36 or later, for tested stats-provider and runtime combinations. `KubeletPSI` is GA and
locked on from 1.36, but that does not guarantee kernel accounting or the counter unit. Per-container PSI
requires cgroup v2. Quick pressure updates need in-place resize (GA since 1.35), and `updateMode:
InPlace` also needs VPA's `InPlace` gate.

## Implementation History

- 2026-09-30: initial AEP, with a proof of concept linked from the pull request

## Alternatives

**Relax the pod-lifetime gate.** `--in-recommendation-bounds-eviction-lifetime-threshold` has been a
global flag since [#3962](https://github.com/kubernetes/autoscaler/pull/3962). Set to `0s`, stock VPA
resized the rollout case slightly faster than quick pressure, but it also resized every young Pod whose
target moved. In a negative control, new Pods spiked during startup and settled without stalling; the
relaxed gate resized them and left them 2.2x over-provisioned, while quick pressure left them alone. PSI
limits the bypass to cohorts with a measured stall.

**Recommendation input only, no updater change.** A raised target inside the recommended range still
waits up to 12 hours, which does not address the issue.

**Rely on OOM samples and existing gates.** Supplies nothing for a container that stalls without being
killed, and a kill is the disruption users such as
[#10174](https://github.com/kubernetes/autoscaler/issues/10174) want to avoid.

**A separate controller that writes Pod resources.** Adds a second writer that can fight VPA's
decisions. Integrating the signal reuses caps, container policies and update modes.

**An alternative recommender or metrics adapter.** A named recommender replaces the stock model rather
than adding an input, and an external metrics adapter adds a deployment without giving the value
pressure semantics. metrics-server does not carry PSI.

**Carry pressure per Pod.** Would limit quick updates to the stalled replica, but puts Pod identities in
VPA status, unbounded for a 5000-node DaemonSet.

**Compare working set with `memory.high`.** Would cover low-factor throttling, but the Summary API does
not expose the effective boundary.
