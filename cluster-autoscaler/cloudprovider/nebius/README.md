# Cluster Autoscaler for Nebius AI Cloud

The cluster autoscaler for [Nebius AI Cloud](https://nebius.com/) scales
worker nodes in Nebius Managed Kubernetes (MK8S) clusters.

## Configuration

The provider is configured via a JSON config file (`--cloud-config`) or
environment variables.

| Env var | Config field | Required | Description |
|---------|-------------|----------|-------------|
| `NEBIUS_IAM_TOKEN` | `iam_token` | Yes | Nebius IAM token for authentication |
| `NEBIUS_CLUSTER_ID` | `cluster_id` | Yes | MK8S cluster ID to manage |
| `NEBIUS_PARENT_ID` | `parent_id` | Yes | Parent folder ID containing compute instances |
| — | `domain` | No | API domain (defaults to `api.eu.nebius.com`) |

### Example config file

```json
{
  "iam_token": "your-iam-token",
  "cluster_id": "your-cluster-id",
  "parent_id": "your-parent-folder-id"
}
```

## Node group discovery

All node groups in the specified cluster are auto-discovered. There is no
support for filtering by `--node-group-auto-discovery` at this time.

## How it works

- **Refresh** — Every scan interval (default 10s), the provider lists all
  node groups from the MK8S API and caches instance membership by querying
  the Compute API.
- **Scale up** — Sets the node group's target size via `FixedNodeCount`.
  The Nebius MK8S API uses a oneOf for size (`Autoscaling` or
  `FixedNodeCount`) and does not expose a desired-count field within the
  autoscaling spec, so setting a specific target requires switching to
  fixed mode.
- **Scale down** — Deletes specific compute instances via the Nebius
  Compute API and updates the node group target size.
- **Node-to-group mapping** — Uses the `nebius.com/node-group-id` label
  on Kubernetes nodes, falling back to cached compute instance provider IDs.

## Important: autoscaling mode conversion

The Nebius MK8S API uses a oneOf for node group size — either
`Autoscaling{min, max}` or `FixedNodeCount`. There is no
"desired count" field within the autoscaling spec.

**The first time the cluster autoscaler scales a node group, it will
permanently convert it from autoscaling mode to fixed mode.** This is
irreversible via the autoscaler — an operator must manually re-enable
autoscaling through the Nebius console or API. A warning is logged
when this conversion occurs.

If your node groups use Nebius-native autoscaling and you do not want
the cluster autoscaler to override it, exclude those node groups from
management.

## Limitations

- **Scale-from-zero** is not supported. `TemplateNodeInfo` is not yet
  implemented, so the autoscaler cannot simulate what a new node would
  look like. Scale-up works when at least one node exists in the group.
- **ListInstances** fetches all compute instances in the parent folder
  because the Nebius API does not support label-based filtering. Use a
  dedicated parent folder to minimize unnecessary API calls.

## Development

```bash
# Build with only the Nebius provider
go build -tags nebius ./...

# Run tests
go test ./cloudprovider/nebius/
```
