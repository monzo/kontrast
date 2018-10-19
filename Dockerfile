FROM golang:1.10-alpine3.7 as builder

RUN apk add --update-cache git curl mercurial build-base

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && chmod +x /usr/local/bin/dep

RUN mkdir -p /out

ADD . /go/src/github.com/monzo/kryp
RUN cd /go/src/github.com/monzo/kryp && \
      make build-in-docker

RUN mkdir -p /go/src/github.com/tomwilkie && \
      cd /go/src/github.com/tomwilkie && \
      git clone https://github.com/tomwilkie/prom-run.git && \
      cd prom-run && \
      make prom-run && \
      mv /go/src/github.com/tomwilkie/prom-run/prom-run /out

FROM scratch

COPY --from=builder /out/prom-run /bin/prom-run
COPY --from=builder /out/kryp /bin/kryp

VOLUME /data

