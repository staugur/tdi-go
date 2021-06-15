ARG buildos=golang:1.16.5-alpine3.13
ARG runos=scratch

# build dependencies with alpine
FROM $buildos AS builder

WORKDIR /build

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o tdi

# run application with a small image
FROM $runos

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /build/tdi /bin/

EXPOSE 13145

ENTRYPOINT ["tdi", "-d", "/tdi"]
