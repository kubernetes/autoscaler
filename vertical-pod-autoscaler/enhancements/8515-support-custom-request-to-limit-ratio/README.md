# AEP-8515: Add support for setting a custom request-to-limit ratio at the VPA object level

<!-- toc -->
- [Summary](#summary)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
    - [API Changes](#api-changes)
    - [Behavior](#behavior)
      - [Current behavior of VPA 1.4.2](#current-behavior-of-vpa-142)
      - [Proposed feature behavior](#proposed-feature-behavior)
    - [Notes/Constraints/Caveats](#notesconstraintscaveats)
    - [Validation](#validation)
      - [Static Validation via CRD Rules](#static-validation-via-crd-rules)
      - [Dynamic Validation via Admission Controller](#dynamic-validation-via-admission-controller)
    - [Feature Enablement and Rollback](#feature-enablement-and-rollback)
      - [Enabling or Disabling the Feature in a Live Cluster](#enabling-or-disabling-the-feature-in-a-live-cluster)
      - [When Enabled](#when-enabled)
      - [When Disabled](#when-disabled)
    - [Kubernetes Version Compatibility](#kubernetes-version-compatibility)
    - [Test Plan](#test-plan)
      - [E2E](#e2e)
    - [Examples](#examples)
  - [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Currently, when VPA is configured to set both requests and limits automatically (i.e. when `controlledValues` is set to `RequestsAndLimits` in the VerticalPodAutoscaler CRD), it adjusts the container limits proportionally based on the original request-to-limit ratio(s) specified by the user in workload API objects such as Deployments or StatefulSets, using the container-level `resources` stanzas - [Ref](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/docs/examples.md#keeping-limit-proportional-to-request). More specifically, VPA reads the request-to-limit ratio from the managed Pods and uses that ratio to compute limits.

If the request-to-limit ratio needs to be updated (for example, because the application's resource usage has changed), users must modify the `resources.requests` or `resources.limits` fields in the workload's API object. However, applying these changes triggers the termination and recreation of existing Pods.

This proposal introduces a new mechanism that allows VPA users to adjust the request-to-limit ratio directly at the VPA CRD level, both for already running workloads and for newly created workloads.

The feature is gated by a new feature gate, `RequestToLimitRatio`, which is disabled by default in alpha.

## Goals

* Allow VPA to update the request-to-limit ratio of a Pod's containers during Pod recreation or in-place updates.  
* Introduce a new `RequestToLimitRatio` block that enables users to adjust the request-to-limit ratio in the following ways:  
  * **Factor**: Multiplies the recommended request by a specified value, and the result is set as the new limit, for example:
    * If the value for `Factor` is set to `2`, the limit will be twice the recommended request.  
    * If the value for `Factor` is set to `1.1`, the limit will be 10% higher than the recommended request.  
  * **Quantity**: Adds a buffer on top of the resource request. This can be expressed as an **absolute value with units** (e.g. `100Mi`, `10m`).  

## Non-Goals

* This proposal does not change the core VPA algorithm or its decision-making process for when to apply the recommended values or set limits proportionally.
* This proposal does not change the default request-to-limit behavior when the feature flag is enabled. Pods managed by VPA objects that do not use the new `RequestToLimitRatio` field will continue to follow the existing behavior. For details, see the [Behavior](#behavior) section.

## Proposal

* Extend [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.4.2/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L197) to allow updating the request-to-limit ratio for individual containers in a Pod targeted by a VPA object. Furthermore, to enable updating the ratio for all containers, a single wildcard entry with `containerName = '*'` can be used. This applies to all containers in the targeted Pod that do not have individual policies.


Some examples of the VPA CRD using the new `RequestToLimitRatio` field are provided in a later [section](#examples).

## Design Details

### API Changes

A new `RequestToLimitRatio` field will be added, with the following sub-fields:

* [Required] `RequestToLimitRatio.CPU.Type` or `RequestToLimitRatio.Memory.Type` (type `string`): Specifies how to apply limits proportionally to the requests. `Type` can have the following values:  
  * `Factor`: Interpreted as a multiplier for the recommended request.  
    * Example: a value of `2` will double the limits.  
  * `Quantity`: Adds an absolute value on top of the requests to determine the new limit.  
    * Example: for memory, a value of `100Mi` means the new limit will be: calculated memory request + `100Mi`. 

* [Optional] `RequestToLimitRatio.CPU.Factor` (type `float`): The factor to apply to the CPU request.
  * If `Type` is `Factor` a value of `3` will triple the CPU limits.  
  * If `Type` is `Quantity`, this field is not allowed.

* [Optional] `RequestToLimitRatio.CPU.Quantity` (type `Quantity`): The value specified in this field is added to the request to calculate the new limit.
  * If `Type` is `Factor`, this field is not allowed.  
  * If `Type` is `Quantity` a CPU resource quantity added. For example, if the value is `200m`, the CPU limit will be calculated as the CPU request plus 200 millicores.

* [Optional] `RequestToLimitRatio.Memory.Factor` (type `float`): Same as `CPU.Factor`.
* [Optional] `RequestToLimitRatio.Memory.Quantity` (type `Quantity`): Similar to `CPU.Quantity` except that for `Quantity` the units are memory-based (e.g., `Mi`, `Gi`) rather than CPU millicores (`m`).

### Behavior

VPA enforces the current request-to-limit ratio while respecting cluster-level constraints, such as a `LimitRange`, even if this requires lowering the resource request to fit within the maximum limit.  

For example, suppose VPA calculates a new recommended CPU request of `200m`, the request-to-limit ratio is set to `1:4`, and a `LimitRange` enforces that a container cannot have more than `600m` CPU. In this case, VPA will set the CPU request to `150m` and the limit to `600m` in order to maintain the `1:4` ratio. This existing behavior is not affected by the new feature.  

#### Current behavior of VPA 1.4.2

1. The user sets the initial resource requests and limits at the workload API level, such as in a Kubernetes Deployment.  
2. When VPA applies new recommended resource request values, it maintains the initially set request-to-limit ratio.

For example, if the original resource request is `1` and the original limit is `2`, then after VPA calculates a new resource request of `10`, the new limit will be updated to `20`. In this version of VPA, the 1:2 ratio is preserved at all times.  

If the user wants to modify the request-to-limit ratio, they must update the Deployment object directly. Since the `resources.requests` and `resources.limits` fields are immutable, this results in the termination and recreation of the existing Pods.

#### Proposed feature behavior

The values specified under `RequestToLimitRatio` in the VPA object will take precedence over the request-to-limit ratio initially set at the workload API level. For example, if the CPU ratio is initially set to `1:2` at the workload API level, but the VPA object sets the CPU request-to-limit ratio to `1:10` using the new `RequestToLimitRatio` field, VPA will use the ratio from the `RequestToLimitRatio` field (`1:10`) when applying new recommended values.

The behavior after implementing this feature is as follows:

1. The user defines a VPA object with the `controlledValues` field set to `RequestsAndLimits` and configures the request-to-limit ratio using the new `RequestToLimitRatio` sub-fields. Based on VPA's mode, the following occurs:
   * **Recreate mode**: When a new request-to-limit ratio is set, the ratio is applied only on Pod creation, after the Updater evicts the running Pod. In this mode, updating the request-to-limit ratio on a running Pod will affect the limits only after the Pod is evicted (either by the Updater or manually, e.g. via `kubectl delete pod`) when the current `resources.requests` differ significantly from the new recommendation.  
   * **InPlaceOrRecreate mode** (alpha in v1.4.0): When a new request-to-limit ratio is set, the VPA Updater will attempt in-place updates using the `/resize` subresource to modify `Pod.Spec.Containers[i].Resources.limits` or `Pod.Spec.Containers[i].Resources.requests` in certain situations. If the in-place update fails, it falls back to evicting the Pod and performing a recreation. For more details, see the [In-Place Updates documentation](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/docs/features.md#in-place-updates-inplaceorrecreate).  
   * **Initial mode**: VPA updates the request-to-limit ratio only during Pod creation and does not change it later.
2. If the `RequestToLimitRatio` feature gate is enabled and a user does not specify the sub-field `RequestToLimitRatio` on a VPA object, the request-to-limit ratio already set in the workload API (e.g. Deployment API) is used.
3. If the `RequestToLimitRatio` feature gate is disabled, the request-to-limit ratio already set in the workload API (e.g. Deployment API) is used.

### Notes/Constraints/Caveats

* This proposal could act as a temporary workaround for users who need to handle requests/limits immutability in their workloads. Note that there is already work underway to relax the immutability of requests and limits at the workload level (e.g. Deployments and StatefulSets), tracked in this issue: [Relaxing immutability of Pod Templates resources/limits](https://github.com/kubernetes/kubernetes/issues/132436)

### Validation

  * To use this functionality, the `RequestToLimitRatio` feature flag must be enabled.

#### Static Validation via CRD Rules

* The `RequestToLimitRatio` configuration will be validated when VPA CRD objects are created or updated. For example:  
  * The `Type` field is marked as required, therefore its presence will be validated.
  * If `Type` is `Factor`, the value must be greater than or equal to 1 (enforced via CRD validation rules).  

#### Dynamic Validation via Admission Controller

* When using the new `RequestToLimitRatio` field, the `controlledValues` field must be set to `RequestsAndLimits`. It does not make sense to specify `RequestToLimitRatio` if VPA is not allowed to update limits. This requirement is enforced by the admission controller.
* Explicitly prohibit the use of `RequestToLimitRatio` for any resource not listed in `controlledResources`. For example, if the intention is to set a custom ratio for CPU, then the value of the `controlledResources` field must include `cpu`.
* If `Type` is set to `Quantity`, then its value will be validated using the [ParseQuantity](https://github.com/kubernetes/apimachinery/blob/v0.34.1/pkg/api/resource/quantity.go#L277) function from `apimachinery`.


### Feature Enablement and Rollback

#### Enabling or Disabling the Feature in a Live Cluster

* Enable the feature by setting the `RequestToLimitRatio` feature gate.  
* Components affected by this feature gate:  
  * admission-controller  
  * updater

#### When Enabled

* The admission controller will **accept** new VPA objects that include a configured `RequestToLimitRatio`.  
* For containers targeted by a VPA object using `RequestToLimitRatio`, the admission controller and/or the updater will enforce the configured ratio. Here are some examples of how this may happen:
  * **From default to a specific ratio**: This occurs when we have a running Pod targeted by a VPA object that does not define `RequestToLimitRatio`. In this case, VPA uses the default ratio derived from the Pod's `resources` stanza (assuming the Pod has a single container). Once we specify a custom ratio using the `RequestToLimitRatio` field, the new ratio is not applied immediately, as the updater still relies on its current behavior to decide when to evict the Pod or perform in-place update. With the `InPlaceOrRecreate` mode, the updater sends patches to the `resize` subresource [as shown here](https://github.com/kubernetes/autoscaler/blob/7b95cb06cb0843c1cd9432a3db893c001e1bc33c/vertical-pod-autoscaler/pkg/updater/restriction/pods_inplace_restriction.go#L137). This can result in two outcomes:
    1. If the in-place update fails for any reason - for example, because the change would alter the Pod's QoS class - the updater evicts the Pod.
    2. If the in-place update succeeds, the updater is finished for that Pod.
  * **From one ratio to another**: In this case, the default ratio defined in the Pod's `resources` stanza is ignored, and the ratio specified in the `RequestToLimitRatio` field is enforced.


#### When Disabled

* The admission controller will **reject** new VPA objects that include a configured `RequestToLimitRatio`.  
  * A descriptive error message should be returned to the user, indicating that the feature is feature-gated.  
* The admission controller and updater will behave as before, according to the behavior described [here](#current-behavior-of-vpa-142).

### Kubernetes Version Compatibility

* Kubernetes version 1.33 or higher is required to use this feature with the VPA mode [`InPlaceOrRecreate`](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/enhancements/4016-in-place-updates-support#kubernetes-version-compatibility).


### Test Plan

* Implement comprehensive unit tests to cover all new functionality.  

#### E2E

* e2e tests with `InPlaceOrRecreate` VPA mode:
  1. Add a test case where the QoS class **changes**. In this scenario, we expect the updater to evict the affected Pods, since the QoS field is immutable. The resulting limits are verified.
  2. Add a test case where the QoS class **does not change**. In this scenario the updater should apply the new ratio using the in-place update mechanism. The resulting limits are verified.

### Examples

Here are some examples of VPA CRDs using the new `RequestToLimitRatio` field in different scenarios.

The following is a sample VPA manifest that targets a specific container named `app` in a Pod. In this manifest:  
* The CPU limit is set to twice the calculated CPU request.  
* The memory limit is set to the calculated memory request plus `200Mi`.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: my-app
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  updatePolicy:
    updateMode: InPlaceOrRecreate
  resourcePolicy:
    containerPolicies:
      - containerName: app
        controlledResources: ["cpu", "memory"]
        controlledValues: RequestsAndLimits
        RequestToLimitRatio:
          cpu:
            type: Factor
            factor: 2
          memory:
            type: Quantity
            quantity: 200Mi
```

In the manifest below, we configure VPA to control only the CPU resource's requests and limits for the container named `app`. The CPU limit is calculated by increasing the recommended CPU request by 20% (i.e. `recommended request × 1.2`).

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: my-app
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  updatePolicy:
    updateMode: InPlaceOrRecreate
  resourcePolicy:
    containerPolicies:
      - containerName: app
        controlledResources: ["cpu"]
        controlledValues: RequestsAndLimits
        RequestToLimitRatio:
          cpu:
            type: Factor
            factor: 1.2
```

## Implementation History

* 2025-10-06: Update the `Type` field to correctly indicate that it is required, not optional. This field has no default value and must be explicitly set to either "Factor" or "Quantity".
* 2025-09-18: Update API for consistency. Add e2e tests and other small updates.
* 2025-09-10: Initial proposal created.