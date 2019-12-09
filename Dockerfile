FROM golang:1.13.4-alpine as builder

RUN apk add --update-cache git curl mercurial build-base

RUN mkdir -p /out

ADD . /go/src/github.com/monzo/kontrast
RUN cd /go/src/github.com/monzo/kontrast && \
      make build-in-docker

FROM scratch

COPY --from=builder /out/kontrastd /bin/kontrastd

VOLUME /data

WORKDIR /web

ADD ./assets /web/assets

CMD ["/bin/kontrastd"]

