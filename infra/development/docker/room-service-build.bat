@echo off
REM Build script for room-service on Windows
setlocal enabledelayedexpansion

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

go build -o build\room-service .\services\room-service\cmd\main.go

if errorlevel 1 (
    echo Build failed for room-service
    exit /b 1
)

echo Build succeeded for room-service
