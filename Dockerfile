FROM golang:1.20.10-bullseye as GO_PROJECT

ARG SSH_PRV_KEY
ARG SSH_PUB_KEY
ARG SSH_KNOWN_HOSTS

RUN mkdir -p /root/.ssh && \
    chmod 0700 /root/.ssh && \
    echo "$SSH_PRV_KEY" > /root/.ssh/id_rsa && \
    echo "$SSH_PUB_KEY" > /root/.ssh/id_rsa.pub && \
    echo "$SSH_KNOWN_HOSTS" > /root/.ssh/known_hosts && \
    chmod 600 /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa.pub && \
    chmod 600 /root/.ssh/known_hosts && \
    mkdir -p /code

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY="https://goproxy.cn,direct"

COPY .  /code
RUN cd /code && \
    go mod tidy

RUN cd /code && \
    go build -buildvcs=false -o sentry-exporter-go .
