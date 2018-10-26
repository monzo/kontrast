FROM golang:1.10-alpine3.7 as builder

RUN apk add --update-cache git curl mercurial build-base

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && chmod +x /usr/local/bin/dep

RUN mkdir -p /out

ADD . /go/src/github.com/monzo/kryp
RUN cd /go/src/github.com/monzo/kryp && \
      make build-in-docker

FROM scratch

COPY --from=builder /out/krypd /bin/krypd

VOLUME /data

