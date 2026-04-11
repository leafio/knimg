#!/bin/bash

# 构建脚本，优化二进制体积

echo "=== KnImg 构建脚本 ==="
echo "开始构建..."

# 清理之前的构建
rm -f knimg

# 构建参数说明：
# -ldflags="-s -w" 移除符号表和调试信息，减小体积
# -buildmode=pie 启用位置无关可执行文件
# -trimpath 移除构建路径信息
go build -o knimg \
  -ldflags="-s -w" \
  -buildmode=pie \
  -trimpath \
  main.go

if [ $? -eq 0 ]; then
    echo "✓ 构建成功！"
    echo "二进制文件大小: $(du -h knimg | cut -f1)"
    
    # 可选：使用 UPX 进一步压缩（如果安装了 UPX）
    if command -v upx &> /dev/null; then
        echo "使用 UPX 压缩二进制文件..."
        upx --best knimg
        echo "✓ 压缩完成！"
        echo "压缩后大小: $(du -h knimg | cut -f1)"
    else
        echo "提示：安装 UPX 可以进一步减小二进制体积"
    fi
else
    echo "✗ 构建失败"
    exit 1
fi

echo "=== 构建完成 ==="
