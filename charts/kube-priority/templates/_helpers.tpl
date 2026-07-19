{{- define "kube-priority.name" -}}kube-priority{{- end -}}
{{- define "kube-priority.namespace" -}}{{ .Release.Namespace }}{{- end -}}
