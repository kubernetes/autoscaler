# Cluster Autoscaler for DigitalOcean

The cluster autoscaler for DigitalOcean scales worker nodes within any
specified DigitalOcean Kubernetes cluster's node pool. This is part of the DOKS
offering which can be enabled/disable dynamically for an existing cluster.

# Configuration

## Cloud config file

The (JSON) configuration file of the DigitalOcean cloud provider supports the
following values:

- `cluster_id`: the ID of the cluster (a UUID)
- `token`: the DigitalOcean access token literally defined
- `token_file`: a file path containing the DigitalOcean access token
- `url`: the DigitalOcean URL (optional; defaults to `https://api.digitalocean.com/`)

Exactly one of `token` or `token_file` must be provided.

## Behavior

Parameters of the autoscaler (such as whether it is on or off, and the
minimum/maximum values) are configured through the public DOKS API and
subsequently reflected by the node pool objects. The cloud provider periodically
picks up the configuration from the API and adjusts the behavior accordingly.

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
