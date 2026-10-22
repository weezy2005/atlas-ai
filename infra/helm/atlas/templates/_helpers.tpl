{{/* Effective config for one service: defaults deep-merged under the service's own values. */}}
{{- define "atlas.svc" -}}
{{- $svc := deepCopy .svc -}}
{{- $merged := mustMergeOverwrite (deepCopy .root.Values.defaults) $svc -}}
{{- if not (hasKey $merged "port") }}{{- $_ := set $merged "port" 8000 }}{{- end -}}
{{- toYaml $merged -}}
{{- end -}}

{{- define "atlas.labels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/part-of: atlas
app.kubernetes.io/instance: {{ .root.Release.Name }}
app.kubernetes.io/managed-by: {{ .root.Release.Service }}
helm.sh/chart: {{ .root.Chart.Name }}-{{ .root.Chart.Version }}
{{- end -}}

{{- define "atlas.selector" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .root.Release.Name }}
{{- end -}}

{{- define "atlas.image" -}}
{{- $img := .root.Values.image -}}
{{- if $img.registry }}{{ $img.registry }}/{{ end }}{{ $img.prefix }}{{ .name }}:{{ $img.tag }}
{{- end -}}
