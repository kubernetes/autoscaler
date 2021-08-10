# VPA Admission Controller

- [Intro](#intro)
- [Running](#running)
- [Implementation](#implmentation)

## Intro

This is a binary that registers itself as a Mutating Admission Webhook
and because of that is on the path of creating all pods.
For each pod creation, it will get a request from the apiserver and it will
either decide there's no matching VPA configuration or find the corresponding
one and use current recommendation to set resource requests in the pod.

## Running

1. You should make sure your API server supports Mutating Webhooks.
Its `--admission-control` flag should have `MutatingAdmissionWebhook` as one of
the values on the list and its `--runtime-config` flag should include
`admissionregistration.k8s.io/v1beta1=true`.
To change those flags, ssh to your API Server instance, edit
`/etc/kubernetes/manifests/kube-apiserver.manifest` and restart kubelet to pick
up the changes: ```sudo systemctl restart kubelet.service```
1. Generate certs by running `bash gencerts.sh`. This will use kubectl to create
   a secret in your cluster with the certs.
1. Create RBAC configuration for the admission controller pod by running
   `kubectl create -f ../deploy/admission-controller-rbac.yaml`
1. Create the pod:
   `kubectl create -f ../deploy/admission-controller-deployment.yaml`.
   The first thing this will do is it will register itself with the apiserver as
   Webhook Admission Controller and start changing resource requirements
   for pods on their creation & updates.
1. You can specify a path for it to register as a part of the installation process
   by setting `--register-by-url=true` and passing `--webhook-address` and `--webhook-port`.

## Implementation

All VPA configurations in the cluster are watched with a lister.
In the context of pod creation, there is an incoming https request from
apiserver.
The logic to serve that request involves finding the appropriate VPA, retrieving
current recommendation from it and encodes the recommendation as a json patch to
the Pod resource.

