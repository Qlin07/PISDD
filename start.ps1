# ============================================================
#  SimpleChat - One-Click Startup Script / 一键启动脚本
#  Author   : PISDD
#
#  【功能 / Features】
#   一键启动「简聊 SimpleChat」全栈环境, 含:
#   Starts the full SimpleChat stack:
#     1. 数据层 Data layer  - MySQL + Redis + MinIO (Docker Compose)
#     2. 后端   Backend     - Go/Gin  on port 8080
#     3. 前端   Frontend    - Vite    on port 5173
#   等待前后端就绪后, 提示前端访问地址。
#   Waits for services, then prints the frontend URL.
#
#  【用法 / Usage】PowerShell 中运行 (Run in PowerShell):
#    正常启动 Start   : powershell -ExecutionPolicy Bypass -File start.ps1
#    重启   Restart   : powershell -ExecutionPolicy Bypass -File start.ps1 -Restart
#    停止   Stop      : powershell -ExecutionPolicy Bypass -File start.ps1 -Stop
#    或右键「使用 PowerShell 运行」
#    Or right-click -> "Run with PowerShell".
#
#  【前置要求 / Prerequisites】
#    - Docker Desktop 已安装并运行
#    - Go 1.22+ 、 Node.js 、 npm 已安装
#    - 首次运行前已在 server/ client/ 执行 `go mod tidy`、`npm install`
#
#  【注意事项 / Notes】
#    - 后端与前端以最小化窗口在后台启动 (launched in minimized windows)
#    - MySQL 使用宿主机端口 13306 (避开 Windows 端口排除范围占用 3306)
#    - MinIO 控制台: http://localhost:9001  (账号 minioadmin / 密码 minioadmin123)
# ============================================================

param(
  [switch]$Stop = $false,
  [switch]$Restart = $false
)

$Root   = $PSScriptRoot
$DbDir  = Join-Path $Root "database"
$SrvDir = Join-Path $Root "server"
$CliDir = Join-Path $Root "client"

$GoBin = Join-Path $env:USERPROFILE ".local\lib\go\bin"
if (Test-Path $GoBin) { $env:Path = "$GoBin;$env:Path" }

function Write-Step([string]$msg) { Write-Host "`n[$(Get-Date -Format 'HH:mm:ss')] $msg" -ForegroundColor Cyan }
function Write-Ok  ([string]$msg) { Write-Host "  OK  -> $msg" -ForegroundColor Green }
function Write-Warn([string]$msg) { Write-Host "  SKIP-> $msg" -ForegroundColor Yellow }

# ---------- Stop everything ----------
function Stop-All {
  Write-Step "Stopping backend (port 8080) and frontend (port 5173)..."
  foreach ($port in 8080, 5173) {
    Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue |
      ForEach-Object { Stop-Process -Id $_.OwningProcess -Force -ErrorAction SilentlyContinue }
  }
  Write-Ok "Backend / frontend stopped."

  Write-Step "Stopping data layer containers..."
  if (Test-Path (Join-Path $DbDir "docker-compose.yml")) {
    Push-Location $DbDir
    docker compose down --remove-orphans 2>&1 | Out-Null
    Pop-Location
    Write-Ok "Docker containers stopped (data volumes kept)."
  }
}

# ---------- 检查 Docker Desktop 是否可用 ----------
function Assert-DockerRunning {
  docker info > $null 2>&1
  if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Warn "Docker 不可用 (docker info 失败)."
    Write-Host "  请先启动 Docker Desktop, 等待其就绪后再运行本脚本." -ForegroundColor Yellow
    return $false
  }
  return $true
}

# ---------- Start data layer ----------
function Start-Database {
  Write-Step "Starting data layer (MySQL + Redis + MinIO) via Docker Compose..."
  if (-not (Test-Path (Join-Path $DbDir "docker-compose.yml"))) {
    Write-Warn "docker-compose.yml not found in $DbDir"
    return
  }
  if (-not (Assert-DockerRunning)) { return }

  Write-Host "  docker compose up -d ..."
  Push-Location $DbDir
  $composeOut = $(docker compose up -d 2>&1)
  $upCode = $LASTEXITCODE
  Pop-Location
  if ($upCode -ne 0) {
    Write-Warn "docker compose 启动失败 (退出码 $upCode):"
    $composeOut | ForEach-Object { "    $_" }
    Write-Host "  常见原因: 端口被占用。若报 3306, 请在 database/docker-compose.yml 中改用 13306。" -ForegroundColor Yellow
    return
  }
  # Docker 会把 "Running/Created" 等提示写到 stderr, 这里统一干净地展示
  $composeOut | Where-Object { $_ -and $_.ToString().Trim() } | ForEach-Object { "    $_" }
  Write-Ok "Containers launched."

  # Wait for MySQL to become healthy (up to 60s)
  Write-Step "Waiting for MySQL to be healthy..."
  $ready = $false
  for ($i = 0; $i -lt 30; $i++) {
    Start-Sleep -Seconds 2
    $state = @(docker inspect --format '{{.State.Health.Status}}' simplechat-mysql 2>$null)
    if ($state -and $state[0] -eq "healthy") { $ready = $true; break }
  }
  if ($ready) {
    Write-Ok "MySQL healthy."
  } else {
    Write-Warn "MySQL 未能在 60s 内变为 healthy。以下是诊断信息:"
    Push-Location $DbDir
    docker compose ps
    Write-Host "MySQL 最近日志 (tail):" -ForegroundColor Yellow
    docker logs --tail 30 simplechat-mysql 2>&1
    Pop-Location
    Write-Host "  提示: MySQL 首次初始化需下载镜像并建表, 通常需要几十秒。" -ForegroundColor Yellow
  }
}

# ---------- Start backend ----------
function Start-Backend {
  Write-Step "Starting backend server (Go/Gin) ..."
  Start-Process -FilePath "cmd.exe" -ArgumentList "/c", "go run main.go 2>&1" `
    -WorkingDirectory $SrvDir -WindowStyle Minimized
  Write-Ok "Backend launching on http://localhost:8080"
}

# ---------- Start frontend ----------
function Start-Frontend {
  Write-Step "Starting frontend client (Vite) ..."
  # npm on Windows is npm.cmd / npm.ps1; launch through cmd to be reliable
  Start-Process -FilePath "cmd.exe" -ArgumentList "/c", "npm run dev 2>&1" `
    -WorkingDirectory $CliDir -WindowStyle Minimized
  Write-Ok "Frontend launching on http://localhost:5173"
}

# ---------- Print result ----------
function Show-Summary {
  Write-Host ""
  Write-Host "==================================================" -ForegroundColor Cyan
  Write-Host "  SimpleChat is starting up" -ForegroundColor Cyan
  Write-Host "==================================================" -ForegroundColor Cyan
  Write-Host "  Frontend  : http://localhost:5173"
  Write-Host "  Backend   : http://localhost:8080/api"
  Write-Host "  MinIO UI  : http://localhost:9001  (minioadmin/minioadmin123)"
  Write-Host "  (Vite and Go are launched in minimized windows.)"
  Write-Host "==================================================" -ForegroundColor Cyan
}

# ---------- Main ----------
if ($Stop) { Stop-All; return }

if ($Restart) {
  Write-Step "Restart requested: stopping existing services first."
  Stop-All
  Write-Host ""
}

Start-Database
Start-Backend
Start-Frontend

# Wait until the frontend is actually serving, then show the address
Write-Step "Waiting for services to come online (up to 90s)..."
function Test-Tcp([int]$port) {
  try {
    $client = New-Object System.Net.Sockets.TcpClient
    $client.Connect("127.0.0.1", $port)
    $client.Close()
    return $true
  } catch { return $false }
}
$frontReady = $false
$backReady = $false
$deadline = (Get-Date).AddSeconds(90)
while ((Get-Date) -lt $deadline) {
  if (-not $frontReady)  { $frontReady  = Test-Tcp 5173 }
  if (-not $backReady)   { $backReady   = Test-Tcp 8080 }
  if ($frontReady -and $backReady) { break }
  Start-Sleep -Seconds 3
}

Show-Summary

# Final prompt with the frontend address
Write-Host ""
Write-Host "=========================================================" -ForegroundColor Green
if ($frontReady) {
  Write-Host "  SUCCESS! SimpleChat is ready." -ForegroundColor Green
  Write-Host "  Open the frontend in your browser now:" -ForegroundColor Green
  Write-Host "      >>>  http://localhost:5173  <<<" -ForegroundColor Green
} else {
  Write-Host "  Frontend not reachable yet (still starting or failed)." -ForegroundColor Yellow
  Write-Host "  Try again in a moment:  http://localhost:5173" -ForegroundColor Yellow
}
Write-Host "=========================================================" -ForegroundColor Green