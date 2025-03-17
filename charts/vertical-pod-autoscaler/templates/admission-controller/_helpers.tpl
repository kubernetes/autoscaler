{{- define "vertical-pod-autoscaler.admissionController.fullname" -}}
{{- printf "%s-%s" .Release.Name "admission-controller" -}}
{{- end -}}


{{- define "vertical-pod-autoscaler.admissionController.serviceAccount.name" -}}
{{- printf "%s-%s" .Release.Name "admission-controller" -}}
{{- end -}}
