# Vertical Pod Autoscaler

## Contents
- [Intro](#intro)
- [Installation](#intallation)
  - [Prerequisites](#prerequisites)
  - [Install command](#install-command)
  - [Quick start](#quick-start)
  - [Test your installation](#test-your-installation)
  - [Example VPA configuration](#example-vpa-configuration)
  - [Troubleshooting](#troubleshooting)
  - [Components of VPA](#component-of-vpa)
  - [Tear down](#tear-down)
- [Known limitations](#known-limitation)
  - [Limitations of beta version](#limitations-of-beta-version)

# Intro

Vertical Pod Autoscaler (VPA) frees the users from necessity of setting
up-to-date resource requests for the containers in their pods.
When configured, it will set the requests automatically based on usage and
thus allow proper scheduling onto nodes so that appropriate resource amount is
available for each pod.

It can both down-scale pods that are over-requesting resources, and also up-scale pods that are under-requesting resources based on their usage over time.

Autoscaling is configured with a
[Custom Resource Definition object](https://kubernetes.io/docs/concepts/api-extension/custom-resources/)
called [VerticalPodAutoscaler](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2/types.go).
It allows to specify which pods should be vertically autoscaled as well as if/how the
resource recommendations are applied.

To enable vertical pod autoscaling on your cluster please follow the installation
procedure described below.

# Installation

The current default version is Vertical Pod Autoscaler 0.4.0

**NOTE:** version 0.4 requires at least Kubernetes 1.11 to work (needs certain
Custom Resource Definition capabilities). With older Kubernetes versions we
suggest using the [latest 0.3 version](https://github.com/kubernetes/autoscaler/blob/vpa-release-0.3/vertical-pod-autoscaler/README.md) 

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

### Prerequisites

* VPA version 0.4+ requires Kubernetes 1.11. For older versions see [latest 0.3 version](https://github.com/kubernetes/autoscaler/blob/vpa-release-0.3/vertical-pod-autoscaler/README.md)
* `kubectl` should be connected to the cluster you want to install VPA in.
* The metrics server must be deployed in your cluster. Read more about [Metrics Server](https://github.com/kubernetes-incubator/metrics-server).
* If you are using a GKE Kubernetes cluster, you will need to grant your current Google
  identity `cluster-admin` role. Otherwise you won't be authorized to grant extra
  privileges to the VPA system components.
  ```console
  $ gcloud info | grep Account    # get current google identity
  Account: [myname@example.org]

  $ kubectl create clusterrolebinding myname-cluster-admin-binding --clusterrole=cluster-admin --user=myname@example.org
  Clusterrolebinding "myname-cluster-admin-binding" created
  ```
* If you already have another version of VPA installed in your cluster, you have to tear down
  the existing installation first with:
  ```
  ./hack/vpa-down.sh
  ```

### Install command

To install VPA, please download the source code of VPA (for example with `git clone https://github.com/kubernetes/autoscaler.git`)
and run the following command inside the `vertical-pod-autoscaler` directory:

```
./hack/vpa-up.sh
```

Note: the script currently reads environment variables: `$REGISTRY` and `$TAG`.
Make sure you leave them unset unless you want to use a non-default version of VPA.

The script issues multiple `kubectl` commands to the
cluster that insert the configuration and start all needed pods (see
[architecture](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/vertical-pod-autoscaler.md#architecture-overview))
in the `kube-system` namespace. It also generates
and uploads a secret (a CA cert) used by VPA Admission Controller when communicating
with the API server.

### Quick start

After [installation](#installation) the system is ready to recommend and set
resource requests for your pods.
In order to use it you need to insert a *Vertical Pod Autoscaler* resource for
each controller that you want to have automatically computed resource requirements.
This will be most commonly a **Deployment**.
There are three modes in which *VPAs* operate:

* `"Auto"`: VPA assigns resource requests on pod creation as well as updates
  them on existing pods using the preferred update mechanism. Currently this is
  equivalent to `"Recreate"` (see below). Once restart free ("in-place") update
  of pod requests is available, it may be used as the preferred update mechanism by
  the `"Auto"` mode. **NOTE:** This feature of VPA is experimental and may cause downtime
  for your applications.
* `"Recreate"`: VPA assigns resource requests on pod creation as well as updates
  them on existing pods by evicting them when the requested resources differ significantly
  from the new recommendation (respecting the Pod Disruption Budget, if defined).
  This mode should be used rarely, only if you need to ensure that the pods are restarted
  whenever the resource request changes. Otherwise prefer the `"Auto"` mode which may take
  advantage of restart free updates once they are available. **NOTE:** This feature of VPA
  is experimental and may cause dowtime for your applications.
* `"Initial"`: VPA only assigns resource requests on pod creation and never changes them
  later.
* `"Off"`: VPA does not automatically change resource requirements of the pods.
  The recommendations are calculated and can be inspected in the VPA object.

### Test your installation

A simple way to check if Vertical Pod Autoscaler is fully operational in your
cluster is to create a sample deployment and a corresponding VPA config:
```
kubectl create -f examples/hamster.yaml
```

The above command creates a deployment with 2 pods, each running a single container
that requests 100 millicores and tries to utilize slightly above 500 millicores.
The command also creates a VPA config pointing at the deployment.
VPA will observe the behavior of the pods and after about 5 minutes they should get
updated with a higher CPU request
(note that VPA does not modify the template in the deployment, but the actual requests
of the pods are updated). To see VPA config and current recommended resource requests run:
```
kubectl describe vpa
```


*Note: if your cluster has little free capacity these pods may be unable to schedule.
You may need to add more nodes or adjust examples/hamster.yaml to use less CPU.*

### Example VPA configuration

```
apiVersion: autoscaling.k8s.io/v1beta2
kind: VerticalPodAutoscaler
metadata:
  name: my-app-vpa
spec:
  targetRef:
    apiVersion: "extensions/v1beta1"
    kind:       Deployment
    name:       my-app
  updatePolicy:
    updateMode: "Auto"
```

### Troubleshooting

To diagnose problems with a VPA installation, perform the following steps:

* Check if all system components are running:
```
kubectl --namespace=kube-system get pods|grep vpa
```
The above command should list 3 pods (recommender, updater and admission-controller)
all in state Running.

* Check if the system components log any errors.
For each of the pods returned by the previous command do:
```
kubectl --namespace=kube-system logs [pod name]| grep -e '^E[0-9]\{4\}'
```

* Check that the VPA Custom Resource Definition was created:
```
kubectl get customresourcedefinition|grep verticalpodautoscalers
```

### Components of VPA

The project consists of 3 components:

* [Recommender](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/recommender/README.md) - it monitors the current and past resource consumption and, based on it,
provides recommended values containers' cpu and memory requests.

* [Updater](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/updater/README.md) - it checks which of the managed pods have correct resources set and, if not,
kills them so that they can be recreated by their controllers with the updated requests.

* [Admission Plugin](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/admission-controller/README.md) - it sets the correct resource requests on new pods (either just created
or recreated by their controller due to Updater's activity).

More on the architecture can be found [HERE](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/vertical-pod-autoscaler.md).

### Tear down

Note that if you stop running VPA in your cluster, the resource requests
for the pods already modified by VPA will not change, but any new pods
will get resources as defined in your controllers (i.e. deployment or
replicaset) and not according to previous recommendations made by VPA.

To stop using Vertical Pod Autoscaling in your cluster:
* If running on GKE, clean up role bindings created in [Prerequisites](#prerequisites):
```
kubectl delete clusterrolebinding myname-cluster-admin-binding
```
* Tear down VPA components:
```
./hack/vpa-down.sh
```

# Known limitations

## Limitations of beta version

* Updating running pods is an experimental feature of VPA. Whenever VPA updates
  the pod resources the pod is recreated, which causes all running containers to
  be restarted. The pod may be recreated on a different node.
* VPA does not evict pods which are not run under a controller. For such pods
  `Auto` mode is currently equivalent to `Initial`.
* Vertical Pod Autoscaler **should not be used with the [Horizontal Pod Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) (HPA) on CPU or memory** at this moment. 
  However, you can use VPA with [HPA on custom and external metrics](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-custom-metrics).
* The VPA admission controller is an admission webhook. If you add other admission webhooks
  to you cluster, it is important to analyze how they interact and whether they may conflict
  with each other. The order of admission controllers is defined by a flag on APIserver.
* VPA reacts to most out-of-memory events, but not in all situations.
* VPA performance has not been tested in large clusters.
* VPA recommendation might exceed available resources (e.g. Node size, available
  size, available quota) and cause **pods to go pending**. This can be partly 
  addressed by using VPA together with [Cluster Autoscaler](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#basics).
* Multiple VPA resources matching the same pod have undefined behavior.
* VPA does not change resource limits. This implies that recommendations are
  capped to limits during actuation.
  **NOTE** This behaviour is likely to change so please don't rely on it.

# Related links

* [FAQ](FAQ.md)
* [Design
  proposal](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/vertical-pod-autoscaler.md)
* [API
  definition](pkg/apis/autoscaling.k8s.io/v1beta2/types.go)
