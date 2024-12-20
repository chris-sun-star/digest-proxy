FROM golang:1.23 AS builder
ARG GOPROXY
WORKDIR /workspace
COPY . .
RUN go build .

FROM ubuntu:22.04
WORKDIR /
COPY --from=builder /workspace/digest-proxy .
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN apt-get update && apt-get install -y ca-certificates
USER 65534:65534

ENTRYPOINT ["/digest-proxy"]
