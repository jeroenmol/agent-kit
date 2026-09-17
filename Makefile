GOCACHE := $(CURDIR)/.cache/go-build
GOMODCACHE := $(CURDIR)/.cache/go-mod
GOPATH := $(CURDIR)/.cache/go-path
GOENV := GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) GOPATH=$(GOPATH)

.PHONY: build test focused-test fmt check lint

build:
	@mkdir -p $(GOCACHE) $(GOMODCACHE) $(GOPATH) bin
	@version=$$(git describe --tags --always --dirty 2>/dev/null || printf 'devel'); \
	commit=$$(git rev-parse --short HEAD 2>/dev/null || printf 'unknown'); \
	dirty=false; \
	if test -n "$$(git status --porcelain 2>/dev/null)"; then dirty=true; fi; \
	$(GOENV) go build -ldflags "-X main.version=$$version -X main.commit=$$commit -X main.dirty=$$dirty" -o bin/agent-kit ./cmd/agent-kit

test:
	@mkdir -p $(GOCACHE) $(GOMODCACHE) $(GOPATH)
	@$(GOENV) go test ./...

focused-test:
	@mkdir -p $(GOCACHE) $(GOMODCACHE) $(GOPATH)
	@$(GOENV) go test ./cmd/agent-kit

fmt:
	@gofmt -w $$(find . -path './.cache' -prune -o -name '*.go' -type f -print)

check:
	@mkdir -p $(GOCACHE) $(GOMODCACHE) $(GOPATH)
	@test -z "$$(gofmt -l $$(find . -path './.cache' -prune -o -name '*.go' -type f -print))"
	@$(GOENV) go vet ./...
	@$(GOENV) go build -o .cache/agent-kit-check ./cmd/agent-kit
	@$(GOENV) go test ./...

lint:
	@mkdir -p $(GOCACHE) $(GOMODCACHE) $(GOPATH) .cache/golangci-lint
	@$(GOENV) GOLANGCI_LINT_CACHE=$(CURDIR)/.cache/golangci-lint golangci-lint run
