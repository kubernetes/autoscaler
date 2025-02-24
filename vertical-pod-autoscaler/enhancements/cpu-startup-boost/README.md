# AEP-: CPU Startup Boost

<!-- toc -->
- [AEP-: CPU Startup Boost](#aep--cpu-startup-boost)
  - [Summary](#summary)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
  - [Design Details](#design-details)
    - [Workflow](#workflow)
    - [API Changes](#api-changes)
      - [Priority of `StartupBoost`](#priority-of-startupboost)
      - [CPU startup boost and VPA autoscaling modes](#cpu-startup-boost-and-vpa-autoscaling-modes)
    - [Validation](#validation)
    - [Startup Boost and Readiness/Startup Probes](#startup-boost-and-readinessstartup-probes)
    - [Failed in-place downsizes](#failed-in-place-downsizes)
  - [Test Plan](#test-plan)
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

* Allow VPA to boost the CPU request and limit of a pod's containers when the pod during the pod startup (from creation time until it becomes `Ready`.)
* Allow VPA to scale pods down to the original CPU resource values, [in-place](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources), as soon as their [`Ready`](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-conditions) condition is true.

### Non-Goals

* Allow VPA to boost CPU resources of pods that are already [`Ready`](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-conditions).
* Allow VPA to boost CPU resources during startup of workloads that have not configured a [Readiness or a Startup probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/).
* Allow VPA to boost memory resources.

## Proposal

To extend [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L190) with a `StartupBoost.CPU.Factor` field, that will instruct VPA the factor by which to multiply the initial resource request and limit of the containers' targeted by the VPA object.

This AEP also proposes to extend [`ContainerScalingMode`](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L230-L235) with a new `StartupBoostOnly` mode to allow users to only enable the startup boost feature and not vanilla VPA altogether.

## Design Details

### Workflow

1.  The user first configures the CPU startup boost factor on their VPA object

1. When a pod targeted by that VPA is created, the kube-apiserver invokes the
VPA Admission Controller

1. The VPA Admission Controller modifies the pod's containers CPU request and
limits to align with its `boostPolicy`, if specified, during the
pod creation.
      * VPA Admission Controller will not modify the CPU resources of a pod if
        it does not configure a Readiness or a Startup probe.

1. The VPA Updater monitors pods targeted by the VPA object and when the pod
condition is `Ready`, it scales down the CPU resources to the appropriate
non-boosted value: `max(existing VPA recommendation for that container, minimum configured value)`.
    * The scale down is applied [in-place](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources).

### API Changes

To extend [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L190) with a `StartupBoost.CPU.Factor` field, that will instruct VPA the factor by which to multiply the initial resource request and limit of the containers' targeted by the VPA object.

This AEP also proposes to extend [`ContainerScalingMode`](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L230-L235) with a new `StartupBoostOnly` mode to allow users to only enable the startup boost feature and not vanilla VPA altogether.


#### Priority of `StartupBoost`

The new `StartupBoost` field will take precedence over the rest of the container resource policy configurations. Functioning independently from all other fields in [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L190), except for [`ContainerName`](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L191-L194) and [`Mode`](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L195-L197). This means that a container's CPU request/limit can be boosted during startup beyond [`MaxAllowed`](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L202-L205), for example, or it will be able to be boosted even if
CPU is  explicitly excluded from [`ControlledResources`](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L207-L211).

#### CPU startup boost and VPA autoscaling modes

Initially, CPU startup boost will only be able to be used if the container or VPA scaling mode is [`InPlaceOrRecreate`](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/enhancements/4016-in-place-updates-support/README.md?plain=1#L61), since it does not make sense to use this feature with disruptive modes of VPA

Here's an example of the VPA CRD, incorporating CPU boosting (only shows relevant sections):

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
    updateMode: "InPlaceOrRecreate"
  resourcePolicy:
    containerPolicies:
      - containerName: 'boosted-container'
        mode: 'StartupBoostOnly'
        startupBoost:
          cpu:
            factor: 2.0
```

### Validation

* We'll also validate that the `startupBoost` configuration is valid when VPA objects are created/updated:
   * The VPA autoscaling mode must be `InPlaceOrRecreate`
   * The boost factor is >= 1

### Startup Boost and Readiness/Startup Probes

The CPU Startup Boost functionality will boost the CPU requests of containers
until the pod has a `Ready` condition. Therefore, in this version of the AEP,
workloads must be configured with a Readiness or a Startup probe to be able to
utilize this feature.

### Failed in-place downsizes

The VPA Updater **will not** evict a pod to actuate a startup CPU boost recommendation if it attempted to apply the recommendation in place and it failed (see the [scenarios](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/enhancements/4016-in-place-updates-support/README.md?plain=1#L164-L169 ) where the VPA updater will consider that the update failed). This is to avoid an eviction loop:

1. A pod is created and has its CPU resources boosted
1. The pod is ready. VPA Updater tries to downscale the pod in-place and it fails.
1. VPA Updater evicts the pod. Logic flow goes back to (1).

## Test Plan

* Add tests to verify the VPA fields validation

We will add the following scenarios to our e2e tests:

* CPU Startup Boost recommendation is applied to pod controlled by VPA until it becomes `Ready`. Then, the pod is scaled back down in-place.
  * Boost is applied to all containers of a pod.
  * Boost is applied to a subset of containers.
* CPU Startup Boost will not be applied if a pod is not configured with a Readiness or a Startup probe.
* Pod should not be evicted if the in-place update fails when scaling the pod back down.


## Implementation History

* 2025-02-24: Initial version.

