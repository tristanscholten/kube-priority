{{- define "kube-priority-manager.name" -}}kube-priority-manager{{- end -}}
{{- define "kube-priority-manager.namespace" -}}{{ .Release.Namespace }}{{- end -}}
