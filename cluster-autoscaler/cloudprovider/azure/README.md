# Cluster Autoscaler on Azure

The cluster autoscaler on Azure scales worker nodes within any specified autoscaling group. It will run as a `Deployment` in your cluster. This README will go over some of the necessary steps required to get the cluster autoscaler up and running.

## Kubernetes Version

Cluster autoscaler must run on Kubernetes with Azure VMSS support ([kubernetes#43287](https://github.com/kubernetes/kubernetes/issues/43287)). It is planed in Kubernetes v1.10.

## Permissions

Get azure credentials by running the following command

```sh
# replace <subscription-id> with yours.
az ad sp create-for-rbac --role="Contributor" --scopes="/subscriptions/<subscription-id>" --output json
```

And fill the values with the result you got into the configmap

```yaml
apiVersion: v1
data:
  ClientID: <client-id>
  ClientSecret: <client-secret>
  ResourceGroup: <resource-group>
  SubscriptionID: <subscription-id>
  TenantID: <tenand-id>
  ScaleSetName: <scale-set-name>
kind: ConfigMap
metadata:
  name: cluster-autoscaler-azure
  namespace: kube-system
```

Create the configmap by running

```sh
kubectl create -f cluster-autoscaler-azure-configmap.yaml
```

## Deployment

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
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
      labels:
        app: cluster-autoscaler
    spec:
      containers:
      - image: gcr.io/google_containers/cluster-autoscaler:{{ ca_version }}
        name: cluster-autoscaler
        resources:
          limits:
            cpu: 100m
            memory: 300Mi
          requests:
            cpu: 100m
            memory: 300Mi
        env:
        - name: ARM_SUBSCRIPTION_ID
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: SubscriptionID
        - name: ARM_RESOURCE_GROUP
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: ResourceGroup
        - name: ARM_TENANT_ID
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: TenantID
        - name: ARM_CLIENT_ID
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: ClientID
        - name: ARM_CLIENT_SECRET
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: ClientSecret
        - name: ARM_SCALE_SET_NAME
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: ScaleSetName
        command:
          - ./cluster-autoscaler
          - --v=4
          - --cloud-provider=azure
          - --skip-nodes-with-local-storage=false
          - --nodes="1:10:$(ARM_SCALE_SET_NAME)"
        volumeMounts:
          - name: ssl-certs
            mountPath: /etc/ssl/certs/ca-certificates.crt
            readOnly: true
        imagePullPolicy: "Always"
      volumes:
      - name: ssl-certs
        hostPath:
          path: "/etc/ssl/certs/ca-certificates.crt"
```

## Deploy in master node

To run a CA pod in master node - CA deployment should tolerate the master `taint` and `nodeSelector` should be used to schedule the pods in master node.

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
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
      labels:
        app: cluster-autoscaler
    spec:
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      nodeSelector:
        kubernetes.io/role: master
      containers:
      - image: gcr.io/google_containers/cluster-autoscaler:{{ ca_version }}
        name: cluster-autoscaler
        resources:
          limits:
            cpu: 100m
            memory: 300Mi
          requests:
            cpu: 100m
            memory: 300Mi
        env:
        - name: ARM_SUBSCRIPTION_ID
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: SubscriptionID
        - name: ARM_RESOURCE_GROUP
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: ResourceGroup
        - name: ARM_TENANT_ID
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: TenantID
        - name: ARM_CLIENT_ID
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: ClientID
        - name: ARM_CLIENT_SECRET
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: ClientSecret
        - name: ARM_SCALE_SET_NAME
          valueFrom:
            configMapKeyRef:
              name: cluster-autoscaler-azure
              key: ScaleSetName
        command:
          - ./cluster-autoscaler
          - --v=4
          - --cloud-provider=azure
          - --skip-nodes-with-local-storage=false
          - --nodes="1:10:$(ARM_SCALE_SET_NAME)"
        volumeMounts:
          - name: ssl-certs
            mountPath: /etc/ssl/certs/ca-certificates.crt
            readOnly: true
        imagePullPolicy: "Always"
      volumes:
      - name: ssl-certs
        hostPath:
          path: "/etc/ssl/certs/ca-certificates.crt"
```
