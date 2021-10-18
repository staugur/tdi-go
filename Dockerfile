ARG buildos=golang:1.17.2-alpine
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
EXPOSE 13145
ENTRYPOINT ["tdi", "-d", "/tdi"]