# Cluster Autoscaler for Hetzner Cloud

The cluster autoscaler for Hetzner Cloud scales worker nodes.

# Configuration

`HCLOUD_TOKEN` Required Hetzner Cloud token.

`HCLOUD_CLOUD_INIT` Base64 encoded Cloud Init yaml with commands to join the cluster, Sample [examples/cloud-init.txt for (Kubernetes 1.20.1)](examples/cloud-init.txt)

`HCLOUD_{POOL_NAME}_CLOUD_INIT` Base64 encoded Cloud Init yaml. Allows for specifying a different cloud init per pool name. `{POOL_NAME}` should be the uppercase equivalent of the node group's name. e.g. `--nodes=0:2:cpx41:fsn1:test-pool` would look for `HCLOUD_TEST-POOL_CLOUD_INIT`. If not set it will default to `HCLOUD_CLOUD_INIT`.

`HCLOUD_IMAGE` Defaults to `ubuntu-20.04`, @see https://docs.hetzner.cloud/#images. You can also use an image ID here (e.g. `15512617`), or a label selector associated with a custom snapshot (e.g. `customized_ubuntu=true`). The most recent snapshot will be used in the latter case.

`HCLOUD_NETWORK` Default empty , The name of the network that is used in the cluster , @see https://docs.hetzner.cloud/#networks

`HCLOUD_FIREWALL` Default empty , The name of the firewall that is used in the cluster , @see https://docs.hetzner.cloud/#firewalls

`HCLOUD_SSH_KEY` Default empty , This SSH Key will have access to the fresh created server, @see https://docs.hetzner.cloud/#ssh-keys

`HCLOUD_PUBLIC_IPV4` Default true , Whether the server is created with a public IPv4 address or not, @see https://docs.hetzner.cloud/#primary-ips

`HCLOUD_{POOL_NAME}_PUBLIC_IPV4` is the pool name equivalent of `HCLOUD_PUBLIC_IPV4` and follows the same rules as `HCLOUD_CLOUD_INIT` above.

`HCLOUD_PUBLIC_IPV6` Default true , Whether the server is created with a public IPv6 address or not, @see https://docs.hetzner.cloud/#primary-ips

`HCLOUD_{POOL_NAME}_PUBLIC_IPV6` is the pool name equivalent of `HCLOUD_PUBLIC_IPV6` and follows the same rules as `HCLOUD_CLOUD_INIT` above.

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
