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

# Build for specific platforms
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cluster-autoscaler-linux-amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o cluster-autoscaler-linux-arm64 .
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o cluster-autoscaler-darwin-amd64 .
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o cluster-autoscaler-darwin-arm64 .
```

Deploy with the VCloud provider:

```bash
# Option 1: Using Static mode
./cluster-autoscaler-darwin-arm64 \
  --cloud-provider=vcloud \
  --cloud-config=/etc/config/cloud-config \
  --kubeconfig=$HOME/.kube/config \
  --nodes=1:3:nodepool-name \
  --v=2 --logtostderr

# Option 2: Auto-discovery mode
./cluster-autoscaler-darwin-arm64 \
  --cloud-provider=vcloud \
  --node-group-auto-discovery=vcloud:tag=autoscaler.k8s.io.infra.vnetwork.dev/enabled=true \
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
  labels:
    app: cluster-autoscaler
spec:
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
      - image: registry.k8s.io/autoscaling/cluster-autoscaler:latest
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
        - --v=4
        - --stderrthreshold=info
        - --cloud-provider=vcloud
        - --skip-nodes-with-local-storage=false
        - --expander=least-waste
        - --node-group-auto-discovery=vcloud:autoscaler.k8s.io.infra.vnetwork.dev/enabled=true
        - --balance-similar-node-groups
        - --skip-nodes-with-system-pods=false
        - --scale-down-enabled=true
        - --scale-down-delay-after-add=10m
        - --scale-down-unneeded-time=10m
        - --scale-down-utilization-threshold=0.5
        - --max-node-provision-time=15m
        - --enforce-node-group-min-size=true
        volumeMounts:
        - name: cloud-config
          mountPath: /etc/config
          readOnly: true
        env:
        - name: CLUSTER_ID
          valueFrom:
            secretKeyRef:
              name: vcloud-config
              key: cluster-id
        - name: CLUSTER_NAME
          valueFrom:
            secretKeyRef:
              name: vcloud-config
              key: cluster-name
        - name: MGMT_URL
          valueFrom:
            secretKeyRef:
              name: vcloud-config
              key: mgmt-url
        - name: PROVIDER_TOKEN
          valueFrom:
            secretKeyRef:
              name: vcloud-config
              key: provider-token
      volumes:
        - name: cloud-config
          hostPath:
            path: /etc/config
            type: Directory
```

### Service Account and RBAC

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cluster-autoscaler
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-autoscaler
rules:
- apiGroups: [""]
  resources: ["events", "endpoints"]
  verbs: ["create", "patch"]
- apiGroups: [""]
  resources: ["pods/eviction"]
  verbs: ["create"]
- apiGroups: [""]
  resources: ["pods/status"]
  verbs: ["update"]
- apiGroups: [""]
  resources: ["endpoints"]
  resourceNames: ["cluster-autoscaler"]
  verbs: ["get", "update"]
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["watch", "list", "get", "update"]
- apiGroups: [""]
  resources: ["namespaces", "pods", "services", "replicationcontrollers", "persistentvolumeclaims", "persistentvolumes"]
  verbs: ["watch", "list", "get"]
- apiGroups: ["extensions"]
  resources: ["replicasets", "daemonsets"]
  verbs: ["watch", "list", "get"]
- apiGroups: ["policy"]
  resources: ["poddisruptionbudgets"]
  verbs: ["watch", "list"]
- apiGroups: ["apps"]
  resources: ["statefulsets", "replicasets", "daemonsets"]
  verbs: ["watch", "list", "get"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses", "csinodes", "csidrivers", "csistoragecapacities"]
  verbs: ["watch", "list", "get"]
- apiGroups: ["batch", "extensions"]
  resources: ["jobs"]
  verbs: ["get", "list", "watch", "patch"]
- apiGroups: ["coordination.k8s.io"]
  resources: ["leases"]
  verbs: ["create"]
- apiGroups: ["coordination.k8s.io"]
  resourceNames: ["cluster-autoscaler"]
  resources: ["leases"]
  verbs: ["get", "update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-autoscaler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-autoscaler
subjects:
- kind: ServiceAccount
  name: cluster-autoscaler
  namespace: kube-system
```

### Configuration Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: vcloud-config
  namespace: kube-system
type: Opaque
stringData:
  cluster-id: "your-cluster-id"
  cluster-name: "your-cluster-name"
  mgmt-url: "https://k8s.io.infra.vnetwork.dev/api/v2/..."
  provider-token: "your-api-token"
```

## Important Configuration Flags

### Essential Flags for VCloud

- `--enforce-node-group-min-size=true`: Ensures minimum node count is maintained
- `--cloud-provider=vcloud`: Specifies VCloud as the cloud provider
- `--v=4`: Sets appropriate logging level for troubleshooting

### Scaling Behavior

- `--scale-down-delay-after-add=10m`: Wait time before considering scale-down after adding nodes
- `--scale-down-unneeded-time=10m`: Time a node should be unneeded before it's eligible for removal
- `--scale-down-utilization-threshold=0.5`: Node utilization threshold for scale-down decisions
- `--max-node-provision-time=15m`: Maximum time to wait for node provisioning

### Node Selection

- `--skip-nodes-with-local-storage=false`: Allow scaling down nodes with local storage
- `--skip-nodes-with-system-pods=false`: Allow scaling down nodes with system pods
- `--balance-similar-node-groups`: Distribute pods across similar node groups

## API Endpoints

The VCloud provider uses the following API endpoints:

- `GET /clusters/{cluster-id}/nodepools` - List all node pools
- `GET /clusters/{cluster-id}/nodepools/{pool-id}` - Get specific node pool info
- `GET /clusters/{cluster-id}/nodepools/{pool-id}/machines` - List instances in pool
- `PUT /clusters/{cluster-id}/nodepools/{pool-id}/scale` - Scale node pool
- `DELETE /clusters/{cluster-id}/nodepools/{pool-id}/machines/{instance-id}` - Delete instance

## Troubleshooting

### Common Issues

#### 1. Excessive API Calls

**Symptoms:**
```
I0624 10:15:23.456789 1 vcloud_manager.go:352] Listing instances for node pool: pool-1
I0624 10:15:23.567890 1 vcloud_manager.go:352] Listing instances for node pool: pool-2
I0624 10:15:23.678901 1 vcloud_manager.go:352] Listing instances for node pool: pool-3
I0624 10:15:23.789012 1 vcloud_manager.go:352] Listing instances for node pool: pool-4
```

**Root Cause:** The `NodeGroupForNode` function was calling `group.Nodes()` for every node group to find which one contains a specific node, causing multiplicative API calls.

**Solution:** Implemented two-level caching:
- Node group level instance caching (30-second TTL)
- Cloud provider level node-to-nodegroup mapping cache

**Performance Impact:** Reduced API calls from 20+ per evaluation cycle to 1-2 per 30 seconds (90%+ reduction).

#### 2. Node Group Not Initialized Errors

**Symptoms:**
```
E0624 10:15:24.123456 1 vcloud_node_group.go:242] node group not initialized
```

**Root Cause:** NodeGroup methods were trying to access a removed `nodePool` field instead of using the `targetSize` field directly.

**Solution:** Updated all NodeGroup methods to work with the simplified struct:
```go
type NodeGroup struct {
    id         string
    clusterID  string
    client     *VCloudAPIClient
    manager    *EnhancedManager
    minSize    int
    maxSize    int
    targetSize int
}
```

#### 3. JSON Parsing Errors

**Symptoms:**
```
E0624 10:15:24.234567 1 vcloud_manager.go:385] Failed to parse instances response as JSON
```

**Root Cause:** API response format changed from flat structure to nested format.

**Before:**
```json
{
  "machines": [{"id": "1", "name": "node1"}]
}
```

**After:**
```json
{
  "status": 200,
  "data": {
    "machines": [{"id": "1", "name": "node1"}]
  }
}
```

**Solution:** Updated JSON parsing structure in `ListNodePoolInstances`:
```go
var instancesResponse struct {
    Status int `json:"status"`
    Data   struct {
        Machines []struct {
            ID        string `json:"id"`
            Name      string `json:"name"`
            State     string `json:"state"`
            // ... other fields
        } `json:"machines"`
    } `json:"data"`
}
```

### Debugging Commands

Check cluster autoscaler logs:
```bash
kubectl logs -n kube-system deployment/cluster-autoscaler -f
```

View node group status:
```bash
kubectl get nodes --show-labels
```

Check autoscaler events:
```bash
kubectl get events --field-selector involvedObject.name=cluster-autoscaler -n kube-system
```

### Performance Monitoring

Monitor API call patterns:
```bash
kubectl logs -n kube-system deployment/cluster-autoscaler | grep "Making.*request to VCloud API"
```

Check cache efficiency:
```bash
kubectl logs -n kube-system deployment/cluster-autoscaler | grep "found cached\|rebuilding.*cache"
```

## Node Pool Requirements

For a node pool to be managed by cluster autoscaler:

1. **Autoscaling Enabled**: `minSize > 0` or `maxSize > 0`
2. **Proper Labels**: Nodes should have the machine ID label: `k8s.io.infra.vnetwork.io/machine-id`
3. **Provider ID**: Nodes must have provider ID in format: `vcloud://instance-id`

## Scaling Behavior

### Scale Up

Triggered when:
- Pods cannot be scheduled due to insufficient resources
- New pods remain in `Pending` state for more than 30 seconds

Process:
1. Evaluate resource requirements of pending pods
2. Select appropriate node group based on resource fit
3. Call VCloud API to increase node pool size
4. Wait for new nodes to join the cluster

### Scale Down

Triggered when:
- Node utilization falls below threshold (default 50%)
- Node has been underutilized for the configured time period
- All pods on the node can be rescheduled elsewhere

Process:
1. Identify underutilized nodes
2. Simulate pod rescheduling to ensure feasibility
3. Gracefully drain pods from the node
4. Call VCloud API to delete the instance

## Best Practices

### Resource Requests

Always set resource requests on pods:
```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: app
    image: myapp:latest
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 200m
        memory: 256Mi
```

### Node Pool Configuration

Configure appropriate min/max sizes:
```yaml
# Example node pool configuration
minSize: 2      # Ensure minimum availability
maxSize: 10     # Prevent runaway scaling
desiredSize: 3  # Initial size
```

### Pod Disruption Budgets

Use PodDisruptionBudgets to control scaling behavior:
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: myapp-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: myapp
```

### Build Instructions

Build the cluster autoscaler with VCloud support:

```bash
cd cluster-autoscaler

# Build VCloud-specific binary
go build -tags vcloud -o cluster-autoscaler-vcloud .

# Build for specific platforms
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cluster-autoscaler-linux-amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o cluster-autoscaler-linux-arm64 .
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o cluster-autoscaler-darwin-amd64 .
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o cluster-autoscaler-darwin-arm64 .
```

## Monitoring and Alerts

### Key Metrics

Monitor these metrics for cluster autoscaler health:

- **Scaling Events**: Frequency of scale up/down operations
- **API Latency**: Response times for VCloud API calls
- **Cache Hit Rate**: Effectiveness of node group caching
- **Pending Pods**: Pods waiting for resources

### Recommended Alerts

1. **High API Error Rate**: > 5% of VCloud API calls failing
2. **Slow Scaling**: Node provisioning taking > 15 minutes
3. **Cache Misses**: Cache rebuild frequency > 1 per minute
4. **Stuck Pods**: Pods pending for > 10 minutes

## Version Compatibility

- **Kubernetes**: 1.25+
- **Cluster Autoscaler**: 1.28+
- **VCloud API**: v2.0+

## Contributing

When modifying the VCloud provider:

1. Follow existing code patterns and error handling
2. Add appropriate logging with `klog`
3. Include unit tests for new functionality
4. Update this README for any configuration changes
5. Test with various node pool configurations

## Support

For issues and support:

1. Check this troubleshooting guide first
2. Review cluster autoscaler logs with `--v=4` flag
3. Verify VCloud API connectivity and authentication
4. Ensure node pools have proper autoscaling configuration

## Features

- Auto-discovery of VCloud NodePools with autoscaling enabled
- Standard Cluster Autoscaler interfaces (CloudProvider, NodeGroup)
- Individual node deletion (follows common cloud provider patterns)
- Provider ID format: `vcloud://instance-uuid`
- VCloud-specific labels: `k8s.io.infra.vnetwork.io/*`
- Retry logic with exponential backoff
- Support for both config files and environment variables
- Two-level caching for performance optimization

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

## Architecture Details

The VCloud provider implements a **hybrid scaling approach** that combines the best practices from other cloud
providers:

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
- **Performance Optimization**: Two-level caching reduces API calls by 90%+

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

- Configuration parsing
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

# Test scaling
kubectl run test-scale --image=nginx --requests=cpu=1000m --replicas=3
kubectl get nodes -w
kubectl delete deployment test-scale
```

## Configuration Setup

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

## Troubleshooting

Common issues and solutions:

- **Node groups not discovered**: Verify NodePools have `min > 0` and `max > min`
- **Scale up fails**: Check provider token permissions and NodePool capacity limits
- **Scale down fails**: Verify nodes belong to the node group and minimum size constraints
- **Individual node deletion fails**: Check instance exists and is in deletable state
- **Configuration errors**:
    - Config file: Check `/etc/config/cloud-config` exists and has correct permissions (600)
- **Config file not found**: Ensure `/etc/config/cloud-config` exists on the node running cluster-autoscaler
- **Permission denied**: Check config file permissions and ownership
- **High API calls**: Use `--v=2` and consider `--scan-interval=30s`

```bash
# Debug logging v4
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