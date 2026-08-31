#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")"
GOOS=wasip1 GOARCH=wasm go build -o ../dist/openapi-go-gin-mock-v1.0.0.wasm ../cmd/plugin/main.go
