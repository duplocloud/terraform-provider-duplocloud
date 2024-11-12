TEST?=$$(go list ./duplocloud/... ./duplosdk/...)
#
HOSTNAME=registry.terraform.io
NAMESPACE=duplocloud
##
#HOSTNAME=duplocloud.net
#NAMESPACE=tf

NAME=duplocloud
BINARY=terraform-provider-${NAME}
VERSION=0.10.54
#mac
#OS_ARCH=darwin_amd64
#OS_ARCH=linux_amd64
OS_ARCH := $(shell go env GOOS)_$(shell go env GOARCH)
OS_ARCH_MAC=darwin_amd64
OS_ARCH_DOCKER=linux_amd64
duplo_host?=http://localhost:60020
duplo_token?=FAKE

default: install

doc:
	go generate

build:
	go build -o ${BINARY}

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
#	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
#	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
#	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
#	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
#	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
#	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mkdir -p ~/.terraform.d/plugin-cache/${OS_ARCH}
	cp ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/${BINARY} ~/.terraform.d/plugin-cache/${OS_ARCH}/terraform-provider-${NAME}_v${VERSION}_x4

install_mac: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH_MAC}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH_MAC}
	mkdir -p ~/.terraform.d/plugin-cache/${OS_ARCH_MAC}
	cp ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH_MAC}/${BINARY} ~/.terraform.d/plugin-cache/${OS_ARCH_MAC}/terraform-provider-${NAME}_v${VERSION}_x4

install_docker: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH_DOCKER}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH_DOCKER}

install_window: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

#	mkdir -p bin
#	mv ${BINARY} bin
test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=90s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m
