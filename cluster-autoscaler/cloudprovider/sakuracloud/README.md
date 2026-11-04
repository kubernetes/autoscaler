# Cluster Autoscaler for SAKURA cloud

The cluster autoscaler for [SAKURA cloud](https://cloud.sakura.ad.jp/) scales
worker nodes in a self-managed cluster. SAKURA cloud has no instance-group /
ASG primitive, so the autoscaler provisions and deletes servers directly
(disk copy from a source archive, server creation on the shared segment,
startup-note bootstrap), the same approach as the Hetzner provider.

## Configuration

Environment variables:

- `SAKURACLOUD_ACCESS_TOKEN` / `SAKURACLOUD_ACCESS_TOKEN_SECRET`: API key.
- `SAKURACLOUD_CLUSTER_CONFIG`: JSON document:

```json
{
  "zone": "is1a",
  "nodeGroups": {
    "sakura-cil": {
      "minSize": 0,
      "maxSize": 2,
      "core": 2,
      "memoryGB": 4,
      "diskGB": 20,
      "sourceArchiveID": "<ubuntu 24.04 archive ID>",
      "startupNoteID": "<startup script note ID>",
      "labels": {"cloud": "sakura", "role": "sakura-spot"},
      "taints": [{"key": "dedicated", "value": "sakura-ops", "effect": "NoSchedule"}]
    }
  }
}
```

Run with `--cloud-provider=sakuracloud`.

## Node group membership and providerID

- Servers belonging to a node group carry the tag `ca-group-<name>`.
- The providerID convention is `sakuracloud://<zone>/<serverName>`. The
  bootstrap startup note must register the kubelet with a matching
  `--provider-id`; the disk config sets the hostname to the server name, so
  the note can derive it as `sakuracloud://<zone>/$(hostname -s)`.
- Nodes with any other providerID scheme (mixed-provider clusters) are
  reported as unmanaged (`NodeGroupForNode` returns nil), per the
  CloudProvider contract.

## Scale-from-zero

`TemplateNodeInfo` advertises cpu/memory from the group config plus the
configured labels and taints, so pods with matching nodeSelector/tolerations
trigger scale-up from zero.

## Notes

- The startup note (SAKURA cloud "note", class `shell`) is executed on first
  boot and must install the kubelet/agent and join the cluster. Keep join
  credentials in the note, not in `SAKURACLOUD_CLUSTER_CONFIG`.
- Scale-down deletes the server together with its disks.

## Implementation notes (observed API behavior)

- The server plan must be specified by spec (`{"CPU": n, "MemoryMB": m}`);
  the plan-ID form returned by `/product/server` is rejected with 400 by the
  current API.
- After `PUT /disk/:id/config` (hostname + startup note injection) the disk
  transiently leaves the `available` state; powering the server on before it
  settles fails with `409 disk_is_not_available`, so the provider waits for
  the disk to become available again.
- The `/server` list response does not reliably include the instance power
  state, so deletion always force-powers-off first and tolerates a 409
  (already down) before deleting the server together with its disks.
