BINARY := txtwordle
SOURCE := main.go go.mod

.PHONY: dist clean all build

build/darwin/$(BINARY): $(SOURCE)
	mkdir -p build/darwin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/darwin/$(BINARY)

build/darwinarm/$(BINARY): $(SOURCE)
	mkdir -p build/darwinarm
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build  -o build/darwinarm/$(BINARY)

build/linux/$(BINARY): $(SOURCE)
	mkdir -p build/linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o build/linux/$(BINARY)

build/linuxarmhf/$(BINARY): $(SOURCE)
	mkdir -p build/linuxarmhf
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build  -o build/linuxarmhf/$(BINARY)

build/linuxarm64/$(BINARY): $(SOURCE)
	mkdir -p build/linuxarm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build  -o build/linuxarm64/$(BINARY)

build/darwinuniversal/$(BINARY): build/darwin/$(BINARY) build/darwinarm/$(BINARY)
	mkdir -p build/darwinuniversal
	lipo -create -output build/darwinuniversal/$(BINARY) build/darwin/$(BINARY) build/darwinarm/$(BINARY)

build/shar.tar.gz: build/linux/$(BINARY) build/linuxarmhf/$(BINARY) build/linuxarm64/$(BINARY) shar/README-shar shar/install.sh
	tar cfz build/shar.tar.gz -C build linux/$(BINARY) linuxarmhf/$(BINARY) linuxarm64/$(BINARY) -C ../shar README-shar install.sh

build: build/darwin/$(BINARY) build/darwinarm/$(BINARY) build/linux/$(BINARY) build/linuxarmhf/$(BINARY) build/linuxarm64/$(BINARY) build/darwinuniversal/$(BINARY) ## Build all binaries

dist/$(BINARY)-install.sh: build/shar.tar.gz shar/sh-header
	mkdir -p dist
	cat shar/sh-header build/shar.tar.gz > dist/$(BINARY)-install.sh
	chmod 755 dist/$(BINARY)-install.sh

all: dist/$(BINARY)-install.sh build/darwinuniversal/$(BINARY) ## Make everything

clean: ## Clean everything
	rm -rf build || true
	rm -rf dist || true

help: ## Show this help
	@echo "These are the make commands for the pwned CLI.\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
