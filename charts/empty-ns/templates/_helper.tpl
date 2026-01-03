{{ define "empty-ns.fullname" -}}
{{ .Chart.Name }}
{{- end }}

{{ define "empty-ns.name" -}}
{{ .Chart.Name }}
{{- end }}

{{ define "empty-ns.chart" -}}
{{ .Chart.Name }}
{{- end }}

{{ define "empty-ns.labels" }}
app.kubernetes.io/name: {{ include "empty-ns.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
kubernetes.wifu.co.uk/app: {{ include "empty-ns.name" . }}
kubernetes.wifu.co.uk/app-version: {{ printf "%s-%s" .Chart.Name .Chart.AppVersion }}
kubernetes.wifu.co.uk/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version }}
{{ end }}

{{ define "empty-ns.sa-name" -}}
{{ include "empty-ns.fullname" . }}
{{- end }}