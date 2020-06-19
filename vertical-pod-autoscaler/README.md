# Vertical Pod Autoscaler

## Contents
- [Contents](#contents)
- [Intro](#intro)
- [Installation](#installation)
  - [Compatibility](#compatibility)
  - [Notice on removal of v1beta1 version (>=0.5.0)](#notice-on-removal-of-v1beta1-version-050)
  - [Prerequisites](#prerequisites)
  - [Install command](#install-command)
  - [Quick start](#quick-start)
  - [Test your installation](#test-your-installation)
  - [Example VPA configuration](#example-vpa-configuration)
  - [Troubleshooting](#troubleshooting)
  - [Components of VPA](#components-of-vpa)
  - [Tear down](#tear-down)
- [Examples](#examples)
  - [Keeping limit proportional to request](#keeping-limit-proportional-to-request)
  - [Capping to Limit Range](#capping-to-limit-range)
  - [Resource Policy Overriding Limit Range](#resource-policy-overriding-limit-range)
- [Known limitations](#known-limitations)
- [Related links](#related-links)

# Intro

Vertical Pod Autoscaler (VPA) frees the users from necessity of setting
up-to-date resource limits and requests for the containers in their pods. When
configured, it will set the requests automatically based on usage and thus
allow proper scheduling onto nodes so that appropriate resource amount is
available for each pod. It will also maintain ratios between limits and
requests that were specified in initial containers configuration.

It can both down-scale pods that are over-requesting resources, and also
up-scale pods that are under-requesting resources based on their usage over
time.


Autoscaling is configured with a
[Custom Resource Definition object](https://kubernetes.io/docs/concepts/api-extension/custom-resources/)
called [VerticalPodAutoscaler](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go).
It allows to specify which pods should be vertically autoscaled as well as if/how the
resource recommendations are applied.

To enable vertical pod autoscaling on your cluster please follow the installation
procedure described below.

# Installation

The current default version is Vertical Pod Autoscaler 0.8.0

### Compatibility

| VPA version | Kubernetes version |
| --- | --- |
| 0.8+ | 1.13+ |
| 0.4 to 0.7 | 1.11+ |
| 0.3.X and lower | 1.7+ |

### Notice on removal of v1beta1 version (>=0.5.0)

**NOTE:** In 0.5.0 we disabled the old version of the API - `autoscaling.k8s.io/v1beta1`.
The VPA objects in this version will no longer receive recommendations and
existing recommendations will be removed. The objects will remain present though
and a ConfigUnsupported condition will be set on them.

This doc is for installing latest VPA. For instructions on migration from older versions see Migration Doc [Migration Doc](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/MIGRATE.md)


### Prerequisites

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
  is experimental and may cause downtime for your applications.
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
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: my-app-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
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

# Limits control

When setting limits VPA will conform to
[resource policies](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L82).
It will maintain limit to request ratio specified for all containers.

VPA will try to cap recommendations between min and max of
[limit ranges](https://kubernetes.io/docs/concepts/policy/limit-range/). If limit range conflicts
and VPA resource policy conflict then VPA will follow VPA policy (and set values outside limit
range).

## Examples

### Keeping limit proportional to request

Container template specifies resource request for 500 milli CPU and 1 GB of RAM. The template also
specifies resource limit of 2 GB RAM. VPA recommendation is 1000 milli CPU and 2 GB of RAM. When VPA
applies the recommendation it will also set memory limit to 4 GB.

### Capping to Limit Range

Container template specifies resource request for 500 milli CPU and 1 GB of RAM. The template also
specifies resource limit of 2 GB RAM. A limit range sets maximum limit to 3 GB RAM per container.
VPA recommendation is 1000 milli CPU and 2 GB of RAM. When VPA applies the recommendation it will
set memory limit to 3 GB (to keep it within the allowed limit range) and memory request to 1.5 GB (
to maintain 2:1 limit/request ratio from the template).

### Resource Policy Overriding Limit Range

Container template specifies resource request for 500 milli CPU and 1 GB of RAM. The template also
specifies resource limit of 2 GB RAM. A limit range sets maximum limit to 3 GB RAM per container.
VPAs Container Resource Policy requires VPA to set containers request to at least 750 milli CPU and
2 GB RAM. VPA recommendation is 1000 milli CPU and 2 GB of RAM. When applying the recommendation
VPA will set RAM request to 2 GB (following the resource policy) and RAM limit to 4 GB (to maintain
2:1 limit/request ratio from the template).

# Known limitations

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

# Related links

* [FAQ](FAQ.md)
* [Design
  proposal](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/vertical-pod-autoscaler.md)
* [API
  definition](pkg/apis/autoscaling.k8s.io/v1/types.go)
