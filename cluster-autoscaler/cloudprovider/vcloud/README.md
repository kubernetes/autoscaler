# VCloud Provider for Cluster Autoscaler

The VCloud provider enables Kubernetes Cluster Autoscaler to automatically scale node groups in VCloud infrastructure
using VCloud NodePool APIs.

## Configuration
Embedded in worker node on deployment-process, you can shell into worker at locate /etc/config/cloud-config

### Configuration Parameters

| Parameter        | Description                               | Required |
|------------------|-------------------------------------------|----------|
| `CLUSTER_ID`     | Unique identifier for your VCloud cluster | Yes      |
| `CLUSTER_NAME`   | Human-readable name of your cluster       | Yes      |
| `MGMT_URL`       | VCloud management API URL                 | Yes      |
| `PROVIDER_TOKEN` | Authentication token for VCloud API       | Yes      |

## Deployment
Build the cluster autoscaler with VCloud support:

```bash
cd cluster-autoscaler

# Build VCloud-specific binary
go build -tags vcloud -o cluster-autoscaler-vcloud .
```

Deploy with the VCloud provider:

```bash
# Option 1: Using config file (hostPath mount)
./cluster-autoscaler-vcloud \
  --cloud-provider=vcloud \
  --cloud-config=/etc/vcloud/config \
  --kubeconfig=$HOME/.kube/config \
  --v=2 --logtostderr

# Option 2: Auto-discovery mode (uses environment variables)
./cluster-autoscaler-vcloud \
  --cloud-provider=vcloud \
  --node-group-auto-discovery=vcloud:tag=k8s.io/cluster-autoscaler/enabled \
  --kubeconfig=$HOME/.kube/config \
  --v=2 --logtostderr
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-autoscaler
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cluster-autoscaler
  template:
    metadata:
      labels:
        app: cluster-autoscaler
    spec:
      serviceAccountName: cluster-autoscaler
      containers:
      - image: cluster-autoscaler:latest
        name: cluster-autoscaler
        resources:
          limits:
            cpu: 100m
            memory: 300Mi
          requests:
            cpu: 100m
            memory: 300Mi
        command:
        - ./cluster-autoscaler
        - --v=2
        - --cloud-provider=vcloud
        - --cloud-config=/etc/config/cloud-config
        - --nodes=1:10:nodepool-name
        volumeMounts:
        - name: cloud-config
          mountPath: /etc/config
          readOnly: true
      volumes:
      - name: cloud-config
        hostPath:
          path: /etc/config
          type: Directory
```

## Features

- Auto-discovery of VCloud NodePools with autoscaling enabled
- Standard Cluster Autoscaler interfaces (CloudProvider, NodeGroup)
- Individual node deletion (follows common cloud provider patterns)
- Provider ID format: `vcloud://instance-uuid`
- VCloud-specific labels: `k8s.io.infra.vnetwork.io/*`
- Retry logic with exponential backoff
- Support for both config files and environment variables

## Scaling Operations

### Scale Up (Node Creation)
- **Method**: Pool capacity increase via `PUT /nodepools/{id}/scale`
- **Behavior**: VCloud creates new instances automatically
- **Control**: Cluster autoscaler specifies desired size, VCloud manages instance details

### Scale Down (Node Deletion)
- **Method**: Individual instance deletion via `DELETE /nodepools/{id}/machines/{instance-id}`
- **Behavior**: Precise targeting of specific nodes for removal
- **Control**: Cluster autoscaler specifies exact instances to delete

### API Payloads

**Scale Up Request:**
```json
{
  "desiredSize": 5,
  "reason": "cluster-autoscaler-scale-up",
  "async": true
}
```

**Scale Down Request:**
```json
{
  "force": false,
  "reason": "cluster-autoscaler-scale-down"
}
```

## Architecture

The VCloud provider implements a **hybrid scaling approach** that combines the best practices from other cloud providers:

### Scaling Strategy
- **Scale Up**: Uses pool capacity increase (like AWS/Azure/DigitalOcean)
- **Scale Down**: Uses individual instance deletion (like GCP/Azure)

### Benefits
- ✅ **Predictable Scale Down**: Exact control over which nodes are removed
- ✅ **Efficient Scale Up**: Let VCloud manage instance provisioning details
- ✅ **Standard Compliance**: Follows cluster-autoscaler patterns
- ✅ **Error Handling**: Comprehensive validation and rollback support

### Implementation Highlights
- **Node Ownership Validation**: Ensures nodes belong to the correct node group
- **Minimum Size Enforcement**: Prevents scaling below configured limits
- **Graceful Deletion**: Uses `force: false` for proper instance shutdown
- **Partial Failure Handling**: Logs progress when some operations succeed
- **Dual Configuration**: Supports both config files and environment variables

## Requirements

- VCloud NodePool APIs available
- NodePools with autoscaling enabled (min/max > 0) 
- Valid VCloud provider token with scaling permissions
- Network connectivity to VCloud management APIs
- API endpoints: `/nodepools`, `/nodepools/{id}`, `/nodepools/{id}/scale`, `/nodepools/{id}/machines/{machine-id}`

## Testing

### Unit Tests

Run the included unit tests to verify core functionality:

```bash
cd cluster-autoscaler
go test ./cloudprovider/vcloud/ -v
```

The test suite includes:
- Configuration parsing (INI files and environment variables)
- Node group properties and validation
- Provider ID format validation
- DeleteNodes implementation patterns
- Enhanced manager creation with error handling

### Integration Testing

```bash
# Test with hostPath config file
./cluster-autoscaler \
  --cloud-provider=vcloud \
  --cloud-config=/etc/config/cloud-config \
  --dry-run=true --v=2

# Test with environment variables
CLUSTER_ID=test CLUSTER_NAME=test MGMT_URL=https://k8s.io.infra.vnetwork.dev PROVIDER_TOKEN=test \
./cluster-autoscaler \
  --cloud-provider=vcloud \
  --dry-run=true --v=2

# Test scaling
kubectl run test-scale --image=nginx --requests=cpu=1000m --replicas=3
kubectl get nodes -w
kubectl delete deployment test-scale
```

## Configuration Setup

**For hostPath config file deployment:**

1. Create the config file on each node:
```bash
sudo mkdir -p /etc/config
sudo cat > /etc/config/cloud-config << EOF
[vCloud]
CLUSTER_ID=your-cluster-id
CLUSTER_NAME=your-cluster-name
MGMT_URL=https://k8s.io.infra.vnetwork.dev
PROVIDER_TOKEN=your-provider-token
EOF
sudo chmod 600 /etc/config/cloud-config
```

2. Ensure the config file is available on all nodes where cluster-autoscaler might run

**For environment variable deployment:**

No additional setup needed - just set the environment variables in the deployment.

## Troubleshooting

Common issues and solutions:

- **Node groups not discovered**: Verify NodePools have `min > 0` and `max > min`
- **Scale up fails**: Check provider token permissions and NodePool capacity limits
- **Scale down fails**: Verify nodes belong to the node group and minimum size constraints
- **Individual node deletion fails**: Check instance exists and is in deletable state
- **Configuration errors**: 
  - Config file: Check `/etc/config/cloud-config` exists and has correct permissions (600)
  - Environment variables: Verify all required env vars are set
- **Config file not found**: Ensure `/etc/config/cloud-config` exists on the node running cluster-autoscaler
- **Permission denied**: Check config file permissions and ownership
- **High API calls**: Use `--v=2` and consider `--scan-interval=30s`

```bash
# Debug logging with hostPath config
./cluster-autoscaler --cloud-provider=vcloud --cloud-config=/etc/config/cloud-config --v=4 --logtostderr

# Check config file
ls -la /etc/config/cloud-config
cat /etc/config/cloud-config

# Test API connectivity
curl -H "X-Provider-Token: $TOKEN" "$MGMT_URL/nodepools"

# Test individual node operations
curl -X DELETE -H "X-Provider-Token: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"force": false, "reason": "cluster-autoscaler-scale-down"}' \
  "$MGMT_URL/nodepools/{pool-id}/machines/{machine-id}"
```