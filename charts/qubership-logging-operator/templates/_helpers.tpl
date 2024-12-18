{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "fluentd.name" -}}
  {{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "fluentd.fullname" -}}
  {{- if .Values.fullnameOverride -}}
    {{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
  {{- else -}}
    {{- $name := default .Chart.Name .Values.nameOverride -}}
    {{- if contains $name .Release.Name -}}
      {{- .Release.Name | trunc 63 | trimSuffix "-" -}}
    {{- else -}}
      {{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "fluentd.chart" -}}
  {{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "helm-chart.serviceAccountName" -}}
  {{- if .Values.serviceAccount.create -}}
    {{ default (include "helm-chart.fullname" .) .Values.serviceAccount.name }}
  {{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
  {{- end -}}
{{- end -}}

{{/*
Check the major version of Graylog and return 'true' if it equal 5
*/}}
{{- define "graylog.isMajorVersion5" -}}
  {{- if regexMatch "^*:5\\.[0-9]+\\.[0-9]+$" (include "graylog.image" . ) -}}
true
  {{- end -}}
{{- end -}}

{{/*
Return true if generateCerts in TLS is enabled for Graylog HTTP.
*/}}
{{- define "graylog.http.generateCerts.enabled" -}}
  {{- if .Values.graylog.install -}}
    {{- if .Values.graylog.tls -}}
      {{- if .Values.graylog.tls.http -}}
        {{- if .Values.graylog.tls.http.generateCerts -}}
          {{- if .Values.graylog.tls.http.generateCerts.enabled -}}
true
          {{- end -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Return true if generateCerts in TLS is enabled for Graylog input.
*/}}
{{- define "graylog.input.generateCerts.enabled" -}}
  {{- if .Values.graylog.install -}}
    {{- if .Values.graylog.tls -}}
      {{- if .Values.graylog.tls.input -}}
        {{- if .Values.graylog.tls.input.generateCerts -}}
          {{- if .Values.graylog.tls.input.generateCerts.enabled -}}
true
          {{- end -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Return true if generateCerts in TLS is enabled for Fluentd.
*/}}
{{- define "fluentd.generateCerts.enabled" -}}
  {{- if .Values.fluentd.install -}}
    {{- if .Values.fluentd.tls -}}
      {{- if .Values.fluentd.tls.generateCerts -}}
        {{- if .Values.fluentd.tls.generateCerts.enabled -}}
true
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Return true if generateCerts in TLS is enabled for Fluent-bit.
*/}}
{{- define "fluentbit.generateCerts.enabled" -}}
  {{- if .Values.fluentbit.install -}}
    {{- if .Values.fluentbit.tls -}}
      {{- if .Values.fluentbit.tls.generateCerts -}}
        {{- if .Values.fluentbit.tls.generateCerts.enabled -}}
true
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Return true if generateCerts in TLS is enabled for Fluent-bit aggregator.
*/}}
{{- define "fluentbit.aggregator.generateCerts.enabled" -}}
  {{- if .Values.fluentbit.install -}}
    {{- if .Values.fluentbit.aggregator -}}
      {{- if .Values.fluentbit.aggregator.install -}}
        {{- if .Values.fluentbit.aggregator.tls -}}
          {{- if .Values.fluentbit.aggregator.tls.generateCerts -}}
            {{- if .Values.fluentbit.aggregator.tls.generateCerts.enabled -}}
true
            {{- end -}}
          {{- end -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Return true if cacerts for HTTP TLS in Graylog is present and mounted.
*/}}
{{- define "graylog.cacerts.present" -}}
  {{- if .Values.graylog.tls.http.cacerts -}}
true
  {{- else -}}
    {{- if ( include "graylog.http.generateCerts.enabled" . ) -}}
true
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Return true if cert for HTTP TLS in Graylog is present and mounted.
*/}}
{{- define "graylog.cert.present" -}}
  {{- if .Values.graylog.tls.http.cert -}}
    {{- if and .Values.graylog.tls.http.cert.secretName .Values.graylog.tls.http.cert.secretKey }}
true
    {{- else -}}
      {{- if ( include "graylog.http.generateCerts.enabled" . ) -}}
true
      {{- end -}}
    {{- end -}}
  {{- else -}}
    {{- if ( include "graylog.http.generateCerts.enabled" . ) -}}
true
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Return true if key for HTTP TLS in Graylog is present and mounted.
*/}}
{{- define "graylog.key.present" -}}
  {{- if .Values.graylog.tls.http.key -}}
    {{- if and .Values.graylog.tls.http.key.secretName .Values.graylog.tls.http.key.secretKey }}
true
    {{- else -}}
      {{- if (include "graylog.http.generateCerts.enabled" . ) -}}
true
      {{- end -}}
    {{- end -}}
  {{- else -}}
    {{- if (include "graylog.http.generateCerts.enabled" . ) -}}
true
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Set default value for graylog host if not specified in Values.
*/}}
{{- define "graylog.host" -}}
  {{- if .Values.graylog.install -}}
    {{- if not .Values.graylog.host -}}
      {{- if .Values.CLOUD_PUBLIC_HOST -}}
        {{- printf "%s-%s.%s/" "https://graylog" .Values.NAMESPACE (trimSuffix "/" .Values.CLOUD_PUBLIC_HOST) -}}
      {{- end -}}
    {{- else -}}
      {{- $host := trimSuffix "/" .Values.graylog.host -}}
      {{- printf "%s/" $host -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Return secretName if generateCerts in TLS is enabled for Graylog input.
*/}}
{{- define "graylog.secretName" -}}
  {{- if .Values.graylog.install -}}
    {{- if .Values.graylog.tls -}}
      {{- if .Values.graylog.tls.input -}}
        {{- if .Values.graylog.tls.input.cert -}}
            {{- printf "%s" (trimSuffix "/" .Values.graylog.tls.input.cert.secretName) -}}
        {{- else -}}
          {{- if .Values.graylog.tls.input.generateCerts -}}
            {{- if .Values.graylog.tls.input.generateCerts.enabled -}}
              {{- printf "%s" (trimSuffix "/" .Values.graylog.tls.input.generateCerts.secretName) -}}
            {{- end -}}
          {{- end -}}
        {{- end -}}
      {{- end -}}
      {{- if .Values.graylog.tls.http -}}
        {{- if .Values.graylog.tls.http.cert -}}
          {{- printf "%s" (trimSuffix "/" .Values.graylog.tls.http.cert.secretName) -}}
        {{- else -}}
          {{- if .Values.graylog.tls.http.generateCerts -}}
            {{- if .Values.graylog.tls.http.generateCerts.enabled -}}
              {{- printf "%s" (trimSuffix "/" .Values.graylog.tls.http.generateCerts.secretName) -}}
            {{- end -}}
          {{- end -}}
        {{- end -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}


{{- define "logging.monitoredImages" -}}
{{- end -}}

{{/******************************************************************************************************************/}}

{{/*
Find a logging-operator image in various places.
Image can be found from:
* specified by user from .Values.operatorImage
* default value
*/}}
{{- define "logging-operator.image" -}}
  {{- if .Values.operatorImage -}}
    {{- printf "%s" .Values.operatorImage -}}
  {{- else -}}
    {{- print "logging-operator:main" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a graylog image in various places.
Image can be found from:
* specified by user from .Values.operatorImage
* default value
*/}}
{{- define "graylog.image" -}}
  {{- if .Values.graylog.dockerImage -}}
    {{- printf "%s" .Values.graylog.dockerImage -}}
  {{- else -}}
    {{- print "graylog/graylog:5.2.7" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a fluentd image in various places.
Image can be found from:
* specified by user from .Values.fluentd.dockerImage
* default value
*/}}
{{- define "fluentd.image" -}}
  {{- if .Values.fluentd.dockerImage -}}
    {{- printf "%s" .Values.fluentd.dockerImage -}}
  {{- else -}}
    {{- print "qubership-fluentd:main" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a Fluentd ConfigMap reload image in various places.
Image can be found from:
* specified by user from Values.fluentd.configmapReload.image
* default value
*/}}
{{- define "fluentd.configmapReload.image" -}}
  {{- if .Values.fluentd.configmapReload.dockerImage -}}
    {{- printf "%s" .Values.fluentd.configmapReload.dockerImage -}}
  {{- else -}}
    {{- print "jimmidyson/configmap-reload:v0.13.1" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a FluentBit image in various places.
Image can be found from:
* specified by user from .Values.fluentbit.dockerImage
* default value
*/}}
{{- define "fluentbit.image" -}}
  {{- if .Values.fluentbit.dockerImage -}}
    {{- printf "%s" .Values.fluentbit.dockerImage -}}
  {{- else -}}
    {{- print "fluent/fluent-bit:3.2.2" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a FluentBit ConfigMap reload image in various places.
Image can be found from:
* specified by user from .Values.fluentbit.configmapReload.dockerImage
* default value
*/}}
{{- define "fluentbit.configmapReload.image" -}}
  {{- if .Values.fluentbit.configmapReload.dockerImage -}}
    {{- printf "%s" .Values.fluentbit.configmapReload.dockerImage -}}
  {{- else -}}
    {{- print "jimmidyson/configmap-reload:v0.13.1" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a cloud-events-reader image in various places.
Image can be found from:
* specified by user from .Values.cloudEventsReader.dockerImage
* default value
*/}}
{{- define "cloud-events-reader.image" -}}
  {{- if .Values.cloudEventsReader.dockerImage -}}
    {{- printf "%s" .Values.cloudEventsReader.dockerImage -}}
  {{- else -}}
    {{- print "k8s-events-reader:main" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a logging-integration-tests image in various places.
Image can be found from:
* specified by user from .Values.integrationTests.image
* default value
*/}}
{{- define "logging-integration-tests.image" -}}
  {{- if .Values.integrationTests.image -}}
    {{- printf "%s" .Values.integrationTests.image -}}
  {{- else -}}
    {{- print "logging-integration-tests:main" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a graylog-plugins-init-container image in various places.
Image can be found from:
* specified by user from .Values.graylog.initContainerDockerImage
* default value
*/}}
{{- define "graylog-plugins-init.image" -}}
  {{- if .Values.graylog.initContainerDockerImage -}}
    {{- printf "%s" .Values.graylog.initContainerDockerImage -}}
  {{- else -}}
    {{- print "graylog-plugins-init-container:main" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a mongodb image in various places.
Image can be found from:
* specified by user from .Values.graylog.mongodbImage
* default value
*/}}
{{- define "mongodb.image" -}}
  {{- if .Values.graylog.mongodbImage -}}
    {{- printf "%s" .Values.graylog.mongodbImage -}}
  {{- else -}}
    {{- print "mongo:5.0.30" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a authProxy image in various places.
Image can be found from:
* specified by user from .Values.graylog.authProxy.image
* default value
*/}}
{{- define "authProxy.image" -}}
  {{- if .Values.graylog.authProxy.image -}}
    {{- printf "%s" .Values.graylog.authProxy.image -}}
  {{- else -}}
    {{- print "graylog-auth-proxy:main" -}}
  {{- end -}}
{{- end -}}

{{/*
Find a init_setup image in various places.
Image can be found from:
* specified by user from .Values.graylog.initSetupImage
* default value
*/}}
{{- define "init-setup.image" -}}
  {{- if .Values.graylog.initSetupImage -}}
    {{- printf "%s" .Values.graylog.initSetupImage -}}
  {{- else -}}
    {{- print "alpine:3.17.2" -}}
  {{- end -}}
{{- end -}}

{{/*
MongoDB images for sequential upgrades.
Upgrade path:
3.6.23 -> 4.0.28 -> 4.2.22 -> 4.4.17 -> 5.0.19
*/}}

{{/*
MongoDB 4.0 image.
*/}}
{{- define "mongodb40.image" -}}
  {{- if .Values.graylog.mongodb40Image -}}
    {{- printf "%s" .Values.graylog.mongodb40Image -}}
  {{- else -}}
    {{- print "mongo:4.0.28" -}}
  {{- end -}}
{{- end -}}

{{/*
MongoDB 4.2 image.
*/}}
{{- define "mongodb42.image" -}}
  {{- if .Values.graylog.mongodb42Image -}}
    {{- printf "%s" .Values.graylog.mongodb42Image -}}
  {{- else -}}
    {{- print "mongo:4.2.22" -}}
  {{- end -}}
{{- end -}}

{{/*
MongoDB 4.4 image.
*/}}
{{- define "mongodb44.image" -}}
  {{- if .Values.graylog.mongodb44Image -}}
    {{- printf "%s" .Values.graylog.mongodb44Image -}}
  {{- else -}}
    {{- print "mongo:4.4.17" -}}
  {{- end -}}
{{- end -}}
