# Cluster Autoscaler on Volcengine

The Cluster Autoscaler on Volcengine dynamically scales Kubernetes worker nodes. It runs as a deployment within your cluster. This README provides a step-by-step guide for setting up cluster autoscaler on your Kubernetes cluster.

## Permissions

### Using Volcengine Credentials

To use Volcengine credentials, create a `Secret` with your access key and access key secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloud-config
  namespace: kube-system
type: Opaque
data:
  access-key: [YOUR_BASE64_AK_ID]
  secret-key: [YOUR_BASE64_AK_SECRET]
  region-id: [YOUR_BASE64_REGION_ID]
```

See the [Volcengine Access Key User Manual](https://www.volcengine.com/docs/6291/65568) and [Volcengine Autoscaling Region](https://www.volcengine.com/docs/6617/87001) for more information.

## Manual Configuration

### Auto Scaling Group Setup

1. Create an Auto Scaling Group in the [Volcengine Console](https://console.volcengine.com/as) with valid configurations, and set the desired instance number to zero.

2. Create a Scaling Configuration for the Scaling Group with valid configurations. In User Data, specify the script to initialize the environment and join this node to the Kubernetes cluster.

### Cluster Autoscaler Deployment

1. Create a service account.

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    k8s-app: cluster-autoscaler
  name: cluster-autoscaler-account
  namespace: kube-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-autoscaler
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    k8s-app: cluster-autoscaler
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
    verbs: ["watch", "list", "get", "update", "delete"]
  - apiGroups: [""]
    resources:
      - "namespaces"
      - "pods"
      - "services"
      - "replicationcontrollers"
      - "persistentvolumeclaims"
      - "persistentvolumes"
    verbs: ["watch", "list", "get"]
  - apiGroups: ["batch", "extensions"]
    resources: ["jobs"]
    verbs: ["watch", "list", "get", "patch"]
  - apiGroups: [ "policy" ]
    resources: [ "poddisruptionbudgets" ]
    verbs: [ "watch", "list" ]
  - apiGroups: ["apps"]
    resources: ["daemonsets", "replicasets", "statefulsets"]
    verbs: ["watch", "list", "get"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses", "csinodes", "csidrivers", "csistoragecapacities"]
    verbs: ["watch", "list", "get"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["create","list","watch"]
  - apiGroups: [""]
    resources: ["configmaps"]
    resourceNames: ["cluster-autoscaler-status", "cluster-autoscaler-priority-expander"]
    verbs: ["delete", "get", "update"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["watch", "list", "get", "create", "update", "patch", "delete", "deletecollection"]
  - apiGroups: ["extensions"]
    resources: ["replicasets", "daemonsets"]
    verbs: ["watch", "list", "get"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: cluster-autoscaler
  namespace: kube-system
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    k8s-app: cluster-autoscaler
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["create","list","watch"]
  - apiGroups: [""]
    resources: ["configmaps"]
    resourceNames: ["cluster-autoscaler-status", "cluster-autoscaler-priority-expander"]
    verbs: ["delete","get","update","watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-autoscaler
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    k8s-app: cluster-autoscaler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-autoscaler
subjects:
  - kind: ServiceAccount
    name: cluster-autoscaler-account
    namespace: kube-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: cluster-autoscaler
  namespace: kube-system
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    k8s-app: cluster-autoscaler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: cluster-autoscaler
subjects:
  - kind: ServiceAccount
    name: cluster-autoscaler
    namespace: kube-system
```

2. Create a deployment.

```yaml
---
kind: Deployment
apiVersion: apps/v1
metadata:
  name: cluster-autoscaler
  namespace: kube-system
  labels:
    app: cluster-autoscaler
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cluster-autoscaler
  template:
    metadata:
      namespace: kube-system
      labels:
        app: cluster-autoscaler
    spec:
      serviceAccountName: cluster-autoscaler-account
      containers:
        - name: cluster-autoscaler
          image: registry.k8s.io/autoscaling/cluster-autoscaler:latest
          imagePullPolicy: Always
          command:
            - ./cluster-autoscaler
            - --alsologtostderr
            - --cloud-config=/config/cloud-config
            - --cloud-provider=volcengine
            - --nodes=[min]:[max]:[ASG_ID]
            - --scale-down-delay-after-add=1m0s
            - --scale-down-unneeded-time=1m0s
          env:
            - name: ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: cloud-config
                  key: access-key
            - name: SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: cloud-config
                  key: secret-key
            - name: REGION_ID
              valueFrom:
                secretKeyRef:
                  name: cloud-config
                  key: region-id
```

## Auto-Discovery Setup

Auto Discovery is not currently supported in Volcengine.