Write-Host "===> Start building frp for Linux with web assets..."

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

# 输出目录
$OutputDir = "C:\code\frp\output"

# 创建 output 目录
if (!(Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# 清理旧文件（防冲突）
if (Test-Path "$OutputDir/server") {
    Remove-Item "$OutputDir/server" -Force
}
if (Test-Path "$OutputDir/client") {
    Remove-Item "$OutputDir/client" -Force
}

# 构建参数
$BuildFlags = @(
    "-trimpath",
    "-ldflags=-s -w"
)

# ========================
# 构建前端 web
# ========================
Write-Host "===> Building web/frps..."
Push-Location "web/frps"
try {
    yarn install
    if ($LASTEXITCODE -ne 0) { throw "web/frps install failed" }

    yarn build
    if ($LASTEXITCODE -ne 0) { throw "web/frps build failed" }
}
finally {
    Pop-Location
}

Write-Host "===> Building web/frpc..."
Push-Location "web/frpc"
try {
    yarn install
    if ($LASTEXITCODE -ne 0) { throw "web/frpc install failed" }

    yarn build
    if ($LASTEXITCODE -ne 0) { throw "web/frpc build failed" }
}
finally {
    Pop-Location
}

# ========================
# 构建 Go 二进制
# ========================
Write-Host "===> Building frps..."
go build @BuildFlags -tags=frps -o "$OutputDir/server" ./cmd/frps
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ frps build failed"
    exit 1
}

Write-Host "===> Building frpc..."
go build @BuildFlags -tags=frpc -o "$OutputDir/client" ./cmd/frpc
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ frpc build failed"
    exit 1
}

Write-Host "✅ Build completed!"
Write-Host "Output:"
Write-Host " - $OutputDir/server"
Write-Host " - $OutputDir/client"