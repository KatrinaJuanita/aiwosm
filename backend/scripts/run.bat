@echo off
echo ==================================
echo       WOSM Go Backend 运行脚本
echo ==================================

echo 1. 检查Redis服务...
tasklist /fi "imagename eq redis-server.exe" | find "redis-server.exe" > nul
if not errorlevel 1 (
    echo Redis服务已运行
) else (
    echo 启动Redis服务...
    start "" "..\..\redis\redis-server.exe" "..\..\redis\redis.windows.conf"
    timeout /t 3 /nobreak > nul
)

echo 2. 下载Go依赖...
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo 依赖下载失败
    goto :end
)

echo 3. 启动WOSM Go Backend...
if exist wosm.exe (
    echo 使用已构建的可执行文件...
    wosm.exe
) else (
    echo 直接运行Go代码...
    go run ./cmd/main.go
)

:end
pause
