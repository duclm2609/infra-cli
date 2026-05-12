$ErrorActionPreference = "Stop"

$Repo = "duclm2609/infra-cli"
$InstallDir = "$env:LOCALAPPDATA\infra"

# Detect architecture
$Arch = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else {
  Write-Error "Unsupported architecture"; exit 1
}

$Platform = "windows-$Arch"

# Get latest release tag
$Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
$Version = $Release.tag_name
if (-not $Version) {
  Write-Error "Failed to fetch latest release"; exit 1
}

$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/infra-$Platform.zip"

Write-Host "Installing infra $Version ($Platform)..."

$Tmp = New-TemporaryFile | ForEach-Object { $_.DirectoryName + "\" + $_.BaseName }
New-Item -ItemType Directory -Path $Tmp -Force | Out-Null

try {
  $ZipPath = "$Tmp\infra.zip"
  Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing
  Expand-Archive -Path $ZipPath -DestinationPath $Tmp -Force

  New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
  Move-Item -Path "$Tmp\infra.exe" -Destination "$InstallDir\infra.exe" -Force
} finally {
  Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
}

# Add to PATH if not already present
$UserPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
  [System.Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
  $env:PATH += ";$InstallDir"
  Write-Host "Added $InstallDir to PATH"
}

Write-Host "Installed to $InstallDir\infra.exe"
& "$InstallDir\infra.exe" --version
