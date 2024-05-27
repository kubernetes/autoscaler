# Cluster Autoscaler for Hetzner Cloud

The cluster autoscaler for Hetzner Cloud scales worker nodes.

## Configuration

`HCLOUD_TOKEN` Required Hetzner Cloud token.

`HCLOUD_CLOUD_INIT` Base64 encoded Cloud Init yaml with commands to join the cluster, Sample [examples/cloud-init.txt for (Kubernetes 1.20.1)](examples/cloud-init.txt)

`HCLOUD_IMAGE` Defaults to `ubuntu-20.04`, @see https://docs.hetzner.cloud/#images. You can also use an image ID here (e.g. `15512617`), or a label selector associated with a custom snapshot (e.g. `customized_ubuntu=true`). The most recent snapshot will be used in the latter case.

`HCLOUD_CLUSTER_CONFIG` This is the new format replacing 
 * `HCLOUD_CLOUD_INIT` 
 * `HCLOUD_IMAGE` 
 
 Base64 encoded JSON according to the following structure

```json
{
    "imagesForArch": { // These should be the same format as HCLOUD_IMAGE
        "arm64": "", 
        "amd64": ""
    },
    "nodeConfigs": {
        "pool1": { // This equals the pool name. Required for each pool that you have
            "cloudInit": "", // HCLOUD_CLOUD_INIT make sure it isn't base64 encoded twice ;]
            "labels": {
                "node.kubernetes.io/role": "autoscaler-node"
            },
            "taints": 
            [
                {
                    "key": "node.kubernetes.io/role",
                    "value": "autoscaler-node",
                    "effect": "NoExecute"
                }
            ]
        }
    }
}
```


`HCLOUD_NETWORK` Default empty , The id or name of the network that is used in the cluster , @see https://docs.hetzner.cloud/#networks

`HCLOUD_FIREWALL` Default empty , The id or name of the firewall that is used in the cluster , @see https://docs.hetzner.cloud/#firewalls

`HCLOUD_SSH_KEY` Default empty , The id or name of SSH Key that will have access to the fresh created server, @see https://docs.hetzner.cloud/#ssh-keys

`HCLOUD_PUBLIC_IPV4` Default true , Whether the server is created with a public IPv4 address or not, @see https://docs.hetzner.cloud/#primary-ips

`HCLOUD_PUBLIC_IPV6` Default true , Whether the server is created with a public IPv6 address or not, @see https://docs.hetzner.cloud/#primary-ips

Node groups must be defined with the `--nodes=<min-servers>:<max-servers>:<instance-type>:<region>:<name>` flag.

Multiple flags will create multiple node pools. For example:
```
--nodes=1:10:CPX51:FSN1:pool1
--nodes=1:10:CPX51:NBG1:pool2
--nodes=1:10:CX41:NBG1:pool3
```

You can find a deployment sample under [examples/cluster-autoscaler-run-on-master.yaml](examples/cluster-autoscaler-run-on-master.yaml). Please be aware that you should change the values within this deployment to reflect your cluster.

## Development

Make sure you're inside the `cluster-autoscaler` root folder.

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

### Updating vendored hcloud-go

To update the vendored `hcloud-go` code, navigate to the directory and run the `hack/update-vendor.sh` script:

```
cd cluster-autoscaler/cloudprovider/hetzner
UPSTREAM_REF=v2.0.0 hack/update-vendor.sh
git add hcloud-go/
```

## Debugging

To enable debug logging, set the log level of the autoscaler to at least level 5 via cli flag: `--v=5`  
The logs will include all requests and responses made towards the Hetzner API including headers and body.