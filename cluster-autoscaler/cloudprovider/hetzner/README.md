# Cluster Autoscaler for Hetzner Cloud

The cluster autoscaler for Hetzner Cloud scales worker nodes.

# Configuration

`HCLOUD_TOKEN` is required

# Development

Make sure you're inside the root path of the [autoscaler
repository](https://github.com/kubernetes/autoscaler)

1.) Build the `cluster-autoscaler` binary:


```
make build-in-docker
```

2.) Build the docker image:

```
docker build -t hetzner/cluster-autoscaler:dev .
```


3.) Push the docker image to Docker hub:

```
docker push hetzner/cluster-autoscaler:dev
```
