# Features

## Contents

- [Limits control](#limits-control)
- [Memory Value Humanization](#memory-value-humanization)
- [CPU Recommendation Rounding](#cpu-recommendation-rounding)
- [Memory Recommendation Rounding](#memory-recommendation-rounding)
- [In-Place Updates](#in-place-updates-inplaceorrecreate)
- [VPA and LimitRange](#vpa-and-limitrange)

## Limits control

When setting limits VPA will conform to
[resource policies](https://github.com/kubernetes/autoscaler/blob/vertical-pod-autoscaler-1.2.1/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L95-L103).
It will maintain limit to request ratio specified for all containers.

VPA will try to cap recommendations between min and max of
[limit ranges](https://kubernetes.io/docs/concepts/policy/limit-range/). If limit range conflicts
with VPA resource policy, VPA will follow VPA policy (and set values outside the limit
range).

To disable getting VPA recommendations for an individual container, set `mode` to `"Off"` in `containerPolicies`.

## Memory Value Humanization

> [!WARNING]
> DEPRECATED: This feature is deprecated as of VPA v1.5.0 and will be removed in a future version. Use `--round-memory-bytes` instead for memory recommendation formatting.

> [!NOTE]
> This feature was added in v1.3.0.


VPA can present memory recommendations in human-readable binary units (KiB, MiB, GiB, TiB) instead of raw bytes, making resource recommendations easier to understand. This feature is controlled by the `--humanize-memory` flag in the recommender component.

When enabled, memory values in recommendations will be:
- Converted to the most appropriate binary unit (KiB, MiB, GiB, or TiB)
- Displayed with up to 2 decimal places for precision
- Applied to target, lower bound, and upper bound recommendations

For example, instead of seeing a memory recommendation of `262144000` bytes, you would see `250.00Mi`.

Note: Due to the conversion to binary units and decimal place rounding, the humanized values may be slightly higher than the raw byte recommendations. For example, 1537 bytes would be shown as "1.50Ki" (1536 bytes). Consider this small difference when doing precise capacity planning.

To enable this feature, set the `--humanize-memory` flag to true when running the VPA recommender:
```bash
--humanize-memory=true
```

## CPU Recommendation Rounding

VPA can provide CPU recommendations rounded up to user-specified values, making it easier to interpret and configure resources. This feature is controlled by the `--round-cpu-millicores` flag in the recommender component.

When enabled, CPU recommendations will be:
- Rounded up to the nearest multiple of the specified millicore value
- Applied to target, lower bound, and upper bound recommendations

For example, with `--round-cpu-millicores=50`, a CPU recommendation of `79m` would be rounded up to `100m`, and a recommendation of `34m` would be rounded up to `50m`.

To enable this feature, set the --round-cpu-millicores flag when running the VPA recommender:

```bash
--round-cpu-millicores=50
```

## Memory Recommendation Rounding

VPA can provide Memory recommendations rounded up to user-specified values, making it easier to interpret and configure resources. This feature is controlled by the `--round-memory-bytes` flag in the recommender component.

When enabled, Memory recommendations will be:
- Rounded up to the nearest multiple of the specified bytes value
- Applied to target, lower bound, and upper bound recommendations

For example, with `--round-memory-bytes=134217728`, a memory recommendation of `200Mi` would be rounded up to `256Mi`, and a recommendation of `80Mi` would be rounded up to `128Mi`.

To enable this feature, set the `--round-memory-bytes` flag when running the VPA recommender:

```bash
--round-memory-bytes=134217728
```

## In-Place Updates (`InPlaceOrRecreate`)

> [!WARNING] 
> FEATURE STATE: VPA v1.4.0 [alpha]
> FEATURE STATE: VPA v1.5.0 [beta]

VPA supports in-place updates to reduce disruption when applying resource recommendations. This feature leverages Kubernetes' in-place update capabilities (which is in beta as of Kubernetes 1.33) to modify container resources without requiring pod recreation.
For more information, see [AEP-4016: Support for in place updates in VPA](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/enhancements/4016-in-place-updates-support)

### Usage

To use in-place updates, set the VPA's `updateMode` to `InPlaceOrRecreate`:
```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: my-vpa
spec:
  updatePolicy:
    updateMode: "InPlaceOrRecreate"
```

### Behavior

When using `InPlaceOrRecreate` mode, VPA will first attempt to apply updates in-place, if in-place update fails, VPA will fall back to pod recreation.
Updates are attempted when:
* Container requests are outside the recommended bounds
* Quick OOM occurs
* For long-running pods (>12h), when recommendations differ significantly (>10%)

Important Notes

* Disruption Possibility: While in-place updates aim to minimize disruption, they cannot guarantee zero disruption as the container runtime is responsible for the actual resize operation.

* Memory Limit Downscaling: In the beta version, memory limit downscaling is not supported for pods with resizePolicy: PreferNoRestart. In such cases, VPA will fall back to pod recreation.

### Requirements:

* Kubernetes 1.33+ with `InPlacePodVerticalScaling` feature gate enabled
* VPA version 1.4.0+ with `InPlaceOrRecreate` feature gate enabled

### Configuration

Enable the feature by setting the following flags in VPA components ( for both updater and admission-controller ):

```bash
--feature-gates=InPlaceOrRecreate=true
``` 

### Limitations

* All containers in a pod are updated together (partial updates not supported)
* Memory downscaling requires careful consideration to prevent OOMs
* Updates still respect VPA's standard update conditions and timing restrictions
* In-place updates will fail if they would result in a change to the pod's QoS class

### Fallback Behavior

VPA will fall back to pod recreation in the following scenarios:

* In-place update is [infeasible](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/1287-in-place-update-pod-resources/README.md#resize-status) (node resources, etc.)
* Update is [deferred](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/1287-in-place-update-pod-resources/README.md#resize-status) for more than 5 minutes
* Update is in progress for more than 1 hour
* [Pod QoS](https://kubernetes.io/docs/concepts/workloads/pods/pod-qos/) class would change due to the update
* Memory limit downscaling is required with [PreferNoRestart policy](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/1287-in-place-update-pod-resources/README.md#container-resize-policy)

### Monitoring

VPA provides metrics to track in-place update operations:

* `vpa_in_place_updatable_pods_total`: Number of pods matching in-place update criteria
* `vpa_in_place_updated_pods_total`: Number of pods successfully updated in-place
* `vpa_vpas_with_in_place_updatable_pods_total`: Number of VPAs with pods eligible for in-place updates
* `vpa_vpas_with_in_place_updated_pods_total`: Number of VPAs with successfully in-place updated pods
* `vpa_updater_failed_in_place_update_attempts_total`: Number of failed attempts to update pods in-place.


## VPA and LimitRange

The Admission Controller and the Updater component post-process recommendations to obey the constraints set in [LimitRange API objects](https://kubernetes.io/docs/concepts/policy/limit-range/). There are some edge cases where they might not. This section provides examples of how these components behave with different constraints in place.

We use this Deployment with a Pod that contains two containers in all examples:

```yaml
containers:
  - name: main
    image: main:latest
    resources:
      limits:
        memory: 200Mi
      requests:
        memory: 100Mi
  - name: sidecar1
    image: alpine/curl:latest
    resources:
      limits:
        memory: 50Mi
      requests:
        memory: 50Mi
```

We use the following `updatePolicy` section of the VPA object, which targets the Deployment object shown above:

```yaml
updatePolicy:
  updateMode: Recreate
  resourcePolicy:
    containerPolicies:
      - containerName: main
        controlledResources:
          - memory
      - containerName: sidecar1
        controlledResources:
          - memory
```

The VPA's recommender calculates the following container-level recommendations. These recommendations are used across all examples. In other words, this is how the `status` section of the VPA object looks - irrelevant fields are omitted to make the manifest more compact:

```yaml
status:
  recommendation:
    containerRecommendations:
    - containerName: main
      target:
        memory: 160Mi
    - containerName: sidecar1
      target:
        memory: 25Mi
```

### Example 1

Here is the defined LimitRange object that applies to our Deployment:

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: limitrange1
  namespace: default
spec:
  limits:
  - type: Container
    min:
      memory: 50Mi
    defaultRequest:
      memory: 50Mi
```

When the Pod is killed/evicted, the admission-controller recreates it to obey the constraints set in the LimitRange object, using the following container-level resource sections:

```yaml
containers:
  - name: main
    image: main:latest
    resources:
      limits:
        memory: 320Mi
      requests:
        memory: 160Mi
  - name: sidecar1
    image: alpine/curl:latest
    resources:
      limits:
        memory: 50Mi
      requests:
        memory: 50Mi
```

There is no constraint in the LimitRange object that is currently violated for the container called `main`. Its request is defined, and the calculated request is greater than the value specified in the `min.memory` field. Similarly, its calculated limit is greater than the value defined in `min.memory`. Therefore, the admission-controller sets the target as the request and increases the limit proportionally to maintain the 1:2 ratio defined in the Deployment object.

For the `sidecar1` container, the recommended memory is 25Mi. However, this violates a constraint defined in the LimitRange object - the request must be higher than the minimum set in the LimitRange (i.e. `min.memory`). Therefore, the admission-controller increases both the requests and limits to make the Pod schedulable. The request-to-limit ratio for this container is 1:1.

### Example 2

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: limitrange1
  namespace: default
spec:
  limits:
  - type: Container
    max:
      memory: 250Mi
    default:
      memory: 25Mi
    defaultRequest:
      memory: 25Mi
```

By using this LimitRange object and the container-level recommendations, when the Pod is killed, the admission-controller sets the following values:
* For the container called `main`, the request-to-limit ratio is still 1:2, which initially results in a 160Mi memory request and a 320Mi memory limit. However, the limit now violates the `max` value in the LimitRange object. The admission-controller therefore decreases the request and limit proportionally to maintain the 1:2 ratio, resulting in a 125Mi request and a 250Mi limit.
* For the container called `sidecar1`, there is no violation based on the LimitRange. Therefore, the admission-controller sets a 25Mi request and a 25Mi limit.

For this example, here are the updated container resource sections:

```yaml
containers:
  - name: main
    image: main:latest
    resources:
      limits:
        memory: 250Mi
      requests:
        memory: 125Mi
  - name: sidecar1
    image: alpine/curl:latest
    resources:
      limits:
        memory: 25Mi
      requests:
        memory: 25Mi
```

### Example 3

In this example, let's show how a `resourcePolicy` section from a VPA object can make the Pod unschedulable, as the constraints defined there take precedence over the constraints in the LimitRange object. For example:

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: limitrange1
  namespace: default
spec:
  limits:
  - type: Container
    limits:
    - type: Container
      min:
        memory: 100Mi
      defaultRequest:
        memory: 100Mi
```

The following shows the updated policy from the VPA object:

```yaml
updatePolicy:
  updateMode: Recreate
  resourcePolicy:
    containerPolicies:
      - containerName: main
        controlledResources:
          - memory
      - containerName: sidecar1
        controlledResources:
          - memory
        maxAllowed:
          memory: 50Mi
```

By using this LimitRange and the slightly modified VPA object, where an upper limit is set for the `sidecar1` container's request, we make this Pod unschedulable. The minimum memory request for a container is 100Mi (`min.memory` from the LimitRange), while the admission-controller is only allowed to set a 50Mi memory request. This implies that extra care is needed when using `maxAllowed` and `minAllowed` if a LimitRange object is in place.