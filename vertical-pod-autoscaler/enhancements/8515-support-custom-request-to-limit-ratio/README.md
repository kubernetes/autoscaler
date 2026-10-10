# AEP-8515: Add support for setting a custom request-to-limit policy at the VPA object level

<!-- toc -->
- [Summary](#summary)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Design Details](#design-details)
  - [API Changes](#api-changes)
  - [Behavior](#behavior)
    - [Current behavior of VPA 1.5.0](#current-behavior-of-vpa-150)
    - [Proposed feature behavior](#proposed-feature-behavior)
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
  - [Example 1](#example-1)
  - [Example 2](#example-2)
- [Implementation History](#implementation-history)
<!-- /toc -->

## Summary

Currently, when a VPA is configured to set both requests and limits automatically (i.e. when `controlledValues` is set to `RequestsAndLimits` in the VerticalPodAutoscaler CRD), it adjusts container limits proportionally based on the original request-to-limit ratio specified by the user in a higher-level controller, such as Job, Deployment, or StatefulSet, using the container-level `resources` stanzas ([Ref](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/docs/examples.md#keeping-limit-proportional-to-request)). More specifically, VPA reads the request-to-limit ratio from the managed Pods and uses that ratio to compute limits.

If the request-to-limit ratio needs to be updated (for example, because the application's resource usage has changed), users must modify the `resources.requests` or `resources.limits` fields in the controller.

This proposal introduces a new mechanism that allows users to adjust the request-to-limit policy directly at the VPA CRD level. Users can apply this mechanism to both newly created and running workloads, with more finer-grained control. The goal of this proposal is to provide a smoother way to set and update the request-to-limit policy than the current mechanism. For example:
- After the VPA starts generating recommendations, users must modify the original resource specifications in the controller to change the ratio. This workflow is not user-friendly, as users may not expect to update the controller instead of adjusting a field in the VPA CRD. Furthermore the proposed approach is more consistent with how VPA handles other fields, such as resource requests and limits, whose behavior - such as setting which resource type should VPA control - can also be configured directly through the VPA.
- The current approach does not support fine-grained control. For example, a user may want to increase the memory limit by a fixed amount, such as 100 MiB, on top of the recommended request. The current approach does not support this use case.

The feature is gated by a new feature gate, `RequestToLimitPolicy`, which is disabled by default in alpha.

## Goals

* Allow VPA to update the request-to-limit policy of a Pod's containers during Pod recreation or in-place updates based on the new `RequestToLimit` stanza.
* Introduce a new `RequestToLimit` block that enables users to adjust how the resource limit is derived from the resource request in the following ways:  
  * **Percentage**: Represents the request-to-limit ratio as a percentage-like integer.
  * **Quantity**: Adds a buffer on top of the resource request.

## Non-Goals

* This proposal does not change the core VPA algorithm or its decision-making process for when to apply the recommended values.
* This proposal does not change the default request-to-limit behavior when the feature flag is enabled. Pods managed by VPA objects that do not use the new `RequestToLimit` field will continue to follow the existing behavior. For details, see the [Behavior](#behavior) section.

## Proposal

* Extend [`ContainerResourcePolicy`](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.4.2/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L197) to allow updating the request-to-limit policy for individual containers in a Pod targeted by a VPA object. Furthermore, to enable updating the request-to-limit policy for all containers, a single wildcard entry with `containerName = '*'` can be used. This applies to all containers in the targeted Pod that do not have individual policies.

Some examples of the VPA CRD using the new `RequestToLimit` field are provided in a later [section](#examples).

## Design Details

### API Changes

A new `RequestToLimit` field will be added, with the following sub-fields:

* [Required] `RequestToLimit.CPU.Type` or `RequestToLimit.Memory.Type` (type `string`): Specifies how to apply limits proportionally to the requests. `Type` can have the following values:  
  * `Percentage`
  * `Quantity`

* [Optional] `RequestToLimit.CPU.Percentage` (type `integer`): Specifies the request-to-limit policy as a percentage-like integer.
  * If `Type` is `Percentage`, the value determines the multiplier applied to the resource request. For example:
    * 100 represents a 1x multiplier, meaning the limit equals the request.
    * 150 represents a 1.5x multiplier, meaning the request is multiplied by 1.5 to determine the limit.
    * Ratios below 1:1 are not permitted. For example, 50, which represents a 0.5x multiplier, is rejected.
  * If `Type` is `Quantity`, this field is not allowed.

* [Optional] `RequestToLimit.CPU.Quantity` (type `resource.Quantity`): The value specified in this field is added to the request to calculate the new limit.
  * If `Type` is `Percentage`, this field is not allowed.  
  * If `Type` is `Quantity`, the specified value is added to the resource request. For example, if the value is 200m, the CPU limit is calculated as the CPU request plus 200m.

* [Optional] `RequestToLimit.Memory.Percentage` (type `integer`): Same as `CPU.Percentage`.
* [Optional] `RequestToLimit.Memory.Quantity` (type `resource.Quantity`): Similar to `CPU.Quantity` except that for `Quantity` the units are memory-based (e.g., `Mi`, `Gi`) rather than CPU millicores (`m`).

Here are the Go struct definitions for the new field and its types. `ContainerResourcePolicy` is extended with a new `RequestToLimit` field (YAML key `requestToLimit`):

```go
// RequestToLimitType defines the policy type for deriving a limit from a request.
// +kubebuilder:validation:Enum=Percentage;Quantity
type RequestToLimitType string

const (
    // PercentageRequestToLimitType specifies that a percentage-like integer is used to determine the limit.
    PercentageRequestToLimitType   RequestToLimitType = "Percentage"
    // QuantityRequestToLimitType specifies that a fixed quantity is added
	  // to the request to determine the limit.
    QuantityRequestToLimitType RequestToLimitType = "Quantity"
)

// ContainerResourcePolicy is extended with the new RequestToLimit field
// (existing fields omitted for brevity).
type ContainerResourcePolicy struct {
    // RequestToLimit specifies how the resource limit should be derived from
    // the resource request for this container.
    // +optional
    RequestToLimit *RequestToLimit `json:"requestToLimit,omitempty"`
}

// RequestToLimit defines the request-to-limit policy.
type RequestToLimit struct {
    // CPU specifies the request-to-limit policy for the CPU resource.
    // If this field is not set, the request-to-limit policy for CPU is determined
    // from the Pod spec.
    // +optional
    CPU *RequestToLimitPolicy `json:"cpu,omitempty"`

    // Memory specifies the request-to-limit policy for the memory resource.
    // If this field is not set, the request-to-limit policy for memory is determined
    // from the Pod spec.
    // +optional
    Memory *RequestToLimitPolicy `json:"memory,omitempty"`
}

// RequestToLimitPolicy defines the request-to-limit policy for a resource.
// +union
// +kubebuilder:validation:XValidation:rule="(self.type == 'Percentage') == has(self.percentage)",message="percentage is required when type is Percentage and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="(self.type == 'Quantity') == has(self.quantity)",message="quantity is required when type is Quantity and forbidden otherwise"
type RequestToLimitPolicy struct {
    // Type specifies the type of request-to-limit policy to apply.
    // +unionDiscriminator
    // +required
    Type RequestToLimitType `json:"type"`

    // Percentage specifies the request-to-limit policy as a percentage-like integer.
    // The value is divided by 100 to determine the multiplier applied to the
    // recommended resource request to calculate the new limit.
    // This field is required when Type is "Percentage".
    // +unionMember=Percentage
    // +optional
    // +kubebuilder:validation:Minimum=100
    Percentage *int32 `json:"percentage,omitempty"`

    // Quantity specifies the absolute resource quantity
    // to add to the recommended resource request to determine the new limit.
    // This field is required when Type is "Quantity".
    // +unionMember=Quantity
    // +optional
    Quantity *resource.Quantity `json:"quantity,omitempty"`
}
```

### Behavior

VPA enforces the current request-to-limit policy while respecting cluster-level constraints, such as a `LimitRange`, even if this requires lowering the resource request to fit within the maximum limit.  

For example, suppose VPA calculates a new recommended CPU request of `200m`, the request-to-limit ratio is set to `1:4`, and a `LimitRange` enforces that a container cannot have more than `600m` CPU. In this case, VPA will set the CPU request to `150m` and the limit to `600m` in order to maintain the `1:4` ratio. This existing behavior is not affected by the new feature.  

#### Current behavior of VPA 1.5.0

1. The user sets the initial resource requests and limits at the controller level, such as in a Kubernetes Deployment.  
2. When VPA applies new recommendations, it preserves the request-to-limit ratio defined in the Pods `resources` stanzas.


For example, if the original resource request is `1` and the original limit is `2`, then after VPA calculates a new resource request of `10`, the new limit will be updated to `20`. In other words, the 1:2 ratio is preserved at all times.  

If the user wants to modify the request-to-limit ratio, they must update the Deployment object directly. Since the `resources.requests` and `resources.limits` fields are immutable, this results in the termination and recreation of the existing Pods.

#### Proposed feature behavior

* Values specified in `RequestToLimit` in a VerticalPodAutoscaler object take precedence over the request-to-limit ratio defined at the Pod level. For example, if a request-to-limit policy exists at the VPA object level for a specific resource of a container, the policy defined in the Pod spec is not checked for that resource.
* The updater and admission controller do not enforce the request-to-limit policy for a container when the recommender has not generated a recommendation for it. This is consistent with the current behavior for request-to-limit ratios defined in the Pod spec.
* This KEP proposes scaling limits when `RequestToLimit` is specified, even when the user omits limits from the parent object that manages the Pods and the limits are therefore absent from the Pod `resources` stanza.

The following part describes the possible approaches for determining when the updater should apply the request-to-limit policy defined in the new stanza. The admission controller's behavior is straightforward: when a Pod creation event occurs, update the limits based on the request-to-limit policy.

* **Proactive:** Apply the policy during the next Updater cycle, regardless of whether the Pod would otherwise qualify for scaling.
* **Event-driven:** React to changes to the VPA object using the `client-go` `ResourceEventHandlerFuncs` mechanism and apply the request-to-limit policy when the VPA object changes. This approach can apply the policy sooner than the previous approach because the updater does not need to wait for the next Updater cycle, which is one minute by default.
* **Recommendation-driven:** Apply the request-to-limit policy only when the Updater determines that the Pod should be updated based on the update-priority logic. This may happen during the next Updater cycle or one of the subsequent cycles, depending on the Pod's current resource requests and the calculated recommendations. For more details, see the [update priority calculator](https://github.com/iamzili/autoscaler/blob/83dc9214a7da4ec30b6fe2e4173ecbfca1c51a9d/vertical-pod-autoscaler/pkg/updater/priority/update_priority_calculator.go#L133-L153).

Although the proactive and event-driven approaches would allow the request-to-limit policy to be applied sooner, the community discussion raised concerns about the additional API server load that VPA could introduce. Therefore, this document proposes applying the request-to-limit policy defined in the VPA object at admission time (i.e. when a Pod is initially created or recreated after eviction) and by the updater using the **recommendation-driven** approach. This preserves the existing VPA update behavior rather than introducing a new trigger for Pod updates.

### Validation

  * To use this functionality, the `RequestToLimitPolicy` feature flag must be enabled.

#### Static Validation via CRD Rules

* The `RequestToLimit` configuration will be validated when VPA CRD objects are created or updated. For example:  
  * The `Type` field is marked as required, therefore its presence will be validated.
  * If `Type` is `Percentage`, the value must be greater than or equal to 100 (enforced via CRD validation rules).  

#### Dynamic Validation via Admission Controller

* When using the new `RequestToLimit` field, the `controlledValues` field must be set to `RequestsAndLimits`. It does not make sense to specify `RequestToLimit` if VPA is not allowed to update limits. This requirement is enforced by the admission controller.
* Explicitly prohibit the use of `RequestToLimit` for any resource not listed in `controlledResources`. For example, if the intention is to set a custom policy for CPU, then the value of the `controlledResources` field must include `cpu`.
* If `Type` is set to `Quantity`, then its value will be validated using the [ParseQuantity](https://github.com/kubernetes/apimachinery/blob/v0.34.1/pkg/api/resource/quantity.go#L277) function from `apimachinery`.


### Feature Enablement and Rollback

#### Enabling or Disabling the Feature in a Live Cluster

* Enable the feature by setting the `RequestToLimitPolicy` feature gate. After enabling the feature, users can define the `RequestToLimit` stanza at the VerticalPodAutoscaler object level.
* Components affected by this feature gate:  
  * admission-controller  
  * updater

#### When Enabled

* The admission controller will **accept** new VPA objects that include a configured `RequestToLimit`.  
* For containers targeted by a VPA object using `RequestToLimit`, the admission controller and/or the updater will enforce the configured request-to-limit policy. Here are some examples of how this may happen:
  * **From default to a specific policy**: This occurs when we have a running Pod targeted by a VPA object that does not define `RequestToLimit`. In this case, VPA uses the default ratio derived from the Pod's `resources` stanza. Once a user specifies a custom policy in the `RequestToLimit` stanza, the admission controller and the updater enforce the new policy according to the configured VPA mode.
  * **From one policy to another**: In this case, the default ratio defined in the Pod's `resources` stanza is ignored, and the policy specified in the `RequestToLimit` field is enforced.

#### When Disabled

* The admission controller will **reject** new VPA objects that include a configured `RequestToLimit`.  
  * A descriptive error message should be returned to the user, indicating that the feature is feature-gated.
* When a user disables the feature gate and at least one VPA object with a `RequestToLimit` stanza exists (because the feature gate was previously enabled), the updater and the admission controller use the ratios from the Pod specifications.

### Kubernetes Version Compatibility

* Kubernetes version 1.33 or higher is required to use this feature with the VPA mode [`InPlaceOrRecreate`](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/enhancements/4016-in-place-updates-support#kubernetes-version-compatibility).


### Test Plan

* Implement comprehensive unit tests to cover all new functionality.  

#### E2E

* e2e tests with `InPlaceOrRecreate` VPA mode:
  1. Add a test case where the QoS class **changes**. In this scenario, the updater should evict the affected Pods. The resulting limits are then verified.
  2. Add a test case where the QoS class **does not change**. In this scenario, the updater should apply the new request-to-limit policy using an in-place update. The resulting limits are then verified.

### Examples

Here are some examples of VPA CRDs using the new `RequestToLimit` field in different scenarios.

### Example 1

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
        requestToLimit:
          cpu:
            type: Percentage
            percentage: 200
          memory:
            type: Quantity
            quantity: 200Mi
```

### Example 2

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
        requestToLimit:
          cpu:
            type: Percentage
            percentage: 120
```

## Implementation History

* 2025-10-06: Update the `Type` field to correctly indicate that it is required, not optional. This field has no default value and must be explicitly set to either "Percentage" or "Quantity".
* 2025-09-18: Update API for consistency. Add e2e tests and other small updates.
* 2025-09-10: Initial proposal created.
