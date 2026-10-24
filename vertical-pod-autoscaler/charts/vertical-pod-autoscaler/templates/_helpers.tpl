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
{{- if .Values.admissionController.tls.secretName -}}
    {{ .Values.admissionController.tls.secretName }}
{{- else -}}
    {{- printf "%s-%s" (include "vertical-pod-autoscaler.admissionController.fullname" .) "tls" | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
admissionController webhook
*/}}

{{- define "vertical-pod-autoscaler.admissionController.webhook.configName" -}}
{{ include "vertical-pod-autoscaler.fullname" . }}-webhook-config
{{- end }}

{{/*
admissionController certManager
*/}}
{{- define "vertical-pod-autoscaler.admissionController.certManager.certName" -}}
{{ include "vertical-pod-autoscaler.fullname" . }}-webhook-cert
{{- end }}

{{- define "vertical-pod-autoscaler.admissionController.certManager.issuerName" -}}
{{ include "vertical-pod-autoscaler.fullname" . }}-selfsigned
{{- end }}

{{- define "vertical-pod-autoscaler.admissionController.certManager.caName" -}}
{{ include "vertical-pod-autoscaler.fullname" . }}-webhook-ca
{{- end }}

{{- define "vertical-pod-autoscaler.admissionController.certManager.caIssuerName" -}}
{{ include "vertical-pod-autoscaler.fullname" . }}-ca
{{- end }}

{{/*
admissionController generated TLS data (tls.create). Computed once per render and cached on
.Values so the Secret and the webhook caBundle always come from the same CA.
Precedence: user-supplied tls.caCert/cert/key, then the existing Secret (lookup), then genCA/genSignedCert.
*/}}
{{- define "vertical-pod-autoscaler.admissionController.tls.data" -}}
{{- $tls := .Values.admissionController.tls -}}
{{- if not (hasKey $tls "_data") -}}
  {{- $svc := .Values.admissionController.service.name -}}
  {{- $ns := .Release.Namespace -}}
  {{- $data := dict -}}
  {{- if and $tls.caCert $tls.cert $tls.key -}}
    {{- $data = dict "ca" ($tls.caCert | b64enc) "cert" ($tls.cert | b64enc) "key" ($tls.key | b64enc) -}}
  {{- else -}}
    {{- $existing := lookup "v1" "Secret" $ns (include "vertical-pod-autoscaler.admissionController.tls.secretName" .) -}}
    {{- $oldCa := dig "data" "ca" "" $existing -}}
    {{- $oldCert := dig "data" "cert" "" $existing -}}
    {{- $oldKey := dig "data" "key" "" $existing -}}
    {{- if and $oldCa $oldCert $oldKey -}}
      {{- $data = dict "ca" $oldCa "cert" $oldCert "key" $oldKey -}}
    {{- else -}}
      {{- $days := int $tls.validityDays -}}
      {{- $names := list $svc (printf "%s.%s" $svc $ns) (printf "%s.%s.svc" $svc $ns) (printf "%s.%s.svc.cluster.local" $svc $ns) -}}
      {{- $ca := genCA (printf "%s-ca" $svc) $days -}}
      {{- $cert := genSignedCert (printf "%s.%s.svc" $svc $ns) nil $names $days $ca -}}
      {{- $data = dict "ca" ($ca.Cert | b64enc) "cert" ($cert.Cert | b64enc) "key" ($cert.Key | b64enc) -}}
    {{- end -}}
  {{- end -}}
  {{- $_ := set $tls "_data" $data -}}
{{- end -}}
{{- end -}}


{{/*
updater
*/}}
{{- define "vertical-pod-autoscaler.updater.fullname" -}}
{{ include "vertical-pod-autoscaler.fullname" . }}-updater
{{- end }}

{{- define "vertical-pod-autoscaler.updater.labels" -}}
{{ include "vertical-pod-autoscaler.labels" . }}
app.kubernetes.io/component: updater
app.kubernetes.io/component-instance: {{ .Release.Name }}-updater
{{- end }}

{{- define "vertical-pod-autoscaler.updater.selectorLabels" -}}
{{ include "vertical-pod-autoscaler.selectorLabels" . }}
app.kubernetes.io/component: updater
{{- end }}

{{- define "vertical-pod-autoscaler.updater.image" -}}
{{- printf "%s:%s" .Values.updater.image.repository (default .Chart.AppVersion .Values.updater.image.tag) }}
{{- end }}

{{- define "vertical-pod-autoscaler.updater.leaderElectionEnabled" -}}
{{- if and (eq .Values.updater.leaderElection.enabled nil) (gt (int .Values.updater.replicas) 1) -}}
true
{{- else if .Values.updater.leaderElection.enabled -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}


{{/*
Create the name of the namespace to use
*/}}
{{- define "common.names.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
recommender
*/}}
{{- define "vertical-pod-autoscaler.recommender.fullname" -}}
{{ include "vertical-pod-autoscaler.fullname" . }}-recommender
{{- end }}

{{- define "vertical-pod-autoscaler.recommender.labels" -}}
{{ include "vertical-pod-autoscaler.labels" . }}
app.kubernetes.io/component: recommender
app.kubernetes.io/component-instance: {{ .Release.Name }}-recommender
{{- end }}

{{- define "vertical-pod-autoscaler.recommender.selectorLabels" -}}
{{ include "vertical-pod-autoscaler.selectorLabels" . }}
app.kubernetes.io/component: recommender
{{- end }}

{{- define "vertical-pod-autoscaler.recommender.image" -}}
{{- printf "%s:%s" .Values.recommender.image.repository (default .Chart.AppVersion .Values.recommender.image.tag) }}
{{- end }}


{{/*
extraObjects
*/}}

{{/*
Render one extraObjects entry: a string is evaluated as a template as-is, a map is converted to YAML first.
The context must be the root context ($) so that tpl can resolve .Release, .Values and the named templates.
*/}}
{{- define "vertical-pod-autoscaler.tplvalues.render" -}}
{{- if typeIs "string" .value -}}
{{- tpl .value .context -}}
{{- else -}}
{{- tpl (.value | toYaml) .context -}}
{{- end -}}
{{- end }}
