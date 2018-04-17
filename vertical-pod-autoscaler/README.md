# Vertical Pod Autoscaler

# Intro

Vertical Pod Autoscaler (VPA) frees the users from necessity of setting
up-to-date resource requests for their containers in pods.
When configured, it will set the requests automatically based on usage and
thus allow proper scheduling onto nodes so that appropriate resource amount is
available for each pod.

Autoscaling is configured with a
[Custom Resource Definition object](https://kubernetes.io/docs/concepts/api-extension/custom-resources/)
called [VerticalPodAutoscaler](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1/types.go).
It allows to specify which pods should be under vertically autoscaled as well as if/how the
resource recommendations are applied.

To enable vertical pod autoscaling on your cluster please follow the installation
procedure described below.


# Installation

### Prerequisites

* It is strongly recommended to use Kubernetes 1.9 or greater.
  Your cluster must support MutatingAdmissionWebhooks, which are enabled by default
  since 1.9 ([#58255](https://github.com/kubernetes/kubernetes/pull/58255)).
  Read more about [VPA Admission Webhook](./admission-controller/README.md#running)).
* `kubectl` should be connected to the cluster you want to install VPA in.
* If you are using a GKE Kubernetes cluster, you will need to grant your current Google
  identity `cluster-admin` role. Otherwise you won't be authorized to grant extra
  privileges to the VPA system components.
  ```console
  $ gcloud info | grep Account    # get current google identity
  Account: [myname@example.org]

  $ kubectl create clusterrolebinding myname-cluster-admin-binding --clusterrole=cluster-admin --user=myname@example.org
  Clusterrolebinding "myname-cluster-admin-binding" created
  ```

### Install command

To install VPA, please download the source code of VPA (for example with `git clone https://github.com/kubernetes/autoscaler.git`)
and run:

```
./hack/vpa-up.sh
```
Inside `vertical-pod-autoscaler` directory.


Note: the script currently reads environment variables: `$REGISTRY` and `$TAG`.
Make sure you leave them unset unless you want to use a non-default version of VPA.

The script issues multiple `kubectl` commands to the
cluster that insert the configuration and start all needed pods (see
[architecture](http://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/vertical-pod-autoscaler.md#architecture-overview)) 
in the `kube-system` namespace. It also generates
and uploads a secret (a CA cert) used by VPA Admission Controller when communicating
with the API server.

### Quick start

After [installation](#installation) the system is ready to recommend and set
resource requests for your pods.
In order to use it you need to insert a *Vertical Pod Autoscaler* resource for
each logical group of pods that have similar resource requirements.
We recommend to insert a *VPA* per each *Deployment* you want to control
automatically and use the same label selector as the *Deployment* uses.
There are three modes in which *VPAs* operate:

* `"Auto"`: VPA assigns resource requests on Pod creation as well as updates
  them on running Pods (only if they differ significantly from the new
  recommendation and only within Eviction API limits). This is the default setting.
* `"Initial"`: VPA only assigns resource requests on Pod creation and never changes them
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
that uses one full CPU core. The request is initially set to 0.5 CPU on each
container. It also creates a VPA config with selector that matches the pods in the
deployment.
VPA will observe the behavior of the pods and after about 5 minutes they should get
updated with the CPU request slightly above 1 core
(note that VPA does not modify the template in the deployment, but the actual requests
of the pods are updated). To see VPA config and current recommended resource requests run:
```
kubectl describe vpa
```


### Example VPA configuration

```
apiVersion: poc.autoscaling.k8s.io/v1alpha1
kind: VerticalPodAutoscaler
metadata:
  name: my-app-vpa
spec:
  selector:
    matchLabels:
      app: my-app
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

* Recommender - it monitors the current and past resource consumption and, based on it,
provides recommended values containers' cpu and memory requests.

* Updater - it checks which of the managed Pods have correct resources set and, if not,
kills them so that they can be recreated by their controllers with the updated requests.

* Admission Plugin - it sets the correct resource requests on new pods (either just created
or recreated by their controller due to Updater's activity).

More on the architecture can be found [HERE](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/vertical-pod-autoscaler.md).

# Known limitations of the alpha version

* Whenever VPA updates the pod resources the pod is recreated, which causes all
  running containers to be restarted.
* VPA in `auto` mode can only be used on pods that run under a controller
  (such as Deployment), which is responsible for restarting deleted pods.
  **Using VPA in `auto` mode with a pod not running under any controller will
  cause the pod to be deleted and not recreated**.
* The VPA admission controller is an admission webhook. If you add other admission webhooks
  to you cluster, it is important to analyze how they interact and whether they may conflict
  with each other. The order of admission controllers is defined by a flag on APIserver.
* VPA reacts to some out-of-memory events, but not in all situations.
* VPA performance has not been tested in large clusters.
* VPA recommendation might exceed available resources (e.g. Node size, available
  size, available quota) and cause Pods to go pending.
* Multiple VPA resources matching the same Pod have undefined behavior.

# Related links

* [Design
  proposal](http://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/vertical-pod-autoscaler.md)
* [API
  definition](http://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1/types.go)
