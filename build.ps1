Write-Host "===> Start building frp for Linux with web assets..."

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

$OutputDir = (Get-Location).Path

$BuildFlags = @(
    "-trimpath",
    "-ldflags=-s -w"
)

# ⚠️ 删除同名目录（关键）
if (Test-Path "$OutputDir/server" -PathType Container) {
    Write-Host "⚠️ Removing existing folder: server"
    Remove-Item "$OutputDir/server" -Recurse -Force
}

if (Test-Path "$OutputDir/client" -PathType Container) {
    Write-Host "⚠️ Removing existing folder: client"
    Remove-Item "$OutputDir/client" -Recurse -Force
}

# ===== 构建 web =====
Write-Host "===> Building web/frps..."
Push-Location "web/frps"
yarn install
yarn build
Pop-Location

Write-Host "===> Building web/frpc..."
Push-Location "web/frpc"
yarn install
yarn build
Pop-Location

# ===== 构建 Go =====
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