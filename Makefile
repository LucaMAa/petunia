# ──────────────────────────────────────────────────────────────────────────
# Makefile - quality/security checks for the Go backend (petunia)
#
# Quick usage:
#   make check         -> runs ALL checks in read-only mode (CI-safe,
#                          does not modify anything, fails if something is wrong)
#   make fmt            -> FORMATS the code and writes changes to disk
#                          (gofmt + goimports). Also run automatically by CI.
#   make fmt-check      -> same as fmt but without writing, only checks
#   make vet            -> go vet
#   make lint           -> golangci-lint (includes staticcheck, gosec,
#                          errcheck, ineffassign, unused, misspell, ...)
#   make vuln           -> govulncheck (known vulnerabilities, official Go database)
#   make secrets        -> gitleaks (hardcoded secrets/credentials in the code)
#   make build          -> go build
#   make test           -> go test (race + coverage)
#   make tidy-check     -> checks that go.mod/go.sum are clean
#   make tools          -> installs all required tools
# ──────────────────────────────────────────────────────────────────────────

.PHONY: check fmt fmt-check vet lint vuln secrets build test tidy-check tools clean

GO       := go
PKG      := ./...
GOBIN    := $(shell go env GOPATH)/bin

check: fmt-check vet lint vuln secrets build test tidy-check
	@echo ""
	@echo "✅ All checks passed."

fmt: tools
	@echo "→ Formatting (gofmt + goimports)..."
	gofmt -w .
	$(GOBIN)/goimports -w .

fmt-check:
	@echo "→ Checking formatting (gofmt)..."
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "❌ Unformatted files:"; \
		echo "$$unformatted"; \
		echo "Run 'make fmt' to fix them."; \
		exit 1; \
	else \
		echo "✅ Formatting OK."; \
	fi

vet:
	@echo "→ go vet..."
	$(GO) vet $(PKG)

lint: tools
	@echo "→ golangci-lint (staticcheck + gosec + errcheck + ...)..."
	$(GOBIN)/golangci-lint run ./...

vuln: tools
	@echo "→ govulncheck..."
	$(GOBIN)/govulncheck $(PKG)

secrets: tools
	@echo "→ gitleaks (searching for hardcoded secrets)..."
	$(GOBIN)/gitleaks detect --source . --no-banner --redact

build:
	@echo "→ go build..."
	$(GO) build ./...

test:
	@echo "→ go test..."
	$(GO) test $(PKG) -race -cover

tidy-check:
	@echo "→ go mod tidy (verification)..."
	@cp go.mod /tmp/go.mod.bak; cp go.sum /tmp/go.sum.bak 2>/dev/null || true
	@$(GO) mod tidy
	@if ! diff -q go.mod /tmp/go.mod.bak >/dev/null 2>&1 || ! diff -q go.sum /tmp/go.sum.bak >/dev/null 2>&1; then \
		echo "❌ go.mod/go.sum were not clean (now fixed). Run 'go mod tidy' and commit the changes."; \
		exit 1; \
	else \
		echo "✅ go.mod/go.sum OK."; \
	fi

tools:
	@command -v $(GOBIN)/goimports >/dev/null 2>&1 || { \
		echo "→ Installing goimports..."; \
		$(GO) install golang.org/x/tools/cmd/goimports@latest; \
	}
	@echo "→ Rebuilding golangci-lint with the active Go toolchain..."
	@rm -f $(GOBIN)/golangci-lint
	@$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@command -v $(GOBIN)/govulncheck >/dev/null 2>&1 || { \
		echo "→ Installing govulncheck..."; \
		$(GO) install golang.org/x/vuln/cmd/govulncheck@latest; \
	}
	@command -v $(GOBIN)/gitleaks >/dev/null 2>&1 || { \
		echo "→ Installing gitleaks..."; \
		$(GO) install github.com/zricethezav/gitleaks/v8@latest; \
	}

clean:
	$(GO) clean
