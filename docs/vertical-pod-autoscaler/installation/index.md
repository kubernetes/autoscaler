# Installation

The current default version is Vertical Pod Autoscaler 1.2.1

## Compatibility

| VPA version     | Kubernetes version |
|-----------------|--------------------|
| 1.2.1           | 1.27+              |
| 1.2.0           | 1.27+              |
| 1.1.2           | 1.25+              |
| 1.1.1           | 1.25+              |
| 1.0             | 1.25+              |
| 0.14            | 1.25+              |
| 0.13            | 1.25+              |
| 0.12            | 1.25+              |
| 0.11            | 1.22 - 1.24        |
| 0.10            | 1.22+              |
| 0.9             | 1.16+              |
| 0.8             | 1.13+              |
| 0.4 to 0.7      | 1.11+              |
| 0.3.X and lower | 1.7+               |

## Notice on CRD update (>=1.0.0)

**NOTE:** In version 1.0.0, we have updated the CRD definition and added RBAC for the
status resource. If you are upgrading from version (<=0.14.0), you must update the CRD
definition and RBAC.

```shell
kubectl apply -f https://raw.githubusercontent.com/kubernetes/autoscaler/vpa-release-1.0/vertical-pod-autoscaler/deploy/vpa-v1-crd-gen.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/autoscaler/vpa-release-1.0/vertical-pod-autoscaler/deploy/vpa-rbac.yaml
```

Another method is to re-execute the ./hack/vpa-process-yamls.sh script.

```shell
git clone https://github.com/kubernetes/autoscaler.git
cd autoscaler/vertical-pod-autoscaler
git checkout origin/vpa-release-1.0
REGISTRY=registry.k8s.io/autoscaling TAG=1.0.0 ./hack/vpa-process-yamls.sh apply
```

If you need to roll back to version (<=0.14.0), please check out the release for your
rollback version and execute ./hack/vpa-process-yamls.sh. For example, to rollback to 0.14.0:

```shell
git checkout origin/vpa-release-0.14
REGISTRY=registry.k8s.io/autoscaling TAG=0.14.0 ./hack/vpa-process-yamls.sh apply
kubectl delete clusterrole system:vpa-status-actor
kubectl delete clusterrolebinding system:vpa-status-actor
```

## Notice on deprecation of v1beta2 version (>=0.13.0)

**NOTE:** In 0.13.0 we deprecate `autoscaling.k8s.io/v1beta2` API. We plan to
remove this API version. While for now you can continue to use `v1beta2` API we
recommend using `autoscaling.k8s.io/v1` instead. `v1` and `v1beta2` APIs are
almost identical (`v1` API has some fields which are not present in `v1beta2`)
so simply changing which API version you're calling should be enough in almost
all cases.

## Notice on removal of v1beta1 version (>=0.5.0)

**NOTE:** In 0.5.0 we disabled the old version of the API - `autoscaling.k8s.io/v1beta1`.
The VPA objects in this version will no longer receive recommendations and
existing recommendations will be removed. The objects will remain present though
and a ConfigUnsupported condition will be set on them.

This doc is for installing latest VPA. For instructions on migration from older versions see [Migration Doc](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/MIGRATE.md).

## Prerequisites

- `kubectl` should be connected to the cluster you want to install VPA.
- The metrics server must be deployed in your cluster. Read more about [Metrics Server](https://github.com/kubernetes-sigs/metrics-server).
- If you are using a GKE Kubernetes cluster, you will need to grant your current Google
  identity `cluster-admin` role. Otherwise, you won't be authorized to grant extra
  privileges to the VPA system components.

  ```console
  $ gcloud info | grep Account    # get current google identity
  Account: [myname@example.org]

  $ kubectl create clusterrolebinding myname-cluster-admin-binding --clusterrole=cluster-admin --user=myname@example.org
  Clusterrolebinding "myname-cluster-admin-binding" created
  ```

- If you already have another version of VPA installed in your cluster, you have to tear down
  the existing installation first with:

  ```console
  ./hack/vpa-down.sh
  ```

## Install command

To install VPA, please download the source code of VPA (for example with `git clone https://github.com/kubernetes/autoscaler.git`)
and run the following command inside the `vertical-pod-autoscaler` directory:

```console
./hack/vpa-up.sh
```

Note: the script currently reads environment variables: `$REGISTRY` and `$TAG`.
Make sure you leave them unset unless you want to use a non-default version of VPA.

Note: If you are seeing following error during this step:

```console
unknown option -addext
```

please upgrade openssl to version 1.1.1 or higher (needs to support -addext option) or use ./hack/vpa-up.sh on the [0.8 release branch](https://github.com/kubernetes/autoscaler/tree/vpa-release-0.8).

The script issues multiple `kubectl` commands to the
cluster that insert the configuration and start all needed pods (see
[architecture](https://github.com/kubernetes/design-proposals-archive/blob/main/autoscaling/vertical-pod-autoscaler.md#architecture-overview))
in the `kube-system` namespace. It also generates
and uploads a secret (a CA cert) used by VPA Admission Controller when communicating
with the API server.

To print YAML contents with all resources that would be understood by
`kubectl diff|apply|...` commands, you can use

```console
./hack/vpa-process-yamls.sh print
```

The output of that command won't include secret information generated by
[pkg/admission-controller/gencerts.sh](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler/pkg/admission-controller/gencerts.sh) script.

## Quick start

After [installation](#installation) the system is ready to recommend and set
resource requests for your pods.
In order to use it, you need to insert a *Vertical Pod Autoscaler* resource for
each controller that you want to have automatically computed resource requirements.
This will be most commonly a **Deployment**.
There are four modes in which *VPAs* operate:

- `"Auto"`: VPA assigns resource requests on pod creation as well as updates
  them on existing pods using the preferred update mechanism. Currently, this is
  equivalent to `"Recreate"` (see below). Once restart free ("in-place") update
  of pod requests is available, it may be used as the preferred update mechanism by
  the `"Auto"` mode.
- `"Recreate"`: VPA assigns resource requests on pod creation as well as updates
  them on existing pods by evicting them when the requested resources differ significantly
  from the new recommendation (respecting the Pod Disruption Budget, if defined).
  This mode should be used rarely, only if you need to ensure that the pods are restarted
  whenever the resource request changes. Otherwise, prefer the `"Auto"` mode which may take
  advantage of restart-free updates once they are available.
- `"Initial"`: VPA only assigns resource requests on pod creation and never changes them
  later.
- `"Off"`: VPA does not automatically change the resource requirements of the pods.
  The recommendations are calculated and can be inspected in the VPA object.

## Test your installation

A simple way to check if Vertical Pod Autoscaler is fully operational in your
cluster is to create a sample deployment and a corresponding VPA config:

```console
kubectl create -f examples/hamster.yaml
```

The above command creates a deployment with two pods, each running a single container
that requests 100 millicores and tries to utilize slightly above 500 millicores.
The command also creates a VPA config pointing at the deployment.
VPA will observe the behaviour of the pods, and after about 5 minutes, they should get
updated with a higher CPU request
(note that VPA does not modify the template in the deployment, but the actual requests
of the pods are updated). To see VPA config and current recommended resource requests run:

```console
kubectl describe vpa
```

*Note: if your cluster has little free capacity these pods may be unable to schedule.
You may need to add more nodes or adjust examples/hamster.yaml to use less CPU.*

## Example VPA configuration

```yaml
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

## Troubleshooting

To diagnose problems with a VPA installation, perform the following steps:

- Check if all system components are running:

```console
kubectl --namespace=kube-system get pods|grep vpa
```

The above command should list 3 pods (recommender, updater and admission-controller)
all in state Running.

- Check if the system components log any errors.
  For each of the pods returned by the previous command do:

```console
kubectl --namespace=kube-system logs [pod name] | grep -e '^E[0-9]\{4\}'
```

- Check that the VPA Custom Resource Definition was created:

```console
kubectl get customresourcedefinition | grep verticalpodautoscalers
```

## Components of VPA

The project consists of 3 components:

- [Recommender](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/recommender/README.md) - monitors the current and past resource consumption and, based on it,
  provides recommended values for the containers' cpu and memory requests.

- [Updater](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/updater/README.md) - checks which of the managed pods have correct resources set and, if not,
  kills them so that they can be recreated by their controllers with the updated requests.

- [Admission Plugin](https://github.com/kubernetes/autoscaler/blob/master/vertical-pod-autoscaler/pkg/admission-controller/README.md) - sets the correct resource requests on new pods (either just created
  or recreated by their controller due to Updater's activity).

More on the architecture can be found [HERE](https://github.com/kubernetes/design-proposals-archive/blob/main/autoscaling/vertical-pod-autoscaler.md).

## Tear down

Note that if you stop running VPA in your cluster, the resource requests
for the pods already modified by VPA will not change, but any new pods
will get resources as defined in your controllers (i.e. deployment or
replicaset) and not according to previous recommendations made by VPA.

To stop using Vertical Pod Autoscaling in your cluster:

- If running on GKE, clean up role bindings created in [Prerequisites](#prerequisites):

```console
kubectl delete clusterrolebinding myname-cluster-admin-binding
```

- Tear down VPA components:

```console
./hack/vpa-down.sh
```
