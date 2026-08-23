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

# 本地打包模式：只为当前系统构建并打包
if [ "$1" = "local" ]; then
    LOCAL_OS=$(go env GOOS)
    LOCAL_ARCH=$(go env GOARCH)
    OUT="knimg-${LOCAL_OS}-${LOCAL_ARCH}"

    echo "=== KnImg 本地打包 (${LOCAL_OS}/${LOCAL_ARCH}) ==="

    LDFLAGS="-s -w"
    if [ "$LOCAL_OS" = "windows" ]; then
        LDFLAGS="-H=windowsgui -s -w"
    fi

    if ! GOOS=$LOCAL_OS GOARCH=$LOCAL_ARCH go build -ldflags="$LDFLAGS" -o "build/$OUT" .; then
        echo "✗ 本地构建失败"
        exit 1
    fi
    echo "✓ 构建成功: build/$OUT"

    if [ "$LOCAL_OS" = "darwin" ]; then
        APP="KnImg-${LOCAL_ARCH}.app"
        TMP_APP="build/${APP}.tmp"
        rm -rf "$TMP_APP"
        mkdir -p "$TMP_APP/Contents/MacOS" "$TMP_APP/Contents/Resources"
        cp "build/$OUT" "$TMP_APP/Contents/MacOS/"
        chmod +x "$TMP_APP/Contents/MacOS/$OUT"

        cat > "$TMP_APP/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>$OUT</string>
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

        rm -rf "build/$APP"
        mv "$TMP_APP" "build/$APP"
        echo "✓ 应用包创建成功: build/$APP"
    elif [ "$LOCAL_OS" = "windows" ]; then
        echo "✓ Windows 可执行文件: build/knimg-windows-amd64.exe"
    else
        echo "✓ 可执行文件: build/$OUT"
    fi

    echo "=== 本地打包完成 ==="
    ls -la build/
    exit 0
fi

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
    # 使用临时目录，确保构建失败时不会影响现有版本
    temp_app_dir="build/KnImg-amd64.app.tmp"
    rm -rf "$temp_app_dir"
    mkdir -p "$temp_app_dir/Contents/MacOS"
    mkdir -p "$temp_app_dir/Contents/Resources"
    cp build/knimg-darwin-amd64 "$temp_app_dir/Contents/MacOS/"
    chmod +x "$temp_app_dir/Contents/MacOS/knimg-darwin-amd64"
    
    # 创建 Info.plist
    cat > "$temp_app_dir/Contents/Info.plist" << EOF
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
    zip -r build/knimg-macos-amd64.app.zip "$temp_app_dir"
    
    # 只有在所有操作都成功后，才替换现有应用程序包
    if [ $? -eq 0 ]; then
        rm -rf build/KnImg-amd64.app
        mv "$temp_app_dir" build/KnImg-amd64.app
        echo "✓ macOS AMD64 应用程序包创建成功"
    else
        echo "✗ macOS AMD64 应用程序包压缩失败"
        rm -rf "$temp_app_dir"
    fi
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
    # 使用临时目录，确保构建失败时不会影响现有版本
    temp_app_dir="build/KnImg-arm64.app.tmp"
    rm -rf "$temp_app_dir"
    mkdir -p "$temp_app_dir/Contents/MacOS"
    mkdir -p "$temp_app_dir/Contents/Resources"
    cp build/knimg-darwin-arm64 "$temp_app_dir/Contents/MacOS/"
    chmod +x "$temp_app_dir/Contents/MacOS/knimg-darwin-arm64"
    
    # 创建 Info.plist
    cat > "$temp_app_dir/Contents/Info.plist" << EOF
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
    zip -r build/knimg-macos-arm64.app.zip "$temp_app_dir"
    
    # 只有在所有操作都成功后，才替换现有应用程序包
    if [ $? -eq 0 ]; then
        rm -rf build/KnImg-arm64.app
        mv "$temp_app_dir" build/KnImg-arm64.app
        echo "✓ macOS ARM64 应用程序包创建成功"
    else
        echo "✗ macOS ARM64 应用程序包压缩失败"
        rm -rf "$temp_app_dir"
    fi
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
