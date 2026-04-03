@echo off

rem 简化版 KnImg 启动脚本

echo =======================
echo KnImg 启动脚本
echo =======================

rem 检查 knimg.exe
if not exist "knimg.exe" (
    echo 错误: knimg.exe 文件不存在！
    pause
    exit /b 1
)

rem 检查 frontend 目录
if not exist "frontend" (
    echo 错误: frontend 目录不存在！
    pause
    exit /b 1
)

rem 检查 index.html
if not exist "frontend\index.html" (
    echo 错误: frontend\index.html 文件不存在！
    pause
    exit /b 1
)

echo 所有文件检查通过！
echo 正在启动 KnImg 服务器...
echo 服务器地址: http://localhost:8080
echo 按 Ctrl+C 停止服务器
echo.

rem 运行 knimg.exe
knimg.exe

pause
