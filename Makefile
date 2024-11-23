include $(CURDIR)/hack/tools.mk

MAIN_PACKAGE_PATH := ./cmd/tfc
BINARY_NAME       := ./bin/tfc

GO_LINT_ERROR_FORMAT ?= colored-line-number

# ============================================================================
# HELPERS
# ============================================================================

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ============================================================================
# QUALITY CONTROL
# ============================================================================

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## lint-go: lint the code
.PHONY: lint-go
lint-go: install-golangci-lint
	$(GOLANGCI_LINT) run --out-format=$(GO_LINT_ERROR_FORMAT)

## lint-go-fix: lint the code, auto-fix if possible
.PHONY: lint-go-fix
lint-go-fix:
	$(GOLANGCI_LINT) run --fix

# ============================================================================
# DEVELOPMENT
# ============================================================================

## test: run all tests
.PHONY: test
test:
	go test -v -race ./...

## build: build the application
.PHONY: build
build:
	CGO_ENABLED=0 go build -o=${BINARY_NAME} ${MAIN_PACKAGE_PATH}

