# Cluster Autoscaler for DigitalOcean

The cluster autoscaler for DigitalOcean scales worker nodes within any
specified DigitalOcean Kubernetes cluster's node pool. This is part of the DOKS
offering which can be enabled/disable dynamically for an existing cluster.

# Configuration

The `cluster-autoscaler` dynamically runs based on tags associated with node
pools. These are the current valid tags:

```
k8s-cluster-autoscaler-enabled:true
k8s-cluster-autoscaler-min:3
k8s-cluster-autoscaler-max:10
```

The syntax is in form of `key:value`.

* If `k8s-cluster-autoscaler-enabled:true` is absent or
  `k8s-cluster-autoscaler-enabled` is **not** set to `true`, the
  `cluster-autoscaler` will not process the node pool by default.
* To set the minimum number of nodes to use `k8s-cluster-autoscaler-min`
* To set the maximum number of nodes to use `k8s-cluster-autoscaler-max`


If you don't set the minimum and maximum tags, node pools will have the
following default limits:

```
minimum number of nodes: 1
maximum number of nodes: 200
```

# Development

Make sure you're inside the root path of the [autoscaler
repository](https://github.com/kubernetes/autoscaler)

1.) Build the `cluster-autoscaler` binary:


```
make build-in-docker
```

2.) Build the docker image:

```
docker build -t digitalocean/cluster-autoscaler:dev .
```


3.) Push the docker image to Docker hub:

```
docker push digitalocean/cluster-autoscaler:dev
```
