# 构建所有服务的可执行文件（Linux 格式，用于 Docker）

Write-Host "Building services for Linux..." -ForegroundColor Green

# 创建 build 目录
if (-not (Test-Path "build")) {
    New-Item -ItemType Directory -Path "build" | Out-Null
}

# 设置环境变量：编译为 Linux 可执行文件
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"

# 要构建的服务列表
$services = @(
    @{ "name" = "api-service"; "path" = "./services/api-service/cmd/main.go"; "output" = "build/api-service" },
    @{ "name" = "danmaku-service"; "path" = "./services/danmaku-service/cmd/main.go"; "output" = "build/danmaku-service" }
    @{ "name" = "user-service"; "path" = "./services/user-service/cmd/main.go"; "output" = "build/user-service" }
    @{ "name" = "room-service"; "path" = "./services/room-service/cmd/main.go"; "output" = "build/room-service" }
)

# 逐个构建服务
$buildSuccess = $true
foreach ($service in $services) {
    Write-Host "Building $($service.name)..." -ForegroundColor Cyan
    
    go build -o $($service.output) $($service.path)
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Build successful: $($service.output)" -ForegroundColor Green
    } else {
        Write-Host "Build failed: $($service.name)" -ForegroundColor Red
        $buildSuccess = $false
    }
}

if ($buildSuccess) {
    Write-Host "`nAll services built successfully!" -ForegroundColor Green
} else {
    Write-Host "`nBuild failed!" -ForegroundColor Red
    exit 1
}
