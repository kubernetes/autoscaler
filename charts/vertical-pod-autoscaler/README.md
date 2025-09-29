# vertical-pod-autoscaler

WARNING: This chart is currently under development and is not ready for production use.

Automatically adjust resources for your workloads

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square)
![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)
![AppVersion: 1.5.0](https://img.shields.io/badge/AppVersion-1.5.0-informational?style=flat-square)

## Introduction
The Vertical Pod Autoscaler (VPA) automatically adjusts the CPU and memory resource requests of pods to match their actual resource utilization.

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| adrianmoisey | kubernetes-sig-autoscaling@googlegroups.com |  |
| omerap12 | kubernetes-sig-autoscaling@googlegroups.com |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| admissionController.enabled | bool | `true` |  |
| admissionController.extraArgs | list | `[]` |  |
| admissionController.extraEnv | list | `[]` |  |
| admissionController.image.pullPolicy | string | `"IfNotPresent"` |  |
| admissionController.image.repository | string | `"registry.k8s.io/autoscaling/vpa-admission-controller"` |  |
| admissionController.image.tag | string | `nil` |  |
| admissionController.nodeSelector | object | `{}` |  |
| admissionController.podAnnotations | object | `{}` |  |
| admissionController.podLabels | object | `{}` |  |
| admissionController.replicas | int | `1` |  |
| admissionController.resources.limits.cpu | string | `"200m"` |  |
| admissionController.resources.limits.memory | string | `"500Mi"` |  |
| admissionController.resources.requests.cpu | string | `"50m"` |  |
| admissionController.resources.requests.memory | string | `"200Mi"` |  |
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
| admissionController.volumeMounts[0].mountPath | string | `"/etc/tls-certs"` |  |
| admissionController.volumeMounts[0].name | string | `"tls-certs"` |  |
| admissionController.volumeMounts[0].readOnly | bool | `true` |  |
| admissionController.volumes[0].name | string | `"tls-certs"` |  |
| admissionController.volumes[0].secret.defaultMode | int | `420` |  |
| admissionController.volumes[0].secret.secretName | string | `"vpa-tls-certs"` |  |
| commonLabels | object | `{}` |  |
| fullnameOverride | string | `nil` |  |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `nil` |  |
| rbac.create | bool | `true` |  |