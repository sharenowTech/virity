PKGS := $(shell go list ./... | grep -v /vendor)
BINARY := virity
OS ?= linux
ARCH ?= amd64
VERSION ?= latest

.PHONY: dep
dep: ## Get the dependencies
	go get -v -d ./...

.PHONY: test
test: dep
	go test $(PKGS)

BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/golangci-lint

$(GOMETALINTER):
	GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.18.0

.PHONY: lint
lint: $(GOMETALINTER)
	golangci-lint run ./... --fast --disable-all --enable govet

.PHONY: docker-agent
docker-agent:
ifeq ($(REGISTRY),)
	docker build --build-arg VERSION=$(VERSION) -t $(BINARY)-agent:$(VERSION) -f Dockerfile.agent .
else
	docker build --build-arg VERSION=$(VERSION) -t $(REGISTRY)/$(BINARY)-agent:$(VERSION) -f Dockerfile.agent .
	docker push $(REGISTRY)/$(BINARY)-agent:$(VERSION)
endif

.PHONY: docker-server
docker-server:
ifeq ($(REGISTRY),)
	docker build --build-arg VERSION=$(VERSION) -t $(BINARY)-server:$(VERSION) -f Dockerfile.server .
else
	docker build --build-arg VERSION=$(VERSION) -t $(REGISTRY)/$(BINARY)-server:$(VERSION) -f Dockerfile.server .
	docker push $(REGISTRY)/$(BINARY)-server:$(VERSION)
endif

.PHONY: docker
docker: docker-agent docker-server

CMDs := agent server
CMD = $(word 1, $@)

.PHONY: $(CMDs)
$(CMDs): dep
	mkdir -p build
	GOOS=$(OS) GOARCH=$(ARCH) CGO_ENABLED=0 go build -v -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o build/$(BINARY)-$(CMD)-$(OS)-$(ARCH)-v$(VERSION) github.com/car2go/$(BINARY)/cmd/$(CMD)

.PHONY: bin
bin: $(CMDs)

.PHONY: all
all: $(CMDs) docker-agent docker-server

.DEFAULT_GOAL := bin
