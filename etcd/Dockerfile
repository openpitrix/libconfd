# Copyright 2017 <chaishushan{AT}gmail.com>. All rights reserved.
# Use of this source code is governed by a BSD license
# that can be found in the LICENSE file.

# https://coreos.com/etcd/docs/latest/op-guide/container.html#docker
# https://coreos.com/etcd/docs/latest/v2/api.html

FROM golang:1.9.2-alpine3.6 as builder

RUN apk add --no-cache curl openssl

# download etcd-v3.2.12
RUN mkdir -p /etcd-download && cd /etcd-download \
	&& curl -L https://github.com/coreos/etcd/releases/download/v3.2.12/etcd-v3.2.12-linux-amd64.tar.gz | tar zx

FROM alpine:3.6

COPY --from=builder /etcd-download/etcd-v3.2.12-linux-amd64/etcd    /usr/local/bin/etcd
COPY --from=builder /etcd-download/etcd-v3.2.12-linux-amd64/etcdctl /usr/local/bin/etcdctl

EXPOSE 2379 2380

RUN mkdir -p /etcd
WORKDIR /etcd

ENTRYPOINT []

CMD [ "/usr/local/bin/etcd", \
    "--data-dir=/etcd/data", \
    "--name=node1", \
    \
    "--initial-advertise-peer-urls=http://0.0.0.0:2380", \
    "--listen-peer-urls=http://0.0.0.0:2380", \
    \
    "-advertise-client-urls", "http://0.0.0.0:2379", \
    "-listen-client-urls", "http://0.0.0.0:2379", \
    \
    "--initial-cluster", "node1=http://0.0.0.0:2380" \
    ]
