# Cluster Autoscaler for VKE

The cluster autoscaler for VKE scales worker nodes within a VKE Kubernetes
cluster node group by calling the VKE control-plane API.

Authentication uses OpenStack application credentials (Keystone). The API base
URL is taken from the `VKE_URL` environment variable.

## Configuration

Pass a JSON cloud config with `--cloud-config`:

```json
{
  "cluster_id": "<cluster-uuid>",
  "tenant_id": "<openstack-project-id>",
  "application_key": "<application-credential-id>",
  "application_secret": "<application-credential-secret>",
  "openstack_auth_url": "https://example.com:5000/v3"
}
```

| Field | Description |
| --- | --- |
| `cluster_id` | VKE cluster UUID |
| `tenant_id` | OpenStack project / tenant ID |
| `application_key` | OpenStack application credential ID |
| `application_secret` | OpenStack application credential secret |
| `openstack_auth_url` | Keystone auth URL |

Also set:

```bash
export VKE_URL=https://vke-api.example.com
```

Example manifests live under [`examples/`](./examples).

Apply order:

```bash
kubectl apply -f examples/cluster-autoscaler-secret.yaml
kubectl apply -f examples/cluster-autoscaler-deployment.yaml
```

Replace placeholders in the secret before applying. Use your own Cluster Autoscaler image that includes the VKE provider (or build with `BUILD_TAGS=vke`).

## Behavior

- Node groups (node pools) are discovered from the VKE API on each `Refresh()`.
- Scale-up adds nodes with the VKE nodegroup add API.
- Scale-down deletes specific nodes with the VKE nodegroup node delete API.
- Node `ProviderID` values are expected in `openstack:///<instance-id>` form.
- Node group auto-provisioning is not supported yet.

## Development

From the `cluster-autoscaler` directory:

```bash
# Full binary (includes VKE among other providers)
make build

# VKE-only binary
BUILD_TAGS=vke make build

# Unit tests for this provider
go test ./cloudprovider/vke/...
```

Build an image:

```bash
make build-in-docker
docker build -t <your-registry>/cluster-autoscaler:dev .
```
