
# KEP - Balancer 

## Introduction

One of the problems that the users are facing when running Kubernetes deployments is how to 
deploy pods across several domains and keep them balanced and autoscaled at the same time. 
These domains may include:

* Cloud provider zones inside a single region, to ensure that the application is still up and running, even if one of the zones has issues.
* Different types of Kubernetes nodes. These may involve nodes that are spot/preemptible, or of different machine families. 

A single Kubernetes deployment may either leave the placement entirely up to the scheduler 
(most likely leading to something not entirely desired, like all pods going to a single domain) or 
focus on a single domain (thus not achieving the goal of being in two or more domains). 

PodTopologySpreading solves the problem a bit, but not completely. It allows only even spreading 
and once the deployment gets skewed it doesn’t do anything to rebalance. Pod topology spreading 
(with skew and/or ScheduleAnyway flag) is also just a hint, if skewed placement is available and 
allowed then Cluster Autoscaler is not triggered and the user ends up with a skewed deployment. 
A user could specify a strict pod topolog spreading but then, in case of problems the deployment
would not move its pods to the domains that are available. The growth of the deployment would also 
be totally blocked as the available domains would be too much skewed.

Thus, if full flexibility is needed, the only option is to have multiple deployments, targeting 
different domains. This setup however creates one big problem. How to consistently autoscale multiple 
deployments? The simplest idea - having multiple HPAs is not stable, due to different loads, race 
conditions or so, some domains may grow while the others are shrunk. As HPAs and deployments are 
not connected anyhow, the skewed setup will not fix itself automatically. It may eventually come to 
a semi-balanced state but it is not guaranteed. 


Thus there is a need for some component that will:

* Keep multiple deployments aligned. For example it may keep an equal ratio between the number of
pods in one deployment and the other. Or put everything to the first and overflow to the second and so on.
* React to individual deployment problems should it be zone outage or lack of spot/preemptible vms. 
* Actively try to rebalance and get to the desired layout.
* Allow to autoscale all deployments with a single target, while maintaining the placement policy.

## Balancer 

Balancer is a stand-alone controller, living in userspace (or in control plane, if needed) exposing 
a CRD API object, also called Balancer. Each balancer object has pointers to multiple deployments 
or other pod-controlling objects that expose the Scale subresource. Balancer periodically checks 
the number of running and problematic pods inside each of the targets, compares it with the desired 
number of replicas, constraints and policies and adjusts the number of replicas on the targets, 
should some of them run too many or too few of them. To allow being an HPA target Balancer itself
exposes the Scale subresource.

## Balancer API

```go
// Balancer is an object used to automatically keep the desired number of
// replicas (pods) distributed among the specified set of targets (deployments
// or other objects that expose the Scale subresource).
type Balancer struct {
   metav1.TypeMeta
   // Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
   // +optional
   metav1.ObjectMeta
   // Specification of the Balancer behavior.
   // More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status.
   Spec BalancerSpec
   // Current information about the Balancer.
   // +optional
   Status BalancerStatus
}
 
// BalancerSpec is the specification of the Balancer behavior.
type BalancerSpec struct {
   // Targets is a list of targets between which Balancer tries to distribute
   // replicas.
   Targets []BalancerTarget
   // Replicas is the number of pods that should be distributed among the
   // declared targets according to the specified policy.
   Replicas int32
   // Selector that groups the pods from all targets together (and only those).
   // Ideally it should match the selector used by the Service built on top of the
   // Balancer. All pods selectable by targets' selector must match to this selector,
   // however target's selector don't have to be a superset of this one (although
   // it is recommended).
   Selector metav1.LabelSelector
   // Policy defines how the balancer should distribute replicas among targets.
   Policy BalancerPolicy
}
 
// BalancerTarget is the declaration of one of the targets between which the balancer
// tries to distribute replicas.
type BalancerTarget struct {
   // Name of the target. The name can be later used to specify
   // additional balancer details for this target.
   Name string
   // ScaleTargetRef is a reference that points to a target resource to balance.
   // The target needs to expose the Scale subresource.
   ScaleTargetRef hpa.CrossVersionObjectReference
   // MinReplicas is the minimum number of replicas inside of this target.
   // Balancer will set at least this amount on the target, even if the total
   // desired number of replicas for Balancer is lower.
   // +optional
   MinReplicas *int32
   // MaxReplicas is the maximum number of replicas inside of this target.
   // Balancer will set at most this amount on the target, even if the total
   // desired number of replicas for the Balancer is higher.
   // +optional
   MaxReplicas *int32
}
 
// BalancerPolicyName is the name of the balancer Policy.
type BalancerPolicyName string
const (
   PriorityPolicyName     BalancerPolicyName = "priority"
   ProportionalPolicyName BalancerPolicyName = "proportional"
)
 
// BalancerPolicy defines Balancer policy for replica distribution.
type BalancerPolicy struct {
   // PolicyName decides how to balance replicas across the targets.
   // Depending on the name one of the fields Priorities or Proportions must be set.
   PolicyName BalancerPolicyName
   // Priorities contains detailed specification of how to balance when balancer
   // policy name is set to Priority.
   // +optional
   Priorities *PriorityPolicy
   // Proportions contains detailed specification of how to balance when
   // balancer policy name is set to Proportional.
   // +optional
   Proportions *ProportionalPolicy
   // Fallback contains specification of how to recognize and what to do if some
   // replicas fail to start in one or more targets. No fallback happens if not-set.
   // +optional
   Fallback *Fallback
}
 
// PriorityPolicy contains details for Priority-based policy for Balancer.
type PriorityPolicy struct {
   // TargetOrder is the priority-based list of Balancer targets names. The first target
   // on the list gets the replicas until its maxReplicas is reached (or replicas
   // fail to start). Then the replicas go to the second target and so on. MinReplicas
   // is guaranteed to be fulfilled, irrespective of the order, presence on the
   // list, and/or total Balancer's replica count.
   TargetOrder []string
}
 
// ProportionalPolicy contains details for Proportion-based policy for Balancer.
type ProportionalPolicy struct {
   // TargetProportions is a map from Balancer targets names to rates. Replicas are
   // distributed so that the max difference between the current replica share
   // and the desired replica share is minimized. Once a target reaches maxReplicas
   // it is removed from the calculations and replicas are distributed with
   // the updated proportions. MinReplicas is guaranteed for a target, irrespective
   // of the total Balancer's replica count, proportions or the presence in the map.
   TargetProportions map[string]int32
}
 
// Fallback contains information how to recognize and handle replicas
// that failed to start within the specified time period.
type Fallback struct {
   // StartupTimeout defines how long will the Balancer wait before considering
   // a pending/not-started pod as blocked and starting another replica in some other
   // target. Once the replica is finally started, replicas in other targets
   // may be stopped.
   StartupTimeout metav1.Duration
}
 
// BalancerStatus describes the Balancer runtime state.
type BalancerStatus struct {
   // Replicas is an actual number of observed pods matching Balancer selector.
   Replicas int32
   // Selector is a query over pods that should match the replicas count. This is same
   // as the label selector but in the string format to avoid introspection
   // by clients. The string will be in the same format as the query-param syntax.
   // More info about label selectors: http://kubernetes.io/docs/user-guide/labels#label-selectors
   Selector string
   // Conditions is the set of conditions required for this Balancer to work properly,
   // and indicates whether or not those conditions are met.
   // +optional
   // +patchMergeKey=type
   // +patchStrategy=merge
   Conditions []metav1.Condition
}
```
