.PHONY: keydropgo all test xgo clean help
.PHONY: keydropgo-android keydropgo-ios

help: ##@other Show this help
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

ifndef GOPATH
	$(error GOPATH not set. Please set GOPATH and make sure keydrop-go is located at $$GOPATH/src/github.com/keydrop-im/keydrop-go. \
	For more information about the GOPATH environment variable, see https://golang.org/doc/code.html#GOPATH)
endif

VERSION := 0.0.1

CGO_CFLAGS=-I/$(JAVA_HOME)/include -I/$(JAVA_HOME)/include/darwin
GOBIN=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))build/bin
GIT_COMMIT := $(shell git rev-parse --short HEAD)

BUILD_FLAGS := $(shell echo "-ldflags '-X main.buildStamp=`date -u '+%Y-%m-%d.%H:%M:%S'` -X main.gitCommit=$(GIT_COMMIT)  -X github.com/keydrop-im/keydrop-go/geth/params.VersionMeta=$(GIT_COMMIT)'")

GO ?= latest
XGOVERSION ?= 1.10.x
XGOIMAGE = statusteam/xgo:$(XGOVERSION)
XGOIMAGEIOSSIM = statusteam/xgo-ios-simulator:$(XGOVERSION)

networkid ?= keydropChain
gotest_extraflags =

DOCKER_IMAGE_NAME ?= keydropteam/keydrop-go

DOCKER_TEST_WORKDIR = /go/src/github.com/keydrop-im/keydrop-go/
DOCKER_TEST_IMAGE = golang:1.10

UNIT_TEST_PACKAGES := $(shell go list ./...  | grep -v /vendor | grep -v /t/e2e | grep -v /t/destructive | grep -v /lib)

# This is a code for automatic help generator.
# It supports ANSI colors and categories.
# To add new item into help output, simply add comments
# starting with '##'. To add category, use @category.
GREEN  := $(shell echo "\e[32m")
WHITE  := $(shell echo "\e[37m")
YELLOW := $(shell echo "\e[33m")
RESET  := $(shell echo "\e[0m")
HELP_FUN = \
		   %help; \
		   while(<>) { push @{$$help{$$2 // 'options'}}, [$$1, $$3] if /^([a-zA-Z0-9\-]+)\s*:.*\#\#(?:@([a-zA-Z\-]+))?\s(.*)$$/ }; \
		   print "Usage: make [target]\n\n"; \
		   for (sort keys %help) { \
			   print "${WHITE}$$_:${RESET}\n"; \
			   for (@{$$help{$$_}}) { \
				   $$sep = " " x (32 - length $$_->[0]); \
				   print "  ${YELLOW}$$_->[0]${RESET}$$sep${GREEN}$$_->[1]${RESET}\n"; \
			   }; \
			   print "\n"; \
		   }

keydropgo: ##@build Build keydrop-go as keydropd server
	go build -i -o $(GOBIN)/keydropd -v -tags '$(BUILD_TAGS)' $(BUILD_FLAGS) ./cmd/keydropd
	@echo "Compilation done."
	@echo "Run \"build/bin/keydropd -h\" to view available commands."

keydropgo-cross: keydropgo-android keydropgo-ios
	@echo "Full cross compilation done."
	@ls -ld $(GOBIN)/keydropgo-*

keydropgo-android: ##@cross-compile Build keydrop-go for Android
	$(GOPATH)/bin/xgo --go=$(GO) -out keydropgo --dest=$(GOBIN) --targets=android-16/aar -v -tags '$(BUILD_TAGS)' $(BUILD_FLAGS) ./lib
	@mv build/bin/keydropgo-android-16.aar build/bin/keydropgo-$(VERSION).aar
	@echo "Android cross compilation done."

keydropgo-ios: xgo	##@cross-compile Build keydrop-go for iOS
	$(GOPATH)/bin/xgo --go=$(GO) -out keydropgo --dest=$(GOBIN) --targets=ios-9.3/framework -v -tags '$(BUILD_TAGS)' $(BUILD_FLAGS) ./lib
	@echo "iOS framework cross compilation done."

keydropgo-ios-simulator:	##@cross-compile Build keydrop-go for iOS Simulator
	$(GOPATH)/bin/xgo --image $(XGOIMAGEIOSSIM) --go=$(GO) -out keydropgo --dest=$(GOBIN) --targets=ios-9.3/framework -v -tags '$(BUILD_TAGS)' $(BUILD_FLAGS) ./lib
	@echo "iOS framework cross compilation done."

keydropgo-library: ##@cross-compile Build keydrop-go as static library for current platform
	@echo "Building static library..."
	go build -buildmode=c-archive -o $(GOBIN)/libkeydrop.a ./lib
	@echo "Static library built:"
	@ls -la $(GOBIN)/libkeydrop.*

docker-image: BUILD_TAGS ?= metrics prometheus
docker-image: ##@docker Build docker image (use DOCKER_IMAGE_NAME to set the image name)
	@echo "Building docker image..."
	docker build --file _assets/build/Dockerfile --build-arg "build_tags=$(BUILD_TAGS)" . -t $(DOCKER_IMAGE_NAME):latest

docker-image-tag: ##@docker Tag DOCKER_IMAGE_NAME:latest with a tag following pattern $GIT_SHA[:8]-$BUILD_TAGS
	@echo "Tagging docker image..."
	docker tag $(DOCKER_IMAGE_NAME):latest $(DOCKER_IMAGE_NAME):$(shell BUILD_TAGS="$(BUILD_TAGS)" ./_assets/ci/get-docker-image-tag.sh)

xgo-docker-images: ##@docker Build xgo docker images
	@echo "Building xgo docker images..."
	docker build _assets/build/xgo/base -t $(XGOIMAGE)
	docker build _assets/build/xgo/ios-simulator -t $(XGOIMAGEIOSSIM)

xgo-docker-simulator-image:
	@echo "Building xgo simulator docker images..."
	docker build assets/ios-simulator -t $(XGOIMAGEIOSSIM)

xgo:
	docker pull $(XGOIMAGE)
	go get github.com/karalabe/xgo

setup: dep-install lint-install mock-install ##@other Prepare project for first build

generate: ##@other Regenerate assets and other auto-generated stuff
	go generate ./static

mock-install: ##@other Install mocking tools
	go get -u github.com/golang/mock/mockgen

mock: ##@other Regenerate mocks
	mockgen -package=fcm          -destination=geth/notifications/push/fcm/client_mock.go -source=geth/notifications/push/fcm/client.go
	mockgen -package=fake         -destination=geth/transactions/fake/mock.go             -source=geth/transactions/fake/txservice.go
	mockgen -package=account      -destination=geth/account/accounts_mock.go              -source=geth/account/accounts.go
	mockgen -package=jail         -destination=geth/jail/cell_mock.go                     -source=geth/jail/cell.go
	mockgen -package=keydrop       -destination=services/keydrop/account_mock.go            -source=services/keydrop/service.go

docker-test: ##@tests Run tests in a docker container with golang.
	docker run --privileged --rm -it -v "$(shell pwd):$(DOCKER_TEST_WORKDIR)" -w "$(DOCKER_TEST_WORKDIR)" $(DOCKER_TEST_IMAGE) go test ${ARGS}

test: test-unit-coverage ##@tests Run basic, short tests during development

test-unit: ##@tests Run unit and integration tests
	go test -v $(UNIT_TEST_PACKAGES) $(gotest_extraflags)

test-unit-coverage: ##@tests Run unit and integration tests with coverage
	go test -coverpkg= $(UNIT_TEST_PACKAGES) $(gotest_extraflags)

test-unit-race: gotest_extraflags=-race
test-unit-race: test-unit ##@tests Run unit and integration tests with -race flag

test-e2e: ##@tests Run e2e tests
	# order: reliability then alphabetical
	# TODO(tiabc): make a single command out of them adding `-p 1` flag.
	go test -timeout 5m ./t/e2e/accounts/... -network=$(networkid) $(gotest_extraflags)
	go test -timeout 5m ./t/e2e/api/... -network=$(networkid) $(gotest_extraflags)
	go test -timeout 5m ./t/e2e/node/... -network=$(networkid) $(gotest_extraflags)
	go test -timeout 50m ./t/e2e/jail/... -network=$(networkid) $(gotest_extraflags)
	go test -timeout 20m ./t/e2e/rpc/... -network=$(networkid) $(gotest_extraflags)
	go test -timeout 20m ./t/e2e/whisper/... -network=$(networkid) $(gotest_extraflags)
	go test -timeout 10m ./t/e2e/transactions/... -network=$(networkid) $(gotest_extraflags)
	go test -timeout 10m ./t/e2e/services/... -network=$(networkid) $(gotest_extraflags)
	# e2e_test tag is required to include some files from ./lib without _test suffix
	go test -timeout 40m -tags e2e_test ./lib -network=$(networkid) $(gotest_extraflags)

test-e2e-race: gotest_extraflags=-race
test-e2e-race: test-e2e ##@tests Run e2e tests with -race flag

lint-install:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

lint:
	@echo "lint"
	@gometalinter ./...

ci: lint mock dep-ensure test-unit test-e2e ##@tests Run all linters and tests at once

clean: ##@other Cleanup
	rm -fr build/bin/*
	rm -f coverage.out coverage-all.out coverage.html

deep-clean: clean
	rm -Rdf .ethereumtest/keydropChain

vendor-check: ##@dependencies Require all new patches and disallow other changes
	./_assets/patches/patcher -c
	./_assets/ci/isolate-vendor-check.sh

dep-ensure: ##@dependencies Dep ensure and apply all patches
	@dep ensure
	./_assets/patches/patcher

dep-install: ##@dependencies Install vendoring tool
	go get -u github.com/golang/dep/cmd/dep

update-geth: ##@dependencies Update geth (use GETH_BRANCH to optionally set the geth branch name)
	./_assets/ci/update-geth.sh $(GETH_BRANCH)

patch: ##@patching Revert and apply all patches
	./_assets/patches/patcher

patch-revert: ##@patching Revert all patches only
	./_assets/patches/patcher -r
