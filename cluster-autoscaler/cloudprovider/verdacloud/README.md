# Cluster Autoscaler for VerdaCloud

The cluster autoscaler for VerdaCloud (formerly DataCrunch.io) scales worker nodes.

## Configuration

`VERDA_CLIENT_ID` Required VerdaCloud OAuth2 client ID.

`VERDA_CLIENT_SECRET` Required VerdaCloud OAuth2 client secret.

`VERDA_BASE_URL` Optional VerdaCloud API base URL. Defaults to `https://api.verda.com/v1`.

`VERDA_CLUSTER_CONFIG` Base64 encoded JSON according to the following structure:

```json
{
  "image": {
    "gpu": "24.04.kubernetes1.31.1.cuda12.9.qcow2",
    "cpu": "24.04.kubernetes1.31.1.cuda12.9.qcow2"
  },
  "sshKeyIDs": ["your-ssh-key-id"],
  "billingConfig": {
    "price": "FIXED_PRICE",
    "contract": "PAY_AS_YOU_GO"
  },
  "debug": false,
  "availableLocations": ["FIN-02", "FIN-03"],
  "osVolumeSize": 50,
  "labels": ["env=production"],
  "startupScript": "base64 encoded cloud init script. refer to example config",
  "startupScriptEnv": {
    "MASTER_IP": "",
    "MASTER_PORT": "",
    "JOIN_TOKEN": "",
    "JOIN_HASH_FULL": ""
  },
  "taints": [],
  "groups": {
    "asg-name-here": {
      "labels": ["hardware=gpu", "team=ai"],
      "taints": [
        { "key": "nvidia.com/gpu", "value": "present", "effect": "NoSchedule" }
      ],
      "billingConfig": {
        "price": "FIXED_PRICE",
        "contract": "SPOT"
      },
      "osVolumeSize": 200,
      "availableLocations": ["FIN-02"]
    }
  }
}

```
## Configuration reference

The JSON above is the authoritative format. This table summarizes top‑level keys for quick reference:

| Key | Type | Required | Default | Notes |
|-----|------|----------|---------|-------|
| image.gpu | string | required | — | Image for GPU nodes provided by VerdaCloud (current: k8s 1.31.1, cuda 12.9). Contact us for other versions. |
| image.cpu | string | required | — | Image for CPU nodes provided by VerdaCloud (current: k8s 1.31.1, cuda 12.9). Contact us for other versions. |
| sshKeyIDs | array<string> | required | — | SSH key IDs to inject. See [Fetching SSH Keys](#fetching-ssh-keys) for details. |
| billingConfig.price | string | required | — | Must be `FIXED_PRICE`. |
| billingConfig.contract | string | optional | — | One of LONG_TERM, PAY_AS_YOU_GO, or SPOT |
| debug | bool | optional | false | Enables additional provider‑side diagnostics |
| availableLocations | array<string> | required | — | Location codes eligible for provisioning. Group config overwrites global. |
| osVolumeSize | int | optional | 50 | Size of the OS volume in GB. Group config overwrites global. |
| labels | array<string> | optional | — | Labels to apply to all nodes. Group config merges with global. |
| startupScript | string (base64) | required | — | Base64‑encoded startup script. Use the default script provided in `examples/config.json`. |
| startupScriptEnv | map<string,string> | required | — | Environment variables for the startup script. Must align with `startupScript`. |
| taints | array<object> | optional | — | Standard k8s taint objects applied to nodes |
| groups | map<string,object> | optional | — | Node group definitions overriding defaults. See below for supported keys. |

### Group Configuration Overrides
The `groups` map allows defining overrides for specific Auto Scaling Groups (ASGs). The format is `"asg-name": { ... }`.
Supported override keys within a group object:
- `labels`: Merges with global labels. Overwrites value if conflict with global label.
- `taints`: Merges with global group. Overwrites value if conflict with global.
- `availableLocations`: Overwrites global available locations.
- `billingConfig`: Overwrites global billing config.
- `osVolumeSize`: Overwrites global OS volume size.


`VERDA_CLUSTER_CONFIG_FILE` Can be used as alternative to `VERDA_CLUSTER_CONFIG`. This is the path to a file containing the JSON structure described above. The file will be read and the contents will be used as the configuration.

**NOTE**: In contrast to `VERDA_CLUSTER_CONFIG`, this file is not base64 encoded.

## Helper Commands

### Fetching SSH Keys

You can fetch your SSH keys using the following command:

```bash
curl https://api.verda.com/v1/sshkeys \
  --header 'Authorization: Bearer YOUR_SECRET_TOKEN'
```

### Fetching Images

To find available images for the `image.gpu` and `image.cpu` configuration fields:

```bash
curl https://api.verda.com/v1/images \
  --header 'Authorization: Bearer YOUR_SECRET_TOKEN'
```

Look for the `image_type` value in the response to use in your configuration.

### Fetching Instance Types

To find the correct `instance-type` for the `--nodes` flag, you can fetch the available instance types:

```bash
curl https://api.verda.com/v1/instance-types \
  --header 'Authorization: Bearer YOUR_SECRET_TOKEN'
```

Look for the `instance_type` value in the response to use in your configuration.

Node groups must be defined with the `--nodes=<min-servers>:<max-servers>:<instance-type>:<asg-name>[:<hostname-prefix>]` flag. See [Fetching Instance Types](#fetching-instance-types) to find valid `instance-type` values.

The `hostname-prefix` parameter is optional. If provided, it will be used as the base name for generated hostnames instead of the ASG name.

Multiple flags will create multiple node pools. For example:
```
--nodes=1:5:1A6000.10V:as-test-a6000
--nodes=0:10:CPU.4V.16G:cpu-workers
--nodes=1:3:1H100.20V:gpu-h100-pool
--nodes=1:5:1A100.22V:as-test-1a10022v:custom-node
```

The last example uses a custom hostname prefix `custom-node`, so instances will be named like `custom-node-vm-fin-03-42` instead of `as-test-1a10022v-vm-fin-03-42`.

You can find a complete deployment sample under [examples/cluster-autoscaler-deployment-example.yaml](examples/cluster-autoscaler-deployment-example.yaml). This single file contains all required Kubernetes resources including namespace, RBAC, secrets, configmap, and deployment. Please be aware that you should change the values within this deployment to reflect your cluster:

- Replace `your-client-id` and `your-client-secret` in the Secret
- Update `your-ssh-key-id` in the ConfigMap
- Configure your cluster join parameters (`MASTER_IP`, `JOIN_TOKEN`, `JOIN_HASH_FULL`)
- Modify the `--nodes` flags to match your desired instance types and scaling limits
- Update the startup script with your actual base64-encoded cluster join script

## Development

Make sure you're inside the `cluster-autoscaler` root folder.

1.) Build the docker image:

```bash
make make-image BUILD_TAGS=verdacloud TAG='dev' REGISTRY='verdacloud'
```

2.) Push the docker image:

```bash
make push-image BUILD_TAGS=verdacloud TAG='dev' REGISTRY='verdacloud'
```

**Note:** The `make-image` command automatically builds the code inside Docker, so no separate build step is needed.

## Support and caveats

- Hostname format: Instances created by this provider include an internal magic separator in their hostname that encodes the ASG name or custom hostname prefix (format: `{hostname-prefix|asg-name}-vm-{location-lowercase}-{random-2-digits}`). The autoscaler relies on this to identify group membership. Examples: `custom-node-vm-fin-03-42` or `as-test-1a100-vm-fin-03-87`.
- No legacy fallback: If instances are created outside this provider with different hostname conventions, they may not be associated with the expected ASG by the autoscaler.
- ProviderID format: `verdacloud://<location>/<hostname>`.

## Debugging

To enable debug logging, run the autoscaler with `--v=4` or higher.
At `--v=7` the autoscaler logs CreateInstance request bodies for troubleshooting; response bodies and headers are not logged.