# image-cli Windows installer（GitHub 镜像加速）
# Usage:
#   irm https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.ps1 | iex
#   irm https://ghproxy.net/https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.ps1 | iex

[CmdletBinding()]
param(
    [string]$Version = $(if ($env:IMAGE_CLI_VERSION) { $env:IMAGE_CLI_VERSION } else { "latest" }),
    [string]$Repo = $(if ($env:IMAGE_CLI_REPO) { $env:IMAGE_CLI_REPO } else { "neko233-com/image-cli" }),
    [string]$InstallDir = $(if ($env:IMAGE_CLI_INSTALL_DIR) { $env:IMAGE_CLI_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "image-cli" }),
    [switch]$NoPathUpdate
)

$ErrorActionPreference = "Stop"
$BinaryName = "image-cli"

function Fail([string]$Message) {
    throw "image-cli installer: $Message"
}

function Get-CandidateUrls([string]$OriginalUrl) {
    $urls = [System.Collections.Generic.List[string]]::new()
    $urls.Add($OriginalUrl)
    if ($env:IMAGE_CLI_DIRECT_ONLY -eq "1") { return $urls }
    $mirrors = if ($null -ne $env:IMAGE_CLI_UPDATE_MIRRORS) { $env:IMAGE_CLI_UPDATE_MIRRORS } else { "https://gh-proxy.com/,https://ghfast.top/,https://ghproxy.net/" }
    foreach ($mirror in ($mirrors -split ",")) {
        if ([string]::IsNullOrWhiteSpace($mirror)) { continue }
        $urls.Add(($mirror.TrimEnd("/") + "/" + $OriginalUrl))
    }
    return $urls
}

function Get-FastestUrl([string]$OriginalUrl) {
    $best = $null
    $bestSeconds = [double]::PositiveInfinity
    foreach ($candidate in (Get-CandidateUrls $OriginalUrl)) {
        $watch = [System.Diagnostics.Stopwatch]::StartNew()
        try {
            $response = Invoke-WebRequest -UseBasicParsing -Method Get -Uri $candidate -Headers @{ Range = "bytes=0-65535" } -MaximumRedirection 5 -TimeoutSec 8
            $watch.Stop()
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 400 -and $watch.Elapsed.TotalSeconds -lt $bestSeconds) {
                $best = $candidate
                $bestSeconds = $watch.Elapsed.TotalSeconds
            }
        } catch {
            $watch.Stop()
        }
    }
    if (-not $best) { Fail "no reachable route for $OriginalUrl" }
    return $best
}

if ([string]::IsNullOrWhiteSpace($Version) -or $Version -eq "latest") {
    $apiUrl = Get-FastestUrl "https://api.github.com/repos/$Repo/releases/latest"
    $release = Invoke-RestMethod -Headers @{ "User-Agent" = "image-cli-installer" } -Uri $apiUrl
    $Version = $release.tag_name
}
if (-not $Version.StartsWith("v")) { $Version = "v$Version" }
$versionNumber = $Version.TrimStart("v")

$arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
switch ($arch) {
    "x64" { $archName = "amd64" }
    "arm64" { $archName = "arm64" }
    default { Fail "unsupported architecture: $arch" }
}

$asset = "${BinaryName}-windows-${archName}.exe"
$baseUrl = "https://github.com/$Repo/releases/download/$Version"
$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("image-cli-install-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tempDir | Out-Null

try {
    Write-Host "Installing image-cli $Version (windows/$archName)"
    $binary = Join-Path $tempDir $asset
    $checksums = Join-Path $tempDir "SHA256SUMS"
    $assetUrl = Get-FastestUrl "$baseUrl/$asset"
    $routeBase = $assetUrl.Substring(0, $assetUrl.Length - ("/" + $asset).Length)
    Invoke-WebRequest -Headers @{ "User-Agent" = "image-cli-installer" } -Uri "$routeBase/$asset" -OutFile $binary
    Invoke-WebRequest -Headers @{ "User-Agent" = "image-cli-installer" } -Uri "$routeBase/SHA256SUMS" -OutFile $checksums

    $line = Get-Content $checksums | Where-Object { $_ -match "\s$([regex]::Escape($asset))$" -or $_ -match "\sdist/$([regex]::Escape($asset))$" } | Select-Object -First 1
    if (-not $line) { Fail "checksum missing for $asset" }
    $expected = ($line -split "\s+")[0].ToLowerInvariant()
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $binary).Hash.ToLowerInvariant()
    if ($expected -ne $actual) { Fail "checksum verification failed" }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $dest = Join-Path $InstallDir "${BinaryName}.exe"
    Copy-Item -LiteralPath $binary -Destination $dest -Force
    Write-Host "Installed: $dest"

    if (-not $NoPathUpdate) {
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        $parts = @($userPath -split ";" | Where-Object { $_ })
        if ($parts -notcontains $InstallDir) {
            [Environment]::SetEnvironmentVariable("Path", (($parts + $InstallDir) -join ";"), "User")
            Write-Host "Added install directory to user PATH. Open a new terminal."
        }
    }
}
finally {
    Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "Run: image-cli version"
