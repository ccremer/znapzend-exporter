FROM docker.io/library/alpine:3.11 as runtime

ENTRYPOINT ["znapzend-exporter"]

RUN \
    apk add --no-cache curl bash

COPY znapzend-exporter /usr/bin/
USER 1000:0
