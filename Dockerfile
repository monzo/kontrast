FROM busybox:latest

COPY bin/kryp /bin/kryp

VOLUME /data

CMD ["/bin/kryp", "/data"]
