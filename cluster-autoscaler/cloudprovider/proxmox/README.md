# Cluster Autoscaler for Proxmox VE

The Cluster Autoscaler for Proxmox VE scales Kubernetes worker nodes running as VMs in a Proxmox VE cluster.

## Configuration

The autoscaler requires a configuration file with the following structure:

```json
{
  "api_endpoint": "https://proxmox.example.com:8006/api2/json",
  "username": "root@pam",
  "password": "your-password",
  "insecure_skip_tls_verify": false,
  "node_groups": [
    {
      "name": "worker-pool-1",
      "min_size": 1,
      "max_size": 10,
      "proxmox_node": "pve-node1",
      "template_id": 9000,
      "vmid_start": 200,
      "vmid_end": 299,
      "vm_config": {
        "cores": 2,
        "memory": 2048,
        "storage": "local-lvm",
        "network": "vmbr0"
      }
    }
  ]
}
```

### Configuration Parameters

- **api_endpoint**: Proxmox VE API endpoint URL
- **username**: Proxmox VE username (e.g., "root@pam")
- **password**: Proxmox VE password
- **token_id**: Alternative to username/password - API token ID
- **token_secret**: Alternative to username/password - API token secret
- **insecure_skip_tls_verify**: Skip TLS certificate verification (for self-signed certs)

### Node Group Configuration

- **name**: Unique identifier for the node group
- **min_size**: Minimum number of nodes in this group
- **max_size**: Maximum number of nodes in this group
- **proxmox_node**: Proxmox cluster node where VMs should be created
- **template_id**: VM template ID to clone from
- **vmid_start**: Starting VM ID range for this node group
- **vmid_end**: Ending VM ID range for this node group
- **vm_config**: VM configuration parameters
  - **cores**: Number of CPU cores
  - **memory**: Memory in MB
  - **storage**: Storage pool name
  - **network**: Network bridge name

## Prerequisites

1. **VM Template**: Create a VM template in Proxmox with:
   - Cloud-init support
   - Kubernetes node components pre-installed
   - Proper SSH key configuration

2. **Proxmox User**: Create a dedicated user with appropriate permissions:
   - VM.Allocate
   - VM.Clone
   - VM.Config.Disk
   - VM.Config.CPU
   - VM.Config.Memory
   - VM.Monitor
   - VM.PowerMgmt

## Usage

### Building with Proxmox Support

```bash
cd cluster-autoscaler
make build-in-docker CLOUD_PROVIDER=proxmox
```

### Deployment

1. Create a secret with your Proxmox configuration:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: proxmox-config
  namespace: kube-system
data:
  config.json: <base64-encoded-config>
```

2. Deploy the cluster autoscaler with Proxmox provider:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-autoscaler
  namespace: kube-system
spec:
  template:
    spec:
      containers:
      - image: k8s.gcr.io/autoscaling/cluster-autoscaler:latest
        name: cluster-autoscaler
        command:
        - ./cluster-autoscaler
        - --v=4
        - --stderrthreshold=info
        - --cloud-provider=proxmox
        - --cloud-config=/etc/config/config.json
        - --nodes=1:10:worker-pool-1
        volumeMounts:
        - name: config
          mountPath: /etc/config
          readOnly: true
      volumes:
      - name: config
        secret:
          secretName: proxmox-config
```

## Implementation Status

This is a basic implementation that provides the foundation for Proxmox VE integration. Current features:

- ✅ Basic cloud provider interface implementation
- ✅ Node group management
- ✅ Configuration parsing
- ⚠️ API client (basic structure, needs full Proxmox API implementation)
- ⚠️ VM lifecycle management (create, delete, start, stop)
- ⚠️ Node discovery and status reporting

## TODO

- [ ] Complete Proxmox API client implementation
- [ ] Add proper authentication handling (tokens vs username/password)
- [ ] Implement VM creation from templates
- [ ] Add node discovery based on VM tags
- [ ] Implement proper error handling and retry logic
- [ ] Add unit tests
- [ ] Add integration tests
- [ ] Improve logging and metrics
- [ ] Add support for multiple Proxmox nodes
- [ ] Add support for different storage backends

## Contributing

This implementation is a starting point. Contributions are welcome to complete the full Proxmox VE API integration.
