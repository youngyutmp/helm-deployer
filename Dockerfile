FROM alpine:3.8

ARG APP_VERSION=0.5.0
ARG DOWNLOAD_URL=https://github.com/entwico/helm-deployer/releases/download/v$APP_VERSION/linux_amd64_helm-deployer

LABEL maintainer="Andrew Tarasenko andrexus@gmail.com"

WORKDIR /app

RUN apk update && \
    apk add ca-certificates wget && \
    update-ca-certificates && \
    wget -q $DOWNLOAD_URL -O /app/helm-deployer && \
    chmod +x /app/helm-deployer && \
    rm -rf /var/cache/apk/*

ADD config.default.yaml /app/config.yaml

EXPOSE 8000

ENTRYPOINT ["/app/helm-deployer"]
