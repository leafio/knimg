#!/bin/bash

# 跨平台编译脚本

echo "=== KnImg 跨平台编译脚本 ==="
echo "开始构建三个平台的可执行文件..."

# 创建输出目录
mkdir -p build

# 设置Go环境变量
export GO111MODULE=on
export GOROOT=/usr/local/go

# 清理之前的构建
rm -f build/*

# 编译Windows 64位
echo "\n编译 Windows 64位..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/knimg-windows-amd64.exe .

# 编译Mac 64位
echo "\n编译 Mac 64位..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/knimg-darwin-amd64 .

# 编译Mac ARM64位 (M1/M2)
echo "\n编译 Mac ARM64位..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o build/knimg-darwin-arm64 .

# 编译Linux 64位
echo "\n编译 Linux 64位..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/knimg-linux-amd64 .

# 检查编译结果
echo "\n=== 编译结果 ==="
ls -la build/

# 检查是否成功
if [ -f "build/knimg-windows-amd64.exe" ] && [ -f "build/knimg-darwin-amd64" ] && [ -f "build/knimg-darwin-arm64" ] && [ -f "build/knimg-linux-amd64" ]; then
    echo "\n✓ 所有平台编译成功!"
    echo "可执行文件位于 build/ 目录"
else
    echo "\n✗ 编译失败!"
    exit 1
fi

echo "\n=== 构建完成 ==="
