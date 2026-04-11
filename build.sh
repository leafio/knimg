#!/bin/bash

# 跨平台编译脚本
echo "=== KnImg 跨平台编译脚本 ==="
echo "开始构建各平台的应用程序..."

# 创建输出目录
mkdir -p build

# 设置Go环境变量
export GO111MODULE=on
export GOROOT=/usr/local/go

# 不清理之前的构建，保留之前成功构建的文件
# rm -rf build/*

# 成功构建的平台列表
success_platforms=()

# 编译Windows 64位 (窗口应用)
echo "\n编译 Windows 64位 (窗口应用)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui -s -w" -o build/knimg-windows-amd64.exe .
if [ $? -eq 0 ]; then
    echo "✓ Windows 64位构建成功"
    success_platforms+=("Windows")
else
    echo "✗ Windows 64位构建失败"
fi

# 编译Mac 64位
echo "\n编译 Mac 64位..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/knimg-darwin-amd64 .
if [ $? -eq 0 ]; then
    echo "✓ Mac 64位构建成功"
    success_platforms+=("Mac_AMD64")
    
    # 创建 macOS 应用程序包 (AMD64)
    echo "创建 macOS 应用程序包 (AMD64)..."
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
    echo "✓ macOS AMD64 应用程序包创建成功"
else
    echo "✗ Mac 64位构建失败"
fi

# 编译Mac ARM64位 (M1/M2)
echo "\n编译 Mac ARM64位..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o build/knimg-darwin-arm64 .
if [ $? -eq 0 ]; then
    echo "✓ Mac ARM64位构建成功"
    success_platforms+=("Mac_ARM64")
    
    # 创建 macOS 应用程序包 (ARM64)
    echo "创建 macOS 应用程序包 (ARM64)..."
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
    echo "✓ macOS ARM64 应用程序包创建成功"
else
    echo "✗ Mac ARM64位构建失败"
fi

# 编译Linux 64位
echo "\n编译 Linux 64位..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/knimg-linux-amd64 .
if [ $? -eq 0 ]; then
    echo "✓ Linux 64位构建成功"
    success_platforms+=("Linux")
    
    # 创建 Linux 桌面条目文件
    echo "创建 Linux 桌面条目文件..."
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
    echo "✓ Linux 构建产物压缩成功"
else
    echo "✗ Linux 64位构建失败"
fi

# 检查编译结果
echo "\n=== 构建结果 ==="
ls -la build/

# 检查是否有成功构建的平台
if [ ${#success_platforms[@]} -eq 0 ]; then
    echo "\n✗ 所有平台构建失败!"
    exit 1
else
    echo "\n✓ 构建完成!"
    echo "成功构建的平台: ${success_platforms[*]}"
    echo "应用程序包位于 build/ 目录"
    
    # 输出成功构建的文件
    if [ -f "build/knimg-windows-amd64.exe" ]; then
        echo "- Windows: knimg-windows-amd64.exe (窗口应用)"
    fi
    if [ -f "build/knimg-macos-amd64.app.zip" ]; then
        echo "- macOS Intel: knimg-macos-amd64.app.zip"
    fi
    if [ -f "build/knimg-macos-arm64.app.zip" ]; then
        echo "- macOS M1/M2: knimg-macos-arm64.app.zip"
    fi
    if [ -f "build/knimg-linux.zip" ]; then
        echo "- Linux: knimg-linux.zip"
    fi
fi

echo "\n=== 构建完成 ==="
