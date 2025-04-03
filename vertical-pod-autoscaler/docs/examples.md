# Examples

## Contents

- [Examples](#examples)
  - [Keeping limit proportional to request](#keeping-limit-proportional-to-request)
  - [Capping to Limit Range](#capping-to-limit-range)
  - [Resource Policy Overriding Limit Range](#resource-policy-overriding-limit-range)
  - [Starting multiple recommenders](#starting-multiple-recommenders)
  - [Using CPU management with static policy](#using-cpu-management-with-static-policy)
  - [Controlling eviction behavior based on scaling direction and resource](#controlling-eviction-behavior-based-on-scaling-direction-and-resource)
  - [Limiting which namespaces are used](#limiting-which-namespaces-are-used)
  - [Setting the webhook failurePolicy](#setting-the-webhook-failurepolicy)

## Keeping limit proportional to request

The container template specifies resource request for 500 milli CPU and 1 GB of RAM. The template also
specifies resource limit of 2 GB RAM. VPA recommendation is 1000 milli CPU and 2 GB of RAM. When VPA
applies the recommendation, it will also set the memory limit to 4 GB.

## Capping to Limit Range

The container template specifies resource request for 500 milli CPU and 1 GB of RAM. The template also
specifies resource limit of 2 GB RAM. A limit range sets a maximum limit to 3 GB RAM per container.
VPA recommendation is 1000 milli CPU and 2 GB of RAM. When VPA applies the recommendation, it will
set the memory limit to 3 GB (to keep it within the allowed limit range) and the memory request to 1.5 GB (
to maintain a 2:1 limit/request ratio from the template).

## Resource Policy Overriding Limit Range

The container template specifies resource request for 500 milli CPU and 1 GB of RAM. The template also
specifies a resource limit of 2 GB RAM. A limit range sets a maximum limit to 3 GB RAM per container.
VPAs Container Resource Policy requires VPA to set containers request to at least 750 milli CPU and
2 GB RAM. VPA recommendation is 1000 milli CPU and 2 GB of RAM. When applying the recommendation,
VPA will set RAM request to 2 GB (following the resource policy) and RAM limit to 4 GB (to maintain
the 2:1 limit/request ratio from the template).

## Starting multiple recommenders

It is possible to start one or more extra recommenders in order to use different percentile on different workload profiles.
For example you could have 3 profiles:  [frugal](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/deploy/recommender-deployment-low.yaml),
[standard](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/deploy/recommender-deployment.yaml) and
[performance](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/deploy/recommender-deployment-high.yaml) which will
use different TargetCPUPercentile (50, 90 and 95) to calculate their recommendations.

Please note the usage of the following arguments to override default names and percentiles:

- --recommender-name=performance
- --target-cpu-percentile=0.95

You can then choose which recommender to use by setting `recommenders` inside the `VerticalPodAutoscaler` spec.

## Custom memory bump-up after OOMKill

After an OOMKill event was observed, VPA increases the memory recommendation based on the observed memory usage in the event according to this formula: `recommendation = max(memory-usage-in-oomkill-event + oom-min-bump-up-bytes, memory-usage-in-oomkill-event * oom-bump-up-ratio)`.
You can configure the minimum bump-up as well as the multiplier by specifying startup arguments for the recommender:
`oom-bump-up-ratio` specifies the memory bump up ratio when OOM occurred, default is `1.2`. This means, memory will be increased by 20% after an OOMKill event.
`oom-min-bump-up-bytes` specifies minimal increase of memory after observing OOM. Defaults to `100 * 1024 * 1024` (=100MiB)

Usage in recommender deployment

```yaml
  containers:
  - name: recommender
    args:
      - --oom-bump-up-ratio=2.0
      - --oom-min-bump-up-bytes=524288000
```

## Using CPU management with static policy

If you are using the [CPU management with static policy](https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/#static-policy) for some containers,
you probably want the CPU recommendation to be an integer. A dedicated recommendation pre-processor can perform a round up on the CPU recommendation. Recommendation capping still applies after the round up.
To activate this feature, pass the flag `--cpu-integer-post-processor-enabled` when you start the recommender.
The pre-processor only acts on containers having a specific configuration. This configuration consists in an annotation on your VPA object for each impacted container.
The annotation format is the following:

```yaml
vpa-post-processor.kubernetes.io/{containerName}_integerCPU=true
```

## Controlling eviction behavior based on scaling direction and resource

To limit disruptions caused by evictions, you can put additional constraints on the Updater's eviction behavior by specifying `.updatePolicy.EvictionRequirements` in the VPA spec. An `EvictionRequirement` contains a resource and a `ChangeRequirement`, which is evaluated by comparing a new recommendation against the currently set resources for a container

Here is an example configuration which allows evictions only when CPU or memory get scaled up, but not when they both are scaled down

```yaml
 updatePolicy:
   evictionRequirements:
     - resources: ["cpu", "memory"]
       changeRequirement: TargetHigherThanRequests
```

Note that this doesn't prevent scaling down entirely, as Pods may get recreated for different reasons, resulting in a new recommendation being applied. See [the original AEP](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/enhancements/4831-control-eviction-behavior) for more context and usage information.

## Limiting which namespaces are used

 By default the VPA will run against all namespaces. You can limit that behaviour by setting the following options:

1. `ignored-vpa-object-namespaces` - A comma separated list of namespaces to ignore
1. `vpa-object-namespace` - A single namespace to monitor

These options cannot be used together and are mutually exclusive.

## Setting the webhook failurePolicy

It is possible to set the failurePolicy of the webhook to `Fail` by passing `--webhook-failure-policy-fail=true` to the VPA admission controller.
Please use this option with caution as it may be possible to break Pod creation if there is a failure with the VPA.
Using it in conjunction with `--ignored-vpa-object-namespaces=kube-system` or `--vpa-object-namespace` to reduce risk.
