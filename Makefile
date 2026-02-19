BINARY_NAME=terraform-provider-clerk
HOSTNAME=registry.terraform.io
NAMESPACE=makolabsai
NAME=clerk
VERSION=0.1.0
OS_ARCH=$$(go env GOOS)_$$(go env GOARCH)

default: build

.PHONY: build
build:
	go build -o $(BINARY_NAME)

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
	cp $(BINARY_NAME) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)/

.PHONY: test
test:
	go test ./... -v -count=1

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -count=1 -timeout 120m

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: generate
generate:
	go generate ./...

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)

.PHONY: check
check: fmt lint test
