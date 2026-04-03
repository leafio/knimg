@echo off

rem KnImg Windows 测试脚本

setlocal

echo KnImg Windows 版本测试
set "EXE_NAME=knimg.exe"
set "FRONTEND_DIR=frontend"

rem 检查必要文件
if not exist "%EXE_NAME%" (
    echo 错误: %EXE_NAME% 文件不存在！
    pause
    exit /b 1
)

if not exist "%FRONTEND_DIR%" (
    echo 错误: %FRONTEND_DIR% 目录不存在！
    pause
    exit /b 1
)

if not exist "%FRONTEND_DIR%\index.html" (
    echo 错误: %FRONTEND_DIR%\index.html 文件不存在！
    pause
    exit /b 1
)

echo 所有文件检查通过！
echo 正在启动 KnImg 服务器...
echo 服务器将运行在 http://localhost:8080
echo 按 Ctrl+C 停止服务器

echo.
echo 启动信息:
%EXE_NAME%

endlocal
