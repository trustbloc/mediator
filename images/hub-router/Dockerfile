#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ARG GO_VER
ARG ALPINE_VER

FROM golang:${GO_VER}-alpine${ALPINE_VER} as base
RUN apk add --no-cache git make
ADD . src/github.com/trustbloc/hub-router
WORKDIR src/github.com/trustbloc/hub-router
RUN make hub-router

FROM alpine:${ALPINE_VER}
COPY --from=base /go/src/github.com/trustbloc/hub-router/.build/bin/hub-router /usr/local/bin

# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

ENTRYPOINT ["hub-router"]
