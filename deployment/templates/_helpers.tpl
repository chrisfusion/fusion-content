{{/*
Standard Helm labels applied to every resource.
*/}}
{{- define "fusion-content.labels" -}}
app.kubernetes.io/name: fusion-content
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end }}

{{/*
ServiceAccount name for the server.
*/}}
{{- define "fusion-content.serverSAName" -}}
fusion-content-server
{{- end }}

{{/*
Name of the Secret that contains repos.yaml.
Prefers existingSecret; falls back to chart-generated "<release>-repos".
*/}}
{{- define "fusion-content.reposSecretName" -}}
{{- if .Values.repos.existingSecret -}}
{{- .Values.repos.existingSecret -}}
{{- else -}}
{{- printf "%s-repos" .Release.Name -}}
{{- end -}}
{{- end }}
