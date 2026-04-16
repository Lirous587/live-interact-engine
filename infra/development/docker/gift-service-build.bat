@echo off
REM Build script for gift-service on Windows
setlocal enabledelayedexpansion

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

go build -o build\gift-service .\services\gift-service\cmd\main.go

if errorlevel 1 (
    echo Build failed for gift-service
    exit /b 1
)

echo Build succeeded for gift-service
