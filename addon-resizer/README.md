# addon-resizer

This container image watches over another container in a deployment, and
vertically scales the dependent container up and down. Currently supported
options are linear and exponential scaling. Addon-resizer only works for a singleton.

## Nanny program and arguments

The nanny scales resources linearly with the number of nodes or containers in the cluster. The base and marginal resource requirements are given as command line arguments, but you cannot give a marginal requirement without a base requirement.

The cluster size is periodically checked, and used to calculate the expected resources. If the expected and actual resources differ by more than the threshold (given as a +/- percent), then the deployment is updated (updating a deployment stops the old pod, and starts a new pod).

```
Usage of pod_nanny:
      --config-dir="": The name of directory used to specify resources for scaled container.
      --container="pod-nanny": The name of the container to watch. This defaults to the nanny itself.
      --cpu="MISSING": The base CPU resource requirement.
      --deployment="": The name of the deployment being monitored. This is required.
      --estimator="linear": "The estimator to use. Currently supported: linear, exponential"
      --extra-cpu="0": The amount of CPU to add per resource (containers or nodes).
      --extra-memory="0Mi": The amount of memory to add per resource.
      --extra-storage="0Gi": The amount of storage to add per resource.
      --log-flush-frequency=5s: Maximum number of seconds between log flushes
      --memory="MISSING": The base memory resource requirement.
      --minClusterSize=16: The smallest number of resources will be scaled to. Must be > 1. This flag is used only when an exponential estimator is used.
      --namespace=$MY_POD_NAMESPACE: The namespace of the ward. This defaults to the nanny's own pod.
      --pod=$MY_POD_NAME: The name of the pod to watch. This defaults to the nanny's own pod.
      --poll-period=10000: The time, in milliseconds, to poll the dependent container.
      --storage="MISSING": The base storage resource requirement.
      --threshold=0: A number between 0-100. The dependent's resources are rewritten when they deviate from expected by more than threshold.
      --use-metrics=false: Whether to use apiserver metrics to detect cluster size instead of the default method of listing resources from the Kubernetes API.
      --expose-metrics=true: Whether to expose Prometheus metrics about addon-resizer on the port specified by the address flag.
      --address=8945: If expose-metrics is true, this is the port where we expose Prometheus metrics.
```

Estimators:
- linear - based on the current number of resources.
- exponential - based on first number in the geometric sequence (a=minClusterSize, r=1.5) greater than or equal to the current number of resources.

## Example deployment file

The [example.yaml](deploy/example.yaml) file is an example deployment where the nanny watches and resizes itself.
