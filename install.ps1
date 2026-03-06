#Requires -Version 5.1
[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$repo    = "danielmrdev/dtasks-cli"
$binary  = "dtasks"
$apiUrl  = "https://api.github.com/repos/$repo/releases/latest"

# ── Detect arch ──────────────────────────────────────────────────────────────
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default {
        Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
        exit 1
    }
}

$asset = "${binary}-windows-${arch}.exe"

# ── Resolve latest version ───────────────────────────────────────────────────
Write-Host "Fetching latest release..."
$release = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
$version = $release.tag_name

if (-not $version) {
    Write-Error "Could not determine latest version"
    exit 1
}

$downloadUrl  = "https://github.com/$repo/releases/download/$version/$asset"
$checksumUrl  = "https://github.com/$repo/releases/download/$version/checksums.txt"

# ── Install dir ──────────────────────────────────────────────────────────────
$installDir = Join-Path $env:LOCALAPPDATA "Programs\dtasks"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}
$dest = Join-Path $installDir "$binary.exe"

# ── Download ──────────────────────────────────────────────────────────────────
Write-Host "Downloading dtasks $version (windows/$arch)..."
$tmp = Join-Path $env:TEMP "$asset"
Invoke-WebRequest -Uri $downloadUrl -OutFile $tmp -UseBasicParsing

# ── Verify checksum ───────────────────────────────────────────────────────────
Write-Host "Verifying checksum..."
try {
    $checksums = (Invoke-WebRequest -Uri $checksumUrl -UseBasicParsing).Content
    $expected  = ($checksums -split "`n" | Where-Object { $_ -match $asset }) -replace '\s+.*', '' | Select-Object -First 1
    if ($expected) {
        $actual = (Get-FileHash -Path $tmp -Algorithm SHA256).Hash.ToLower()
        if ($actual -ne $expected.ToLower()) {
            Write-Error "Checksum mismatch — aborting"
            Remove-Item $tmp -Force
            exit 1
        }
        Write-Host "Checksum OK"
    }
} catch {
    Write-Warning "Could not verify checksum: $_"
}

# ── Install ───────────────────────────────────────────────────────────────────
Move-Item -Path $tmp -Destination $dest -Force
Write-Host "Installed dtasks $version -> $dest"

# ── Add to user PATH if needed ───────────────────────────────────────────────
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$userPath;$installDir", "User")
    Write-Host ""
    Write-Host "$installDir added to your PATH."
    Write-Host "Restart your terminal for the change to take effect."
}

# ── Shell completions ─────────────────────────────────────────────────────────
if ($Host.UI.RawUI -ne $null -and [System.Environment]::UserInteractive) {
    $answer = Read-Host "Install PowerShell completions? [y/N]"
    if ($answer -match "^[Yy]") {
        if (-not (Test-Path $PROFILE)) {
            New-Item -ItemType File -Path $PROFILE -Force | Out-Null
        }
        $completionLine = "`n# dtasks shell completion`n& `"$dest`" completion powershell | Out-String | Invoke-Expression"
        $profileContent = Get-Content $PROFILE -Raw -ErrorAction SilentlyContinue
        if ($profileContent -notlike "*dtasks completion*") {
            Add-Content -Path $PROFILE -Value $completionLine
            Write-Host "PowerShell completions added to $PROFILE"
        } else {
            Write-Host "PowerShell completions already in $PROFILE"
        }
    }
}
