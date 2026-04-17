#!/bin/bash

echo "=== DevOps Platform Build Script ==="

# 后端构建
echo ""
echo "1. Building Backend..."
cd backend

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}')
echo "   Go version: $GO_VERSION"

# 整理依赖
echo "   Running go mod tidy..."
go mod tidy

# 编译
echo "   Compiling..."
if go build -o devops-backend .; then
    echo "   ✅ Backend build success: backend/devops-backend"
else
    echo "   ❌ Backend build failed"
    echo "   Note: If you encounter 'bad g in signal handler', try using Go 1.22+"
    exit 1
fi

cd ..

# 前端构建
echo ""
echo "2. Building Frontend..."
cd frontend

# 检查node
if ! command -v npm &> /dev/null; then
    echo "   ❌ npm not found"
    exit 1
fi

echo "   Node version: $(node -v)"
echo "   npm version: $(npm -v)"

# 安装依赖
echo "   Installing dependencies..."
npm install

# 构建
echo "   Building..."
if npm run build; then
    echo "   ✅ Frontend build success: frontend/dist/"
else
    echo "   ❌ Frontend build failed"
    exit 1
fi

cd ..

echo ""
echo "=== Build Complete ==="
echo ""
echo "Backend binary: backend/devops-backend"
echo "Frontend dist:  frontend/dist/"
echo ""
echo "To start the application:"
echo "  1. Start MySQL and configure backend/config.yaml"
echo "  2. Run: cd backend && ./devops-backend"
echo "  3. Or serve frontend/dist with nginx/caddy"