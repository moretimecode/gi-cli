@echo off
setlocal
REM ============================================
REM 编译 .NET 分析器 (多平台)
REM ============================================

set "SCRIPT_DIR=%~dp0"

pushd "%SCRIPT_DIR%" >nul

if not exist "%SCRIPT_DIR%internal\dotnets\DeadCodeAnalyzer" (
    echo [错误] 找不到项目目录: %SCRIPT_DIR%internal\dotnets\DeadCodeAnalyzer
    exit /b 1
)

cd /d "%SCRIPT_DIR%internal\dotnets\DeadCodeAnalyzer"

echo.
echo 编译 C# 分析器 (多平台)...
echo.

REM 编译 Windows 版本
echo [1/3] 编译 Windows x64...
dotnet publish -c Release -r win-x64 --self-contained true --nologo /v:minimal
if errorlevel 1 (
    echo [错误] Windows 编译失败！
    popd >nul
    pause
    exit /b 1
)
copy /Y "bin\Release\net10.0\win-x64\publish\DeadCodeAnalyzer.exe" "%SCRIPT_DIR%internal\dotnets\deadcode_windows_amd64.exe" >nul
echo [✓] Windows 完成

REM 编译 Linux 版本
echo [2/3] 编译 Linux x64...
dotnet publish -c Release -r linux-x64 --self-contained true --nologo /v:minimal
if errorlevel 1 (
    echo [错误] Linux 编译失败！
    popd >nul
    pause
    exit /b 1
)
copy /Y "bin\Release\net10.0\linux-x64\publish\DeadCodeAnalyzer" "%SCRIPT_DIR%internal\dotnets\deadcode_linux_amd64" >nul
echo [✓] Linux 完成

REM 编译 macOS 版本
echo [3/3] 编译 macOS x64...
dotnet publish -c Release -r osx-x64 --self-contained true --nologo /v:minimal
if errorlevel 1 (
    echo [错误] macOS 编译失败！
    popd >nul
    pause
    exit /b 1
)
copy /Y "bin\Release\net10.0\osx-x64\publish\DeadCodeAnalyzer" "%SCRIPT_DIR%internal\dotnets\deadcode_darwin_amd64" >nul
echo [✓] macOS 完成

popd >nul

echo.
echo 所有平台编译完成！
echo 文件位置: internal\dotnets\
echo   - deadcode_windows_amd64.exe
echo   - deadcode_linux_amd64
echo   - deadcode_darwin_amd64
echo.
echo 现在可以运行 build.bat 编译 Go 项目
echo.
