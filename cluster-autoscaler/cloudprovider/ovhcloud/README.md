# Cluster Autoscaler for OVHcloud

The cluster autoscaler for OVHclud scales worker nodes within any
OVHcloud Kubernetes cluster's node pool. This autoscaler should only watch and
offer scaling options on node pools with the `autoscaling` optional parameter enabled.

## Configuration

The `cluster-autoscaler` with OVHcloud needs a configuration file to work by using `--cloud-config` parameter.

Here is an sample:

```json
{
  "project_id": "<my_project_id>",
  "cluster_id": "<my_cluster_id>",

  "authentication_type": "consumer",

  "application_endpoint": "ovh-eu",
  "application_key": "key",
  "application_secret": "secret",
  "application_consumer_key": "consumer_key"
}
```

`cluster_id` and `project_id` can be found on your OVHcloud manager.

For application tokens, you should visit: https://api.ovh.com/createToken/

## Host specification

At OVHcloud, we offer the `cluster-autoscaler` to run on the Kubernetes cluster control-plane.

The `cluster-autoscaler` is free to use, and we recommend not to use this project unless you want to try out specific configurations.

## Environment

You should be able to find the custom resource definition with:

```
kubectl get crd nodepools.kube.cloud.ovh.com -o json

{
    "apiVersion": "apiextensions.k8s.io/v1",
    "kind": "CustomResourceDefinition",
    "metadata": {
        "name": "nodepools.kube.cloud.ovh.com",
        "selfLink": "/apis/apiextensions.k8s.io/v1/customresourcedefinitions/nodepools.kube.cloud.ovh.com",
        ...
    },
    ...
}
```

To know if your node pools auto-scaling is enabled, you can simply output the resources:

```
kubectl get nodepools

NAME             FLAVOR   AUTO-SCALING     MONTHLY BILLED   ANTI AFFINITY   AGE
nodepool-b2-7    b2-7     true             false            false           140d
...
```

You should be able to edit the auto-scaling parameters using `kubectl edit` or by requesting the OVHcloud API.

## Development

Make sure you're inside the root path of the [autoscaler repository](https://github.com/kubernetes/autoscaler)

1.) Build the `cluster-autoscaler` binary:


```
make build-in-docker
```

2.) Build the docker image:

```
docker build -t ovhcloud/cluster-autoscaler:dev .
```


3.) Push the docker image to Docker hub:

```
docker push ovhcloud/cluster-autoscaler:dev
```
