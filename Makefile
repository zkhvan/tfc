include $(CURDIR)/hack/tools.mk

MAIN_PACKAGE_PATH := ./cmd/tfc
BINARY_NAME       ?= ./bin/tfc
VERSION_PACKAGE   := github.com/zkhvan/tfc/internal/build

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

# --------------------------------------------------------------------------
# HELPERS
# --------------------------------------------------------------------------

## tidy: tidy the code
.PHONY: tidy
tidy: tidy-go

## tidy-go: format code and tidy modfile
.PHONY: tidy-go
tidy-go: format-go lint-go-fix
	go mod tidy -v

# --------------------------------------------------------------------------
# LINTERS
# --------------------------------------------------------------------------

## lint: lint the code
.PHONY: lint
lint: lint-go

## lint-go: lint the go code
.PHONY: lint-go
lint-go: install-golangci-lint
	$(GOLANGCI_LINT) run --out-format=$(GO_LINT_ERROR_FORMAT)

## lint-go-fix: lint the go code, auto-fix if possible
.PHONY: lint-go-fix
lint-go-fix:
	$(GOLANGCI_LINT) run --fix

# --------------------------------------------------------------------------
# FORMATTERS
# --------------------------------------------------------------------------

.PHONY: format-go
format-go:
	go fmt ./...

# ============================================================================
# DEVELOPMENT
# ============================================================================

## test: run all tests
.PHONY: test
test:
	go test \
		-v \
		-timeout=300s \
		-coverprofile=coverage.txt \
		-covermode=atomic \
		-race \
		./...

## build: build the application
.PHONY: build
build:
	CGO_ENABLED=0 go build \
		-ldflags "-w -X $(VERSION_PACKAGE).Version=$(VERSION) -X $(VERSION_PACKAGE).Date=$$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
		-o=${BINARY_NAME} \
		${MAIN_PACKAGE_PATH}

