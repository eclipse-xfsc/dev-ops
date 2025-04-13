#Custom Kaniko (Works just under windows)
#FROM gcr.io/kaniko-project/executor:debug AS kaniko
#FROM alpine:3.14.2
#COPY --from=kaniko /kaniko/ /kaniko/

#ENV PATH $PATH:/usr/local/bin:/kaniko
#ENV DOCKER_CONFIG /kaniko/.docker/
#ENV DOCKER_CREDENTIAL_GCR_CONFIG /kaniko/.config/gcloud/docker_credential_gcr_config.json
#ENV SSL_CERT_DIR /kaniko/ssl/certs



FROM python:3.11-alpine
RUN apk update
RUN apk upgrade
RUN apk add bash
RUN apk add sqlite
RUN apk add npm
RUN apk add nodejs
RUN apk add go
RUN apk add git
RUN apk add curl
RUN apk add openssl
RUN apk add maven
RUN apk add make
RUN apk add build-base
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 && \
    chmod +x get_helm.sh && \
    ./get_helm.sh

