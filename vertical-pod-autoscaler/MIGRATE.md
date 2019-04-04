# Vertical Pod Autoscaler Migration

### Notice on switching to v1beta2 version (0.3.X to >=0.4.0)

In 0.4.0 we introduced a new version of the API - `autoscaling.k8s.io/v1beta2`.
Full API is accessible [here](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2/types.go).

The change introduced is in the way you express which pods should be scaled by a
given Vertical Pod Autoscaler. In short we are moving from label selectors to
controller references. This change is introduced due to two main reasons:
* Use of selectors is prone to misconfigurations - e.g. VPA objects targeting
all pods, overlapping VPA objects
* This change aligns VPA with [Horizontal Pod Autoscaler
  API](https://github.com/kubernetes/api/blob/master/autoscaling/v1/types.go)

Let's see an example ilustrating the change:

**[DEPRECATED]** In `v1beta1` pods to scale by VPA are specified by a
[kubernetes label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors).

```yaml
apiVersion: "autoscaling.k8s.io/v1beta1"
kind: VerticalPodAutoscaler
metadata:
  name: hamster-vpa-deprecated
spec:
  selector: # selector is the deprecated way
    matchLabels:
      app: hamster
```

**[RECOMMENDED]** In `v1beta2` pods to scale by VPA are specified by a
target reference. This target will usually be a Deployment, as configured in the
example below.

```yaml
apiVersion: "autoscaling.k8s.io/v1beta2"
kind: VerticalPodAutoscaler
metadata:
  name: hamster-vpa
spec:
  targetRef:
    apiVersion: "extensions/v1beta1"
    kind:       Deployment
    name:       hamster
```

The target object can be a well known controller (Deployment, ReplicaSet, DaemonSet, StatefulSet etc.)
or any object that implements the scale subresource. VPA uses ScaleStatus to
retrieve the pod set controlled by this object.
If VerticalPodAutoscaler cannot use specified target it will report
ConfigUnsupported condition.

Note that VerticalPodAutoscaler does not require full implementation
of scale subresource - it will not use it to modify the replica count.
The only thing retrieved is a label selector matching pods grouped by this controller.

See complete examples:
* [v1beta2](./examples/hamster.yaml)
* [v1beta1](./examples/hamster-deprecated.yaml)

You can perform a 0.3 to 0.4 upgrade without losing your VPA objects.
The recommended way is as follows:

1. Run `./hack/vpa-apply-upgrade.sh` - this will restart your VPA installation with
a new version, add the new API and keep all your VPA objects.
1. Your `v1beta1` objects will be marked as deprecated but still work
1. Switch your VPA definition to
`apiVersion: "autoscaling.k8s.io/v1beta2"`
1. Modify the VPA spec to:
```yaml
spec:
  # Note the empty selector field - this is needed to remove previously defined selector
  selector:
  targetRef:
    apiVersion: "extensions/v1beta1"
    kind:       "Deployment"
    name:       "<deployment_name>" # This matches the deployment name
```
5. Kubectl apply -f the above

You can also first try the new API in the `"Off"` mode.

### Notice on switching from alpha to beta (<0.3.0 to 0.4.0+)

**NOTE:** We highly recommend switching to the 0.4.X version. However,
for instructions on switching to 0.3.X see the [0.3 version README](https://github.com/kubernetes/autoscaler/blob/vpa-release-0.3/vertical-pod-autoscaler/README.md)

Between versions 0.2.x and 0.4.x there is an alpha to beta switch which includes
a change of VPA apiVersion. The safest way to switch is to use `vpa-down.sh`
script to tear down the old installation of VPA first. This will delete your old
VPA objects that have been defined with `poc.autoscaling.k8s.io/v1alpha1`
apiVersion. Then use `vpa-up.sh` to bring up the new version of VPA and create
your VPA objects from the scratch, passing apiVersion
`autoscaling.k8s.io/v1beta2` and switching from selector to targetRef, as
described in the prevous section.
