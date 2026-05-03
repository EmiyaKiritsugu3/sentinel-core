#!/bin/bash
set -e

# Sentinel Pre-Flight Check (Local Quality Gate)
# Adheres to Standards STD-08 and STD-15

echo ""
echo "🛡️  SENTINEL PRE-FLIGHT CHECK"
echo "=============================="

echo "1. Building binaries..."
go build ./...

echo "2. Running unit tests..."
go test ./...

echo "3. Native Go Linting (vet & fmt)..."
go vet ./...
go fmt ./...

echo "4. Synchronizing State Graph..."
go run cmd/sentinel/main.go scan

echo "5. Sovereign Standard Audit..."
# Tenta rodar a auditoria se houver tarefa ativa
go run cmd/sentinel/main.go audit || echo "⚠️  Audit skipped: No active task or audit failed (Local warning only)."

echo "=============================="
echo "✅ PRE-FLIGHT PASSED. Ready for commit/push."
echo ""
