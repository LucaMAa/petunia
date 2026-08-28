#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

if ! command -v curl >/dev/null 2>&1; then
  echo "Error: curl is not installed."
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "Error: python3 is not installed."
  exit 1
fi

JSON_PAYLOAD="$(curl -fsSL "https://go.dev/dl/?mode=json")"
TARGET_VERSION="$(printf '%s' "$JSON_PAYLOAD" | python3 -c 'import json,sys; data=json.load(sys.stdin); versions=[item["version"] for item in data if item.get("stable") is True]; print(versions[0] if versions else "")')"

if [[ -z "$TARGET_VERSION" ]]; then
  echo "Unable to determine the latest stable Go version."
  exit 1
fi

if [[ ! "$TARGET_VERSION" =~ ^go[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Unexpected response from go.dev: $TARGET_VERSION"
  exit 1
fi

GO_BASE="${TARGET_VERSION#go}"
GO_BASE="${GO_BASE%.*}.0"
TARGET_GO="${TARGET_VERSION#go}"

echo "Updating Go to $TARGET_GO"

go mod edit -go="$GO_BASE" -toolchain=go"$TARGET_GO"

go mod tidy

go version

echo ""
echo "Rebuilding Go tooling for $TARGET_GO"
rm -f "$(go env GOPATH)/bin/golangci-lint"
$(go env GOPATH)/bin/golangci-lint version >/dev/null 2>&1 || true
GOBIN="$(go env GOPATH)/bin" && \
  GOFLAGS='' go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest && \
  go install golang.org/x/tools/cmd/goimports@latest && \
  go install golang.org/x/vuln/cmd/govulncheck@latest

echo ""
echo "OK: the module now targets Go $TARGET_GO and toolchain binaries were rebuilt with the new version."
