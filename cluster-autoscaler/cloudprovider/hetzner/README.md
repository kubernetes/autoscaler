# Cluster Autoscaler for Hetzner Cloud

The cluster autoscaler for Hetzner Cloud scales worker nodes.

# Configuration

`HCLOUD_TOKEN` Required Hetzner Cloud token.
`HCLOUD_CLOUD_INIT` Base64 encoded Cloud Init yaml with commands to join the cluster, Sample [examples/cloud-init.txt for (Kubernetes 1.20.1)](examples/cloud-init.txt)
`HCLOUD_IMAGE` Defaults to `ubuntu-20.04`, @see https://docs.hetzner.cloud/#images
`HCLOUD_NETWORK` Default empty , The name of the network that is used in the cluster , @see https://docs.hetzner.cloud/#networks
`HCLOUD_SSH_KEY` Default empty , This SSH Key will have access to the fresh created server, @see https://docs.hetzner.cloud/#ssh-keys
Node groups must be defined with the `--nodes=<min-servers>:<max-servers>:<instance-type>:<region>:<name>` flag.
Multiple flags will create multiple node pools. For example:
```
--nodes=1:10:CPX51:FSN1:pool1
--nodes=1:10:CPX51:NBG1:pool2
--nodes=1:10:CX41:NBG1:pool3
```

You can find a deployment sample under [examples/cluster-autoscaler-run-on-master.yaml](examples/cluster-autoscaler-run-on-master.yaml). Please be aware that you should change the values within this deployment to reflect your cluster.

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
