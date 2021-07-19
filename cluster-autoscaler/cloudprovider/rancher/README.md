# Cluster Autoscaler for Rancher

The cluster autoscaler for Rancher scales nodes within any specified Rancher Kubernetes Engine cluster's node pool.

# Requirements

Rancher version >= 2.5.6

# Configuration

The cluster-autoscaler for Rancher needs a configuration file to work by using --cloud-config parameter.

Here an [example](examples/autoscaler-config-example.yaml).

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cluster-autoscaler-cloud-config
  namespace: kube-system
type: Opaque
stringData:
  cloud-config: |-
    [Global]
    url=https://rancherapi.com/v3
    access=your-token
    secret=your-secret
    cluster-id=c-abcdef
  autoscaler_node_arg: "2:6:c-abcdef:np-abcde" # Your NodePool ID
```

You have to create a new API Key from your Rancher Dashboard to get the `access` and `secret` values to use the Autoscaler.

# Development

Make sure you're inside the root path of the [autoscaler repository](https://github.com/kubernetes/autoscaler)

1.) Build the `cluster-autoscaler` binary:

```
make build-in-docker
```

2.) Build the docker image:

```
docker build -t rancher/cluster-autoscaler:dev .
```

3.) Push the docker image to Docker hub:

```
docker push rancher/cluster-autoscaler:dev
```
