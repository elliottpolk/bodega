BIN=bodega
PKG=github.com/elliottpolk/bodega
CLI_VERSION=`cat .version`
COMPILED=`date +%s`
GIT_HASH=`git rev-parse --short HEAD`
GOOS?=linux
BUILD_DIR=./build/bin

M = $(shell printf "\033[34;1m◉\033[0m")

default: clean build ;                                              @ ## defaulting to clean and build

.PHONY: all
all: clean build

.PHONY: clean
clean: ; $(info $(M) running clean ...)                             @ ## clean up the old build dir
	@rm -vrf build

.PHONY: test
test: unit-test;													@ ## wrapper to run all testing

.PHONY: unit-test
unit-test: ; $(info $(M) running unit tests...)                     @ ## run the unit tests
	@go get -v -u
	@go test -cover ./...

.PHONY: build
build: build-dir; $(info $(M) building ...)                         @ ## build the binary
	@GOOS=$(GOOS) go build \
		-ldflags "-X main.version=$(CLI_VERSION) -X main.compiled=$(COMPILED) -X main.githash=$(GIT_HASH)" \
		-o ./build/bin/$(BIN) ./cmd/main.go

.PHONEY: build-dir
build-dir: ;
	@[ ! -d "${BUILD_DIR}" ] && mkdir -vp "${BUILD_DIR}/public" || true

.PHONEY: install
install: ; $(info $(M) installing locally ...) 						@ ## install binary locally
	@GOOS=$(GOOS) go build \
		-ldflags "-X main.version=$(CLI_VERSION) -X main.compiled=$(COMPILED) -X main.githash=$(GIT_HASH)" \
		-o $(GOPATH)/bin/$(BIN) ./cmd/main.go

.PHONY: proto
proto: ; $(info $(M) running protoc commands...)                    @ ## code generation from .proto files
	@for i in `find proto -type f | awk -F '/' '{print $2}' | sort -u`; \
		do \
			rm -rf ${i}/*.pb*.go 2> /dev/null || true; \
		done
	@for i in `find proto -type f`;   \
		do                  \
			protoc			\
			-Iproto			\
			-I$(GOPATH)/src \
			-I$(GOPATH)/src/$(PKG)/proto \
			-I$(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
			--go_out=plugins=grpc,paths=source_relative:. \
			--grpc-gateway_out=logtostderr=true,paths=source_relative,allow_delete_body=true:. \
			"$${i}"; 	\
		done
	@sed -i 's/json:"id,omitempty"/json:"id,omitempty" bson:"_id"/g' **/*.pb.go

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

