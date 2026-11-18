{{- define "xfsc.dbSecret" -}}{{ default (printf "%s-keycloak-db" .Release.Name) .Values.keycloak.database.passwordSecretName }}{{- end -}}
{{- define "xfsc.adminSecret" -}}{{ default (printf "%s-keycloak-admin" .Release.Name) .Values.keycloak.bootstrapAdmin.passwordSecretName }}{{- end -}}
{{- define "xfsc.clientSecret" -}}{{ default (printf "%s-identity-controller-credentials" .Release.Name) .Values.controller.clientSecret.secretName }}{{- end -}}
{{- define "xfsc.annotations" -}}
inject.xfsc.io/enabled: {{ .enabled | quote }}
inject.xfsc.io/needs: {{ .needs | quote }}
inject.xfsc.io/env-replacements: {{ .replacements | toJson | quote }}
{{- end -}}
