FROM golang:1.16 AS builder
WORKDIR /usr/local/src/journal3
COPY . .
RUN make

FROM registry.access.redhat.com/ubi8/ubi
COPY --from=builder /usr/local/src/journal3/.output/journal3 /usr/local/bin/journal3
RUN yum update -y && \
    yum install -y sqlite sqlite-devel && \
    yum clean all && \
    rm -rf /var/cache/yum
ENTRYPOINT "/usr/local/bin/journal3"
EXPOSE 8080
