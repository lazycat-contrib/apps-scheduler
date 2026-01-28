#!/bin/bash
set -e

echo "=== Meow App Operator Build Script ==="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}[1/4] Preparing Go modules...${NC}"
go mod tidy && go mod download

echo -e "${YELLOW}[2/4] Generating Ent code...${NC}"
go generate ./internal/ent

echo -e "${YELLOW}[3/4] Building binary...${NC}"
mkdir -p dist
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/apps-scheduler ./cmd/apps-scheduler

echo -e "${YELLOW}[4/4] Setting permissions...${NC}"
chmod +x dist/apps-scheduler

echo -e "${GREEN}=== Build completed! ===${NC}"
echo "Binary: dist/apps-scheduler"
ls -lh dist/apps-scheduler
