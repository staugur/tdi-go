# -- build dependencies with alpine --
FROM golang:1.16.2-alpine3.13 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go build -ldflags "-s -w" -o tdi .

# run application with a small image
FROM scratch

COPY --from=builder /build/tdi /bin/

EXPOSE 13145

ENTRYPOINT ["tdi", "-d", "/tdi"]
