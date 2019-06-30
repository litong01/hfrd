# HFRD API server
#
# To build this image, run the following command:
#   docker build --build-arg user=$USER \
#      --build-arg uid=$(id -u) \
#      --build-arg gid=$(id -g) \
#      --build-arg gitid= --build-arg gittoken= \
#      -t hfrdserver  .
#
FROM golang:alpine AS BUILD

ARG gitid=
ARG gittoken=

USER root

RUN apk add --no-cache --update curl git && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

RUN cd $GOPATH/src && \
    git clone https://${gitid}:${gittoken}@github.ibm.com/IBMCode/hfrd.git

RUN echo "Ready to build " && \
    cd $GOPATH/src/hfrd/api && \
    dep ensure && \
    go build -o /usr/local/bin/hfrdserver

FROM alpine

ARG user=ubuntu
ARG uid=1000
ARG gid=1000

USER root

RUN set -eux; \
    addgroup -g ${gid} ${user} \
    && adduser -D -h /opt/hfrd -s /bin/bash -u ${uid} -G ${user} ${user} \
    && adduser ${user} ping && export GOPATH=/opt/hfrd \
    && mkdir -p /opt/hfrd/var

COPY --from=build /usr/local/bin/hfrdserver /usr/local/bin/hfrdserver
ENV HOME /opt/hfrd
ENV WORKDIR /opt/hfrd
ENV GIN_MODE release
WORKDIR /opt/hfrd
ENTRYPOINT ["hfrdserver"]
USER ${user}
