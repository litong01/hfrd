FROM golang:alpine

RUN set -e; \
    apk update; \
    apk add --no-cache --update git texlive;\
    go get github.com/IzakMarais/reporter/...; \
    go install -v github.com/IzakMarais/reporter/cmd/grafana-reporter





