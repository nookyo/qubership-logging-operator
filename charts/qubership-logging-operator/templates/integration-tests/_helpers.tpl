{{/*
Get Custom Resource plural from path in Values
*/}}
{{- define "integrationTests.plural_custom_resource" -}}
{{- printf "%v" (index (regexSplit "/" .Values.integrationTests.statusWriting.customResourcePath 5) 3) }}
{{- end -}}

{{/*
Get Custom Resource apiGroup from path in Values
*/}}
{{- define "integrationTests.apigroup_custom_resource" -}}
{{- printf "%v" (index (regexSplit "/" .Values.integrationTests.statusWriting.customResourcePath 5) 0) }}
{{- end -}}
