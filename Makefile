.PHONY: help

BINARY=tdi
Version=$(shell grep "const version" main.go | tr -d '"' | awk '{print $$NF}')

help:
	@echo "  make clean  - Remove binaries and vim swap files"
	@echo "  make gotool - Run go tool 'fmt' and 'vet'"
	@echo "  make build  - Compile go code and generate binary file"
	@echo "  make release- Format go code and compile to generate binary release"

gotool:
	go fmt ./
	go vet ./

build:
	CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/$(BINARY) && chmod +x bin/$(BINARY)

build-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/$(BINARY) && chmod +x bin/$(BINARY)

docker:
	docker build -t staugur/tdi-go .

release: gotool build-amd64
	cd bin/ && tar zcvf $(BINARY).$(Version)-linux-amd64.tar.gz $(BINARY) && rm $(BINARY)
