# 构建 api-service 可执行文件（Linux 格式，用于 Docker）
param(
    [string]$Output = "build/api-service"
)

Write-Host "Building api-service for Linux..." -ForegroundColor Green

# 创建 build 目录
if (-not (Test-Path "build")) {
    New-Item -ItemType Directory -Path "build" | Out-Null
}

# 设置环境变量：编译为 Linux 可执行文件
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"

go build -o $Output ./services/api-service/cmd/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful: $Output (Linux binary)" -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
