param(
    [string]$DatabaseHost = "117.72.123.14",
    [int]$DatabasePort = 3306,
    [string]$DatabaseName = "steellink",
    [string]$DatabaseUsername = "root",
    [string]$DatabasePassword = "",
    [string]$RedisHost = "117.72.123.14",
    [int]$RedisPort = 6379,
    [string]$RedisPassword = "",
    [string]$JwtSecret = "steellink-dev-secret",
    [int]$ServerPort = 8080,
    [string]$ServerMode = "debug"
)

$ErrorActionPreference = "Stop"

function Resolve-GoExe {
    $candidates = @(
        (Join-Path $env:LOCALAPPDATA "Programs\Go-1.26.2\go\bin\go.exe"),
        "C:\Program Files\Go\bin\go.exe",
        "C:\Go\bin\go.exe"
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    throw "Go executable not found. Install Go locally or place it under $env:LOCALAPPDATA\Programs\Go-1.26.2."
}

Set-Location -LiteralPath (Split-Path -Parent $PSScriptRoot)

$goExe = Resolve-GoExe

$env:DATABASE_HOST = $DatabaseHost
$env:DATABASE_PORT = "$DatabasePort"
$env:DATABASE_NAME = $DatabaseName
$env:DATABASE_USERNAME = $DatabaseUsername
$env:DATABASE_PASSWORD = $DatabasePassword
$env:REDIS_HOST = $RedisHost
$env:REDIS_PORT = "$RedisPort"
$env:REDIS_PASSWORD = $RedisPassword
$env:JWT_SECRET = $JwtSecret
$env:SERVER_PORT = "$ServerPort"
$env:SERVER_MODE = $ServerMode

Write-Host "Starting SteelLink backend with remote services..." -ForegroundColor Cyan
Write-Host "MySQL: $DatabaseHost:$DatabasePort / $DatabaseName" -ForegroundColor DarkGray
Write-Host "Redis: $RedisHost:$RedisPort" -ForegroundColor DarkGray
Write-Host "API: http://localhost:$ServerPort" -ForegroundColor DarkGray

& $goExe run .\cmd\api\main.go
