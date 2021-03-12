# -- build dependencies with alpine --
FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    shorturl_host=0.0.0.0

WORKDIR /build

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go build -ldflags "-s -w -X main.built=$(date -u '+%Y-%m-%dT%H:%M:%SZ')" -o shorturl .

# run application with a small image
FROM scratch

COPY --from=builder /build/shorturl /bin/

EXPOSE 16001

ENTRYPOINT ["shorturl"]
