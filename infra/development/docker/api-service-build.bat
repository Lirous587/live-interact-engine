@echo off
REM Build script for api-service on Windows
setlocal enabledelayedexpansion

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

go build -o build\api-service .\services\api-service\cmd\main.go

if errorlevel 1 (
    echo Build failed for api-service
    exit /b 1
)

echo Build succeeded for api-service
