@echo off

rem 测试脚本 - 运行 knimg.exe 并显示输出

echo 正在测试 KnImg Windows 版本...
echo ===================================

rem 检查 knimg.exe 是否存在
if not exist "dist\windows\knimg.exe" (
    echo 错误：knimg.exe 文件不存在！
    pause
    exit /b 1
)

rem 复制到当前目录进行测试
copy "dist\windows\knimg.exe" .\ /Y
if errorlevel 1 (
    echo 错误：无法复制 knimg.exe 文件！
    pause
    exit /b 1
)

rem 复制前端目录
if not exist "frontend" (
    echo 错误：frontend 目录不存在！
    pause
    exit /b 1
)

rem 运行 knimg.exe
 echo 正在运行 knimg.exe...
 .\knimg.exe

rem 等待用户按任意键
pause
