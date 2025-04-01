# Features

## Contents

- [Limits control](#limits-control)
- [Memory Value Humanization](#memory-value-humanization)
- [CPU Recommendation Rounding](#cpu-recommendation-rounding)
- [In-Place Updates](#in-place-updates)

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

## In-Place Updates

VPA supports in-place updates to reduce disruption when applying resource recommendations. This feature leverages Kubernetes' in-place update capabilities (beta in 1.33) to modify container resources without requiring pod recreation.

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

* Kubernetes 1.33+ with in-place update feature enabled
* VPA version 1.4.0+ with `InPlaceOrRecreate` feature gate enabled
* Cluster feature gate InPlacePodVerticalScaling enabled

### Configuration

Enable the feature by setting the following flags in VPA components ( for both updater and admission-controller ):

```bash
--feature-gates=InPlaceOrRecreate=true
``` 

### Limitations

* All containers in a pod are updated together (partial updates not supported)
* Memory downscaling requires careful consideration to prevent OOMs
* Updates still respect VPA's standard update conditions and timing restrictions
* In-place updates will fail for pods with Guaranteed QoS class (requires pod recreation)

### Fallback Behavior

VPA will fall back to pod recreation in the following scenarios:

* In-place update is [infeasible](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/1287-in-place-update-pod-resources/README.md#resize-status) (node resources, etc.)
* Update is [deferred](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/1287-in-place-update-pod-resources/README.md#resize-status) for more than 5 minutes
* Update is in progress for more than 1 hour
* [Pod QoS](https://kubernetes.io/docs/concepts/workloads/pods/pod-qos/) class would change due to the update
* Memory limit downscaling is required with [PreferNoRestart policy](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/1287-in-place-update-pod-resources/README.md#container-resize-policy)

### Monitoring

VPA provides metrics to track in-place update operations:

* vpa_in_place_updatable_pods_total: Number of pods matching in-place update criteria
* vpa_in_place_updated_pods_total: Number of pods successfully updated in-place
* vpa_vpas_with_in_place_updatable_pods_total: Number of VPAs with pods eligible for in-place updates
* vpa_vpas_with_in_place_updated_pods_total: Number of VPAs with successfully in-place updated pods