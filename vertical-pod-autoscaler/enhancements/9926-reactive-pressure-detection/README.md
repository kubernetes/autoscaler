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
  - [Flags](#flags)
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
bounded window after that, the updater applies the current recommendation in place, never lowering a
resource, to any Pod whose memory request is below the target, without waiting for the usual gates. There is no new update mode,
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
- When pressure is observed in a cohort, apply the target in place, within a bounded window, to that
  cohort's Pods whose memory request is below it, instead of waiting for the pod-lifetime gate
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
for the 12-hour gate. Measured on a single node with MemoryQoS enabled, a synthetic workload holding
~503Mi over two replicas, and 8 days of history injected through a checkpoint:

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

The data is per-container memory PSI from the kubelet Summary API, which VPA does not read today. The
recommender fetches it per node, either directly from the kubelet or through the API server proxy.
The default install grants neither permission: direct needs `get nodes/stats`, the proxy needs
`get nodes/proxy`. [Scalability](#scalability) covers node discovery, the transport trade-offs and
the 5000-node measurements.

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

### Flags

Alpha adds eight flags. Every one of them needs the `ReactiveMemoryPressureDetection` gate and is marked
`[ALPHA]`.

| Flag | Component | Default | What it sets |
|---|---|---|---|
| `--pressure-stall-ratio` | recommender | `0.10` | Stall rate a window must exceed, in `[0, 1)` |
| `--pressure-working-set-fraction` | recommender | `0.90` | Working set over limit a window must reach, in `(0, 1]` |
| `--pressure-bump-up-ratio` | recommender | `1.15` | Multiplier for the pressure peak, as `--oom-bump-up-ratio` is for the OOM peak |
| `--pressure-min-bump-up-bytes` | recommender | `64Mi` | Floor for the same raise, as `--oom-min-bump-up-bytes` is for the OOM peak |
| `--pressure-transport` | recommender | `direct` | `direct` or `apiserver-proxy` |
| `--pressure-kubelet-ca-file` | recommender | empty | CA bundle that verifies kubelet serving certificates on the direct transport. Empty keeps the API client CA |
| `--pressure-quick-update-window` | updater | `10m` | How long after an accepted sample a Pod below its target is updated in place |
| `--pressure-quick-update-fraction` | updater | `0.1` | Fraction of a VPA's Pods, at least one, that quick pressure updates per loop |

The two thresholds are flags because the right value depends on the workload. The two bump-up values are
flags because their OOM counterparts are. Everything else this document describes as a tunable is a
constant: the 60s poll interval, the 60s startup grace, the 10-minute cooldown, the 2-minute sample age
limit and 50 concurrent fetches. They stay constants in alpha, and one becomes a flag when a cluster
shows it must.

### Detection Rule

For each eligible container, both conditions must hold over two consecutive windows (three fresh
samples):

1. **Stall rate** above `--pressure-stall-ratio` (default `0.10`).
2. **Working-set fraction** at the end of each window at least `--pressure-working-set-fraction`
   (default `0.90`).

**Choosing thresholds.** The defaults are starting points from one synthetic workload, not calibrated
values. In the rollout case a stalled cohort held a `some` rate near 0.6, and a startup spike that never
stalled held 0. That shows 0.10 separates those two cases, not that it fits every workload, and no
measurement supports 0.90 yet. Both are recommender-wide, so an admin sets them, not a VPA owner. Before
opting a workload in, an admin can read its stall rate as
`rate(container_pressure_memory_waiting_seconds_total[5m])` from the kubelet's `/metrics/cadvisor`
endpoint, which reports the same `some` counter. Too low a ratio raises memory for containers that are
only busy. Too high a ratio leaves `vpa_recommender_pressure_containers` empty while containers stall.
Beta justifies the defaults against two consuming workloads.

A sample that is incomplete, stale or absent drops the count of consecutive windows, because two
windows separated by a gap are not two windows in a row. A gap longer than two poll intervals, a
container restart, a changed limit or a counter that went backwards also starts a new baseline. Neither
shortens a cooldown that is already running. After an accepted sample the container is quiet for 10 minutes, then needs two fresh
windows. While a resize is pending or in progress, including one started outside VPA, nothing is
injected and the container is reported as `resize_in_progress`, `resize_deferred` or
`resize_infeasible`.

With the defaults, the earliest detection is about two minutes after a container starts: a baseline
sample and two 60s windows. The 60s startup grace gates the emission, not the baseline, so it is
already satisfied by then and adds nothing to the floor. Recommender and updater loops each add up to
a minute.

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

**Identity and freshness.** The observer matches Summary entries by Pod UID and container name. A new
container start, a changed limit or a decreasing counter resets the baseline. A sample older than 2
minutes is discarded. A sample ahead of the recommender's clock is discarded only beyond 5 seconds of
tolerance, because node and recommender clocks drift. Inside that tolerance the observer uses the
sample as it is. A node whose samples fall outside these bounds is reported as `stale`. A gap longer
than two poll intervals starts a new baseline. The counter is cumulative, so a shorter gap still
measures the whole window, and the sweep can revisit a node up to two intervals apart.

**Units.** The Summary PSI total is in microseconds on the cadvisor stats path and nanoseconds on the
CRI stats path, on the same kubelet version
([k/k#142599](https://github.com/kubernetes/kubernetes/issues/142599)). Microseconds is canonical and
sig-node's direction is to convert on the CRI path, so the observer reads microseconds and the unit is
not configurable. A flag could not express it anyway, because the unit varies per node, not per cluster,
and a cluster-wide setting of `nanoseconds` would silence detection on every microsecond node. A node on
the other unit computes a rate above 1, which is impossible for a stall rate. The observer discards that
sample, resets the container's baseline and reports the node as `unit_mismatch`. The exact bound is
`1 + 1s/Δt`, because Summary timestamps have one-second resolution. Without the guard, a nanosecond
counter makes every busy container qualify.

**Delivery.** The observer runs outside the recommender loop. It keeps at most one pending observation
per container, and a newer one replaces an older one that was not drained yet.
`LoadRealTimeMetrics` drains them after usage and OOM samples, drops an observation whose container has
left the target set since emission (an opt-out or a deleted Pod), and calls `RecordPressure`. The
pending set is bounded by the target set, so it never drops an observation. This differs from the OOM
observer's 5000-event channel on purpose: a fixed bound would drop events in a cluster-wide burst. The
detector drops stale samples, replaced containers and containers with a resize outstanding before
emitting. Emission starts the cooldown. If `RecordPressure` rejects the observation, the feeder
releases that cooldown, so an injection that never happened does not silence the container for ten
minutes. Gate-off constructs no observer.

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
	sample := ContainerUsageSample{
		MeasureStart: timestamp,
		Usage:        bumpUpMemory(memoryUsed, GetAggregationsConfig().PressureBumpUpRatio, GetAggregationsConfig().PressureMinBumpUp),
		Resource:     ResourceMemory,
	}
	if !container.addMemorySample(&sample, sampleSourcePressure) {
		return errors.New("adding pressure sample failed")
	}
	container.aggregator.RecordPressureObserved(timestamp)
	return nil
}
```

`bumpUpMemory` is the `max(used + minBumpUp, used * ratio)` formula lifted out of `RecordOOM`, so both
sources raise a peak the same way. The pressure path reads the global aggregations config where the OOM
path reads per-container accessors, because `PerVPAConfig` does not cover the pressure values yet.
`RecordPressureObserved` is what stamps `lastPressureTime` (see **Quick Pressure Update**).

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

The recommender does this in every update mode, so a mode changes actuation, not detection:

| `updateMode` | Pressure peak in the recommendation | Quick pressure update |
|---|---|---|
| `InPlace`, `InPlaceOrRecreate` | yes | yes |
| `Recreate`, `Auto` (deprecated) | yes | no. The raised target applies under the normal gates, which evict. |
| `Initial` | yes | no. The raised target applies to new Pods. |
| `Off` | yes, in `status.recommendation` only | no. Nothing is applied. |

The gap is deliberate. Quick pressure never evicts (see **Why never evict** below), so in an evicting
mode it has nothing to do that the normal gates do not already do. A user does not have to infer this:
when a container opts in under any other mode, the admission controller returns a warning that names the
mode and says that only `InPlace` and `InPlaceOrRecreate` apply the raise to running Pods without
waiting for the usual thresholds. It warns rather than rejects, because the
recommendation change on its own is useful, and because the user may switch mode later.

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
| Trigger lasts | while the OOM ended within `--evict-after-oom-threshold` of now ([#10325](https://github.com/kubernetes/autoscaler/pull/10325)) | a window after the last accepted sample |
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
informer for addresses. That informer is cluster-wide, so it caches every Node object even when
`--vpa-object-namespace` restricts the other informers. It is the recommender's first Node watch, and
the direct transport is the only thing that needs it: the proxy transport addresses nodes by name and
starts no Node informer, and with the gate off neither exists. Only nodes with eligible containers are
queried, and with no opt-ins there are no fetches. This does not depend on VPASlice
([#10229](https://github.com/kubernetes/autoscaler/pull/10229)).

**Building the target set.** The recommender rebuilds the target set each loop by walking the Pods it
already tracks once, and it evaluates a VPA selector only for a Pod in a namespace that holds an
opted-in VPA. Asking the cluster state for each VPA's matching Pods instead costs one scan of every Pod
per opted-in VPA. With no opt-ins the walk stops at the first check.

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

Both paths ship, because neither covers every cluster: direct needs `serverTLSBootstrap`, and the proxy
is the answer for kubeadm-style per-node self-signed certificates. Most of the code is in direct: it
copies the REST config with the kubelet CA, dials the node's InternalIP while verifying the certificate
against the node name, refuses redirects, reads the kubelet port from `DaemonEndpoints`, and handles
status codes and `Retry-After` itself. The proxy is one request on the API client the recommender
already has. Direct is still the default, because `get nodes/proxy` also permits command execution in
any container on any node, and every proxied fetch also goes through the API server.

**Scheduling.** One fetch per node per 60s poll interval, at most 50 in flight, 5s per request, 45s per
sweep, 16MiB per response. Failing nodes back off exponentially with jitter up to 5 minutes, honouring `Retry-After`.
The observer runs outside the recommender loop.

**Coverage.** Detection needs two windows, so a node revisited less often than every two poll intervals
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
| Observation windows ([AEP-9936](../9936-vpa-observation-window/README.md)) | Collection continues, so a window can accumulate pressure peaks. A `lastPressureTime` older than the quick window when actuation reopens gives no bypass. |
| Per-VPA percentiles ([AEP-9970](../9970-per-vpa-target-percentiles/README.md)) | A low memory percentile can absorb a single pressure peak, as it can an OOM peak. |
| Request/limit ratio ([AEP-8515](https://github.com/kubernetes/autoscaler/pull/8516)) | Only request recommendations change; limit derivation stays with existing policy. `RequestsOnly` can cap the usable request at an unchanged limit. |
| Resize preemption ([KEP-5836](https://github.com/kubernetes/enhancements/issues/5836)) | Honoured as configured; never enabled by this feature. |
| Swap | `disallowResizeForSwappableContainers` still rejects restart-free memory resizes of swappable containers. |
| Guaranteed QoS | MemoryQoS sets no container `memory.high` when request equals limit, so a Guaranteed container stalls only in limit reclaim shortly before an OOM kill. The detector does not need `memory.high`, but has not been measured on Guaranteed Pods. An [opt-in `memory.high` for Guaranteed](https://github.com/kubernetes/enhancements/pull/6141#issuecomment-4908522670) would give it a throttle window to act in. |
| Proactive reclaim | Out of scope ([enhancements#6398](https://github.com/kubernetes/enhancements/issues/6398) was closed leaving it to node agents). Without swap, `memory.reclaim` cannot lower the working set. |
| MemoryQoS | A VPA resize re-runs the kubelet's `memory.high` formula, so VPA should be the only actor on a container's boundary. Three kubelet gaps limit the result, measured on v1.37.0: the Summary API does not expose effective `memory.high`; after a resize that made request equal limit the container kept its old `memory.high` ([k/k#142569](https://github.com/kubernetes/kubernetes/issues/142569)); and under `memoryReservationPolicy: TieredReservation` a request-only resize leaves pod-level `memory.low` at the old request ([k/k#142572](https://github.com/kubernetes/kubernetes/issues/142572)). |

### Observability

| Metric | Type | Labels | Description |
|---|---|---|---|
| `vpa_recommender_pressure_events_total` | Counter | `result` | Detector: `stale`, `identity_changed`, `rate_out_of_range`, `resize_pending`. Feeder: `accepted`, `policy_changed`, `window_mismatch`, `unknown_container` |
| `vpa_recommender_pressure_containers` | Gauge | `state` | Containers meeting both conditions in the latest window: `injectable`, `cooldown`, `resize_in_progress`, `resize_deferred`, `resize_infeasible` |
| `vpa_recommender_pressure_estimate_to_recommendation_ratio` | Histogram | none | Accepted pressure estimate divided by the last memory recommendation |
| `vpa_recommender_pressure_fetch_duration_seconds` | Histogram | `result` | Per-node fetch latency: `success`, `timeout`, `throttled`, `error` |
| `vpa_recommender_pressure_nodes` | Gauge | `state` | Nodes by last fetch: `observed`, `missing_psi`, `failed`, `stale`, `undersampled`, `unit_mismatch` |
| `vpa_recommender_pressure_node_revisit_seconds` | Histogram | none | Time between consecutive successful fetches of a node |
| `vpa_updater_pressure_quick_updates_total` | Counter | `result` | `admitted`, `capped`, `no_change`, `eviction_skipped`, `decrease_blocked` when caps would still lower a resource, and `decrease_clamped` per resource held at its request |

`vpa_updater_pressure_quick_updates_total` does not partition the Pods it sees. One Pod can add to
`decrease_clamped` once per resource and then to `admitted`, so the results sum to more than the number
of Pods. Read `admitted` on its own, and the rest as reasons rather than as shares of a whole.

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
CRD schema validate the mode enum, and the webhook warns about an opt-in under a mode that cannot act on
the signal right away. Each component validates its flags at startup: bump ratio in `[1, 2]`, minimum
bump not negative, stall ratio in `[0, 1)`, working-set fraction in `(0, 1]`, a known transport, a
positive quick-update window and a quick-update fraction in `(0, 1]`. The two thresholds exclude the
value that would turn the condition off silently: a stall rate cannot reach 1, and a fraction of 0
passes every container.

Flag validation does not read the gate. A bad pressure flag stops the recommender even with the feature
off, which is the same treatment the OOM and Prometheus flags already have. The two bump-up values also
reach the aggregations config either way, where nothing reads them until a pressure sample arrives.

### Test Plan

Time-dependent tests use an injected clock.

1. `RecordPressure` uses `max(workingSet, memoryPeak)`, writes only `pressurePeak`, replaces the interval
   peak once and never reinflates itself; removing the source enum or the `GetMaxMemoryPeak` inclusion
   fails the test.
2. Pressure never rolls an interval or moves the usage watermark, and is recorded when older than the
   latest usage sample. Gate-off usage and OOM results match
   the pre-feature code, including checkpoint exclusion of the current peak.
3. Detector: both conditions, three samples, grace, cooldown, post-resize reset, stale, duplicate and
   future timestamps, identity changes, zero PSI followed by real pressure, and a nanosecond counter,
   which is discarded and reports the node as `unit_mismatch`.
4. Two emissions before a drain leave one observation, the later one. An observation is dropped once
   its container leaves the target set, and a rejected `RecordPressure` releases the cooldown. Run under
   the race detector.
5. An accepted sample sets `lastPressureTime` whether or not it moves the target; the stamp is carried
   forward and cleared on opt-out.
6. Updater: a young Pod below its target after a recent sample is updated in place within the window and
   not after it, including a replica that did not produce the sample; no update when resources would not
   change; no update with a resize pending; decreases are clamped; the per-loop cap holds; a failed or
   evicting in-place decision does not evict; `Recreate` mode and gate-off never apply quick pressure; a
   Pod the stock gates admit keeps its eviction fallback.
7. Webhook and CRD parity for the mode enum and the gate-off rules below, and the update-mode warning:
   one per opted-in container under a non-in-place mode, none under `InPlace` or `InPlaceOrRecreate`,
   and none for an opt-in the webhook already rejects.
8. Transport: direct TLS verified by node name, explicit proxy and RBAC, 429 and timeout handling,
   oversized responses, and coverage across fake nodes with slow responders. The 5000-node figure in
   [Scalability](#scalability) needs a harness the repository does not have: the KWOK benchmark under
   `test/benchmark/` creates one fake node and serves no Summary API. Alpha adds the harness with the
   feature.
9. e2e: a workload stalled near its limit under a VPA with an aged checkpoint is resized in place within
   the window with the gate on, and not with the gate off. The VPA e2e jobs run on GCE clusters built by
   kube-up, and none sets a MemoryQoS throttling factor, so no node writes `memory.high`. Page-cache
   churn alone stalls too briefly on recent kernels to qualify, so the test needs `memory.high`. Alpha
   adds a variant of an existing alpha-beta job that sets `memoryThrottlingFactor`; elsewhere the test
   skips with a clear message when Summary reports no PSI or the node sets no `memory.high`.

### Feature Enablement and Rollback

- Feature gate: `ReactiveMemoryPressureDetection`, alpha, off by default
- Components: admission-controller, recommender, updater

Enabling starts the observer, allows pressure injection and `lastPressureTime`, lets the updater apply
quick pressure, and lets the admission controller accept `pressureDetection`.

Enabling it on the direct transport can stop the recommender from starting. The transport is built
once, during startup, and a failure there fails recommender construction rather than falling back to
the proxy, because a silent fallback would widen RBAC from `get nodes/stats` to `get nodes/proxy`
without the operator asking. An operator who cannot supply kubelet serving certificates from a trusted
CA sets `--pressure-transport=apiserver-proxy` or leaves the gate off. Fetch failures after startup are
not fatal: the node backs off and is reported as `failed`.

Disabling stops polling and injection, and the updater ignores `lastPressureTime`. The admission
controller rejects new opt-ins but allows a container policy that already opts in to stay, be removed,
or change to Disabled.
Existing peaks are not removed (see [Disabling does not undo a raise](#disabling-does-not-undo-a-raise)).

### Graduation Criteria

Alpha:

- Gate registered, off by default
- Observer, `RecordPressure`, `lastPressureTime` and quick pressure update, with the unit, integration
  and e2e tests above
- Bounded concurrency and response size, with sweep rotation and carry-over, so the next sweep fetches
  a node cut by the sweep deadline first
- Verified direct TLS and opt-in proxy RBAC
- `pressureDetection` with webhook and schema validation, including the update-mode warning
- The eight flags, the metrics above and user documentation

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

**Recommendation input only, no updater change.** This is what every mode other than `InPlace` and
`InPlaceOrRecreate` already gets, and it is not enough for Story 1: a raised target inside the
recommended range still waits up to 12 hours for the pod-lifetime gate. Shipping only the input would
leave the motivating case unfixed.

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
