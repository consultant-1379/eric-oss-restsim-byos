{{/* vim: set filetype=mustache: */}}

{{/*
Create a map from ".Values.global" with defaults if missing in values file.
This hides defaults from values file.
*/}}
{{ define "eric-oss-byos-buildmanager.global" }}
  {{- $globalDefaults := dict "security" (dict "tls" (dict "enabled" true)) -}}
  {{- $globalDefaults := merge $globalDefaults (dict "nodeSelector" (dict)) -}}
  {{- $globalDefaults := merge $globalDefaults (dict "registry" (dict "url" "armdocker.rnd.ericsson.se")) -}}
  {{- $globalDefaults := merge $globalDefaults (dict "pullSecret" "") -}}
  {{- $globalDefaults := merge $globalDefaults (dict "timezone" "UTC") -}}
  {{- $globalDefaults := merge $globalDefaults (dict "externalIPv4" (dict "enabled")) -}}
  {{- $globalDefaults := merge $globalDefaults (dict "externalIPv6" (dict "enabled")) -}}
  {{ if .Values.global }}
    {{- mergeOverwrite $globalDefaults .Values.global | toJson -}}
  {{ else }}
    {{- $globalDefaults | toJson -}}
  {{ end }}
{{ end }}
{{- define "eric-oss-byos-buildmanager.mainImagePath" }}
    {{- $productInfo := fromYaml (.Files.Get "eric-product-info.yaml") -}}
    {{- $registryUrl := (index $productInfo "images" "eric-oss-byos-buildmanager" "registry") -}}
    {{- $repoPath := (index $productInfo "images" "eric-oss-byos-buildmanager" "repoPath") -}}
    {{- $name := (index $productInfo "images" "eric-oss-byos-buildmanager" "name") -}}
    {{- $tag := (index $productInfo "images" "eric-oss-byos-buildmanager" "tag") -}}
    {{- if .Values.global -}}
        {{- if .Values.global.registry -}}
            {{- if .Values.global.registry.url -}}
                {{- $registryUrl = .Values.global.registry.url -}}
            {{- end -}}
            {{- if not (kindIs "invalid" .Values.global.registry.repoPath) -}}
                {{- $repoPath = .Values.global.registry.repoPath -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
    {{- if .Values.imageCredentials -}}
        {{- if .Values.imageCredentials.registry -}}
            {{- if .Values.imageCredentials.registry.url -}}
                {{- $registryUrl = .Values.imageCredentials.registry.url -}}
            {{- end -}}
        {{- end -}}
        {{- if not (kindIs "invalid" .Values.imageCredentials.repoPath) -}}
            {{- $repoPath = .Values.imageCredentials.repoPath -}}
        {{- end -}}
        {{- if .Values.imageCredentials.mainImage -}}
            {{- if .Values.imageCredentials.mainImage.registry -}}
                {{- if .Values.imageCredentials.mainImage.registry.url -}}
                    {{- $registryUrl = .Values.imageCredentials.mainImage.registry.url -}}
                {{- end -}}
            {{- end -}}
            {{- if not (kindIs "invalid" .Values.imageCredentials.mainImage.repoPath) -}}
                {{- $repoPath = .Values.imageCredentials.mainImage.repoPath -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
    {{- if $repoPath -}}
        {{- $repoPath = printf "%s/" $repoPath -}}
    {{- end -}}
    {{- printf "%s/%s%s:%s" $registryUrl $repoPath $name $tag -}}
{{- end -}}


{{/*
Expand the name of the chart.
*/}}
{{- define "eric-oss-byos-buildmanager.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "eric-oss-byos-buildmanager.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Create chart version as used by the chart label.
*/}}
{{- define "eric-oss-byos-buildmanager.version" -}}
{{- printf "%s" .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Create image registry url

{{- define "eric-oss-byos-buildmanager.registryUrl" -}}
    {{- $g := fromJson (include "eric-oss-byos-buildmanager.global" .) -}}
    {{- $registryURL := $g.registry.url -}}
    {{- if .Values.imageCredentials.registry -}}
        {{- if .Values.imageCredentials.registry.url -}}
            {{- $registryURL = .Values.imageCredentials.registry.url -}}
        {{- end -}}
    {{- end -}}
    {{- print $registryURL -}}
{{- end -}}
*/}}


{{/*
Create image pull secrets
*/}}

{{/*
{{- define "eric-oss-byos-buildmanager.pullSecrets" -}}
  {{- $g := fromJson (include "eric-oss-byos-buildmanager.global" .) -}}
  {{- $pullSecret := $g.pullSecret -}}
  {{- if .Values.imageCredentials -}}
      {{- if .Values.imageCredentials.pullSecret -}}
          {{- $pullSecret = .Values.imageCredentials.pullSecret -}}
      {{- end -}}
  {{- end -}}
  {{- print $pullSecret -}}
{{- end -}}
*/}}

{{/*
Create image pull secrets
*/}}

{{- define "eric-oss-byos-buildmanager.pullSecrets" -}}
{{- if .Values.imageCredentials.pullSecret -}}
{{- print .Values.imageCredentials.pullSecret -}}
{{- else if .Values.global.pullSecret -}}
{{- print .Values.global.pullSecret -}}
{{- end -}}
{{- end -}}


{{/*
Timezone variable
*/}}
{{- define "eric-oss-byos-buildmanager.timezone" -}}
{{- $g := fromJson (include "eric-oss-byos-buildmanager.global" .) -}}
{{- print $g.timezone | quote -}}
{{- end -}}


{{/*
Create image repo path

{{- define "eric-oss-byos-buildmanager.repoPath" -}}
{{- if .Values.imageCredentials.repoPath -}}
{{- print .Values.imageCredentials.repoPath "/" -}}
{{- end -}}
{{- end -}}
*/}}

{{/*
    Define Image Pull Policy
*/}}
{{- define "eric-oss-byos-buildmanager.registryImagePullPolicy" -}}
    {{- $globalRegistryPullPolicy := "IfNotPresent" -}}
    {{- if .Values.global -}}
        {{- if .Values.global.registry -}}
            {{- if .Values.global.registry.imagePullPolicy -}}
                {{- $globalRegistryPullPolicy = .Values.global.registry.imagePullPolicy -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
    {{- print $globalRegistryPullPolicy -}}
{{- end -}}

Create a user defined label (DR-D1121-068, DR-D1121-060)
*/}}
{{ define "eric-oss-byos-buildmanager.config-labels" }}
  {{- $g := fromJson (include "eric-oss-byos-buildmanager.global" .) -}}
  {{- $global := $g.labels -}}
  {{- $service := .Values.labels -}}
  {{- include "eric-oss-byos-buildmanager.mergeLabels" (dict "location" .Template.Name "sources" (list $global $service)) -}}
{{- end }}

{{/*
Define the labels
*/}}
{{- define "eric-oss-byos-buildmanager.labels" -}}
  {{- $productLabels := include "eric-oss-byos-buildmanager.product-labels" . | fromYaml -}}
  {{- $config := include "eric-oss-byos-buildmanager.config-labels" . | fromYaml -}}
  {{- include "eric-oss-byos-buildmanager.mergeLabels" (dict "location" .Template.Name "sources" (list $productLabels $config))  -}}
{{- end -}}

{{/*
Create a user defined annotation (DR-D1121-065, DR-D1121-060)
*/}}
{{ define "eric-oss-byos-buildmanager.config-annotations" }}
  {{- $global := (.Values.global).annotations -}}
  {{- $service := .Values.annotations -}}
  {{- include "eric-oss-byos-buildmanager.mergeAnnotations" (dict "location" .Template.Name "sources" (list $global $service)) -}}
{{- end }}

{{/*
Create Ericsson product specific annotations
*/}}
{{- define "eric-oss-byos-buildmanager.product-info" }}
ericsson.com/product-name: {{ (fromYaml (.Files.Get "eric-product-info.yaml")).productName | quote }}
ericsson.com/product-number: {{ (fromYaml (.Files.Get "eric-product-info.yaml")).productNumber | quote }}
ericsson.com/product-revision: {{regexReplaceAll "(.*)[+-].*" .Chart.Version "${1}" }}
{{- end -}}

{{- define "eric-oss-byos-buildmanager.annotations" -}}
{{- $productInfo := include "eric-oss-byos-buildmanager.product-info" . | fromYaml -}}
{{- $config := include "eric-oss-byos-buildmanager.config-annotations" . | fromYaml -}}
{{- include "eric-oss-byos-buildmanager.mergeAnnotations" (dict "location" .Template.Name "sources" (list $productInfo $config)) }}
{{- end -}}

{{/* Opentelemetry tracer configuration env
*/}}
{{- define "eric-oss-byos-buildmanager.traceEnv" }}
{{- if .Values.env.trace.enabled }}
- name: TRACE_ENABLED
  value: "true"
- name: OTEL_EXPORTER_JAEGER_AGENT_HOST
  value: {{ .Values.env.trace.agent.host | quote }}
- name: OTEL_EXPORTER_JAEGER_AGENT_PORT
  value: {{ .Values.env.trace.agent.port | quote }}
- name: OTEL_TRACES_SAMPLER
  value: {{ .Values.env.trace.sampler.type | quote }}
- name: OTEL_TRACES_SAMPLER_ARG
  value: .Values.env.trace.sampler.arg
- name: OTEL_LOG_LEVEL
  value: {{ .Values.env.trace.logLevel | quote }}
- name: OTEL_RESOURCE_ATTRIBUTES
  value: {{ .Values.env.trace.tags | quote }}
{{- end }}
{{- end }}

{{/*
When factfinder endpoint tls is required we override the global flag
*/}}
{{- define "eric-oss-byos-buildmanager.tls" -}}
{{- $g := fromJson (include "eric-oss-byos-buildmanager.global" .) -}}
{{- $tls := $g.security.tls.enabled }}
{{- $tls -}}
{{- end -}}

{{/* Hardcode health check port
*/}}
{{- define "eric-oss-byos-buildmanager.probePort" -}}
{{- $probePort := 9797 -}}
{{- $probePort -}}
{{- end -}}

{{/*
Create a merged set of nodeSelectors from global and service level.
*/}}
{{ define "eric-oss-byos-buildmanager.nodeSelector" }}
  {{- $g := fromJson (include "eric-oss-byos-buildmanager.global" .) -}}
  {{- if .Values.nodeSelector -}}
    {{- range $key, $localValue := .Values.nodeSelector -}}
      {{- if hasKey $g.nodeSelector $key -}}
          {{- $globalValue := index $g.nodeSelector $key -}}
          {{- if ne $globalValue $localValue -}}
            {{- printf "nodeSelector \"%s\" is specified in both global (%s: %s) and service level (%s: %s) with differing values which is not allowed." $key $key $globalValue $key $localValue | fail -}}
          {{- end -}}
      {{- end -}}
    {{- end -}}
    {{- toYaml (merge $g.nodeSelector .Values.nodeSelector) | trim -}}
  {{- else -}}
    {{- toYaml $g.nodeSelector | trim -}}
  {{- end -}}
{{ end }}

{{/*
Define RoleBinding value, note: returns boolean as string
*/}}
{{- define "eric-oss-byos-buildmanager.roleBinding" -}}
{{- $rolebinding := true -}}
{{- if .Values.global -}}
    {{- if .Values.global.security -}}
        {{- if .Values.global.security.policyBinding -}}
            {{- if hasKey .Values.global.security.policyBinding "create" -}}
                {{- $rolebinding = .Values.global.security.policyBinding.create -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
{{- end -}}
{{- $rolebinding -}}
{{- end -}}

{{/*
Define reference to SecurityPolicy
*/}}
{{- define "eric-oss-byos-buildmanager.securityPolicyReference" -}}
{{- $policyreference := "default-restricted-security-policy" -}}
{{- if .Values.global -}}
    {{- if .Values.global.security -}}
        {{- if .Values.global.security.policyReferenceMap -}}
            {{- if hasKey .Values.global.security.policyReferenceMap "default-restricted-security-policy" -}}
                {{- $policyreference = index .Values "global" "security" "policyReferenceMap" "default-restricted-security-policy" -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
{{- end -}}
{{- $policyreference -}}
{{- end -}}

{{/*  Prometheus annotations
*/}}



