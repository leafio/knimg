#!/bin/bash

# 跨平台编译脚本
echo "=== KnImg 跨平台编译脚本 ==="
echo "开始构建三个平台的应用程序..."

# 创建输出目录
mkdir -p build

# 设置Go环境变量
export GO111MODULE=on
export GOROOT=/usr/local/go

# 清理之前的构建
rm -rf build/*

# 编译Windows 64位 (窗口应用)
echo "\n编译 Windows 64位 (窗口应用)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui -s -w" -o build/knimg-windows-amd64.exe .

# 编译Mac 64位
echo "\n编译 Mac 64位..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/knimg-darwin-amd64 .

# 编译Mac ARM64位 (M1/M2)
echo "\n编译 Mac ARM64位..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o build/knimg-darwin-arm64 .

# 编译Linux 64位
echo "\n编译 Linux 64位..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/knimg-linux-amd64 .

# 创建 macOS 应用程序包 (AMD64)
echo "\n=== 创建 macOS 应用程序包 (AMD64) ==="
mkdir -p build/KnImg-amd64.app/Contents/MacOS
mkdir -p build/KnImg-amd64.app/Contents/Resources
cp build/knimg-darwin-amd64 build/KnImg-amd64.app/Contents/MacOS/
chmod +x build/KnImg-amd64.app/Contents/MacOS/knimg-darwin-amd64

# 创建 Info.plist
cat > build/KnImg-amd64.app/Contents/Info.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>knimg-darwin-amd64</string>
    <key>CFBundleIdentifier</key>
    <string>com.knimg.app</string>
    <key>CFBundleName</key>
    <string>KnImg</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

# 压缩应用程序包
zip -r build/knimg-macos-amd64.app.zip build/KnImg-amd64.app

# 创建 macOS 应用程序包 (ARM64)
echo "\n=== 创建 macOS 应用程序包 (ARM64) ==="
mkdir -p build/KnImg-arm64.app/Contents/MacOS
mkdir -p build/KnImg-arm64.app/Contents/Resources
cp build/knimg-darwin-arm64 build/KnImg-arm64.app/Contents/MacOS/
chmod +x build/KnImg-arm64.app/Contents/MacOS/knimg-darwin-arm64

# 创建 Info.plist
cat > build/KnImg-arm64.app/Contents/Info.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>knimg-darwin-arm64</string>
    <key>CFBundleIdentifier</key>
    <string>com.knimg.app</string>
    <key>CFBundleName</key>
    <string>KnImg</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

# 压缩应用程序包
zip -r build/knimg-macos-arm64.app.zip build/KnImg-arm64.app

# 创建 Linux 桌面条目文件
echo "\n=== 创建 Linux 桌面条目文件 ==="
cat > build/knimg.desktop << EOF
[Desktop Entry]
Name=KnImg
Exec=./knimg-linux-amd64
Terminal=false
Type=Application
Categories=Utility;
EOF
chmod +x build/knimg.desktop

# 压缩 Linux 构建产物
zip -r build/knimg-linux.zip build/knimg-linux-amd64 build/knimg.desktop

# 检查编译结果
echo "\n=== 构建结果 ==="
ls -la build/

# 检查是否成功
if [ -f "build/knimg-windows-amd64.exe" ] && [ -f "build/knimg-macos-amd64.app.zip" ] && [ -f "build/knimg-macos-arm64.app.zip" ] && [ -f "build/knimg-linux.zip" ]; then
    echo "\n✓ 所有平台构建成功!"
    echo "应用程序包位于 build/ 目录"
    echo "- Windows: knimg-windows-amd64.exe (窗口应用)"
    echo "- macOS Intel: knimg-macos-amd64.app.zip"
    echo "- macOS M1/M2: knimg-macos-arm64.app.zip"
    echo "- Linux: knimg-linux.zip"
else
    echo "\n✗ 构建失败!"
    exit 1
fi

echo "\n=== 构建完成 ==="
