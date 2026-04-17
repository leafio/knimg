#!/bin/bash

# KnImg 前端开发服务器启动脚本
# 由于使用了 ES6 模块,需要通过本地服务器访问

echo "🚀 启动 KnImg 前端开发服务器..."
echo ""

# 检查 Python3 是否可用
if command -v python3 &> /dev/null; then
    echo "✅ 使用 Python3 HTTP 服务器"
    echo "📍 访问地址: http://localhost:8080"
    echo "💡 按 Ctrl+C 停止服务器"
    echo ""
    cd "$(dirname "$0")"
    python3 -m http.server 8080
elif command -v python &> /dev/null; then
    echo "✅ 使用 Python HTTP 服务器"
    echo "📍 访问地址: http://localhost:8080"
    echo "💡 按 Ctrl+C 停止服务器"
    echo ""
    cd "$(dirname "$0")"
    python -m SimpleHTTPServer 8080
else
    echo "❌ 未找到 Python,请安装 Python 或使用其他 HTTP 服务器"
    echo "💡 推荐使用: npx http-server -p 8080"
    exit 1
fi
