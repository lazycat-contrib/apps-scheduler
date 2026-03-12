#!/bin/bash
set -e

echo "=== Meow App Operator Build Script ==="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Get version from manifest.yml
VERSION=$(grep '^version:' manifest.yml | awk '{print $2}')
if [ -z "$VERSION" ]; then
    VERSION="dev"
fi

# Get git info
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo -e "${YELLOW}[1/4] Preparing Go modules...${NC}"
go mod tidy && go mod download

echo -e "${YELLOW}[2/4] Generating Ent code...${NC}"
go generate ./internal/ent

echo -e "${YELLOW}[3/4] Building binary...${NC}"
echo "  Version:    ${VERSION}"
echo "  Git Commit: ${GIT_COMMIT}"
echo "  Build Time: ${BUILD_TIME}"

mkdir -p dist

# Build with version info via ldflags
LDFLAGS="-s -w \
-X apps-scheduler/internal/version.Version=${VERSION} \
-X apps-scheduler/internal/version.GitCommit=${GIT_COMMIT} \
-X apps-scheduler/internal/version.BuildTime=${BUILD_TIME}"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o dist/apps-scheduler ./cmd/apps-scheduler

echo -e "${YELLOW}[4/4] Setting permissions...${NC}"
chmod +x dist/apps-scheduler

echo -e "${GREEN}=== Build completed! ===${NC}"
echo "Binary: dist/apps-scheduler"
ls -lh dist/apps-scheduler