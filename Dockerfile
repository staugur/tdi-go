ARG buildos=golang:1.20-alpine
ARG runos=scratch

# -- build dependencies with alpine --
FROM $buildos AS builder
WORKDIR /build
COPY . .
ARG goproxy
ARG TARGETARCH
RUN if [ "x$goproxy" != "x" ]; then go env -w GOPROXY=${goproxy},direct; fi ; \
    CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags "-s -w" .

# -- run application with a small image --
FROM $runos
COPY --from=builder /build/tdi /bin/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 13145
ENTRYPOINT ["tdi"]
CMD ["-d", "/tdi"]
