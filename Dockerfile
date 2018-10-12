FROM busybox:latest

COPY bin/prom-run /bin/prom-run
COPY bin/kryp /bin/kryp

VOLUME /data

