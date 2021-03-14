.PHONY: help

BINARY=tdi
Version=$(shell grep "const version" main.go | tr -d '"' | awk '{print $NF}')

help:
	@echo "  make clean  - Remove binaries and vim swap files"
	@echo "  make gotool - Run go tool 'fmt' and 'vet'"
	@echo "  make test   - Run go test"
	@echo "  make build  - Compile go code and generate binary file"
	@echo "  make release- Format go code and compile to generate binary release"

gotool:
	go fmt ./
	go vet ./

test:
	@go test -count=1 .

build: gotool
	go build -ldflags "-s -w" -o bin/$(BINARY) && chmod +x bin/$(BINARY)

docker:
	docker build -t staugur/tdi-go .

release: gotool build
	cd bin/ && tar zcvf $(BINARY).$(Version)-linux-amd64.tar.gz $(BINARY) && rm $(BINARY)
