# Known limitations

- Whenever VPA updates the pod resources, the pod is recreated, which causes all
  running containers to be recreated. The pod may be recreated on a different
  node.
- VPA cannot guarantee that pods it evicts or deletes to apply recommendations
  (when configured in `Auto` and `Recreate` modes) will be successfully
  recreated. This can be partly
  addressed by using VPA together with [Cluster Autoscaler](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#basics).
- VPA does not update resources of pods which are not run under a controller.
- Vertical Pod Autoscaler **should not be used with the [Horizontal Pod
  Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-resource-metrics)
  (HPA) on the same resource metric (CPU or memory)** at this moment. However, you can use [VPA with
  HPA on separate resource metrics](https://github.com/kubernetes/autoscaler/issues/6247) (e.g. VPA
  on memory and HPA on CPU) as well as with [HPA on custom and external
  metrics](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#scaling-on-custom-metrics).
- The VPA admission controller is an admission webhook. If you add other admission webhooks
  to your cluster, it is important to analyze how they interact and whether they may conflict
  with each other. The order of admission controllers is defined by a flag on API server.
- VPA reacts to most out-of-memory events, but not in all situations.
- VPA performance has not been tested in large clusters.
- VPA recommendation might exceed available resources (e.g. Node size, available
  size, available quota) and cause **pods to go pending**. This can be partly
  addressed by using VPA together with [Cluster Autoscaler](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#basics).
- Multiple VPA resources matching the same pod have undefined behavior.
- Running the vpa-recommender with leader election enabled (`--leader-elect=true`) in a GKE cluster
  causes contention with a lease called `vpa-recommender` held by the GKE system component of the
  same name. To run your own VPA in GKE, make sure to specify a different lease name using
  `--leader-elect-resource-name=vpa-recommender-lease` (or specify your own lease name).
