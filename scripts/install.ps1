# image-cli Windows installer（自动 GitHub / 镜像加速，适配国内网络）
# Usage:
#   irm https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.ps1 | iex
#   irm https://ghproxy.net/https://raw.githubusercontent.com/neko233-com/image-cli/main/scripts/install.ps1 | iex
#
# 环境变量:
#   IMAGE_CLI_DOWNLOAD_SOURCE = auto | github | mirror  （默认 auto）
#   IMAGE_CLI_UPDATE_MIRRORS  = 镜像前缀列表（逗号分隔，默认 gh-proxy/ghfast/ghproxy）
#   IMAGE_CLI_PREFER_MIRROR   = 1 时 auto 模式优先走镜像
#   IMAGE_CLI_DIRECT_ONLY     = 1 时仅直连 GitHub
# 一键安装: irm .../install.ps1 | iex  （脚本内勿使用 CmdletBinding，否则 iex 会失败）

param(
    [string]$Version = $(if ($env:IMAGE_CLI_VERSION) { $env:IMAGE_CLI_VERSION } else { "latest" }),
    [string]$Repo = $(if ($env:IMAGE_CLI_REPO) { $env:IMAGE_CLI_REPO } else { "neko233-com/image-cli" }),
    [string]$InstallDir = $(if ($env:IMAGE_CLI_INSTALL_DIR) { $env:IMAGE_CLI_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "image-cli" }),
    [ValidateSet("auto", "github", "mirror")]
    [string]$Source = $(if ($env:IMAGE_CLI_DOWNLOAD_SOURCE) { $env:IMAGE_CLI_DOWNLOAD_SOURCE } else { "auto" }),
    [switch]$NoPathUpdate
)

$ErrorActionPreference = "Stop"
$BinaryName = "image-cli"
$UserAgent = "image-cli-installer"

function Fail([string]$Message) {
    throw "image-cli installer: $Message"
}

function Get-MirrorPrefixes {
    $raw = if ($null -ne $env:IMAGE_CLI_UPDATE_MIRRORS) {
        $env:IMAGE_CLI_UPDATE_MIRRORS
    } else {
        "https://gh-proxy.com/,https://ghfast.top/,https://ghproxy.net/,https://mirror.ghproxy.com/"
    }
    $list = [System.Collections.Generic.List[string]]::new()
    foreach ($part in ($raw -split ",")) {
        $part = $part.Trim()
        if ([string]::IsNullOrWhiteSpace($part)) { continue }
        if ($part -notmatch "^https://") { continue }
        $list.Add($part.TrimEnd("/"))
    }
    return $list
}

function Expand-Routes([string]$OriginalUrl) {
    $routes = [System.Collections.Generic.List[string]]::new()
    $seen = @{}

    function Add-Route([string]$Url) {
        if ([string]::IsNullOrWhiteSpace($Url)) { return }
        if ($seen.ContainsKey($Url)) { return }
        $seen[$Url] = $true
        [void]$routes.Add($Url)
    }

    $mirrors = Get-MirrorPrefixes
    $preferMirror = ($Source -eq "mirror") -or ($env:IMAGE_CLI_PREFER_MIRROR -eq "1")

    switch ($Source) {
        "github" {
            Add-Route $OriginalUrl
        }
        "mirror" {
            foreach ($mirror in $mirrors) {
                if ($mirror -match "\{url\}") {
                    Add-Route ($mirror -replace "\{url\}", $OriginalUrl)
                } else {
                    Add-Route ($mirror + "/" + $OriginalUrl)
                }
            }
        }
        default {
            if ($env:IMAGE_CLI_DIRECT_ONLY -eq "1") {
                Add-Route $OriginalUrl
            } elseif ($preferMirror) {
                foreach ($mirror in $mirrors) {
                    if ($mirror -match "\{url\}") {
                        Add-Route ($mirror -replace "\{url\}", $OriginalUrl)
                    } else {
                        Add-Route ($mirror + "/" + $OriginalUrl)
                    }
                }
                Add-Route $OriginalUrl
            } else {
                Add-Route $OriginalUrl
                foreach ($mirror in $mirrors) {
                    if ($mirror -match "\{url\}") {
                        Add-Route ($mirror -replace "\{url\}", $OriginalUrl)
                    } else {
                        Add-Route ($mirror + "/" + $OriginalUrl)
                    }
                }
            }
        }
    }
    return $routes
}

function Test-Route([string]$Url) {
    $headers = @{
        "User-Agent" = $UserAgent
        "Accept"     = "*/*"
    }
    foreach ($mode in @("head", "range", "gettiny")) {
        try {
            switch ($mode) {
                "head" {
                    $response = Invoke-WebRequest -UseBasicParsing -Method Head -Uri $Url -Headers $headers -TimeoutSec 12 -MaximumRedirection 8
                }
                "range" {
                    $headersRange = @{
                        "User-Agent" = $UserAgent
                        "Accept"     = "*/*"
                        "Range"      = "bytes=0-8191"
                    }
                    $response = Invoke-WebRequest -UseBasicParsing -Method Get -Uri $Url -Headers $headersRange -TimeoutSec 12 -MaximumRedirection 8
                }
                default {
                    $response = Invoke-WebRequest -UseBasicParsing -Method Get -Uri $Url -Headers $headers -TimeoutSec 12 -MaximumRedirection 8
                }
            }
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 400) {
                return $true
            }
        } catch {
            continue
        }
    }
    return $false
}

function Select-Routes([string]$OriginalUrl) {
    $routes = Expand-Routes $OriginalUrl
    $probed = [System.Collections.Generic.List[object]]::new()
    foreach ($candidate in $routes) {
        $watch = [System.Diagnostics.Stopwatch]::StartNew()
        if (Test-Route $candidate) {
            $watch.Stop()
            [void]$probed.Add([pscustomobject]@{ Url = $candidate; Ms = $watch.Elapsed.TotalMilliseconds })
        }
    }
    if ($probed.Count -gt 0) {
        return @($probed | Sort-Object Ms | ForEach-Object { $_.Url })
    }
    return @($routes)
}

function Invoke-DownloadFile([string]$OriginalUrl, [string]$Destination) {
    $ordered = Select-Routes $OriginalUrl
    $lastError = $null
    foreach ($url in $ordered) {
        try {
            Invoke-WebRequest -UseBasicParsing -Method Get -Uri $url -Headers @{ "User-Agent" = $UserAgent } -OutFile $Destination -TimeoutSec 300 -MaximumRedirection 8
            Write-Host "Download route: $url"
            return $url
        } catch {
            $lastError = $_
        }
    }
    Fail ("unable to download " + $OriginalUrl + " (tried " + $ordered.Count + " routes): " + $lastError.Exception.Message)
}

function Invoke-ReleaseJson([string]$OriginalApiUrl) {
    $ordered = Select-Routes $OriginalApiUrl
    $headers = @{
        "User-Agent" = $UserAgent
        "Accept"     = "application/vnd.github+json"
    }
    $lastError = $null
    foreach ($url in $ordered) {
        try {
            $release = Invoke-RestMethod -Uri $url -Headers $headers -TimeoutSec 30 -MaximumRedirection 8
            Write-Host "Release metadata route: $url"
            return $release
        } catch {
            $lastError = $_
        }
    }
    Fail ("unable to fetch latest release metadata: " + $lastError.Exception.Message)
}

try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 -bor [Net.SecurityProtocolType]::Tls13
} catch {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
}

if ([string]::IsNullOrWhiteSpace($Version) -or $Version -eq "latest") {
    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
    $release = Invoke-ReleaseJson $apiUrl
    $Version = $release.tag_name
}
if (-not $Version.StartsWith("v")) { $Version = "v$Version" }

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
    Write-Host "Installing image-cli $Version (windows/$archName) [source=$Source]"
    $binary = Join-Path $tempDir $asset
    $checksums = Join-Path $tempDir "SHA256SUMS"

    $assetDownloaded = Invoke-DownloadFile "$baseUrl/$asset" $binary
    $routeBase = $assetDownloaded.Substring(0, $assetDownloaded.Length - ("/" + $asset).Length)
    try {
        Invoke-WebRequest -UseBasicParsing -Method Get -Uri "$routeBase/SHA256SUMS" -Headers @{ "User-Agent" = $UserAgent } -OutFile $checksums -TimeoutSec 120 -MaximumRedirection 8
    } catch {
        Invoke-DownloadFile "$baseUrl/SHA256SUMS" $checksums | Out-Null
    }

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
