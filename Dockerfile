FROM andrexus/baseimage

ARG APP_VERSION=0.6.0
ARG DOWNLOAD_URL=https://github.com/entwico/helm-deployer/releases/download/v$APP_VERSION/linux_amd64_helm-deployer

LABEL maintainer="Andrew Tarasenko andrexus@gmail.com"

WORKDIR /srv

ADD config.default.yaml /srv/config.yaml

RUN apk update && \
    apk add ca-certificates wget && \
    update-ca-certificates && \
    wget -q $DOWNLOAD_URL -O /srv/helm-deployer && \
    chmod +x /srv/helm-deployer && \
    chown -R app:app /srv && \
    ln -s /srv/helm-deployer /usr/bin/helm-deployer && \
    rm -rf /var/cache/apk/*

EXPOSE 8000

USER app

CMD ["helm-deployer"]
