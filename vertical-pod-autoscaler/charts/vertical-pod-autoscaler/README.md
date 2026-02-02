# vertical-pod-autoscaler

WARNING: This chart is currently under development and is not ready for production use.

Automatically adjust resources for your workloads

![Version: 0.8.0](https://img.shields.io/badge/Version-0.8.0-informational?style=flat-square)
![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)
![AppVersion: 1.6.0](https://img.shields.io/badge/AppVersion-1.6.0-informational?style=flat-square)

## Introduction
The Vertical Pod Autoscaler (VPA) automatically adjusts the CPU and memory resource requests of pods to match their actual resource utilization.

## Helm Installation & upgrade

```bash
helm repo add autoscalers https://kubernetes.github.io/autoscaler
helm upgrade -i vertical-pod-autoscaler autoscalers/vertical-pod-autoscaler
```

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| adrianmoisey | <kubernetes-sig-autoscaling@googlegroups.com> |  |
| omerap12 | <kubernetes-sig-autoscaling@googlegroups.com> |  |

## Webhook Management
The admission controller requires a `MutatingWebhookConfiguration` and TLS certificates. This chart supports two mutually exclusive modes:

### Helm-managed (default)
```yaml
admissionController:
  registerWebhook: false
  certGen:
    enabled: true
```
In this mode:
- Helm creates the MutatingWebhookConfiguration
- The kube-webhook-certgen job generates TLS certificates and stores them in a Secret
- The certificates are automatically injected into the webhook configuration

### Application-managed
```yaml
admissionController:
  registerWebhook: true
  certGen:
    enabled: false
```
In this mode:
- The VPA admission controller creates and manages the webhook itself
Important: You are responsible for creating the TLS secret before or after installing the chart. The admission controller will only create the `MutatingWebhookConfiguration` once the secret exists.
If the secret is created after the Helm install, you must restart the admission controller pod to trigger webhook registration.

## Migration Guides

### Migrating from vpa-up.sh script
TBD

### Migrating from Application-managed to Helm-managed webhook
If you previously deployed with registerWebhook: true and want to switch to Helm-managed:
- Delete the existing webhook:
```bash
kubectl delete mutatingwebhookconfiguration vpa-webhook-config
```
- Delete the existing secret (to allow certgen to create new certificates):
```bash
kubectl delete secret -n <namespace> vpa-tls-certs
```
- Upgrade with the new values:
```bash
helm upgrade <release-name> <chart> \
  --set admissionController.registerWebhook=false \
  --set admissionController.certGen.enabled=true
```
## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| admissionController.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].key | string | `"app.kubernetes.io/component"` |  |
| admissionController.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].operator | string | `"In"` |  |
| admissionController.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].values[0] | string | `"admission-controller"` |  |
| admissionController.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.topologyKey | string | `"kubernetes.io/hostname"` |  |
| admissionController.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight | int | `100` |  |
| admissionController.certGen.affinity | object | `{}` |  |
| admissionController.certGen.enabled | bool | `true` |  |
| admissionController.certGen.env | object | `{}` | Additional environment variables to be added to the certgen container. Format is KEY: Value format |
| admissionController.certGen.image.pullPolicy | string | `"IfNotPresent"` | The pull policy for the certgen image. Recommend not changing this |
| admissionController.certGen.image.repository | string | `"registry.k8s.io/ingress-nginx/kube-webhook-certgen"` | An image that contains certgen for creating certificates. |
| admissionController.certGen.image.tag | string | `"v20231011-8b53cabe0"` | An image tag for the admissionController.certGen.image.repository image. |
| admissionController.certGen.nodeSelector | object | `{}` |  |
| admissionController.certGen.podSecurityContext | object | `{"runAsNonRoot":true,"runAsUser":65534,"seccompProfile":{"type":"RuntimeDefault"}}` | The securityContext block for the certgen pod(s) |
| admissionController.certGen.resources | object | `{}` | The resources block for the certgen pod |
| admissionController.certGen.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true}` | The securityContext block for the certgen container(s) |
| admissionController.certGen.tolerations | list | `[]` |  |
| admissionController.enabled | bool | `true` |  |
| admissionController.extraArgs | list | `[]` |  |
| admissionController.extraEnv | list | `[]` |  |
| admissionController.hostNetwork | bool | `false` | Enable host network for the admission controller pod. Set to true when the pod needs direct access to the host's network namespace. Note: this bypasses Kubernetes network isolation and may cause port conflicts if multiple replicas run on the same node. |
| admissionController.image.pullPolicy | string | `"IfNotPresent"` |  |
| admissionController.image.repository | string | `"registry.k8s.io/autoscaling/vpa-admission-controller"` |  |
| admissionController.image.tag | string | `nil` |  |
| admissionController.mutatingWebhookConfiguration.annotations | object | `{}` | Additional annotations for the MutatingWebhookConfiguration |
| admissionController.mutatingWebhookConfiguration.failurePolicy | string | `"Ignore"` | The failurePolicy for the mutating webhook. Allowed values are: Ignore, Fail |
| admissionController.mutatingWebhookConfiguration.namespaceSelector | object | `{}` | The namespaceSelector controls which namespaces are affected by the webhook |
| admissionController.mutatingWebhookConfiguration.objectSelector | object | `{}` | The objectSelector can filter objects on e.g. labels |
| admissionController.mutatingWebhookConfiguration.timeoutSeconds | int | `5` | Sets the amount of time the API server will wait on a response from the webhook service |
| admissionController.nodeSelector | object | `{}` |  |
| admissionController.podAnnotations | object | `{}` |  |
| admissionController.podDisruptionBudget.enabled | bool | `true` |  |
| admissionController.podDisruptionBudget.maxUnavailable | int or string | `nil` | Maximum number/percentage of pods that can be unavailable after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| admissionController.podDisruptionBudget.minAvailable | int or string | `1` | Minimum number/percentage of pods that must be available after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| admissionController.podLabels | object | `{}` |  |
| admissionController.priorityClassName | string | `nil` |  |
| admissionController.registerWebhook | bool | `false` | Whether to register webhook via the application itself or via Helm. Set to false when using Helm-managed webhook. Security issue: granting delete on mutatingwebhookconfigurations is a potential security risk as it allows the admission controller to remove any webhook configurations. |
| admissionController.replicas | int | `2` |  |
| admissionController.resources | object | `{}` |  |
| admissionController.service.annotations | object | `{}` |  |
| admissionController.service.name | string | `"vpa-webhook"` |  |
| admissionController.service.ports[0].port | int | `443` |  |
| admissionController.service.ports[0].protocol | string | `"TCP"` |  |
| admissionController.service.ports[0].targetPort | int | `8000` |  |
| admissionController.serviceAccount.annotations | object | `{}` |  |
| admissionController.serviceAccount.create | bool | `true` |  |
| admissionController.serviceAccount.labels | object | `{}` |  |
| admissionController.tls.caCert | string | `""` |  |
| admissionController.tls.cert | string | `""` |  |
| admissionController.tls.create | bool | `false` |  |
| admissionController.tls.key | string | `""` |  |
| admissionController.tls.secretName | string | `"vpa-tls-certs"` |  |
| admissionController.tolerations | list | `[]` |  |
| admissionController.volumeMounts[0].mountPath | string | `"/etc/tls-certs"` |  |
| admissionController.volumeMounts[0].name | string | `"tls-certs"` |  |
| admissionController.volumeMounts[0].readOnly | bool | `true` |  |
| admissionController.volumes[0].name | string | `"tls-certs"` |  |
| admissionController.volumes[0].secret.defaultMode | int | `420` |  |
| admissionController.volumes[0].secret.items[0].key | string | `"ca"` |  |
| admissionController.volumes[0].secret.items[0].path | string | `"caCert.pem"` |  |
| admissionController.volumes[0].secret.items[1].key | string | `"cert"` |  |
| admissionController.volumes[0].secret.items[1].path | string | `"serverCert.pem"` |  |
| admissionController.volumes[0].secret.items[2].key | string | `"key"` |  |
| admissionController.volumes[0].secret.items[2].path | string | `"serverKey.pem"` |  |
| admissionController.volumes[0].secret.secretName | string | `"vpa-tls-certs"` |  |
| commonLabels | object | `{}` |  |
| containerSecurityContext | object | `{}` |  |
| fullnameOverride | string | `nil` |  |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `nil` |  |
| podSecurityContext.runAsNonRoot | bool | `true` |  |
| podSecurityContext.runAsUser | int | `65534` |  |
| rbac.create | bool | `true` |  |
| rbac.extraRules | list | `[]` |  |
| recommender.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].key | string | `"app.kubernetes.io/component"` |  |
| recommender.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].operator | string | `"In"` |  |
| recommender.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].values[0] | string | `"recommender"` |  |
| recommender.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.topologyKey | string | `"kubernetes.io/hostname"` |  |
| recommender.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight | int | `100` |  |
| recommender.enabled | bool | `true` |  |
| recommender.extraArgs | list | `[]` |  |
| recommender.extraEnv | list | `[]` |  |
| recommender.image.pullPolicy | string | `"IfNotPresent"` |  |
| recommender.image.repository | string | `"registry.k8s.io/autoscaling/vpa-recommender"` |  |
| recommender.image.tag | string | `nil` |  |
| recommender.leaderElection.enabled | string | `nil` |  |
| recommender.leaderElection.leaseDuration | string | `"15s"` |  |
| recommender.leaderElection.renewDeadline | string | `"10s"` |  |
| recommender.leaderElection.resourceName | string | `"vpa-recommender-lease"` |  |
| recommender.leaderElection.resourceNamespace | string | `""` |  |
| recommender.leaderElection.retryPeriod | string | `"2s"` |  |
| recommender.nodeSelector | object | `{}` |  |
| recommender.podAnnotations | object | `{}` |  |
| recommender.podDisruptionBudget.enabled | bool | `true` |  |
| recommender.podDisruptionBudget.maxUnavailable | int or string | `nil` | Maximum number/percentage of pods that can be unavailable after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| recommender.podDisruptionBudget.minAvailable | int or string | `1` | Minimum number/percentage of pods that must be available after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| recommender.podLabels | object | `{}` |  |
| recommender.priorityClassName | string | `nil` |  |
| recommender.replicas | int | `2` |  |
| recommender.resources | object | `{}` |  |
| recommender.serviceAccount.annotations | object | `{}` |  |
| recommender.serviceAccount.create | bool | `true` |  |
| recommender.serviceAccount.labels | object | `{}` |  |
| recommender.tolerations | list | `[]` |  |
| updater.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].key | string | `"app.kubernetes.io/component"` |  |
| updater.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].operator | string | `"In"` |  |
| updater.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].values[0] | string | `"updater"` |  |
| updater.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.topologyKey | string | `"kubernetes.io/hostname"` |  |
| updater.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight | int | `100` |  |
| updater.enabled | bool | `true` |  |
| updater.extraArgs | list | `[]` |  |
| updater.image.pullPolicy | string | `"IfNotPresent"` |  |
| updater.image.repository | string | `"registry.k8s.io/autoscaling/vpa-updater"` |  |
| updater.image.tag | string | `nil` |  |
| updater.leaderElection.enabled | string | `nil` |  |
| updater.leaderElection.leaseDuration | string | `"15s"` |  |
| updater.leaderElection.renewDeadline | string | `"10s"` |  |
| updater.leaderElection.resourceName | string | `"vpa-updater-lease"` |  |
| updater.leaderElection.resourceNamespace | string | `""` |  |
| updater.leaderElection.retryPeriod | string | `"2s"` |  |
| updater.nodeSelector | object | `{}` |  |
| updater.podAnnotations | object | `{}` |  |
| updater.podDisruptionBudget.enabled | bool | `true` |  |
| updater.podDisruptionBudget.maxUnavailable | int or string | `nil` | Maximum number/percentage of pods that can be unavailable after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| updater.podDisruptionBudget.minAvailable | int or string | `1` | Minimum number/percentage of pods that must be available after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| updater.podLabels | object | `{}` |  |
| updater.priorityClassName | string | `nil` |  |
| updater.replicas | int | `2` |  |
| updater.resources | object | `{}` |  |
| updater.serviceAccount.annotations | object | `{}` |  |
| updater.serviceAccount.create | bool | `true` |  |
| updater.serviceAccount.labels | object | `{}` |  |
| updater.tolerations | list | `[]` |  |
