$ErrorActionPreference = "Stop"

$Repository = "futrx-com/remote.futrx-cli"
$Version = if ($env:REMOTE_VERSION) { $env:REMOTE_VERSION } else { "latest" }
$ReleaseBaseUrl = if ($env:REMOTE_RELEASE_BASE_URL) {
    $env:REMOTE_RELEASE_BASE_URL.TrimEnd("/")
} else {
    "https://github.com/$Repository/releases"
}

function Fail([string]$Message) {
    throw "remote installer: $Message"
}

if ($Version -ne "latest" -and $Version -notmatch '^[A-Za-z0-9._-]+$') {
    Fail "invalid version: $Version"
}

$Architecture = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()) {
    "X64" { "amd64" }
    "Arm64" { "arm64" }
    default { Fail "unsupported architecture: $_" }
}

$DownloadRoot = if ($Version -eq "latest") {
    "$ReleaseBaseUrl/latest/download"
} else {
    "$ReleaseBaseUrl/download/$Version"
}

$InstallDir = if ($env:REMOTE_INSTALL_DIR) {
    $env:REMOTE_INSTALL_DIR
} else {
    Join-Path ([Environment]::GetFolderPath("LocalApplicationData")) "Programs\Remote"
}

$Asset = "remote-windows-$Architecture.exe"
$TemporaryDir = Join-Path ([System.IO.Path]::GetTempPath()) ("remote-install-" + [guid]::NewGuid())

try {
    New-Item -ItemType Directory -Path $TemporaryDir | Out-Null
    $BinaryPath = Join-Path $TemporaryDir $Asset
    $ChecksumsPath = Join-Path $TemporaryDir "checksums.txt"

    Invoke-WebRequest -UseBasicParsing -Uri "$DownloadRoot/$Asset" -OutFile $BinaryPath
    Invoke-WebRequest -UseBasicParsing -Uri "$DownloadRoot/checksums.txt" -OutFile $ChecksumsPath

    $ChecksumLine = Get-Content $ChecksumsPath | Where-Object {
        $_ -match "^([0-9A-Fa-f]{64})\s+\*?$([regex]::Escape($Asset))$"
    } | Select-Object -First 1
    if (-not $ChecksumLine) {
        Fail "checksums.txt has no valid checksum for $Asset"
    }

    $ExpectedChecksum = ($ChecksumLine -split '\s+')[0]
    $ActualChecksum = (Get-FileHash -Algorithm SHA256 $BinaryPath).Hash
    if ($ActualChecksum -ne $ExpectedChecksum) {
        Fail "checksum verification failed for $Asset"
    }

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $Destination = Join-Path $InstallDir "remote.exe"
    Copy-Item -Force $BinaryPath $Destination

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $PathEntries = @($UserPath -split ';' | Where-Object { $_ })
    if ($PathEntries -notcontains $InstallDir) {
        $NewUserPath = (@($PathEntries) + $InstallDir) -join ';'
        [Environment]::SetEnvironmentVariable("Path", $NewUserPath, "User")
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Added $InstallDir to your user PATH. Open a new terminal to use it."
    }

    Write-Host "Installed remote to $Destination"
} finally {
    if (Test-Path $TemporaryDir) {
        Remove-Item -Recurse -Force $TemporaryDir
    }
}
