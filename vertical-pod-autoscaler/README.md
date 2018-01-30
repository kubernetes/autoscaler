# Vertical Pod Autoscaler

# Intro

Vertical Pod Autoscaler (VPA) frees the users from necessity of setting
up-to-date resource requests for their containers in pods.
When configured, it will set the requests automatically based on usage and
thus allow proper scheduling onto nodes so that appropriate resource amount is
available for each pod.

# For users

### Installation

**Prerequisites** (to be automatized):

* Install Prometheus (otherwise VPA will only have current usage data, no
  history).
* Make sure your cluster supports MutatingAdmissionWebhooks (see
  [here](./admission-controller/README.md#running)).
* `kubectl` should be connected to the cluster you want to install VPA in.

To install VPA, run:

```
./hack/vpa-up.sh
```

Note: the script currently depends on environment variables: `$REGISTRY` and `$TAG`.
Make sure you don't set them if you want the released version.

The script issues multiple `kubectl` commands to the
cluster that insert the configuration and start all needed pods (see
[architecture](#architecture)) in the `kube-system` namespace.

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
  recommendation and only within Eviction API limits).
* `"Initial"`: VPA only assigns resource requests on Pod creation and never changes them
  later.
* `"Off"`: VPA does not automatically change resource requirements of the pods.
  The recommendations are calculated and can be inspected in the VPA object.

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

### Known limitations

* The VPA admission controller is an admission webhook. The order of admission
  controllers is defined by flag on APIserver.
  VPA admission controller might have **relations to other admission controllers**,
  e.g. running it after quota admission might cause making incorrect decisions.
* Out-of-memory events / pod evictions are not taken into account for memory
  usage data. **Containers dying because of lack of memory might not get bigger
  recommendations**.
* Recommender reads some amount of history (currently one day) and treats all
  samples from that period identically, **no matter how recent they are**. Also, it
  does not forget samples after they go out of the one day window, so the
  **history length will grow during the lifetime of the recommender binary**.

# For developers

### Architecture

The system consists of three separate binaries:
[recommender](./recommender/), [updater](./updater/) and
[admission controller](./admission-controller/).

### How to plug in a modified recommender

First, make any changes you like in recommender code.
Then, build it with
```
make --directory recommender build docker
```
Remember the command puts your build docker image into your GCR registry
and tags it using env variables: `$REGISTRY`, e.g. `gcr.io/my-project` and
`$TAG`, e.g. `my-latest-release`.
To deploy that version, follow [installation](#installation).
If you already had VPA installed, you can run:
```
./hack/vpa-down.sh recommender
./hack/vpa-up.sh recommender
```
to only recreate the recommender deployment and keep the rest of VPA system as
it was.

### How to modify other components

Updater and admission controller can be modified, built and deployed similarly
to [recommender](#how-to-plug-in-a-modified-recommender).

# Related links

* [Design
  proposal](http://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/vertical-pod-autoscaler.md)
* [API
  definition](http://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1/types.go)
