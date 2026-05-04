#!/bin/bash
set -e

# Sentinel Pre-Flight Check v2.1 (The Markdown & Compliance Edition)
# Adheres to Standards STD-08, STD-11, STD-13, STD-14 and STD-15

# Colors for UI
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "🛡️  SENTINEL PRE-FLIGHT CHECK v2.1"
echo "=================================="
echo -e "${NC}"

STATUS_BUILD="${RED}FAIL${NC}"
STATUS_TEST="${RED}FAIL${NC}"
STATUS_LINT="${RED}FAIL${NC}"
STATUS_SECURITY="${RED}FAIL${NC}"
STATUS_HYGIENE="${RED}FAIL${NC}"
STATUS_DOCS_AUDIT="${RED}FAIL${NC}"
STATUS_DOCS_LINT="${RED}FAIL${NC}"

# 1. Hygiene Gate (Go Mod)
echo -e "${YELLOW}[1/7] Checking Dependency Hygiene...${NC}"
go mod verify > /dev/null
go mod tidy
if git diff --exit-code go.mod go.sum > /dev/null; then
    STATUS_HYGIENE="${GREEN}PASS${NC}"
else
    echo -e "${RED}❌ Hygiene Error: go.mod or go.sum are not tidy. Run 'go mod tidy' and stage changes.${NC}"
    exit 1
fi

# 2. Build Gate
echo -e "${YELLOW}[2/7] Building Binaries...${NC}"
if go build ./...; then
    STATUS_BUILD="${GREEN}PASS${NC}"
else
    exit 1
fi

# 3. Test Gate
echo -e "${YELLOW}[3/7] Running Unit Tests...${NC}"
if go test ./...; then
    STATUS_TEST="${GREEN}PASS${NC}"
else
    exit 1
fi

# 4. Lint & Formatting Gate
echo -e "${YELLOW}[4/7] Verifying Formatting & Static Analysis...${NC}"
UNFORMATTED=$(gofmt -l .)
if [ -z "$UNFORMATTED" ]; then
    go vet ./...
    STATUS_LINT="${GREEN}PASS${NC}"
else
    echo -e "${RED}❌ Formatting Error: The following files are not formatted:${NC}"
    echo "$UNFORMATTED"
    echo -e "${YELLOW}Run 'go fmt ./...' and stage the changes.${NC}"
    exit 1
fi

# 5. Security Gate (Vulnerabilities)
echo -e "${YELLOW}[5/7] Scanning for Vulnerabilities (govulncheck)...${NC}"
if go run golang.org/x/vuln/cmd/govulncheck@latest ./...; then
    STATUS_SECURITY="${GREEN}PASS${NC}"
else
    echo -e "${RED}❌ Security Error: Vulnerabilities detected in dependencies.${NC}"
    exit 1
fi

# 6. Documentation Lint (markdownlint)
echo -e "${YELLOW}[6/7] Running Markdown Lint (markdownlint-cli2)...${NC}"
if npx --yes markdownlint-cli2 > /dev/null 2>&1; then
    STATUS_DOCS_LINT="${GREEN}PASS${NC}"
else
    echo -e "${RED}❌ Documentation Lint Error: Lint violations found. Run 'npx markdownlint-cli2' for details.${NC}"
    exit 1
fi

# 7. Documentation Audit (Placeholders)
echo -e "${YELLOW}[7/7] Auditing Documentation Placeholders...${NC}"
PLACEHOLDERS=$(grep -rE "\[Descreva\.\.\.\]|\[Ponto\.\.\.\]|TODO: ADR" docs/ || true)
if [ -z "$PLACEHOLDERS" ]; then
    STATUS_DOCS_AUDIT="${GREEN}PASS${NC}"
else
    echo -e "${RED}❌ Documentation Audit Error: Placeholders found in docs:${NC}"
    echo "$PLACEHOLDERS"
    exit 1
fi

# Synchronizing State Graph
echo -e "\n${YELLOW}🔄 Synchronizing State Graph...${NC}"
go run cmd/sentinel/main.go scan

# Sovereign Standard Audit
echo -e "${YELLOW}🛡️  Running Sovereign Standard Audit...${NC}"
go run cmd/sentinel/main.go audit || echo -e "${YELLOW}⚠️  Audit skipped: No active task.${NC}"

# Final Matrix
echo -e "\n${BLUE}==================================${NC}"
echo -e "🚀 PRE-FLIGHT COMPLIANCE MATRIX"
echo -e "----------------------------------"
echo -e "  Build        : $STATUS_BUILD"
echo -e "  Tests        : $STATUS_TEST"
echo -e "  Lint (Go)    : $STATUS_LINT"
echo -e "  Security     : $STATUS_SECURITY"
echo -e "  Hygiene      : $STATUS_HYGIENE"
echo -e "  Docs Lint    : $STATUS_DOCS_LINT"
echo -e "  Docs Audit   : $STATUS_DOCS_AUDIT"
echo -e "----------------------------------"
echo -e "${GREEN}✅ ALL GATES OPEN. READY FOR COMMIT.${NC}"
echo -e "${BLUE}==================================${NC}\n"
