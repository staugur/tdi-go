.PHONY: help

BINARY=tdi

help:
	@echo "  make clean  - Remove binaries and vim swap files"
	@echo "  make gotool - Run go tool 'fmt' and 'vet'"
	@echo "  make build  - Compile go code and generate binary file"

gotool:
	go fmt ./
	go vet ./

build: gotool
	go build -ldflags "-s -w" -o bin/$(BINARY) && chmod +x bin/$(BINARY)

docker:
	docker build -t staugur/tdi-go .
