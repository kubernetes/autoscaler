# AEP-7862: CPU Startup Boost

<!-- TOC -->

- [AEP-7862: CPU Startup Boost](#aep-7862-cpu-startup-boost)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
  - [Design Details](#design-details)
    - [Workflow](#workflow)
    - [API Changes](#api-changes)
      - [Priority of StartupBoost](#priority-of-startupboost)
    - [Validation](#validation)
      - [Static Validation](#static-validation)
      - [Dynamic Validation](#dynamic-validation)
    - [Mitigating Failed In-Place Downsizes](#mitigating-failed-in-place-downsizes)
    - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
      - [How can this feature be enabled / disabled in a live cluster?](#how-can-this-feature-be-enabled--disabled-in-a-live-cluster)
    - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
  - [Test Plan](#test-plan)
  - [Examples](#examples)
    - [Per-pod configurations startupBoost configured in VerticalPodAutoscalerSpec](#per-pod-configurations-startupboost-configured-in-verticalpodautoscalerspec)
      - [Startup CPU Boost Enabled & VPA Disabled](#startup-cpu-boost-enabled--vpa-disabled)
      - [Startup CPU Boost Disabled & VPA Enabled](#startup-cpu-boost-disabled--vpa-enabled)
      - [Startup CPU Boost Enabled & VPA Enabled](#startup-cpu-boost-enabled--vpa-enabled)
    - [Per-container configurations startupBoost configured in ContainerPolicies](#per-container-configurations-startupboost-configured-in-containerpolicies)
      - [Startup CPU Boost Enabled & VPA Disabled](#startup-cpu-boost-enabled--vpa-disabled)
      - [Startup CPU Boost Disabled & VPA Enabled](#startup-cpu-boost-disabled--vpa-enabled)
      - [Startup CPU Boost Enabled & VPA Enabled](#startup-cpu-boost-enabled--vpa-enabled)
  - [Implementation History](#implementation-history)

<!-- /TOC -->

Long application start time is a known problem for more traditional workloads
running in containerized applications, especially Java workloads. This delay can
negatively impact the user experience and overall application performance. One
potential solution is to provide additional CPU resources to pods during their
startup phase, but this can lead to waste if the extra CPU resources are not
set back to their original values after the pods have started up.

This proposal allows VPA to boost the CPU request and limit of containers during
the pod startup and to scale the CPU resources back down when the pod is
`Ready` or after certain time has elapsed, leveraging the
[in-place pod resize Kubernetes feature](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources).

### Goals

* Allow VPA to boost the CPU request and limit of a pod's containers during the
pod (re-)creation time.
* Allow VPA to scale pods down [in-place](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources)
to the existing VPA recommendation for that container, if any, or to the CPU
resources configured in the pod spec, as soon as their [`Ready`](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-conditions)
condition is true and `StartupBoost.CPU.Duration` has elapsed.

### Non-Goals

* Allow VPA to boost CPU resources of pods outside of the pod (re-)creation
time.
* Allow VPA to boost memory resources.
  * This is out of scope for now because the in-place pod resize feature
  [does not support memory limit decrease yet.](https://github.com/kubernetes/enhancements/tree/758ea034908515a934af09d03a927b24186af04c/keps/sig-node/1287-in-place-update-pod-resources#memory-limit-decreases)

## Proposal

* To extend [`VerticalPodAutoscalerSpec`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L75)
with a new `StartupBoost` field to allow users to configure the CPU startup
boost.

* To extend [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L191)
with a new `StartupBoost` field to allow users to optionally customize the
startup boost behavior for individual containers.

* To enable only startup boost (if the `StartupBoost` config is present in the
VPA object) without having to ALSO use the traditional VPA functionality.

## Design Details

### Workflow

1.  The user first configures the CPU startup boost on their VPA object

1. When a pod targeted by that VPA is created, the kube-apiserver invokes the
VPA Admission Controller

1. The VPA Admission Controller modifies the pod's containers CPU request and
limits to align with its `StartupBoost` policy, if specified, during the pod
creation. The boosted value is based on the VPA recommendation available at the
time of admission. During the boost period, no resizing will take place.

1. The VPA Updater monitors pods targeted by the VPA object and when the pod
condition is `Ready` and `StartupBoost.CPU.Duration` has elapsed, it scales
down the CPU resources to the appropriate non-boosted value. This "unboosting"
resizes the pod to whatever the recommendation is at that moment. The specific
behavior is determined by the VPA `updatePolicy`:
    * If `updatePolicy` is `Auto`, `Recreate` or `InPlaceOrRecreate`, the VPA
    Updater will apply the current VPA recommendation, even if it's higher than
    the boosted value.
    * If `updatePolicy` is `Off` for the VPA object, or `mode` is `Off` in a
    container policy, the VPA Updater will revert the CPU resources to the
    values specified in the pod spec.
    * The scale down is applied [in-place](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources).

### API Changes

The new `StartupBoost` parameter will be added to both:
   * [`VerticalPodAutoscalerSpec`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L75):
   Will allow users to specify the default CPU startup boost for all containers of the pod targeted by the VPA object.
   * [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L191):
   Will allow users to optionally customize the startup boost behavior for individual containers.

Here is the Go struct definition for `CPUStartupBoost`:

```go
// +enum
type CPUBoostType string

const (
    FactorType   CPUBoostType = "Factor"
    QuantityType CPUBoostType = "Quantity"
)

type CPUStartupBoost struct {
    // +unionDiscriminator
    // +required
    Type CPUBoostType `json:"type"`

    // +unionMember=Factor
    // +optional
    Factor *int32 `json:"factor,omitempty"`

    // +unionMember=Quantity
    // +optional
    Quantity *resource.Quantity `json:"quantity,omitempty"`

    // +optional
    Duration *metav1.Duration `json:"duration,omitempty"`
}
```

`StartupBoost` will contain the following fields:
  * [Required] `StartupBoost.CPU.Type` (type: `string`): A string that specifies
  the kind of boost to apply. Supported values are:
    * `Factor`: The `StartupBoost.CPU.Factor` field will be interpreted as a
    multiplier for the recommended CPU request. For example, a value of `2` will
    double the CPU request.
    * `Quantity`: The `StartupBoost.CPU.Quantity` field will be interpreted as an
    additional CPU resource quantity (e.g., `"500m"`, `"1"`) to be added to the existing CPU
    request or limit during the boost phase.

  * [Optional] `StartupBoost.CPU.Factor`: (type: `integer`): The factor to apply to the CPU request. Defaults to 1 if not specified.
     * If `StartupBoost.CPU.Type`is `Factor`, this field is required.
     * If `StartupBoost.CPU.Type`is `Quantity`, this field is not allowed.
  * [Optional] `StartupBoost.CPU.Quantity`: (type: `resource.Quantity`): The additional CPU resource quantity.
     * If `StartupBoost.CPU.Type`is `Quantity`, this field is required.
     * If `StartupBoost.CPU.Type`is `Factor`, this field is not allowed.
  * [Optional] `StartupBoost.CPU.Duration` (type: `duration`): if specified, it
  indicates for how long to keep the pod boosted **after** it goes to `Ready`.
     * It defaults to `0s` if not specified.

> [!IMPORTANT]
> The boosted CPU value will be capped by
> [`--max-allowed-cpu-boost`](https://github.com/kubernetes/autoscaler/blob/4b40a55bebd2ce184b289cd028969182d15f412c/vertical-pod-autoscaler/pkg/admission-controller/main.go#L86C1-L86C2)
> flag value, if set.

> [!NOTE]
> To ensure that containers are unboosted only after their applications are
> started and ready, it is recommended to configure a
> [Readiness or a Startup probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
> for the containers that will be CPU boosted. Check the [Test Plan](#test-plan)
> section for more details on this feature's behavior for different combinations
> of probers + `StartupBoost.CPU.Duration`.

#### Priority of `StartupBoost`

The new `StartupBoost` field will take precedence over the rest of the fields
in  [`VerticalPodAutoscalerSpec`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L75)
and [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L191),
**except for**:
   * [`VerticalPodAutoscalerSpec.TargetRef`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L88)
   * [`ContainerResourcePolicy.ContainerName`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L192-L195)
   * [`ContainerResourcePolicy.ControlledValues`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L214-L217)

This means that a container's CPU request/limit can be boosted during startup
beyond [`MaxAllowed`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L203-L206),
for example, or it will be able to be boosted even if CPU is explicitly
excluded from [`ControlledResources`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.3.0/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L208-L212).

### Validation

#### Static Validation

* We will check that the `startupBoost` configuration is valid when VPA objects
are created/updated:
   * The boost factor value is >= 1 (via CRD validation rules)
   * The [feature enablement](#feature-enablement-and-rollback) flags must be
   on.


#### Dynamic Validation

* The boosted CPU value must be greater than the CPU request or limit of the
  container during the boost phase, otherwise we risk downscaling the container.
* If `Type` is set to `Factor` and `Value` is set to a value that can't be
  parsed as a `float64` (e.g., `500m`), the API must reject the `startupBoost`
  configuration as invalid.

### Mitigating Failed In-Place Downsizes

The VPA Updater **will not** evict a pod if it attempted to scaled the pod down
in place (to unboost its CPU resources) and the update failed (see the
[scenarios](https://github.com/kubernetes/autoscaler/blob/0a34bf5d3a71b486bdaa440f1af7f8d50dc8e391/vertical-pod-autoscaler/enhancements/4016-in-place-updates-support/README.md?plain=1#L164-L169) where the VPA
updater will consider that the update failed). This is to avoid an eviction
loop:

1. A pod is created and has its CPU resources boosted
1. The pod meets the conditions to be unboosted. VPA Updater tries to downscale
the pod in-place and it fails.
1. VPA Updater evicts the pod. Logic flow goes back to (1).

### Feature Enablement and Rollback

#### How can this feature be enabled / disabled in a live cluster?

* Feature gates name: `CPUStartupBoost`
* Components depending on the feature gates:
  * admission-controller
  * updater

Enabling of feature gates `CPUStartupBoost` will cause the following to happen:
  * admission-controller to **accept** new VPA objects being created with
`StartupBoost` configured.
  * admission-controller to **boost** CPU resources.
  * updater to **unboost** the CPU resources.

Disabling of feature gates `CPUStartupBoost` will cause the following to happen:
  * admission-controller to **reject** new VPA objects being created with
  `StartupBoost` configured.
    * A descriptive error message should be returned to the user letting them
    know that they are using a feature gated feature.
  * admission-controller **to not** boost CPU resources, should it encounter a
  VPA configured with a `StartupBoost` config.
  * updater **to not** unboost CPU resources when pods meet the scale down
  requirements, should it encounter a VPA configured with a `StartupBoost`
  config.

### Kubernetes Version Compatibility

Similarly to [AEP-4016](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/enhancements/4016-in-place-updates-support#kubernetes-version-compatibility),
`StartupBoost` configuration is built assuming that VPA will be running on a
Kubernetes 1.33+ with the beta version of
[KEP-1287: In-Place Update of Pod Resources](https://github.com/kubernetes/enhancements/issues/1287)
enabled. If this is not the case, VPA's attempt to unboost pods may fail and the
pods may remain boosted for their whole lifecycle.

## Test Plan

Other than comprehensive unit tests, we will also add the following scenarios to
our e2e tests:

* CPU Startup Boost recommendation is applied to pod controlled by VPA until it
becomes `Ready` and `StartupBoost.CPU.Duration` has elapsed. Then, the pod is
scaled back down in-place. We'll also test the following sub-cases:
  * Boost is applied to all containers of a pod.
  * Boost is applied only to a subset of containers in a pod.
  * Combinations of probes + `StartupBoost.CPU.Duration`:
    * No probes and no `StartupBoost.CPU.Duration` specified: unboost will
    likely happen immediately.
    * No probes and a 60s `StartupBoost.CPU.Duration`: unboost will likely
    happen after 60s.
    * A readiness/startup probe and no `StartupBoost.CPU.Duration` specified:
    unboost will likely as soon as the pod becomes `Ready`.
    *  A readiness/startup probe and a 60s `StartupBoost.CPU.Duration`
    specified: unboost will likely happen 60s **after** the pod becomes `Ready`.

* Pod is not evicted if the in-place update fails when scaling the pod back
down.

## Examples

Here are some examples of the VPA CR incorporating CPU boosting for different
scenarios.

### Per-pod configurations (`startupBoost` configured in `VerticalPodAutoscalerSpec`)

#### Startup CPU Boost Enabled & VPA Disabled

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
    # This only disables VPA actuations. It doesn't disable
    # startup boost configurations.
    updateMode: "Off"
  startupBoost:
    cpu:
      type: "Factor"
      factor: 3
      duration: 10s
```

#### Startup CPU Boost Disabled & VPA Enabled

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
    updateMode: "Auto"
```

#### Startup CPU Boost Enabled & VPA Enabled

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
    updateMode: "Auto"
  startupBoost:
    cpu:
      type: "Factor"
      factor: 3
      duration: 10s
```

### Per-container configurations (`startupBoost` configured in `ContainerPolicies`)

#### Startup CPU Boost Enabled & VPA Disabled

All containers under `example` deployment will receive "regular" VPA updates
(VPA is in `"Auto"` mode in this example), **except for**
`boosted-container-name`. `boosted-container-name` will only be CPU
boosted/unboosted (`StartupBoost` is enabled and VPA `Mode` is set to `Off`).

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
  resourcePolicy:
    containerPolicies:
      - containerName: "boosted-container-name"
        # VPA mode is set to Off, so it never changes pod resources for this
        # container. This setting is independent from the startup boost mode.
        # CPU startup boost changes will still be applied.
        mode: "Off"
        startupBoost:
          cpu:
            type: "Quantity"
            quantity: "2"
```

#### Startup CPU Boost Disabled & VPA Enabled

All containers under `example` deployment will receive "regular" VPA updates
and be CPU boosted/unboosted, except for `disable-cpu-boost-for-this-container`.
It has a `containerPolicy` `startupBoost` overriding the global VPA config that
sets the boost factor to 1.

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
  startupBoost:
    cpu:
      type: "Factor"
      factor: 2
  resourcePolicy:
    containerPolicies:
      - containerName: "disable-cpu-boost-for-this-container"
        startupBoost:
          cpu:
            type: "Factor"
            factor: 1
```

#### Startup CPU Boost Enabled & VPA Enabled

All containers under `example` deployment will receive "regular" VPA updates,
**including** `boosted-container-name`. Additionally, `boosted-container-name`
will be CPU boosted/unboosted, because it has a `StartupBoost` config in its
container policy.

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
  resourcePolicy:
    containerPolicies:
      - containerName: "boosted-container-name"
        minAllowed:
          cpu: "250m"
          memory: "100Mi"
        maxAllowed:
          cpu: "500m"
          memory: "600Mi"
        # The CPU boosted resources can go beyond maxAllowed.
        startupBoost:
          cpu:
            type: "Quantity"
            quantity: "4"
```

## Implementation History

* 2025-10-04: Update `startupBoost.cpu.type` field to correctly indicate it is a required field, not optional. The field has no default value and must be explicitly set to either "Factor" or "Quantity".
* 2025-08-05: Make some API changes and clarify behavior during and after boost period in the workflow section.
* 2025-06-23: Decouple Startup CPU Boost from InPlaceOrRecreate mode, allow
users to specify a `startupBoost` config in `VerticalPodAutoscalerSpec` and in
`ContainerPolicies` to make the API simpler and add more yaml examples.
* 2025-03-20: Initial version.

