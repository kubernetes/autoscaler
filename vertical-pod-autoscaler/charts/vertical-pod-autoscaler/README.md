# vertical-pod-autoscaler

WARNING: This chart is currently under development and is not ready for production use.

Automatically adjust resources for your workloads

![Version: 0.12.0](https://img.shields.io/badge/Version-0.12.0-informational?style=flat-square)
![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)
![AppVersion: 1.7.1](https://img.shields.io/badge/AppVersion-1.7.1-informational?style=flat-square)

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
The admission controller requires a `MutatingWebhookConfiguration` and TLS certificates. This chart supports three mutually exclusive modes:

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

### cert-manager managed

Using an existing `Issuer` or `ClusterIssuer`:
```yaml
admissionController:
  registerWebhook: false
  certGen:
    enabled: false
  certManager:
    enabled: true
    issuerRef:
      name: my-issuer
      kind: ClusterIssuer
```

Or letting the chart create a namespaced self-signed issuer:
```yaml
admissionController:
  registerWebhook: false
  certGen:
    enabled: false
  certManager:
    enabled: true
    createSelfSignedIssuer:
      enabled: true
```
In this mode:
- Helm creates the MutatingWebhookConfiguration
- cert-manager issues and renews TLS certificates automatically
- cert-manager's cainjector injects the CA into the webhook configuration

By default, you must provide an existing `Issuer` or `ClusterIssuer` via `admissionController.certManager.issuerRef`. Alternatively, enable `admissionController.certManager.createSelfSignedIssuer.enabled: true` to let the chart create a namespaced self-signed issuer automatically.

## Custom Resource Definitions

The VPA CRDs are managed as regular chart templates and are kept in sync automatically on `helm upgrade` (set `crds.enabled: false` if you manage CRDs separately, e.g. via Argo CD or Fleet).

By default, the CRDs are annotated with `helm.sh/resource-policy: keep`, so `helm uninstall` will not remove them, protecting any existing VPA objects from being deleted. Set `crds.keep: false` to disable this.

> [!WARNING]
> If you're upgrading from a chart version before 0.12.0 (when CRDs moved from `crds/` to `templates/`) and `crds.enabled` is `true`, the upgrade will fail with an "invalid ownership metadata" error unless you use `--take-ownership` or apply the manual fix below first. If `crds.enabled` is `false`, these CRDs are not rendered by the chart. This is because Helm never took ownership of CRDs installed via the old `crds/` folder, so it refuses to adopt them without one of these steps. This only needs to be done once, on your first upgrade past 0.12.0.

If you're on Helm 3.17+, pass `--take-ownership` on that upgrade:

```bash
helm upgrade <release-name> <chart> --namespace <namespace> --take-ownership
```

On older Helm versions `--take-ownership` isn't available, so add the ownership metadata manually first:

```bash
kubectl label crd verticalpodautoscalercheckpoints.autoscaling.k8s.io app.kubernetes.io/managed-by=Helm --overwrite
kubectl annotate crd verticalpodautoscalercheckpoints.autoscaling.k8s.io meta.helm.sh/release-name=<release-name> meta.helm.sh/release-namespace=<namespace> --overwrite

kubectl label crd verticalpodautoscalers.autoscaling.k8s.io app.kubernetes.io/managed-by=Helm --overwrite
kubectl annotate crd verticalpodautoscalers.autoscaling.k8s.io meta.helm.sh/release-name=<release-name> meta.helm.sh/release-namespace=<namespace> --overwrite
```

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
| admissionController.annotations | object | `{}` |  |
| admissionController.certGen.affinity | object | `{}` |  |
| admissionController.certGen.enabled | bool | `true` |  |
| admissionController.certGen.env | object | `{}` | Additional environment variables to be added to the certgen container. Format is KEY: Value format |
| admissionController.certGen.image.pullPolicy | string | `"IfNotPresent"` | The pull policy for the certgen image. Recommend not changing this |
| admissionController.certGen.image.repository | string | `"registry.k8s.io/ingress-nginx/kube-webhook-certgen"` | An image that contains certgen for creating certificates. |
| admissionController.certGen.image.tag | string | `"v20231011-8b53cabe0"` | An image tag for the admissionController.certGen.image.repository image. |
| admissionController.certGen.nodeSelector | object | `{}` |  |
| admissionController.certGen.podSecurityContext | object | `{"runAsNonRoot":true,"runAsUser":65534,"seccompProfile":{"type":"RuntimeDefault"}}` | The securityContext block for the certgen pod(s) |
| admissionController.certGen.priorityClassName | string | `""` | Priority class name for the certgen job pods. These jobs gate the release as pre-install/pre-upgrade and post-install/post-upgrade hooks, so a priority class can help them schedule on a busy cluster. |
| admissionController.certGen.resources | object | `{}` | The resources block for the certgen pod |
| admissionController.certGen.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true}` | The securityContext block for the certgen container(s) |
| admissionController.certGen.tolerations | list | `[]` |  |
| admissionController.certManager.annotations | object | `{}` | Annotations to add to all cert-manager resources created by Chart. |
| admissionController.certManager.createSelfSignedIssuer | object | `{"duration":"8760h","enabled":false,"renewBefore":"720h"}` | Optionally create a SelfSigned Issuer for CA generation. When enabled, a namespaced SelfSigned Issuer is created in the VPA namespace to issue the intermediate CA certificate, which in turn signs the webhook TLS certificate. |
| admissionController.certManager.createSelfSignedIssuer.duration | string | `"8760h"` | Lifetime of the intermediate CA certificate. |
| admissionController.certManager.createSelfSignedIssuer.renewBefore | string | `"720h"` | Time before expiry to renew the CA certificate. |
| admissionController.certManager.duration | string | `"168h"` | Lifetime of the webhook TLS certificate. |
| admissionController.certManager.enabled | bool | `false` | If true, cert-manager manages the webhook certificate lifecycle. cert-manager must be installed in the cluster, see https://cert-manager.io/docs/installation. Mutually exclusive with certGen.enabled, registerWebhook, and tls.create. |
| admissionController.certManager.issuerRef | object | `{"group":"cert-manager.io","kind":"ClusterIssuer","name":""}` | Reference to an existing issuer for signing the webhook TLS certificate. Required when createSelfSignedIssuer.enabled is false. |
| admissionController.certManager.issuerRef.group | string | `"cert-manager.io"` | API group of the issuer. |
| admissionController.certManager.issuerRef.kind | string | `"ClusterIssuer"` | Kind of the issuer (ClusterIssuer or Issuer). |
| admissionController.certManager.issuerRef.name | string | `""` | Name of the issuer. |
| admissionController.certManager.privateKey.algorithm | string | `"RSA"` | Key algorithm for certificates (RSA, ECDSA, Ed25519). |
| admissionController.certManager.privateKey.size | int | `2048` | Key size for RSA or ECDSA. Ignored for Ed25519. |
| admissionController.certManager.renewBefore | string | `"24h"` | Time before expiry to renew the TLS certificate. |
| admissionController.enabled | bool | `true` |  |
| admissionController.extraArgs | list | `[]` |  |
| admissionController.extraEnv | list | `[]` |  |
| admissionController.hostNetwork | bool | `false` | Enable host network for the admission controller pod. Set to true when the pod needs direct access to the host's network namespace. Note: this bypasses Kubernetes network isolation and may cause port conflicts if multiple replicas run on the same node. |
| admissionController.image.pullPolicy | string | `"IfNotPresent"` |  |
| admissionController.image.repository | string | `"registry.k8s.io/autoscaling/vpa-admission-controller"` |  |
| admissionController.image.tag | string | `nil` |  |
| admissionController.logLevel | int | `4` | Log verbosity for the Admission Controller (klog -v). |
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
| admissionController.revisionHistoryLimit | int | `10` |  |
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
| admissionController.topologySpreadConstraints | list | `[]` | Topology spread constraints for scheduling the Admission Controller, used to spread replicas across failure domains such as zones. |
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
| crds.enabled | bool | `true` | Whether to install and manage the VPA CRDs. Disable if you manage CRDs separately. |
| crds.keep | bool | `true` | Whether to add the helm.sh/resource-policy: keep annotation to the CRDs, so they are not removed by `helm uninstall`. |
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
| recommender.annotations | object | `{}` |  |
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
| recommender.logLevel | int | `4` | Log verbosity for the Recommender (klog -v). |
| recommender.nodeSelector | object | `{}` |  |
| recommender.podAnnotations | object | `{}` |  |
| recommender.podDisruptionBudget.enabled | bool | `true` |  |
| recommender.podDisruptionBudget.maxUnavailable | int or string | `nil` | Maximum number/percentage of pods that can be unavailable after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| recommender.podDisruptionBudget.minAvailable | int or string | `1` | Minimum number/percentage of pods that must be available after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| recommender.podLabels | object | `{}` |  |
| recommender.priorityClassName | string | `nil` |  |
| recommender.replicas | int | `2` |  |
| recommender.resources | object | `{}` |  |
| recommender.revisionHistoryLimit | int | `10` |  |
| recommender.serviceAccount.annotations | object | `{}` |  |
| recommender.serviceAccount.create | bool | `true` |  |
| recommender.serviceAccount.labels | object | `{}` |  |
| recommender.tolerations | list | `[]` |  |
| recommender.topologySpreadConstraints | list | `[]` | Topology spread constraints for scheduling the Recommender, used to spread replicas across failure domains such as zones. |
| updater.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].key | string | `"app.kubernetes.io/component"` |  |
| updater.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].operator | string | `"In"` |  |
| updater.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.labelSelector.matchExpressions[0].values[0] | string | `"updater"` |  |
| updater.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].podAffinityTerm.topologyKey | string | `"kubernetes.io/hostname"` |  |
| updater.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].weight | int | `100` |  |
| updater.annotations | object | `{}` |  |
| updater.enabled | bool | `true` |  |
| updater.extraArgs[0] | string | `"--in-place-skip-disruption-budget=true"` |  |
| updater.image.pullPolicy | string | `"IfNotPresent"` |  |
| updater.image.repository | string | `"registry.k8s.io/autoscaling/vpa-updater"` |  |
| updater.image.tag | string | `nil` |  |
| updater.leaderElection.enabled | string | `nil` |  |
| updater.leaderElection.leaseDuration | string | `"15s"` |  |
| updater.leaderElection.renewDeadline | string | `"10s"` |  |
| updater.leaderElection.resourceName | string | `"vpa-updater-lease"` |  |
| updater.leaderElection.resourceNamespace | string | `""` |  |
| updater.leaderElection.retryPeriod | string | `"2s"` |  |
| updater.logLevel | int | `4` | Log verbosity for the Updater (klog -v). |
| updater.nodeSelector | object | `{}` |  |
| updater.podAnnotations | object | `{}` |  |
| updater.podDisruptionBudget.enabled | bool | `true` |  |
| updater.podDisruptionBudget.maxUnavailable | int or string | `nil` | Maximum number/percentage of pods that can be unavailable after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| updater.podDisruptionBudget.minAvailable | int or string | `1` | Minimum number/percentage of pods that must be available after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| updater.podLabels | object | `{}` |  |
| updater.priorityClassName | string | `nil` |  |
| updater.replicas | int | `2` |  |
| updater.resources | object | `{}` |  |
| updater.revisionHistoryLimit | int | `10` |  |
| updater.serviceAccount.annotations | object | `{}` |  |
| updater.serviceAccount.create | bool | `true` |  |
| updater.serviceAccount.labels | object | `{}` |  |
| updater.tolerations | list | `[]` |  |
| updater.topologySpreadConstraints | list | `[]` | Topology spread constraints for scheduling the Updater, used to spread replicas across failure domains such as zones. |
