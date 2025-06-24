@echo off
echo ==================================
echo       WOSM Go Backend 构建脚本
echo ==================================

echo 1. 检查Go环境...
go version
if %ERRORLEVEL% NEQ 0 (
    echo Go环境未安装或未配置PATH
    goto :end
)

echo 2. 下载依赖...
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo 依赖下载失败
    goto :end
)

echo 3. 构建应用程序...
go build -o wosm.exe ./cmd/main.go
if %ERRORLEVEL% NEQ 0 (
    echo 构建失败
    goto :end
)

echo 4. 构建完成！
echo 可执行文件: wosm.exe

:end
pause
