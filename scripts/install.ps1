#
# install.ps1 - Script d'installation de xsh pour Windows (PowerShell)
# Usage: iwr -useb https://raw.githubusercontent.com/benoitpetit/xsh/main/core/scripts/install.ps1 | iex
#
# Ce script télécharge et installe xsh automatiquement sur Windows
#

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "benoitpetit/xsh"
$BinaryName = "xsh.exe"
$InstallDir = "$env:LOCALAPPDATA\xsh"

# Couleurs
function Write-Success { param($Message) Write-Host "✅ $Message" -ForegroundColor Green }
function Write-Error { param($Message) Write-Host "❌ $Message" -ForegroundColor Red }
function Write-Info { param($Message) Write-Host "ℹ️  $Message" -ForegroundColor Cyan }
function Write-Warning { param($Message) Write-Host "⚠️  $Message" -ForegroundColor Yellow }

# Détection de l'OS et architecture
function Detect-Platform {
    $arch = $env:PROCESSOR_ARCHITECTURE
    
    switch ($arch) {
        "AMD64" { $script:Arch = "amd64" }
        "ARM64" { $script:Arch = "arm64" }
        default { 
            Write-Error "Architecture non supportée: $arch"
            exit 1
        }
    }
    
    Write-Info "Plateforme détectée: windows-$script:Arch"
}

# Vérification des dépendances
function Check-Dependencies {
    try {
        $null = Invoke-WebRequest -Uri "https://github.com" -Method HEAD -TimeoutSec 5
    } catch {
        Write-Error "Impossible de se connecter à internet"
        exit 1
    }
}

# Récupération de la dernière version
function Get-LatestVersion {
    Write-Info "Récupération de la dernière release..."
    
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -TimeoutSec 10
        $script:Version = $release.tag_name
        Write-Info "Version: $script:Version"
    } catch {
        Write-Error "Impossible de récupérer la dernière version"
        exit 1
    }
}

# Installation
function Install-xsh {
    $assetName = "xsh-windows-$script:Arch.exe"
    $downloadUrl = "https://github.com/$Repo/releases/download/$script:Version/$assetName"
    
    Write-Info "Téléchargement depuis: $downloadUrl"
    
    # Créer le répertoire d'installation
    if (!(Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    $tempFile = "$env:TEMP\$BinaryName"
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -TimeoutSec 60
    } catch {
        Write-Error "Échec du téléchargement: $_"
        exit 1
    }
    
    # Déplacer vers le répertoire d'installation
    $installPath = "$InstallDir\$BinaryName"
    Move-Item -Path $tempFile -Destination $installPath -Force
    
    Write-Success "xsh installé dans $installPath"
    
    # Vérifier si le répertoire est dans le PATH
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($userPath -notlike "*$InstallDir*") {
        Write-Info "Ajout de $InstallDir au PATH utilisateur..."
        [Environment]::SetEnvironmentVariable("PATH", "$userPath;$InstallDir", "User")
        Write-Success "PATH mis à jour. Redémarrez PowerShell pour utiliser 'xsh'"
    } else {
        Write-Info "Le répertoire est déjà dans le PATH"
    }
}

# Message de bienvenue
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

# Message post-installation
function Show-PostInstall {
    Write-Host ""
    Write-Success "Installation terminée !"
    Write-Host ""
    Write-Host "Commandes rapides:"
    Write-Host "  xsh auth login       # Authentification"
    Write-Host "  xsh feed             # Timeline"
    Write-Host "  xsh --help           # Aide"
    Write-Host ""
    Write-Host "Documentation: https://github.com/$Repo"
    Write-Host ""
    Write-Warning "Redémarrez PowerShell pour utiliser 'xsh' si c'est la première installation"
    Write-Host ""
}

# Main
Show-Welcome
Check-Dependencies
Detect-Platform
Get-LatestVersion
Install-xsh
Show-PostInstall
