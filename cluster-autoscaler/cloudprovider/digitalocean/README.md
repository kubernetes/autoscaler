# Cluster Autoscaler for DigitalOcean

The cluster autoscaler for DigitalOcean scales worker nodes within any
specified DigitalOcean Kubernetes cluster's node pool. This is part of the DOKS
offering which can be enabled/disable dynamically for an existing cluster.

# Configuration

The `cluster-autoscaler` accepts the `--nodes` flag in the format of:

```
--nodes "1,10,foo" 
```

This is a comma separated value, in the format of: `min,max,tagValue` with the
following meanings:

```
  1: minimum number of nodes
  5: maximum number of nodes
foo: tag value assigned for the node pool at DigitalOcean
```

This can be defined multiple times for multiple node pools:

```
--nodes "1,10,foo" --nodes "5,20,bar" 
```

If you don't set this flag, all node pools will have the following limits:

```
minimum number of nodes: 1
maximum number of nodes: 200
```

In order to pick up the correct node pool, it needs to have a tag associated in
the form of: `k8s-cluster-autoscaler:value`. For example if you have specified:


```
--nodes "1,10,foo" 
```

Make sure your node-pool has the following tag associated with it:

```
k8s-cluster-autoscaler:foo
```

# Development

Make sure you're inside the root path of the [autoscaler
repository](https://github.com/kubernetes/autoscaler)

1.) Build the `cluster-autoscaler` binary:


```
make build-binary
```

2.) Build the docker image:

```
docker build -t digitalocean/cluster-autoscaler:dev .
```


3.) Push the docker image to Docker hub:

```
docker push digitalocean/cluster-autoscaler:dev
```
