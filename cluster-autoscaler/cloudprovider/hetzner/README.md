# Cluster Autoscaler for Hetzner Cloud

The cluster autoscaler for Hetzner Cloud scales worker nodes.

# Configuration

`HCLOUD_TOKEN` Required Hetzner Cloud token.
`HCLOUD_CLOUD_INIT` Base64 encoded Cloud Init yaml with commands to join the cluster
`HCLOUD_IMAGE` Defaults to `ubuntu-20.04`, @see https://docs.hetzner.cloud/#images

Node groups must be defined with the `--nodes=<min-servers>:<max-servers>:<instance-type>:<region>:<name>` flag.
Multiple flags will create multiple node pools. For example:
```
--nodes=1:10:CPX51:FSN1:pool1
--nodes=1:10:CPX51:NBG1:pool2
--nodes=1:10:CX41:NBG1:pool3
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
docker build -t hetzner/cluster-autoscaler:dev .
```


3.) Push the docker image to Docker hub:

```
docker push hetzner/cluster-autoscaler:dev
```
