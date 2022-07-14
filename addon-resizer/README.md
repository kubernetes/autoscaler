# addon-resizer

This container image watches over another container in a deployment, and
vertically scales the dependent container up and down. Currently the only
option is to scale it linearly based on the number of nodes, and it only works
for a singleton.

Currently recommended version is 1.8, on addon-resizer-release-1.8 branch. The latest version and Docker images are 2.3 pushed to:

* gcr.io/google-containers/addon-resizer-amd64:2.3
* gcr.io/google-containers/addon-resizer-arm64:2.3
* gcr.io/google-containers/addon-resizer-arm:2.3

## Nanny program and arguments

The nanny scales resources linearly with the number of nodes in the cluster. The base and marginal resource requirements are given as command line arguments, but you cannot give a marginal requirement without a base requirement.

The cluster size is periodically checked, and used to calculate the expected resources. If the expected and actual resources differ by more than the threshold (given as a +/- percent), then the deployment is updated (updating a deployment stops the old pod, and starts a new pod).

```
Usage of ./pod_nanny:
      --acceptance-offset=20: A number from range 0-100. The dependent's resources are rewritten when they deviate from expected by a percentage that is higher than this threshold. Can't be lower than recommendation-offset.
      --alsologtostderr[=false]: log to standard error as well as files
      --container="pod-nanny": The name of the container to watch. This defaults to the nanny itself.
      --cpu="MISSING": The base CPU resource requirement.
      --deployment="": The name of the deployment being monitored. This is required.
      --extra-cpu="0": The amount of CPU to add per node.
      --extra-memory="0Mi": The amount of memory to add per node.
      --extra-storage="0Gi": The amount of storage to add per node.
      --log-flush-frequency=5s: Maximum number of seconds between log flushes
      --log_backtrace_at=:0: when logging hits line file:N, emit a stack trace
      --log_dir="": If non-empty, write log files in this directory
      --logtostderr[=true]: log to standard error instead of files
      --memory="MISSING": The base memory resource requirement.
      --namespace="": The namespace of the ward. This defaults to the nanny pod's own namespace.
      --pod="": The name of the pod to watch. This defaults to the nanny's own pod.
      --poll-period=10000: The time, in milliseconds, to poll the dependent container.
      --recommendation-offset=10: A number from range 0-100. When the dependent's resources are rewritten, they are set to the closer end of the range defined by this percentage threshold.
      --stderrthreshold=2: logs at or above this threshold go to stderr
      --storage="MISSING": The base storage resource requirement.
      --v=0: log level for V logs
      --vmodule=: comma-separated list of pattern=N settings for file-filtered logging
```

## Example deployment file

You can take a look at an [example deployment](./deploy/example.yaml) where the nanny watches and resizes itself.

## Addon resizer configuration

To follow instructions in this section, set environment variable `ADDON_NAME` to
name of your addon, for example:

```
ADDON_NAME=heapster
```

Currently Addon Resizer is used to scale addons: `heapster`, `metrics-server`, `kube-state-metrics`.

### Overview

Some addons are scaled by Addon Resizer running as a sidecar container. The default
configuration is passed to Addon Resizer via command-line flags and consists of
parameters:

1. CPU parameters:
  ```
  --cpu
  --extra-cpu
  ```

  On n-node cluster, the total CPU assigned to the addon will be computed as:
  ```
  cpu + n * extra-cpu
  ```

  *Note: Addon Resizer uses buckets of cluster sizes, so it will use n larger
  than the cluster size by up to 50% for clusters larger than 16 nodes. For
  smaller clusters, n = 16 will be used.*
 
2. Memory parameters:
  ```
  --memory
  --extra-memory
  ```

  On n-node cluster, the total memory assigned to the addon will be computed as:
  ```
  memory + n * extra-memory
  ```

  *Note: Addon Resizer uses buckets of cluster sizes, so it will use n larger
  than the cluster size by up to 50% for clusters larger than 16 nodes. For
  smaller clusters, n = 16 will be used.*

These resources are overwritten by analogous values specified in a ConfigMap
`$ADDON_NAME-config` in kube-system namespace. By default the ConfigMap is empty.

### View current defaults

Find version of addon running in your cluster:
```
kubectl get deployments -n kube-system -l k8s-app=$ADDON_NAME -L version

# set env variables to received values, for example:
ADDON_DEPLOYMENT=heapster-v1.5.0
ADDON_VERSION=v1.5.0
```

Find the resource parameters for version of the addon running in your clusters,
for example:

```
kubectl logs -n kube-system -l k8s-app=$ADDON_NAME -l version=$ADDON_VERSION -c
$ADDON_NAME-nanny | head -n 1
```

You can also see total values computed for addon container CPU and memory
requirements by inspecting full deployment specification:
```
kubectl get deployment -n kube-system $ADDON_DEPLOYMENT -o yaml
```

### View current configuration

```
kubectl get configmap -n kube-system $ADDON_NAME-config -o yaml
```

By default the configuration is empty:

```
kind: ConfigMap
apiVersion: v1
data:
  NannyConfiguration: |-
    apiVersion: nannyconfig/v1alpha1
    kind: NannyConfiguration
metadata:
[...]
```

### Edit a configuration

Before you edit a configuration, make sure to check the default values. Don’t
overwrite addons configuration with value lower than defaults, otherwise you may
cause some Kubernetes components to stop working. If the configuration is
missing, empty or incorrect, Addon Resizer will fall back to default
configuration.

```
kubectl edit configmap -n kube-system $ADDON_NAME-config -o yaml
```

Set values for CPU and memory configuration using predefined options (they are
specified as NannyConfiguration API):

```
baseCPU
cpuPerNode
baseMemory
memoryPerNode
```

Example of edited configuration:

```
kind: ConfigMap
apiVersion: v1
data:
  NannyConfiguration: |-
    apiVersion: nannyconfig/v1alpha1
    kind: NannyConfiguration
    baseCPU: 100m
    cpuPerNode: 1m
    baseMemory: 200Mi
metadata:
[...]
```

Restart the addon. One way to do it is to delete addon deployment and wait for
controller manager to re-create it:
```
kubectl delete deployment -n kube-system $ADDON_DEPLOYMENT
```

### Reset a configuration to default values

To reset a configuration, just remove it and let it be recreated by addon
manager:

```
kubectl delete configmap -n kube-system $ADDON_NAME-config
```

Then reset the addon itself using the same method:

```
kubectl delete deployment -n kube-system $ADDON_DEPLOYMENT
```
