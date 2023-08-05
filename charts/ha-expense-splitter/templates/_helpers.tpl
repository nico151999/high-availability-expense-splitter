{{/*
Function naming convention (though not always possible to apply):
(what is it used for in helm)-(what does the function return)[-(K8s type if previous parameters cannot ensure uniqueness)]
*/}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name" -}}
{{ . }}-svc
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name-hpa" -}}
{{ include "service-name" . }}-hpa
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name-deployment" -}}
{{ include "service-name" . }}-dpl
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name-pod" -}}
{{ include "service-name" . }}-pod
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name-port" -}}
{{- $svcName := printf "%s-prt" (include "service-name" .) -}}
{{- if gt (len $svcName) 15 -}}
{{- $sub := (sub (len $svcName) 15) | int -}}
{{ substr $sub (len $svcName) $svcName }}
{{- else -}}
{{ $svcName }}
{{- end -}}
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name-service" -}}
{{ include "service-name" . }}-svc
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name-serviceaccount" -}}
{{ include "service-name" . }}-svcacc
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name-ingress" -}}
{{ include "service-name" . }}-ing
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name-networkPolicy" -}}
{{ include "service-name" . }}-npl
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name-configMap" -}}
{{ include "service-name" . }}-cfg
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-name-secret" -}}
{{ include "service-name" . }}-sec
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-serverHostnameKeyName" -}}
{{ . | upper }}_{{ include "service-serverHostnameKey" . }}
{{- end}}

{{- define "service-serverHostnameKey" -}}
SERVER_HOSTNAME
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-serverPortKeyName" -}}
{{ . | upper }}_{{ include "service-serverPortKey" . }}
{{- end}}

{{- define "service-serverPortKey" -}}
SERVER_PORT
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-corsConfigFileName" -}}
{{ . }}CorsConfig.yaml
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-corsConfigVolumeName" -}}
{{ . }}-crs-vol
{{- end}}

{{- define "service-corsConfigVolumeMountPath" -}}
/etc/haExpenseSplitter
{{- end}}

{{/* Accepts the short name of the service as parameter */}}
{{- define "service-corsConfigFilePath" -}}
{{ include "service-corsConfigVolumeMountPath" . }}/{{ include "service-corsConfigFileName" . }}
{{- end}}

{{/* Accepts the short name of the service as parameter "shortName" and the release name as "releaseName" */}}
{{- define "service-selectorLabels-deployment" -}}
app.kubernetes.io/name: {{ include "service-name-deployment" .shortName }}
app.kubernetes.io/instance: {{ .releaseName }}
{{- end}}

{{/* Accepts the short name of the service as parameter "shortName" and the release name as "releaseName" */}}
{{- define "service-selectorlabels-pod" -}}
app.kubernetes.io/name: {{ include "service-name-pod" .shortName }}
app.kubernetes.io/instance: {{ .releaseName }}
{{- end}}

{{- define "service-serverPort" -}}
8080
{{- end}}

{{- define "reflectionService-shortName" -}}
reflection
{{- end}}





{{/* Accepts the short name of the processor as parameter */}}
{{- define "processor-name" -}}
{{ . }}-proc
{{- end}}

{{/* Accepts the short name of the processor as parameter */}}
{{- define "processor-name-hpa" -}}
{{ include "processor-name" . }}-hpa
{{- end}}

{{/* Accepts the short name of the processor as parameter */}}
{{- define "processor-name-deployment" -}}
{{ include "processor-name" . }}-dpl
{{- end}}

{{/* Accepts the short name of the processor as parameter */}}
{{- define "processor-name-pod" -}}
{{ include "processor-name" . }}-pod
{{- end}}

{{/* Accepts the short name of the processor as parameter "shortName" and the release name as "releaseName" */}}
{{- define "processor-selectorLabels-deployment" -}}
app.kubernetes.io/name: {{ include "processor-name-deployment" .shortName }}
app.kubernetes.io/instance: {{ .releaseName }}
{{- end}}

{{/* Accepts the short name of the processor as parameter "shortName" and the release name as "releaseName" */}}
{{- define "processor-selectorlabels-pod" -}}
app.kubernetes.io/name: {{ include "processor-name-pod" .shortName }}
app.kubernetes.io/instance: {{ .releaseName }}
{{- end}}

{{/* Accepts the short name of the processor as parameter */}}
{{- define "processor-name-serviceaccount" -}}
{{ include "processor-name" . }}-svcacc
{{- end}}

{{/* Accepts the short name of the processor as parameter */}}
{{- define "processor-name-secret" -}}
{{ include "processor-name" . }}-sec
{{- end}}





{{/* Accepts the short name of the frontend as parameter */}}
{{- define "frontend-name" -}}
{{ . }}-fe
{{- end}}

{{/* Accepts the short name of the frontend as parameter */}}
{{- define "frontend-name-deployment" -}}
{{ include "frontend-name" . }}-dpl
{{- end}}

{{/* Accepts the short name of the frontend as parameter */}}
{{- define "frontend-name-pod" -}}
{{ include "frontend-name" . }}-pod
{{- end}}

{{/* Accepts the short name of the frontend as parameter */}}
{{- define "frontend-name-port" -}}
{{- $svcName := printf "%s-prt" (include "frontend-name" .) -}}
{{- if gt (len $svcName) 15 -}}
{{- $sub := (sub (len $svcName) 15) | int -}}
{{ substr $sub (len $svcName) $svcName }}
{{- else -}}
{{ $svcName }}
{{- end -}}
{{- end}}

{{/* Accepts the short name of the frontend as parameter */}}
{{- define "frontend-name-service" -}}
{{ include "frontend-name" . }}-svc
{{- end}}

{{/* Accepts the short name of the frontend as parameter */}}
{{- define "frontend-name-ingress" -}}
{{ include "frontend-name" . }}-ing
{{- end}}

{{/* Accepts the short name of the frontend as parameter "shortName" and the release name as "releaseName" */}}
{{- define "frontend-selectorLabels-deployment" -}}
app.kubernetes.io/name: {{ include "frontend-name-deployment" .shortName }}
app.kubernetes.io/instance: {{ .releaseName }}
{{- end}}

{{/* Accepts the short name of the frontend as parameter "shortName" and the release name as "releaseName" */}}
{{- define "frontend-selectorLabels-pod" -}}
app.kubernetes.io/name: {{ include "frontend-name-pod" .shortName }}
app.kubernetes.io/instance: {{ .releaseName }}
{{- end}}

{{- define "frontend-port" -}}
8080
{{- end}}

{{- define "expenseSplitterFrontend-shortName" -}}
expense-splitter
{{- end}}





{{/* An UPPER_SNAKE_CASE reason for an error occurred while trying to publish a task on the MQ */}}
{{- define "global-messagePublicationErrorReason" -}}
TASK_PUBLICATION_ERROR
{{- end}}

{{/* An UPPER_SNAKE_CASE reason for an error occurred while performing a DB select */}}
{{- define "global-dbSelectErrorReason" -}}
DB_SELECT_ERROR
{{- end}}

{{/* An UPPER_SNAKE_CASE reason for an error occurred while performing a DB insert */}}
{{- define "global-dbInsertErrorReason" -}}
DB_INSERT_ERROR
{{- end}}

{{- define "global-globalDomainKey" -}}
GLOBAL_DOMAIN
{{- end}}

{{- define "global-messagePublicationErrorReasonKey" -}}
MESSAGE_PUBLICATION_ERROR_REASON
{{- end}}

{{- define "global-dbSelectErrorReasonKey" -}}
DB_SELECT_ERROR_REASON
{{- end}}

{{- define "global-dbInsertErrorReasonKey" -}}
DB_INSERT_ERROR_REASON
{{- end}}

{{- define "global-natsServerHostKey" -}}
NATS_SERVER_HOST
{{- end}}

{{- define "global-natsServerPortKey" -}}
NATS_SERVER_PORT
{{- end}}

{{- define "global-traceCollectorHostKey" -}}
TRACE_COLLECTOR_HOST
{{- end}}

{{- define "global-traceCollectorPortKey" -}}
TRACE_COLLECTOR_PORT
{{- end}}

{{- define "global-name-configMap" -}}
global-cfg
{{- end}}

{{- define "global-service" -}}
service
{{- end}}

{{- define "global-service-role" -}}
service-role
{{- end}}

{{- define "global-service-rolebinding" -}}
service-rolebinding
{{- end}}

{{- define "global-processor" -}}
processor
{{- end}}

{{- define "global-processor-role" -}}
processor-role
{{- end}}

{{- define "global-processor-rolebinding" -}}
processor-rolebinding
{{- end}}





{{- define "db-name" -}}
{{ .Release.Name }}-db
{{- end}}





{{/* Accepts the name of the table as parameter */}}
{{- define "table-name" -}}
{{ . }}-table
{{- end}}





{{/* Accepts the primary and the fallback image pull secrets as parameters "primary" and "fallback" respectively */}}
{{- define "imagepullsecrets" -}}
{{- $secrets := default .fallback .primary -}}
{{- if not (empty $secrets) -}}
imagePullSecrets:
{{ toYaml $secrets }}
{{- end}}
{{- end}}

{{/* Accepts the primary and the fallback security context as parameters "primary" and "fallback" respectively */}}
{{- define "securitycontext" -}}
{{- $secContext := default .fallback .primary -}}
{{- if not (empty $secContext) -}}
securityContext:
{{- range $k, $v := $secContext}}
  {{$k}}: {{$v}}
{{- end}}
{{- end}}
{{- end}}

{{/*Accepts the root as parameter */}}
{{- define "busyboxImage" -}}
{{ .Values.global.busybox.image.repository }}
{{- if not (empty .Values.global.busybox.image.tag) -}}
:{{ .Values.global.busybox.image.tag }}
{{- end }}
{{- end}}

{{- define "global-dbNameKey" -}}
DB_NAME
{{- end}}

{{- define "dbUserKey" -}}
DB_USER
{{- end}}

{{- define "dbPasswordKey" -}}
DB_PASSWORD
{{- end}}

{{- define "global-dbHostKey" -}}
DB_HOST
{{- end}}

{{- define "global-dbPortKey" -}}
DB_PORT
{{- end}}