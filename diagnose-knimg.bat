@echo off

rem KnImg 诊断脚本

setlocal enabledelayedexpansion

echo =======================
echo KnImg 诊断脚本
echo =======================
echo 运行时间: %date% %time%
echo =======================
echo.

rem 显示当前目录
echo 1. 当前工作目录:
echo    %cd%
echo.

rem 检查文件结构
echo 2. 文件结构检查:

if exist "knimg.exe" (
    echo    [✓] knimg.exe 文件存在
    for %%f in (knimg.exe) do (
        echo    大小: %%~zf 字节
        echo    路径: "%%~dpnxf"
    )
) else (
    echo    [✗] knimg.exe 文件不存在！
)
echo.

if exist "frontend" (
    echo    [✓] frontend 目录存在
    if exist "frontend\index.html" (
        echo    [✓] frontend\index.html 文件存在
        for %%f in (frontend\index.html) do (
            echo    大小: %%~zf 字节
        )
    ) else (
        echo    [✗] frontend\index.html 文件不存在！
    )
) else (
    echo    [✗] frontend 目录不存在！
)
echo.

rem 显示目录内容
echo 3. 目录内容:
dir /b
echo.

echo 4. 环境变量:
echo    USERPROFILE: %USERPROFILE%
echo    TEMP: %TEMP%
echo    PATH: %PATH:~0,100%...
echo.

rem 检查端口
echo 5. 端口检查:
netstat -ano | findstr :8080
echo.

echo 6. 诊断结论:
if exist "knimg.exe" ( 
    if exist "frontend" ( 
        if exist "frontend\index.html" ( 
            echo    [✓] 所有必要文件都存在
            echo    [✓] 可以运行 knimg.exe
            echo.
            echo    建议: 双击运行 run-knimg.bat 或直接运行 knimg.exe
        ) else (
            echo    [✗] 缺少 frontend\index.html 文件
        )
    ) else (
        echo    [✗] 缺少 frontend 目录
    )
) else (
    echo    [✗] 缺少 knimg.exe 文件
)
echo.
echo 7. 解决方案:
echo    1. 确保 knimg.exe 和 frontend 文件夹在同一目录

echo    2. 确保 frontend 目录包含 index.html 文件

echo    3. 如果文件缺失，请重新解压 knimg-windows.zip

echo.
echo =======================
echo 诊断完成

echo 按任意键退出...
pause >nul
endlocal
