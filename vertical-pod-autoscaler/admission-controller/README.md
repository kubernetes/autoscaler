# VPA Admission Controller

# Running

1. You should make sure your API server supports Mutating Webhooks.
Its `--admission-control` flag should have `MutatingAdmissionWebhook` as one of
the values on the list and its `--runtime-config` flag should include
`admissionregistration.k8s.io/v1beta1=true`.
1. Generate certs by running `bash gencerts.sh`. This will use kubectl to create
   a secret in your cluster with the certs.
1. Create RBAC configuration for the admission controller pod by running
   `kubectl create -f ../deploy/admission-controller-rbac.yaml`
1. Create the pod:
   `kubectl cerate -f ../deploy/admission-controlller-deployment.yaml`.
   The first thing this will do is it will register itself with the apiserver as
   an Webhook Admission Controller and start changing resource requirements
   for pods on their creation & updates.
