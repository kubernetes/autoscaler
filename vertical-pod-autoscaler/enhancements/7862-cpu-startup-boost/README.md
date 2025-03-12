# AEP-7862: CPU Startup Boost

<!-- toc -->
- [AEP-7862: CPU Startup Boost](#aep-7862-cpu-startup-boost)
  - [Summary](#summary)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
  - [Design Details](#design-details)
    - [Workflow](#workflow)
    - [API Changes](#api-changes)
      - [Priority of `StartupBoost`](#priority-of-startupboost)
    - [Validation](#validation)
      - [Static Validation](#static-validation)
      - [Dynamic Validation](#dynamic-validation)
    - [Feature Enablement](#feature-enablement)
    - [Mitigating Failed In-Place Downsizes](#mitigating-failed-in-place-downsizes)
  - [Test Plan](#test-plan)
  - [Examples](#examples)
    - [CPU Boost Only](#cpu-boost-only)
    - [CPU Boost and Vanilla VPA](#cpu-boost-and-vanilla-vpa)
  - [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Long application start time is a known problem for more traditional workloads
running in containerized applications, especially Java workloads. This delay can
negatively impact the user experience and overall application performance. One
potential solution is to provide additional CPU resources to pods during their
startup phase, but this can lead to waste if the extra CPU resources are not
set back to their original values after the pods are ready.

This proposal allows VPA to boost the CPU request and limit of containers during
the pod startup and to scale the CPU resources back down when the pod is `Ready`,
leveraging the [in-place pod resize Kubernetes feature](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources).


### Goals

* Allow VPA to boost the CPU request and limit of a pod's containers during the
pod startup (from creation time until it becomes `Ready`).
* Allow VPA to scale pods down [in-place](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources)
to the existing VPA recommendation for that container, if any, or to the CPU
resources configured in the pod spec, as soon as their [`Ready`](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-conditions)
condition is true.

### Non-Goals

* Allow VPA to boost CPU resources of pods that are already [`Ready`](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-conditions).
* Allow VPA to boost CPU resources during startup of workloads that have not
configured a [Readiness or a Startup probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/).
* Allow VPA to boost memory resources.
  * This is out of scope for now because the in-place pod resize feature
  [does not support memory limit decrease yet.](https://github.com/kubernetes/enhancements/tree/758ea034908515a934af09d03a927b24186af04c/keps/sig-node/1287-in-place-update-pod-resources#memory-limit-decreases)

## Proposal

* To extend [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L191)
with a new `StartupBoost` field to allow users to configure the CPU startup
boost.

* To extend [`ContainerScalingMode`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L231-L236)
with a new `StartupBoostOnly` mode to allow users to only enable the startup
boost feature and not vanilla VPA altogether.

## Design Details

### Workflow

1.  The user first configures the CPU startup boost on their VPA object

1. When a pod targeted by that VPA is created, the kube-apiserver invokes the
VPA Admission Controller

1. The VPA Admission Controller modifies the pod's containers CPU request and
limits to align with its `StartupBoost` policy, if specified, during the pod
creation.

1. The VPA Updater monitors pods targeted by the VPA object and when the pod
condition is `Ready`, it scales down the CPU resources to the appropriate
non-boosted value: `existing VPA recommendation for that container` (if any) OR
the `CPU resources configured in the pod spec`.
    * The scale down is applied [in-place](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources).

### API Changes

The new `StartupBoost` parameter will be added to the [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L191)
and contain the following fields:
  * `StartupBoost.CPU.Factor`: the factor by which to multiply the initial
  resource request and limit of the containers' targeted by the VPA object.
  * `StartupBoost.CPU.Value`: the target value of the CPU request or limit
  during the startup boost phase.

NOTE: The boosted CPU value will be capped by
[`--container-recommendation-max-allowed-cpu`](https://github.com/kubernetes/autoscaler/blob/4d294562e505431d518a81e8833accc0ec99c9b8/vertical-pod-autoscaler/pkg/recommender/main.go#L122)
flag value, if set.

NOTE: Only one of `Factor` or `Value` may be specified per container policy.

#### Priority of `StartupBoost`

The new `StartupBoost` field will take precedence over the rest of the container
resource policy configurations. Functioning independently from all other fields
in [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L191),
**except for**:
   * [`ContainerName`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L192-L195)
   * [`Mode`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L196-L198)
   * [`ControlledValues`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L214-L217)

This means that a container's CPU request/limit can be boosted during startup
beyond [`MaxAllowed`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L203-L206),
for example, or it will be able to be boosted even if CPU is explicitly
excluded from [`ControlledResources`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L208-L212).

### Validation

#### Static Validation

* We will check that the `startupBoost` configuration is valid when VPA objects
are created/updated:
   * The VPA autoscaling mode must be `InPlaceOrRecreate` (since it does not
   make sense to use this feature with disruptive modes of VPA).
   * The boost factor is >= 1
   * Only one of `StartupBoost.CPU.Factor` or `StartupBoost.CPU.Value` is
   specified

#### Dynamic Validation

* `StartupBoost.CPU.Value` must be greater than the CPU request or limit of the
  container during the boost phase, otherwise we risk downscaling the container.

* Workloads must be configured with a Readiness or a Startup probe to be able to
utilize this feature. Therefore, VPA will not boost CPU resources of workloads
that do not configure a Readiness or a Startup probe.

### Feature Enablement

During the Alpha launch of this feature, users will need to:
* Enable the CPU startup boost feature via a binary flag.
* Set the VPA autoscaling mode to `InPlaceOrRecreate`
* Set the [`ContainerScalingMode`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L231-L236)
to either:
  * **NEW**: `StartupBoostOnly`: this mode will allow users to only enable the
  startup boost feature for a container and not vanilla VPA altogether.
  * **NEW**: `Auto`: in this mode, both vanilla VPA and CPU Startup Boost will
  be enabled.


### Mitigating Failed In-Place Downsizes

The VPA Updater will evict a pod to actuate a startup CPU boost
recommendation if it attempted to apply the recommendation in place and it
failed (see the [scenarios](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/enhancements/4016-in-place-updates-support/README.md?plain=1#L164-L169)
where the VPA updater will consider that the update failed).

We will cache pods that have failed the in-place downsize and not reattempt to
CPU boost them again for the next hour, to **avoid** an eviction loop scenario
like this:

1. A pod is created and has its CPU resources boosted
2. The pod is ready. VPA Updater tries to downscale the pod in-place and it
fails.
1. VPA Updater evicts the pod. (If we do nothing, the logic flow goes back to
(1)).

## Test Plan

Other than comprehensive unit tests, we will also add the following scenarios to
our e2e tests:

* CPU Startup Boost recommendation is applied to pod controlled by VPA until it
becomes `Ready`. Then, the pod is scaled back down in-place.
  * Boost is applied to all containers of a pod.
  * Boost is applied to a subset of containers.
* CPU Startup Boost will not be applied if a pod is not configured with a
Readiness or a Startup probe.
* Pod is evicted the first time that an in-place update fails when scaling the
pod back down. And a new CPU boost is not attempted when the pod is recreated.


## Examples

Here are some examples of the VPA CR incorporating CPU boosting for different
scenarios.

### CPU Boost Only

```yaml
apiVersion: "autoscaling.k8s.io/v1"
kind: VerticalPodAutoscaler
metadata:
  name: example-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: example
  updatePolicy:
    # VPA Update mode must be InPlaceOrRecreate
    updateMode: "InPlaceOrRecreate"
  resourcePolicy:
    containerPolicies:
      - containerName: "boosted-container-name"
        mode: "StartupBoostOnly"
        startupBoost:
          cpu:
            factor: 2.0
```

### CPU Boost and Vanilla VPA

```yaml
apiVersion: "autoscaling.k8s.io/v1"
kind: VerticalPodAutoscaler
metadata:
  name: example-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: example
  updatePolicy:
    # VPA Update mode must be InPlaceOrRecreate
    updateMode: "InPlaceOrRecreate"
  resourcePolicy:
    containerPolicies:
      - containerName: "boosted-container-name"
        mode: "Auto" # Vanilla VPA mode + Startup Boost
        minAllowed:
          cpu: "250m"
          memory: "100Mi"
        maxAllowed:
          cpu: "500m"
          memory: "600Mi"
        # The CPU boosted resources can go beyond maxAllowed.
        startupBoost:
          cpu:
            value: 4
```

## Implementation History

* 2025-03-12: Initial version.

