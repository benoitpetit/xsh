#
# install.ps1 - xsh Installation Script for Windows (PowerShell)
# Usage: iwr -useb https://raw.githubusercontent.com/benoitpetit/xsh/master/scripts/install.ps1 | iex
#
# This script downloads and installs xsh automatically on Windows
#

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "benoitpetit/xsh"
$BinaryName = "xsh.exe"
$InstallDir = "$env:LOCALAPPDATA\xsh"

# Colors
function Write-Success { param($Message) Write-Host "✅ $Message" -ForegroundColor Green }
function Write-Error { param($Message) Write-Host "❌ $Message" -ForegroundColor Red }
function Write-Info { param($Message) Write-Host "ℹ️  $Message" -ForegroundColor Cyan }
function Write-Warning { param($Message) Write-Host "⚠️  $Message" -ForegroundColor Yellow }

# OS and architecture detection
function Detect-Platform {
    $arch = $env:PROCESSOR_ARCHITECTURE

    switch ($arch) {
        "AMD64" { $script:Arch = "amd64" }
        "ARM64" { $script:Arch = "arm64" }
        default {
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }

    Write-Info "Detected platform: windows-$script:Arch"
}

# Dependency check
function Check-Dependencies {
    try {
        $null = Invoke-WebRequest -Uri "https://github.com" -Method HEAD -TimeoutSec 5
    } catch {
        Write-Error "Unable to connect to the internet"
        exit 1
    }
}

# Get latest version
function Get-LatestVersion {
    Write-Info "Fetching latest release..."

    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -TimeoutSec 10
        $script:Version = $release.tag_name
        Write-Info "Version: $script:Version"
    } catch {
        Write-Error "Unable to fetch latest version"
        exit 1
    }
}

# Installation
function Install-xsh {
    $assetName = "xsh-windows-$script:Arch.exe"
    $downloadUrl = "https://github.com/$Repo/releases/download/$script:Version/$assetName"

    Write-Info "Downloading from: $downloadUrl"

    # Create installation directory
    if (!(Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $tempFile = "$env:TEMP\$BinaryName"

    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -TimeoutSec 60
    } catch {
        Write-Error "Download failed: $_"
        exit 1
    }

    # Move to installation directory
    $installPath = "$InstallDir\$BinaryName"
    Move-Item -Path $tempFile -Destination $installPath -Force

    Write-Success "xsh installed in $installPath"

    # Check if directory is in PATH
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($userPath -notlike "*$InstallDir*") {
        Write-Info "Adding $InstallDir to user PATH..."
        [Environment]::SetEnvironmentVariable("PATH", "$userPath;$InstallDir", "User")
        Write-Success "PATH updated. Restart PowerShell to use 'xsh'"
    } else {
        Write-Info "Directory is already in PATH"
    }
}

# Welcome message
function Show-Welcome {
    Write-Host ""
    Write-Host "╔══════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║                                      ║" -ForegroundColor Cyan
    Write-Host "║   xsh - Twitter/X CLI Installer      ║" -ForegroundColor Cyan
    Write-Host "║           (Windows/PowerShell)       ║" -ForegroundColor Cyan
    Write-Host "║                                      ║" -ForegroundColor Cyan
    Write-Host "╚══════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
}

# Post-installation message
function Show-PostInstall {
    Write-Host ""
    Write-Success "Installation complete!"
    Write-Host ""
    Write-Host "Quick commands:"
    Write-Host "  xsh auth login       # Authentication"
    Write-Host "  xsh feed             # Timeline"
    Write-Host "  xsh --help           # Help"
    Write-Host ""
    Write-Host "Documentation: https://github.com/$Repo"
    Write-Host ""
    Write-Warning "Restart PowerShell to use 'xsh' if this is the first installation"
    Write-Host ""
}

# Main
Show-Welcome
Check-Dependencies
Detect-Platform
Get-LatestVersion
Install-xsh
Show-PostInstall
