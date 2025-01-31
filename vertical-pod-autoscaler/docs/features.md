# Features

## Contents

- [Limits control](#limits-control)
- [Memory Value Humanization](#memory-value-humanization)

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