---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    k8s-app: cluster-autoscaler
  annotations:
    azure.workload.identity/client-id: "${CAS_UAI_CLIENTID}"
  name: cluster-autoscaler
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
    verbs: ["watch", "list", "get", "update"]
  - apiGroups: [""]
    resources:
      - "namespaces"
      - "pods"
      - "services"
      - "replicationcontrollers"
      - "persistentvolumeclaims"
      - "persistentvolumes"
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
    verbs: ["get", "list", "watch"]
  - apiGroups: ["batch"]
    resources: ["jobs", "cronjobs"]
    verbs: ["watch", "list", "get"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["create"]
  - apiGroups: ["coordination.k8s.io"]
    resourceNames: ["cluster-autoscaler"]
    resources: ["leases"]
    verbs: ["get", "update", "patch", "delete"]
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
    name: cluster-autoscaler
    namespace: kube-system
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
    resourceNames:
      - "cluster-autoscaler-status"
      - "cluster-autoscaler-priority-expander"
    verbs: ["delete", "get", "update", "watch"]
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
---
apiVersion: v1
data:
  ResourceGroup: "${RESOURCE_GROUP_MC_B64}"
  TenantID: "${TENANT_ID_B64}"
  SubscriptionID: "${SUBSCRIPTION_ID_B64}"
  VMType: dm1zcw==
kind: Secret
metadata:
  name: cluster-autoscaler-azure
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: cluster-autoscaler
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
        azure.workload.identity/use: "true"
    spec:
      serviceAccountName: cluster-autoscaler
      containers:
        - image: cluster-autoscaler
          imagePullPolicy: Always
          name: cluster-autoscaler
          args:
            - ./cluster-autoscaler
            - --v=3
            - --logtostderr=true
            - --cloud-provider=azure
            - --skip-nodes-with-local-storage=false
            - --nodes=1:10:${VMSS_NAME}
          env:
            - name: ARM_SUBSCRIPTION_ID
              valueFrom:
                secretKeyRef:
                  key: SubscriptionID
                  name: cluster-autoscaler-azure
            - name: ARM_RESOURCE_GROUP
              valueFrom:
                secretKeyRef:
                  key: ResourceGroup
                  name: cluster-autoscaler-azure
            - name: ARM_TENANT_ID
              valueFrom:
                secretKeyRef:
                  key: TenantID
                  name: cluster-autoscaler-azure
            - name: ARM_USE_WORKLOAD_IDENTITY_EXTENSION
              value: "true"
            - name: ARM_VM_TYPE
              valueFrom:
                secretKeyRef:
                  key: VMType
                  name: cluster-autoscaler-azure
          resources:
            limits:
              cpu: 100m
              memory: 300Mi
            requests:
              cpu: 100m
              memory: 300Mi
          volumeMounts:
            - mountPath: /etc/ssl/certs/ca-certificates.crt
              name: ssl-certs
              readOnly: true
      # use system nodepool only
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                - key: kubernetes.azure.com/mode
                  operator: In
                  values: ["system"]
      restartPolicy: Always
      volumes: 
        - hostPath:
            path: /etc/ssl/certs/ca-certificates.crt
            type: ""
          name: ssl-certs