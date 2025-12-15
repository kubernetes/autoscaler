# vertical-pod-autoscaler

WARNING: This chart is currently under development and is not ready for production use.

Automatically adjust resources for your workloads

![Version: 0.8.0](https://img.shields.io/badge/Version-0.8.0-informational?style=flat-square)
![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)
![AppVersion: 1.5.1](https://img.shields.io/badge/AppVersion-1.5.1-informational?style=flat-square)

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

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| admissionController.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].key | string | `"app.kubernetes.io/component"` |  |
| admissionController.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].operator | string | `"In"` |  |
| admissionController.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].values[0] | string | `"admission-controller"` |  |
| admissionController.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].topologyKey | string | `"kubernetes.io/hostname"` |  |
| admissionController.enabled | bool | `true` |  |
| admissionController.extraArgs | list | `[]` |  |
| admissionController.extraEnv | list | `[]` |  |
| admissionController.image.pullPolicy | string | `"IfNotPresent"` |  |
| admissionController.image.repository | string | `"registry.k8s.io/autoscaling/vpa-admission-controller"` |  |
| admissionController.image.tag | string | `nil` |  |
| admissionController.nodeSelector | object | `{}` |  |
| admissionController.podAnnotations | object | `{}` |  |
| admissionController.podDisruptionBudget.enabled | bool | `true` |  |
| admissionController.podDisruptionBudget.maxUnavailable | int or string | `nil` | Maximum number/percentage of pods that can be unavailable after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| admissionController.podDisruptionBudget.minAvailable | int or string | `1` | Minimum number/percentage of pods that must be available after the eviction. IMPORTANT: You can specify either 'minAvailable' or 'maxUnavailable', but not both. |
| admissionController.podLabels | object | `{}` |  |
| admissionController.priorityClassName | string | `nil` |  |
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
| admissionController.tls.existingSecret | string | `""` |  |
| admissionController.tls.key | string | `""` |  |
| admissionController.tls.secretName | string | `"vpa-tls-certs"` |  |
| admissionController.tolerations | list | `[]` |  |
| admissionController.volumeMounts[0].mountPath | string | `"/etc/tls-certs"` |  |
| admissionController.volumeMounts[0].name | string | `"tls-certs"` |  |
| admissionController.volumeMounts[0].readOnly | bool | `true` |  |
| admissionController.volumes[0].name | string | `"tls-certs"` |  |
| admissionController.volumes[0].secret.defaultMode | int | `420` |  |
| admissionController.volumes[0].secret.secretName | string | `"vpa-tls-certs"` |  |
| commonLabels | object | `{}` |  |
| containerSecurityContext | object | `{}` |  |
| fullnameOverride | string | `nil` |  |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `nil` |  |
| podSecurityContext.runAsNonRoot | bool | `true` |  |
| podSecurityContext.runAsUser | int | `65534` |  |
| rbac.create | bool | `true` |  |
| recommender.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].key | string | `"app.kubernetes.io/component"` |  |
| recommender.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].operator | string | `"In"` |  |
| recommender.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].values[0] | string | `"recommender"` |  |
| recommender.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].topologyKey | string | `"kubernetes.io/hostname"` |  |
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
| updater.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].key | string | `"app.kubernetes.io/component"` |  |
| updater.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].operator | string | `"In"` |  |
| updater.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].values[0] | string | `"updater"` |  |
| updater.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].topologyKey | string | `"kubernetes.io/hostname"` |  |
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
