{{/*
Chart
*/}}
{{- define "vertical-pod-autoscaler.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "vertical-pod-autoscaler.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "vertical-pod-autoscaler.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "vertical-pod-autoscaler.labels" -}}
helm.sh/chart: {{ include "vertical-pod-autoscaler.chart" . }}
{{ include "vertical-pod-autoscaler.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{- define "vertical-pod-autoscaler.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vertical-pod-autoscaler.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{/*
admissionController
*/}}
{{- define "vertical-pod-autoscaler.admissionController.fullname" -}}
{{ include "vertical-pod-autoscaler.fullname" . }}-admission-controller
{{- end }}

{{- define "vertical-pod-autoscaler.admissionController.labels" -}}
{{ include "vertical-pod-autoscaler.labels" . }}
app.kubernetes.io/component: admission-controller
app.kubernetes.io/component-instance: {{ .Release.Name }}-admission-controller
{{- end }}

{{- define "vertical-pod-autoscaler.admissionController.selectorLabels" -}}
{{ include "vertical-pod-autoscaler.selectorLabels" . }}
app.kubernetes.io/component: admission-controller
{{- end }}

{{- define "vertical-pod-autoscaler.admissionController.image" -}}
{{- printf "%s:%s" .Values.admissionController.image.repository (default .Chart.AppVersion .Values.admissionController.image.tag) }}
{{- end }}

{{/*
Create the name of the tls secret to use
*/}}
{{- define "vertical-pod-autoscaler.admissionController.tls.secretName" -}}
{{- if .Values.admissionController.tls.existingSecret -}}
    {{ .Values.admissionController.tls.existingSecret }}
{{- else -}}
    {{- printf "%s-%s" (include "vertical-pod-autoscaler.admissionController.fullname" .) "tls" | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}


{{/*
Create the name of the namespace to use
*/}}
{{- define "common.names.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}