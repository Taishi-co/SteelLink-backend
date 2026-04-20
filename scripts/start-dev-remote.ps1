param(
    [string]$DatabaseHost = "",
    [int]$DatabasePort = 3306,
    [string]$DatabaseName = "steellink",
    [string]$DatabaseUsername = "root",
    [string]$DatabasePassword = "",
    [string]$RedisHost = "",
    [int]$RedisPort = 6379,
    [string]$RedisPassword = "",
    [string]$JwtSecret = "steellink-dev-secret",
    [int]$ServerPort = 8080,
    [string]$ServerMode = "debug"
)

$ErrorActionPreference = "Stop"

function Resolve-GoExe {
    $goCommand = Get-Command go -ErrorAction SilentlyContinue
    if ($goCommand) {
        return $goCommand.Source
    }

    $candidates = @()

    $localProgramRoot = Join-Path $env:LOCALAPPDATA "Programs"
    if (Test-Path $localProgramRoot) {
        $candidates += Get-ChildItem -Path $localProgramRoot -Directory -Filter "Go*" -ErrorAction SilentlyContinue |
            ForEach-Object { Join-Path $_.FullName "go\bin\go.exe" }
    }

    $candidates += @(
        "C:\Program Files\Go\bin\go.exe",
        "C:\Go\bin\go.exe"
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    throw "Go executable not found. Install Go and ensure `go` is on PATH, or install it into a standard Windows location."
}

Set-Location -LiteralPath (Split-Path -Parent $PSScriptRoot)

$goExe = Resolve-GoExe
$resolvedDatabaseHost = if ($DatabaseHost) { $DatabaseHost } elseif ($env:STEELLINK_REMOTE_DB_HOST) { $env:STEELLINK_REMOTE_DB_HOST } else { "" }
$resolvedRedisHost = if ($RedisHost) { $RedisHost } elseif ($env:STEELLINK_REMOTE_REDIS_HOST) { $env:STEELLINK_REMOTE_REDIS_HOST } elseif ($resolvedDatabaseHost) { $resolvedDatabaseHost } else { "" }

if (-not $resolvedDatabaseHost) {
    throw "DatabaseHost is required. Pass -DatabaseHost or set STEELLINK_REMOTE_DB_HOST first."
}

if (-not $resolvedRedisHost) {
    throw "RedisHost is required. Pass -RedisHost, or set STEELLINK_REMOTE_REDIS_HOST."
}

$env:DATABASE_HOST = $resolvedDatabaseHost
$env:DATABASE_PORT = "$DatabasePort"
$env:DATABASE_NAME = $DatabaseName
$env:DATABASE_USERNAME = $DatabaseUsername
$env:DATABASE_PASSWORD = $DatabasePassword
$env:REDIS_HOST = $resolvedRedisHost
$env:REDIS_PORT = "$RedisPort"
$env:REDIS_PASSWORD = $RedisPassword
$env:JWT_SECRET = $JwtSecret
$env:SERVER_PORT = "$ServerPort"
$env:SERVER_MODE = $ServerMode

Write-Host "Starting SteelLink backend with remote services..." -ForegroundColor Cyan
Write-Host "MySQL: ${resolvedDatabaseHost}:$DatabasePort / $DatabaseName" -ForegroundColor DarkGray
Write-Host "Redis: ${resolvedRedisHost}:$RedisPort" -ForegroundColor DarkGray
Write-Host "API: http://localhost:$ServerPort" -ForegroundColor DarkGray

& $goExe run .\cmd\api\main.go
