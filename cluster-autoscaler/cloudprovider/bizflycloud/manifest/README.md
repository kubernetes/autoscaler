# Phương án triển khai cluster-autoscaler(CA) cho BKE

Mô hình triển khai

![model](https://raw.githubusercontent.com/lmq1999/123/master/CA.png)

Pod CA sẽ được đặt ở namespace của cluster nó quản lý

Ở shoot-cluster cần apply RBAC trong manifest

## Manifest

**Deployment:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-autoscaler
  namespace: shoot-ke1m4ea3j8w6kkvp
  labels:
    app: cluster-autoscaler
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
        - image: cr-hn-1.bizflycloud.vn/1e7f10a9850b45b488a3f0417ccb60e0/cluster-autoscaler:test
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
            - --cloud-provider=bizflycloud
            - --skip-nodes-with-local-storage=false
            - --leader-elect=true
            - --expander=least-waste
            - --kubeconfig=/etc/kubernetes/shoot-ke1m4ea3j8w6kkvp.kubeconfig
          env:
            - name: BIZFLYCLOUD_AUTH_METHOD
              value: password #application_credential
            - name: BIZFLYCLOUD_EMAIL
              value: xxxxxxxxxxxxxxxxx
            - name: BIZFLYCLOUD_PASSWORD
              value: xxxxxxxxxxxxxxxxx
            # - name: BIZFLYCLOUD_APP_CREDENTIAL_ID
            #   value: xxxxxxxxxxxxxxxxx
            # - name: BIZFLYCLOUD_APP_CREDENTIAL_SECRET
            #   value: xxxxxxxxxxxxxxxxx
            # - name: BIZFLYCLOUD_PROJECT_ID
            #   value: xxxxxxxxxxxxxxxxx
            # - name: BIZFLYCLOUD_TENANT_ID
            #   value: xxxxxxxxxxxxxxxxx
            - name: BIZFLYCLOUD_REGION
              value: HN
            - name: BIZFLYCLOUD_API_URL
              value: https://staging.bizflycloud.vn
            - name: CLUSTER_NAME
              value: ke1m4ea3j8w6kkvp
          volumeMounts:
          - mountPath: /etc/kubernetes/
            name: k8s-config
            readOnly: true
      volumes:
      - configMap:
          defaultMode: 420
          name: k8s-config
        name: k8s-config
```

**RBAC:**

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    k8s-addon: cluster-autoscaler.addons.k8s.io
    k8s-app: cluster-autoscaler
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
    resources: ["storageclasses", "csinodes"]
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
    verbs: ["delete", "get", "update", "watch"]

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

## Các bổ sung cần thiết

### Autoscaling-group

Autoscaling-group không tự tạo policy để CA có thể hoạt động, vẫn yêu cầu bắt buộc phải enable

Nếu không enable sẽ lấy từ autoscaling-group với desire=min=max từ đó không scale được.

### Labels

Trong các node worker cần bổ sung thêm 2 labels như sau

```yaml
bke.bizflycloud.vn/node-id: xxxxxxxxxxxx #Để CA xác định được chính xác node cần xóa
bke.bizflycloud.vn/pool-name: xxxxxxxxxxx #Để thuận tiện trong việc sử dụng node-selector nhằm scale vào đúng pool đó
```

### kubeconfig

Trong pod CA cần mount configmaps chưa kubeconfig của cluster để trỏ đến và ghi **cluster-autoscaler-status** vào configmap trong shoot-cluster (khách hàng theo dõi được)

Demo:

```yaml
quanlm@quanlm-desktop:~$ kubectl describe configmaps cluster-autoscaler-status -n kube-system 
Name:         cluster-autoscaler-status
Namespace:    kube-system
Labels:       <none>
Annotations:  cluster-autoscaler.kubernetes.io/last-updated: 2021-04-05 08:39:21.577044595 +0000 UTC

Data
====
status:
----
Cluster-autoscaler status at 2021-04-05 08:39:21.577044595 +0000 UTC:
Cluster-wide:
  Health:      Healthy (ready=3 unready=0 notStarted=0 longNotStarted=0 registered=3 longUnregistered=0)
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:32:31.664863279 +0000 UTC m=+42.861518513
  ScaleUp:     InProgress (ready=3 registered=3)
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:32:31.664863279 +0000 UTC m=+42.861518513
  ScaleDown:   NoCandidates (candidates=0)
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:38:33.283990811 +0000 UTC m=+404.480646061

NodeGroups:
  Name:        606ac80ef9253182e9bc8f64
  Health:      Healthy (ready=1 unready=0 notStarted=0 longNotStarted=0 registered=1 longUnregistered=0 cloudProviderTarget=3 (minSize=1, maxSize=3))
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:32:31.664863279 +0000 UTC m=+42.861518513
  ScaleUp:     InProgress (ready=1 cloudProviderTarget=3)
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:32:31.664863279 +0000 UTC m=+42.861518513
  ScaleDown:   NoCandidates (candidates=0)
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:36:52.537836401 +0000 UTC m=+303.734491633

  Name:        606ac820f93c27e8eb77df8d
  Health:      Healthy (ready=1 unready=0 notStarted=0 longNotStarted=0 registered=1 longUnregistered=0 cloudProviderTarget=3 (minSize=1, maxSize=3))
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:32:31.664863279 +0000 UTC m=+42.861518513
  ScaleUp:     InProgress (ready=1 cloudProviderTarget=3)
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:32:55.216254754 +0000 UTC m=+66.412910020
  ScaleDown:   NoCandidates (candidates=0)
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:38:33.283990811 +0000 UTC m=+404.480646061

  Name:        606ac847d7048cfe5d14ad11
  Health:      Healthy (ready=1 unready=0 notStarted=0 longNotStarted=0 registered=1 longUnregistered=0 cloudProviderTarget=2 (minSize=1, maxSize=3))
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:32:31.664863279 +0000 UTC m=+42.861518513
  ScaleUp:     InProgress (ready=1 cloudProviderTarget=2)
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:33:17.962127761 +0000 UTC m=+89.158782997
  ScaleDown:   NoCandidates (candidates=0)
               LastProbeTime:      2021-04-05 08:39:11.702759465 +0000 UTC m=+442.899414715
               LastTransitionTime: 2021-04-05 08:37:53.299217703 +0000 UTC m=+364.495872923


Events:
  Type    Reason         Age    From                Message
  ----    ------         ----   ----                -------
  Normal  ScaledUpGroup  6m49s  cluster-autoscaler  Scale-up: setting group 606ac80ef9253182e9bc8f64 size to 3
```
