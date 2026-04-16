@echo off
REM Build script for user-service on Windows
setlocal enabledelayedexpansion

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

go build -o build\user-service .\services\user-service\cmd\main.go

if errorlevel 1 (
    echo Build failed for user-service
    exit /b 1
)

echo Build succeeded for user-service
