{{/*
Expand the name of the chart.
*/}}
{{- define "synergy-wholesale-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "synergy-wholesale-exporter.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "synergy-wholesale-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "synergy-wholesale-exporter.labels" -}}
helm.sh/chart: {{ include "synergy-wholesale-exporter.chart" . }}
{{ include "synergy-wholesale-exporter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "synergy-wholesale-exporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "synergy-wholesale-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "synergy-wholesale-exporter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "synergy-wholesale-exporter.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the secret containing Synergy Wholesale credentials.
*/}}
{{- define "synergy-wholesale-exporter.secretName" -}}
{{- default (include "synergy-wholesale-exporter.fullname" .) .Values.secret.name }}
{{- end }}

{{/*
Renders a value that may contain a template.
*/}}
{{- define "synergy-wholesale-exporter.tplRender" -}}
{{- $value := typeIs "string" .value | ternary .value (.value | toYaml) }}
{{- if contains "{{" $value }}
{{- tpl $value .context }}
{{- else }}
{{- $value }}
{{- end }}
{{- end -}}
