@echo off
REM Build script for danmaku-service on Windows
setlocal enabledelayedexpansion

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

go build -o build\danmaku-service .\services\danmaku-service\cmd\main.go

if errorlevel 1 (
    echo Build failed for danmaku-service
    exit /b 1
)

echo Build succeeded for danmaku-service
