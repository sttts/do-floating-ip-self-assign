FROM alpine:3.1
MAINTAINER Dr. Stefan Schimanski <stefan.schimanski@gmail.com>

RUN apk add -U ca-certificates && rm -rf /var/cache/apk/*

COPY do-floating-ip-self-assign /do-floating-ip-self-assign

ENTRYPOINT ["/do-floating-ip-self-assign"]
