@echo off
chcp 65001 >nul
setlocal

REM 获取版本号 (从 git tag 或手动设置)
for /f "delims=" %%i in ('git describe --tags --abbrev=0 2^>nul') do set "VERSION=%%i"
if not defined VERSION set "VERSION=dev"

REM 获取构建时间
for /f "delims=" %%i in ('powershell -Command "Get-Date -Format \"yyyy-MM-dd\""') do set "DATE=%%i"

echo ============================================
echo Git Tools 构建脚本
echo 版本: %VERSION%
echo 构建时间: %DATE%
echo ============================================

set OUTPUT_DIR=bin

if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

echo 正在使用 wails build...
wails build -platform windows/amd64 -o "%OUTPUT_DIR%\git-tools-%VERSION%-windows-x64.exe"

if %ERRORLEVEL% equ 0 (
    echo ============================================
    echo 构建成功！
    echo 输出: %OUTPUT_DIR%\git-tools-%VERSION%-windows-x64.exe
    echo ============================================
) else (
    echo 构建失败！
    exit /b 1
)

endlocal
pause
