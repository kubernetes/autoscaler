{{- define "vertical-pod-autoscaler.updater.fullname" -}}
{{- printf "%s-%s" .Release.Name "updater" -}}
{{- end -}}


{{- define "vertical-pod-autoscaler.updater.serviceAccount.name" -}}
{{- printf "%s-%s" .Release.Name "updater" -}}
{{- end -}}
